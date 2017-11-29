package controller

import (
	"errors"

	"github.com/appscode/go/log"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) Exists(om *metav1.ObjectMeta) (bool, error) {
	if _, err := c.ExtClient.MySQLs(om.Namespace).Get(om.Name, metav1.GetOptions{}); err != nil {
		if !kerr.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func (c *Controller) PauseDatabase(dormantDb *api.DormantDatabase) error {
	// Delete Service
	if err := c.DeleteService(dormantDb.Name, dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	if err := c.DeleteStatefulSet(dormantDb.OffshootName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	mysql := &api.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		},
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
		if err := c.deleteSecret(dormantDb); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) deleteSecret(dormantDb *api.DormantDatabase) error {
	var secretFound bool = false

	dormantDatabaseSecret := dormantDb.Spec.Origin.Spec.MySQL.DatabaseSecret

	mysqlList, err := c.ExtClient.MySQLs(dormantDb.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, mysql := range mysqlList.Items {
		databaseSecret := mysql.Spec.DatabaseSecret
		if databaseSecret != nil {
			if databaseSecret.SecretName == dormantDatabaseSecret.SecretName {
				secretFound = true
				break
			}
		}
	}

	if !secretFound {
		labelMap := map[string]string{
			api.LabelDatabaseKind: api.ResourceKindMySQL,
		}
		dormantDatabaseList, err := c.ExtClient.DormantDatabases(dormantDb.Namespace).List(
			metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(labelMap).String(),
			},
		)
		if err != nil {
			return err
		}

		for _, ddb := range dormantDatabaseList.Items {
			if ddb.Name == dormantDb.Name {
				continue
			}

			databaseSecret := ddb.Spec.Origin.Spec.MySQL.DatabaseSecret
			if databaseSecret != nil {
				if databaseSecret.SecretName == dormantDatabaseSecret.SecretName {
					secretFound = true
					break
				}
			}
		}
	}

	if !secretFound {
		if err := c.DeleteSecret(dormantDatabaseSecret.SecretName, dormantDb.Namespace); err != nil {
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
