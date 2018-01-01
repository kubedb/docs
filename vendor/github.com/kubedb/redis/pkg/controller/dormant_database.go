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
	redis, err := c.ExtClient.Redises(om.Namespace).Get(om.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return redis.DeletionTimestamp == nil, nil
}

func (c *Controller) PauseDatabase(dormantDb *api.DormantDatabase) error {
	redis := &api.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		},
	}
	// Delete Service
	if err := c.deleteService(redis.OffshootName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	if err := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
		Name:      redis.OffshootName(),
		Namespace: dormantDb.Namespace,
	}); err != nil {
		log.Errorln(err)
		return err
	}

	if err := c.deleteRBACStuff(redis); err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) WipeOutDatabase(dormantDb *api.DormantDatabase) error {
	labelMap := map[string]string{
		api.LabelDatabaseName: dormantDb.Name,
		api.LabelDatabaseKind: api.ResourceKindRedis,
	}

	labelSelector := labels.SelectorFromSet(labelMap)

	log.Info("No snapshot for Redis.")

	if err := c.DeletePersistentVolumeClaims(dormantDb.Namespace, labelSelector); err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) ResumeDatabase(dormantDb *api.DormantDatabase) error {
	origin := dormantDb.Spec.Origin
	objectMeta := origin.ObjectMeta

	redis := &api.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:        objectMeta.Name,
			Namespace:   objectMeta.Namespace,
			Labels:      objectMeta.Labels,
			Annotations: objectMeta.Annotations,
		},
		Spec: *origin.Spec.Redis,
	}

	if redis.Annotations == nil {
		redis.Annotations = make(map[string]string)
	}

	for key, val := range dormantDb.Annotations {
		redis.Annotations[key] = val
	}

	_, err := c.ExtClient.Redises(redis.Namespace).Create(redis)
	return err
}

func (c *Controller) createDormantDatabase(redis *api.Redis) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.Name,
			Namespace: redis.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindRedis,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        redis.Name,
					Namespace:   redis.Namespace,
					Labels:      redis.Labels,
					Annotations: redis.Annotations,
				},
				Spec: api.OriginSpec{
					Redis: &redis.Spec,
				},
			},
		},
	}

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
