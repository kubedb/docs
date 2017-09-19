package controller

import (
	"errors"
	"reflect"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/wait"
	kutildb "github.com/appscode/kutil/kubedb/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	tapi_v1alpha1 "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Deleter interface {
	// Check Database CRD
	Exists(*metav1.ObjectMeta) (bool, error)
	// Pause operation
	PauseDatabase(*tapi.DormantDatabase) error
	// Wipe out operation
	WipeOutDatabase(*tapi.DormantDatabase) error
	// Resume operation
	ResumeDatabase(*tapi.DormantDatabase) error
}

type DormantDbController struct {
	// Kubernetes client
	client clientset.Interface
	// Api Extension Client
	apiExtKubeClient apiextensionsclient.Interface
	// ThirdPartyExtension client
	extClient tcs.KubedbV1alpha1Interface
	// Deleter interface
	deleter Deleter
	// ListerWatcher
	lw *cache.ListWatch
	// Event Recorder
	recorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
}

// NewDormantDbController creates a new DormantDatabase Controller
func NewDormantDbController(
	client clientset.Interface,
	apiExtKubeClient apiextensionsclient.Interface,
	extClient tcs.KubedbV1alpha1Interface,
	deleter Deleter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *DormantDbController {
	// return new DormantDatabase Controller
	return &DormantDbController{
		client:           client,
		apiExtKubeClient: apiExtKubeClient,
		extClient:        extClient,
		deleter:          deleter,
		lw:               lw,
		recorder:         eventer.NewEventRecorder(client, "DormantDatabase Controller"),
		syncPeriod:       syncPeriod,
	}
}

func (c *DormantDbController) Run() {
	// Ensure DormantDatabase CRD
	c.ensureCustomResourceDefinition()
	// Watch DormantDatabase with provided ListerWatcher
	c.watch()
}

// Ensure DormantDatabase CustomResourceDefinition
func (c *DormantDbController) ensureCustomResourceDefinition() {
	log.Infoln("Ensuring DormantDatabase CustomResourceDefinition")

	resourceName := tapi.ResourceTypeDormantDatabase + "." + tapi_v1alpha1.SchemeGroupVersion.Group
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
			Group:   tapi_v1alpha1.SchemeGroupVersion.Group,
			Version: tapi_v1alpha1.SchemeGroupVersion.Version,
			Scope:   extensionsobj.NamespaceScoped,
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural:     tapi.ResourceTypeDormantDatabase,
				Kind:       tapi.ResourceKindDormantDatabase,
				ShortNames: []string{tapi.ResourceCodeDormantDatabase},
			},
		},
	}

	if _, err = c.apiExtKubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
		log.Fatalln(err)
	}
}

