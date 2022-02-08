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

	"github.com/Masterminds/semver/v3"
	"github.com/divideandconquer/go-merge/merge"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/exec"
)

func (r *Reconciler) CreateTLSUsers(db *api.MongoDB) error {
	secretName := db.GetCertSecretName(api.MongoDBClientCert, "")
	certSecret, err := r.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
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
		err = r.CreateTLSUser(db, strings.Join(db.MongosHosts(), ","), "", tlsUserName)
		if err != nil {
			return err
		}

		for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
			err = r.CreateTLSUser(db, strings.Join(db.ShardHosts(i), ","), db.ShardRepSetName(i), tlsUserName)
			if err != nil {
				return err
			}
		}
	} else {
		err = r.CreateTLSUser(db, strings.Join(db.Hosts(), ","), db.RepSetName(), tlsUserName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) CreateTLSUser(db *api.MongoDB, url, repSetName, tlsUserName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	dbClient, err := mongodb.NewKubeDBClientBuilder(db, r.Client).
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

func (r *Reconciler) SetupReplicaSetsConfig(db *api.MongoDB) error {
	if db.Spec.ShardTopology != nil {
		err := r.SetupReplicaSetConfig(db, strings.Join(db.ConfigSvrHosts(), ","), db.ConfigSvrRepSetName(), db.Spec.ShardTopology.ConfigServer.ConfigSecret)
		if err != nil {
			return err
		}

		for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
			err = r.SetupReplicaSetConfig(db, strings.Join(db.ShardHosts(i), ","), db.ShardRepSetName(i), db.Spec.ShardTopology.Shard.ConfigSecret)
			if err != nil {
				return err
			}
		}
	} else if db.Spec.ReplicaSet != nil {
		err := r.SetupReplicaSetConfig(db, strings.Join(db.Hosts(), ","), db.RepSetName(), db.Spec.ConfigSecret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) SetupReplicaSetConfig(db *api.MongoDB, url, repSetName string, configSecretRef *core.LocalObjectReference) error {
	if configSecretRef == nil {
		return nil
	}

	secret, err := r.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), configSecretRef.Name, metav1.GetOptions{})
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
	dbClient, err := mongodb.NewKubeDBClientBuilder(db, r.Client).
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

func (c *Reconciler) ApplyConfigJsFiles(db *api.MongoDB) error {
	if db.Spec.ShardTopology != nil {
		if c.hasConfigJs(db, db.Spec.ShardTopology.ConfigServer.ConfigSecret) {
			for i := int32(0); i < db.Spec.ShardTopology.ConfigServer.Replicas; i++ {
				podName := fmt.Sprintf("%s-%d", db.ConfigSvrNodeName(), i)
				isMaster, err := c.isMaster(podName, db)
				if err != nil {
					return err
				}

				if isMaster {
					err := c.ApplyConfigJsFile(db, podName)
					if err != nil {
						return err
					}
					break
				}
			}
		}

		if c.hasConfigJs(db, db.Spec.ShardTopology.Shard.ConfigSecret) {
			for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
				for j := int32(0); j < db.Spec.ShardTopology.Shard.Replicas; j++ {
					podName := fmt.Sprintf("%s-%d", db.ShardNodeName(i), j)
					isMaster, err := c.isMaster(podName, db)
					if err != nil {
						return err
					}

					if isMaster {
						err := c.ApplyConfigJsFile(db, podName)
						if err != nil {
							return err
						}
						break
					}
				}
			}
		}

		if c.hasConfigJs(db, db.Spec.ShardTopology.Mongos.ConfigSecret) {
			mongosApplied := false
			for i := int32(0); i < db.Spec.ShardTopology.Mongos.Replicas; i++ {
				podName := fmt.Sprintf("%s-%d", db.MongosNodeName(), i)
				pod, err := c.Client.CoreV1().Pods(db.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
				if err != nil && !errors.IsNotFound(err) {
					return err
				}
				if pod == nil || pod.Status.Phase != core.PodRunning {
					continue
				}

				err = c.ApplyConfigJsFile(db, podName)
				if err != nil {
					return err
				}
				mongosApplied = true
			}
			if !mongosApplied {
				return fmt.Errorf("failed to apply configuration.js in mongos, reason: no active mongos found")
			}
		}
	} else if db.Spec.ReplicaSet != nil {
		if c.hasConfigJs(db, db.Spec.ConfigSecret) {
			for i := int32(0); i < *db.Spec.Replicas; i++ {
				podName := fmt.Sprintf("%s-%d", db.OffshootName(), i)
				isMaster, err := c.isMaster(podName, db)
				if err != nil {
					return err
				}

				if isMaster {
					err := c.ApplyConfigJsFile(db, podName)
					if err != nil {
						return err
					}
					break
				}
			}
		}
	} else {
		if c.hasConfigJs(db, db.Spec.ConfigSecret) {
			podName := fmt.Sprintf("%s-%d", db.OffshootName(), 0)

			err := c.ApplyConfigJsFile(db, podName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Reconciler) ApplyConfigJsFile(db *api.MongoDB, podName string) error {
	pod, err := c.Client.CoreV1().Pods(db.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cmd, err := c.getCmdForConfigJs(db)
	if err != nil {
		return err
	}

	out, err := exec.ExecIntoPod(c.ClientConfig, pod, exec.Command(cmd...))
	if out != "" {
		klog.Infof("configuration.js applied, output: %s", out)
	}
	return err
}

func (c *Reconciler) hasConfigJs(db *api.MongoDB, configSecretRef *core.LocalObjectReference) bool {
	if configSecretRef == nil {
		return false
	}

	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), configSecretRef.Name, metav1.GetOptions{})
	if err != nil {
		klog.Infof("failed to get secret %s/%s", db.Namespace, configSecretRef.Name)
		return false
	}

	_, ok := secret.Data[api.MongoDBConfigurationJSFile]
	return ok
}

func (c *Reconciler) getCmdForConfigJs(db *api.MongoDB) ([]string, error) {
	mgVersion, err := c.DBClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), db.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var sslArgs string
	if db.Spec.SSLMode == api.SSLModeRequireSSL {
		sslArgs = fmt.Sprintf("--tls --tlsCAFile=%v/%v --tlsCertificateKeyFile=%v/%v",
			api.MongoCertDirectory, api.TLSCACertFileName, api.MongoCertDirectory, api.MongoClientFileName)

		breakingVer, _ := semver.NewVersion("4.1")
		exceptionVer, _ := semver.NewVersion("4.1.4")
		currentVer, err := semver.NewVersion(mgVersion.Spec.Version)
		if err != nil {
			return nil, fmt.Errorf("MongoDB %s/%s: unable to parse version. reason: %s", db.Namespace, db.Name, err.Error())
		}
		if currentVer.Equal(exceptionVer) {
			sslArgs = fmt.Sprintf("--tls --tlsCAFile=%v/%v --tlsPEMKeyFile=%v/%v", api.MongoCertDirectory, api.TLSCACertFileName, api.MongoCertDirectory, api.MongoClientFileName)
		} else if currentVer.LessThan(breakingVer) {
			sslArgs = fmt.Sprintf("--ssl --sslCAFile=%v/%v --sslPEMKeyFile=%v/%v", api.MongoCertDirectory, api.TLSCACertFileName, api.MongoCertDirectory, api.MongoClientFileName)
		}
	}

	return []string{
		"bash",
		"-c",
		fmt.Sprintf(`mongo admin --host=localhost %v --username=$MONGO_INITDB_ROOT_USERNAME --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin --quiet %v`, sslArgs, api.MongoDBConfigDirectoryPath+"/"+api.MongoDBConfigurationJSFile),
	}, nil
}

func (c *Reconciler) isMaster(clientPodName string, db *api.MongoDB) (bool, error) {
	client, err := mongodb.NewKubeDBClientBuilder(db, c.Client).
		WithPod(clientPodName).
		GetMongoClient()
	if err != nil {
		return false, err
	}
	defer func() {
		client.Close()
	}()

	res := make(map[string]interface{})

	err = client.Database("admin").RunCommand(context.Background(), bson.D{{Key: "isMaster", Value: "1"}}).Decode(&res)
	if err != nil {
		return false, err
	}

	if val, ok := res["ismaster"]; ok && val == true {
		return true, nil
	}
	return false, nil
}
