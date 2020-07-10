/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

// PerconaXtraDBMutator implements the AdmissionHook interface to mutate the PerconaXtraDB resources
type PerconaXtraDBMutator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &PerconaXtraDBMutator{}

// Resource is the resource to use for hosting mutating admission webhook.
func (a *PerconaXtraDBMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.MutatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralPerconaXtraDB,
		},
		api.ResourceSingularPerconaXtraDB
}

// Initialize is called as a post-start hook
func (a *PerconaXtraDBMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

// Admit is called to decide whether to accept the admission request.
// The returned response may use the Patch field to mutate the object.
func (a *PerconaXtraDBMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindPerconaXtraDB {
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
	perconaxtradbMod, err := setDefaultValues(obj.(*api.PerconaXtraDB).DeepCopy())
	if err != nil {
		return hookapi.StatusForbidden(err)
	} else if perconaxtradbMod != nil {
		patch, err := meta_util.CreateJSONPatch(req.Object.Raw, perconaxtradbMod)
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
func setDefaultValues(px *api.PerconaXtraDB) (runtime.Object, error) {
	if px.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	if px.Spec.Halted {
		if px.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
			return nil, errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		px.Spec.TerminationPolicy = api.TerminationPolicyHalt
	}

	px.SetDefaults()

	// If monitoring spec is given without port,
	// set default Listening port
	setMonitoringPort(px)

	return px, nil
}

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func setMonitoringPort(px *api.PerconaXtraDB) {
	if px.Spec.Monitor != nil &&
		px.GetMonitoringVendor() == mona.VendorPrometheus {
		if px.Spec.Monitor.Prometheus == nil {
			px.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if px.Spec.Monitor.Prometheus.Exporter == nil {
			px.Spec.Monitor.Prometheus.Exporter = &mona.PrometheusExporterSpec{}
		}
		if px.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			px.Spec.Monitor.Prometheus.Exporter.Port = api.PrometheusExporterPortNumber
		}
	}
}
