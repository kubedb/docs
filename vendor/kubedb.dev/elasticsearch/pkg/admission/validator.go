package admission

import (
	"fmt"
	"strings"
	"sync"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type ElasticsearchValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &ElasticsearchValidator{}

var forbiddenEnvVars = []string{
	"NODE_NAME",
	"NODE_MASTER",
	"NODE_DATA",
}

func (a *ElasticsearchValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "validators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "elasticsearchvalidators",
		},
		"elasticsearchvalidator"
}

func (a *ElasticsearchValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *ElasticsearchValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindElasticsearch {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return hookapi.StatusUninitialized()
	}

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			// req.Object.Raw = nil, so read from kubernetes
			obj, err := a.extClient.KubedbV1alpha1().Elasticsearches(req.Namespace).Get(req.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					break
				}
				return hookapi.StatusInternalServerError(err)
			} else if obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`elasticsearch "%v/%v" can't be paused. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
			}
		}
	default:
		obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}
		if req.Operation == admission.Update {
			// validate changes made by user
			oldObject, err := meta_util.UnmarshalFromJSON(req.OldObject.Raw, api.SchemeGroupVersion)
			if err != nil {
				return hookapi.StatusBadRequest(err)
			}

			elasticsearch := obj.(*api.Elasticsearch).DeepCopy()
			oldElasticsearch := oldObject.(*api.Elasticsearch).DeepCopy()
			oldElasticsearch.SetDefaults()
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldElasticsearch.Spec.DatabaseSecret == nil {
				oldElasticsearch.Spec.DatabaseSecret = elasticsearch.Spec.DatabaseSecret
			}

			// Allow changing CertificateSecret only if there was no secret have set up yet.
			if oldElasticsearch.Spec.CertificateSecret == nil {
				oldElasticsearch.Spec.CertificateSecret = elasticsearch.Spec.CertificateSecret
			}

			if err := validateUpdate(elasticsearch, oldElasticsearch); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateElasticsearch(a.client, a.extClient, obj.(*api.Elasticsearch), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// ValidateElasticsearch checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateElasticsearch(client kubernetes.Interface, extClient cs.Interface, elasticsearch *api.Elasticsearch, strictValidation bool) error {
	if elasticsearch.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	if _, err := extClient.CatalogV1alpha1().ElasticsearchVersions().Get(string(elasticsearch.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if elasticsearch.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if elasticsearch.Spec.StorageType != api.StorageTypeDurable && elasticsearch.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, elasticsearch.Spec.StorageType)
	}

	topology := elasticsearch.Spec.Topology
	if topology != nil {
		if elasticsearch.Spec.Replicas != nil {
			return errors.New("doesn't support spec.replicas when spec.topology is set")
		}
		if elasticsearch.Spec.Storage != nil {
			return errors.New("doesn't support spec.storage when spec.topology is set")
		}
		if elasticsearch.Spec.PodTemplate.Spec.Resources.Size() != 0 {
			return errors.New("doesn't support spec.resources when spec.topology is set")
		}

		if topology.Client.Prefix == topology.Master.Prefix {
			return errors.New("client & master node should not have same prefix")
		}
		if topology.Client.Prefix == topology.Data.Prefix {
			return errors.New("client & data node should not have same prefix")
		}
		if topology.Master.Prefix == topology.Data.Prefix {
			return errors.New("master & data node should not have same prefix")
		}

		if topology.Client.Replicas == nil || *topology.Client.Replicas < 1 {
			return fmt.Errorf(`topology.client.replicas "%v" invalid. Must be greater than zero`, topology.Client.Replicas)
		}
		if err := amv.ValidateStorage(client, elasticsearch.Spec.StorageType, topology.Client.Storage); err != nil {
			return err
		}

		if topology.Master.Replicas == nil || *topology.Master.Replicas < 1 {
			return fmt.Errorf(`topology.master.replicas "%v" invalid. Must be greater than zero`, topology.Master.Replicas)
		}
		if err := amv.ValidateStorage(client, elasticsearch.Spec.StorageType, topology.Master.Storage); err != nil {
			return err
		}

		if topology.Data.Replicas == nil || *topology.Data.Replicas < 1 {
			return fmt.Errorf(`topology.data.replicas "%v" invalid. Must be greater than zero`, topology.Data.Replicas)
		}
		if err := amv.ValidateStorage(client, elasticsearch.Spec.StorageType, topology.Data.Storage); err != nil {
			return err
		}
	} else {
		if elasticsearch.Spec.Replicas == nil || *elasticsearch.Spec.Replicas < 1 {
			return fmt.Errorf(`spec.replicas "%v" invalid. Must be greater than zero`, elasticsearch.Spec.Replicas)
		}

		if err := amv.ValidateStorage(client, elasticsearch.Spec.StorageType, elasticsearch.Spec.Storage); err != nil {
			return err
		}
	}

	if err := amv.ValidateEnvVar(elasticsearch.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindElasticsearch); err != nil {
		return err
	}

	if strictValidation {
		databaseSecret := elasticsearch.Spec.DatabaseSecret
		if databaseSecret != nil {
			if _, err := client.CoreV1().Secrets(elasticsearch.Namespace).Get(databaseSecret.SecretName, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		certificateSecret := elasticsearch.Spec.CertificateSecret
		if certificateSecret != nil {
			if _, err := client.CoreV1().Secrets(elasticsearch.Namespace).Get(certificateSecret.SecretName, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if elasticsearchVersion is deprecated.
		// If deprecated, return error
		elasticsearchVersion, err := extClient.CatalogV1alpha1().ElasticsearchVersions().Get(string(elasticsearch.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}

		if elasticsearchVersion.Spec.Deprecated {
			return fmt.Errorf("elasticsearch %s/%s is using deprecated version %v. Skipped processing", elasticsearch.Namespace,
				elasticsearch.Name, elasticsearchVersion.Name)
		}

		if err := elasticsearchVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("elasticsearch %s/%s is using invalid elasticsearchVersion %v. Skipped processing. reason: %v", elasticsearch.Namespace,
				elasticsearch.Name, elasticsearchVersion.Name, err)
		}
	}

	backupScheduleSpec := elasticsearch.Spec.BackupSchedule
	if backupScheduleSpec != nil {
		if err := amv.ValidateBackupSchedule(client, backupScheduleSpec, elasticsearch.Namespace); err != nil {
			return err
		}
	}

	if elasticsearch.Spec.UpdateStrategy.Type == "" {
		return fmt.Errorf(`'spec.updateStrategy.type' is missing`)
	}

	if elasticsearch.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if elasticsearch.Spec.StorageType == api.StorageTypeEphemeral && elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyPause {
		return fmt.Errorf(`'spec.terminationPolicy: Pause' can not be used for 'Ephemeral' storage`)
	}

	if elasticsearch.Spec.DisableSecurity && elasticsearch.Spec.EnableSSL {
		return fmt.Errorf(`to enable 'spec.enableSSL', 'spec.disableSecurity' needs to be set to false`)
	}

	monitorSpec := elasticsearch.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	if err := matchWithDormantDatabase(extClient, elasticsearch); err != nil {
		return err
	}

	return nil
}

func matchWithDormantDatabase(extClient cs.Interface, elasticsearch *api.Elasticsearch) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := extClient.KubedbV1alpha1().DormantDatabases(elasticsearch.Namespace).Get(elasticsearch.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}

	// Check DatabaseKind
	if value, _ := meta_util.GetStringValue(dormantDb.Labels, api.LabelDatabaseKind); value != api.ResourceKindElasticsearch {
		return errors.New(fmt.Sprintf(`invalid Elasticsearch: "%v/%v". Exists DormantDatabase "%v/%v" of different Kind`, elasticsearch.Namespace, elasticsearch.Name, dormantDb.Namespace, dormantDb.Name))
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.Elasticsearch
	drmnOriginSpec.SetDefaults()
	originalSpec := elasticsearch.Spec

	// Skip checking UpdateStrategy
	drmnOriginSpec.UpdateStrategy = originalSpec.UpdateStrategy

	// Skip checking ServiceAccountName
	drmnOriginSpec.PodTemplate.Spec.ServiceAccountName = elasticsearch.Spec.PodTemplate.Spec.ServiceAccountName

	// Skip checking TerminationPolicy
	drmnOriginSpec.TerminationPolicy = originalSpec.TerminationPolicy

	// Skip checking Monitoring
	drmnOriginSpec.Monitor = originalSpec.Monitor

	// Skip Checking BackUP Scheduler
	drmnOriginSpec.BackupSchedule = originalSpec.BackupSchedule

	if !meta_util.Equal(drmnOriginSpec, &originalSpec) {
		diff := meta_util.Diff(drmnOriginSpec, &originalSpec)
		log.Errorf("elasticsearch spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff)
		return errors.New(fmt.Sprintf("elasticsearch spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff))
	}

	return nil
}

func validateUpdate(obj, oldObj runtime.Object) error {
	preconditions := getPreconditionFunc()
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError())
		}
		return err
	}
	return nil
}

func getPreconditionFunc() []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
	}

	for _, field := range preconditionSpecFields {
		preconditions = append(preconditions,
			meta_util.RequireChainKeyUnchanged(field),
		)
	}
	return preconditions
}

var preconditionSpecFields = []string{
	"spec.topology.*.prefix",
	"spec.topology.*.storage",
	"spec.enableSSL",
	"spec.certificateSecret",
	"spec.authPlugin",
	"spec.databaseSecret",
	"spec.storageType",
	"spec.storage",
	"spec.init",
	"spec.podTemplate.spec.nodeSelector",
}

func preconditionFailedError() error {
	str := preconditionSpecFields
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}
