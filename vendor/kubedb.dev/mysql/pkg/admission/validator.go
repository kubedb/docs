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
	"sync"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type MySQLValidator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &MySQLValidator{}

var forbiddenEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_ALLOW_EMPTY_PASSWORD",
	"MYSQL_RANDOM_ROOT_PASSWORD",
	"MYSQL_ONETIME_PASSWORD",
}

func (a *MySQLValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralMySQL,
		},
		api.ResourceSingularMySQL
}

func (a *MySQLValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *MySQLValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
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

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			// req.Object.Raw = nil, so read from kubernetes
			obj, err := a.extClient.KubedbV1alpha2().MySQLs(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
			} else if err == nil && obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`mysql "%v/%v" can't be terminated. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
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

			mysql := obj.(*api.MySQL).DeepCopy()
			oldMySQL := oldObject.(*api.MySQL).DeepCopy()
			oldMySQL.SetDefaults(a.ClusterTopology)
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldMySQL.Spec.AuthSecret == nil {
				oldMySQL.Spec.AuthSecret = mysql.Spec.AuthSecret
			}

			if err := validateUpdate(mysql, oldMySQL); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateMySQL(a.client, a.extClient, obj.(*api.MySQL), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

func validateGroupReplicas(replicas int32) error {
	if replicas < 3 || replicas > api.MySQLMaxGroupMembers {
		return fmt.Errorf("accepted value of 'spec.replicas' for group replication is in range [3, %d]",
			api.MySQLMaxGroupMembers)
	}
	return nil
}

func validateMySQLGroup(replicas int32, group api.MySQLGroupSpec) error {
	if err := validateGroupReplicas(replicas); err != nil {
		return err
	}

	// validate group name whether it is a valid uuid
	if _, err := uuid.Parse(group.Name); err != nil {
		return errors.Wrapf(err, "invalid group name is set")
	}

	return nil
}

// ValidateMySQL checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateMySQL(client kubernetes.Interface, extClient cs.Interface, mysql *api.MySQL, strictValidation bool) error {
	if mysql.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	mysqlVersion, err := extClient.CatalogV1alpha1().MySQLVersions().Get(context.TODO(), mysql.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mysql.Spec.Replicas == nil {
		return fmt.Errorf(`spec.replicas "%v" invalid. Value must be greater than 0, but for group replication this value shouldn't be more than %d'`,
			mysql.Spec.Replicas, api.MySQLMaxGroupMembers)
	}

	if mysql.Spec.Topology != nil {
		if mysql.Spec.Topology.Mode == nil {
			return errors.New("a valid 'spec.topology.mode' must be set for MySQL clustering")
		}

		if mysql.IsInnoDBCluster() && mysqlVersion.Spec.Router.Image == "" {
			return errors.Errorf("InnoDBCluster mode is not supported for MySQL version %s", mysqlVersion.Name)
		}

		// validation for group configuration is performed only when
		// 'spec.topology.mode' is set to "GroupReplication"
		if *mysql.Spec.Topology.Mode == api.MySQLModeGroupReplication {
			// if spec.topology.mode is "GroupReplication", spec.topology.group is set to default during mutating
			if err := validateMySQLGroup(*mysql.Spec.Replicas, *mysql.Spec.Topology.Group); err != nil {
				return err
			}
		}
	}

	if err := amv.ValidateEnvVar(mysql.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindMySQL); err != nil {
		return err
	}

	if mysql.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if mysql.Spec.StorageType != api.StorageTypeDurable && mysql.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, mysql.Spec.StorageType)
	}
	if err := amv.ValidateStorage(client, mysql.Spec.StorageType, mysql.Spec.Storage); err != nil {
		return err
	}

	authSecret := mysql.Spec.AuthSecret

	if strictValidation {
		if authSecret != nil {
			if _, err := client.CoreV1().Secrets(mysql.Namespace).Get(context.TODO(), authSecret.Name, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if mysqlVersion is deprecated.
		// If deprecated, return error
		if mysqlVersion.Spec.Deprecated {
			return fmt.Errorf("mysql %s/%s is using deprecated version %v. Skipped processing", mysql.Namespace, mysql.Name, mysqlVersion.Name)
		}
		if err := mysqlVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("mysql %s/%s is using invalid mysqlVersion %v. Skipped processing. reason: %v", mysql.Namespace,
				mysql.Name, mysqlVersion.Name, err)
		}
	}

	if mysql.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if mysql.Spec.StorageType == api.StorageTypeEphemeral && mysql.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		return fmt.Errorf(`'spec.terminationPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := mysql.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdate(obj, oldObj *api.MySQL) error {
	preconditions := meta_util.PreConditionSet{
		String: sets.NewString(
			"spec.storageType",
			"spec.authSecret",
			"spec.podTemplate.spec.nodeSelector",
		),
	}
	// Once the database has been initialized, don't let update the "spec.init" section
	if oldObj.Spec.Init != nil && oldObj.Spec.Init.Initialized {
		preconditions.Insert("spec.init")
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
