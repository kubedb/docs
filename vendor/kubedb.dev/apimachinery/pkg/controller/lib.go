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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/log"
	_ "gomodules.xyz/stow/azure"
	_ "gomodules.xyz/stow/google"
	_ "gomodules.xyz/stow/s3"
	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	core_util "kmodules.xyz/client-go/core/v1"
	policy_util "kmodules.xyz/client-go/policy/v1beta1"
)

const UtilVolumeName = "util-volume"

func (c *Controller) checkGoverningService(name, namespace string) (bool, error) {
	_, err := c.Client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func (c *Controller) CreateGoverningService(name, namespace string) error {
	// Check if service name exists
	found, err := c.checkGoverningService(name, namespace)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: core.ServiceSpec{
			Type:      core.ServiceTypeClusterIP,
			ClusterIP: core.ClusterIPNone,
		},
	}
	_, err = c.Client.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	return err
}

// GetVolumeForSnapshot returns pvc or empty directory depending on StorageType.
// In case of PVC, this function will create a PVC then returns the volume.
func (c *Controller) GetVolumeForSnapshot(st api.StorageType, pvcSpec *core.PersistentVolumeClaimSpec, jobName, namespace string) (*core.Volume, error) {
	if st == api.StorageTypeEphemeral {
		ed := core.EmptyDirVolumeSource{}
		if pvcSpec != nil {
			if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
				ed.SizeLimit = &sz
			}
		}
		return &core.Volume{
			Name: UtilVolumeName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &ed,
			},
		}, nil
	}

	volume := &core.Volume{
		Name: UtilVolumeName,
	}
	if len(pvcSpec.AccessModes) == 0 {
		pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
			core.ReadWriteOnce,
		}
		log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, *pvcSpec)
	}

	claim := &core.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: *pvcSpec,
	}
	if pvcSpec.StorageClassName != nil {
		claim.Annotations = map[string]string{
			"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
		}
	}

	if _, err := c.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Create(context.TODO(), claim, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	volume.PersistentVolumeClaim = &core.PersistentVolumeClaimVolumeSource{
		ClaimName: claim.Name,
	}

	return volume, nil
}

func (c *Controller) CreateStatefulSetPodDisruptionBudget(sts *appsv1.StatefulSet) error {
	owner := metav1.NewControllerRef(sts, appsv1.SchemeGroupVersion.WithKind("StatefulSet"))

	m := metav1.ObjectMeta{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}
	_, _, err := policy_util.CreateOrPatchPodDisruptionBudget(context.TODO(), c.Client, m,
		func(in *policyv1beta1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
			in.Labels = sts.Labels
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: sts.Spec.Template.Labels,
			}

			maxUnavailable := int32(math.Max(1, math.Floor((float64(*sts.Spec.Replicas)-1.0)/2.0)))
			in.Spec.MaxUnavailable = &intstr.IntOrString{IntVal: maxUnavailable}

			in.Spec.MinAvailable = nil
			return in
		}, metav1.PatchOptions{})
	return err
}

func (c *Controller) CreateDeploymentPodDisruptionBudget(deployment *appsv1.Deployment) error {
	owner := metav1.NewControllerRef(deployment, appsv1.SchemeGroupVersion.WithKind("Deployment"))

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
