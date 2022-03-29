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

	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func (c *Controller) RunSentinelHealthChecker(stopCh <-chan struct{}) {
	// As CheckRedisHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckSentinelHealth(stopCh)
}

func (c *Controller) CheckSentinelHealth(stopCh <-chan struct{}) {
	klog.Info("Starting Sentinel health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down Sentinel health checker...")
			break
		default:
			c.CheckSentinelHealthOnce()
			time.Sleep(api.HealthCheckInterval)
		}
	}
}

func (c *Controller) CheckSentinelHealthOnce() {
	dbList, err := c.rsLister.RedisSentinels(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list Sentinel objects with: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	for idx := range dbList {
		db := dbList[idx]
		if db.DeletionTimestamp != nil || db.Spec.Halted {
			continue
		}

		wg.Add(1)
		go func(db *api.RedisSentinel) {
			defer func() {
				wg.Done()
			}()
			for i := 0; i < int(pointer.Int32(db.Spec.Replicas)); i++ {
				dnsName := fmt.Sprintf("%s-%v.%s.%s.svc", db.Name, i, db.GoverningServiceName(), db.Namespace)
				client, err := c.getRedisSentinelClient(db, dnsName, api.RedisSentinelPort)
				if err != nil {
					klog.Errorf("Failed to get Sentinel client for Sentinel: %s/%s error: %s", db.Namespace, db.Name, err.Error())
					// Since the client was unable to connect the database,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					c.updateErrorAcceptingConnectionsSentinel(db, err)
					// Since the client isn't created, skip rest operations.
					return
				}
				defer client.Close()
				// If the client was created without any error,
				// the database is accepting connection.
				// Update "AcceptingConnection" to "true".
				_, err = util.UpdateRedisSentinelStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseAcceptingConnection,
								Status:             core.ConditionTrue,
								Reason:             api.DatabaseAcceptingConnectionRequest,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("The Sentinel: %s/%s is accepting client requests.", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					klog.Errorf("Failed to update status for Sentinel: %s/%s", db.Namespace, db.Name)
					// Since condition update failed, skip remaining operations.
					return
				}

				pingResult, err := client.Ping().Result()
				if err != nil {
					c.updateSentinelNotReady(db)
					klog.Errorf("Failed to ping the Sentinel: %s/%s error: %s", db.Namespace, db.Name, err.Error())
					return
				} else if !strings.Contains(pingResult, "PONG") {
					c.updateSentinelNotReady(db)
					klog.Errorf("Ping returned unexpected reply for the database: %s/%s reply: %s", db.Namespace, db.Name, pingResult)
					return
				}
			}

			c.updateSentinelReady(db)
		}(db)
	}
	// Wait until all go-routine complete executions
	wg.Wait()
}

func (c *Controller) updateErrorAcceptingConnectionsSentinel(db *api.RedisSentinel, connectionErr error) {
	_, err := util.UpdateRedisSentinelStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseAcceptingConnection,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseNotAcceptingConnectionRequest,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Sentinel: %s/%s is not accepting client requests. error: %s", db.Namespace, db.Name, connectionErr),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Sentinel: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for Sentinel: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateSentinelReady(db *api.RedisSentinel) {
	_, err := util.UpdateRedisSentinelStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionTrue,
					Reason:             api.ReadinessCheckSucceeded,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Sentinel: %s/%s is ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for Sentinel: %s/%s", db.Namespace, db.Name)
	}
}

func (c *Controller) updateSentinelNotReady(db *api.RedisSentinel) {
	_, err := util.UpdateRedisSentinelStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.ReadinessCheckFailed,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Sentinel: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		klog.Errorf("Failed to update status for Sentinel: %s/%s", db.Namespace, db.Name)
	}
}
