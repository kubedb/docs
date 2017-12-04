package controller

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) Exists(om *metav1.ObjectMeta) (bool, error) {
	if _, err := c.ExtClient.Elasticsearchs(om.Namespace).Get(om.Name, metav1.GetOptions{}); err != nil {
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

	topology := dormantDb.Spec.Origin.Spec.Elasticsearch.Topology
	if topology != nil {
		clientName := dormantDb.OffshootName()
		if topology.Client.Prefix != "" {
			clientName = fmt.Sprintf("%v-%v", topology.Client.Prefix, clientName)
		}
		if err := c.DeleteStatefulSet(clientName, dormantDb.Namespace); err != nil {
			return err
		}

		masterName := dormantDb.OffshootName()
		if topology.Master.Prefix != "" {
			masterName = fmt.Sprintf("%v-%v", topology.Master.Prefix, masterName)
		}
		if err := c.DeleteStatefulSet(masterName, dormantDb.Namespace); err != nil {
			return err
		}

		dataName := dormantDb.OffshootName()
		if topology.Data.Prefix != "" {
			dataName = fmt.Sprintf("%v-%v", topology.Data.Prefix, dataName)
		}
		if err := c.DeleteStatefulSet(dataName, dormantDb.Namespace); err != nil {
			return err
		}
	} else {
		if err := c.DeleteStatefulSet(dormantDb.OffshootName(), dormantDb.Namespace); err != nil {
			return err
		}
	}

	elasticsearch := &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		},
	}
	if err := c.deleteRBACStuff(elasticsearch); err != nil {
		return err
	}
	return nil
}

func (c *Controller) WipeOutDatabase(dormantDb *api.DormantDatabase) error {
	labelMap := map[string]string{
		api.LabelDatabaseName: dormantDb.Name,
		api.LabelDatabaseKind: api.ResourceKindElasticsearch,
	}

	labelSelector := labels.SelectorFromSet(labelMap)

	if err := c.DeleteSnapshots(dormantDb.Namespace, labelSelector); err != nil {
		return err
	}

	if err := c.DeletePersistentVolumeClaims(dormantDb.Namespace, labelSelector); err != nil {
		return err
	}
	return nil
}

func (c *Controller) ResumeDatabase(dormantDb *api.DormantDatabase) error {
	origin := dormantDb.Spec.Origin
	objectMeta := origin.ObjectMeta

	if origin.Spec.Elasticsearch.Init != nil {
		return errors.New("do not support InitSpec in spec.origin.elasticsearch")
	}

	elasticsearch := &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:        objectMeta.Name,
			Namespace:   objectMeta.Namespace,
			Labels:      objectMeta.Labels,
			Annotations: objectMeta.Annotations,
		},
		Spec: *origin.Spec.Elasticsearch,
	}

	if elasticsearch.Annotations == nil {
		elasticsearch.Annotations = make(map[string]string)
	}

	for key, val := range dormantDb.Annotations {
		elasticsearch.Annotations[key] = val
	}

	_, err := c.ExtClient.Elasticsearchs(elasticsearch.Namespace).Create(elasticsearch)
	return err
}

func (c *Controller) createDormantDatabase(elasticsearch *api.Elasticsearch) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      elasticsearch.Name,
			Namespace: elasticsearch.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindElasticsearch,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        elasticsearch.Name,
					Namespace:   elasticsearch.Namespace,
					Labels:      elasticsearch.Labels,
					Annotations: elasticsearch.Annotations,
				},
				Spec: api.OriginSpec{
					Elasticsearch: &elasticsearch.Spec,
				},
			},
		},
	}

	if elasticsearch.Spec.Init != nil {
		initSpec, err := json.Marshal(elasticsearch.Spec.Init)
		if err == nil {
			dormantDb.Annotations = map[string]string{
				api.ElasticsearchInitSpec: string(initSpec),
			}
		}
	}

	dormantDb.Spec.Origin.Spec.Elasticsearch.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
