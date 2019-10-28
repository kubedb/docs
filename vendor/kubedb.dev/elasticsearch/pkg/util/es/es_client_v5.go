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
