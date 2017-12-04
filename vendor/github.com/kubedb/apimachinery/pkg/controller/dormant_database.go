package controller

import (
	"errors"
	"reflect"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/wait"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Deleter interface {
	// Check Database CRD
	Exists(*metav1.ObjectMeta) (bool, error)
	// Pause operation
	PauseDatabase(*api.DormantDatabase) error
	// Wipe out operation
	WipeOutDatabase(*api.DormantDatabase) error
	// Resume operation
	ResumeDatabase(*api.DormantDatabase) error
}

type DormantDbController struct {
	// Kubernetes client
	client kubernetes.Interface
	// Api Extension Client
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	// ThirdPartyExtension client
	extClient cs.KubedbV1alpha1Interface
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
	client kubernetes.Interface,
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface,
	extClient cs.KubedbV1alpha1Interface,
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

	resourceName := api.ResourceTypeDormantDatabase + "." + api.SchemeGroupVersion.Group
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
				Plural:     api.ResourceTypeDormantDatabase,
				Kind:       api.ResourceKindDormantDatabase,
				ShortNames: []string{api.ResourceCodeDormantDatabase},
			},
		},
	}

	if _, err = c.apiExtKubeClient.CustomResourceDefinitions().Create(crd); err != nil {
		log.Fatalln(err)
	}
}

func (c *DormantDbController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&api.DormantDatabase{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				dormantDb := obj.(*api.DormantDatabase)
				util.AssignTypeKind(dormantDb)
				if dormantDb.Status.CreationTime == nil {
					if err := c.create(dormantDb); err != nil {
						log.Errorln(err)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				dormantDb := obj.(*api.DormantDatabase)
				util.AssignTypeKind(dormantDb)
				if err := c.delete(dormantDb); err != nil {
					log.Errorln(err)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldDormantDb, ok := old.(*api.DormantDatabase)
				if !ok {
					return
				}
				newDormantDb, ok := new.(*api.DormantDatabase)
				if !ok {
					return
				}
				// TODO: Find appropriate checking
				// Only allow if Spec varies
				util.AssignTypeKind(oldDormantDb)
				util.AssignTypeKind(newDormantDb)
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

func (c *DormantDbController) create(dormantDb *api.DormantDatabase) error {
	_, err := util.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
		t := metav1.Now()
		in.Status.CreationTime = &t
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
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
			core.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	_, err = util.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Status.Phase = api.DormantDatabasePhasePausing
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	c.recorder.Event(dormantDb, core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Database")

	// Pause Database workload
	if err := c.deleter.PauseDatabase(dormantDb); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to pause. Reason: %v",
			err,
		)
		return err
	}

	c.recorder.Event(
		dormantDb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulPause,
		"Successfully paused Database workload",
	)

	_, err = util.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
		t := metav1.Now()
		in.Status.PausingTime = &t
		in.Status.Phase = api.DormantDatabasePhasePaused
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *DormantDbController) delete(dormantDb *api.DormantDatabase) error {
	phase := dormantDb.Status.Phase
	if phase != api.DormantDatabasePhaseResuming && phase != api.DormantDatabasePhaseWipedOut {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			`DormantDatabase "%v" is not %v.`,
			dormantDb.Name,
			api.DormantDatabasePhaseWipedOut,
		)

		if err := c.reCreateDormantDatabase(dormantDb); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				core.EventTypeWarning,
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

func (c *DormantDbController) update(oldDormantDb, updatedDormantDb *api.DormantDatabase) error {
	if oldDormantDb.Spec.WipeOut != updatedDormantDb.Spec.WipeOut && updatedDormantDb.Spec.WipeOut {
		return c.wipeOut(updatedDormantDb)
	}

	if oldDormantDb.Spec.Resume != updatedDormantDb.Spec.Resume && updatedDormantDb.Spec.Resume {
		if oldDormantDb.Status.Phase == api.DormantDatabasePhasePaused {
			return c.resume(updatedDormantDb)
		} else {
			message := "Failed to resume Database. " +
				"Only DormantDatabase of \"Paused\" Phase can be resumed"
			c.recorder.Event(
				updatedDormantDb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				message,
			)
		}
	}
	return nil
}

func (c *DormantDbController) wipeOut(dormantDb *api.DormantDatabase) error {
	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
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
			core.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	_, err = util.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Status.Phase = api.DormantDatabasePhaseWipingOut
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Wipe out Database workload
	c.recorder.Event(dormantDb, core.EventTypeNormal, eventer.EventReasonWipingOut, "Wiping out Database")
	if err := c.deleter.WipeOutDatabase(dormantDb); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			"Failed to wipeOut. Reason: %v",
			err,
		)
		return err
	}

	c.recorder.Event(
		dormantDb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulWipeOut,
		"Successfully wiped out Database workload",
	)

	_, err = util.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
		t := metav1.Now()
		in.Status.WipeOutTime = &t
		in.Status.Phase = api.DormantDatabasePhaseWipedOut
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *DormantDbController) resume(dormantDb *api.DormantDatabase) error {
	c.recorder.Event(
		dormantDb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonResuming,
		"Resuming DormantDatabase",
	)

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
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
			core.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			message,
		)
		return errors.New(message)
	}

	_, err = util.TryPatchDormantDatabase(c.extClient, dormantDb.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Status.Phase = api.DormantDatabasePhaseResuming
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err = c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
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
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate DormantDatabase: "%v". Reason: %v`,
				dormantDb.Name,
				err,
			)
			return err
		}

		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			"Failed to resume Database. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *DormantDbController) reCreateDormantDatabase(dormantDb *api.DormantDatabase) error {
	_dormantDb := &api.DormantDatabase{
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
