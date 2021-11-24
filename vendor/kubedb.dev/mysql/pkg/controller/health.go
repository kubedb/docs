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
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	_ "github.com/go-sql-driver/mysql"
	sql_driver "github.com/go-sql-driver/mysql"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	"xorm.io/xorm"
)

const (
	TLSValueCustom     = "custom"
	TLSValueSkipVerify = "skip-verify"
)

// health checker algorithm

// run health checker after every certain  period of time
// list all databases
// for db in databases
//	-> if halted || has deletion timestamp
//		-> continue
//	->get all pods of db
//		->set pod condition
//			-> ready || not ready?
//	->create engine that can connect to a db server using services
//	-> update AcceptingConnection
//    ->check for cluster health
//	-> if healthy
//		-> update server ready
//	   else
//		-> serverReady->false
//		-> dbReady->false
//	->if innodb cluster
//		-> checkHealth from router services
//			if healthy:
//				serverReady->true
//				dbReady-true
//			else
//				dbReady->false
//				serverReady->false
//	else
//		if healthy && serverReady
//			dbReady->true

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckMySQLHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckMySQLHealth(stopCh)
}

func (c *Controller) CheckMySQLHealth(stopCh <-chan struct{}) {
	klog.Info("Starting MySQL health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down MySQL health checker...")
			break
		default:
			c.CheckMySQLHealthOnce()
			time.Sleep(api.HealthCheckInterval)
		}
	}
}

