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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/appscode/go/types"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type ElasticsearchMutator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &ElasticsearchMutator{}

func (a *ElasticsearchMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.MutatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralElasticsearch,
		},
		api.ResourceSingularElasticsearch
}

func (a *ElasticsearchMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *ElasticsearchMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
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
	obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
	if err != nil {
		return hookapi.StatusBadRequest(err)
	}
	mod, err := setDefaultValues(obj.(*api.Elasticsearch).DeepCopy(), a.ClusterTopology)
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

// setDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a Elasticsearch database
func setDefaultValues(elasticsearch *api.Elasticsearch, clusterTopology *core_util.Topology) (runtime.Object, error) {
	if elasticsearch.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	if elasticsearch.Spec.Halted {
		if elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
			return nil, errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyHalt
	}

	topology := elasticsearch.Spec.Topology
	if topology != nil {
		if topology.Client.Replicas == nil {
			topology.Client.Replicas = types.Int32P(1)
		}

		if topology.Master.Replicas == nil {
			topology.Master.Replicas = types.Int32P(1)
		}

		if topology.Data.Replicas == nil {
			topology.Data.Replicas = types.Int32P(1)
		}
	} else {
		if elasticsearch.Spec.Replicas == nil {
			elasticsearch.Spec.Replicas = types.Int32P(1)
		}
	}
	elasticsearch.SetDefaults(clusterTopology)

	// If monitoring spec is given without port,
	// set default Listening port
	setMonitoringPort(elasticsearch)

	return elasticsearch, nil
}

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func setMonitoringPort(db *api.Elasticsearch) {
	if db.Spec.Monitor != nil &&
		db.GetMonitoringVendor() == mona.VendorPrometheus {
		if db.Spec.Monitor.Prometheus == nil {
			db.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if db.Spec.Monitor.Prometheus.Exporter == nil {
			db.Spec.Monitor.Prometheus.Exporter = &mona.PrometheusExporterSpec{}
		}
		if db.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			db.Spec.Monitor.Prometheus.Exporter.Port = api.PrometheusExporterPortNumber
		}
	}
}
