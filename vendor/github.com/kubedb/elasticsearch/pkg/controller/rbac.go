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

func (c *Controller) deleteRole(elastic *api.Elasticsearch) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(elastic.Namespace).Delete(elastic.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureRole(elastic *api.Elasticsearch) error {
	// Create new Roles
	_, _, err := kutilrbac.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      elastic.Name,
			Namespace: elastic.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{api.ResourceTypeElasticsearch},
					ResourceNames: []string{elastic.Name},
					Verbs:         []string{"get"},
				},
				{
					APIGroups: []string{core.GroupName},
					Resources: []string{"services", "endpoints"},
					Verbs:     []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(elastic *api.Elasticsearch) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(elastic.Namespace).Delete(elastic.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureServiceAccount(elastic *api.Elasticsearch) error {
	// Create new ServiceAccount
	_, _, err := kutilcore.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      elastic.OffshootName(),
			Namespace: elastic.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(elastic *api.Elasticsearch) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(elastic.Namespace).Delete(elastic.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) ensureRoleBinding(elastic *api.Elasticsearch) error {
	// Ensure new RoleBindings
	_, _, err := kutilrbac.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      elastic.Name,
			Namespace: elastic.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     elastic.Name,
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      elastic.Name,
					Namespace: elastic.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) ensureRBACStuff(elastic *api.Elasticsearch) error {
	// Create New Role
	if err := c.ensureRole(elastic); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.ensureServiceAccount(elastic); err != nil {
		return err
	}

	// Create New RoleBinding
	if err := c.ensureRoleBinding(elastic); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteRBACStuff(elastic *api.Elasticsearch) error {
	// Delete Existing Role
	if err := c.deleteRole(elastic); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(elastic); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(elastic); err != nil {
		return err
	}
	return nil
}
