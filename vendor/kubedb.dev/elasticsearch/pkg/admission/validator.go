/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package admission

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	"gomodules.xyz/sets"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type ElasticsearchValidator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &ElasticsearchValidator{}

var forbiddenEnvVars = []string{
	"node.name",
	"node.ingest",
	"node.master",
	"node.data",
}

func (a *ElasticsearchValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralElasticsearch,
		},
		api.ResourceSingularElasticsearch
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
			obj, err := a.extClient.KubedbV1alpha2().Elasticsearches(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					break
				}
				return hookapi.StatusInternalServerError(err)
			} else if obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`elasticsearch "%v/%v" can't be terminated. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
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
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldElasticsearch.Spec.AuthSecret == nil {
				oldElasticsearch.Spec.AuthSecret = elasticsearch.Spec.AuthSecret
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
func ValidateElasticsearch(client kubernetes.Interface, extClient cs.Interface, db *api.Elasticsearch, strictValidation bool) error {
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	if _, err := extClient.CatalogV1alpha1().ElasticsearchVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if db.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if db.Spec.StorageType != api.StorageTypeDurable && db.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, db.Spec.StorageType)
	}

	topology := db.Spec.Topology
	if topology != nil {
		if db.Spec.Replicas != nil {
			return errors.New("doesn't support spec.replicas when spec.topology is set")
		}
		if db.Spec.Storage != nil {
			return errors.New("doesn't support spec.storage when spec.topology is set")
		}
		if db.Spec.PodTemplate.Spec.Resources.Size() != 0 {
			return errors.New("doesn't support spec.resources when spec.topology is set")
		}

		if topology.Ingest.Suffix == topology.Master.Suffix {
			return errors.New("ingest & master node should not have same suffix")
		}
		if topology.Ingest.Suffix == topology.Data.Suffix {
			return errors.New("ingest & data node should not have same suffix")
		}
		if topology.Master.Suffix == topology.Data.Suffix {
			return errors.New("master & data node should not have same suffix")
		}

		if topology.Ingest.Replicas == nil || *topology.Ingest.Replicas < 1 {
			return fmt.Errorf(`topology.ingest.replicas "%v" invalid. Must be greater than zero`, topology.Ingest.Replicas)
		}
		if err := amv.ValidateStorage(client, db.Spec.StorageType, topology.Ingest.Storage); err != nil {
			return err
		}

		if topology.Master.Replicas == nil || *topology.Master.Replicas < 1 {
			return fmt.Errorf(`topology.master.replicas "%v" invalid. Must be greater than zero`, topology.Master.Replicas)
		}
		if err := amv.ValidateStorage(client, db.Spec.StorageType, topology.Master.Storage); err != nil {
			return err
		}

		if topology.Data.Replicas == nil || *topology.Data.Replicas < 1 {
			return fmt.Errorf(`topology.data.replicas "%v" invalid. Must be greater than zero`, topology.Data.Replicas)
		}
		if err := amv.ValidateStorage(client, db.Spec.StorageType, topology.Data.Storage); err != nil {
			return err
		}

		// Resources validation
		// Heap size is the 50% of memory & it cannot be less than 128Mi(some say 97Mi)
		// So, minimum memory request should be twice of 128Mi, i.e. 256Mi.
		if value, ok := topology.Master.Resources.Requests[core.ResourceMemory]; ok && value.Value() < 2*api.ElasticsearchMinHeapSize {
			return fmt.Errorf("master.resources.reqeusts.memory cannot be less than %dMi, given %dMi", (2*api.ElasticsearchMinHeapSize)/(1024*1024), value.Value()/(1024*1024))
		}
		if value, ok := topology.Data.Resources.Requests[core.ResourceMemory]; ok && value.Value() < 2*api.ElasticsearchMinHeapSize {
			return fmt.Errorf("data.resources.reqeusts.memory cannot be less than %dMi, given %dMi", (2*api.ElasticsearchMinHeapSize)/(1024*1024), value.Value()/(1024*1024))
		}
		if value, ok := topology.Ingest.Resources.Requests[core.ResourceMemory]; ok && value.Value() < 2*api.ElasticsearchMinHeapSize {
			return fmt.Errorf("ingest.resources.reqeusts.memory cannot be less than %dMi, given %dMi", (2*api.ElasticsearchMinHeapSize)/(1024*1024), value.Value()/(1024*1024))
		}
	} else {
		if db.Spec.Replicas == nil || *db.Spec.Replicas < 1 {
			return fmt.Errorf(`spec.replicas "%v" invalid. Must be greater than zero`, db.Spec.Replicas)
		}

		if err := amv.ValidateStorage(client, db.Spec.StorageType, db.Spec.Storage); err != nil {
			return err
		}

		// Resources validation
		// Heap size is the 50% of memory & it cannot be less than 128Mi(some say 97Mi)
		// So, minimum memory request should be twice of 128Mi, i.e. 256Mi.
		if value, ok := db.Spec.PodTemplate.Spec.Resources.Requests[core.ResourceMemory]; ok && value.Value() < 2*api.ElasticsearchMinHeapSize {
			return fmt.Errorf("PodTemplate.Spec.Resources.Requests.memory cannot be less than %dMi, given %dMi", (2*api.ElasticsearchMinHeapSize)/(1024*1024), value.Value()/(1024*1024))
		}
	}

	if err := amv.ValidateEnvVar(db.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindElasticsearch); err != nil {
		return err
	}

	if strictValidation {
		authSecret := db.Spec.AuthSecret
		if authSecret != nil {
			if _, err := client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), authSecret.Name, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if elasticsearchVersion is deprecated.
		// If deprecated, return error
		elasticsearchVersion, err := extClient.CatalogV1alpha1().ElasticsearchVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}

		if elasticsearchVersion.Spec.Deprecated {
			return fmt.Errorf("elasticsearch %s/%s is using deprecated version %v. Skipped processing", db.Namespace,
				db.Name, elasticsearchVersion.Name)
		}

		if err := elasticsearchVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("elasticsearch %s/%s is using invalid elasticsearchVersion %v. Skipped processing. reason: %v", db.Namespace,
				db.Name, elasticsearchVersion.Name, err)
		}
	}

	if db.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if db.Spec.StorageType == api.StorageTypeEphemeral && db.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		return fmt.Errorf(`'spec.terminationPolicy: Halt' can not be set for 'Ephemeral' storage`)
	}

	if db.Spec.DisableSecurity && db.Spec.EnableSSL {
		return fmt.Errorf(`to enable 'spec.enableSSL', 'spec.disableSecurity' needs to be set to false`)
	}

	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdate(obj, oldObj *api.Elasticsearch) error {
	preconditions := getPreconditionFunc(oldObj)
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError())
		}
		return err
	}
	return nil
}

func getPreconditionFunc(db *api.Elasticsearch) []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
	}

	// Once the database has been initialized, don't let update the "spec.init" section
	if db.Spec.Init != nil && db.Spec.Init.Initialized {
		preconditionSpecFields.Insert("spec.init")
	}

	for _, field := range preconditionSpecFields.List() {
		preconditions = append(preconditions,
			meta_util.RequireChainKeyUnchanged(field),
		)
	}
	return preconditions
}

var preconditionSpecFields = sets.NewString(
	"spec.topology.*.prefix",
	"spec.authSecret",
	"spec.storageType",
	"spec.podTemplate.spec.nodeSelector",
)

func preconditionFailedError() error {
	str := preconditionSpecFields.List()
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}
