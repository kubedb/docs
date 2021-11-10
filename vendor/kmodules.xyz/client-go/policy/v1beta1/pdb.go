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

package v1beta1

import (
	"context"

	"kmodules.xyz/client-go/discovery"

	"github.com/pkg/errors"
	policy "k8s.io/api/policy/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchPodDisruptionBudget(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget, opts metav1.PatchOptions) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	cur, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating PodDisruptionBudget %s/%s.", meta.Namespace, meta.Name)
		out, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Create(ctx, transform(&policy.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policy.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	mod := transform(cur.DeepCopy())
	if !apiequality.Semantic.DeepEqual(cur.Spec, mod.Spec) {
		// ref: https://github.com/kubernetes/kubernetes/issues/45398
		if ok, err := discovery.CheckAPIVersion(c.Discovery(), ">= 1.15"); err == nil && ok {
			return PatchPodDisruptionBudget(ctx, c, cur, transform, opts)
		}
		// PDBs dont have the specs, Specs can't be modified once created, so we have to delete first, then recreate with correct  spec
		klog.Warningf("Spec of PodDisruptionBudget %s/%s is modified, deleting existing one first.", meta.Namespace, meta.Name)
		err = c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Delete(ctx, meta.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		klog.V(3).Infof("Creating PodDisruptionBudget %s/%s.", mod.Namespace, mod.Name)
		out, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Create(ctx, transform(&policy.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policy.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		return out, kutil.VerbPatched, err
	}
	return cur, kutil.VerbUnchanged, nil
}

func PatchPodDisruptionBudget(ctx context.Context, c kubernetes.Interface, cur *policy.PodDisruptionBudget, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget, opts metav1.PatchOptions) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	return PatchPodDisruptionBudgetObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchPodDisruptionBudgetObject(ctx context.Context, c kubernetes.Interface, cur, mod *policy.PodDisruptionBudget, opts metav1.PatchOptions) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, policy.PodDisruptionBudget{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	klog.V(3).Infof("Patching PodDisruptionBudget %s with %s.", cur.Name, string(patch))
	out, err := c.PolicyV1beta1().PodDisruptionBudgets(cur.Namespace).Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdatePodDisruptionBudget(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget, opts metav1.UpdateOptions) (result *policy.PodDisruptionBudget, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		klog.Errorf("Attempt %d failed to update PodDisruptionBudget %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update PodDisruptionBudget %s after %d attempts due to %v", meta.Name, attempt, err)
	}
	return
}
