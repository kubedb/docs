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

func (c *Controller) deleteRole(elastic *tapi.Elasticsearch) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(elastic.Namespace).Delete(elastic.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRole(elastic *tapi.Elasticsearch) error {
	// Create new Roles
	_, err := kutilrbac.EnsureRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      elastic.Name,
			Namespace: elastic.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{tapi.ResourceTypeElasticsearch},
					ResourceNames: []string{elastic.Name},
					Verbs:         []string{"get"},
				},
				{
					APIGroups: []string{apiv1.GroupName},
					Resources: []string{"services", "endpoints"},
					Verbs:     []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(elastic *tapi.Elasticsearch) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(elastic.Namespace).Delete(elastic.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(elastic *tapi.Elasticsearch) error {
	// Create new ServiceAccount
	_, err := kutilcore.EnsureServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      elastic.OffshootName(),
			Namespace: elastic.Namespace,
		},
		func(in *apiv1.ServiceAccount) *apiv1.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(elastic *tapi.Elasticsearch) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(elastic.Namespace).Delete(elastic.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(elastic *tapi.Elasticsearch) error {
	// Ensure new RoleBindings
	_, err := kutilrbac.EnsureRoleBinding(
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

func (c *Controller) createRBACStuff(elastic *tapi.Elasticsearch) error {
	// Delete Existing Role
	if err := c.deleteRole(elastic); err != nil {
		return err
	}
	// Create New Role
	if err := c.createRole(elastic); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(elastic); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(elastic); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (c *Controller) deleteRBACStuff(elastic *tapi.Elasticsearch) error {
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
