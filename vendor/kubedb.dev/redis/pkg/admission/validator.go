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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	version "gomodules.xyz/version"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type RedisValidator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &RedisValidator{}

var forbiddenEnvVars = []string{
	// No forbidden envs yet
}

func (a *RedisValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralRedis,
		},
		api.ResourceSingularRedis
}

func (a *RedisValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *RedisValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindRedis {
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
			obj, err := a.extClient.KubedbV1alpha1().Redises(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
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
			redis := obj.(*api.Redis).DeepCopy()
			oldRedis := oldObject.(*api.Redis).DeepCopy()
			oldRedis.SetDefaults(a.ClusterTopology)
			if err := validateUpdate(redis, oldRedis); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateRedis(a.client, a.extClient, obj.(*api.Redis), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// ValidateRedis checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateRedis(client kubernetes.Interface, extClient cs.Interface, redis *api.Redis, strictValidation bool) error {
	if redis.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	if _, err := extClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(redis.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if redis.Spec.Mode != api.RedisModeStandalone && redis.Spec.Mode != api.RedisModeCluster {
		return fmt.Errorf(`spec.mode "%v" invalid. Value must be one of "%v" or "%v"`,
			redis.Spec.Mode, api.RedisModeStandalone, api.RedisModeCluster)
	}

	if redis.Spec.Mode == api.RedisModeStandalone && *redis.Spec.Replicas != 1 {
		return fmt.Errorf(`spec.replicas "%v" invalid for standalone mode. Value must be one`, redis.Spec.Replicas)
	}

	if redis.Spec.Mode == api.RedisModeCluster && *redis.Spec.Cluster.Master < 3 {
		return fmt.Errorf(`spec.cluster.master "%v" invalid. Value must be >= 3`, redis.Spec.Cluster.Master)
	}

	if redis.Spec.Mode == api.RedisModeCluster && *redis.Spec.Cluster.Replicas == 0 {
		return fmt.Errorf(`spec.cluster.replicas "%v" invalid. Value must be > 0`, redis.Spec.Cluster.Replicas)
	}

	if redis.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if redis.Spec.StorageType != api.StorageTypeDurable && redis.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, redis.Spec.StorageType)
	}
	if err := amv.ValidateStorage(client, redis.Spec.StorageType, redis.Spec.Storage); err != nil {
		return err
	}

	if strictValidation {
		// Check if redisVersion is deprecated.
		// If deprecated, return error
		redisVersion, err := extClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(redis.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}
		if redisVersion.Spec.Deprecated {
			return fmt.Errorf("redis %s/%s is using deprecated version %v. Skipped processing",
				redis.Namespace, redis.Name, redisVersion.Name)
		}

		if err := redisVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("redis %s/%s is using invalid redisVersion %v. Skipped processing. reason: %v", redis.Namespace,
				redis.Name, redisVersion.Name, err)
		}
	}

	if redis.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if redis.Spec.StorageType == api.StorageTypeEphemeral && redis.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		return fmt.Errorf(`'spec.terminationPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	if redis.Spec.TLS != nil {
		redisVersion, err := extClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(redis.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}
		if _, err := checkTLSSupport(redisVersion.Spec.Version); err != nil {
			return err
		}
	}

	if err := amv.ValidateEnvVar(redis.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindRedis); err != nil {
		return err
	}

	monitorSpec := redis.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
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
	"spec.storageType",
	"spec.storage",
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

func checkTLSSupport(v string) (bool, error) {
	rdVersion, err := version.NewVersion(v)
	if err != nil {
		return false, err
	}
	if rdVersion.Major() < 6 {
		return false, fmt.Errorf("ssl support is available only for v6 or later versions")
	}
	return true, nil
}
