package controller

import (
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	rbac_util "github.com/appscode/kutil/rbac/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs_util "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) WaitUntilPaused(drmn *api.DormantDatabase) error {
	db := &api.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:      drmn.OffshootName(),
			Namespace: drmn.Namespace,
		},
	}

	if err := core_util.WaitUntilPodDeletedBySelector(c.Client, db.Namespace, metav1.SetAsLabelSelector(db.OffshootSelectors())); err != nil {
		return err
	}

	if err := core_util.WaitUntilServiceDeletedBySelector(c.Client, db.Namespace, metav1.SetAsLabelSelector(db.OffshootSelectors())); err != nil {
		return err
	}

	if err := c.waitUntilRBACStuffDeleted(db.ObjectMeta); err != nil {
		return err
	}

	if err := c.deleteLeaderLockConfigMap(db.ObjectMeta); err != nil {
		return err
	}

	return nil
}

func (c *Controller) waitUntilRBACStuffDeleted(meta metav1.ObjectMeta) error {
	// Delete Existing Role
	if err := rbac_util.WaitUntillRoleDeleted(c.Client, meta); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := rbac_util.WaitUntillRoleBindingDeleted(c.Client, meta); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := core_util.WaitUntillServiceAccountDeleted(c.Client, meta); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteMatchingDormantDatabase(postgres *api.Postgres) error {
	// Check if DormantDatabase exists or not
	ddb, err := c.ExtClient.DormantDatabases(postgres.Namespace).Get(postgres.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}

	// Set WipeOut to false
	if _, _, err := cs_util.PatchDormantDatabase(c.ExtClient, ddb, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Spec.WipeOut = false
		return in
	}); err != nil {
		return err
	}

	// Delete  Matching dormantDatabase
	if err := c.ExtClient.DormantDatabases(postgres.Namespace).Delete(postgres.Name,
		meta_util.DeleteInBackground()); err != nil && !kerr.IsNotFound(err) {
		return err
	}

	return nil
}

func (c *Controller) createDormantDatabase(postgres *api.Postgres) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgres.Name,
			Namespace: postgres.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindPostgres,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:              postgres.Name,
					Namespace:         postgres.Namespace,
					Labels:            postgres.Labels,
					Annotations:       postgres.Annotations,
					CreationTimestamp: postgres.CreationTimestamp,
				},
				Spec: api.OriginSpec{
					Postgres: &(postgres.Spec),
				},
			},
		},
	}

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
