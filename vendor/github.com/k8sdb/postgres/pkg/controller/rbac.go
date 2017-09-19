package controller

import (
	kutilcore "github.com/appscode/kutil/core/v1"
	kutilrbac "github.com/appscode/kutil/rbac/v1beta1"
	"github.com/k8sdb/apimachinery/apis/kubedb"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	rbac "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

func (c *Controller) deleteRole(postgres *tapi.Postgres) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(postgres.Namespace).Delete(postgres.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRole(postgres *tapi.Postgres) error {
	// Create new Roles
	_, err := kutilrbac.EnsureRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{tapi.ResourceTypePostgres},
					ResourceNames: []string{postgres.Name},
					Verbs:         []string{"get"},
				},
				{
					APIGroups:     []string{apiv1.GroupName},
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

func (c *Controller) deleteServiceAccount(postgres *tapi.Postgres) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(postgres.Namespace).Delete(postgres.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(postgres *tapi.Postgres) error {
	// Create new ServiceAccount
	_, err := kutilcore.EnsureServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      postgres.OffshootName(),
			Namespace: postgres.Namespace,
		},
		func(in *apiv1.ServiceAccount) *apiv1.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(postgres *tapi.Postgres) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(postgres.Namespace).Delete(postgres.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(postgres *tapi.Postgres) error {
	// Ensure new RoleBindings
	_, err := kutilrbac.EnsureRoleBinding(
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

func (c *Controller) createRBACStuff(postgres *tapi.Postgres) error {
	// Delete Existing Role
	if err := c.deleteRole(postgres); err != nil {
		return err
	}
	// Create New Role
	if err := c.createRole(postgres); err != nil {
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
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (c *Controller) deleteRBACStuff(postgres *tapi.Postgres) error {
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
