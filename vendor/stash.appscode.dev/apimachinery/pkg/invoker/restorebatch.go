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
	"fmt"
	"time"

	"stash.appscode.dev/apimachinery/apis"
	"stash.appscode.dev/apimachinery/apis/stash/v1alpha1"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	cs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stash_scheme "stash.appscode.dev/apimachinery/client/clientset/versioned/scheme"
	v1beta1_util "stash.appscode.dev/apimachinery/client/clientset/versioned/typed/stash/v1beta1/util"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/reference"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

type RestoreBatchInvoker struct {
	kubeClient   kubernetes.Interface
	stashClient  cs.Interface
	restoreBatch *v1beta1.RestoreBatch
}

func NewRestoreBatchInvoker(kubeClient kubernetes.Interface, stashClient cs.Interface, restoreBatch *v1beta1.RestoreBatch) RestoreInvoker {
	return &RestoreBatchInvoker{
		kubeClient:   kubeClient,
		stashClient:  stashClient,
		restoreBatch: restoreBatch,
	}
}

func (inv *RestoreBatchInvoker) GetObjectMeta() metav1.ObjectMeta {
	return inv.restoreBatch.ObjectMeta
}

func (inv *RestoreBatchInvoker) GetTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       v1beta1.ResourceKindRestoreBatch,
		APIVersion: v1beta1.SchemeGroupVersion.String(),
	}
}

func (inv *RestoreBatchInvoker) GetObjectRef() (*core.ObjectReference, error) {
	return reference.GetReference(stash_scheme.Scheme, inv.restoreBatch)
}

func (inv *RestoreBatchInvoker) GetOwnerRef() *metav1.OwnerReference {
	return metav1.NewControllerRef(inv.restoreBatch, v1beta1.SchemeGroupVersion.WithKind(v1beta1.ResourceKindRestoreBatch))
}

func (inv *RestoreBatchInvoker) GetLabels() map[string]string {
	return inv.restoreBatch.OffshootLabels()
}

