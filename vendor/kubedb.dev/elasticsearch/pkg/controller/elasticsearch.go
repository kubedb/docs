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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/elasticsearch/pkg/admission"
	"kubedb.dev/elasticsearch/pkg/distribution"

	"github.com/appscode/go/log"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) create(elasticsearch *api.Elasticsearch) error {
	if err := validator.ValidateElasticsearch(c.Client, c.DBClient, elasticsearch, true); err != nil {
		c.Recorder.Event(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil
	}

	// create Governing Service
	if err := c.ensureElasticGvrSvc(elasticsearch); err != nil {
		return fmt.Errorf(`failed to create governing Service for "%v/%v". Reason: %v`, elasticsearch.Namespace, elasticsearch.Name, err)
	}

	// ensure database Service
	vt1, err := c.ensureService(elasticsearch)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	elasticsearch, vt2, err := c.ensureElasticsearchNode(elasticsearch)
	if err != nil {
		return err
	}

	// If both err==nil & elasticsearch == nil,
	// the object was dropped from the work-queue, to process later.
	// return nil.
	if elasticsearch == nil {
		return nil
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.Recorder.Event(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Elasticsearch",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.Recorder.Event(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Elasticsearch",
		)
	}

	// ensure appbinding before ensuring Restic scheduler and restore
	_, err = c.ensureAppBinding(elasticsearch)
	if err != nil {
		log.Errorln(err)
		return err
	}

	//======================== Wait for the initial restore =====================================
	if elasticsearch.Spec.Init != nil && elasticsearch.Spec.Init.WaitForInitialRestore {
		// Only wait for the first restore.
		// For initial restore,  elasticsearch.Spec.Init.Initialized will be "false" and "DataRestored" condition either won't exist or will be "False".
		if !elasticsearch.Spec.Init.Initialized &&
			!kmapi.IsConditionTrue(elasticsearch.Status.Conditions, api.DatabaseDataRestored) {
			// write log indicating that the database is waiting for the data to be restored by external initializer
			log.Infof("Database %s %s/%s is waiting for data to be restored by external initializer",
				elasticsearch.Kind,
				elasticsearch.Namespace,
				elasticsearch.Name,
			)
			// Rest of the processing will execute after the the restore process completed. So, just return for now.
			return nil
		}
	}

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(elasticsearch); err != nil {
		c.Recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(elasticsearch); err != nil {
		c.Recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	// Check: ReplicaReady --> AcceptingConnection --> Ready --> Provisioned
	// If spec.Init.WaitForInitialRestore is true, but data wasn't restored successfully,
	// process won't reach here (returned nil at the beginning). As it is here, that means data was restored successfully.
	// No need to check for IsConditionTrue(DataRestored).
	if kmapi.IsConditionTrue(elasticsearch.Status.Conditions, api.DatabaseReplicaReady) &&
		kmapi.IsConditionTrue(elasticsearch.Status.Conditions, api.DatabaseAcceptingConnection) &&
		kmapi.IsConditionTrue(elasticsearch.Status.Conditions, api.DatabaseReady) &&
		!kmapi.IsConditionTrue(elasticsearch.Status.Conditions, api.DatabaseProvisioned) {
		_, err := util.UpdateElasticsearchStatus(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			elasticsearch.ObjectMeta,
			func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
				in.Conditions = kmapi.SetCondition(in.Conditions,
					kmapi.Condition{
						Type:               api.DatabaseProvisioned,
						Status:             core.ConditionTrue,
						Reason:             api.DatabaseSuccessfullyProvisioned,
						ObservedGeneration: elasticsearch.Generation,
						Message:            fmt.Sprintf("The Elasticsearch: %s/%s is successfully provisioned.", elasticsearch.Namespace, elasticsearch.Name),
					})
				return in
			},
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
	}

	// If the database is successfully provisioned,
	// Set spec.Init.Initialized to true, if init!=nil.
	// This will prevent the operator from re-initializing the database.
	if elasticsearch.Spec.Init != nil &&
		!elasticsearch.Spec.Init.Initialized &&
		kmapi.IsConditionTrue(elasticsearch.Status.Conditions, api.DatabaseProvisioned) {
		_, _, err := util.CreateOrPatchElasticsearch(context.TODO(), c.DBClient.KubedbV1alpha2(), elasticsearch.ObjectMeta, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.Init.Initialized = true
			return in
		}, metav1.PatchOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) ensureElasticsearchNode(es *api.Elasticsearch) (*api.Elasticsearch, kutil.VerbType, error) {
	if es == nil {
		return nil, kutil.VerbUnchanged, errors.New("Elasticsearch object is empty")
	}

	elastic, err := distribution.NewElasticsearch(c.Client, c.DBClient, es)
	if err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to get elasticsearch distribution")
	}

	// Create/sync certificate secrets
	// But if  the tls.issuerRef is set, do nothing (i.e. should be handled from enterprise operator).
	if err = elastic.EnsureCertSecrets(); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to ensure certificates secret")
	}

	// Create/sync user credential (ie. username, password) secrets
	if err = elastic.EnsureAuthSecret(); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to ensure database credential secret")
	}

	// Get the cert secret names
	// List varies depending on the elasticsearch distribution & configuration.
	sNames := elastic.RequiredCertSecretNames()
	// Check whether the secrets are available or not.
	ok, err := dynamic_util.ResourcesExists(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		es.Namespace,
		sNames...,
	)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if !ok {
		// If the certificates are managed by the enterprise operator,
		// It takes some time for the secrets to get ready.
		// If any required secret is yet to get ready,
		// drop the elasticsearch object from work queue (i.e. return nil with no error).
		// When any secret owned by this elasticsearch object is created/updated,
		// this elasticsearch object will be enqueued again for processing.
		log.Infoln(fmt.Sprintf("Required secrets for Elasticsearch: %s/%s are not ready yet", es.Namespace, es.Name))
		return nil, kutil.VerbUnchanged, nil
	}

	if err = elastic.EnsureDefaultConfig(); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to ensure default configuration for elasticsearch")
	}

	// Ensure Service account, role, rolebinding, and PSP for database statefulsets
	if err := c.ensureDatabaseRBAC(elastic.UpdatedElasticsearch()); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to create RBAC role or roleBinding")
	}

	vt := kutil.VerbUnchanged
	topology := elastic.UpdatedElasticsearch().Spec.Topology
	if topology != nil {
		vt1, err := elastic.EnsureIngestNodes()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		vt2, err := elastic.EnsureMasterNodes()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		vt3, err := elastic.EnsureDataNodes()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}

		if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated && vt3 == kutil.VerbCreated {
			vt = kutil.VerbCreated
		} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched || vt3 == kutil.VerbPatched {
			vt = kutil.VerbPatched
		}
	} else {
		vt, err = elastic.EnsureCombinedNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
	}

	return elastic.UpdatedElasticsearch(), vt, nil
}

