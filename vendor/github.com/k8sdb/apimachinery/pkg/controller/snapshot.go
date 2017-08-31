package controller

import (
	"errors"
	"fmt"
	"time"

	"github.com/appscode/go/wait"
	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	"github.com/k8sdb/apimachinery/pkg/analytics"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	batch "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Snapshotter interface {
	ValidateSnapshot(*tapi.Snapshot) error
	GetDatabase(*tapi.Snapshot) (runtime.Object, error)
	GetSnapshotter(*tapi.Snapshot) (*batch.Job, error)
	WipeOutSnapshot(*tapi.Snapshot) error
}

type SnapshotController struct {
	// Kubernetes client
	client clientset.Interface
	// Api Extension Client
	apiExtKubeClient apiextensionsclient.Interface
	// ThirdPartyExtension client
	extClient tcs.ExtensionInterface
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
	client clientset.Interface,
	apiExtKubeClient apiextensionsclient.Interface,
	extClient tcs.ExtensionInterface,
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

	resourceName := tapi.ResourceTypeSnapshot + "." + tapi.V1alpha1SchemeGroupVersion.Group
	var err error
	if _, err = c.apiExtKubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(resourceName, metav1.GetOptions{}); err == nil {
		return
	}
	if !kerr.IsNotFound(err) {
		log.Fatalln(err)
	}

	crd := &extensionsobj.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   tapi.V1alpha1SchemeGroupVersion.Group,
			Version: tapi.V1alpha1SchemeGroupVersion.Version,
			Scope:   extensionsobj.NamespaceScoped,
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural:     tapi.ResourceTypeSnapshot,
				Kind:       tapi.ResourceKindSnapshot,
				ShortNames: []string{tapi.ResourceCodeSnapshot},
			},
		},
	}

	if _, err = c.apiExtKubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
		log.Fatalln(err)
	}
}

