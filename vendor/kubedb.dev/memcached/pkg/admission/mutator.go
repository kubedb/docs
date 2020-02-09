/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package admission

import (
	"fmt"
	"sync"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/appscode/go/types"
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

type MemcachedMutator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &MemcachedMutator{}

func (a *MemcachedMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "mutators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "memcachedmutators",
		},
		"memcachedmutator"
}

func (a *MemcachedMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *MemcachedMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindMemcached {
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
	mod, err := setDefaultValues(obj.(*api.Memcached).DeepCopy())
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

// setDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a Memcached database
func setDefaultValues(memcached *api.Memcached) (runtime.Object, error) {
	if memcached.Spec.Version == "" {
		return nil, fmt.Errorf(`object 'Version' is missing in '%v'`, memcached.Spec)
	}

	if memcached.Spec.Halted {
		if memcached.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
			return nil, errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		memcached.Spec.TerminationPolicy = api.TerminationPolicyHalt
	}

	if memcached.Spec.Replicas == nil {
		memcached.Spec.Replicas = types.Int32P(1)
	}
	memcached.SetDefaults()

	// If monitoring spec is given without port,
	// set default Listening port
	setMonitoringPort(memcached)

	return memcached, nil
}

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func setMonitoringPort(memcached *api.Memcached) {
	if memcached.Spec.Monitor != nil &&
		memcached.GetMonitoringVendor() == mona.VendorPrometheus {
		if memcached.Spec.Monitor.Prometheus == nil {
			memcached.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if memcached.Spec.Monitor.Prometheus.Exporter == nil {
			memcached.Spec.Monitor.Prometheus.Exporter = &mona.PrometheusExporterSpec{}
		}
		if memcached.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			memcached.Spec.Monitor.Prometheus.Exporter.Port = api.PrometheusExporterPortNumber
		}
	}
}