func (c *Controller) halt(db *api.Elasticsearch) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	glog.Infof("Elasticsearch %v/%v is halting...", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilHalted(db); err != nil {
		return err
	}
	glog.Infof("Elasticsearch %v/%v is Halted.", db.Namespace, db.Name)
	if _, err := util.UpdateElasticsearchStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
			in.Conditions = kmapi.SetCondition(in.Conditions, kmapi.Condition{
				Type:               api.DatabaseHalted,
				Status:             core.ConditionTrue,
				Reason:             api.DatabaseHaltedSuccessfully,
				ObservedGeneration: db.Generation,
				Message:            fmt.Sprintf("Elasticseach %s/%s successfully halted.", db.Namespace, db.Name),
			})
			// make "AcceptingConnection" and "Ready" conditions false.
			// Because these are handled from health checker at a certain interval,
			// if consecutive halt and un-halt occurs in the meantime,
			// phase might still be on the "Ready" state.
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseAcceptingConnection,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseHaltedSuccessfully,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Elasticsearch: %s/%s is not accepting client requests.", db.Namespace, db.Name),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseHaltedSuccessfully,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The Elasticsearch: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return in
		},
		metav1.UpdateOptions{},
	); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(elasticsearch *api.Elasticsearch) error {
	owner := metav1.NewControllerRef(elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	// If TerminationPolicy is "halt", keep PVCs,Secrets intact.
	// TerminationPolicyPause is deprecated and will be removed in future.
	if elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		if err := c.removeOwnerReferenceFromOffshoots(elasticsearch); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(elasticsearch, owner); err != nil {
			return err
		}
	}

	if elasticsearch.Spec.Monitor != nil {
		if err := c.deleteMonitor(elasticsearch); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(elasticsearch *api.Elasticsearch, owner *metav1.OwnerReference) error {
	selector := labels.SelectorFromSet(elasticsearch.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := c.wipeOutDatabase(elasticsearch.ObjectMeta, elasticsearch.GetPersistentSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			context.TODO(),
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			elasticsearch.Namespace,
			elasticsearch.GetPersistentSecrets(),
			elasticsearch); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		elasticsearch.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(elasticsearch *api.Elasticsearch) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(elasticsearch.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		elasticsearch.Namespace,
		labelSelector,
		elasticsearch); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		elasticsearch.Namespace,
		elasticsearch.GetPersistentSecrets(),
		elasticsearch); err != nil {
		return err
	}
	return nil
}
