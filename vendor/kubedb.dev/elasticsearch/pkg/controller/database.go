/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"fmt"
	"strings"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/elasticsearch/pkg/util/es"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/portforward"
)

func (c *Controller) getAllIndices(elasticsearch *api.Elasticsearch) (string, error) {
	var url string
	if meta.PossiblyInCluster() {
		url = elasticsearch.GetConnectionURL()
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
			c.ClientConfig,
			elasticsearch.Namespace,
			clientPodName,
			api.ElasticsearchRestPort,
		)
		if err := tunnel.ForwardPort(); err != nil {
			return "", err
		}
		url = fmt.Sprintf("%v://127.0.0.1:%d", elasticsearch.GetConnectionScheme(), tunnel.Local)
	}

	var reason error
	var indices []string
	err := wait.PollImmediate(time.Second*30, time.Minute*5, func() (bool, error) {
		client, err := es.GetElasticClient(c.Client, c.ExtClient, elasticsearch, url)
		if err != nil {
			log.Warningln(err)
			reason = err
			return false, nil
		}
		defer client.Stop()
		indices, err = client.GetIndexNames()
		if err != nil {
			log.Warningln(err)
			reason = err
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to get Elasticsearch indices. Reason: %v", reason)
	}
	return strings.Join(indices, ","), nil
}
