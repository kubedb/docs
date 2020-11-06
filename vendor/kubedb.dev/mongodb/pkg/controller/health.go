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
	"errors"
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/mongo"
	mgoptions "go.mongodb.org/mongo-driver/mongo/options"
	"gomodules.xyz/x/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/tools/certholder"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckMongoDBHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckMongoDBHealth(stopCh)
}

func (c *Controller) CheckMongoDBHealth(stopCh <-chan struct{}) {
	glog.Info("Starting MongoDB health checker...")

	go wait.Until(func() {
		dbList, err := c.mgLister.MongoDBs(core.NamespaceAll).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to list MongoDB objects with: %s", err.Error())
			return
		}

		for _, db := range dbList {
			var err error
			var dbClient, configSvrClient, mongosClient *mongo.Client
			shardClient := make([]*mongo.Client, 0)
			// Create database client
			if db.Spec.ShardTopology == nil {
				dbClient, err = c.GetMongoClient(db, strings.Join(db.Hosts(), ","))
				if err != nil {
					// Since the client was unable to connect the database,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnections(db, err)
					// Since the client isn't created, skip rest operations.
					continue
				}
			} else {
				configSvrClient, err = c.GetMongoClient(db, strings.Join(db.ConfigSvrHosts(), ","))
				if err != nil {
					// Since the client was unable to connect to the config server,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnections(db, err)
					// Since the client isn't created, skip rest operations.
					continue
				}

				cont := false
				shardClient = make([]*mongo.Client, db.Spec.ShardTopology.Shard.Shards)
				for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
					shardClient[i], err = c.GetMongoClient(db, strings.Join(db.ShardHosts(i), ","))
					if err != nil {
						// Since the client was unable to connect to the shard nodes,
						// update "AcceptingConnection" to "false".
						// update "Ready" to "false"
						c.updateErrorAcceptingConnections(db, err)
						// Since the client isn't created, skip rest operations.
						cont = true
						break
					}
				}

				if cont {
					continue
				}

				mongosClient, err = c.GetMongoClient(db, strings.Join(db.MongosHosts(), ","))
				if err != nil {
					// Since the client was unable to connect to the mongos,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnections(db, err)
					// Since the client isn't created, skip rest operations.
					continue
				}
			}

			// While creating the client, we perform a health check along with it.
			// If the client is created without any error,
			// the database is accepting connection.
			// Update "AcceptingConnection" to "true".
			_, err = util.UpdateMongoDBStatus(
				context.TODO(),
				c.DBClient.KubedbV1alpha2(),
				db.ObjectMeta,
				func(in *api.MongoDBStatus) (types.UID, *api.MongoDBStatus) {
					in.Conditions = kmapi.SetCondition(in.Conditions,
						kmapi.Condition{
							Type:               api.DatabaseAcceptingConnection,
							Status:             core.ConditionTrue,
							Reason:             api.DatabaseAcceptingConnectionRequest,
							ObservedGeneration: db.Generation,
							Message:            fmt.Sprintf("The MongoDB: %s/%s is accepting client requests.", db.Namespace, db.Name),
						})
					return db.UID, in
				},
				metav1.UpdateOptions{},
			)
			if err != nil {
				glog.Errorf("Failed to update status for MongoDB: %s/%s", db.Namespace, db.Name)
				// Since condition update failed, skip remaining operations.
				continue
			}

			if db.Spec.ShardTopology == nil {
				// Update to "Ready" condition to "true" only if the database ping is successful.
				err = dbClient.Ping(context.TODO(), nil)
				if err != nil {
					glog.Errorf("Failed to ping database for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since the get status failed, skip remaining operations.
					continue
				}

				c.updateDatabaseReady(db)
			} else {
				// Update to "Ready" condition to "true" only if the config server and shard ping is successful.
				err = configSvrClient.Ping(context.TODO(), nil)
				if err != nil {
					glog.Errorf("Failed to ping config server for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since the get status failed, skip remaining operations.
					continue
				}

				cont := false
				for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
					err = shardClient[i].Ping(context.TODO(), nil)
					if err != nil {
						glog.Errorf("Failed to ping shard%d for MongoDB: %s/%s with: %s", i, db.Namespace, db.Name, err.Error())
						// Since the get status failed, skip remaining operations.
						cont = true
						break
					}
				}

				if cont {
					continue
				}

				err = mongosClient.Ping(context.TODO(), nil)
				if err != nil {
					glog.Errorf("Failed to ping mongos for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since the get status failed, skip remaining operations.
					continue
				}

				c.updateDatabaseReady(db)
			}
		}
	}, c.ReadinessProbeInterval, stopCh)

	// will wait here until stopCh is closed.
	<-stopCh
	glog.Info("Shutting down MongoDB health checker...")
}

