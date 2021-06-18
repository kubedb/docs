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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	_ "github.com/go-sql-driver/mysql"
	sql_driver "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	TLSValueCustom     = "custom"
	TLSValueSkipVerify = "skip-verify"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckMySQLHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckMariaDBHealth(stopCh)
}

func (c *Controller) CheckMariaDBHealth(stopCh <-chan struct{}) {
	klog.Info("Starting MariaDB health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down MariaDB health checker...")
			break
		default:
			c.CheckMariaDBHealthOnce()
			time.Sleep(api.HealthCheckInterval)
		}
	}
}

func (c *Controller) CheckMariaDBHealthOnce() {
	dbList, err := c.mdLister.MariaDBs(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list MariaDB objects with: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	for idx := range dbList {
		db := dbList[idx]

		// If the DB object is deleted, no need to perform health check.
		if db.DeletionTimestamp != nil || db.Spec.Halted {
			continue
		}

		wg.Add(1)
		go func(db *api.MariaDB) {
			defer func() {
				wg.Done()
			}()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()

			// 1st insure all the pods are going to join the cluster(offline/online) to form a group replication
			// then check if the db is going to accepting connection and in ready state.

			// verifying all pods are going Online

			podList, err := c.Client.CoreV1().Pods(db.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: labels.Set(db.OffshootSelectors()).String(),
			})

			if err != nil {
				klog.Warning("failed to list DB pod with ", err.Error())
				return
			}
			for _, pod := range podList.Items {
				if core_util.IsPodConditionTrue(pod.Status.Conditions, core_util.PodConditionTypeReady) {
					continue
				}
				engine, err := c.getMariaDBClient(ctx, db, getMariaDBHostDNS(db, pod.ObjectMeta), api.MySQLDatabasePort)

				if err != nil {
					klog.Warning("failed to get db client for host ", pod.Namespace, "/", pod.Name)
					return
				}
				func(engine *xorm.Engine) {
					defer func() {
						if engine != nil {
							err = engine.Close()
							if err != nil {
								klog.Errorf("can't close the engine. error: %v,", err)
							}
						}
					}()
					isHostOnline, err := c.isHostOnline(db, engine)
					if err != nil {
						klog.Warning("host is not online ", err.Error())
					}
					// update pod status if specific host get online
					if isHostOnline {
						pod.Status.Conditions = core_util.SetPodCondition(pod.Status.Conditions, core.PodCondition{
							Type:               core_util.PodConditionTypeReady,
							Status:             core.ConditionTrue,
							LastTransitionTime: metav1.Now(),
							Reason:             "DBConditionTypeReadyAndServerOnline",
							Message:            "DB is ready because of server getting Online and Running state",
						})
						_, err = c.Client.CoreV1().Pods(pod.Namespace).UpdateStatus(ctx, &pod, metav1.UpdateOptions{})
						if err != nil {
							klog.Warning("failed to update pod status with: ", err.Error())
						}
					}
				}(engine)
			}
			// Create database client
			port, err := c.GetPrimaryServicePort(db)
			if err != nil {
				klog.Warning("Failed to primary service port with: ", err.Error())
				return
			}

			engine, err := c.getMariaDBClient(ctx, db, db.PrimaryServiceDNS(), port)
			if err != nil {
				// Since the client was unable to connect the database,
				// update "AcceptingConnection" to "false".
				// update "Ready" to "false"
				_, err = util.UpdateMariaDBStatus(
					ctx,
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.MariaDBStatus) (types.UID, *api.MariaDBStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseAcceptingConnection,
								Status:             core.ConditionFalse,
								Reason:             api.DatabaseNotAcceptingConnectionRequest,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("The MariaDB: %s/%s is not accepting client requests, reason: %s", db.Namespace, db.Name, err.Error()),
							})
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseReady,
								Status:             core.ConditionFalse,
								Reason:             api.ReadinessCheckFailed,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("The MariaDB: %s/%s is not ready.", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					klog.Errorf("failed to update status for MariaDB: %s/%s", db.Namespace, db.Name)
				}
				// Since the client isn't created, skip rest operations.
				return
			}

			defer func() {
				if engine != nil {
					err = engine.Close()
					if err != nil {
						klog.Errorf("Can't close the engine. error: %v", err)
					}
				}
			}()

			// While creating the client, we perform a health check along with it.
			// If the client is created without any error,
			// the database is accepting connection.
			// Update "AcceptingConnection" to "true".
			_, err = util.UpdateMariaDBStatus(
				ctx,
				c.DBClient.KubedbV1alpha2(),
				db.ObjectMeta,
				func(in *api.MariaDBStatus) (types.UID, *api.MariaDBStatus) {
					in.Conditions = kmapi.SetCondition(in.Conditions,
						kmapi.Condition{
							Type:               api.DatabaseAcceptingConnection,
							Status:             core.ConditionTrue,
							Reason:             api.DatabaseAcceptingConnectionRequest,
							ObservedGeneration: db.Generation,
							Message:            fmt.Sprintf("The MariaDB: %s/%s is accepting client requests.", db.Namespace, db.Name),
						})
					return db.UID, in
				},
				metav1.UpdateOptions{},
			)
			if err != nil {
				klog.Errorf("Failed to update status for MariaDB: %s/%s", db.Namespace, db.Name)
				// Since condition update failed, skip remaining operations.
				return
			}
			// check MariaDB database health
			if db.IsCluster() {
				err = c.checkMariaDBClusterHealth(db, engine)
				if err != nil {
					klog.Errorf("MariaDB Cluster %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
					return
				}
			} else {
				err = c.checkMariaDBStandaloneHealth(engine)
				if err != nil {
					klog.Errorf("MariaDB standalone %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
					return
				}
			}

			// database is healthy. So update to "Ready" condition to "true"
			_, err = util.UpdateMariaDBStatus(
				ctx,
				c.DBClient.KubedbV1alpha2(),
				db.ObjectMeta,
				func(in *api.MariaDBStatus) (types.UID, *api.MariaDBStatus) {
					in.Conditions = kmapi.SetCondition(in.Conditions,
						kmapi.Condition{
							Type:               api.DatabaseReady,
							Status:             core.ConditionTrue,
							Reason:             api.ReadinessCheckSucceeded,
							ObservedGeneration: db.Generation,
							Message:            fmt.Sprintf("The MariaDB: %s/%s is ready.", db.Namespace, db.Name),
						})
					return db.UID, in
				},
				metav1.UpdateOptions{},
			)
			if err != nil {
				klog.Errorf("Failed to update status for MariaDB: %s/%s", db.Namespace, db.Name)
				return
			}
		}(db)
	}

	wg.Wait()

}

func (c *Controller) checkMariaDBClusterHealth(db *api.MariaDB, engine *xorm.Engine) error {
	// sql queries for checking cluster healthiness

	// 1. check all nodes are ONLINE
	_, err := engine.QueryString("SELECT 1;")
	if err != nil {
		return err
	}

	// 2. check all nodes are ONLINE
	result, err := engine.QueryString("SHOW STATUS LIKE 'wsrep_cluster_size';")
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("wsrep_cluster_size, query result is nil")
	}
	dbStatus, ok := result[0]["Value"]
	if !ok {
		return fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, strconv.Itoa(int(*db.Spec.Replicas))) != 0 {
		return fmt.Errorf("all group member are not online yet")
	}

	// 3. Internal Galera Cluster FSM state number.
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_local_state_comment';")
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("wsrep_local_state_comment, query result is nil")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "Synced") != 0 {
		return fmt.Errorf("all group member are not synced")
	}

	// 4. Shows the internal state of the EVS Protocol.
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_evs_state';")
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("wsrep_evs_state, query result is nil")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "OPERATIONAL") != 0 {
		return fmt.Errorf("wsrep_evs_state is not operational")
	}

	// 5.  Cluster component status. Possible values are PRIMARY (primary group configuration, quorum present),
	// NON_PRIMARY (non-primary group configuration, quorum lost) or DISCONNECTED (not connected to group, retrying).
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_cluster_status';")
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("wsrep_cluster_status, query result is nil")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "Primary") != 0 {
		return fmt.Errorf("cluster component status is not primary")
	}

	// 6. Whether or not MariaDB is connected to the wsrep provider. Possible values are ON or OFF.
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_connected';")
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("wsrep_connected, query result is nil")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "ON") != 0 {
		return fmt.Errorf("all group member are not connceted")
	}

	// 7. Whether or not the Galera wsrep provider is ready. Possible values are ON or OFF
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_ready';")
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("wsrep_ready, query result is nil")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "ON") != 0 {
		return fmt.Errorf("galera wsrep provider is not ready")
	}

	return nil
}

