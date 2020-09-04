/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package elastic_stack

import (
	"context"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	core_util "kmodules.xyz/client-go/core/v1"
	policy_util "kmodules.xyz/client-go/policy/v1beta1"
)

func (es *Elasticsearch) createPodDisruptionBudget(sts *appsv1.StatefulSet, maxUnavailable *intstr.IntOrString) error {
	if sts == nil {
		return errors.New("statefulSet is empty")
	}
	owner := metav1.NewControllerRef(sts, appsv1.SchemeGroupVersion.WithKind("StatefulSet"))

	m := metav1.ObjectMeta{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}
	_, _, err := policy_util.CreateOrPatchPodDisruptionBudget(context.TODO(), es.kClient, m,
		func(in *policyv1beta1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
			in.Labels = sts.Labels

			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: sts.Spec.Template.Labels,
			}

			in.Spec.MaxUnavailable = maxUnavailable
			in.Spec.MinAvailable = nil
			return in
		}, metav1.PatchOptions{})
	return err
}
