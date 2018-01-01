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

func (c *Controller) deleteRole(redis *api.Redis) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(redis.Namespace).Delete(redis.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureRole(redis *api.Redis) error {
	// Create new Roles
	_, _, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      redis.OffshootName(),
			Namespace: redis.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{api.ResourceTypeRedis},
					ResourceNames: []string{redis.Name},
					Verbs:         []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(redis *api.Redis) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(redis.Namespace).Delete(redis.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureServiceAccount(redis *api.Redis) error {
	// Create new ServiceAccount
	_, _, err := core_util.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      redis.OffshootName(),
			Namespace: redis.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(redis *api.Redis) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(redis.Namespace).Delete(redis.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureRoleBinding(redis *api.Redis) error {
	// Ensure new RoleBindings
	_, _, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      redis.OffshootName(),
			Namespace: redis.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     redis.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      redis.OffshootName(),
					Namespace: redis.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) ensureRBACStuff(redis *api.Redis) error {
	// Create New Role
	if err := c.ensureRole(redis); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.ensureServiceAccount(redis); err != nil {
		return err
	}

	// Create New RoleBinding
	if err := c.ensureRoleBinding(redis); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteRBACStuff(redis *api.Redis) error {
	// Delete Existing Role
	if err := c.deleteRole(redis); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(redis); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(redis); err != nil {
		return err
	}

	return nil
}
