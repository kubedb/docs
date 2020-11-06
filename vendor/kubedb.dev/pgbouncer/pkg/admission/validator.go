/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	cm_api "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	"gomodules.xyz/x/log"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type PgBouncerValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &PgBouncerValidator{}

func (a *PgBouncerValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralPgBouncer,
		},
		api.ResourceSingularPgBouncer
}

func (a *PgBouncerValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (pbValidator *PgBouncerValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindPgBouncer {
		status.Allowed = true
		return status
	}

	pbValidator.lock.RLock()
	defer pbValidator.lock.RUnlock()
	if !pbValidator.initialized {
		return hookapi.StatusUninitialized()
	}

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			// req.Object.Raw = nil, so read from kubernetes
			obj, err := pbValidator.extClient.KubedbV1alpha2().PgBouncers(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
			if kerr.IsNotFound(err) {
				log.Infoln("obj ", obj.Name, " already deleted")
			}
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
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

			pgbouncer := obj.(*api.PgBouncer).DeepCopy()
			oldPgBouncer := oldObject.(*api.PgBouncer).DeepCopy()
			oldPgBouncer.SetDefaults()

			if err := validateUpdate(pgbouncer, oldPgBouncer, req.Kind.Kind); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidatePgBouncer(pbValidator.client, pbValidator.extClient, obj.(*api.PgBouncer), true); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// ValidatePgBouncer checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidatePgBouncer(client kubernetes.Interface, extClient cs.Interface, db *api.PgBouncer, strictValidation bool) error {
	if db.Spec.Replicas == nil || *db.Spec.Replicas < 1 {
		return fmt.Errorf(`spec.replicas "%v" invalid. Value must be greater than zero`, db.Spec.Replicas)
	}

	if db.Spec.Version == "" {
		return fmt.Errorf(`spec.Version can't be empty`)
	}

	if db.Spec.TLS != nil {
		if db.Spec.TLS != nil {
			if *db.Spec.TLS.IssuerRef.APIGroup != cm_api.SchemeGroupVersion.Group {
				return fmt.Errorf(`spec.tls.client.issuerRef.apiGroup must be %s`, cm_api.SchemeGroupVersion.Group)
			}
			if (db.Spec.TLS.IssuerRef.Kind != cm_api.IssuerKind) && (db.Spec.TLS.IssuerRef.Kind != cm_api.ClusterIssuerKind) {
				return fmt.Errorf(`spec.tls.client.issuerRef.issuerKind must be either %s or %s`, cm_api.IssuerKind, cm_api.ClusterIssuerKind)
			}
		}
	}

	if strictValidation {
		// Check if pgbouncerVersion is absent or deprecated.
		// If deprecated, return error
		pgbouncerVersion, err := extClient.CatalogV1alpha1().PgBouncerVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}
		if pgbouncerVersion.Spec.Deprecated {
			return fmt.Errorf("pgbouncer %s/%s is using deprecated version %v. Skipped processing",
				db.Namespace, db.Name, pgbouncerVersion.Name)
		}
	}
	return nil
}

func validateUpdate(obj, oldObj runtime.Object, kind string) error {
	preconditions := getPreconditionFunc()
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError(kind))
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
	// nothing to check yet
}

func preconditionFailedError(kind string) error {
	str := preconditionSpecFields
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{"For resource kind ", kind, `, at least one of the following was changed for resource
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}
