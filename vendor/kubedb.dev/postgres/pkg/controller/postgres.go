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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/postgres/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) create(postgres *api.Postgres) error {
	if err := validator.ValidatePostgres(c.Client, c.DBClient, postgres, true); err != nil {
		c.Recorder.Event(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	if postgres.Status.Phase == "" {
		pg, err := util.UpdatePostgresStatus(context.TODO(), c.DBClient.KubedbV1alpha1(), postgres.ObjectMeta, func(in *api.PostgresStatus) *api.PostgresStatus {
			in.Phase = api.DatabasePhaseProvisioning
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		postgres.Status = pg.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, postgres.Namespace); err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, postgres.Namespace, governingService, err)
	}

	// ensure database Service
	vt1, err := c.ensureService(postgres)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	postgresVersion, err := c.DBClient.CatalogV1alpha1().PostgresVersions().Get(context.TODO(), string(postgres.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}
	vt2, err := c.ensurePostgresNode(postgres, postgresVersion)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.Recorder.Event(
			postgres,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Postgres",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.Recorder.Event(
			postgres,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Postgres",
		)
	}

	// ensure appbinding before ensuring Restic scheduler and restore
	_, err = c.ensureAppBinding(postgres, postgresVersion)
	if err != nil {
		log.Errorln(err)
		return err
	}

	//======================== Wait for the initial restore =====================================
	if postgres.Spec.Init != nil && postgres.Spec.Init.WaitForInitialRestore {
		// Only wait for the first restore.
		// For initial restore, "Provisioned" condition won't exist and "DataRestored" condition either won't exist or will be "False".
		if !kmapi.HasCondition(postgres.Status.Conditions, api.DatabaseProvisioned) &&
			!kmapi.IsConditionTrue(postgres.Status.Conditions, api.DatabaseDataRestored) {
			// write log indicating that the database is waiting for the data to be restored by external initializer
			log.Infof("Database %s %s/%s is waiting for data to be restored by external initializer",
				postgres.Kind,
				postgres.Namespace,
				postgres.Name,
			)
			// Rest of the processing will execute after the the restore process completed. So, just return for now.
			return nil
		}
	}

	pg, err := util.UpdatePostgresStatus(context.TODO(), c.DBClient.KubedbV1alpha1(), postgres.ObjectMeta, func(in *api.PostgresStatus) *api.PostgresStatus {
		in.Phase = api.DatabasePhaseReady
		in.ObservedGeneration = postgres.Generation
		return in
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	postgres.Status = pg.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(postgres); err != nil {
		c.Recorder.Eventf(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(postgres); err != nil {
		c.Recorder.Eventf(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	return nil
}

func (c *Controller) ensurePostgresNode(postgres *api.Postgres, postgresVersion *catalog.PostgresVersion) (kutil.VerbType, error) {
	var err error

	if err = c.ensureDatabaseSecret(postgres); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Ensure Service account, role, rolebinding, and PSP for database statefulsets
	if err := c.ensureDatabaseRBAC(postgres); err != nil {
		return kutil.VerbUnchanged, err
	}

	vt, err := c.ensureCombinedNode(postgres, postgresVersion)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	return vt, nil
}

func (c *Controller) halt(db *api.Postgres) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	log.Infof("Halting Postgres %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilPaused(db); err != nil {
		return err
	}
	log.Infof("update status of Postgres %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdatePostgresStatus(context.TODO(), c.DBClient.KubedbV1alpha1(), db.ObjectMeta, func(in *api.PostgresStatus) *api.PostgresStatus {
		in.Phase = api.DatabasePhaseHalted
		in.ObservedGeneration = db.Generation
		return in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(postgres *api.Postgres) error {
	owner := metav1.NewControllerRef(postgres, api.SchemeGroupVersion.WithKind(api.ResourceKindPostgres))

	// If TerminationPolicy is "halt", keep PVCs and Secrets intact.
	// TerminationPolicyPause is deprecated and will be removed in future.
	if postgres.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		if err := c.removeOwnerReferenceFromOffshoots(postgres); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots,WAL-data).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets, wal-data intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(postgres, owner); err != nil {
			return err
		}
	}

	if postgres.Spec.Monitor != nil {
		if err := c.deleteMonitor(postgres); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(postgres *api.Postgres, owner *metav1.OwnerReference) error {
	selector := labels.SelectorFromSet(postgres.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if postgres.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		// at first, pause the database transactions by deleting the statefulsets. otherwise wiping out may not be accurate.
		// because, while operator is trying to delete the wal data, the database pod may still trying to push new data.
		policy := metav1.DeletePropagationForeground
		if err := c.Client.
			AppsV1().
			StatefulSets(postgres.Namespace).
			DeleteCollection(
				context.TODO(),
				metav1.DeleteOptions{PropagationPolicy: &policy},
				metav1.ListOptions{LabelSelector: selector.String()},
			); err != nil && !kerr.IsNotFound(err) {
			return errors.Wrap(err, "error in deletion of statefulsets")
		}
		// Let's give statefulsets some time to breath and then be deleted.
		if err := wait.PollImmediate(kutil.RetryInterval, kutil.GCTimeout, func() (bool, error) {
			podList, err := c.Client.CoreV1().Pods(postgres.Namespace).List(context.TODO(), metav1.ListOptions{
				LabelSelector: selector.String(),
			})
			return len(podList.Items) == 0, err
		}); err != nil {
			fmt.Printf("got error while waiting for db pods to be deleted: %v. coninuing with further deletion steps.\n", err.Error())
		}
		if err := c.wipeOutDatabase(postgres.ObjectMeta, postgres.Spec.GetSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
		// if wal archiver was configured, remove wal data from backend
		if postgres.Spec.Archiver != nil {
			if err := c.wipeOutWalData(postgres.ObjectMeta, &postgres.Spec); err != nil {
				return err
			}
		}
	} else {
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			context.TODO(),
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			postgres.Namespace,
			postgres.Spec.GetSecrets(),
			postgres); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		postgres.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(postgres *api.Postgres) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(postgres.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		postgres.Namespace,
		labelSelector,
		postgres); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		postgres.Namespace,
		postgres.Spec.GetSecrets(),
		postgres); err != nil {
		return err
	}
	return nil
}
