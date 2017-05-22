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
	if _, err := c.ExtClient.Elastics(om.Namespace).Get(om.Name); err != nil {
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
		amc.LabelDatabaseKind: tapi.ResourceKindElastic,
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
	return nil
}

func (c *Controller) ResumeDatabase(dormantDb *tapi.DormantDatabase) error {
	origin := dormantDb.Spec.Origin
	objectMeta := origin.ObjectMeta
	elastic := &tapi.Elastic{
		ObjectMeta: kapi.ObjectMeta{
			Name:        objectMeta.Name,
			Namespace:   objectMeta.Namespace,
			Labels:      objectMeta.Labels,
			Annotations: objectMeta.Annotations,
		},
		Spec: *origin.Spec.Elastic,
	}
	_, err := c.ExtClient.Elastics(elastic.Namespace).Create(elastic)
	return err
}
