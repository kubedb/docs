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

package controller

import (
	"context"
	"math"

	_ "gomodules.xyz/stow/azure"
	_ "gomodules.xyz/stow/google"
	_ "gomodules.xyz/stow/s3"
	apps "k8s.io/api/apps/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	core_util "kmodules.xyz/client-go/core/v1"
	policy_util "kmodules.xyz/client-go/policy/v1beta1"
)

// SyncStatefulSetPodDisruptionBudget syncs the PDB with the current state of the statefulSet.
// The maxUnavailable is calculated based statefulSet replica count, maxUnavailable = (replicas-1)/2.
// Also cleanup the PDB, when replica count is 1 or less.
func (c *Controller) SyncStatefulSetPodDisruptionBudget(sts *apps.StatefulSet) error {
	if sts == nil {
		return nil
	}
	// CleanUp PDB for statefulSet with replica 1
	if *sts.Spec.Replicas <= 1 {
		// pdb name & namespace is same as the corresponding statefulSet's name & namespace.
		err := c.Client.PolicyV1beta1().PodDisruptionBudgets(sts.Namespace).Delete(context.TODO(), sts.Name, metav1.DeleteOptions{})
		if !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}
	return c.SyncStatefulSetPDBWithCustomLabelSelectors(sts, *sts.Spec.Replicas, sts.Labels, sts.Spec.Selector.MatchLabels)
}

// Deprecated: CreateStatefulSetPodDisruptionBudget is deprecated. Use SyncStatefulSetPodDisruptionBudget instead.
func (c *Controller) CreateStatefulSetPodDisruptionBudget(sts *apps.StatefulSet) error {
	owner := metav1.NewControllerRef(sts, apps.SchemeGroupVersion.WithKind("StatefulSet"))

	m := metav1.ObjectMeta{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}
	_, _, err := policy_util.CreateOrPatchPodDisruptionBudget(context.TODO(), c.Client, m,
		func(in *policyv1beta1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
			in.Labels = sts.Labels
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: sts.Spec.Selector.MatchLabels,
			}

			maxUnavailable := int32(math.Max(1, math.Floor((float64(*sts.Spec.Replicas)-1.0)/2.0)))
			in.Spec.MaxUnavailable = &intstr.IntOrString{IntVal: maxUnavailable}

			in.Spec.MinAvailable = nil
			return in
		}, metav1.PatchOptions{})
	return err
}

func (c *Controller) SyncStatefulSetPDBWithCustomLabelSelectors(sts *apps.StatefulSet, replicas int32, labels map[string]string, matchLabelSelectors map[string]string) error {
	if sts == nil {
		return nil
	}
	pdbRef := metav1.ObjectMeta{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}

	r := int32(math.Max(1, math.Floor(float64(replicas-1)/2.0)))
	maxUnavailable := &intstr.IntOrString{IntVal: r}

	owner := metav1.NewControllerRef(sts, apps.SchemeGroupVersion.WithKind("StatefulSet"))
	_, _, err := policy_util.CreateOrPatchPodDisruptionBudget(context.TODO(), c.Client, pdbRef,
		func(in *policyv1beta1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
			in.Labels = labels
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: matchLabelSelectors,
			}
			in.Spec.MaxUnavailable = maxUnavailable
			in.Spec.MinAvailable = nil
			return in
		}, metav1.PatchOptions{})
	return err
}

func (c *Controller) CreateDeploymentPodDisruptionBudget(deployment *apps.Deployment) error {
	owner := metav1.NewControllerRef(deployment, apps.SchemeGroupVersion.WithKind("Deployment"))

	m := metav1.ObjectMeta{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
	}

	_, _, err := policy_util.CreateOrPatchPodDisruptionBudget(context.TODO(), c.Client, m,
		func(in *policyv1beta1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
			in.Labels = deployment.Labels
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: deployment.Spec.Template.Labels,
			}

			in.Spec.MaxUnavailable = nil

			in.Spec.MinAvailable = &intstr.IntOrString{IntVal: 1}
			return in
		}, metav1.PatchOptions{})
	return err
}