func (c *Controller) CheckMySQLHealthOnce() {
	dbList, err := c.myLister.MySQLs(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list MySQL objects with: %s", err.Error())
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

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()

			// 1st insure all the pods are going to join the cluster(offline/online) to form a group replication
			// then check if the db is going to accepting connection and in ready state.

			// verifying all pods are going Online
			podList, err := c.Client.CoreV1().Pods(db.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: labels.Set(db.OffshootSelectors()).String(),
			})
			if err != nil {
				klog.Warning("Failed to list DB pod with ", err.Error())
				return
			}

			dbPods, err := api.GetDatabasePods(db, c.StsLister, podList.Items)
			if err != nil {
				klog.Warning("Failed filter database pods. Reason: ", err.Error())
				return
			}
			isHealthy := false
			for _, pod := range dbPods {

				engine, err := c.getMySQLClient(ctx, db, HostDNS(db, pod.ObjectMeta), api.MySQLDatabasePort)
				if err != nil {

					// Since the client was unable to connect the database,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					err := c.updateMySQLStatusConditions(ctx, db,
						kmapi.Condition{
							Type:    api.DatabaseAcceptingConnection,
							Status:  core.ConditionFalse,
							Reason:  api.DatabaseNotAcceptingConnectionRequest,
							Message: fmt.Sprintf("Error while creating mysql client %s", err),
						},
						kmapi.Condition{
							Type:    api.ServerReady,
							Status:  core.ConditionFalse,
							Reason:  api.DatabaseNotAcceptingConnectionRequest,
							Message: fmt.Sprintf("Error while creating mysql client %s", err),
						},
						kmapi.Condition{
							Type:    api.DatabaseReady,
							Status:  core.ConditionFalse,
							Reason:  api.DatabaseNotAcceptingConnectionRequest,
							Message: fmt.Sprintf("Error while creating mysql client %s", err),
						},
					)
					if err != nil {
						klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					}
					// Since the client isn't created, skip rest operations.
					return
				}

				//// While creating the client, we perform a health check along with it.
				//// If the client is created without any error,
				//// the database is accepting connection.
				//// Update "AcceptingConnection" to "true"
				err = c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
					Type:    api.DatabaseAcceptingConnection,
					Status:  core.ConditionTrue,
					Reason:  api.DatabaseAcceptingConnectionRequest,
					Message: fmt.Sprintf("MySQL %s/%s is accepting connection", db.Name, db.Namespace),
				})
				if err != nil {
					klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					// Since condition update failed, skip remaining operations.
					return
				}

				func(engine *xorm.Engine) {
					defer closeClientEngine(engine)
					if *db.Spec.Replicas > int32(1) && db.Spec.Topology != nil && (db.UsesGroupReplication() || db.IsInnoDBCluster()) {
						isHealthy, err = c.checkMySQLClusterHealth(ctx, len(dbPods), engine)
						if err != nil {
							klog.Errorf("MySQL Cluster %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
							err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
								Type:    api.DatabaseReady,
								Status:  core.ConditionFalse,
								Reason:  api.SomeReplicasAreNotReady,
								Message: fmt.Sprintf("MySQL %s/%s not all the replicas are joined in cluster", db.Name, db.Namespace),
							})
							if err != nil {
								klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
							}
						}

					} else {
						isHealthy, err = c.checkMySQLStandaloneHealth(ctx, engine)
						if err != nil {
							klog.Errorf("MySQL standalone %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
							err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
								Type:    api.DatabaseReady,
								Status:  core.ConditionFalse,
								Reason:  api.SomeReplicasAreNotReady,
								Message: fmt.Sprintf("MySQL  stand alone %s/%s is not ready", db.Name, db.Namespace),
							})
							if err != nil {
								klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
							}
						}
					}

				}(engine)
				if isHealthy {
					break
				}

			}

			if isHealthy {
				// database is healthy. So update to "Ready" condition to "true"
				err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
					Type:    api.ServerReady,
					Status:  core.ConditionTrue,
					Reason:  api.AllReplicasAreReady,
					Message: fmt.Sprintf("MySQL %s/%s  all the replicas are joined in cluster", db.Name, db.Namespace),
				})
				if err != nil {
					klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
				}
			} else {
				// database is not healthy. So update to "Ready" condition to "false"
				err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
					Type:    api.ServerReady,
					Status:  core.ConditionFalse,
					Reason:  api.SomeReplicasAreNotReady,
					Message: fmt.Sprintf("MySQL %s/%s  not all the replicas are joined in cluster", db.Name, db.Namespace),
				})
				if err != nil {
					klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
				}
			}

			//innodb cluster has a load balancer called mysql-router
			//make sure its possible to connect from router before declaring database ready
			if db.IsInnoDBCluster() {
				engine, err := c.getMySQLClient(ctx, db, db.PrimaryServiceDNS(), api.MySQLDatabasePort)
				defer closeClientEngine(engine)
				if err != nil {
					klog.Errorf("Error while creating mysql client engine ", err.Error())
					err := c.updateMySQLStatusConditions(ctx, db,
						kmapi.Condition{
							Type:    api.DatabaseAcceptingConnection,
							Status:  core.ConditionFalse,
							Reason:  api.DatabaseNotAcceptingConnectionRequest,
							Message: fmt.Sprintf("Error while creating mysql client %s", err),
						})
					if err != nil {
						klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)

					}
					return
				}

				isRouterHealthy, err := c.checkMySQLClusterHealth(ctx, len(dbPods), engine)

				if err != nil {
					klog.Errorf("MySQL Innodb Cluster %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
					err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
						Type:    api.ServerReady,
						Status:  core.ConditionFalse,
						Reason:  api.SomeReplicasAreNotReady,
						Message: fmt.Sprintf("MySQL %s/%s  not all the replicas are joined in cluster", db.Name, db.Namespace),
					})
					if err != nil {
						klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					}
				}
				if isRouterHealthy {
					err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
						Type:    api.DatabaseReady,
						Status:  core.ConditionTrue,
						Reason:  api.AllReplicasAreReady,
						Message: fmt.Sprintf("MySQL %s/%s   all the replicas are joined in cluster", db.Name, db.Namespace),
					})
					if err != nil {
						klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					}
				} else {
					// database is not healthy. So update to "Ready" condition to "false"
					err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
						Type:    api.DatabaseReady,
						Status:  core.ConditionFalse,
						Reason:  api.SomeReplicasAreNotReady,
						Message: fmt.Sprintf("MySQL %s/%s  not all the replicas are joined in cluster", db.Name, db.Namespace),
					})
					if err != nil {
						klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					}
				}
			} else {
				if isHealthy {
					_, err := c.getMySQLClient(ctx, db, db.PrimaryServiceDNS(), api.MySQLDatabasePort)
					if err != nil {

						err = c.updateMySQLStatusConditions(ctx, db,
							kmapi.Condition{
								Type:    api.DatabaseAcceptingConnection,
								Status:  core.ConditionFalse,
								Reason:  api.DatabaseNotAcceptingConnectionRequest,
								Message: fmt.Sprintf("Error while creating mysql client %s", err),
							})
						if err != nil {
							klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)

						}
					}

					err = c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
						Type:    api.DatabaseReady,
						Status:  core.ConditionTrue,
						Reason:  api.AllReplicasAreReady,
						Message: fmt.Sprintf("MySQL %s/%s all replicas joined in cluster and accepting connection", db.Name, db.Namespace),
					})
					if err != nil {
						klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
					}
				}
			}
		}()
	}
	wg.Wait()
}

