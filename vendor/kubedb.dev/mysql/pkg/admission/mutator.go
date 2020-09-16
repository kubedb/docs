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
	"github.com/google/uuid"
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

type MySQLMutator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &MySQLMutator{}

func (a *MySQLMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.MutatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralMySQL,
		},
		api.ResourceSingularMySQL
}

func (a *MySQLMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *MySQLMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindMySQL {
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
	mysqlMod, err := setDefaultValues(obj.(*api.MySQL).DeepCopy())
	if err != nil {
		return hookapi.StatusForbidden(err)
	} else if mysqlMod != nil {
		patch, err := meta_util.CreateJSONPatch(req.Object.Raw, mysqlMod)
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
func setDefaultValues(mysql *api.MySQL) (runtime.Object, error) {
	if mysql.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	if mysql.Spec.Halted {
		if mysql.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
			return nil, errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		mysql.Spec.TerminationPolicy = api.TerminationPolicyHalt
	}

	if mysql.Spec.Topology != nil && mysql.Spec.Topology.Mode != nil &&
		*mysql.Spec.Topology.Mode == api.MySQLClusterModeGroup {
		if mysql.Spec.Topology.Group == nil {
			mysql.Spec.Topology.Group = &api.MySQLGroupSpec{}
		}

		if mysql.Spec.Topology.Group.Name == "" {
			grName, err := uuid.NewRandom()
			if err != nil {
				return nil, errors.New("failed to generate a new group name")
			}
			mysql.Spec.Topology.Group.Name = grName.String()
		}

		if mysql.Spec.Topology.Group.BaseServerID == nil {
			mysql.Spec.Topology.Group.BaseServerID = types.Int64P(api.MySQLDefaultBaseServerID)
		}
	}

	mysql.SetDefaults()

	// If monitoring spec is given without port,
	// set default Listening port
	setMonitoringPort(mysql)

	return mysql, nil
}

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func setMonitoringPort(mysql *api.MySQL) {
	if mysql.Spec.Monitor != nil &&
		mysql.GetMonitoringVendor() == mona.VendorPrometheus {
		if mysql.Spec.Monitor.Prometheus == nil {
			mysql.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if mysql.Spec.Monitor.Prometheus.Exporter == nil {
			mysql.Spec.Monitor.Prometheus.Exporter = &mona.PrometheusExporterSpec{}
		}
		if mysql.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			mysql.Spec.Monitor.Prometheus.Exporter.Port = api.PrometheusExporterPortNumber
		}
	}
}
