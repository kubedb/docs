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

func (c *Controller) deleteRole(mongodb *api.MongoDB) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(mongodb.Namespace).Delete(mongodb.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRole(mongodb *api.MongoDB) error {
	// Create new Roles
	_, err := kutilrbac.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      mongodb.OffshootName(),
			Namespace: mongodb.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{api.ResourceTypeMongoDB},
					ResourceNames: []string{mongodb.Name},
					Verbs:         []string{"get"},
				},
				{
					// TODO. Use this if secret is necessary, Otherwise remove it
					APIGroups:     []string{core.GroupName},
					Resources:     []string{"secrets"},
					ResourceNames: []string{mongodb.Spec.DatabaseSecret.SecretName},
					Verbs:         []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(mongodb *api.MongoDB) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(mongodb.Namespace).Delete(mongodb.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(mongodb *api.MongoDB) error {
	// Create new ServiceAccount
	_, err := kutilcore.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      mongodb.OffshootName(),
			Namespace: mongodb.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(mongodb *api.MongoDB) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(mongodb.Namespace).Delete(mongodb.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(mongodb *api.MongoDB) error {
	// Ensure new RoleBindings
	_, err := kutilrbac.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      mongodb.OffshootName(),
			Namespace: mongodb.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     mongodb.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      mongodb.OffshootName(),
					Namespace: mongodb.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createRBACStuff(mongodb *api.MongoDB) error {
	// Delete Existing Role
	if err := c.deleteRole(mongodb); err != nil {
		return err
	}
	// Create New Role
	if err := c.createRole(mongodb); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(mongodb); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(mongodb); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (c *Controller) deleteRBACStuff(mongodb *api.MongoDB) error {
	// Delete Existing Role
	if err := c.deleteRole(mongodb); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(mongodb); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(mongodb); err != nil {
		return err
	}

	return nil
}
