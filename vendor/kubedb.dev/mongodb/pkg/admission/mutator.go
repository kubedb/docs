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
	"sync"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1"
	"kmodules.xyz/webhook-runtime/builder"
)

type MongoDBMutator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	dbClient    cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &MongoDBMutator{}

func (a *MongoDBMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return builder.MutatorResource(api.Kind(api.ResourceKindMongoDB))
}

func (a *MongoDBMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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
	mongoMod, err := a.setDefaultValues(a.dbClient, obj.(*api.MongoDB).DeepCopy())
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
func (a *MongoDBMutator) setDefaultValues(extClient cs.Interface, db *api.MongoDB) (runtime.Object, error) {
	if db.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Halted {
		if db.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
			return nil, errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.TerminationPolicy = api.TerminationPolicyHalt
	}

	mgVersion, err := getMongoDBVersion(extClient, db.Spec.Version)
	if err != nil {
		return nil, err
	}

	db.SetDefaults(mgVersion, a.ClusterTopology)

	return db, nil
}

// getMongoDBVersion returns MongoDBVersion.
// If MongoDBVersion doesn't exists return 0 valued MongoDBVersion (not nil)
func getMongoDBVersion(extClient cs.Interface, ver string) (*v1alpha1.MongoDBVersion, error) {
	mgVersion, err := extClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), ver, metav1.GetOptions{})
	if err != nil {
		return &v1alpha1.MongoDBVersion{}, err
	}
	return mgVersion, err
}
