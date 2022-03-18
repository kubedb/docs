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
	"sync"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	_ "github.com/go-sql-driver/mysql"
	sql_driver "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
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

// CheckMySQLHealthOnce check the database health in every sudden interval (10s).
// it will list all the databases and then run the health check and update the status accordingly
// for any database topology it will query in the database server and update the conditions like databaseReady,AcceptingConnection

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

			// loop over all pods and check db health and update db conditions
			replicaNotReadyCount := 0
			for _, pod := range dbPods {

				dns := HostDNS(db, pod.ObjectMeta)
				//innodbCluster uses router as a load balancer
				// which is connected to router Primary service
				if db.IsInnoDBCluster() {
					dns = db.PrimaryServiceDNS()
				}

				engine, err := c.getMySQLClient(ctx, db, dns, api.MySQLDatabasePort)
				//unable get clientConnection means database server is not healthy
				if err != nil {
					klog.Error(err)
					err = c.updateConditionsForNotHealthy(ctx, db, err)
					if err != nil {
						klog.Error(err)
					}
					return
				}

				func(engine *xorm.Engine) {
					defer closeClientEngine(engine)

					//checkHeath for StandAlone and update status
					if db.Spec.Topology == nil {
						healthy, err := c.checkMySQLStandaloneHealth(ctx, engine)

						if err != nil || !healthy {
							klog.Errorf("database instance %s is not healthy. Reason %v", db.GetNameSpacedName(), err)
							err = c.updateConditionsForNotHealthy(ctx, db, err)
							if err != nil {
								klog.Error(err)
							}
							return
						}
						//db is healthy
						err = c.updateConditionsForHealthy(ctx, db)
						if err != nil {
							klog.Error(err)
							return
						}
					}

					//check Health for Read Replica and update status
					if db.IsReadReplica() {
						healthy, err := c.checkMySQLReadReplicaHealth(ctx, engine, db)
						if err != nil || !healthy {
							klog.Errorf("read replica %s is not healthy.Reason %v", db.GetNameSpacedName(), err)
							err := c.updateConditionsForNotHealthy(ctx, db, err)
							if err != nil {
								klog.Error(err)
							}
							return
						}
						//db is healthy
						err = c.updateConditionsForHealthy(ctx, db)

						if err != nil {
							klog.Error(err)
							return
						}
					}

					//check Health for GroupReplication and update status
					//check Health for InnodbCluster and update status
					if db.UsesGroupReplication() || db.IsInnoDBCluster() {
						healthy, err := c.checkMySQLClusterHealth(ctx, len(dbPods), engine)
						if err != nil {
							replicaNotReadyCount++
						}
						if err != nil || !healthy {
							err = c.updateMySQLStatusConditions(ctx, db,
								kmapi.Condition{
									Type:    api.DatabaseReady,
									Status:  core.ConditionFalse,
									Reason:  api.SomeReplicasAreNotReady,
									Message: fmt.Sprintf("database %s in not accepting connection ,%v", db.GetNameSpacedName(), err),
								},
								kmapi.Condition{
									Type:    api.DatabaseReplicaReady,
									Status:  core.ConditionFalse,
									Reason:  api.SomeReplicasAreNotReady,
									Message: fmt.Sprintf("replica is not accepting connection %v", pod.ObjectMeta.Name),
								},
							)
							if err != nil {
								klog.Error(err)
								return
							}
							return
						}

						err = c.updateConditionsForHealthy(ctx, db)
						if err != nil {
							klog.Error(err)
							return
						}
					}
				}(engine)

			}

			if db.UsesGroupReplication() || db.IsInnoDBCluster() {
				if replicaNotReadyCount == len(dbPods) {
					err := c.updateConditionsForNotHealthy(ctx, db, errors.New("all replica is not ready"))
					if err != nil {
						klog.Error(err)
						return
					}
				} else {
					err := c.updateMySQLStatusConditions(ctx, db, kmapi.Condition{
						Type:   api.DatabaseAcceptingConnection,
						Status: core.ConditionTrue,
						Reason: api.DatabaseAcceptingConnectionRequest,
					})
					if err != nil {
						klog.Error(err)
						return
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

func (c *Controller) checkMySQLReadReplicaHealth(ctx context.Context, engine *xorm.Engine, db *api.MySQL) (bool, error) {
	//check is online
	session := engine.NewSession().Context(ctx)
	defer func(session *xorm.Session) {
		err := session.Close()
		if err != nil {
			klog.Error(err)
		}
	}(session)

	_, err := session.QueryString("select 1;")
	if err != nil {
		return false, err
	}

	// check is replication running
	res, err := session.QueryString("show slave status;")
	if err != nil {
		return false, errors.Wrap(err, "error query replica status")
	}
	if res != nil {
		if res[0]["Slave_SQL_Running"] == "Yes" && res[0]["Slave_IO_Running"] == "Yes" {
			return true, nil
		}
	}
	return false, fmt.Errorf("read replica %v is not healthy", db.GetNameSpacedName())
}

func (c *Controller) checkMySQLClusterHealth(ctx context.Context, members int, engine *xorm.Engine) (bool, error) {
	session := engine.NewSession().Context(ctx)
	defer func(session *xorm.Session) {
		err := session.Close()
		if err != nil {
			klog.Error(err)
		}
	}(session)

	// 2. check all nodes are in ONLINE
	result, err := session.QueryString("SELECT count(MEMBER_STATE) as online FROM performance_schema.replication_group_members WHERE MEMBER_STATE = 'ONLINE';")
	if err != nil {
		return false, err
	}

	if result == nil {
		return false, fmt.Errorf("query result is nil")
	}
	online, ok := result[0]["online"]
	if ok && online == strconv.Itoa(members) {
		return true, nil
	} else {
		return false, nil
	}

}

func (c *Controller) checkMySQLStandaloneHealth(ctx context.Context, engine *xorm.Engine) (bool, error) {
	session := engine.NewSession().Context(ctx)
	defer func(session *xorm.Session) {
		err := session.Close()
		if err != nil {
			klog.Error(err)
		}
	}(session)
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

func (c *Controller) updateConditionsForNotHealthy(ctx context.Context, db *api.MySQL, reason error) error {

	err := c.updateMySQLStatusConditions(ctx, db,
		kmapi.Condition{
			Type:    api.DatabaseAcceptingConnection,
			Status:  core.ConditionFalse,
			Reason:  api.DatabaseNotAcceptingConnectionRequest,
			Message: fmt.Sprintf("database is not accepting connection , %v", reason),
		},
		kmapi.Condition{
			Type:    api.DatabaseReady,
			Status:  core.ConditionFalse,
			Reason:  api.DatabaseNotAcceptingConnectionRequest,
			Message: fmt.Sprintf("database in not accepting connection %v", reason),
		},
		kmapi.Condition{
			Type:    api.DatabaseReplicaReady,
			Status:  core.ConditionFalse,
			Reason:  api.SomeReplicasAreNotReady,
			Message: fmt.Sprintf("database is not accepting connection , %v", reason),
		},
	)
	return err
}

func (c *Controller) updateConditionsForHealthy(ctx context.Context, db *api.MySQL) error {

	err := c.updateMySQLStatusConditions(ctx, db,
		kmapi.Condition{
			Type:    api.DatabaseAcceptingConnection,
			Status:  core.ConditionTrue,
			Reason:  api.DatabaseAcceptingConnection,
			Message: fmt.Sprintf("database %s is accepting connection", db.GetNameSpacedName()),
		},
		kmapi.Condition{
			Type:    api.DatabaseReady,
			Status:  core.ConditionTrue,
			Reason:  api.AllReplicasAreReady,
			Message: fmt.Sprintf("database %s is ready", db.GetNameSpacedName()),
		},
		kmapi.Condition{
			Type:    api.DatabaseReplicaReady,
			Status:  core.ConditionTrue,
			Reason:  api.AllReplicasAreReady,
			Message: fmt.Sprintf("database %s is ready", db.GetNameSpacedName()),
		},
	)
	return err
}