func (c *Controller) checkMariaDBStandaloneHealth(engine *xorm.Engine) error {
	// sql queries for checking standalone healthiness
	// 1. ping database
	_, err := engine.QueryString("SELECT 1;")
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) isHostOnline(db *api.MariaDB, engine *xorm.Engine) (bool, error) {
	// 1. ping database
	_, err := engine.QueryString("SELECT 1;")
	if err != nil {
		return false, fmt.Errorf("can't ping to mariadb server, reason: %v", err)
	}

	if !db.IsCluster() {
		return true, nil
	}

	// 2. wsrep_local_state
	result, err := engine.QueryString("SHOW STATUS LIKE 'wsrep_local_state';")
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("empty result on query: \"SHOW STATUS LIKE 'wsrep_local_state';\"")
	}
	dbStatus, ok := result[0]["Value"]
	if !ok {
		return false, fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "2") != 0 && strings.Compare(dbStatus, "4") != 0 {
		return false, fmt.Errorf("expected 2 or 4 in wsrep_local_state, got: %s", dbStatus)
	}

	// 3. auto-eviction - https://galeracluster.com/library/documentation/auto-eviction.html
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_evs_state';")
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("empty result on query: \"SHOW STATUS LIKE 'wsrep_evs_state';\"")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return false, fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "OPERATIONAL") != 0 {
		return false, fmt.Errorf("expected OPERATIONAL in wsrep_evs_state, got: %s", dbStatus)
	}

	// 4. wsrep_connected
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_connected';")
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("empty result on query: \"SHOW STATUS LIKE 'wsrep_connected';\"")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return false, fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "ON") != 0 {
		return false, fmt.Errorf("expected ON in wsrep_connected, got: %s", dbStatus)
	}

	//5. wsrep_ready
	result, err = engine.QueryString("SHOW STATUS LIKE 'wsrep_ready';")
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("empty result on query: \"SHOW STATUS LIKE 'wsrep_ready';\"")
	}
	dbStatus, ok = result[0]["Value"]
	if !ok {
		return false, fmt.Errorf("can not read status from QueryString map")
	}
	if strings.Compare(dbStatus, "ON") != 0 {
		return false, fmt.Errorf("expected ON in wsrep_ready, got: %s", dbStatus)
	}
	return true, nil

}

