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
	"context"
	"fmt"
	"strings"
	"sync"

	cat_api "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/coreos/go-semver/semver"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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

type MySQLValidator struct {
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
			Group:    "validators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "mysqlvalidators",
		},
		"mysqlvalidator"
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
			obj, err := a.extClient.KubedbV1alpha1().MySQLs(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
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
			oldMySQL.SetDefaults()
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldMySQL.Spec.DatabaseSecret == nil {
				oldMySQL.Spec.DatabaseSecret = mysql.Spec.DatabaseSecret
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

// recursivelyVersionCompare() receives two slices versionA and versionB of size 3 containing
// major, minor and patch parts of the given versions (versionA and versionB) in indices
// 0, 1 and 2 respectively. This function compares these parts of versionA and versionB. It returns,
//
// 		0;	if all parts of versionA are equal to corresponding parts of versionB
//		1;	if for some i, version[i] > versionB[i] where from j = 0 to i-1, versionA[j] = versionB[j]
//	   -1;	if for some i, version[i] < versionB[i] where from j = 0 to i-1, versionA[j] = versionB[j]
//
// ref: https://github.com/coreos/go-semver/blob/568e959cd89871e61434c1143528d9162da89ef2/semver/semver.go#L126-L141
func recursivelyVersionCompare(versionA []int64, versionB []int64) int {
	if len(versionA) == 0 {
		return 0
	}

	a := versionA[0]
	b := versionB[0]

	if a > b {
		return 1
	} else if a < b {
		return -1
	}

	return recursivelyVersionCompare(versionA[1:], versionB[1:])
}

// Currently, we support Group Replication for version 5.7.25. validateVersion()
// checks whether the given version has exactly these major (5), minor (7) and patch (25).
func validateGroupServerVersion(version string) error {
	recommended, err := semver.NewVersion(api.MySQLGRRecommendedVersion)
	if err != nil {
		return fmt.Errorf("unable to parse recommended MySQL version %s: %v", api.MySQLGRRecommendedVersion, err)
	}

	given, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("unable to parse given MySQL version %s: %v", version, err)
	}

	if cmp := recursivelyVersionCompare(recommended.Slice(), given.Slice()); cmp != 0 {
		return fmt.Errorf("currently supported MySQL server version for group replication is %s, but used %s",
			api.MySQLGRRecommendedVersion, version)
	}

	return nil
}

// On a replication master and each replication slave, the --server-id
// option must be specified to establish a unique replication ID in the
// range from 1 to 2^32 − 1. “Unique”, means that each ID must be different
// from every other ID in use by any other replication master or slave.
// ref: https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_server_id
//
// We calculate a unique server-id for each server using baseServerID field in MySQL CRD.
// Moreover we can use maximum of 9 servers in a group. So the baseServerID should be in
// range [0, (2^32 - 1) - 9]
func validateGroupBaseServerID(baseServerID int64) error {
	if 0 < baseServerID && baseServerID <= api.MySQLMaxBaseServerID {
		return nil
	}
	return fmt.Errorf("invalid baseServerId specified, should be in range [1, %d]", api.MySQLMaxBaseServerID)
}

func validateGroupReplicas(replicas int32) error {
	if replicas == 1 {
		return fmt.Errorf("group shouldn't start with 1 member, accepted value of 'spec.replicas' for group replication is in range [2, %d], default is %d if not specified",
			api.MySQLMaxGroupMembers, api.MySQLDefaultGroupSize)
	}

	if replicas > api.MySQLMaxGroupMembers {
		return fmt.Errorf("group size can't be greater than max size %d (see https://dev.mysql.com/doc/refman/5.7/en/group-replication-frequently-asked-questions.html",
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

	if err := validateGroupBaseServerID(*group.BaseServerID); err != nil {
		return err
	}

	return nil
}

// ValidateMySQL checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateMySQL(client kubernetes.Interface, extClient cs.Interface, mysql *api.MySQL, strictValidation bool) error {
	var (
		err   error
		myVer *cat_api.MySQLVersion
	)

	if mysql.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	if myVer, err = extClient.CatalogV1alpha1().MySQLVersions().Get(context.TODO(), string(mysql.Spec.Version), metav1.GetOptions{}); err != nil {
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

		// currently supported cluster mode for MySQL is "GroupReplication". So
		// '.spec.topology.mode' has been validated only for value "GroupReplication"
		if *mysql.Spec.Topology.Mode != api.MySQLClusterModeGroup {
			return errors.Errorf("currently supported cluster mode for MySQL is %[1]q, spec.topology.mode must be %[1]q",
				api.MySQLClusterModeGroup)
		}

		// validation for group configuration is performed only when
		// 'spec.topology.mode' is set to "GroupReplication"
		if *mysql.Spec.Topology.Mode == api.MySQLClusterModeGroup {
			// if spec.topology.mode is "GroupReplication", spec.topology.group is set to default during mutating
			if err = validateMySQLGroup(*mysql.Spec.Replicas, *mysql.Spec.Topology.Group); err != nil {
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

	databaseSecret := mysql.Spec.DatabaseSecret

	if strictValidation {
		if databaseSecret != nil {
			if _, err := client.CoreV1().Secrets(mysql.Namespace).Get(context.TODO(), databaseSecret.SecretName, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if mysqlVersion is deprecated.
		// If deprecated, return error
		mysqlVersion, err := extClient.CatalogV1alpha1().MySQLVersions().Get(context.TODO(), string(mysql.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}

		if mysqlVersion.Spec.Deprecated {
			return fmt.Errorf("mysql %s/%s is using deprecated version %v. Skipped processing", mysql.Namespace, mysql.Name, mysqlVersion.Name)
		}

		if err := mysqlVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("mysql %s/%s is using invalid mysqlVersion %v. Skipped processing. reason: %v", mysql.Namespace,
				mysql.Name, mysqlVersion.Name, err)
		}

		if mysql.Spec.Topology != nil && mysql.Spec.Topology.Mode != nil &&
			*mysql.Spec.Topology.Mode == api.MySQLClusterModeGroup {
			if err = validateGroupServerVersion(myVer.Spec.Version); err != nil {
				return err
			}
		}
	}

	if mysql.Spec.Init != nil &&
		mysql.Spec.Init.StashRestoreSession != nil &&
		databaseSecret == nil {
		return fmt.Errorf("for Snapshot init, 'spec.databaseSecret.secretName' of %v/%v needs to be similar to older database of restoesession %v",
			mysql.Namespace, mysql.Name, mysql.Spec.Init.StashRestoreSession.Name)
	}

	if mysql.Spec.UpdateStrategy.Type == "" {
		return fmt.Errorf(`'spec.updateStrategy.type' is missing`)
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

func validateUpdate(obj, oldObj runtime.Object) error {
	preconditions := getPreconditionFunc()
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError())
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
	"spec.storageType",
	"spec.storage",
	"spec.databaseSecret",
	"spec.init",
	"spec.podTemplate.spec.nodeSelector",
}

func preconditionFailedError() error {
	str := preconditionSpecFields
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}
