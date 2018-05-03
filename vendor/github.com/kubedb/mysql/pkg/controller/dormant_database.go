package controller

import (
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs_util "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) WaitUntilPaused(drmn *api.DormantDatabase) error {
	db := &api.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      drmn.OffshootName(),
			Namespace: drmn.Namespace,
		},
	}

	if err := core_util.WaitUntilPodDeletedBySelector(c.Client, db.Namespace, metav1.SetAsLabelSelector(db.StatefulSetLabels())); err != nil {
		return err
	}

	if err := core_util.WaitUntilServiceDeletedBySelector(c.Client, db.Namespace, metav1.SetAsLabelSelector(db.OffshootLabels())); err != nil {
		return err
	}

	return nil
}

func (c *Controller) deleteMatchingDormantDatabase(mysql *api.MySQL) error {
	// Check if DormantDatabase exists or not
	ddb, err := c.ExtClient.DormantDatabases(mysql.Namespace).Get(mysql.Name, metav1.GetOptions{})
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
	if err := c.ExtClient.DormantDatabases(mysql.Namespace).Delete(mysql.Name,
		meta_util.DeleteInBackground()); err != nil && !kerr.IsNotFound(err) {
		return err
	}

	return nil
}

func (c *Controller) createDormantDatabase(mysql *api.MySQL) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.Name,
			Namespace: mysql.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMySQL,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:              mysql.Name,
					Namespace:         mysql.Namespace,
					Labels:            mysql.Labels,
					Annotations:       mysql.Annotations,
					CreationTimestamp: mysql.CreationTimestamp,
				},
				Spec: api.OriginSpec{
					MySQL: &mysql.Spec,
				},
			},
		},
	}

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
