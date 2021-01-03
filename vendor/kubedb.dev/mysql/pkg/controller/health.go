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
	"strings"
	"sync"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	_ "github.com/go-sql-driver/mysql"
	sql_driver "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/golang/glog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	TLSValueCustom     = "custom"
	TLSValueSkipVerify = "skip-verify"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckMySQLHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckMySQLHealth(stopCh)
}

func (c *Controller) CheckMySQLHealth(stopCh <-chan struct{}) {
	glog.Info("Starting MySQL health checker...")

	go wait.Until(func() {
		dbList, err := c.myLister.MySQLs(core.NamespaceAll).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to list MySQL objects with: %s", err.Error())
			return
		}

		var wg sync.WaitGroup
		for idx := range dbList {
			db := dbList[idx]

			if db.DeletionTimestamp != nil {
				continue
			}
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
				}()
				// Create database client
				engine, err := c.getMySQLClient(db)
				if err != nil {
					// Since the client was unable to connect the database,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					_, err = util.UpdateMySQLStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						db.ObjectMeta,
						func(in *api.MySQLStatus) (types.UID, *api.MySQLStatus) {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:               api.DatabaseAcceptingConnection,
									Status:             core.ConditionFalse,
									Reason:             api.DatabaseNotAcceptingConnectionRequest,
									ObservedGeneration: db.Generation,
									Message:            fmt.Sprintf("The MySQL: %s/%s is not accepting client requests, reason: %s", db.Namespace, db.Name, err.Error()),
								})
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:               api.DatabaseReady,
									Status:             core.ConditionFalse,
									Reason:             api.ReadinessCheckFailed,
									ObservedGeneration: db.Generation,
									Message:            fmt.Sprintf("The MySQL: %s/%s is not ready.", db.Namespace, db.Name),
								})
							return db.UID, in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						glog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					}
					// Since the client isn't created, skip rest operations.
					return
				}

				defer func() {
					if engine != nil {
						err = engine.Close()
						if err != nil {
							glog.Errorf("Can't close the engine. error: %v", err)
						}
					}
				}()

				// While creating the client, we perform a health check along with it.
				// If the client is created without any error,
				// the database is accepting connection.
				// Update "AcceptingConnection" to "true".
				_, err = util.UpdateMySQLStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.MySQLStatus) (types.UID, *api.MySQLStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseAcceptingConnection,
								Status:             core.ConditionTrue,
								Reason:             api.DatabaseAcceptingConnectionRequest,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("The MySQL: %s/%s is accepting client requests.", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					glog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					// Since condition update failed, skip remaining operations.
					return
				}

				// check MySQL database health
				var isHealthy bool
				if *db.Spec.Replicas > int32(1) && db.Spec.Topology != nil && db.Spec.Topology.Group != nil {
					isHealthy, err = c.checkMySQLClusterHealth(db, engine)
					if err != nil {
						glog.Errorf("MySQL Cluster %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
					}
				} else {
					isHealthy, err = c.checkMySQLStandaloneHealth(engine)
					if err != nil {
						glog.Errorf("MySQL standalone %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
					}
				}

				if !isHealthy {
					// Since the get status failed, skip remaining operations.
					return
				}
				// database is healthy. So update to "Ready" condition to "true"
				_, err = util.UpdateMySQLStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.MySQLStatus) (types.UID, *api.MySQLStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseReady,
								Status:             core.ConditionTrue,
								Reason:             api.ReadinessCheckSucceeded,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("The MySQL: %s/%s is ready.", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					glog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
				}

			}()
		}
		wg.Wait()
	}, c.ReadinessProbeInterval, stopCh)

	// will wait here until stopCh is closed.
	<-stopCh
	glog.Info("Shutting down MySQL health checker...")
}

func (c *Controller) checkMySQLClusterHealth(db *api.MySQL, engine *xorm.Engine) (bool, error) {
	// sql queries for checking cluster healthiness
	// 1. ping database
	_, err := engine.QueryString("SELECT 1;")
	if err != nil {
		return false, err
	}

	// 2. check all nodes are in ONLINE
	result, err := engine.QueryString("SELECT MEMBER_STATE FROM performance_schema.replication_group_members;")
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("query result is nil")
	}

	if len(result) != int(*db.Spec.Replicas) {
		return false, fmt.Errorf("not all members have joined in the group yet")
	}

	for j := range result {
		memberState, ok := result[j]["MEMBER_STATE"]
		if !ok || strings.Compare(memberState, "ONLINE") != 0 {
			return false, fmt.Errorf("all group member are not online yet")
		}
	}

	// 2. check replicas data sync with master
	//TODO

	return true, nil
}

func (c *Controller) checkMySQLStandaloneHealth(engine *xorm.Engine) (bool, error) {
	// sql queries for checking standalone healthiness
	// 1. ping database
	_, err := engine.QueryString("SELECT 1;")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Controller) getMySQLClient(db *api.MySQL) (*xorm.Engine, error) {
	port, err := c.GetPrimaryServicePort(db)
	if err != nil {
		return nil, err
	}

	user, pass, err := c.getMySQLBasicAuth(db)
	if err != nil {
		return nil, fmt.Errorf("password basic auth for MySQL %v/%v", db.Namespace, db.Name)
	}
	tlsConfig := ""
	if db.Spec.TLS != nil {
		serverSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.MustCertSecretName(api.MySQLServerCert), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		cacrt := serverSecret.Data["ca.crt"]
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cacrt)

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

	cnnstr := fmt.Sprintf("%v:%v@tcp(%s:%d)/%s?%s", user, pass, db.PrimaryServiceDNS(), port, api.ResourceSingularMySQL, tlsConfig)
	return xorm.NewEngine(api.ResourceSingularMySQL, cnnstr)
}

func (c *Controller) getMySQLBasicAuth(db *api.MySQL) (string, string, error) {
	var secretName string
	if db.Spec.AuthSecret != nil {
		secretName = db.GetAuthSecretName()
	}
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return string(secret.Data[core.BasicAuthUsernameKey]), string(secret.Data[core.BasicAuthPasswordKey]), nil
}
