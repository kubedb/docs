package report

import (
	"encoding/json"
	"fmt"

	tapi "github.com/k8sdb/apimachinery/api"
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

func getDataFromIndex(client *elastic.Client, indexName string) (*tapi.ElasticSummary, error) {
	esSummary := &tapi.ElasticSummary{
		IdCount: make(map[string]int64),
	}

	// Get analyzer
	analyzerData, err := client.IndexGetSettings(indexName).Do()
	if err != nil {
		return &tapi.ElasticSummary{}, err
	}

	dataByte, err := json.Marshal(analyzerData[indexName].Settings["index"])
	if err != nil {
		return &tapi.ElasticSummary{}, err
	}

	if err := json.Unmarshal(dataByte, &esSummary.Setting); err != nil {
		return &tapi.ElasticSummary{}, err
	}

	// get mappings
	mappingData, err := client.GetMapping().Index(indexName).Do()
	if err != nil {
		return &tapi.ElasticSummary{}, err
	}
	esSummary.Mapping = mappingData

	// Count Ids
	mappingDataBype, err := json.Marshal(mappingData[indexName])
	if err != nil {
		return &tapi.ElasticSummary{}, err
	}
	type esTypes struct {
		Mappings map[string]interface{} `json:"mappings"`
	}
	var esType esTypes
	if err := json.Unmarshal(mappingDataBype, &esType); err != nil {
		return &tapi.ElasticSummary{}, err
	}
	for key := range esType.Mappings {
		counts, err := client.Count(indexName).Type(key).Do()
		if err != nil {
			return &tapi.ElasticSummary{}, err
		}
		esSummary.IdCount[key] = counts
	}
	return esSummary, nil
}