func (c *SnapshotController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&tapi.Snapshot{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				snapshot := obj.(*tapi.Snapshot)
				if snapshot.Status.StartTime == nil {
					if err := c.create(snapshot); err != nil {
						snapshotFailedToCreate()
						log.Errorln(err)
					} else {
						snapshotSuccessfullyCreated()
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				snapshot := obj.(*tapi.Snapshot)
				if err := c.delete(snapshot); err != nil {
					snapshotFailedToDelete()
					log.Errorln(err)
				} else {
					snapshotSuccessfullyDeleted()
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

const (
	durationCheckSnapshotJob = time.Minute * 30
)

func (c *SnapshotController) create(snapshot *tapi.Snapshot) error {
	err := c.UpdateSnapshot(snapshot.ObjectMeta, func(in tapi.Snapshot) tapi.Snapshot {
		t := metav1.Now()
		in.Status.StartTime = &t
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Validate DatabaseSnapshot spec
	if err := c.snapshoter.ValidateSnapshot(snapshot); err != nil {
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}

	// Check running snapshot
	if err := c.checkRunningSnapshot(snapshot); err != nil {
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, err.Error())
		return err
	}

	runtimeObj, err := c.snapshoter.GetDatabase(snapshot)
	if err != nil {
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonFailedToGet, err.Error())
		return err
	}

	err = c.UpdateSnapshot(snapshot.ObjectMeta, func(in tapi.Snapshot) tapi.Snapshot {
		in.Labels[tapi.LabelDatabaseName] = snapshot.Spec.DatabaseName
		in.Labels[tapi.LabelSnapshotStatus] = string(tapi.SnapshotPhaseRunning)
		in.Status.Phase = tapi.SnapshotPhaseRunning
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	c.eventRecorder.Event(runtimeObj, apiv1.EventTypeNormal, eventer.EventReasonStarting, "Backup running")
	c.eventRecorder.Event(snapshot, apiv1.EventTypeNormal, eventer.EventReasonStarting, "Backup running")

	secret, err := storage.NewOSMSecret(c.client, snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to generate osm secret. Reason: %v", err)
		c.eventRecorder.Event(runtimeObj, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}
	_, err = c.client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		message := fmt.Sprintf("Failed to create osm secret. Reason: %v", err)
		c.eventRecorder.Event(runtimeObj, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	job, err := c.snapshoter.GetSnapshotter(snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(runtimeObj, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}
	if _, err := c.client.BatchV1().Jobs(snapshot.Namespace).Create(job); err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(runtimeObj, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	go func() {
		if err := c.checkSnapshotJob(snapshot, job.Name, durationCheckSnapshotJob); err != nil {
			log.Errorln(err)
		}
	}()

	return nil
}

func (c *SnapshotController) delete(snapshot *tapi.Snapshot) error {
	runtimeObj, err := c.snapshoter.GetDatabase(snapshot)
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.eventRecorder.Event(
				snapshot,
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				err.Error(),
			)
			return err
		}
	}

	if runtimeObj != nil {
		c.eventRecorder.Eventf(
			runtimeObj,
			apiv1.EventTypeNormal,
			eventer.EventReasonWipingOut,
			"Wiping out Snapshot: %v",
			snapshot.Name,
		)
	}

	if err := c.snapshoter.WipeOutSnapshot(snapshot); err != nil {
		if runtimeObj != nil {
			c.eventRecorder.Eventf(
				runtimeObj,
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToWipeOut,
				"Failed to  wipeOut. Reason: %v",
				err,
			)
		}
		return err
	}

	if runtimeObj != nil {
		c.eventRecorder.Eventf(
			runtimeObj,
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulWipeOut,
			"Successfully wiped out Snapshot: %v",
			snapshot.Name,
		)
	}
	return nil
}

func (c *SnapshotController) checkRunningSnapshot(snapshot *tapi.Snapshot) error {
	labelMap := map[string]string{
		tapi.LabelDatabaseKind:   snapshot.Labels[tapi.LabelDatabaseKind],
		tapi.LabelDatabaseName:   snapshot.Spec.DatabaseName,
		tapi.LabelSnapshotStatus: string(tapi.SnapshotPhaseRunning),
	}

	snapshotList, err := c.extClient.Snapshots(snapshot.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	})
	if err != nil {
		return err
	}

	if len(snapshotList.Items) > 0 {
		err := c.UpdateSnapshot(snapshot.ObjectMeta, func(in tapi.Snapshot) tapi.Snapshot {
			t := metav1.Now()
			in.Status.StartTime = &t
			in.Status.CompletionTime = &t
			in.Status.Phase = tapi.SnapshotPhaseFailed
			in.Status.Reason = "One Snapshot is already Running"
			return in
		})
		if err != nil {
			c.eventRecorder.Eventf(snapshot, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		return errors.New("One Snapshot is already Running")
	}

	return nil
}

func (c *SnapshotController) checkSnapshotJob(snapshot *tapi.Snapshot, jobName string, checkDuration time.Duration) error {

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
				snapshot,
				apiv1.EventTypeWarning,
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

	err = c.client.CoreV1().Secrets(snapshot.Namespace).Delete(snapshot.Name, &metav1.DeleteOptions{})
	if err != nil && !kerr.IsNotFound(err) {
		c.eventRecorder.Eventf(
			snapshot,
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete Secret. Reason: %v",
			err,
		)
		log.Errorln(err)
	}

	runtimeObj, err := c.snapshoter.GetDatabase(snapshot)
	if err != nil {
		c.eventRecorder.Event(snapshot, apiv1.EventTypeWarning, eventer.EventReasonFailedToGet, err.Error())
		return err
	}

	if jobSuccess {
		snapshot.Status.Phase = tapi.SnapshotPhaseSuccessed
		c.eventRecorder.Event(
			runtimeObj,
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulSnapshot,
			"Successfully completed snapshot",
		)
		c.eventRecorder.Event(
			snapshot,
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulSnapshot,
			"Successfully completed snapshot",
		)
	} else {
		snapshot.Status.Phase = tapi.SnapshotPhaseFailed
		c.eventRecorder.Event(
			runtimeObj,
			apiv1.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			"Failed to complete snapshot",
		)
		c.eventRecorder.Event(
			snapshot,
			apiv1.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			"Failed to complete snapshot",
		)
	}

	err = c.UpdateSnapshot(snapshot.ObjectMeta, func(in tapi.Snapshot) tapi.Snapshot {
		t := metav1.Now()
		in.Status.CompletionTime = &t
		delete(in.Labels, tapi.LabelSnapshotStatus)
		in.Status.Phase = snapshot.Status.Phase
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func snapshotSuccessfullyCreated() {
	analytics.SendEvent(tapi.ResourceNameSnapshot, "created", "success")
}

func snapshotFailedToCreate() {
	analytics.SendEvent(tapi.ResourceNameSnapshot, "created", "failure")
}

func snapshotSuccessfullyDeleted() {
	analytics.SendEvent(tapi.ResourceNameSnapshot, "deleted", "success")
}

func snapshotFailedToDelete() {
	analytics.SendEvent(tapi.ResourceNameSnapshot, "deleted", "failure")
}
