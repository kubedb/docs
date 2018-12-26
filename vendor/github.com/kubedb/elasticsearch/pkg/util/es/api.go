package es

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	esv5 "gopkg.in/olivere/elastic.v5"
	esv6 "gopkg.in/olivere/elastic.v6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	KeyAdminUserName = "ADMIN_USERNAME"
	KeyAdminPassword = "ADMIN_PASSWORD"
)

type ESClient interface {
	CreateIndex(count int) error
	CountIndex() (int, error)
	GetIndexNames() ([]string, error)
	GetAllNodesInfo() ([]NodeInfo, error)
	GetElasticsearchSummary(indexName string) (*api.ElasticsearchSummary, error)
	Stop()
}

type NodeSetting struct {
	Name   string `json:"name,omitempty"`
	Data   string `json:"data,omitempty"`
	Ingest string `json:"ingest,omitempty"`
	Master string `json:"master,omitempty"`
}

type PathSetting struct {
	Data []string `json:"data,omitempty"`
	Logs string   `json:"logs,omitempty"`
	Home string   `json:"home,omitempty"`
}

type Setting struct {
	Node *NodeSetting `json:"node,omitempty"`
	Path *PathSetting `json:"path,omitempty"`
}

type NodeInfo struct {
	Name     string   `json:"name,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	Settings *Setting `json:"settings,omitempty"`
}

func GetElasticClient(kc kubernetes.Interface, db *api.Elasticsearch, url string) (ESClient, error) {
	secret, err := kc.CoreV1().Secrets(db.Namespace).Get(db.Spec.DatabaseSecret.SecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	switch {
	case strings.HasPrefix(string(db.Spec.Version), "5."):
		client, err := esv5.NewClient(
			esv5.SetHttpClient(&http.Client{
				Timeout: 0,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}),
			esv5.SetBasicAuth(string(secret.Data[KeyAdminUserName]), string(secret.Data[KeyAdminPassword])),
			esv5.SetURL(url),
			esv5.SetHealthcheck(false), // don't check health here. otherwise error message can be misleading for invalid credentials
			esv5.SetSniff(false),
		)
		if err != nil {
			return nil, err
		}

		// do a manual health check to test client
		_, err = client.ClusterHealth().Do(context.Background())
		if err != nil {
			return nil, err
		}

		return &ESClientV5{client: client}, nil
	case strings.HasPrefix(string(db.Spec.Version), "6."):
		client, err := esv6.NewClient(
			esv6.SetHttpClient(&http.Client{
				Timeout: 0,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}),
			esv6.SetBasicAuth(string(secret.Data[KeyAdminUserName]), string(secret.Data[KeyAdminPassword])),
			esv6.SetURL(url),
			esv6.SetHealthcheck(false), // don't check health here. otherwise error message can be misleading for invalid credentials
			esv6.SetSniff(false),
		)
		if err != nil {
			return nil, err
		}

		// do a manual health check to test client
		_, err = client.ClusterHealth().Do(context.Background())
		if err != nil {
			return nil, err
		}

		return &ESClientV6{client: client}, nil
	}
	return nil, fmt.Errorf("unknown database verserion: %s", db.Spec.Version)
}
