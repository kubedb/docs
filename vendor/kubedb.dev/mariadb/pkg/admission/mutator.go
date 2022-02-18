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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1"
	"kmodules.xyz/webhook-runtime/builder"
)

// MariaDBMutator implements the AdmissionHook interface to mutate the MariaDB resources
type MariaDBMutator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	dbClient    cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &MariaDBMutator{}

// Resource is the resource to use for hosting mutating admission webhook.
func (a *MariaDBMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return builder.MutatorResource(api.Kind(api.ResourceKindMariaDB))
}

// Initialize is called as a post-start hook
func (a *MariaDBMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

// Admit is called to decide whether to accept the admission request.
// The returned response may use the Patch field to mutate the object.
func (a *MariaDBMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindMariaDB {
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
	mariadbMod, err := setDefaultValues(obj.(*api.MariaDB).DeepCopy(), a.ClusterTopology)
	if err != nil {
		return hookapi.StatusForbidden(err)
	} else if mariadbMod != nil {
		patch, err := meta_util.CreateJSONPatch(req.Object.Raw, mariadbMod)
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

// setDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a MySQL database
func setDefaultValues(db *api.MariaDB, topology *core_util.Topology) (runtime.Object, error) {
	if db.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
			return nil, errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.TerminationPolicy = api.TerminationPolicyHalt
	}

	db.SetDefaults(topology)

	return db, nil
}
