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

			for _, pod := range podList.Items {
				if core_util.IsPodConditionTrue(pod.Status.Conditions, core_util.PodConditionTypeReady) {
					continue
				}

				engine, err := c.getMySQLClient(ctx, db, HostDNS(db, pod.ObjectMeta), api.MySQLDatabasePort)
				if err != nil {
					klog.Warning("Failed to get db client for host ", pod.Namespace, "/", pod.Name)
					continue
				}

				func(engine *xorm.Engine) {
					defer func() {
						if engine != nil {
							err = engine.Close()
							if err != nil {
								klog.Errorf("Can't close the engine. error: %v", err)
							}
						}
					}()

					isHostOnline, err := c.isHostOnline(ctx, db, engine)
					if err != nil {
						klog.Warning("Host is not online ", err.Error())
					}

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
							klog.Warning("Failed to update pod status with: ", err.Error())
						}
					}
				}(engine)
			}

			// verify db is going to accepting connection and in ready state
			port, err := c.GetPrimaryServicePort(db)
			if err != nil {
				klog.Warning("Failed to primary service port with: ", err.Error())
				return
			}
			engine, err := c.getMySQLClient(ctx, db, db.PrimaryServiceDNS(), port)
			if err != nil {
				// Since the client was unable to connect the database,
				// update "AcceptingConnection" to "false".
				// update "Ready" to "false"
				_, err = util.UpdateMySQLStatus(
					ctx,
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
					klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
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
			_, err = util.UpdateMySQLStatus(
				ctx,
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
				klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
				// Since condition update failed, skip remaining operations.
				return
			}

			// check MySQL database health
			var isHealthy bool
			if *db.Spec.Replicas > int32(1) && db.Spec.Topology != nil && db.Spec.Topology.Group != nil {
				isHealthy, err = c.checkMySQLClusterHealth(ctx, len(podList.Items), engine)
				if err != nil {
					klog.Errorf("MySQL Cluster %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
				}
			} else {
				isHealthy, err = c.checkMySQLStandaloneHealth(ctx, engine)
				if err != nil {
					klog.Errorf("MySQL standalone %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
				}
			}

			if isHealthy {
				// database is healthy. So update to "Ready" condition to "true"
				_, err = util.UpdateMySQLStatus(
					ctx,
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
					klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
				}
			} else {
				// database is not healthy. So update to "Ready" condition to "false"
				_, err = util.UpdateMySQLStatus(
					ctx,
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.MySQLStatus) (types.UID, *api.MySQLStatus) {
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
					klog.Errorf("Failed to update status for MySQL: %s/%s", db.Namespace, db.Name)
				}
			}
		}()
	}
	wg.Wait()
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
		return false, fmt.Errorf("Not all members have joined into the group yet")
	}

	for j := range result {
		memberState, ok := result[j]["MEMBER_STATE"]
		if !ok || strings.Compare(memberState, "ONLINE") != 0 {
			return false, fmt.Errorf("All group member are not online yet")
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

func (c *Controller) isHostOnline(ctx context.Context, db *api.MySQL, engine *xorm.Engine) (bool, error) {

	session := engine.NewSession()
	session.Context(ctx)
	defer session.Close()
	// 1. ping for both standalone and group replication member
	_, err := session.QueryString("SELECT 1;")
	if err != nil {
		return false, err
	}

	if db.UsesGroupReplication() {
		result, err := session.QueryString("select member_state from performance_schema.replication_group_members where member_id=@@server_uuid;")
		if err != nil {
			return false, err
		}
		if result == nil {
			return false, fmt.Errorf("Checking member state, query result is nil")
		}
		memberState, ok := result[0]["member_state"]
		if !ok || strings.Compare(memberState, "ONLINE") != 0 {
			return false, fmt.Errorf("The member is not online yet")
		}
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
		serverSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, db.MustCertSecretName(api.MySQLServerCert), metav1.GetOptions{})
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
	engine, err := xorm.NewEngine("mysql", cnnstr)
	if err != nil {
		return nil, fmt.Errorf("failed to create xorm engine")
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
