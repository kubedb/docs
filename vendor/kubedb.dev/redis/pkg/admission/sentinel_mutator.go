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
	"sync"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type RedisSentinelMutator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	dbClient    cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &RedisSentinelMutator{}

func (a *RedisSentinelMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.MutatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralRedisSentinel,
		},
		api.ResourceSingularRedisSentinel
}

func (a *RedisSentinelMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	if a.client, err = kubernetes.NewForConfig(config); err != nil {
		return err
	}
	if a.dbClient, err = cs.NewForConfig(config); err != nil {
		return err
	}
	return err
}

func (a *RedisSentinelMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
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
	obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
	if err != nil {
		return hookapi.StatusBadRequest(err)
	}
	mod, err := setSentinelDefaultValues(obj.(*api.RedisSentinel).DeepCopy(), a.ClusterTopology)
	if err != nil {
		return hookapi.StatusForbidden(err)
	} else if mod != nil {
		patch, err := meta_util.CreateJSONPatch(req.Object.Raw, mod)
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

// setSentinelDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a Redis Sentinel database
func setSentinelDefaultValues(sentinel *api.RedisSentinel, clusterTopology *core_util.Topology) (runtime.Object, error) {
	if sentinel.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	if sentinel.Spec.Halted {
		if sentinel.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
			return nil, errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		sentinel.Spec.TerminationPolicy = api.TerminationPolicyHalt
	}

	if sentinel.Spec.Replicas == nil {
		sentinel.Spec.Replicas = pointer.Int32P(1)
	}

	sentinel.SetDefaults(clusterTopology)

	return sentinel, nil
}
