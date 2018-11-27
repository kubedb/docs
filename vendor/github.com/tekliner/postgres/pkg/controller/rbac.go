package controller

import (
	core_util "github.com/appscode/kutil/core/v1"
	rbac_util "github.com/appscode/kutil/rbac/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	le "github.com/kubedb/postgres/pkg/leader_election"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

func (c *Controller) deleteRole(postgres *api.Postgres) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(postgres.Namespace).Delete(postgres.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureRole(postgres *api.Postgres) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return rerr
	}

	// Create new Roles
	_, _, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{apps.GroupName},
					Resources:     []string{"statefulsets"},
					Verbs:         []string{"get"},
					ResourceNames: []string{postgres.OffshootName()},
				},
				{
					APIGroups: []string{core.GroupName},
					Resources: []string{"pods"},
					Verbs:     []string{"list", "patch"},
				},
				{
					APIGroups: []string{core.GroupName},
					Resources: []string{"configmaps"},
					Verbs:     []string{"create"},
				},
				{
					APIGroups:     []string{core.GroupName},
					Resources:     []string{"configmaps"},
					Verbs:         []string{"get", "update"},
					ResourceNames: []string{le.GetLeaderLockName(postgres.OffshootName())},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(postgres *api.Postgres) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(postgres.Namespace).Delete(postgres.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(postgres *api.Postgres) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return rerr
	}
	// Create new ServiceAccount
	_, _, err := core_util.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(postgres *api.Postgres) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(postgres.Namespace).Delete(postgres.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(postgres *api.Postgres) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return rerr
	}
	// Ensure new RoleBindings
	_, _, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     postgres.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      postgres.OffshootName(),
					Namespace: postgres.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) ensureRBACStuff(postgres *api.Postgres) error {
	// Create New Role
	if err := c.ensureRole(postgres); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(postgres); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(postgres); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteRBACStuff(postgres *api.Postgres) error {
	// Delete Existing Role
	if err := c.deleteRole(postgres); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(postgres); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(postgres); err != nil {
		return err
	}

	return nil
}