func (inv *RestoreBatchInvoker) AddFinalizer() error {
	updatedRestoreBatch, _, err := v1beta1_util.PatchRestoreBatch(context.TODO(), inv.stashClient.StashV1beta1(), inv.restoreBatch, func(in *v1beta1.RestoreBatch) *v1beta1.RestoreBatch {
		in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, v1beta1.StashKey)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	inv.restoreBatch = updatedRestoreBatch
	return nil
}

func (inv *RestoreBatchInvoker) RemoveFinalizer() error {
	updatedRestoreBatch, _, err := v1beta1_util.PatchRestoreBatch(context.TODO(), inv.stashClient.StashV1beta1(), inv.restoreBatch, func(in *v1beta1.RestoreBatch) *v1beta1.RestoreBatch {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, v1beta1.StashKey)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	inv.restoreBatch = updatedRestoreBatch
	return nil
}

func (inv *RestoreBatchInvoker) HasCondition(target *v1beta1.TargetRef, conditionType string) (bool, error) {
	restoreBatch, err := inv.stashClient.StashV1beta1().RestoreBatches(inv.restoreBatch.Namespace).Get(context.TODO(), inv.restoreBatch.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if target != nil {
		return hasRestoreMemberCondition(restoreBatch.Status.Members, *target, conditionType), nil
	}
	return kmapi.HasCondition(restoreBatch.Status.Conditions, conditionType), nil
}

func (inv *RestoreBatchInvoker) GetCondition(target *v1beta1.TargetRef, conditionType string) (int, *kmapi.Condition, error) {
	restoreBatch, err := inv.stashClient.StashV1beta1().RestoreBatches(inv.restoreBatch.Namespace).Get(context.TODO(), inv.restoreBatch.Name, metav1.GetOptions{})
	if err != nil {
		return -1, nil, err
	}
	if target != nil {
		idx, cond := getRestoreMemberCondition(restoreBatch.Status.Members, *target, conditionType)
		return idx, cond, nil
	}
	idx, cond := kmapi.GetCondition(restoreBatch.Status.Conditions, conditionType)
	return idx, cond, nil
}

func (inv *RestoreBatchInvoker) SetCondition(target *v1beta1.TargetRef, newCondition kmapi.Condition) error {
	status := inv.GetStatus()

	if target != nil {
		status.TargetStatus = setRestoreMemberCondition(status.TargetStatus, *target, newCondition)
	} else {
		status.Conditions = kmapi.SetCondition(status.Conditions, newCondition)
	}
	return inv.UpdateStatus(status)
}

func (inv *RestoreBatchInvoker) IsConditionTrue(target *v1beta1.TargetRef, conditionType string) (bool, error) {
	restoreBatch, err := inv.stashClient.StashV1beta1().RestoreBatches(inv.restoreBatch.Namespace).Get(context.TODO(), inv.restoreBatch.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if target != nil {
		return isRestoreMemberConditionTrue(restoreBatch.Status.Members, *target, conditionType), nil
	}
	return kmapi.IsConditionTrue(restoreBatch.Status.Conditions, conditionType), nil
}

func (inv *RestoreBatchInvoker) GetTargetInfo() []RestoreTargetInfo {
	var targetInfo []RestoreTargetInfo
	for _, member := range inv.restoreBatch.Spec.Members {
		targetInfo = append(targetInfo, RestoreTargetInfo{
			Task:                  member.Task,
			Target:                member.Target,
			RuntimeSettings:       member.RuntimeSettings,
			TempDir:               member.TempDir,
			InterimVolumeTemplate: member.InterimVolumeTemplate,
			Hooks:                 member.Hooks,
		})
	}
	return targetInfo
}

func (inv *RestoreBatchInvoker) GetDriver() v1beta1.Snapshotter {
	driver := inv.restoreBatch.Spec.Driver
	if driver == "" {
		driver = v1beta1.ResticSnapshotter
	}
	return driver
}

func (inv *RestoreBatchInvoker) GetRepoRef() kmapi.ObjectReference {
	var repo kmapi.ObjectReference
	repo.Name = inv.restoreBatch.Spec.Repository.Name
	repo.Namespace = inv.restoreBatch.Spec.Repository.Namespace
	if repo.Namespace == "" {
		repo.Namespace = inv.restoreBatch.Namespace
	}
	return repo
}

func (inv *RestoreBatchInvoker) GetRepository() (*v1alpha1.Repository, error) {
	repo := inv.GetRepoRef()
	return inv.stashClient.StashV1alpha1().Repositories(repo.Namespace).Get(context.TODO(), repo.Name, metav1.GetOptions{})
}

func (inv *RestoreBatchInvoker) GetGlobalHooks() *v1beta1.RestoreHooks {
	return inv.restoreBatch.Spec.Hooks
}

func (inv *RestoreBatchInvoker) GetExecutionOrder() v1beta1.ExecutionOrder {
	return inv.restoreBatch.Spec.ExecutionOrder
}

func (inv *RestoreBatchInvoker) NextInOrder(curTarget v1beta1.TargetRef, targetStatus []v1beta1.RestoreMemberStatus) bool {
	for _, t := range inv.GetTargetInfo() {
		if t.Target != nil {
			if TargetMatched(t.Target.Ref, curTarget) {
				return true
			}
			if !TargetRestoreCompleted(t.Target.Ref, targetStatus) {
				return false
			}
		}
	}
	// By default, return true so that nil target(i.e. cluster restore) does not get stuck here.
	return true
}

func (inv *RestoreBatchInvoker) GetHash() string {
	return inv.restoreBatch.GetSpecHash()
}

func (inv *RestoreBatchInvoker) GetObjectJSON() (string, error) {
	jsonObj, err := meta.MarshalToJson(inv.restoreBatch, v1beta1.SchemeGroupVersion)
	if err != nil {
		return "", err
	}
	return string(jsonObj), nil
}

func (inv *RestoreBatchInvoker) CreateEvent(eventType, source, reason, message string) error {
	objRef, err := inv.GetObjectRef()
	if err != nil {
		return err
	}

	t := metav1.Time{Time: time.Now()}
	if source == "" {
		source = EventSourceRestoreBatchController
	}
	_, err = inv.kubeClient.CoreV1().Events(inv.restoreBatch.Namespace).Create(context.TODO(), &core.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v.%x", inv.restoreBatch.Name, t.UnixNano()),
			Namespace: inv.restoreBatch.Namespace,
		},
		InvolvedObject: *objRef,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: t,
		LastTimestamp:  t,
		Count:          1,
		Type:           eventType,
		Source:         core.EventSource{Component: source},
	}, metav1.CreateOptions{})
	return err
}

func (inv *RestoreBatchInvoker) EnsureKubeDBIntegration(appClient appcatalog_cs.Interface) error {
	for i := range inv.restoreBatch.Spec.Members {
		target := inv.restoreBatch.Spec.Members[i].Target
		// Don't do anything if the target is not an AppBinding
		if target == nil || !TargetOfGroupKind(target.Ref, appcat.SchemeGroupVersion.Group, appcat.ResourceKindApp) {
			continue
		}

		appBinding, err := appClient.AppcatalogV1alpha1().AppBindings(inv.restoreBatch.Namespace).Get(context.TODO(), target.Ref.Name, metav1.GetOptions{})
		if err != nil {
			// If the AppBinding does not exist, then don't do anything.
			if kerr.IsNotFound(err) {
				continue
			}
			return err
		}
		// If the AppBinding is not managed by KubeDB, then don't do anything
		if manager, err := meta.GetStringValue(appBinding.Labels, meta.ManagedByLabelKey); err != nil || manager != "kubedb.com" {
			continue
		}
		// Extract the name, and managed-by labels. We are not passing "instance" label because there could be multiple AppBindings.
		appLabels, err := extractLabels(appBinding.Labels, meta.ManagedByLabelKey, meta.NameLabelKey)
		if err != nil {
			return err
		}

		// Add the labels to the invoker
		updatedRestoreBatch, _, err := v1beta1_util.PatchRestoreBatch(context.TODO(), inv.stashClient.StashV1beta1(), inv.restoreBatch, func(in *v1beta1.RestoreBatch) *v1beta1.RestoreBatch {
			in.Labels = meta.OverwriteKeys(in.Labels, appLabels)
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		inv.restoreBatch = updatedRestoreBatch
		return nil
	}
	return nil
}

func (inv *RestoreBatchInvoker) GetStatus() RestoreInvokerStatus {
	return getInvokerStatusFromRestoreBatch(inv.restoreBatch)
}

func (inv *RestoreBatchInvoker) UpdateStatus(status RestoreInvokerStatus) error {
	startTime := inv.GetObjectMeta().CreationTimestamp
	totalTargets := len(inv.GetTargetInfo())
	updatedRestoreBatch, err := v1beta1_util.UpdateRestoreBatchStatus(
		context.TODO(),
		inv.stashClient.StashV1beta1(),
		inv.restoreBatch.ObjectMeta,
		func(in *v1beta1.RestoreBatchStatus) (types.UID, *v1beta1.RestoreBatchStatus) {
			if len(status.Conditions) > 0 {
				in.Conditions = upsertConditions(in.Conditions, status.Conditions)
			}
			if len(status.TargetStatus) > 0 {
				for i := range status.TargetStatus {
					in.Members = upsertRestoreMemberStatus(in.Members, status.TargetStatus[i])
				}
			}

			in.Phase = calculateRestoreBatchPhase(in, totalTargets)
			if IsRestoreCompleted(in.Phase) && in.SessionDuration == "" {
				duration := time.Since(startTime.Time)
				in.SessionDuration = duration.Round(time.Second).String()
			}
			return inv.restoreBatch.ObjectMeta.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	inv.restoreBatch = updatedRestoreBatch
	return nil
}

func upsertRestoreMemberStatus(cur []v1beta1.RestoreMemberStatus, new v1beta1.RestoreMemberStatus) []v1beta1.RestoreMemberStatus {
	// if the member status already exist, then update it
	for i := range cur {
		if TargetMatched(cur[i].Ref, new.Ref) {
			cur[i] = upsertRestoreTargetStatus(cur[i], new)
			return cur
		}
	}
	// the member status does not exist. so, add new entry.
	cur = append(cur, new)
	return cur
}

func calculateRestoreBatchPhase(status *v1beta1.RestoreBatchStatus, totalTargets int) v1beta1.RestorePhase {
	if len(status.Conditions) == 0 || len(status.Members) == 0 ||
		kmapi.IsConditionFalse(status.Conditions, apis.RepositoryFound) ||
		kmapi.IsConditionFalse(status.Conditions, apis.BackendSecretFound) {
		return v1beta1.RestorePending
	}

	if kmapi.IsConditionFalse(status.Conditions, apis.ValidationPassed) {
		return v1beta1.RestorePhaseInvalid
	}

	failedTargetCount := 0
	unknownTargetCount := 0
	successfulTargetCount := 0

	for _, m := range status.Members {
		switch m.Phase {
		case v1beta1.TargetRestoreFailed:
			failedTargetCount++
		case v1beta1.TargetRestorePhaseUnknown:
			unknownTargetCount++
		case v1beta1.TargetRestoreSucceeded:
			successfulTargetCount++
		}
	}
	completedTargets := successfulTargetCount + failedTargetCount + unknownTargetCount
	if completedTargets < len(status.Members) || completedTargets < totalTargets {
		return v1beta1.RestoreRunning
	}
	if failedTargetCount > 0 {
		return v1beta1.RestoreFailed
	}
	if unknownTargetCount > 0 {
		return v1beta1.RestorePhaseUnknown
	}

	return v1beta1.RestoreSucceeded
}
