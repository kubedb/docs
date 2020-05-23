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
	"context"
	"fmt"
	"strings"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"gomodules.xyz/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) checkTLS(mongodb *api.MongoDB) error {
	if mongodb.Spec.TLS == nil {
		return nil
	}

	if mongodb.Spec.ReplicaSet == nil && mongodb.Spec.ShardTopology == nil {
		_, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), mongodb.Name+api.MongoDBServerSecretSuffix, metav1.GetOptions{})
		if err != nil {
			return err
		}
	} else if mongodb.Spec.ReplicaSet != nil && mongodb.Spec.ShardTopology == nil {
		// ReplicaSet
		for i := 0; i < int(*mongodb.Spec.Replicas); i++ {
			_, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), fmt.Sprintf("%v-%d", mongodb.Name, i), metav1.GetOptions{})
			if err != nil {
				return err
			}
		}
		return nil
	} else if mongodb.Spec.ShardTopology != nil {
		// for config server
		for i := 0; i < int(mongodb.Spec.ShardTopology.ConfigServer.Replicas); i++ {
			_, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), fmt.Sprintf("%v-%d", mongodb.ConfigSvrNodeName(), i), metav1.GetOptions{})
			if err != nil {
				return err
			}
		}

		//for shards
		for i := 0; i < int(mongodb.Spec.ShardTopology.Shard.Shards); i++ {
			shardName := mongodb.ShardNodeName(int32(i))
			for j := 0; j < int(mongodb.Spec.ShardTopology.Shard.Replicas); j++ {
				_, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), fmt.Sprintf("%v-%d", shardName, j), metav1.GetOptions{})
				if err != nil {
					return err
				}
			}
		}
		//for mongos
		for i := 0; i < int(mongodb.Spec.ShardTopology.Mongos.Replicas); i++ {
			_, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), fmt.Sprintf("%v-%d", mongodb.MongosNodeName(), i), metav1.GetOptions{})
			if err != nil {
				return err
			}
		}
	}
	// for stash/user
	_, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), mongodb.Name+api.MongoDBExternalClientSecretSuffix+api.MongoDBPEMSecretSuffix, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// for prometheus exporter
	_, err = c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), mongodb.Name+api.MongoDBExporterClientSecretSuffix, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) getTLSArgs(mongoDB *api.MongoDB, mgVersion *v1alpha1.MongoDBVersion) ([]string, error) {
	var sslArgs []string
	sslMode := string(mongoDB.Spec.SSLMode)
	breakingVer, err := version.NewVersion("4.2")
	if err != nil {
		return nil, err
	}
	currentVer, err := version.NewVersion(mgVersion.Spec.Version)
	if err != nil {
		return nil, err
	}

	//xREF: https://github.com/docker-library/mongo/issues/367
	if currentVer.GreaterThanOrEqual(breakingVer) {
		var tlsMode = sslMode
		if strings.Contains(sslMode, "SSL") {
			tlsMode = strings.Replace(sslMode, "SSL", "TLS", 1)
		} //ie. requireSSL => requireTLS

		sslArgs = []string{
			fmt.Sprintf("--tlsMode=%v", tlsMode),
		}

		if mongoDB.Spec.SSLMode != api.SSLModeDisabled {
			//xREF: https://github.com/docker-library/mongo/issues/367
			sslArgs = append(sslArgs, []string{
				fmt.Sprintf("--tlsCAFile=%v/%v", api.MongoCertDirectory, api.TLSCACertFileName),
				fmt.Sprintf("--tlsCertificateKeyFile=%v/%v", api.MongoCertDirectory, api.MongoPemFileName),
			}...)
		}
	} else {
		sslArgs = []string{
			fmt.Sprintf("--sslMode=%v", sslMode),
		}
		if mongoDB.Spec.SSLMode != api.SSLModeDisabled {
			sslArgs = append(sslArgs, []string{
				fmt.Sprintf("--sslCAFile=%v/%v", api.MongoCertDirectory, api.TLSCACertFileName),
				fmt.Sprintf("--sslPEMKeyFile=%v/%v", api.MongoCertDirectory, api.MongoPemFileName),
			}...)
		}
	}

	return sslArgs, nil
}
