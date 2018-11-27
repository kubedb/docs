package framework

import (
	"fmt"

	"github.com/appscode/kutil/tools/portforward"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/elasticsearch/pkg/controller"
	"github.com/kubedb/elasticsearch/pkg/util/es"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetClientPodName(elasticsearch *api.Elasticsearch) string {
	clientName := elasticsearch.Name

	if elasticsearch.Spec.Topology != nil {
		if elasticsearch.Spec.Topology.Client.Prefix != "" {
			clientName = fmt.Sprintf("%v-%v", elasticsearch.Spec.Topology.Client.Prefix, clientName)
		}
	}
	return fmt.Sprintf("%v-0", clientName)
}

func (f *Framework) GetElasticClient(meta metav1.ObjectMeta) (es.ESClient, error) {
	db, err := f.GetElasticsearch(meta)
	if err != nil {
		return nil, err
	}
	clientPodName := f.GetClientPodName(db)
	f.Tunnel = portforward.NewTunnel(
		f.kubeClient.CoreV1().RESTClient(),
		f.restConfig,
		db.Namespace,
		clientPodName,
		api.ElasticsearchRestPort,
	)
	if err := f.Tunnel.ForwardPort(); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%v://127.0.0.1:%d", db.GetConnectionScheme(), f.Tunnel.Local)
	c := controller.New(nil, f.kubeClient, nil, nil, nil, nil, nil, amc.Config{})
	return es.GetElasticClient(c.Client, db, url)
}
