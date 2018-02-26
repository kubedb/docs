package controller

import (
	"github.com/appscode/go/log"
	apps_util "github.com/appscode/kutil/apps/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) Exists(om *metav1.ObjectMeta) (bool, error) {
	postgres, err := c.ExtClient.Postgreses(om.Namespace).Get(om.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return postgres.DeletionTimestamp == nil, nil
}

func (c *Controller) PauseDatabase(dormantDb *api.DormantDatabase) error {
	postgres := &api.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		},
	}
	// Delete Service
	if err := c.deleteService(postgres.ServiceName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}
	if err := c.deleteService(postgres.ReplicasServiceName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	err := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
		Name:      dormantDb.OffshootName(),
		Namespace: dormantDb.Namespace,
	})
	if err != nil {
		return err
	}

	if c.opt.EnableRbac {
		if err := c.deleteRBACStuff(postgres); err != nil {
			return err
		}
	}

	if err := c.deleteLeaderLockConfigMap(dormantDb.ObjectMeta); err != nil {
		return err
	}

	return nil
}

func (c *Controller) WipeOutDatabase(dormantDb *api.DormantDatabase) error {
	labelMap := map[string]string{
		api.LabelDatabaseName: dormantDb.Name,
		api.LabelDatabaseKind: api.ResourceKindPostgres,
	}

	labelSelector := labels.SelectorFromSet(labelMap)

	if err := c.DeleteSnapshots(dormantDb.Namespace, labelSelector); err != nil {
		return err
	}

	if err := c.DeletePersistentVolumeClaims(dormantDb.Namespace, labelSelector); err != nil {
		return err
	}

	if dormantDb.Spec.Origin.Spec.Postgres.DatabaseSecret != nil {
		if err := c.deleteSecret(dormantDb, dormantDb.Spec.Origin.Spec.Postgres.DatabaseSecret); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) ResumeDatabase(dormantDb *api.DormantDatabase) error {
	origin := dormantDb.Spec.Origin
	objectMeta := origin.ObjectMeta

	postgres := &api.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:        objectMeta.Name,
			Namespace:   objectMeta.Namespace,
			Labels:      objectMeta.Labels,
			Annotations: objectMeta.Annotations,
		},
		Spec: *origin.Spec.Postgres,
	}

	_, err := c.ExtClient.Postgreses(postgres.Namespace).Create(postgres)
	return err
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
					Name:        postgres.Name,
					Namespace:   postgres.Namespace,
					Labels:      postgres.Labels,
					Annotations: postgres.Annotations,
				},
				Spec: api.OriginSpec{
					Postgres: &(postgres.Spec),
				},
			},
		},
	}

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
