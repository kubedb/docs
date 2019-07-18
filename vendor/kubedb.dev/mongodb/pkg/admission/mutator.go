package admission

import (
	"fmt"
	"sync"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
)

type MongoDBMutator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &MongoDBMutator{}

func (a *MongoDBMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "mutators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "mongodbmutators",
		},
		"mongodbmutator"
}

func (a *MongoDBMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *MongoDBMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindMongoDB {
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
	mongoMod, err := setDefaultValues(a.extClient, obj.(*api.MongoDB).DeepCopy(), req.Operation)
	if err != nil {
		return hookapi.StatusForbidden(err)
	} else if mongoMod != nil {
		patch, err := meta_util.CreateJSONPatch(req.Object.Raw, mongoMod)
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

// setDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a MongoDB database
func setDefaultValues(extClient cs.Interface, mongodb *api.MongoDB, op admission.Operation) (runtime.Object, error) {
	if mongodb.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	mongodb.SetDefaults()

	if err := setDefaultsFromDormantDB(extClient, mongodb, op); err != nil {
		return nil, err
	}

	// If monitoring spec is given without port,
	// set default Listening port
	setMonitoringPort(mongodb)

	return mongodb, nil
}

// setDefaultsFromDormantDB takes values from Similar Dormant Database
func setDefaultsFromDormantDB(extClient cs.Interface, mongodb *api.MongoDB, op admission.Operation) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := getDormantDB(extClient, mongodb)
	if err != nil {
		return err
	}

	// If dormantDb doesn't exist, then SetSecurityContext and return.
	if dormantDb == nil {
		if op != admission.Create {
			return nil
		}
		if mongodb.Spec.ShardTopology != nil {
			mongodb.Spec.SetSecurityContext(&mongodb.Spec.ShardTopology.Shard.PodTemplate)
			mongodb.Spec.SetSecurityContext(&mongodb.Spec.ShardTopology.ConfigServer.PodTemplate)
			mongodb.Spec.SetSecurityContext(&mongodb.Spec.ShardTopology.Mongos.PodTemplate)
		} else {
			mongodb.Spec.SetSecurityContext(mongodb.Spec.PodTemplate)
		}
		return nil
	}

	// Check Origin Spec
	ddbOriginSpec := dormantDb.Spec.Origin.Spec.MongoDB
	ddbOriginSpec.SetDefaults()

	// If DatabaseSecret of new object is not given,
	// Take dormantDatabaseSecretName
	if mongodb.Spec.DatabaseSecret == nil {
		mongodb.Spec.DatabaseSecret = ddbOriginSpec.DatabaseSecret
	}

	if mongodb.Spec.CertificateSecret == nil {
		mongodb.Spec.CertificateSecret = ddbOriginSpec.CertificateSecret
	}

	if mongodb.Spec.ConfigSource == nil {
		mongodb.Spec.ConfigSource = ddbOriginSpec.ConfigSource
	}

	// If Monitoring Spec of new object is not given,
	// Take Monitoring Settings from Dormant
	if mongodb.Spec.Monitor == nil {
		mongodb.Spec.Monitor = ddbOriginSpec.Monitor
	} else {
		ddbOriginSpec.Monitor = mongodb.Spec.Monitor
	}

	// If SecurityContext of new object is not given,
	// Take dormantDatabase's Security Context
	setSecurityContextFromDormantDB(mongodb, ddbOriginSpec)

	// If Backup Scheduler of new object is not given,
	// Take Backup Scheduler Settings from Dormant
	if mongodb.Spec.BackupSchedule == nil {
		mongodb.Spec.BackupSchedule = ddbOriginSpec.BackupSchedule
	} else {
		ddbOriginSpec.BackupSchedule = mongodb.Spec.BackupSchedule
	}

	// Skip checking UpdateStrategy
	ddbOriginSpec.UpdateStrategy = mongodb.Spec.UpdateStrategy

	// Skip checking TerminationPolicy
	ddbOriginSpec.TerminationPolicy = mongodb.Spec.TerminationPolicy

	if ddbOriginSpec.ShardTopology != nil && mongodb.Spec.ShardTopology != nil {
		// Skip checking strategy of mongos
		ddbOriginSpec.ShardTopology.Mongos.Strategy = mongodb.Spec.ShardTopology.Mongos.Strategy
		// Skip checking ServiceAccountName of ConfigServer
		ddbOriginSpec.ShardTopology.ConfigServer.PodTemplate.Spec.ServiceAccountName = mongodb.Spec.ShardTopology.ConfigServer.PodTemplate.Spec.ServiceAccountName
		// Skip checking ServiceAccountName of Mongos
		ddbOriginSpec.ShardTopology.Mongos.PodTemplate.Spec.ServiceAccountName = mongodb.Spec.ShardTopology.Mongos.PodTemplate.Spec.ServiceAccountName
		// Skip checking ServiceAccountName of Shard
		ddbOriginSpec.ShardTopology.Shard.PodTemplate.Spec.ServiceAccountName = mongodb.Spec.ShardTopology.Shard.PodTemplate.Spec.ServiceAccountName
	}

	if ddbOriginSpec.PodTemplate != nil && mongodb.Spec.PodTemplate != nil {
		// Skip checking ServiceAccountName
		ddbOriginSpec.PodTemplate.Spec.ServiceAccountName = mongodb.Spec.PodTemplate.Spec.ServiceAccountName
	}

	if !meta_util.Equal(ddbOriginSpec, &mongodb.Spec) {
		diff := meta_util.Diff(ddbOriginSpec, &mongodb.Spec)
		log.Errorf("mongodb spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff)
		return errors.New(fmt.Sprintf("mongodb spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff))
	}

	if _, err := meta_util.GetString(mongodb.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mongodb.Spec.Init != nil &&
		(mongodb.Spec.Init.SnapshotSource != nil || mongodb.Spec.Init.StashRestoreSession != nil) {
		mongodb.Annotations = core_util.UpsertMap(mongodb.Annotations, map[string]string{
			api.AnnotationInitialized: "",
		})
	}

	// Delete  Matching dormantDatabase in Controller

	return nil
}

// getDormantDB returns Dormant database that exists
// with same name in same namespace with label set to 'MongoDB'
func getDormantDB(extClient cs.Interface, mongodb *api.MongoDB) (*api.DormantDatabase, error) {
	// Check if DormantDatabase exists or not
	dormantDb, err := extClient.KubedbV1alpha1().DormantDatabases(mongodb.Namespace).Get(mongodb.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return nil, err
		}
		return nil, nil
	}

	// Check DatabaseKind
	if value, _ := meta_util.GetStringValue(dormantDb.Labels, api.LabelDatabaseKind); value != api.ResourceKindMongoDB {
		return nil, errors.New(fmt.Sprintf(`invalid MongoDB: "%v/%v". Exists DormantDatabase "%v/%v" of different Kind`, mongodb.Namespace, mongodb.Name, dormantDb.Namespace, dormantDb.Name))
	}
	return dormantDb, nil
}

func setSecurityContextFromDormantDB(mongodb *api.MongoDB, ddbOriginSpec *api.MongoDBSpec) {
	fn := func(mongoPt *ofst.PodTemplateSpec, drmPt *ofst.PodTemplateSpec) {
		if drmPt == nil || drmPt.Spec.SecurityContext == nil {
			return
		}
		if mongoPt.Spec.SecurityContext == nil {
			mongoPt.Spec.SecurityContext = new(core.PodSecurityContext)
		}
		if mongoPt.Spec.SecurityContext.FSGroup == nil {
			mongoPt.Spec.SecurityContext.FSGroup = drmPt.Spec.SecurityContext.FSGroup
		}
		if mongoPt.Spec.SecurityContext.RunAsNonRoot == nil {
			mongoPt.Spec.SecurityContext.RunAsNonRoot = drmPt.Spec.SecurityContext.RunAsNonRoot
		}
		if mongoPt.Spec.SecurityContext.RunAsUser == nil {
			mongoPt.Spec.SecurityContext.RunAsUser = drmPt.Spec.SecurityContext.RunAsUser
		}
	}

	if mongodb.Spec.ShardTopology != nil && ddbOriginSpec.ShardTopology != nil {
		fn(&mongodb.Spec.ShardTopology.Shard.PodTemplate, &ddbOriginSpec.ShardTopology.Shard.PodTemplate)
		fn(&mongodb.Spec.ShardTopology.ConfigServer.PodTemplate, &ddbOriginSpec.ShardTopology.ConfigServer.PodTemplate)
		fn(&mongodb.Spec.ShardTopology.Mongos.PodTemplate, &ddbOriginSpec.ShardTopology.Mongos.PodTemplate)
	} else if ddbOriginSpec.PodTemplate != nil {
		if mongodb.Spec.PodTemplate == nil {
			mongodb.Spec.PodTemplate = new(ofst.PodTemplateSpec)
		}
		fn(mongodb.Spec.PodTemplate, ddbOriginSpec.PodTemplate)
	}
}

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func setMonitoringPort(mongodb *api.MongoDB) {
	if mongodb.Spec.Monitor != nil &&
		mongodb.GetMonitoringVendor() == mona.VendorPrometheus {
		if mongodb.Spec.Monitor.Prometheus == nil {
			mongodb.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if mongodb.Spec.Monitor.Prometheus.Port == 0 {
			mongodb.Spec.Monitor.Prometheus.Port = api.PrometheusExporterPortNumber
		}
	}
}
