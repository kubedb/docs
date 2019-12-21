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
package controller

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	rbac_util "kmodules.xyz/client-go/rbac/v1beta1"
)

func (c *Controller) deleteRole(etcd *api.Etcd) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(etcd.Namespace).Delete(etcd.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureRole(etcd *api.Etcd) error {
	owner := metav1.NewControllerRef(etcd, api.SchemeGroupVersion.WithKind(api.ResourceKindEtcd))

	// Create new Roles
	_, _, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      etcd.OffshootName(),
			Namespace: etcd.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{apps.GroupName},
					Resources:     []string{"statefulsets"},
					Verbs:         []string{"get"},
					ResourceNames: []string{etcd.OffshootName()},
				},
				{
					APIGroups: []string{core.GroupName},
					Resources: []string{"pods"},
					Verbs:     []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(etcd *api.Etcd) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(etcd.Namespace).Delete(etcd.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(etcd *api.Etcd) error {
	owner := metav1.NewControllerRef(etcd, api.SchemeGroupVersion.WithKind(api.ResourceKindEtcd))

	// Create new ServiceAccount
	_, _, err := core_util.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      etcd.OffshootName(),
			Namespace: etcd.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(etcd *api.Etcd) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(etcd.Namespace).Delete(etcd.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(etcd *api.Etcd) error {
	owner := metav1.NewControllerRef(etcd, api.SchemeGroupVersion.WithKind(api.ResourceKindEtcd))

	// Ensure new RoleBindings
	_, _, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      etcd.OffshootName(),
			Namespace: etcd.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     etcd.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      etcd.OffshootName(),
					Namespace: etcd.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) ensureRBACStuff(etcd *api.Etcd) error {
	// Create New Role
	if err := c.ensureRole(etcd); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(etcd); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(etcd); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteRBACStuff(etcd *api.Etcd) error {
	// Delete Existing Role
	if err := c.deleteRole(etcd); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(etcd); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(etcd); err != nil {
		return err
	}

	return nil
}