func (c *Controller) getMariaDBClient(ctx context.Context, db *api.MariaDB, dns string, port int32) (*xorm.Engine, error) {
	user, pass, err := c.getMariaDBRootCredential(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("DB basic auth is not found for MariaDB %v/%v", db.Namespace, db.Name)
	}

	tlsConfig := ""
	if db.Spec.TLS != nil {
		serverSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, db.GetCertSecretName(api.MariaDBServerCert), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		caCert := serverSecret.Data["ca.crt"]
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCert)

		// tls custom setup
		if db.Spec.RequireSSL {
			err = sql_driver.RegisterTLSConfig(TLSValueCustom, &tls.Config{
				RootCAs: certPool,
			})
			if err != nil {
				return nil, err
			}
			tlsConfig = fmt.Sprintf("tls=%s", TLSValueCustom)
		} else {
			tlsConfig = fmt.Sprintf("tls=%s", TLSValueSkipVerify)
		}
	}

	cnnstr := fmt.Sprintf("%v:%v@tcp(%s:%d)/%s?%s", user, pass, dns, port, "mysql", tlsConfig)
	engine, err := xorm.NewEngine("mysql", cnnstr)
	if err != nil {
		return nil, fmt.Errorf("failed to create xorm engine")
	}
	engine.SetDefaultContext(ctx)
	return engine, nil
}

func (c *Controller) getMariaDBRootCredential(ctx context.Context, db *api.MariaDB) (string, string, error) {
	var secretName string
	if db.Spec.AuthSecret != nil {
		secretName = db.GetAuthSecretName()
	}
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	user, ok := secret.Data[core.BasicAuthUsernameKey]
	if !ok {
		return "", "", fmt.Errorf("DB root user is not set")
	}
	pass, ok := secret.Data[core.BasicAuthPasswordKey]
	if !ok {
		return "", "", fmt.Errorf("DB root password is not set")
	}
	return string(user), string(pass), nil
}

func getMariaDBHostDNS(db *api.MariaDB, podMeta metav1.ObjectMeta) string {
	return fmt.Sprintf("%v.%v.%v.svc", podMeta.Name, db.GoverningServiceName(), podMeta.Namespace)
}
