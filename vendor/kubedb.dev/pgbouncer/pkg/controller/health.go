/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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
	"kmodules.xyz/client-go/tools/certholder"
	"xorm.io/xorm"
)

func (c *Controller) RunHealthChecker(stopch <-chan struct{}) {
	// As CheckPgBouncerDBHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckPgBouncerDBHealth(stopch)
}

func (c *Controller) CheckPgBouncerDBHealth(stopch <-chan struct{}) {
	klog.Info("Starting PgBouncer health checker...")
	for {
		select {
		case <-stopch:
			klog.Info("Shutting down PgBouncer health checker...")
			break
		default:
			c.CheckPgBouncerDBHealthOnce()
			time.Sleep(api.HealthCheckInterval)
		}
	}
}

func (c *Controller) CheckPgBouncerDBHealthOnce() {
	dbList, err := c.pbLister.PgBouncers(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list PgBouncer objects with: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	for idx := range dbList {
		db := dbList[idx]
		// If the DB object is deleted , no need to perform health check.
		if db.DeletionTimestamp != nil {
			continue
		}

		wg.Add(1)
		go func(db *api.PgBouncer) {
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
			err = c.IsPgBouncerServerOnline(ctx, db, PrimaryServiceDNS(db), port)
			if err != nil {
				// Since the client was unable to connect the database,
				// update "AcceptingConnection" to "false".
				// update "Ready" to "false"
				c.updateErrorAcceptingConnections(ctx, db, err)
				return
			}

			// While creating the client, we perform a health check along with it.
			// If the client is created without any error,
			// the database is accepting connection.
			// Update "AcceptingConnection" to "true".
			_, err = util.UpdatePgBouncerStatus(
				ctx,
				c.DBClient.KubedbV1alpha2(),
				db.ObjectMeta,
				func(in *api.PgBouncerStatus) (types.UID, *api.PgBouncerStatus) {
					in.Conditions = kmapi.SetCondition(in.Conditions,
						kmapi.Condition{
							Type:               api.DatabaseAcceptingConnection,
							Status:             core.ConditionTrue,
							Reason:             api.DatabaseAcceptingConnectionRequest,
							ObservedGeneration: db.Generation,
							Message:            fmt.Sprintf("The PgBouncer: %s/%s is accepting client requests.", db.Namespace, db.Name),
						})
					return db.UID, in
				},
				metav1.UpdateOptions{},
			)
			if err != nil {
				klog.Errorf("Failed to update status for PgBouncer: %s/%s", db.Namespace, db.Name)
				// Since condition update failed, skip remaining operations.
				return
			}

			var isHealthy bool
			if *db.Spec.Replicas > int32(1) {
				isHealthy, err = c.checkPgBouncerClusterHealth(ctx, db)
				if err != nil {
					klog.Errorf("PgBouncer Cluster %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
				}
			} else {
				isHealthy, err = c.checkPgBOuncerStandaloneHealth(ctx, db)
				if err != nil {
					klog.Errorf("PgBouncer standalone %s/%s is not healthy, reason: %s", db.Namespace, db.Name, err.Error())
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

func (c *Controller) updateDatabaseNotReady(ctx context.Context, db *api.PgBouncer, errMsg error) {
	_, err := util.UpdatePgBouncerStatus(
		ctx,
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.PgBouncerStatus) (types.UID, *api.PgBouncerStatus) {
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
					Message:            fmt.Sprintf("The PgBouncer: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for PgBouncer: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateDatabaseReady(ctx context.Context, db *api.PgBouncer) {
	_, err := util.UpdatePgBouncerStatus(
		ctx,
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.PgBouncerStatus) (types.UID, *api.PgBouncerStatus) {
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
		klog.Errorf("Failed to update status for PgBouncer: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) checkPgBouncerClusterHealth(ctx context.Context, db *api.PgBouncer) (bool, error) {
	err := c.IsPgBouncerServerOnline(ctx, db, PrimaryServiceDNS(db), api.PgBouncerDatabasePort)
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
		err := c.IsPgBouncerServerOnline(ctx, db, HostDNS(db, pod.ObjectMeta), api.PgBouncerDatabasePort)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// check if the server is ready to accept connections
func (c *Controller) checkPgBOuncerStandaloneHealth(ctx context.Context, db *api.PgBouncer) (bool, error) {
	err := c.IsPgBouncerServerOnline(ctx, db, PrimaryServiceDNS(db), api.PgBouncerDatabasePort)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Controller) IsPgBouncerServerOnline(ctx context.Context, db *api.PgBouncer, dnsName string, port int32) error {
	var err error
	eng, err := c.GetPgBouncerClient(ctx, db, dnsName, port)
	if err != nil {
		return err
	}
	defer eng.Close()

	queryString := "SHOW LISTS;"

	res, err := eng.QueryString(queryString)
	if err != nil {
		return err
	}

	if len(res[0]["list"]) > 0 {
		return nil
	} else {
		return fmt.Errorf("can't get query value")
	}
}

func (c *Controller) GetPgBouncerAuthCredentials(ctx context.Context, db *api.PgBouncer) (string, string, error) {
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, db.AuthSecretName(), metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return api.PgBouncerAdminUsername, string(secret.Data[pbAdminPasswordKey]), nil
}

func (c *Controller) GetPgBouncerClient(ctx context.Context, db *api.PgBouncer, dnsName string, port int32) (*xorm.Engine, error) {
	user, pass, err := c.GetPgBouncerAuthCredentials(ctx, db)
	if err != nil {
		return nil, err
	}
	cnnstr := ""
	sslMode := db.Spec.SSLMode
	if sslMode == "" {
		sslMode = api.PgBouncerSSLModeDisable
	}

	if db.Spec.TLS != nil {
		secretName := db.GetCertSecretName(api.PgBouncerClientCert)

		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "failed to get certificate secret.", secretName)
			return nil, err
		}

		certs, _ := certholder.DefaultHolder.ForResource(api.SchemeGroupVersion.WithResource(api.ResourcePluralPgBouncer), db.ObjectMeta)
		paths, err := certs.Save(certSecret)
		if err != nil {
			klog.Error(err, "failed to save certificate")
			return nil, err
		}
		if db.Spec.ConnectionPool.AuthType == api.PgBouncerClientAuthModeCert || db.Spec.SSLMode == api.PgBouncerSSLModeVerifyCA || db.Spec.SSLMode == api.PgBouncerSSLModeVerifyFull {
			cnnstr = fmt.Sprintf("user=%s password=%s host=%s port=%d connect_timeout=15 dbname=pgbouncer sslmode=%s sslrootcert=%s sslcert=%s sslkey=%s", user, pass, dnsName, port, sslMode, paths.CACert, paths.Cert, paths.Key)
		} else {
			cnnstr = fmt.Sprintf("user=%s password=%s host=%s port=%d connect_timeout=15 dbname=pgbouncer sslmode=%s sslrootcert=%s", user, pass, dnsName, port, sslMode, paths.CACert)
		}
	} else {
		cnnstr = fmt.Sprintf("user=%s password=%s host=%s port=%d connect_timeout=15 dbname=pgbouncer sslmode=%s", user, pass, dnsName, port, sslMode)
	}
	eng, err := xorm.NewEngine("postgres", cnnstr)
	if err != nil {
		return nil, err
	}
	eng.SetDefaultContext(ctx)
	return eng, nil
}

// if the master is not accepting connection then set database ready condition and accepting connection to false
func (c *Controller) updateErrorAcceptingConnections(ctx context.Context, db *api.PgBouncer, connectionErr error) {
	_, err := util.UpdatePgBouncerStatus(
		ctx,
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.PgBouncerStatus) (types.UID, *api.PgBouncerStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseAcceptingConnection,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseNotAcceptingConnectionRequest,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The PgBouncer: %s/%s is not accepting client requests. error: %s", db.Namespace, db.Name, connectionErr),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The PgBouncer: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for PgBouncer: %s/%s", db.Namespace, db.Name)
	}
}

// make host dns with require template
func HostDNS(db *api.PgBouncer, podMeta metav1.ObjectMeta) string {
	return fmt.Sprintf("%v.%v.%v.svc", podMeta.Name, db.GoverningServiceName(), podMeta.Namespace)
}

// make primary host dns with require template
func PrimaryServiceDNS(db *api.PgBouncer) string {
	return fmt.Sprintf("%v.%v.svc", db.ServiceName(), db.Namespace)
}
