package controller

import (
	core_util "github.com/appscode/kutil/core/v1"
	rbac_util "github.com/appscode/kutil/rbac/v1beta1"
	"github.com/kubedb/apimachinery/apis/kubedb"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// Create new Roles
	_, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{api.ResourceTypePostgres},
					ResourceNames: []string{postgres.Name},
					Verbs:         []string{"get"},
				},
				{
					APIGroups:     []string{core.GroupName},
					Resources:     []string{"secrets"},
					ResourceNames: []string{postgres.Spec.DatabaseSecret.SecretName},
					Verbs:         []string{"get"},
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
	// Create new ServiceAccount
	_, err := core_util.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
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
	// Ensure new RoleBindings
	_, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
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
