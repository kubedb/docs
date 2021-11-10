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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/db-client-go/mongodb"

	"github.com/divideandconquer/go-merge/merge"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (c *Reconciler) CreateTLSUsers(db *api.MongoDB) error {
	secretName := db.GetCertSecretName(api.MongoDBClientCert, "")
	certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		klog.Error(err, "failed to get certificate secret", "Secret", secretName)
		return err
	}

	blk, _ := pem.Decode(certSecret.Data[core.TLSCertKey])

	clientCert, err := x509.ParseCertificate(blk.Bytes)
	if err != nil {
		klog.Error(err, "failed to get certificate secret", "Secret", secretName)
		return err
	}
	tlsUserName := clientCert.Subject.String()

	if db.Spec.ShardTopology != nil {
		err = c.CreateTLSUser(db, strings.Join(db.MongosHosts(), ","), "", tlsUserName)
		if err != nil {
			return err
		}

		for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
			err = c.CreateTLSUser(db, strings.Join(db.ShardHosts(i), ","), db.ShardRepSetName(i), tlsUserName)
			if err != nil {
				return err
			}
		}
	} else {
		err = c.CreateTLSUser(db, strings.Join(db.Hosts(), ","), db.RepSetName(), tlsUserName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Reconciler) CreateTLSUser(db *api.MongoDB, url, repSetName, tlsUserName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	dbClient, err := mongodb.NewKubeDBClientBuilder(db, c.Client).
		WithContext(ctx).
		WithURL(url).
		WithReplSet(repSetName).
		GetMongoClient()
	if err != nil {
		return err
	}
	defer func() {
		dbClient.Close()
	}()

	res := make(map[string]interface{})
	err = dbClient.Database("$external").RunCommand(context.Background(), bson.D{{Key: "usersInfo", Value: tlsUserName}}).Decode(&res)
	if err != nil {
		klog.Error("Failed to get user info. error: ", err)
		return err
	}
	users, ok := res["users"].(primitive.A)
	if ok && len(users) == 0 {
		klog.Info("Creating TLS user with name ", tlsUserName)
		err = dbClient.Database("$external").RunCommand(context.Background(),
			bson.D{
				{
					Key:   "createUser",
					Value: tlsUserName,
				},
				{
					Key: "roles",
					Value: bson.A{
						bson.D{
							{Key: "role", Value: "root"},
							{Key: "db", Value: "admin"},
						},
					},
				},
			}).Decode(&res)
		if err != nil {
			klog.Error("Failed to create tls user. error: ", err)
			return err
		}
	}

	return nil
}

func (c *Reconciler) SetupReplicaSetsConfig(db *api.MongoDB) error {
	if db.Spec.ShardTopology != nil {
		err := c.SetupReplicaSetConfig(db, strings.Join(db.ConfigSvrHosts(), ","), db.ConfigSvrRepSetName(), db.Spec.ShardTopology.ConfigServer.ConfigSecret)
		if err != nil {
			return err
		}

		for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
			err = c.SetupReplicaSetConfig(db, strings.Join(db.ShardHosts(i), ","), db.ShardRepSetName(i), db.Spec.ShardTopology.Shard.ConfigSecret)
			if err != nil {
				return err
			}
		}
	} else if db.Spec.ReplicaSet != nil {
		err := c.SetupReplicaSetConfig(db, strings.Join(db.Hosts(), ","), db.RepSetName(), db.Spec.ConfigSecret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Reconciler) SetupReplicaSetConfig(db *api.MongoDB, url, repSetName string, configSecretRef *core.LocalObjectReference) error {
	if configSecretRef == nil {
		return nil
	}

	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), configSecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	providedConfigByte, ok := secret.Data["replicaset.json"]
	if !ok {
		return nil
	}

	providedConfig := make(map[string]interface{})
	err = json.Unmarshal(providedConfigByte, &providedConfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	dbClient, err := mongodb.NewKubeDBClientBuilder(db, c.Client).
		WithContext(ctx).
		WithReplSet(repSetName).
		WithURL(url).
		GetMongoClient()
	if err != nil {
		return err
	}
	defer func() {
		dbClient.Close()
	}()

	info := make(map[string]interface{})
	err = dbClient.Database("admin").RunCommand(context.Background(), bson.D{{Key: "replSetGetConfig", Value: 1.0}}).Decode(&info)
	if err != nil {
		return err
	}
	if val, ok := info["ok"]; ok && val != 1.0 {
		return fmt.Errorf("failed to get replset config. err: %v", info["errmsg"])
	}

	currentConfig := info["config"].(map[string]interface{})
	config := merge.Merge(currentConfig, providedConfig).(map[string]interface{})
	config["version"] = config["version"].(int32) + 1

	info = make(map[string]interface{})
	err = dbClient.Database("admin").RunCommand(context.Background(), bson.D{{Key: "replSetReconfig", Value: config}}).Decode(&info)
	if err != nil {
		return err
	}

	if val, ok := info["ok"]; ok && val == 1.0 {
		return nil
	}

	return fmt.Errorf("failed to run reconfig for mongodb database %s/%s, response: %v", db.Namespace, db.Name, info)
}
