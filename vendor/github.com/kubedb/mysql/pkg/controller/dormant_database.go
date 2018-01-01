package controller

import (
	"encoding/json"
	"errors"

	"github.com/appscode/go/log"
	apps_util "github.com/appscode/kutil/apps/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) Exists(om *metav1.ObjectMeta) (bool, error) {
	mysql, err := c.ExtClient.MySQLs(om.Namespace).Get(om.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return mysql.DeletionTimestamp == nil, nil
}

func (c *Controller) PauseDatabase(dormantDb *api.DormantDatabase) error {
	mysql := &api.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		},
	}
	// Delete Service
	if err := c.deleteService(mysql.OffshootName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	if err := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
		Name:      mysql.OffshootName(),
		Namespace: dormantDb.Namespace,
	}); err != nil {
		log.Errorln(err)
		return err
	}

	if err := c.deleteRBACStuff(mysql); err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) WipeOutDatabase(dormantDb *api.DormantDatabase) error {
	labelMap := map[string]string{
		api.LabelDatabaseName: dormantDb.Name,
		api.LabelDatabaseKind: api.ResourceKindMySQL,
	}

	labelSelector := labels.SelectorFromSet(labelMap)

	if err := c.DeleteSnapshots(dormantDb.Namespace, labelSelector); err != nil {
		log.Errorln(err)
		return err
	}

	if err := c.DeletePersistentVolumeClaims(dormantDb.Namespace, labelSelector); err != nil {
		log.Errorln(err)
		return err
	}

	if dormantDb.Spec.Origin.Spec.MySQL.DatabaseSecret != nil {
		if err := c.deleteSecret(dormantDb, dormantDb.Spec.Origin.Spec.MySQL.DatabaseSecret); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) ResumeDatabase(dormantDb *api.DormantDatabase) error {
	origin := dormantDb.Spec.Origin
	objectMeta := origin.ObjectMeta

	if origin.Spec.MySQL.Init != nil {
		return errors.New("do not support InitSpec in spec.origin.mysql")
	}

	mysql := &api.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:        objectMeta.Name,
			Namespace:   objectMeta.Namespace,
			Labels:      objectMeta.Labels,
			Annotations: objectMeta.Annotations,
		},
		Spec: *origin.Spec.MySQL,
	}

	if mysql.Annotations == nil {
		mysql.Annotations = make(map[string]string)
	}

	for key, val := range dormantDb.Annotations {
		mysql.Annotations[key] = val
	}

	_, err := c.ExtClient.MySQLs(mysql.Namespace).Create(mysql)
	return err
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
					Name:        mysql.Name,
					Namespace:   mysql.Namespace,
					Labels:      mysql.Labels,
					Annotations: mysql.Annotations,
				},
				Spec: api.OriginSpec{
					MySQL: &mysql.Spec,
				},
			},
		},
	}

	if mysql.Spec.Init != nil {
		if initSpec, err := json.Marshal(mysql.Spec.Init); err == nil {
			dormantDb.Annotations = map[string]string{
				api.MySQLInitSpec: string(initSpec),
			}
		}
	}

	dormantDb.Spec.Origin.Spec.MySQL.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
