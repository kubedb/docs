package controller

import (
	kutilcore "github.com/appscode/kutil/core/v1"
	kutilrbac "github.com/appscode/kutil/rbac/v1beta1"
	"github.com/kubedb/apimachinery/apis/kubedb"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) deleteRole(mysql *api.MySQL) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(mysql.Namespace).Delete(mysql.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRole(mysql *api.MySQL) error {
	// Create new Roles
	_, err := kutilrbac.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      mysql.OffshootName(),
			Namespace: mysql.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{api.ResourceTypeMySQL},
					ResourceNames: []string{mysql.Name},
					Verbs:         []string{"get"},
				},
				{
					// TODO. Use this if secret is necessary, Otherwise remove it
					APIGroups:     []string{core.GroupName},
					Resources:     []string{"secrets"},
					ResourceNames: []string{mysql.Spec.DatabaseSecret.SecretName},
					Verbs:         []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(mysql *api.MySQL) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(mysql.Namespace).Delete(mysql.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(mysql *api.MySQL) error {
	// Create new ServiceAccount
	_, err := kutilcore.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      mysql.OffshootName(),
			Namespace: mysql.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(mysql *api.MySQL) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(mysql.Namespace).Delete(mysql.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(mysql *api.MySQL) error {
	// Ensure new RoleBindings
	_, err := kutilrbac.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      mysql.OffshootName(),
			Namespace: mysql.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     mysql.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      mysql.OffshootName(),
					Namespace: mysql.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createRBACStuff(mysql *api.MySQL) error {
	// Delete Existing Role
	if err := c.deleteRole(mysql); err != nil {
		return err
	}
	// Create New Role
	if err := c.createRole(mysql); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(mysql); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(mysql); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (c *Controller) deleteRBACStuff(mysql *api.MySQL) error {
	// Delete Existing Role
	if err := c.deleteRole(mysql); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(mysql); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(mysql); err != nil {
		return err
	}

	return nil
}
