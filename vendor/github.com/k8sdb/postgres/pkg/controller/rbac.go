package controller

import (
	tapi "github.com/k8sdb/apimachinery/api"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	rbac "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

func (c *Controller) deleteRole(postgres *tapi.Postgres) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(postgres.Namespace).Delete(postgres.Name, nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRole(postgres *tapi.Postgres) error {
	// Create new Roles
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgres.Name,
			Namespace: postgres.Namespace,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups:     []string{tapi.GroupName},
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
		},
	}
	if _, err := c.Client.RbacV1beta1().Roles(role.Namespace).Create(role); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteServiceAccount(postgres *tapi.Postgres) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(postgres.Namespace).Delete(postgres.Name, nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(postgres *tapi.Postgres) error {
	// Create new ServiceAccount
	sa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgres.Name,
			Namespace: postgres.Namespace,
		},
	}
	if _, err := c.Client.CoreV1().ServiceAccounts(sa.Namespace).Create(sa); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteRoleBinding(postgres *tapi.Postgres) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(postgres.Namespace).Delete(postgres.Name, nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(postgres *tapi.Postgres) error {
	// Create new RoleBindings
	roleBinding := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgres.Name,
			Namespace: postgres.Namespace,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "Role",
			Name:     postgres.Name,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      postgres.Name,
				Namespace: postgres.Namespace,
			},
		},
	}
	if _, err := c.Client.RbacV1beta1().RoleBindings(roleBinding.Namespace).Create(roleBinding); err != nil {
		return err
	}

	return nil
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
