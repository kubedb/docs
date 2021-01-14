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
	"sync"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	go_es "kubedb.dev/elasticsearch/pkg/util/go-es"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func (c *Controller) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckElasticsearchHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckElasticsearchHealth(stopCh)
}

func (c *Controller) CheckElasticsearchHealth(stopCh <-chan struct{}) {
	glog.Info("Starting Elasticsearch health checker...")

	go wait.Until(func() {
		dbList, err := c.esLister.Elasticsearches(core.NamespaceAll).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to list Elasticsearch objects with: %s", err.Error())
			return
		}

		var wg sync.WaitGroup
		for idx := range dbList {
			db := dbList[idx]

			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
				}()

				// Create database client
				dbClient, err := c.GetElasticsearchClient(db)
				if err != nil {
					glog.Warningf("The Elasticsearch: %s/%s client is not ready with %s", db.Namespace, db.Name, err.Error())
					// Since the client was unable to connect the database,
					// update "AcceptingConnection" to "false".
					// update "Ready" to "false"
					_, err = util.UpdateElasticsearchStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						db.ObjectMeta,
						func(in *api.ElasticsearchStatus) (types.UID, *api.ElasticsearchStatus) {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:               api.DatabaseAcceptingConnection,
									Status:             core.ConditionFalse,
									Reason:             api.DatabaseNotAcceptingConnectionRequest,
									ObservedGeneration: db.Generation,
									Message:            fmt.Sprintf("The Elasticsearch: %s/%s is not accepting client requests.", db.Namespace, db.Name),
								})
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:               api.DatabaseReady,
									Status:             core.ConditionFalse,
									Reason:             api.ReadinessCheckFailed,
									ObservedGeneration: db.Generation,
									Message:            fmt.Sprintf("The Elasticsearch: %s/%s is not ready.", db.Namespace, db.Name),
								})
							return db.UID, in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						glog.Errorf("Failed to update status for Elasticsearch: %s/%s with %s", db.Namespace, db.Name, err.Error())
					}
					// Since the client isn't created, skip rest operations.
					return
				}

				// While creating the client, we perform a health check along with it.
				// If the client is created without any error,
				// the database is accepting connection.
				// Update "AcceptingConnection" to "true".
				_, err = util.UpdateElasticsearchStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.ElasticsearchStatus) (types.UID, *api.ElasticsearchStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseAcceptingConnection,
								Status:             core.ConditionTrue,
								Reason:             api.DatabaseAcceptingConnectionRequest,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("The Elasticsearch: %s/%s is accepting client requests.", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					glog.Errorf("Failed to update status for Elasticsearch: %s/%s with %s", db.Namespace, db.Name, err.Error())
					// Since condition update failed, skip remaining operations.
					return
				}

				// Get database status, could be red, green or yellow.
				status, err := dbClient.ClusterStatus()
				if err != nil {
					glog.Errorf("Failed to get cluster status for Elasticsearch: %s/%s with: %s", db.Namespace, db.Name, err.Error())
					// Since the get status failed, skip remaining operations.
					return
				}

				// Update to "Ready" condition to "true" only if the status is "green".
				// For standalone data node cluster (could be combined/topology), consider status "yellow".
				// check if:
				//	( status == green ) || ( status == yellow && (standalone-data-topology || standalone-data-combined))
				if status == api.ElasticsearchStatusGreen ||
					(status == api.ElasticsearchStatusYellow &&
						((db.Spec.Topology != nil && db.Spec.Topology.Data.Replicas != nil && *db.Spec.Topology.Data.Replicas == int32(1)) ||
							(db.Spec.Topology == nil && db.Spec.Replicas != nil && *db.Spec.Replicas == int32(1)))) {
					// Update "Ready" to "true".
					_, err = util.UpdateElasticsearchStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						db.ObjectMeta,
						func(in *api.ElasticsearchStatus) (types.UID, *api.ElasticsearchStatus) {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:               api.DatabaseReady,
									Status:             core.ConditionTrue,
									Reason:             api.ReadinessCheckSucceeded,
									ObservedGeneration: db.Generation,
									Message:            fmt.Sprintf("The Elasticsearch: %s/%s is ready.", db.Namespace, db.Name),
								})
							return db.UID, in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						glog.Errorf("Failed to update status for Elasticsearch: %s/%s with %s", db.Namespace, db.Name, err.Error())
					}
				} else {
					// Update "Ready" to "false".
					_, err = util.UpdateElasticsearchStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						db.ObjectMeta,
						func(in *api.ElasticsearchStatus) (types.UID, *api.ElasticsearchStatus) {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:               api.DatabaseReady,
									Status:             core.ConditionFalse,
									Reason:             api.ReadinessCheckFailed,
									ObservedGeneration: db.Generation,
									Message:            fmt.Sprintf("The Elasticsearch: %s/%s is not ready with cluster status: %s", db.Namespace, db.Name, status),
								})
							return db.UID, in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						glog.Errorf("Failed to update status for Elasticsearch: %s/%s with %s", db.Namespace, db.Name, err.Error())
					}
				}
			}()
		}
		wg.Wait()
	}, c.ReadinessProbeInterval, stopCh)

	// will wait here until stopCh is closed.
	<-stopCh
	glog.Info("Shutting down Elasticsearch health checker...")
}

func (c *Controller) GetElasticsearchClient(db *api.Elasticsearch) (go_es.ESClient, error) {
	url := fmt.Sprintf("%v://%s.%s.svc:%d", db.GetConnectionScheme(), db.ServiceName(), db.GetNamespace(), api.ElasticsearchRestPort)

	// Get original Elasticsearch version, since the client is version specific
	esVersion, err := c.DBClient.CatalogV1alpha1().ElasticsearchVersions().Get(context.TODO(), db.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get elasticsearchVersion")
	}

	dbClient, err := go_es.GetElasticClient(c.Client, db, esVersion.Spec.Version, url)
	if err != nil {
		return nil, err
	}
	return dbClient, nil
}
