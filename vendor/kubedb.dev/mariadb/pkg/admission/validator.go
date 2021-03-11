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
	"fmt"
	"strings"
	"sync"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/pkg/errors"
	"gomodules.xyz/sets"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

// MariaDBValidator implements the AdmissionHook interface to validate the MariaDB resources
type MariaDBValidator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &MariaDBValidator{}

var forbiddenEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_ALLOW_EMPTY_PASSWORD",
	"MYSQL_RANDOM_ROOT_PASSWORD",
	"MYSQL_ONETIME_PASSWORD",
}

// Resource is the resource to use for hosting validating admission webhook.
func (a *MariaDBValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralMariaDB,
		},
		api.ResourceSingularMariaDB
}

// Initialize is called as a post-start hook
func (a *MariaDBValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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
func (a *MariaDBValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
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

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			// req.Object.Raw = nil, so read from kubernetes
			obj, err := a.extClient.KubedbV1alpha2().MariaDBs(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
			} else if err == nil && obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`mariadb "%v/%v" can't be paused. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
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

			md := obj.(*api.MariaDB).DeepCopy()
			oldMD := oldObject.(*api.MariaDB).DeepCopy()
			oldMD.SetDefaults(a.ClusterTopology)
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldMD.Spec.AuthSecret == nil {
				oldMD.Spec.AuthSecret = md.Spec.AuthSecret
			}

			if err := validateUpdate(md, oldMD); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateMariaDB(a.client, a.extClient, obj.(*api.MariaDB), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// validateCluster checks whether the configurations for MariaDB Cluster are ok
func validateCluster(db *api.MariaDB) error {
	if db.IsCluster() {
		clusterName := db.ClusterName()
		if len(clusterName) > api.MariaDBMaxClusterNameLength {
			return errors.Errorf(`'spec.md.clusterName' "%s" shouldn't have more than %d characters'`,
				clusterName, api.MariaDBMaxClusterNameLength)
		}
		if db.Spec.Init != nil && db.Spec.Init.Script != nil {
			return fmt.Errorf("`.spec.init.scriptSource` is not supported for cluster. For MariaDB cluster initialization see https://stash.run/docs/latest/addons/mariadb/guides/5.7/clusterd/")
		}
	}

	return nil
}

// ValidateMariaDB checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateMariaDB(client kubernetes.Interface, extClient cs.Interface, db *api.MariaDB, strictValidation bool) error {
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Replicas == nil {
		return fmt.Errorf(`'spec.replicas' "%v" invalid. Value must be 1 for standalone mariadb server, but for mariadb cluster, value must be greater than 0`,
			*db.Spec.Replicas)
	}

	if _, err := extClient.CatalogV1alpha1().MariaDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if db.IsCluster() && *db.Spec.Replicas < api.MariaDBDefaultClusterSize {
		return fmt.Errorf(`'spec.replicas' "%v" invalid. Value must be %d for mariadb cluster`,
			db.Spec.Replicas, api.MariaDBDefaultClusterSize)
	}

	if err := validateCluster(db); err != nil {
		return err
	}

	if err := amv.ValidateEnvVar(db.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindMariaDB); err != nil {
		return err
	}

	if db.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if db.Spec.StorageType != api.StorageTypeDurable && db.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, db.Spec.StorageType)
	}
	if err := amv.ValidateStorage(client, db.Spec.StorageType, db.Spec.Storage); err != nil {
		return err
	}

	authSecret := db.Spec.AuthSecret

	if strictValidation {
		if authSecret != nil {
			if _, err := client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), authSecret.Name, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if mariadb Version is deprecated.
		// If deprecated, return error
		dbVersion, err := extClient.CatalogV1alpha1().MariaDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}

		if dbVersion.Spec.Deprecated {
			return fmt.Errorf("mariadb %s/%s is using deprecated version %v. Skipped processing", db.Namespace, db.Name, dbVersion.Name)
		}

		if err := dbVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("mariadbVersion %s/%s is using invalid mariadbVersion %v. Skipped processing. reason: %v", dbVersion.Namespace,
				dbVersion.Name, dbVersion.Name, err)
		}
	}

	if db.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if db.Spec.StorageType == api.StorageTypeEphemeral && db.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		return fmt.Errorf(`'spec.terminationPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdate(obj, oldObj *api.MariaDB) error {
	preconditions := getPreconditionFunc(oldObj)
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError())
		}
		return err
	}
	return nil
}

func getPreconditionFunc(db *api.MariaDB) []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
	}
	// Once the database has been initialized, don't let update the "spec.init" section
	if db.Spec.Init != nil && db.Spec.Init.Initialized {
		preconditionSpecFields.Insert("spec.init")
	}
	for _, field := range preconditionSpecFields.List() {
		preconditions = append(preconditions,
			meta_util.RequireChainKeyUnchanged(field),
		)
	}
	return preconditions
}

var preconditionSpecFields = sets.NewString(
	"spec.storageType",
	"spec.storage",
	"spec.authSecret",
	"spec.podTemplate.spec.nodeSelector",
)

func preconditionFailedError() error {
	str := preconditionSpecFields.List()
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}