func (c *Controller) GetMongoClient(db *api.MongoDB, url string) (*mongo.Client, error) {
	clientOpts, err := c.GetMongoDBClientOpts(db, url)
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Controller) GetURL(db *api.MongoDB, clientPodName string) string {
	nodeType := clientPodName[:strings.LastIndex(clientPodName, "-")]
	return fmt.Sprintf("%s.%s.%s.svc", clientPodName, db.GoverningServiceName(nodeType), db.Namespace)
}

func (c *Controller) GetMongoDBClientOpts(db *api.MongoDB, url string, isReplSet ...bool) (*mgoptions.ClientOptions, error) {
	var clientOpts *mgoptions.ClientOptions
	if db.Spec.SSLMode == api.SSLModeRequireSSL {
		secretName := db.MustCertSecretName(api.MongoDBClientCert, "")
		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			log.Error(err, "failed to get certificate secret", "Secret", secretName)
			return nil, err
		}

		certs, _ := certholder.DefaultHolder.
			ForResource(api.SchemeGroupVersion.WithResource(api.ResourcePluralMongoDB), db.ObjectMeta)
		_, err = certs.Save(certSecret)
		if err != nil {
			log.Error(err, "failed to save certificate")
			return nil, err
		}

		paths, err := certs.Get(db.MustCertSecretName(api.MongoDBClientCert, ""))
		if err != nil {
			return nil, err
		}

		uri := fmt.Sprintf("mongodb://%s/admin?tls=true&authMechanism=MONGODB-X509&tlsCAFile=%v&tlsCertificateKeyFile=%v", url, paths.CACert, paths.Pem)
		clientOpts = mgoptions.Client().ApplyURI(uri)
	} else {
		user, pass, err := c.GetMongoDBRootCredentials(db)
		if err != nil {
			return nil, err
		}
		clientOpts = mgoptions.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s", user, pass, url))
	}

	clientOpts.SetDirect(true)

	return clientOpts, nil
}

func (c *Controller) GetMongoDBRootCredentials(db *api.MongoDB) (string, string, error) {
	if db.Spec.AuthSecret == nil {
		return "", "", errors.New("no database secret")
	}
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return string(secret.Data[core.BasicAuthUsernameKey]), string(secret.Data[core.BasicAuthPasswordKey]), nil
}

func (c *Controller) updateErrorAcceptingConnections(db *api.MongoDB, connectionErr error) {
	_, err := util.UpdateMongoDBStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.MongoDBStatus) (types.UID, *api.MongoDBStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseAcceptingConnection,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseNotAcceptingConnectionRequest,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The MongoDB: %s/%s is not accepting client requests. error: %s", db.Namespace, db.Name, connectionErr),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The MongoDB: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		glog.Errorf("Failed to update status for MongoDB: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateDatabaseReady(db *api.MongoDB) {
	_, err := util.UpdateMongoDBStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.MongoDBStatus) (types.UID, *api.MongoDBStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionTrue,
					Reason:             api.ReadinessCheckSucceeded,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The MongoDB: %s/%s is ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		glog.Errorf("Failed to update status for MongoDB: %s/%s", db.Namespace, db.Name)
	}
}