func closeClientEngine(engine *xorm.Engine) {
	if engine != nil {
		err := engine.Close()
		if err != nil {
			klog.Errorf("Can't close the engine. error: %v", err)
		}
	}
}

func (c *Controller) checkMySQLClusterHealth(ctx context.Context, members int, engine *xorm.Engine) (bool, error) {
	session := engine.NewSession()
	session.Context(ctx)
	defer session.Close()
	// sql queries for checking cluster healthiness
	// 1. ping database
	_, err := session.QueryString("SELECT 1;")
	if err != nil {
		return false, err
	}

	// 2. check all nodes are in ONLINE
	result, err := session.QueryString("SELECT MEMBER_STATE FROM performance_schema.replication_group_members;")
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("query result is nil")
	}
	if len(result) != members {
		return false, fmt.Errorf("not all members have joined into the group yet")
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

func (c *Controller) checkMySQLStandaloneHealth(ctx context.Context, engine *xorm.Engine) (bool, error) {
	session := engine.NewSession()
	session.Context(ctx)
	defer session.Close()
	// sql queries for checking standalone healthiness
	// 1. ping database
	_, err := session.QueryString("SELECT 1;")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Controller) getMySQLClient(ctx context.Context, db *api.MySQL, dns string, port int32) (*xorm.Engine, error) {
	user, pass, err := c.getDBRootCredential(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("DB basic auth is not found for MySQL %v/%v", db.Namespace, db.Name)
	}
	tlsParam := ""
	if db.Spec.TLS != nil {
		serverSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, db.MustCertSecretName(api.MySQLClientCert), metav1.GetOptions{})
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
			tlsParam = fmt.Sprintf("tls=%s", TLSValueCustom)
		} else {
			tlsParam = fmt.Sprintf("tls=%s", TLSValueSkipVerify)
		}
	}

	cnnstr := fmt.Sprintf("%v:%v@tcp(%s:%d)/%s?%s", user, pass, dns, port, api.ResourceSingularMySQL, tlsParam)
	engine, err := xorm.NewEngine(api.ResourceSingularMySQL, cnnstr)
	if err != nil {

		return engine, err
	}
	engine.SetDefaultContext(ctx)
	return engine, nil
}

func (c *Controller) getDBRootCredential(ctx context.Context, db *api.MySQL) (string, string, error) {
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

func HostDNS(db *api.MySQL, podMeta metav1.ObjectMeta) string {
	return fmt.Sprintf("%v.%v.%v.svc", podMeta.Name, db.GoverningServiceName(), podMeta.Namespace)
}

func (c *Controller) updateMySQLStatusConditions(ctx context.Context, db *api.MySQL, conditions ...kmapi.Condition) error {
	_, err := util.UpdateMySQLStatus(
		ctx,
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.MySQLStatus) (types.UID, *api.MySQLStatus) {
			for _, con := range conditions {
				in.Conditions = kmapi.SetCondition(in.Conditions, con)
			}
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	return err
}
