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
package es

import (
	"context"
	"encoding/json"

	"github.com/appscode/go/crypto/rand"
	esv5 "gopkg.in/olivere/elastic.v5"
	"sigs.k8s.io/yaml"
)

type ESClientV5 struct {
	client *esv5.Client
}

var _ ESClient = &ESClientV5{}

func (c *ESClientV5) CreateIndex(count int) error {
	for i := 0; i < count; i++ {
		_, err := c.client.CreateIndex(rand.Characters(5)).Do(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ESClientV5) CountIndex() (int, error) {
	indices, err := c.client.IndexNames()
	if err != nil {
		return 0, err
	}
	return len(indices), nil
}

func (c *ESClientV5) GetIndexNames() ([]string, error) {
	return c.client.IndexNames()
}

func (c *ESClientV5) GetAllNodesInfo() ([]NodeInfo, error) {
	data, err := c.client.NodesInfo().Metric("settings").Do(context.Background())

	nodesInfo := make([]NodeInfo, 0)
	for _, v := range data.Nodes {
		var info NodeInfo
		info.Name = v.Name

		js, err := json.Marshal(v.Settings)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(js, &info.Settings)
		if err != nil {
			return nil, err
		}
		nodesInfo = append(nodesInfo, info)
	}
	return nodesInfo, err
}

func (c *ESClientV5) Stop() {
	c.client.Stop()
}
