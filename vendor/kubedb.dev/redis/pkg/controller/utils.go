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

package controller

import (
	"context"
	"fmt"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/Masterminds/semver/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetTLSArgs(curVersion *semver.Version, standAlone bool, port int) []string {
	tlsArgs := []string{
		fmt.Sprintf("--tls-port %v", port),
		"--port 0",
		"--tls-cert-file /certs/server.crt",
		"--tls-key-file /certs/server.key",
		"--tls-ca-cert-file /certs/ca.crt",
	}
	if !standAlone {
		tlsArgs = append(tlsArgs, "--tls-replication yes")
	}
	if curVersion.Major() >= 6 && curVersion.Minor() >= 2 {
		tlsArgs = append(tlsArgs, "--tls-auth-clients optional")
	} else if curVersion.Major() >= 6 {
		tlsArgs = append(tlsArgs, "--tls-auth-clients no")
	}
	return tlsArgs
}

func (c *Controller) getRedisVersion(db *api.Redis) (*v1alpha1.RedisVersion, *semver.Version, error) {
	redisVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), db.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	curVersion, err := semver.NewVersion(redisVersion.Spec.Version)
	if err != nil {
		return nil, nil, err
	}
	return redisVersion, curVersion, nil
}

func (c *Controller) getRedisSentinelVersion(sentinel *api.RedisSentinel) (*v1alpha1.RedisVersion, *semver.Version, error) {
	redisVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), sentinel.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	curVersion, err := semver.NewVersion(redisVersion.Spec.Version)
	if err != nil {
		return nil, nil, err
	}
	return redisVersion, curVersion, nil
}

func GetRdClusterRegisteredNameInSentinel(db *api.Redis) string {
	return fmt.Sprintf("%s/%s", db.Namespace, db.Name)
}
