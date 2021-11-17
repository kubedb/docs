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

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
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

type RedisSentinelValidator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &RedisSentinelValidator{}

func (a *RedisSentinelValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralRedisSentinel,
		},
		api.ResourceSingularRedisSentinel
}
func (a *RedisSentinelValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *RedisSentinelValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindRedisSentinel {
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
			obj, err := a.extClient.KubedbV1alpha2().RedisSentinels(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
			} else if err == nil && obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`redis "%v/%v" can't be terminated. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
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
			sentinel := obj.(*api.RedisSentinel).DeepCopy()
			oldSentinel := oldObject.(*api.RedisSentinel).DeepCopy()
			oldSentinel.SetDefaults(a.ClusterTopology)
			if err := validateSentinelUpdate(sentinel, oldSentinel); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateRedisSentinel(a.client, a.extClient, obj.(*api.RedisSentinel), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// ValidateRedisSentinel checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateRedisSentinel(client kubernetes.Interface, extClient cs.Interface, sentinel *api.RedisSentinel, strictValidation bool) error {
	if sentinel.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	if _, err := extClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(sentinel.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if sentinel.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if sentinel.Spec.StorageType != api.StorageTypeDurable && sentinel.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, sentinel.Spec.StorageType)
	}
	if err := amv.ValidateStorage(client, sentinel.Spec.StorageType, sentinel.Spec.Storage); err != nil {
		return err
	}

	if strictValidation {
		// Check if redisVersion is deprecated.
		// If deprecated, return error
		redisVersion, err := extClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(sentinel.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}
		if redisVersion.Spec.Deprecated {
			return fmt.Errorf("redis Sentinel %s/%s is using deprecated version %v. Skipped processing",
				sentinel.Namespace, sentinel.Name, redisVersion.Name)
		}

		if err := redisVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("redis Sentinel %s/%s is using invalid redisVersion %v. Skipped processing. reason: %v", sentinel.Namespace,
				sentinel.Name, redisVersion.Name, err)
		}
	}

	if sentinel.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if sentinel.Spec.StorageType == api.StorageTypeEphemeral && sentinel.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		return fmt.Errorf(`'spec.terminationPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	if sentinel.Spec.TLS != nil {
		redisVersion, err := extClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(sentinel.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}
		if err := checkTLSSupport(redisVersion.Spec.Version); err != nil {
			return err
		}
	}

	if err := amv.ValidateEnvVar(sentinel.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindRedis); err != nil {
		return err
	}

	monitorSpec := sentinel.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateSentinelUpdate(obj, oldObj *api.RedisSentinel) error {
	preconditions := meta_util.PreConditionSet{
		String: sets.NewString(
			"spec.storageType",
			"spec.storage",
			"spec.podTemplate.spec.nodeSelector",
		),
	}

	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}
