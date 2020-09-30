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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/mysql/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Controller) create(mysql *api.MySQL) error {
	if err := validator.ValidateMySQL(c.Client, c.ExtClient, mysql, true); err != nil {
		c.Recorder.Event(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil
	}

	if mysql.Status.Phase == "" {
		my, err := util.UpdateMySQLStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), mysql.ObjectMeta, func(in *api.MySQLStatus) *api.MySQLStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		mysql.Status = my.Status
	}

	// create Governing Service
	governingService, err := c.createMySQLGoverningService(mysql)
	if err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, mysql.Namespace, governingService, err)
	}

	// Ensure Service account, role, rolebinding, and PSP for database statefulsets
	if err := c.ensureDatabaseRBAC(mysql); err != nil {
		return err
	}

	// ensure database Service
	vt1, err := c.ensurePrimaryService(mysql)
	if err != nil {
		return err
	}

	// create Service only for master/primary pod
	if mysql.UsesGroupReplication() {
		vt, err := c.ensureSecondaryService(mysql)
		if err != nil {
			return err
		}
		if vt == kutil.VerbCreated {
			c.Recorder.Event(
				mysql,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully created service for secondary replicas",
			)
		} else if vt == kutil.VerbPatched {
			c.Recorder.Event(
				mysql,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully patched service for secondary replicas",
			)
		}
	}

	if err := c.ensureDatabaseSecret(mysql); err != nil {
		return err
	}

	// wait for certificates
	if mysql.Spec.TLS != nil && mysql.Spec.TLS.IssuerRef != nil {
		ok, err := dynamic_util.ResourcesExists(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			mysql.Namespace,
			mysql.MustCertSecretName(api.MySQLServerCert),
			mysql.MustCertSecretName(api.MySQLClientCert),
			mysql.MustCertSecretName(api.MySQLMetricsExporterCert),
			meta_util.NameWithSuffix(mysql.Name, api.MySQLMetricsExporterConfigSecretSuffix),
		)
		if err != nil {
			return err
		}
		if !ok {
			log.Infoln(fmt.Sprintf("wait for all necessary secrets for mysql %s/%s", mysql.Namespace, mysql.Name))
			return nil
		}
	}

	// ensure database StatefulSet
	vt2, err := c.ensureStatefulSet(mysql)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.Recorder.Event(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created MySQL",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.Recorder.Event(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched MySQL",
		)
	}

	// ensure appbinding before ensuring Restic scheduler and restore
	_, err = c.ensureAppBinding(mysql)
	if err != nil {
		log.Errorln(err)
		return err
	}

	if mysql.Spec.Init != nil && mysql.Spec.Init.Initializer != nil {
		// If "Initialized" condition is not present, it means restore process hasn't completed yet.
		// In this case, make database phase "Initializing".
		if !kmapi.HasCondition(mysql.Status.Conditions, api.DatabaseInitialized) {
			mysql, err := util.UpdateMySQLStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), mysql.ObjectMeta, func(in *api.MySQLStatus) *api.MySQLStatus {
				in.Phase = api.DatabasePhaseInitializing
				in.ObservedGeneration = mysql.Generation
				return in
			}, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			// write log indicating that the database is waiting for the data to be restored by external initializer
			log.Infof("Database %s %s/%s is waiting for data to be restored by initializer %s/%s/%s",
				mysql.Kind,
				mysql.Namespace,
				mysql.Name,
				*mysql.Spec.Init.Initializer.APIGroup,
				mysql.Spec.Init.Initializer.Kind,
				mysql.Spec.Init.Initializer.Name,
			)
			// Rest of the processing will execute after the the restore process completed. So, just return for now.
			return nil
		} else {
			// Restore process has completed. It has either succeeded or failed. Update database phase accordingly.
			dbPhase := api.DatabasePhaseRunning
			if !kmapi.IsConditionTrue(mysql.Status.Conditions, api.DatabaseInitialized) {
				dbPhase = api.DatabasePhaseFailed
			}
			mysql, err = util.UpdateMySQLStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), mysql.ObjectMeta, func(in *api.MySQLStatus) *api.MySQLStatus {
				in.Phase = dbPhase
				in.ObservedGeneration = mysql.Generation
				return in
			}, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	}

	my, err := util.UpdateMySQLStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), mysql.ObjectMeta, func(in *api.MySQLStatus) *api.MySQLStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = mysql.Generation
		return in
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	mysql.Status = my.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(mysql); err != nil {
		c.Recorder.Eventf(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(mysql); err != nil {
		c.Recorder.Eventf(
			mysql,
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

func (c *Controller) halt(db *api.MySQL) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	log.Infof("Halting MySQL %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilHalted(db); err != nil {
		return err
	}
	log.Infof("update status of MySQL %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdateMySQLStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), db.ObjectMeta, func(in *api.MySQLStatus) *api.MySQLStatus {
		in.Phase = api.DatabasePhaseHalted
		in.ObservedGeneration = db.Generation
		return in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(mysql *api.MySQL) error {
	owner := metav1.NewControllerRef(mysql, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	// If TerminationPolicy is "halt", keep PVCs and Secrets intact.
	// TerminationPolicyHalt is deprecated and will be removed in future.
	if mysql.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		if err := c.removeOwnerReferenceFromOffshoots(mysql); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(mysql, owner); err != nil {
			return err
		}
	}

	if mysql.Spec.Monitor != nil {
		if err := c.deleteMonitor(mysql); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(mysql *api.MySQL, owner *metav1.OwnerReference) error {
	selector := labels.SelectorFromSet(mysql.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if mysql.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := c.wipeOutDatabase(mysql.ObjectMeta, mysql.Spec.GetSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			context.TODO(),
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			mysql.Namespace,
			mysql.Spec.GetSecrets(),
			mysql); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mysql.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(mysql *api.MySQL) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(mysql.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mysql.Namespace,
		labelSelector,
		mysql); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		mysql.Namespace,
		mysql.Spec.GetSecrets(),
		mysql); err != nil {
		return err
	}
	return nil
}
