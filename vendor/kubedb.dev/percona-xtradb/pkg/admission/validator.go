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
	"strings"
	"sync"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"

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

// PerconaXtraDBValidator implements the AdmissionHook interface to validate the PerconaXtraDB resources
type PerconaXtraDBValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &PerconaXtraDBValidator{}

var forbiddenEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_ALLOW_EMPTY_PASSWORD",
	"MYSQL_RANDOM_ROOT_PASSWORD",
	"MYSQL_ONETIME_PASSWORD",
}

// Resource is the resource to use for hosting validating admission webhook.
func (a *PerconaXtraDBValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "validators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "perconaxtradbvalidators",
		},
		"perconaxtradbvalidator"
}

// Initialize is called as a post-start hook
func (a *PerconaXtraDBValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
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
func (a *PerconaXtraDBValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
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

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			// req.Object.Raw = nil, so read from kubernetes
			obj, err := a.extClient.KubedbV1alpha1().PerconaXtraDBs(req.Namespace).Get(req.Name, metav1.GetOptions{})
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
			} else if err == nil && obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`percona-xtradb "%v/%v" can't be paused. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
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

			px := obj.(*api.PerconaXtraDB).DeepCopy()
			oldPXC := oldObject.(*api.PerconaXtraDB).DeepCopy()
			oldPXC.SetDefaults()
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldPXC.Spec.DatabaseSecret == nil {
				oldPXC.Spec.DatabaseSecret = px.Spec.DatabaseSecret
			}

			if err := validateUpdate(px, oldPXC); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidatePerconaXtraDB(a.client, a.extClient, obj.(*api.PerconaXtraDB), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// validateCluster checks whether the configurations for PerconaXtraDB Cluster are ok
func validateCluster(px *api.PerconaXtraDB) error {
	if px.IsCluster() {
		clusterName := px.ClusterName()
		if len(clusterName) > api.PerconaXtraDBMaxClusterNameLength {
			return errors.Errorf(`'spec.px.clusterName' "%s" shouldn't have more than %d characters'`,
				clusterName, api.PerconaXtraDBMaxClusterNameLength)
		}
		if px.Spec.Init != nil && px.Spec.Init.ScriptSource != nil {
			return fmt.Errorf("`.spec.init.scriptSource` is not supported for cluster. For PerconaXtraDB cluster initialization see https://stash.run/docs/latest/addons/percona-xtradb/guides/5.7/clusterd/")
		}
	}

	return nil
}

// ValidatePerconaXtraDB checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidatePerconaXtraDB(client kubernetes.Interface, extClient cs.Interface, px *api.PerconaXtraDB, strictValidation bool) error {
	if px.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if px.Spec.Replicas == nil {
		return fmt.Errorf(`'spec.replicas' "%v" invalid. Value must be 1 for standalone percona-xtradb server, but for percona-xtradb cluster, value must be greater than 0`,
			*px.Spec.Replicas)
	}

	if pxVersion, err := extClient.CatalogV1alpha1().PerconaXtraDBVersions().Get(string(px.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	} else if px.IsCluster() && pxVersion.Spec.Version != api.PerconaXtraDBClusterRecommendedVersion {
		return errors.Errorf("unsupported version for xtradb cluster, recommended version is %s",
			api.PerconaXtraDBClusterRecommendedVersion)
	}

	if px.IsCluster() && *px.Spec.Replicas < api.PerconaXtraDBDefaultClusterSize {
		return fmt.Errorf(`'spec.replicas' "%v" invalid. Value must be %d for xtradb cluster`,
			px.Spec.Replicas, api.PerconaXtraDBDefaultClusterSize)
	}

	if err := validateCluster(px); err != nil {
		return err
	}

	if err := amv.ValidateEnvVar(px.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindPerconaXtraDB); err != nil {
		return err
	}

	if px.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if px.Spec.StorageType != api.StorageTypeDurable && px.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, px.Spec.StorageType)
	}
	if err := amv.ValidateStorage(client, px.Spec.StorageType, px.Spec.Storage); err != nil {
		return err
	}

	databaseSecret := px.Spec.DatabaseSecret

	if strictValidation {
		if databaseSecret != nil {
			if _, err := client.CoreV1().Secrets(px.Namespace).Get(databaseSecret.SecretName, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if percona-xtradb Version is deprecated.
		// If deprecated, return error
		pxVersion, err := extClient.CatalogV1alpha1().PerconaXtraDBVersions().Get(string(px.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pxVersion.Spec.Deprecated {
			return fmt.Errorf("percona-xtradb %s/%s is using deprecated version %v. Skipped processing", px.Namespace, px.Name, pxVersion.Name)
		}

		if err := pxVersion.ValidateSpecs(); err != nil {
			return fmt.Errorf("perconaXtraDBVersion %s/%s is using invalid perconaXtraDBVersion %v. Skipped processing. reason: %v", pxVersion.Namespace,
				pxVersion.Name, pxVersion.Name, err)
		}
	}

	if px.Spec.Init != nil &&
		px.Spec.Init.StashRestoreSession != nil &&
		databaseSecret == nil {
		return fmt.Errorf("for Stash RestoreSession init, 'spec.databaseSecret.secretName' of %v/%v needs to be similar to older database of restoression %v",
			px.Namespace, px.Name, px.Spec.Init.StashRestoreSession.Name)
	}

	if px.Spec.UpdateStrategy.Type == "" {
		return fmt.Errorf(`'spec.updateStrategy.type' is missing`)
	}

	if px.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if px.Spec.StorageType == api.StorageTypeEphemeral && px.Spec.TerminationPolicy == api.TerminationPolicyPause {
		return fmt.Errorf(`'spec.terminationPolicy: Pause' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := px.Spec.Monitor
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
