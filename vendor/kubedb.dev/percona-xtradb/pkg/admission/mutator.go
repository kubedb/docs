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

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
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
			Group:    "mutators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "perconaxtradbmutators",
		},
		"perconaxtradbmutator"
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
	perconaxtradbMod, err := setDefaultValues(a.extClient, obj.(*api.PerconaXtraDB).DeepCopy())
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
func setDefaultValues(extClient cs.Interface, px *api.PerconaXtraDB) (runtime.Object, error) {
	if px.Spec.Version == "" {
		return nil, errors.New(`'spec.version' is missing`)
	}

	px.SetDefaults()

	if err := setDefaultsFromDormantDB(extClient, px); err != nil {
		return nil, err
	}

	// If monitoring spec is given without port,
	// set default Listening port
	setMonitoringPort(px)

	return px, nil
}

// setDefaultsFromDormantDB takes values from Similar Dormant Database
func setDefaultsFromDormantDB(extClient cs.Interface, px *api.PerconaXtraDB) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := extClient.KubedbV1alpha1().DormantDatabases(px.Namespace).Get(px.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}

	// Check DatabaseKind
	if value, _ := meta_util.GetStringValue(dormantDb.Labels, api.LabelDatabaseKind); value != api.ResourceKindPerconaXtraDB {
		return errors.New(fmt.Sprintf(`invalid PerconaXtraDB: "%v/%v". Exists DormantDatabase "%v/%v" of different Kind`, px.Namespace, px.Name, dormantDb.Namespace, dormantDb.Name))
	}

	// Check Origin Spec
	ddbOriginSpec := dormantDb.Spec.Origin.Spec.PerconaXtraDB
	ddbOriginSpec.SetDefaults()

	// If DatabaseSecret of new object is not given,
	// Take dormantDatabaseSecretName
	if px.Spec.DatabaseSecret == nil {
		px.Spec.DatabaseSecret = ddbOriginSpec.DatabaseSecret
	}

	// If Monitoring Spec of new object is not given,
	// Take Monitoring Settings from Dormant
	if px.Spec.Monitor == nil {
		px.Spec.Monitor = ddbOriginSpec.Monitor
	} else {
		ddbOriginSpec.Monitor = px.Spec.Monitor
	}

	// Skip checking UpdateStrategy
	ddbOriginSpec.UpdateStrategy = px.Spec.UpdateStrategy

	// Skip checking TerminationPolicy
	ddbOriginSpec.TerminationPolicy = px.Spec.TerminationPolicy

	if !meta_util.Equal(ddbOriginSpec, &px.Spec) {
		diff := meta_util.Diff(ddbOriginSpec, &px.Spec)
		log.Errorf("percona-xtradb spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff)
		return errors.New(fmt.Sprintf("percona-xtradb spec mismatches with OriginSpec in DormantDatabases. Diff: %v", diff))
	}

	if _, err := meta_util.GetString(px.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		px.Spec.Init != nil &&
		(px.Spec.Init.SnapshotSource != nil || px.Spec.Init.StashRestoreSession != nil) {
		px.Annotations = core_util.UpsertMap(px.Annotations, map[string]string{
			api.AnnotationInitialized: "",
		})
	}

	// Delete  Matching dormantDatabase in Controller

	return nil
}

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func setMonitoringPort(px *api.PerconaXtraDB) {
	if px.Spec.Monitor != nil &&
		px.GetMonitoringVendor() == mona.VendorPrometheus {
		if px.Spec.Monitor.Prometheus == nil {
			px.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if px.Spec.Monitor.Prometheus.Port == 0 {
			px.Spec.Monitor.Prometheus.Port = api.PrometheusExporterPortNumber
		}
	}
}
