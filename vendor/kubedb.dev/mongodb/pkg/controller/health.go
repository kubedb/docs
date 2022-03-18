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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
			var shardErrors []error

			ctx, cancel := context.WithTimeout(context.Background(), api.HealthCheckInterval)
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
					c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to connect to the database client, error: %v", err.Error()))
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
					c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to connect to the configsvr client, error: %v", err.Error()))
					// Since the client isn't created, skip rest operations.
					return
				}
				defer func() {
					configSvrClient.Close()
				}()

				shardErrors = make([]error, db.Spec.ShardTopology.Shard.Shards)
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
						c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to connect to the shard-%d client, error: %v", i, err.Error()))
						// Since the client isn't created, skip rest operations.
						return
					}
					func(client *mongodb.Client) {
						defer func() {
							client.Close()
						}()

						err = checkReadWrite(ctx, client)
						if err != nil {
							shardErrors[i] = err
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
					c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to connect to the mongos client, error: %v", err.Error()))
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
				err = checkReadWrite(ctx, dbClient, true)
				if err != nil {
					klog.Errorf("health check failed for database for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since read/write operations failed, skip remaining operations.
					c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to check read/write of the database, error: %v", err.Error()))
					return
				}

				c.updateDatabaseReady(db)
			} else {
				for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
					if shardErrors[i] != nil {
						klog.Errorf("health check failed for shard%d for MongoDB: %s/%s with: %s", i, db.Namespace, db.Name, err.Error())
						// Since the read/write operations failed, skip remaining operations.
						c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to check read/write of the shard-%d, error: %v", i, err.Error()))
						return
					}
				}

				err = checkReadWrite(ctx, configSvrClient, true)
				if err != nil {
					klog.Errorf("health check failed for config server for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since read/write operations failed, skip remaining operations.
					c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to check read/write of the configsvr, error: %v", err.Error()))
					return
				}

				err = checkReadWrite(ctx, mongosClient)
				if err != nil {
					klog.Errorf("health check failed for mongos for MongoDB: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since read/write operations failed, skip remaining operations.
					c.updateErrorAcceptingConnections(db, fmt.Errorf("unable to check read/write of the mongos, error: %v", err.Error()))
					return
				}

				c.updateDatabaseReady(db)
			}
		}()
	}

	wg.Wait()
}

func (c *Controller) updateErrorAcceptingConnections(db *api.MongoDB, connectionErr error) {
	klog.Errorf("Failed to accept connections for MongoDB: %s/%s", db.Namespace, db.Name)
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

func checkReadWrite(ctx context.Context, client *mongodb.Client, checkPingOnly ...bool) error {
	err := client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping database with error: %s", err.Error())
	}
	if len(checkPingOnly) == 1 && checkPingOnly[0] {
		return nil
	}

	valTrue := true
	_, err = client.Database("kubedb-system").Collection("health-check").UpdateOne(
		ctx,
		bson.M{"id": "1"},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "id", Value: "1"},
				{Key: "health", Value: "Ok"},
			}},
		},
		&options.UpdateOptions{
			Upsert: &valTrue,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to write on database with error: %s", err.Error())
	}

	c, err := client.Database("kubedb-system").Collection("health-check").Find(
		ctx,
		bson.D{
			{Key: "id", Value: "1"},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to read from database with error: %s", err.Error())
	}
	defer c.Close(context.TODO())

	var res bson.A
	err = c.All(context.TODO(), &res)
	if err != nil {
		return fmt.Errorf("failed to read from database with error: %s", err.Error())
	}

	return nil
}
