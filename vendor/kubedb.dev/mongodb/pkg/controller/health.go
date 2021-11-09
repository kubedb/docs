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
	"strings"
	"sync"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/db-client-go/mongodb"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckMongoDBHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckMongoDBHealth(stopCh)
}

func (c *Controller) CheckMongoDBHealth(stopCh <-chan struct{}) {
	klog.Info("Starting MongoDB health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down MongoDB health checker...")
			break
		default:
			c.CheckMongoDBHealthOnce()
			time.Sleep(api.HealthCheckInterval)
		}
	}
}

func (c *Controller) CheckMongoDBHealthOnce() {
	dbList, err := c.mgLister.MongoDBs(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list MongoDB objects with: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	for idx := range dbList {
		db := dbList[idx]

		if db.DeletionTimestamp != nil || db.Spec.Halted {
			continue
		}

		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()
			var err error
			var dbClient, configSvrClient, mongosClient *mongodb.Client
			var shardPingErrors []error

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()

			// Create database client
			if db.Spec.ShardTopology == nil {
				dbClient, err = mongodb.NewKubeDBClientBuilder(db, c.Client).
					WithContext(ctx).
					WithURL(strings.Join(db.Hosts(), ",")).
					WithReplSet(db.RepSetName()).
					GetMongoClient()
				if err != nil {
					// Since the client was unable to connect the database,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnections(db, err)
					// Since the client isn't created, skip rest operations.
					return
				}
				defer func() {
					dbClient.Close()
				}()
			} else {
				configSvrClient, err = mongodb.NewKubeDBClientBuilder(db, c.Client).
					WithContext(ctx).
					WithURL(strings.Join(db.ConfigSvrHosts(), ",")).
					WithReplSet(db.ConfigSvrRepSetName()).
					GetMongoClient()
				if err != nil {
					// Since the client was unable to connect to the config server,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnections(db, err)
					// Since the client isn't created, skip rest operations.
					return
				}
				defer func() {
					configSvrClient.Close()
				}()

				shardPingErrors = make([]error, db.Spec.ShardTopology.Shard.Shards)
				for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
					shardClient, err := mongodb.NewKubeDBClientBuilder(db, c.Client).
						WithContext(ctx).
						WithURL(strings.Join(db.ShardHosts(i), ",")).
						WithReplSet(db.ShardRepSetName(i)).
						GetMongoClient()
					if err != nil {
						// Since the client was unable to connect to the shard nodes,
						// update "AcceptingConnection" to "false".
						// update "Ready" to "false"
						c.updateErrorAcceptingConnections(db, err)
						// Since the client isn't created, skip rest operations.
						return
					}
					func(client *mongodb.Client) {
						defer func() {
							client.Close()
						}()
						err = client.Ping(ctx, nil)
						if err != nil {
							shardPingErrors[i] = err
							klog.Errorf("Failed to ping shard%d for MongoDB: %s/%s with: %s", i, db.Namespace, db.Name, err.Error())
							// Since the get status failed, skip remaining operations.
							return
						}
					}(shardClient)
				}

				mongosClient, err = mongodb.NewKubeDBClientBuilder(db, c.Client).
					WithContext(ctx).
					WithURL(strings.Join(db.MongosHosts(), ",")).
					WithReplSet("").
					GetMongoClient()
				if err != nil {
					// Since the client was unable to connect to the mongos,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnections(db, err)
					// Since the client isn't created, skip rest operations.
					return
				}
				defer func() {
					mongosClient.Close()
				}()
			}

			// While creating the client, we perform a health check along with it.
			// If the client is created without any error,
			// the database is accepting connection.
			// Update "AcceptingConnection" to "true".
			_, err = util.UpdateMongoDBStatus(
				ctx,
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
				klog.Errorf("Failed to update status for MongoDB: %s/%s", db.Namespace, db.Name)
				// Since condition update failed, skip remaining operations.
				return
			}

			if db.Spec.ShardTopology == nil {
				// Update to "Ready" condition to "true" only if the database ping is successful.
				err = dbClient.Ping(ctx, nil)
				if err != nil {
					klog.Errorf("Failed to ping database for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since the get status failed, skip remaining operations.
					return
				}

				c.updateDatabaseReady(db)
			} else {
				for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
					if shardPingErrors[i] != nil {
						klog.Errorf("Failed to ping shard%d for MongoDB: %s/%s with: %s", i, db.Namespace, db.Name, shardPingErrors[i].Error())
						// Since the get status failed, skip remaining operations.
						return
					}
				}

				// Update to "Ready" condition to "true" only if the config server and shard ping is successful.
				err = configSvrClient.Ping(ctx, nil)
				if err != nil {
					klog.Errorf("Failed to ping config server for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since the get status failed, skip remaining operations.
					return
				}

				err = mongosClient.Ping(ctx, nil)
				if err != nil {
					klog.Errorf("Failed to ping mongos for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since the get status failed, skip remaining operations.
					return
				}

				c.updateDatabaseReady(db)
			}
		}()
	}

	wg.Wait()
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
		klog.Errorf("Failed to update status for MongoDB: %s/%s", db.Namespace, db.Name)
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
		klog.Errorf("Failed to update status for MongoDB: %s/%s", db.Namespace, db.Name)
	}
}
