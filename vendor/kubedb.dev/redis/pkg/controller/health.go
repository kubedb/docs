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

	rd "github.com/go-redis/redis"
	"github.com/golang/glog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckRedisHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckRedisHealth(stopCh)
}

func (c *Controller) CheckRedisHealth(stopCh <-chan struct{}) {
	go wait.Until(func() {
		dbList, err := c.rdLister.Redises(core.NamespaceAll).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to list Redis objects with: %s", err.Error())
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

				client, err := c.getRedisClient(db)
				if err != nil {
					glog.Errorf("Failed to get redis client for Redis: %s/%s error: %s", db.Namespace, db.Name, err.Error())
					// Since the client was unable to connect the database,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnections(db, err)
					// Since the client isn't created, skip rest operations.
					return
				}
				// If the client was created without any error,
				// the database is accepting connection.
				// Update "AcceptingConnection" to "true".
				_, err = util.UpdateRedisStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseAcceptingConnection,
								Status:             core.ConditionTrue,
								Reason:             api.DatabaseAcceptingConnectionRequest,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("The Redis: %s/%s is accepting client requests.", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					glog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
					// Since condition update failed, skip remaining operations.
					return
				}

				pingResult, err := client.Ping().Result()
				if err != nil {
					c.updateDatabaseNotReady(db)
					glog.Errorf("Failed to ping the database: %s/%s error: %s", db.Namespace, db.Name, err.Error())
					return
				} else if !strings.Contains(pingResult, "PONG") {
					c.updateDatabaseNotReady(db)
					glog.Errorf("Ping returned unexpected reply for the database: %s/%s reply: %s", db.Namespace, db.Name, pingResult)
					return
				}

				c.updateDatabaseReady(db)
			}()
		}
	}, c.ReadinessProbeInterval, stopCh)

	// will wait here until stopCh is closed.
	<-stopCh
	glog.Info("Shutting down Redis health checker...")
}

func (c *Controller) getRedisClient(db *api.Redis) (*rd.Client, error) {
	rdOpts := &rd.Options{
		Addr: db.Address(),
	}
	if db.Spec.TLS != nil {
		sec, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.CertificateName(api.RedisClientCert), metav1.GetOptions{})
		if err != nil {
			glog.Error(err, "error in getting the secret")
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(sec.Data["ca.crt"])
		cert, err := tls.X509KeyPair(sec.Data["tls.crt"], sec.Data["tls.key"])
		if err != nil {
			glog.Error(err, "error in making certificate")
			return nil, err
		}
		rdOpts.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},
			ClientCAs: pool,
			RootCAs:   pool,
		}
	}
	rdClient := rd.NewClient(rdOpts)
	return rdClient, nil
}

func (c *Controller) updateErrorAcceptingConnections(db *api.Redis, connectionErr error) {
	_, err := util.UpdateRedisStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseAcceptingConnection,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseNotAcceptingConnectionRequest,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Redis: %s/%s is not accepting client requests. error: %s", db.Namespace, db.Name, connectionErr),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Redis: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		glog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateDatabaseReady(db *api.Redis) {
	_, err := util.UpdateRedisStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionTrue,
					Reason:             api.ReadinessCheckSucceeded,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Redis: %s/%s is ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		glog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateDatabaseNotReady(db *api.Redis) {
	_, err := util.UpdateRedisStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Redis: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		glog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
	}
}
