package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/portforward"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/elasticsearch/pkg/util/es"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (c *Controller) getAllIndices(elasticsearch *api.Elasticsearch) (string, error) {
	var url string
	if meta.PossiblyInCluster() {
		url = fmt.Sprintf("https://%s.%s:%d", elasticsearch.OffshootName(), elasticsearch.Namespace, ElasticsearchRestPort)
	} else {
		clientName := elasticsearch.OffshootName()
		if elasticsearch.Spec.Topology != nil {
			if elasticsearch.Spec.Topology.Client.Prefix != "" {
				clientName = fmt.Sprintf("%v-%v", elasticsearch.Spec.Topology.Client.Prefix, clientName)
			}
		}
		clientPodName := fmt.Sprintf("%v-0", clientName)
		tunnel := portforward.NewTunnel(
			c.Client.CoreV1().RESTClient(),
			c.restConfig,
			elasticsearch.Namespace,
			clientPodName,
			ElasticsearchRestPort,
		)
		if err := tunnel.ForwardPort(); err != nil {
			return "", err
		}
		url = fmt.Sprintf("https://127.0.0.1:%d", tunnel.Local)
	}

	var indices []string
	err := wait.PollImmediate(time.Second*30, time.Minute*5, func() (bool, error) {
		client, err := es.GetElasticClient(c.Client, elasticsearch, url)
		if err != nil {
			return false, nil
		}
		defer client.Stop()
		indices, err = client.GetIndexNames()
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", errors.New("failed to get Elasticsearch indices")
	}
	return strings.Join(indices, ","), nil
}
