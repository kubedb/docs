package controller

import (
	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/labels"
)

func (c *Controller) Exists(om *kapi.ObjectMeta) (bool, error) {
	if _, err := c.ExtClient.Postgreses(om.Namespace).Get(om.Name); err != nil {
		if !k8serr.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func (c *Controller) PauseDatabase(dormantDb *tapi.DormantDatabase) error {
	// Delete Service
	if err := c.DeleteService(dormantDb.Name, dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	statefulSetName := getStatefulSetName(dormantDb.Name)
	if err := c.DeleteStatefulSet(statefulSetName, dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) WipeOutDatabase(dormantDb *tapi.DormantDatabase) error {
	labelMap := map[string]string{
		amc.LabelDatabaseName: dormantDb.Name,
		amc.LabelDatabaseKind: tapi.ResourceKindPostgres,
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

	if dormantDb.Spec.Origin.Spec.Postgres.DatabaseSecret != nil {
		if err := c.deleteSecret(dormantDb); err != nil {
			return err
		}

	}

	return nil
}

func (c *Controller) deleteSecret(dormantDb *tapi.DormantDatabase) error {

	var secretFound bool = false
	dormantDatabaseSecret := dormantDb.Spec.Origin.Spec.Postgres.DatabaseSecret

	postgresList, err := c.ExtClient.Postgreses(dormantDb.Namespace).List(kapi.ListOptions{})
	if err != nil {
		return err
	}

	for _, postgres := range postgresList.Items {
		databaseSecret := postgres.Spec.DatabaseSecret
		if databaseSecret != nil {
			if databaseSecret.SecretName == dormantDatabaseSecret.SecretName {
				secretFound = true
				break
			}
		}
	}

	if !secretFound {
		labelMap := map[string]string{
			amc.LabelDatabaseKind: tapi.ResourceKindPostgres,
		}

		labelSelector := labels.SelectorFromSet(labelMap)

		dormantDatabaseList, err := c.ExtClient.DormantDatabases(dormantDb.Namespace).List(
			kapi.ListOptions{
				LabelSelector: labelSelector,
			},
		)
		if err != nil {
			return err
		}

		for _, ddb := range dormantDatabaseList.Items {
			if ddb.Name == dormantDb.Name {
				continue
			}

			databaseSecret := ddb.Spec.Origin.Spec.Postgres.DatabaseSecret
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

func (c *Controller) ResumeDatabase(dormantDb *tapi.DormantDatabase) error {
	origin := dormantDb.Spec.Origin
	objectMeta := origin.ObjectMeta
	postgres := &tapi.Postgres{
		ObjectMeta: kapi.ObjectMeta{
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
