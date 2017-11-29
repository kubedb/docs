package controller

import (
	"errors"
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/wait"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Snapshotter interface {
	ValidateSnapshot(*api.Snapshot) error
	GetDatabase(*api.Snapshot) (runtime.Object, error)
	GetSnapshotter(*api.Snapshot) (*batch.Job, error)
	WipeOutSnapshot(*api.Snapshot) error
}

type SnapshotController struct {
	// Kubernetes client
	client kubernetes.Interface
	// Api Extension Client
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	// ThirdPartyExtension client
	extClient cs.KubedbV1alpha1Interface
	// Snapshotter interface
	snapshoter Snapshotter
	// ListerWatcher
	lw *cache.ListWatch
	// Event Recorder
	eventRecorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
}

// NewSnapshotController creates a new SnapshotController
func NewSnapshotController(
	client kubernetes.Interface,
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface,
	extClient cs.KubedbV1alpha1Interface,
	snapshoter Snapshotter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *SnapshotController {

	// return new DormantDatabase Controller
	return &SnapshotController{
		client:           client,
		apiExtKubeClient: apiExtKubeClient,
		extClient:        extClient,
		snapshoter:       snapshoter,
		lw:               lw,
		eventRecorder:    eventer.NewEventRecorder(client, "Snapshot Controller"),
		syncPeriod:       syncPeriod,
	}
}

func (c *SnapshotController) Run() {
	// Ensure DormantDatabase TPR
	c.ensureCustomResourceDefinition()
	// Watch DormantDatabase with provided ListerWatcher
	c.watch()
}

// Ensure Snapshot CustomResourceDefinition
func (c *SnapshotController) ensureCustomResourceDefinition() {
	log.Infoln("Ensuring DormantDatabase CustomResourceDefinition")

	resourceName := api.ResourceTypeSnapshot + "." + api.SchemeGroupVersion.Group
	var err error
	if _, err = c.apiExtKubeClient.CustomResourceDefinitions().Get(resourceName, metav1.GetOptions{}); err == nil {
		return
	}
	if !kerr.IsNotFound(err) {
		log.Fatalln(err)
	}

	crd := &crd_api.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: crd_api.CustomResourceDefinitionSpec{
			Group:   api.SchemeGroupVersion.Group,
			Version: api.SchemeGroupVersion.Version,
			Scope:   crd_api.NamespaceScoped,
			Names: crd_api.CustomResourceDefinitionNames{
				Plural:     api.ResourceTypeSnapshot,
				Kind:       api.ResourceKindSnapshot,
				ShortNames: []string{api.ResourceCodeSnapshot},
			},
		},
	}

	if _, err = c.apiExtKubeClient.CustomResourceDefinitions().Create(crd); err != nil {
		log.Fatalln(err)
	}
}

