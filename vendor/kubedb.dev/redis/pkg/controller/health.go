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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	configgenerator "kubedb.dev/apimachinery/pkg/config_generator"

	rd "github.com/go-redis/redis"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckRedisHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckRedisHealth(stopCh)
}

func (c *Controller) CheckRedisHealth(stopCh <-chan struct{}) {
	klog.Info("Starting Redis health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down Redis health checker...")
			break
		default:
			c.CheckRedisHealthOnce()
			time.Sleep(api.HealthCheckInterval)
		}
	}
}

func (c *Controller) CheckRedisHealthOnce() {
	dbList, err := c.rdLister.Redises(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list Redis objects with: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	for idx := range dbList {
		db := dbList[idx]
		if db.DeletionTimestamp != nil || db.Spec.Halted {
			continue
		}

		wg.Add(1)
		go func(db *api.Redis) {
			defer func() {
				wg.Done()
			}()

			rdClient, err := c.getRedisClient(db, nil)
			if err != nil {
				klog.Errorf("Failed to get redis rdClient for Redis: %s/%s error: %s", db.Namespace, db.Name, err.Error())
				// Since the rdClient was unable to connect the database,
				// update "AcceptingConnection" to "false".
				// update "Ready" to "false"
				c.updateErrorAcceptingConnections(db, err)
				// Since the rdClient isn't created, skip rest operations.
				return
			}
			defer rdClient.Close()

			if db.Spec.Mode == api.RedisModeCluster {
				err = c.CheckClusterSlotsForClusterMode(db)
				if err != nil {
					klog.Errorf("failed on cluster slots check. error:", err.Error())
					c.updateErrorAcceptingConnections(db, err)
					return
				}
			}

			// If the rdClient was created without any error, and for cluster mode if the slot is ok then,
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
							Message:            fmt.Sprintf("The Redis: %s/%s is accepting rdClient requests.", db.Namespace, db.Name),
						})
					return db.UID, in
				},
				metav1.UpdateOptions{},
			)
			if err != nil {
				klog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
				// Since condition update failed, skip remaining operations.
				return
			}

			pingResult, err := rdClient.Ping().Result()
			if err != nil {
				c.updateDatabaseNotReady(db)
				klog.Errorf("Failed to ping the database: %s/%s error: %s", db.Namespace, db.Name, err.Error())
				return
			} else if !strings.Contains(pingResult, "PONG") {
				c.updateDatabaseNotReady(db)
				klog.Errorf("Ping returned unexpected reply for the database: %s/%s reply: %s", db.Namespace, db.Name, pingResult)
				return
			}

			if db.Spec.Mode == api.RedisModeCluster || db.Spec.Mode == api.RedisModeSentinel {
				isHealthy, err := c.checkRedisClusterHealth(db)
				if err != nil {
					klog.Errorf("Failed to ping the database Replicas: %s/%s error: %s", db.Namespace, db.Name, err.Error())
					c.updateDatabaseNotReady(db)
					return
				}
				if !isHealthy {
					klog.Errorf("Failed to ping the database Replicas: %s/%s ", db.Namespace, db.Name)
					c.updateDatabaseNotReady(db)
					return
				}
			}
			c.updateDatabaseReady(db)
		}(db)
	}
	// Wait until all go-routine complete executions
	wg.Wait()
}

func (c *Controller) getRedisClient(db *api.Redis, dnsNames []string) (*rd.Client, error) {
	address := db.Address()
	if len(dnsNames) > 0 {
		address = dnsNames[0]
	}
	rdOpts := &rd.Options{
		DialTimeout: 15 * time.Second,
		IdleTimeout: 3 * time.Second,
		PoolSize:    1,
		Addr:        address,
	}
	if !db.Spec.DisableAuth {
		if db.Spec.AuthSecret == nil {
			return nil, errors.New("no database secret")
		}
		authSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		rdOpts.Password = string(authSecret.Data[core.BasicAuthPasswordKey])
	}
	if db.Spec.TLS != nil {
		sec, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.CertificateName(api.RedisClientCert), metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "error in getting the secret")
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(sec.Data["ca.crt"])
		cert, err := tls.X509KeyPair(sec.Data["tls.crt"], sec.Data["tls.key"])
		if err != nil {
			klog.Error(err, "error in making certificate")
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
		klog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
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
		klog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
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
		klog.Errorf("Failed to update status for Redis: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) checkRedisClusterHealth(db *api.Redis) (bool, error) {
	podList, err := c.Client.CoreV1().Pods(db.Namespace).List(context.TODO(), metav1.ListOptions{
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
		err := c.IsRedisServerOnline(db, HostDNS(db, pod.ObjectMeta))
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// try to query in server if failed return err that means not online
func (c *Controller) IsRedisServerOnline(db *api.Redis, dnsName string) error {
	var err error
	client, err := c.getRedisClient(db, []string{dnsName})
	if err != nil {
		return err
	}
	defer client.Close()

	pingResult, err := client.Ping().Result()
	if err != nil {
		return err
	} else if !strings.Contains(pingResult, "PONG") {
		return fmt.Errorf("ping returned unexpected reply for the database: %s/%s reply: %s", db.Namespace, db.Name, pingResult)
	}
	return nil
}

// make host dns with require template
func HostDNS(db *api.Redis, podMeta metav1.ObjectMeta) string {
	return fmt.Sprintf("%v.%v.%v.svc:%d", podMeta.Name, db.GoverningServiceName(), podMeta.Namespace, api.RedisDatabasePort)
}

func (c *Controller) CheckClusterSlotsForClusterMode(db *api.Redis) error {
	if db.Spec.Mode != api.RedisModeCluster {
		return nil
	}
	rdClient, err := c.getRedisClient(db, nil)
	if err != nil {
		return fmt.Errorf("failed to get redis rdClient for Redis: %s/%s error: %s", db.Namespace, db.Name, err.Error())
	}
	defer rdClient.Close()

	res, err := rdClient.ClusterInfo().Result()
	if err != nil {
		return fmt.Errorf("failed to get cluster info from the database: %s/%s error: %s", db.Namespace, db.Name, err.Error())
	}
	clusterInfos := configgenerator.ConvertStringInToMap(res, []string{":", "="})
	state, _ := clusterInfos.Get("cluster_state")
	ClusterState := state.(*configgenerator.ValueGenerator).Value
	aSlots, _ := clusterInfos.Get("cluster_slots_assigned") // this will parse the total number of slots
	assignedSlots, err := strconv.Atoi(aSlots.(*configgenerator.ValueGenerator).Value)
	if err != nil {
		return fmt.Errorf("failed to get cluster assigned slots from the database: %s/%s error: %s", db.Namespace, db.Name, err.Error())
	}
	slots, _ := clusterInfos.Get("cluster_slots_ok") // this will parse the ok slots
	okSlots, err := strconv.Atoi(slots.(*configgenerator.ValueGenerator).Value)
	if err != nil {
		return fmt.Errorf("failed to get cluster ok slots from the database: %s/%s error: %s", db.Namespace, db.Name, err.Error())
	}
	fSlots, _ := clusterInfos.Get("cluster_slots_fail") // this will parse the missing number of slots
	failedSlots, err := strconv.Atoi(fSlots.(*configgenerator.ValueGenerator).Value)
	if err != nil {
		return fmt.Errorf("failed to get cluster failed slots from the database: %s/%s error: %s", db.Namespace, db.Name, err.Error())
	}
	if ClusterState != "ok" && okSlots == assignedSlots || failedSlots > 0 {
		return fmt.Errorf("the redis cluster is not in healthy state %s/%s", db.Namespace, db.Name)
	}
	return nil
}
