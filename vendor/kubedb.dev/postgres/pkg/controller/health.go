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
	"sync"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	_ "github.com/lib/pq"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckPostgresDBHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckPostgresDBHealth(stopCh)
}

func (c *Controller) CheckPostgresDBHealth(stopCh <-chan struct{}) {
	klog.Info("Starting Postgres health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down Postgres health checker...")
			break
		default:
			c.CheckPostgresDBHealthOnce()
			time.Sleep(api.HealthCheckInterval)
		}
	}
}

func (c *Controller) CheckPostgresDBHealthOnce() {
	dbList, err := c.pgLister.Postgreses(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list PostgreSQL objects with: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	for idx := range dbList {
		db := dbList[idx]
		// If the DB object is deleted or halted, no need to perform health check.
		if db.DeletionTimestamp != nil || db.Spec.Halted {
			continue
		}

		wg.Add(1)
		go func(db *api.Postgres) {
			defer func() {
				wg.Done()
			}()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			// verify db is going to accepting connection and in ready state
			port, err := c.GetPrimaryServicePort(db)
			if err != nil {
				klog.Warning("Failed to get primary service port with: ", err.Error())
				return
			}
			err = c.IsPostgreSQLServerOnline(ctx, db, PrimaryServiceDNS(db), port)
			if err != nil {
				// Since the client was unable to connect the database,
				// update "AcceptingConnection" to "false".
				// update "Ready" to "false"
				c.updateErrorAcceptingConnections(ctx, db, err)
				// Since the client isn't created, skip rest operations.
				return
			}

			// While creating the client, we perform a health check along with it.
			// If the client is created without any error,
			// the database is accepting connection.
			// Update "AcceptingConnection" to "true".
			_, err = util.UpdatePostgresStatus(
				ctx,
				c.DBClient.KubedbV1alpha2(),
				db.ObjectMeta,
				func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
					in.Conditions = kmapi.SetCondition(in.Conditions,
						kmapi.Condition{
							Type:               api.DatabaseAcceptingConnection,
							Status:             core.ConditionTrue,
							Reason:             api.DatabaseAcceptingConnectionRequest,
							ObservedGeneration: db.Generation,
							Message:            fmt.Sprintf("The PostgreSQL: %s/%s is accepting client requests.", db.Namespace, db.Name),
						})
					return db.UID, in
				},
				metav1.UpdateOptions{},
			)
			if err != nil {
				klog.Errorf("Failed to update status for PostgreSQL: %s/%s", db.Namespace, db.Name)
				// Since condition update failed, skip remaining operations.
				return
			}

			// check PostgreSQL database health
			var isHealthy bool
			if *db.Spec.Replicas > int32(1) {
				isHealthy, err = c.checkPostgreSQLClusterHealth(ctx, db)
				if err != nil {
					klog.Errorf("PostgreSQL Cluster %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
				}
			} else {
				isHealthy, err = c.checkPostgreSQLStandaloneHealth(ctx, db)
				if err != nil {
					klog.Errorf("PostgreSQL standalone %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
				}
			}

			if !isHealthy {
				c.updateDatabaseNotReady(ctx, db, err)
			} else {
				// database is healthy. So update to "Ready" condition to "true"
				c.updateDatabaseReady(ctx, db)
			}

		}(db)
	}
	// Wait until all go-routine complete executions
	wg.Wait()

}

//check if the cluster's every replica is active and in sync with master
func (c *Controller) checkPostgreSQLClusterHealth(ctx context.Context, db *api.Postgres) (bool, error) {

	err := c.IsPostgreSQLServerOnline(ctx, db, PrimaryServiceDNS(db), api.PostgresDatabasePort)
	if err != nil {
		return false, err
	}
	// 2. check all nodes are in ONLINE
	podList, err := c.Client.CoreV1().Pods(db.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(db.OffshootSelectors()).String(),
	})
	if err != nil {
		return false, err
	}

	dbPods, err := api.GetDatabasePods(db, c.StsLister, podList.Items)
	if err != nil {
		return false, fmt.Errorf("failed filter database pods. Reason: %v", err)
	}
	for _, pod := range dbPods {
		err := c.IsPostgreSQLServerOnline(ctx, db, HostDNS(db, pod.ObjectMeta), api.PostgresDatabasePort)

		if err != nil {
			return false, err
		}
	}

	// 3. check replicas data sync with master
	//TODO
	return true, nil
}

// check if the server is ready to accept connections
func (c *Controller) checkPostgreSQLStandaloneHealth(ctx context.Context, db *api.Postgres) (bool, error) {
	err := c.IsPostgreSQLServerOnline(ctx, db, PrimaryServiceDNS(db), api.PostgresDatabasePort)
	if err != nil {
		return false, err
	}
	return true, nil
}

// get user and pass from auth secret
func (c *Controller) GetPostgresAuthCredentials(ctx context.Context, db *api.Postgres) (string, string, error) {
	if db.Spec.AuthSecret == nil {
		return "", "", errors.New("no database secret")
	}
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, db.Spec.AuthSecret.Name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return string(secret.Data[core.BasicAuthUsernameKey]), string(secret.Data[core.BasicAuthPasswordKey]), nil
}

// if the master is not accepting connection then set database ready condition and accepting connection to false
func (c *Controller) updateErrorAcceptingConnections(ctx context.Context, db *api.Postgres, connectionErr error) {
	_, err := util.UpdatePostgresStatus(
		ctx,
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseAcceptingConnection,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseNotAcceptingConnectionRequest,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The PostgreSQL: %s/%s is not accepting client requests. error: %s", db.Namespace, db.Name, connectionErr),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The PostgreSQL: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for PostgreSQL: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateDatabaseReady(ctx context.Context, db *api.Postgres) {
	_, err := util.UpdatePostgresStatus(
		ctx,
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReplicaReady,
					Status:             core.ConditionTrue,
					Reason:             api.AllReplicasAreReady,
					ObservedGeneration: db.Generation,
					Message:            "All replicas are ready and in Running state",
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionTrue,
					Reason:             api.ReadinessCheckSucceeded,
					ObservedGeneration: db.Generation,
					Message:            "DB is ready because of server getting Online and Running state",
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for PostgreSQL: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateDatabaseNotReady(ctx context.Context, db *api.Postgres, errMsg error) {
	_, err := util.UpdatePostgresStatus(
		ctx,
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReplicaReady,
					Status:             core.ConditionFalse,
					Reason:             api.SomeReplicasAreNotReady,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("Some Replicas are not ready. error: %s", errMsg),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The PostgreSQL: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for PostgreSQL: %s/%s", db.Namespace, db.Name)
	}
}

//try to query in server if failed return err that means not online
func (c *Controller) IsPostgreSQLServerOnline(ctx context.Context, db *api.Postgres, dnsName string, port int32) error {
	var err error
	eng, err := c.GetPostgresClient(ctx, db, dnsName, port)

	if err != nil {
		return err
	}
	defer eng.Close()

	queryString := "SELECT now();"

	res, err := eng.QueryString(queryString)
	if err != nil {
		return err
	}

	if len(res[0]["now"]) > 0 {
		return nil
	} else {
		return fmt.Errorf("can't get query value")
	}
}

// make host dns with require template
func HostDNS(db *api.Postgres, podMeta metav1.ObjectMeta) string {
	return fmt.Sprintf("%v.%v.%v.svc", podMeta.Name, db.GoverningServiceName(), podMeta.Namespace)
}

// make primary host dns with require template
func PrimaryServiceDNS(db *api.Postgres) string {
	return fmt.Sprintf("%v.%v.svc", db.ServiceName(), db.Namespace)
}
