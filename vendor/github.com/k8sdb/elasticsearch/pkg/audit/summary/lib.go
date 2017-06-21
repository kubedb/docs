package summary

import (
	"encoding/json"
	"fmt"

	"github.com/k8sdb/elasticsearch/pkg/audit/type"
	elastic "gopkg.in/olivere/elastic.v3"
)

func newClient(host, port string) (*elastic.Client, error) {
	return elastic.NewClient(
		elastic.SetURL(fmt.Sprintf("http://%v:%v", host, port)),
		elastic.SetMaxRetries(10),
		elastic.SetSniff(false),
	)
}

func getAllIndices(client *elastic.Client) ([]string, error) {
	return client.IndexNames()
}

func getDataFromIndex(client *elastic.Client, indexName string) (*types.IndexInfo, error) {
	indexInfo := &types.IndexInfo{
		IdCount: make(map[string]int64),
	}

	// Get analyzer
	analyzerData, err := client.IndexGetSettings(indexName).Do()
	if err != nil {
		return &types.IndexInfo{}, err
	}

	dataByte, err := json.Marshal(analyzerData[indexName].Settings["index"])
	if err != nil {
		return &types.IndexInfo{}, err
	}

	if err := json.Unmarshal(dataByte, &indexInfo.Setting); err != nil {
		return &types.IndexInfo{}, err
	}

	// get mappings
	mappingData, err := client.GetMapping().Index(indexName).Do()
	if err != nil {
		return &types.IndexInfo{}, err
	}
	indexInfo.Mapping = mappingData

	// Count Ids
	mappingDataBype, err := json.Marshal(mappingData[indexName])
	if err != nil {
		return &types.IndexInfo{}, err
	}
	type esTypes struct {
		Mappings map[string]interface{} `json:"mappings"`
	}
	var esType esTypes
	if err := json.Unmarshal(mappingDataBype, &esType); err != nil {
		return &types.IndexInfo{}, err
	}
	for key := range esType.Mappings {
		counts, err := client.Count(indexName).Type(key).Do()
		if err != nil {
			return &types.IndexInfo{}, err
		}
		indexInfo.IdCount[key] = counts
	}
	return indexInfo, nil
}
