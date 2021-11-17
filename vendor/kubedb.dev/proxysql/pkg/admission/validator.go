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
	"sync"

	catalog_api "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

// ProxySQLValidator implements the AdmissionHook interface to validate the ProxySQL resources
type ProxySQLValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &ProxySQLValidator{}

var forbiddenEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_PROXY_USER",
	"MYSQL_PROXY_PASSWORD",
}

// Resource is the resource to use for hosting validating admission webhook.
func (a *ProxySQLValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralProxySQL,
		},
		api.ResourceSingularProxySQL
}

// Initialize is called as a post-start hook
func (a *ProxySQLValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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
func (a *ProxySQLValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindProxySQL {
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
			_, err := a.extClient.KubedbV1alpha2().ProxySQLs(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
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

			proxysql := obj.(*api.ProxySQL).DeepCopy()
			oldProxysql := oldObject.(*api.ProxySQL).DeepCopy()
			oldProxysql.SetDefaults()
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldProxysql.Spec.AuthSecret == nil {
				oldProxysql.Spec.AuthSecret = proxysql.Spec.AuthSecret
			}

			if err := validateUpdate(proxysql, oldProxysql); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateProxySQL(a.client, a.extClient, obj.(*api.ProxySQL), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// validateBackendWithMode checks whether the backend configurations for ProxySQL are ok
func validateBackendWithMode(extClient cs.Interface, db *api.ProxySQL) error {
	if db.Spec.Mode == nil {
		return errors.New("'.spec.mode' is missing")
	}
	if mode := db.Spec.Mode; *mode != api.LoadBalanceModeGalera &&
		*mode != api.LoadBalanceModeGroupReplication {
		return errors.Errorf("'.spec.mode' must be either %q or %q",
			api.LoadBalanceModeGalera, api.LoadBalanceModeGroupReplication)
	}

	backend := db.Spec.Backend
	if backend == nil || backend.Replicas == nil || backend.Ref == nil || backend.Ref.APIGroup == nil {
		return errors.New(`'.spec.backend' and all of its subfields are required`)
	}

	var err error
	var requiredMode api.LoadBalanceMode
	gk := schema.GroupKind{Group: *backend.Ref.APIGroup, Kind: backend.Ref.Kind}

	switch gk {
	case api.Kind(api.ResourceKindPerconaXtraDB):
		requiredMode = api.LoadBalanceModeGalera
		_, err = extClient.KubedbV1alpha2().PerconaXtraDBs(db.Namespace).Get(context.TODO(), backend.Ref.Name, metav1.GetOptions{})

	case api.Kind(api.ResourceKindMySQL):
		requiredMode = api.LoadBalanceModeGroupReplication
		_, err = extClient.KubedbV1alpha2().MySQLs(db.Namespace).Get(context.TODO(), backend.Ref.Name, metav1.GetOptions{})

	// TODO: add other cases for MySQL and MariaDB when they will be configured

	default:
		return errors.Errorf("invalid group kind '%v' is specified", gk.String())
	}

	if *db.Spec.Mode != requiredMode {
		return errors.Errorf("'.spec.mode' must be %q for %v",
			requiredMode, backend.Ref.Kind)
	}

	if err != nil && kerr.IsNotFound(err) {
		return errors.Errorf("%v object named '%v' is not found",
			backend.Ref.Kind, backend.Ref.Name)
	}
	return errors.Wrapf(err, "failed to get %v object named '%v'",
		backend.Ref.Kind, backend.Ref.Name)
}

// ValidateProxySQL checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateProxySQL(client kubernetes.Interface, extClient cs.Interface, db *api.ProxySQL, strictValidation bool) error {
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	var proxysqlVersion *catalog_api.ProxySQLVersion
	var err error
	if proxysqlVersion, err = extClient.CatalogV1alpha1().ProxySQLVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if db.Spec.Replicas == nil {
		return errors.New("'.spec.replicas' is missing")
	}

	if *db.Spec.Replicas != 1 {
		return errors.Errorf(`'.spec.replicas' "%v" is invalid. Currently, supported replicas for proxysql is 1`,
			*db.Spec.Replicas)
	}

	if err = validateBackendWithMode(extClient, db); err != nil {
		return err
	}

	if err = amv.ValidateEnvVar(db.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindProxySQL); err != nil {
		return err
	}

	authSecret := db.Spec.AuthSecret

	if strictValidation {
		if authSecret != nil {
			if _, err = client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), authSecret.Name, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if proxysql Version is deprecated.
		// If deprecated, return error
		if proxysqlVersion.Spec.Deprecated {
			return fmt.Errorf("proxysql %s/%s is using deprecated version %v. Skipped processing", db.Namespace, db.Name, proxysqlVersion.Name)
		}
	}

	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err = amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdate(obj, oldObj runtime.Object) error {
	preconditions := meta_util.PreConditionSet{
		String: sets.NewString(
			"spec.proxysqlSecret",
			"spec.podTemplate.spec.nodeSelector",
		),
	}
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}
