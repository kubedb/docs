package admission

import (
	"fmt"
	"strings"
	"sync"

	"github.com/appscode/go/log"
	hookapi "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	kubedbv1alpha1 "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	amv "github.com/kubedb/apimachinery/pkg/validator"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ElasticsearchValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &ElasticsearchValidator{}

func (a *ElasticsearchValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "admission.kubedb.com",
			Version:  "v1alpha1",
			Resource: "elasticsearchvalidationreviews",
		},
		"elasticsearchvalidationreview"
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
		// req.Object.Raw = nil, so read from kubernetes
		obj, err := a.extClient.KubedbV1alpha1().Elasticsearches(req.Namespace).Get(req.Name, metav1.GetOptions{})
		if err != nil && !kerr.IsNotFound(err) {
			return hookapi.StatusInternalServerError(err)
		} else if err == nil && obj.Spec.DoNotPause {
			return hookapi.StatusBadRequest(fmt.Errorf(`elasticsearch "%s" can't be paused. To continue delete, unset spec.doNotPause and retry`, req.Name))
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
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldElasticsearch.Spec.DatabaseSecret == nil {
				oldElasticsearch.Spec.DatabaseSecret = elasticsearch.Spec.DatabaseSecret
			}

			// Allow changing CertificateSecret only if there was no secret have set up yet.
			if oldElasticsearch.Spec.CertificateSecret == nil {
				oldElasticsearch.Spec.CertificateSecret = elasticsearch.Spec.CertificateSecret
			}

			if err := validateUpdate(elasticsearch, oldElasticsearch, req.Kind.Kind); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateElasticsearch(a.client, a.extClient.KubedbV1alpha1(), obj.(*api.Elasticsearch)); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

var (
	elasticVersions = sets.NewString("5.6", "5.6.4")
)

// ValidateElasticsearch checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateElasticsearch(client kubernetes.Interface, extClient kubedbv1alpha1.KubedbV1alpha1Interface, elasticsearch *api.Elasticsearch) error {
	if elasticsearch.Spec.Version == "" {
		return fmt.Errorf(`object 'Version' is missing in '%v'`, elasticsearch.Spec)
	}

	// check Elasticsearch version validation
	if !elasticVersions.Has(string(elasticsearch.Spec.Version)) {
		return fmt.Errorf(`KubeDB doesn't support Elasticsearch version: %s`, string(elasticsearch.Spec.Version))
	}

	topology := elasticsearch.Spec.Topology
	if topology != nil {
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
			return fmt.Errorf(`topology.client.replicas "%d" invalid. Must be greater than zero`, topology.Client.Replicas)
		}

		if topology.Master.Replicas == nil || *topology.Master.Replicas < 1 {
			return fmt.Errorf(`topology.master.replicas "%d" invalid. Must be greater than zero`, topology.Master.Replicas)
		}

		if topology.Data.Replicas == nil || *topology.Data.Replicas < 1 {
			return fmt.Errorf(`topology.data.replicas "%d" invalid. Must be greater than zero`, topology.Data.Replicas)
		}
	} else {
		if elasticsearch.Spec.Replicas == nil || *elasticsearch.Spec.Replicas < 1 {
			return fmt.Errorf(`spec.replicas "%d" invalid. Must be greater than zero`, elasticsearch.Spec.Replicas)
		}
	}

	if elasticsearch.Spec.Storage != nil {
		var err error
		if err = amv.ValidateStorage(client, elasticsearch.Spec.Storage); err != nil {
			return err
		}
	}

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

	backupScheduleSpec := elasticsearch.Spec.BackupSchedule
	if backupScheduleSpec != nil {
		if err := amv.ValidateBackupSchedule(client, backupScheduleSpec, elasticsearch.Namespace); err != nil {
			return err
		}
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

func matchWithDormantDatabase(extClient kubedbv1alpha1.KubedbV1alpha1Interface, elasticsearch *api.Elasticsearch) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := extClient.DormantDatabases(elasticsearch.Namespace).Get(elasticsearch.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}

	// Check DatabaseKind
	if value, _ := meta_util.GetStringValue(dormantDb.Labels, api.LabelDatabaseKind); value != api.ResourceKindElasticsearch {
		return errors.New(fmt.Sprintf(`invalid Elasticsearch: "%v". Exists DormantDatabase "%v" of different Kind`, elasticsearch.Name, dormantDb.Name))
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.Elasticsearch
	originalSpec := elasticsearch.Spec

	// Skip checking doNotPause
	drmnOriginSpec.DoNotPause = originalSpec.DoNotPause

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

func validateUpdate(obj, oldObj runtime.Object, kind string) error {
	preconditions := getPreconditionFunc()
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError(kind))
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
	"spec.version",
	"spec.topology.*.prefix",
	"spec.enableSSL",
	"spec.certificateSecret",
	"spec.databaseSecret",
	"spec.storage",
	"spec.nodeSelector",
	"spec.init",
}

func preconditionFailedError(kind string) error {
	str := preconditionSpecFields
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}
