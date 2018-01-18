package controller

import (
	"errors"
	"fmt"

	"github.com/appscode/go/log"
	apps_util "github.com/appscode/kutil/apps/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) Exists(om *metav1.ObjectMeta) (bool, error) {
	elasticsearch, err := c.ExtClient.Elasticsearchs(om.Namespace).Get(om.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return elasticsearch.DeletionTimestamp == nil, nil
}

func (c *Controller) PauseDatabase(dormantDb *api.DormantDatabase) error {
	elasticsearch := &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		},
	}
	// Delete Service
	if err := c.deleteService(elasticsearch.OffshootName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}
	if err := c.deleteService(elasticsearch.MasterServiceName(), dormantDb.Namespace); err != nil {
		log.Errorln(err)
		return err
	}

	topology := dormantDb.Spec.Origin.Spec.Elasticsearch.Topology
	if topology != nil {

		deleteStatefulSet := func(err chan<- error) error {
			clientName := dormantDb.OffshootName()
			if topology.Client.Prefix != "" {
				clientName = fmt.Sprintf("%v-%v", topology.Client.Prefix, clientName)
			}
			go func() {
				err2 := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
					Name:      clientName,
					Namespace: dormantDb.Namespace,
				})
				err <- err2
			}()

			masterName := dormantDb.OffshootName()
			if topology.Master.Prefix != "" {
				masterName = fmt.Sprintf("%v-%v", topology.Master.Prefix, masterName)
			}
			go func() {
				err2 := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
					Name:      masterName,
					Namespace: dormantDb.Namespace,
				})
				err <- err2
			}()

			dataName := dormantDb.OffshootName()
			if topology.Data.Prefix != "" {
				dataName = fmt.Sprintf("%v-%v", topology.Data.Prefix, dataName)
			}
			go func() {
				err2 := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
					Name:      dataName,
					Namespace: dormantDb.Namespace,
				})
				err <- err2
			}()

			return nil
		}

		errors := make(chan error, 3)

		go deleteStatefulSet(errors)

		for i := 1; i <= 3; i++ {
			err := <-errors
			if err != nil {
				return err
			}
		}

	} else {
		err := apps_util.DeleteStatefulSet(c.Client, metav1.ObjectMeta{
			Name:      dormantDb.OffshootName(),
			Namespace: dormantDb.Namespace,
		})
		if err != nil {
			return err
		}
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

	if dormantDb.Spec.Origin.Spec.Elasticsearch.DatabaseSecret != nil {
		if err := c.deleteSecret(dormantDb, dormantDb.Spec.Origin.Spec.Elasticsearch.DatabaseSecret); err != nil {
			return err
		}
	}

	if dormantDb.Spec.Origin.Spec.Elasticsearch.CertificateSecret != nil {
		if err := c.deleteSecret(dormantDb, dormantDb.Spec.Origin.Spec.Elasticsearch.CertificateSecret); err != nil {
			return err
		}
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
					Elasticsearch: &(elasticsearch.Spec),
				},
			},
		},
	}

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}