func (c *SnapshotController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&api.Snapshot{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				snapshot := obj.(*api.Snapshot)
				util.AssignTypeKind(snapshot)
				if snapshot.Status.StartTime == nil {
					if err := c.create(snapshot); err != nil {
						log.Errorln(err)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				snapshot := obj.(*api.Snapshot)
				util.AssignTypeKind(snapshot)
				if err := c.delete(snapshot); err != nil {
					log.Errorln(err)
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

const (
	durationCheckSnapshotJob = time.Minute * 30
)

func (c *SnapshotController) create(snapshot *api.Snapshot) error {
	_, err := util.TryPatchSnapshot(c.extClient, snapshot.ObjectMeta, func(in *api.Snapshot) *api.Snapshot {
		t := metav1.Now()
		in.Status.StartTime = &t
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Validate DatabaseSnapshot spec
	if err := c.snapshoter.ValidateSnapshot(snapshot); err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}

	// Check running snapshot
	if err := c.checkRunningSnapshot(snapshot); err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, err.Error())
		return err
	}

	runtimeObj, err := c.snapshoter.GetDatabase(snapshot)
	if err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToGet, err.Error())
		return err
	}

	_, err = util.TryPatchSnapshot(c.extClient, snapshot.ObjectMeta, func(in *api.Snapshot) *api.Snapshot {
		in.Labels[api.LabelDatabaseName] = snapshot.Spec.DatabaseName
		in.Labels[api.LabelSnapshotStatus] = string(api.SnapshotPhaseRunning)
		in.Status.Phase = api.SnapshotPhaseRunning
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeNormal, eventer.EventReasonStarting, "Backup running")
	c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeNormal, eventer.EventReasonStarting, "Backup running")

	secret, err := storage.NewOSMSecret(c.client, snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to generate osm secret. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}
	_, err = c.client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		message := fmt.Sprintf("Failed to create osm secret. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	job, err := c.snapshoter.GetSnapshotter(snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}
	if _, err := c.client.BatchV1().Jobs(snapshot.Namespace).Create(job); err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	go func() {
		if err := c.checkSnapshotJob(snapshot, job.Name, durationCheckSnapshotJob); err != nil {
			log.Errorln(err)
		}
	}()

	return nil
}

func (c *SnapshotController) delete(snapshot *api.Snapshot) error {
	runtimeObj, err := c.snapshoter.GetDatabase(snapshot)
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.eventRecorder.Event(
				snapshot.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				err.Error(),
			)
			return err
		}
	}

	if runtimeObj != nil {
		c.eventRecorder.Eventf(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeNormal,
			eventer.EventReasonWipingOut,
			"Wiping out Snapshot: %v",
			snapshot.Name,
		)
	}

	if err := c.snapshoter.WipeOutSnapshot(snapshot); err != nil {
		if runtimeObj != nil {
			c.eventRecorder.Eventf(
				api.ObjectReferenceFor(runtimeObj),
				core.EventTypeWarning,
				eventer.EventReasonFailedToWipeOut,
				"Failed to  wipeOut. Reason: %v",
				err,
			)
		}
		return err
	}

	if runtimeObj != nil {
		c.eventRecorder.Eventf(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulWipeOut,
			"Successfully wiped out Snapshot: %v",
			snapshot.Name,
		)
	}
	return nil
}

func (c *SnapshotController) checkRunningSnapshot(snapshot *api.Snapshot) error {
	labelMap := map[string]string{
		api.LabelDatabaseKind:   snapshot.Labels[api.LabelDatabaseKind],
		api.LabelDatabaseName:   snapshot.Spec.DatabaseName,
		api.LabelSnapshotStatus: string(api.SnapshotPhaseRunning),
	}

	snapshotList, err := c.extClient.Snapshots(snapshot.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	})
	if err != nil {
		return err
	}

	if len(snapshotList.Items) > 0 {
		_, err = util.TryPatchSnapshot(c.extClient, snapshot.ObjectMeta, func(in *api.Snapshot) *api.Snapshot {
			t := metav1.Now()
			in.Status.StartTime = &t
			in.Status.CompletionTime = &t
			in.Status.Phase = api.SnapshotPhaseFailed
			in.Status.Reason = "One Snapshot is already Running"
			return in
		})
		if err != nil {
			c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		return errors.New("one Snapshot is already Running")
	}

	return nil
}

func (c *SnapshotController) checkSnapshotJob(snapshot *api.Snapshot, jobName string, checkDuration time.Duration) error {

	var jobSuccess bool = false
	var job *batch.Job
	var err error
	then := time.Now()
	now := time.Now()
	for now.Sub(then) < checkDuration {
		log.Debugln("Checking for Job ", jobName)
		job, err = c.client.BatchV1().Jobs(snapshot.Namespace).Get(jobName, metav1.GetOptions{})
		if err != nil {
			if kerr.IsNotFound(err) {
				time.Sleep(sleepDuration)
				now = time.Now()
				continue
			}
			c.eventRecorder.Eventf(
				snapshot.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToList,
				"Failed to get Job. Reason: %v",
				err,
			)
			return err
		}
		log.Debugf("Pods Statuses:	%d Running / %d Succeeded / %d Failed",
			job.Status.Active, job.Status.Succeeded, job.Status.Failed)
		// If job is success
		if job.Status.Succeeded > 0 {
			jobSuccess = true
			break
		} else if job.Status.Failed > 0 {
			break
		}

		time.Sleep(sleepDuration)
		now = time.Now()
	}

	if err != nil {
		return err
	}

	deleteJobResources(c.client, c.eventRecorder, snapshot, job)

	err = c.client.CoreV1().Secrets(snapshot.Namespace).Delete(snapshot.OSMSecretName(), &metav1.DeleteOptions{})
	if err != nil && !kerr.IsNotFound(err) {
		c.eventRecorder.Eventf(
			snapshot.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete Secret. Reason: %v",
			err,
		)
		log.Errorln(err)
	}

	runtimeObj, err := c.snapshoter.GetDatabase(snapshot)
	if err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToGet, err.Error())
		return err
	}

	if jobSuccess {
		snapshot.Status.Phase = api.SnapshotPhaseSuccessed
		c.eventRecorder.Event(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulSnapshot,
			"Successfully completed snapshot",
		)
		c.eventRecorder.Event(
			snapshot.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulSnapshot,
			"Successfully completed snapshot",
		)
	} else {
		snapshot.Status.Phase = api.SnapshotPhaseFailed
		c.eventRecorder.Event(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			"Failed to complete snapshot",
		)
		c.eventRecorder.Event(
			snapshot.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			"Failed to complete snapshot",
		)
	}

	_, err = util.TryPatchSnapshot(c.extClient, snapshot.ObjectMeta, func(in *api.Snapshot) *api.Snapshot {
		t := metav1.Now()
		in.Status.CompletionTime = &t
		delete(in.Labels, api.LabelSnapshotStatus)
		in.Status.Phase = snapshot.Status.Phase
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}
