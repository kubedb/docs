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

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
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

type PostgresValidator struct {
	ClusterTopology *core_util.Topology

	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &PostgresValidator{}

var forbiddenEnvVars = []string{
	"POSTGRES_PASSWORD",
	"POSTGRES_USER",
}

func (a *PostgresValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    kubedb.ValidatorGroupName,
			Version:  "v1alpha1",
			Resource: api.ResourcePluralPostgres,
		},
		api.ResourceSingularPostgres
}

func (a *PostgresValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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

func (a *PostgresValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindPostgres {
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
			obj, err := a.extClient.KubedbV1alpha2().Postgreses(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
			} else if err == nil && obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`postgres "%v/%v" can't be terminated. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
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

			postgres := obj.(*api.Postgres).DeepCopy()
			oldPostgres := oldObject.(*api.Postgres).DeepCopy()
			postgresVersion, err := a.extClient.CatalogV1alpha1().PostgresVersions().Get(context.TODO(), oldPostgres.Spec.Version, metav1.GetOptions{})
			if err != nil {
				return hookapi.StatusBadRequest(errors.Wrapf(err, "failed to get PostgresVersion: %s", oldPostgres.Spec.Version))
			}
			oldPostgres.SetDefaults(postgresVersion, a.ClusterTopology)
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldPostgres.Spec.AuthSecret == nil {
				oldPostgres.Spec.AuthSecret = postgres.Spec.AuthSecret
			}

			if err := validateUpdate(postgres, oldPostgres); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidatePostgres(a.client, a.extClient, obj.(*api.Postgres), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// ValidatePostgres checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidatePostgres(client kubernetes.Interface, extClient cs.Interface, postgres *api.Postgres, strictValidation bool) error {
	if postgres.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	if _, err := extClient.CatalogV1alpha1().PostgresVersions().Get(context.TODO(), string(postgres.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if postgres.Spec.Replicas == nil || *postgres.Spec.Replicas < 1 {
		return fmt.Errorf(`spec.replicas "%v" invalid. Value must be greater than zero`, postgres.Spec.Replicas)
	}

	if err := amv.ValidateEnvVar(postgres.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindPostgres); err != nil {
		return err
	}

	if postgres.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if postgres.Spec.StorageType != api.StorageTypeDurable && postgres.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, postgres.Spec.StorageType)
	}
	if err := amv.ValidateStorage(client, postgres.Spec.StorageType, postgres.Spec.Storage); err != nil {
		return err
	}

	if postgres.Spec.StandbyMode != nil {
		standByMode := *postgres.Spec.StandbyMode
		if standByMode != api.HotPostgresStandbyMode &&
			standByMode != api.WarmPostgresStandbyMode {
			return fmt.Errorf(`spec.standbyMode "%s" invalid`, standByMode)
		}
	}

	if postgres.Spec.StreamingMode != nil {
		streamingMode := *postgres.Spec.StreamingMode
		// TODO: synchronous Streaming is unavailable due to lack of support
		if streamingMode != api.AsynchronousPostgresStreamingMode &&
			streamingMode != api.SynchronousPostgresStreamingMode {
			return fmt.Errorf(`spec.streamingMode "%s" invalid`, streamingMode)
		}
	}

	pgVersion, err := extClient.CatalogV1alpha1().PostgresVersions().Get(context.TODO(), string(postgres.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}
	if postgres.Spec.ClientAuthMode == api.ClientAuthModeScram {
		if _, err := checkScramAuthMethodSupport(pgVersion.Spec.Version); err != nil {
			return err
		}
	}
	err = validateSpecForDB(postgres, pgVersion)
	if err != nil {
		return err
	}

	databaseSecret := postgres.Spec.AuthSecret
	if strictValidation {
		if databaseSecret != nil {
			if _, err := client.CoreV1().Secrets(postgres.Namespace).Get(context.TODO(), databaseSecret.Name, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if postgresVersion is deprecated.
		// If deprecated, return error
		postgresVersion, err := extClient.CatalogV1alpha1().PostgresVersions().Get(context.TODO(), string(postgres.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}
		if postgresVersion.Spec.Deprecated {
			return fmt.Errorf("postgres %s/%s is using deprecated version %v. Skipped processing",
				postgres.Namespace, postgres.Name, postgresVersion.Name)
		}

		if err := postgresVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("postgres %s/%s is using invalid postgresVersion %v. Skipped processing. reason: %v", postgres.Namespace,
				postgres.Name, postgresVersion.Name, err)
		}
	}

	if postgres.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if postgres.Spec.StorageType == api.StorageTypeEphemeral && postgres.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		return fmt.Errorf(`'spec.terminationPolicy: Halt' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := postgres.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdate(obj, oldObj *api.Postgres) error {
	preconditions := meta_util.PreConditionSet{
		String: sets.NewString(
			"spec.standby",
			"spec.streaming",
			"spec.databaseSecret",
			"spec.storageType",
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

func checkScramAuthMethodSupport(v string) (bool, error) {
	pgVersion, err := semver.NewVersion(v)
	if err != nil {
		return false, err
	}
	if pgVersion.Major() < 11 {
		return false, fmt.Errorf("scram auth method is available only for 11 or higher Versions")
	}
	return true, nil
}

func validateSpecForDB(postgres *api.Postgres, pgVersion *v1alpha1.PostgresVersion) error {
	// need to set the UserID and GroupID
	if pgVersion.Spec.SecurityContext.RunAsUser != nil &&
		postgres.Spec.PodTemplate.Spec.ContainerSecurityContext != nil &&
		pointer.Int64(postgres.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser) != pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser) &&
		!pgVersion.Spec.SecurityContext.RunAsAnyNonRoot {
		return fmt.Errorf("can't change ContainerSecurityContext's RunAsUser for this Postgres Version. It has to be the defualt UserID. The default UserID for this Postgres Version is %v but Container's security context UserID is %v", pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser), pointer.Int64(postgres.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser))
	}
	if pgVersion.Spec.SecurityContext.RunAsUser != nil &&
		postgres.Spec.PodTemplate.Spec.ContainerSecurityContext != nil &&
		pointer.Int64(postgres.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup) != pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser) &&
		!pgVersion.Spec.SecurityContext.RunAsAnyNonRoot {
		return fmt.Errorf("can't change ContainerSecurityContext's RunAsGroup for this Postgres Version. It has to be the defualt GroupID. The default GroupID for this Postgres Version is %v but Container's security context GroupID is %v", pointer.Int64(pgVersion.Spec.SecurityContext.RunAsUser), pointer.Int64(postgres.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup))
	}
	if postgres.Spec.PodTemplate.Spec.ContainerSecurityContext != nil &&
		postgres.Spec.PodTemplate.Spec.SecurityContext != nil &&
		(pointer.Int64(postgres.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser) != pointer.Int64(postgres.Spec.PodTemplate.Spec.SecurityContext.RunAsUser) ||
			pointer.Int64(postgres.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup) != pointer.Int64(postgres.Spec.PodTemplate.Spec.SecurityContext.RunAsGroup)) {
		return fmt.Errorf("Both pod's SecurityContext And ContainerSecurityContext's  RunAsGroup or RunAsOwner need to be equal")
	}
	if (postgres.Spec.ClientAuthMode == api.ClientAuthModeCert) &&
		(postgres.Spec.SSLMode == api.PostgresSSLModeDisable) {
		return fmt.Errorf("can't have %v set to postgres.spec.sslMode when postgres.spec.ClientAuthMode is set to %v",
			postgres.Spec.SSLMode, postgres.Spec.ClientAuthMode)
	}
	if (postgres.Spec.TLS != nil) &&
		(postgres.Spec.SSLMode == api.PostgresSSLModeDisable) {
		return fmt.Errorf("can't have %v set to postgres.spec.sslMode when postgres.spec.TLS is set ",
			postgres.Spec.SSLMode)
	}
	if (postgres.Spec.SSLMode != "" && postgres.Spec.SSLMode != api.PostgresSSLModeDisable) && postgres.Spec.TLS == nil {
		return fmt.Errorf("can't have %v set to postgres.Spec.SSLMode when postgres.Spec.TLS is null",
			postgres.Spec.SSLMode)
	}

	// validate leader election configs
	// ==============> start
	lec := postgres.Spec.LeaderElection
	if lec != nil {
		if lec.ElectionTick <= lec.HeartbeatTick {
			return fmt.Errorf("ElectionTick must be greater than HeartbeatTick")
		}
		if lec.ElectionTick < 1 {
			return fmt.Errorf("ElectionTick must be greater than zero")
		}
		if lec.HeartbeatTick < 1 {
			return fmt.Errorf("HeartbeatTick must be greater than zero")
		}
	}
	// end <==============
	return nil
}
