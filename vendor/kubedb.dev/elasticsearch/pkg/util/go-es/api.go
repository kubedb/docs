/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package go_es

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	esv6 "github.com/elastic/go-elasticsearch/v6"
	esv7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type ESClient interface {
	ClusterStatus() (string, error)
}

var response map[string]interface{}

func GetElasticClient(kc kubernetes.Interface, db *api.Elasticsearch, esVersion, url string) (ESClient, error) {
	var username, password string
	if !db.Spec.DisableSecurity && db.Spec.AuthSecret != nil {
		secret, err := kc.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get secret: %s for Elasticsearch: %s/%s with: %s", db.Spec.AuthSecret.Name, db.Namespace, db.Name, err.Error())
			return nil, errors.Wrap(err, "failed to get the secret")
		}

		if value, ok := secret.Data[core.BasicAuthUsernameKey]; ok {
			username = string(value)
		} else {
			klog.Errorf("Failed for secret: %s/%s, username is missing", secret.Namespace, secret.Name)
			return nil, errors.New("username is missing")
		}

		if value, ok := secret.Data[core.BasicAuthPasswordKey]; ok {
			password = string(value)
		} else {
			klog.Errorf("Failed for secret: %s/%s, password is missing", secret.Namespace, secret.Name)
			return nil, errors.New("password is missing")
		}
	}

	switch {
	// for Elasticsearch 6.x.x
	case strings.HasPrefix(esVersion, "6."):
		client, err := esv6.NewClient(esv6.Config{
			Addresses:         []string{url},
			Username:          username,
			Password:          password,
			EnableDebugLogger: true,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					MaxVersion:         tls.VersionTLS12,
				},
			},
		})
		if err != nil {
			klog.Errorf("Failed to create HTTP client for Elasticsearch: %s/%s with: %s", db.Namespace, db.Name, err.Error())
			return nil, err
		}
		// do a manual health check to test client
		res, err := client.Cluster.Health(
			client.Cluster.Health.WithPretty(),
		)
		if err != nil {
			return nil, err
		}
		if res.IsError() {
			return nil, fmt.Errorf("health check failed with status code: %d", res.StatusCode)
		}
		return &ESClientV6{client: client}, nil

	// for Elasticsearch 7.x.x
	case strings.HasPrefix(esVersion, "7."):
		client, err := esv7.NewClient(esv7.Config{
			Addresses:         []string{url},
			Username:          username,
			Password:          password,
			EnableDebugLogger: true,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					MaxVersion:         tls.VersionTLS12,
				},
			},
		})
		if err != nil {
			klog.Errorf("Failed to create HTTP client for Elasticsearch: %s/%s with: %s", db.Namespace, db.Name, err.Error())
			return nil, err
		}
		// do a manual health check to test client
		res, err := client.Cluster.Health(
			client.Cluster.Health.WithPretty(),
		)
		if err != nil {
			return nil, err
		}
		if res.IsError() {
			return nil, fmt.Errorf("health check failed with status code: %d", res.StatusCode)
		}
		return &ESClientV7{client: client}, nil
	}

	return nil, fmt.Errorf("unknown database verseion: %s", db.Spec.Version)
}