func (c *DormantDbController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&tapi.DormantDatabase{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				dormantDb := obj.(*tapi.DormantDatabase)
				if dormantDb.Status.CreationTime == nil {
					if err := c.create(dormantDb); err != nil {
						log.Errorln(err)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if err := c.delete(obj.(*tapi.DormantDatabase)); err != nil {
					log.Errorln(err)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldDormantDb, ok := old.(*tapi.DormantDatabase)
				if !ok {
					return
				}
				newDormantDb, ok := new.(*tapi.DormantDatabase)
				if !ok {
					return
				}
				// TODO: Find appropriate checking
				// Only allow if Spec varies
				if !reflect.DeepEqual(oldDormantDb.Spec, newDormantDb.Spec) {
					if err := c.update(oldDormantDb, newDormantDb); err != nil {
						log.Errorln(err)
					}
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

func (c *DormantDbController) create(dormantDb *tapi.DormantDatabase) error {
	_, err := kutildb.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
		t := metav1.Now()
		in.Status.CreationTime = &t
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			"Failed to pause Database. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "Failed to pause Database. Delete Database TPR object first"
		c.recorder.Event(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	_, err = kutildb.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
		in.Status.Phase = tapi.DormantDatabasePhasePausing
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	c.recorder.Event(dormantDb, apiv1.EventTypeNormal, eventer.EventReasonPausing, "Pausing Database")

	// Pause Database workload
	if err := c.deleter.PauseDatabase(dormantDb); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to pause. Reason: %v",
			err,
		)
		return err
	}

	c.recorder.Event(
		dormantDb.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulPause,
		"Successfully paused Database workload",
	)

	_, err = kutildb.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
		t := metav1.Now()
		in.Status.PausingTime = &t
		in.Status.Phase = tapi.DormantDatabasePhasePaused
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *DormantDbController) delete(dormantDb *tapi.DormantDatabase) error {
	phase := dormantDb.Status.Phase
	if phase != tapi.DormantDatabasePhaseResuming && phase != tapi.DormantDatabasePhaseWipedOut {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			`DormantDatabase "%v" is not %v.`,
			dormantDb.Name,
			tapi.DormantDatabasePhaseWipedOut,
		)

		if err := c.reCreateDormantDatabase(dormantDb); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate DormantDatabase: "%v". Reason: %v`,
				dormantDb.Name,
				err,
			)
			return err
		}
	}
	return nil
}

func (c *DormantDbController) update(oldDormantDb, updatedDormantDb *tapi.DormantDatabase) error {
	if oldDormantDb.Spec.WipeOut != updatedDormantDb.Spec.WipeOut && updatedDormantDb.Spec.WipeOut {
		return c.wipeOut(updatedDormantDb)
	}

	if oldDormantDb.Spec.Resume != updatedDormantDb.Spec.Resume && updatedDormantDb.Spec.Resume {
		if oldDormantDb.Status.Phase == tapi.DormantDatabasePhasePaused {
			return c.resume(updatedDormantDb)
		} else {
			message := "Failed to resume Database. " +
				"Only DormantDatabase of \"Paused\" Phase can be resumed"
			c.recorder.Event(
				updatedDormantDb.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				message,
			)
		}
	}
	return nil
}

func (c *DormantDbController) wipeOut(dormantDb *tapi.DormantDatabase) error {
	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to wipeOut Database. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "Failed to wipeOut Database. Delete Database TPR object first"
		c.recorder.Event(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	_, err = kutildb.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
		in.Status.Phase = tapi.DormantDatabasePhaseWipingOut
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Wipe out Database workload
	c.recorder.Event(dormantDb, apiv1.EventTypeNormal, eventer.EventReasonWipingOut, "Wiping out Database")
	if err := c.deleter.WipeOutDatabase(dormantDb); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			"Failed to wipeOut. Reason: %v",
			err,
		)
		return err
	}

	c.recorder.Event(
		dormantDb.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulWipeOut,
		"Successfully wiped out Database workload",
	)

	_, err = kutildb.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
		t := metav1.Now()
		in.Status.WipeOutTime = &t
		in.Status.Phase = tapi.DormantDatabasePhaseWipedOut
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *DormantDbController) resume(dormantDb *tapi.DormantDatabase) error {
	c.recorder.Event(
		dormantDb.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonResuming,
		"Resuming DormantDatabase",
	)

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			"Failed to resume DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "Failed to resume DormantDatabase. One Database TPR object exists with same name"
		c.recorder.Event(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			message,
		)
		return errors.New(message)
	}

	_, err = kutildb.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
		in.Status.Phase = tapi.DormantDatabasePhaseResuming
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err = c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	if err = c.deleter.ResumeDatabase(dormantDb); err != nil {
		if err := c.reCreateDormantDatabase(dormantDb); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate DormantDatabase: "%v". Reason: %v`,
				dormantDb.Name,
				err,
			)
			return err
		}

		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			"Failed to resume Database. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *DormantDbController) reCreateDormantDatabase(dormantDb *tapi.DormantDatabase) error {
	_dormantDb := &tapi.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:        dormantDb.Name,
			Namespace:   dormantDb.Namespace,
			Labels:      dormantDb.Labels,
			Annotations: dormantDb.Annotations,
		},
		Spec:   dormantDb.Spec,
		Status: dormantDb.Status,
	}

	if _, err := c.extClient.DormantDatabases(_dormantDb.Namespace).Create(_dormantDb); err != nil {
		return err
	}

	return nil
}
