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
	mongodb, err := c.ExtClient.MongoDBs(om.Namespace).Get(om.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return mongodb.DeletionTimestamp == nil, nil
}

func (c *Controller) PauseDatabase(dormantDb *api.DormantDatabase) error {
	mongodb := &api.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		},
	}
	// Delete Service
	if err := c.deleteService(mongodb.OffshootName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	if err := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
		Name:      mongodb.OffshootName(),
		Namespace: dormantDb.Namespace,
	}); err != nil {
		log.Errorln(err)
		return err
	}

	if err := c.deleteRBACStuff(mongodb); err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) WipeOutDatabase(dormantDb *api.DormantDatabase) error {
	labelMap := map[string]string{
		api.LabelDatabaseName: dormantDb.Name,
		api.LabelDatabaseKind: api.ResourceKindMongoDB,
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

	if dormantDb.Spec.Origin.Spec.MongoDB.DatabaseSecret != nil {
		if err := c.deleteSecret(dormantDb, dormantDb.Spec.Origin.Spec.MongoDB.DatabaseSecret); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) ResumeDatabase(dormantDb *api.DormantDatabase) error {
	origin := dormantDb.Spec.Origin
	objectMeta := origin.ObjectMeta

	if origin.Spec.MongoDB.Init != nil {
		return errors.New("do not support InitSpec in spec.origin.mongodb")
	}

	mongodb := &api.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:        objectMeta.Name,
			Namespace:   objectMeta.Namespace,
			Labels:      objectMeta.Labels,
			Annotations: objectMeta.Annotations,
		},
		Spec: *origin.Spec.MongoDB,
	}

	if mongodb.Annotations == nil {
		mongodb.Annotations = make(map[string]string)
	}

	for key, val := range dormantDb.Annotations {
		mongodb.Annotations[key] = val
	}

	_, err := c.ExtClient.MongoDBs(mongodb.Namespace).Create(mongodb)
	return err
}

func (c *Controller) createDormantDatabase(mongodb *api.MongoDB) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongodb.Name,
			Namespace: mongodb.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMongoDB,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        mongodb.Name,
					Namespace:   mongodb.Namespace,
					Labels:      mongodb.Labels,
					Annotations: mongodb.Annotations,
				},
				Spec: api.OriginSpec{
					MongoDB: &mongodb.Spec,
				},
			},
		},
	}

	if mongodb.Spec.Init != nil {
		if initSpec, err := json.Marshal(mongodb.Spec.Init); err == nil {
			dormantDb.Annotations = map[string]string{
				api.MongoDBInitSpec: string(initSpec),
			}
		}
	}

	dormantDb.Spec.Origin.Spec.MongoDB.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
