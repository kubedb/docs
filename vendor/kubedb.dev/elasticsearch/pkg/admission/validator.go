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
	"sync"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
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
	"node.ml",
	"node.data_hot",
	"node.data_warm",
	"node.data_cold",
	"node.data_frozen",
	"node.data_content",
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
			// Skip validation, if UPDATE operation is called after deletion.
			// Case: Removing Finalizer
			if elasticsearch.DeletionTimestamp != nil {
				break
			}
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
	esVersion, err := extClient.CatalogV1alpha1().ElasticsearchVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}

	if db.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if db.Spec.StorageType != api.StorageTypeDurable && db.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, db.Spec.StorageType)
	}

	err = validateSecureConfig(db, esVersion)
	if err != nil {
		return err
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
		err = validateNodeRoles(topology, esVersion)
		if err != nil {
			return err
		}
		// Check node name suffix
		err = validateNodeSuffix(topology)
		if err != nil {
			return err
		}
		err = validateNodeReplicas(topology)
		if err != nil {
			return err
		}
		tMap := topology.ToMap()
		for nodeRole, node := range tMap {
			if err := amv.ValidateStorage(client, db.Spec.StorageType, node.Storage); err != nil {
				return err
			}
			// Resources validation
			// Heap size is the 50% of memory & it cannot be less than 128Mi(some say 97Mi)
			// So, minimum memory request should be twice of 128Mi, i.e. 256Mi.
			if value, ok := node.Resources.Requests[core.ResourceMemory]; ok && value.Value() < 2*api.ElasticsearchMinHeapSize {
				return fmt.Errorf("%s.resources.reqeusts.memory cannot be less than %dMi, given %dMi", string(nodeRole), (2*api.ElasticsearchMinHeapSize)/(1024*1024), value.Value()/(1024*1024))
			}
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

	// TODO:
	//		- OpenSearch provision fails with security plugin disabled.
	//		- Remove the validation, once the issue is fixed.
	//		- Issue Ref: https://github.com/opensearch-project/security/issues/1481
	if db.Spec.DisableSecurity && esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginOpenSearch {
		return fmt.Errorf(`'spec.disableSecurity' cannot be 'true' for opensearch`)
	}

	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	if err = validateContainerSecurityContext(db.Spec.PodTemplate.Spec.ContainerSecurityContext, esVersion); err != nil {
		return err
	}

	return nil
}

func validateUpdate(obj, oldObj *api.Elasticsearch) error {
	preconditions := meta_util.PreConditionSet{
		String: sets.NewString(
			"spec.topology.*.suffix",
			"spec.authSecret",
			"spec.storageType",
			"spec.podTemplate.spec.nodeSelector",
		),
	}
	// Once the database has been initialized, don't let update the "spec.init" section
	if oldObj.Spec.Init != nil && oldObj.Spec.Init.Initialized {
		preconditions.Insert("spec.init")
	}
	_, err := meta_util.CreateJSONMergePatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func validateContainerSecurityContext(sc *core.SecurityContext, esVersion *catalog.ElasticsearchVersion) error {
	if sc == nil {
		return nil
	}

	// if RunAsAnyNonRoot == false
	//		only allow default UID (runAsUser)
	// else
	//		allow any UID but root (0)
	if !esVersion.Spec.SecurityContext.RunAsAnyNonRoot {
		// if default RunAsUser is missing, user isn't allowed to user RunAsUser.
		if esVersion.Spec.SecurityContext.RunAsUser == nil && sc.RunAsUser != nil {
			return fmt.Errorf("not allowed to set containerSecurityContext.runAsUser for ElasticsearchVersion: %s", esVersion.Name)
		}
		// if default RunAsUser is set, validate it.
		if sc.RunAsUser != nil && esVersion.Spec.SecurityContext.RunAsUser != nil &&
			*sc.RunAsUser != *esVersion.Spec.SecurityContext.RunAsUser {
			return fmt.Errorf("containerSecurityContext.runAsUser must be %d for ElasticsearchVersion: %s", *esVersion.Spec.SecurityContext.RunAsUser, esVersion.Name)
		}
	} else {
		if sc.RunAsUser != nil && *sc.RunAsUser == 0 {
			return fmt.Errorf("not allowed to set containerSecurityContext.runAsUser to root (0) for ElasticsearchVersion: %s", esVersion.Name)
		}
	}

	return nil
}

func validateNodeSuffix(topology *api.ElasticsearchClusterTopology) error {
	tMap := topology.ToMap()
	names := make(map[string]bool)
	for _, value := range tMap {
		names[value.Suffix] = true
	}
	if len(tMap) != len(names) {
		return errors.New("two or more node cannot have same suffix")
	}
	return nil
}

func validateNodeReplicas(topology *api.ElasticsearchClusterTopology) error {
	tMap := topology.ToMap()
	for key, node := range tMap {
		if pointer.Int32(node.Replicas) <= 0 {
			return errors.Errorf("replicas for node role %s must be alteast 1", string(key))
		}
	}
	return nil
}

func validateNodeRoles(topology *api.ElasticsearchClusterTopology, esVersion *catalog.ElasticsearchVersion) error {
	if esVersion.Spec.Distribution == catalog.ElasticsearchDistroOpenDistro || esVersion.Spec.Distribution == catalog.ElasticsearchDistroOpenSearch {
		if topology.ML != nil || topology.DataContent != nil || topology.DataCold != nil || topology.DataFrozen != nil ||
			topology.Coordinating != nil || topology.Transform != nil {
			return errors.Errorf("node role: ml, data_cold, data_frozen, data_content, transform, coordinating are not supported for ElasticsearchVersion %s", esVersion.Name)
		}
	} else if esVersion.Spec.Distribution == catalog.ElasticsearchDistroSearchGuard {
		if topology.Data == nil {
			return errors.New("topology.data cannot be empty")
		}
		if topology.ML != nil || topology.DataHot != nil || topology.DataContent != nil ||
			topology.DataCold != nil || topology.DataWarm != nil || topology.DataFrozen != nil ||
			topology.Coordinating != nil || topology.Transform != nil {
			return errors.Errorf("node role: ml, data_hot, data_cold, data_warm, data_frozen, data_content, transform, coordinating are not supported for ElasticsearchVersion %s", esVersion.Name)
		}
	}

	// Every cluster requires the following node roles:
	//	- (data_content and data_hot) OR (data)
	//	- ref: https://www.elastic.co/guide/en/elasticsearch/reference/7.14/modules-node.html#node-roles
	if topology.Data == nil && (topology.DataHot == nil || topology.DataContent == nil) {
		return errors.New("when data node is empty, you need to have both dataHot and dataContent nodes")
	}

	return nil
}

func validateSecureConfig(db *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion) error {
	dbVersion, err := semver.NewVersion(esVersion.Spec.Version)
	if err != nil {
		return err
	}
	if db.Spec.SecureConfigSecret != nil {
		// Elasticsearch keystore is not supported for OpenDistro
		if esVersion.Spec.Distribution == catalog.ElasticsearchDistroOpenDistro || esVersion.Spec.Distribution == catalog.ElasticsearchDistroOpenSearch {
			return errors.New("secureConfigSecret is not supported for Opendistro/OpenSearch of Elasticsearch")
		}
		// KEYSTORE_PASSWORD is supported since ES version 7.9
		if dbVersion.Major() < 7 || (dbVersion.Major() == 7 && dbVersion.Minor() < 9) {
			return errors.Errorf("secureConfigSecret is not supported for ElasticsearchVersion %s, try with latest versions", esVersion.Name)
		}
	}
	return nil
}
