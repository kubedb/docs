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
	"net"
	"net/http"
	"time"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/Masterminds/semver/v3"
	esv5 "github.com/elastic/go-elasticsearch/v5"
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

func GetElasticClient(kc kubernetes.Interface, db *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion, url string) (ESClient, error) {
	if db == nil || esVersion == nil {
		return nil, errors.New("db or esVersion is empty")
	}

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

	// parse version
	version, err := semver.NewVersion(esVersion.Spec.Version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse version")
	}

	switch {
	// For Elasticsearch 5.x.x
	case version.Major() == 5:
		client, err := esv5.NewClient(esv5.Config{
			Addresses: []string{url},
			Username:  username,
			Password:  password,
			Transport: &http.Transport{
				IdleConnTimeout: 3 * time.Second,
				DialContext: (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext,
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
		defer res.Body.Close()

		if res.IsError() {
			return nil, fmt.Errorf("health check failed with status code: %d", res.StatusCode)
		}
		return &ESClientV5{client: client}, nil

	// for Elasticsearch 6.x.x
	case version.Major() == 6:
		client, err := esv6.NewClient(esv6.Config{
			Addresses:         []string{url},
			Username:          username,
			Password:          password,
			EnableDebugLogger: true,
			DisableRetry:      true,
			Transport: &http.Transport{
				IdleConnTimeout: 3 * time.Second,
				DialContext: (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext,
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
		defer res.Body.Close()

		if res.IsError() {
			return nil, fmt.Errorf("health check failed with status code: %d", res.StatusCode)
		}
		return &ESClientV6{client: client}, nil

	// for Elasticsearch 7.x.x and OpenSearch 1.x.x
	case version.Major() == 7 || (esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginOpenSearch && version.Major() == 1):
		client, err := esv7.NewClient(esv7.Config{
			Addresses:         []string{url},
			Username:          username,
			Password:          password,
			EnableDebugLogger: true,
			DisableRetry:      true,
			Transport: &http.Transport{
				IdleConnTimeout: 3 * time.Second,
				DialContext: (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext,
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
		defer res.Body.Close()

		if res.IsError() {
			return nil, fmt.Errorf("health check failed with status code: %d", res.StatusCode)
		}
		return &ESClientV7{client: client}, nil
	}

	return nil, fmt.Errorf("unknown database verseion: %s", db.Spec.Version)
}
