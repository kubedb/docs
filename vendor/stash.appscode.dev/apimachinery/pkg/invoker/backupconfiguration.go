/*
Copyright AppsCode Inc. and Contributors

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

package invoker

import (
	"context"

	"stash.appscode.dev/apimachinery/apis/stash/v1alpha1"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	cs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stash_scheme "stash.appscode.dev/apimachinery/client/clientset/versioned/scheme"
	v1beta1_util "stash.appscode.dev/apimachinery/client/clientset/versioned/typed/stash/v1beta1/util"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/reference"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/meta"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type BackupConfigurationInvoker struct {
	backupConfig *v1beta1.BackupConfiguration
	stashClient  cs.Interface
}

func NewBackupConfigurationInvoker(stashClient cs.Interface, backupConfig *v1beta1.BackupConfiguration) BackupInvoker {
	return &BackupConfigurationInvoker{
		stashClient:  stashClient,
		backupConfig: backupConfig,
	}
}

func (inv *BackupConfigurationInvoker) GetObjectMeta() metav1.ObjectMeta {
	return inv.backupConfig.ObjectMeta
}

func (inv *BackupConfigurationInvoker) GetTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       v1beta1.ResourceKindBackupConfiguration,
		APIVersion: v1beta1.SchemeGroupVersion.String(),
	}
}

func (inv *BackupConfigurationInvoker) GetObjectRef() (*core.ObjectReference, error) {
	return reference.GetReference(stash_scheme.Scheme, inv.backupConfig)
}

func (inv *BackupConfigurationInvoker) GetOwnerRef() *metav1.OwnerReference {
	return metav1.NewControllerRef(inv.backupConfig, v1beta1.SchemeGroupVersion.WithKind(v1beta1.ResourceKindBackupConfiguration))
}

func (inv *BackupConfigurationInvoker) GetLabels() map[string]string {
	return inv.backupConfig.OffshootLabels()
}
func (inv *BackupConfigurationInvoker) AddFinalizer() error {
	updatedBackupConfig, _, err := v1beta1_util.PatchBackupConfiguration(context.TODO(), inv.stashClient.StashV1beta1(), inv.backupConfig, func(in *v1beta1.BackupConfiguration) *v1beta1.BackupConfiguration {
		in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, v1beta1.StashKey)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	inv.backupConfig = updatedBackupConfig
	return nil
}

func (inv *BackupConfigurationInvoker) RemoveFinalizer() error {
	updatedBackupConfig, _, err := v1beta1_util.PatchBackupConfiguration(context.TODO(), inv.stashClient.StashV1beta1(), inv.backupConfig, func(in *v1beta1.BackupConfiguration) *v1beta1.BackupConfiguration {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, v1beta1.StashKey)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	inv.backupConfig = updatedBackupConfig
	return nil
}

func (inv *BackupConfigurationInvoker) HasCondition(target *v1beta1.TargetRef, conditionType string) (bool, error) {
	backupConfig, err := inv.stashClient.StashV1beta1().BackupConfigurations(inv.backupConfig.Namespace).Get(context.TODO(), inv.backupConfig.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return kmapi.HasCondition(backupConfig.Status.Conditions, conditionType), nil
}

func (inv *BackupConfigurationInvoker) GetCondition(target *v1beta1.TargetRef, conditionType string) (int, *kmapi.Condition, error) {
	backupConfig, err := inv.stashClient.StashV1beta1().BackupConfigurations(inv.backupConfig.Namespace).Get(context.TODO(), inv.backupConfig.Name, metav1.GetOptions{})
	if err != nil {
		return -1, nil, err
	}
	idx, cond := kmapi.GetCondition(backupConfig.Status.Conditions, conditionType)
	return idx, cond, nil
}

func (inv *BackupConfigurationInvoker) SetCondition(target *v1beta1.TargetRef, newCondition kmapi.Condition) error {
	updatedBackupConfig, err := v1beta1_util.UpdateBackupConfigurationStatus(context.TODO(), inv.stashClient.StashV1beta1(), inv.backupConfig.ObjectMeta, func(in *v1beta1.BackupConfigurationStatus) (types.UID, *v1beta1.BackupConfigurationStatus) {
		in.Conditions = kmapi.SetCondition(in.Conditions, newCondition)
		in.Phase = calculateBackupInvokerPhase(inv.GetDriver(), in.Conditions)
		return inv.backupConfig.UID, in
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	inv.backupConfig = updatedBackupConfig
	return nil
}

func (inv *BackupConfigurationInvoker) IsConditionTrue(target *v1beta1.TargetRef, conditionType string) (bool, error) {
	backupConfig, err := inv.stashClient.StashV1beta1().BackupConfigurations(inv.backupConfig.Namespace).Get(context.TODO(), inv.backupConfig.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return kmapi.IsConditionTrue(backupConfig.Status.Conditions, conditionType), nil
}

func (inv *BackupConfigurationInvoker) GetTargetInfo() []BackupTargetInfo {
	return []BackupTargetInfo{
		{
			Task:                  inv.backupConfig.Spec.Task,
			Target:                inv.backupConfig.Spec.Target,
			RuntimeSettings:       inv.backupConfig.Spec.RuntimeSettings,
			TempDir:               inv.backupConfig.Spec.TempDir,
			InterimVolumeTemplate: inv.backupConfig.Spec.InterimVolumeTemplate,
			Hooks:                 inv.backupConfig.Spec.Hooks,
		},
	}
}

func (inv *BackupConfigurationInvoker) GetDriver() v1beta1.Snapshotter {
	driver := inv.backupConfig.Spec.Driver
	if driver == "" {
		driver = v1beta1.ResticSnapshotter
	}
	return driver
}

func (inv *BackupConfigurationInvoker) GetRepoRef() kmapi.ObjectReference {
	var repo kmapi.ObjectReference
	repo.Name = inv.backupConfig.Spec.Repository.Name
	repo.Namespace = inv.backupConfig.Spec.Repository.Namespace
	if repo.Namespace == "" {
		repo.Namespace = inv.backupConfig.Namespace
	}
	return repo
}

func (inv *BackupConfigurationInvoker) GetRepository() (*v1alpha1.Repository, error) {
	repo := inv.GetRepoRef()
	return inv.stashClient.StashV1alpha1().Repositories(repo.Namespace).Get(context.TODO(), repo.Name, metav1.GetOptions{})
}

func (inv *BackupConfigurationInvoker) GetRuntimeSettings() ofst.RuntimeSettings {
	return inv.backupConfig.Spec.RuntimeSettings
}

func (inv *BackupConfigurationInvoker) GetSchedule() string {
	return inv.backupConfig.Spec.Schedule
}

func (inv *BackupConfigurationInvoker) IsPaused() bool {
	return inv.backupConfig.Spec.Paused
}

func (inv *BackupConfigurationInvoker) GetBackupHistoryLimit() *int32 {
	return inv.backupConfig.Spec.BackupHistoryLimit
}

func (inv *BackupConfigurationInvoker) GetGlobalHooks() *v1beta1.BackupHooks {
	return nil
}

func (inv *BackupConfigurationInvoker) GetExecutionOrder() v1beta1.ExecutionOrder {
	return v1beta1.Sequential
}

func (inv *BackupConfigurationInvoker) NextInOrder(curTarget v1beta1.TargetRef, targetStatus []v1beta1.BackupTargetStatus) bool {
	for _, t := range inv.GetTargetInfo() {
		if t.Target != nil {
			if TargetMatched(t.Target.Ref, curTarget) {
				return true
			}
			if !TargetBackupCompleted(t.Target.Ref, targetStatus) {
				return false
			}
		}
	}
	// By default, return true so that nil target(i.e. cluster backup) does not get stuck here.
	return true
}

func (inv *BackupConfigurationInvoker) GetHash() string {
	return inv.backupConfig.GetSpecHash()
}

func (inv *BackupConfigurationInvoker) GetObjectJSON() (string, error) {
	jsonObj, err := meta.MarshalToJson(inv.backupConfig, v1beta1.SchemeGroupVersion)
	if err != nil {
		return "", err
	}
	return string(jsonObj), nil
}

func (inv *BackupConfigurationInvoker) GetRetentionPolicy() v1alpha1.RetentionPolicy {
	return inv.backupConfig.Spec.RetentionPolicy
}

func (inv *BackupConfigurationInvoker) GetPhase() v1beta1.BackupInvokerPhase {
	return inv.backupConfig.Status.Phase
}
