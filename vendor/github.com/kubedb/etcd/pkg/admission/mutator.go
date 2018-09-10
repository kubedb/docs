package admission

import (
	"fmt"
	"sync"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	hookapi "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

type EtcdMutator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &EtcdMutator{}

func (a *EtcdMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "mutators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "etcds",
		},
		"etcd"
}

func (a *EtcdMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	if a.client, err = kubernetes.NewForConfig(config); err != nil {
		return err
	}
	if a.extClient, err = cs.NewForConfig(config); err != nil {
		return err
	}
	return err
}

func (a *EtcdMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindEtcd {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return hookapi.StatusUninitialized()
	}
	obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
	if err != nil {
		return hookapi.StatusBadRequest(err)
	}
	etcdMod, err := setDefaultValues(a.client, a.extClient, obj.(*api.Etcd).DeepCopy())
	if err != nil {
		return hookapi.StatusForbidden(err)
	} else if etcdMod != nil {
		patch, err := meta_util.CreateJSONPatch(obj, etcdMod)
		if err != nil {
			return hookapi.StatusInternalServerError(err)
		}
		status.Patch = patch
		patchType := admission.PatchTypeJSONPatch
		status.PatchType = &patchType
	}

	status.Allowed = true
	return status
}

// setDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a Etcd database
func setDefaultValues(client kubernetes.Interface, extClient cs.Interface, etcd *api.Etcd) (runtime.Object, error) {
	if etcd.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	if etcd.Spec.StorageType == "" {
		etcd.Spec.StorageType = api.StorageTypeDurable
	}

	if etcd.Spec.TerminationPolicy == "" {
		etcd.Spec.TerminationPolicy = api.TerminationPolicyPause
	}

	if etcd.Spec.Replicas == nil {
		etcd.Spec.Replicas = types.Int32P(1)
	}

	if err := setDefaultsFromDormantDB(extClient, etcd); err != nil {
		return nil, err
	}

	// If monitoring spec is given without port,
	// set default Listening port
	setMonitoringPort(etcd)

	return etcd, nil
}

// setDefaultsFromDormantDB takes values from Similar Dormant Database
func setDefaultsFromDormantDB(extClient cs.Interface, etcd *api.Etcd) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := extClient.KubedbV1alpha1().DormantDatabases(etcd.Namespace).Get(etcd.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}

	// Check DatabaseKind
	if value, _ := meta_util.GetStringValue(dormantDb.Labels, api.LabelDatabaseKind); value != api.ResourceKindEtcd {
		return errors.New(fmt.Sprintf(`invalid Etcd: "%v". Exists DormantDatabase "%v" of different Kind`, etcd.Name, dormantDb.Name))
	}

	// Check Origin Spec
	ddbOriginSpec := dormantDb.Spec.Origin.Spec.Etcd

	// If DatabaseSecret of new object is not given,
	// Take dormantDatabaseSecretName
	if etcd.Spec.DatabaseSecret == nil {
		etcd.Spec.DatabaseSecret = ddbOriginSpec.DatabaseSecret
	} else {
		ddbOriginSpec.DatabaseSecret = etcd.Spec.DatabaseSecret
	}

	// If Monitoring Spec of new object is not given,
	// Take Monitoring Settings from Dormant
	if etcd.Spec.Monitor == nil {
		etcd.Spec.Monitor = ddbOriginSpec.Monitor
	} else {
		ddbOriginSpec.Monitor = etcd.Spec.Monitor
	}

	// If Backup Scheduler of new object is not given,
	// Take Backup Scheduler Settings from Dormant
	if etcd.Spec.BackupSchedule == nil {
		etcd.Spec.BackupSchedule = ddbOriginSpec.BackupSchedule
	} else {
		ddbOriginSpec.BackupSchedule = etcd.Spec.BackupSchedule
	}

	// Skip checking DoNotPause
	ddbOriginSpec.DoNotPause = etcd.Spec.DoNotPause

	if !meta_util.Equal(ddbOriginSpec, &etcd.Spec) {
		diff := meta_util.Diff(ddbOriginSpec, &etcd.Spec)
		log.Errorf("etcd spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff)
		return errors.New(fmt.Sprintf("etcd spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff))
	}

	if _, err := meta_util.GetString(etcd.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		etcd.Spec.Init != nil &&
		etcd.Spec.Init.SnapshotSource != nil {
		etcd.Annotations = core_util.UpsertMap(etcd.Annotations, map[string]string{
			api.AnnotationInitialized: "",
		})
	}

	// Delete  Matching dormantDatabase in Controller

	return nil
}

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func setMonitoringPort(etcd *api.Etcd) {
	if etcd.Spec.Monitor != nil &&
		etcd.GetMonitoringVendor() == mona.VendorPrometheus {
		if etcd.Spec.Monitor.Prometheus == nil {
			etcd.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if etcd.Spec.Monitor.Prometheus.Port == 0 {
			etcd.Spec.Monitor.Prometheus.Port = 2379 //api.PrometheusExporterPortNumber
		}
	}
}
