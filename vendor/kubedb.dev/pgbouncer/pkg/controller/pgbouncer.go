/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

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
	validator "kubedb.dev/pgbouncer/pkg/admission"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	meta_util "kmodules.xyz/client-go/meta"
)

const (
	AuthSecretSuffix = "-auth"
)

func (c *Controller) managePgBouncerEvent(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.pbInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Infof("PgBouncer %s does not exist anymore\n", key)
		log.Debugf("PgBouncer %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a PgBouncer was recreated with the same name
		pgbouncer := obj.(*api.PgBouncer).DeepCopy()

		if err := c.manageCreateOrPatchEvent(pgbouncer); err != nil {
			log.Errorln(err)
			c.pushFailureEvent(pgbouncer, err.Error())
			return err
		}
	}
	return nil
}

func (c *Controller) manageCreateOrPatchEvent(pgbouncer *api.PgBouncer) error {
	if err := c.manageValidation(pgbouncer); err != nil {
		log.Infoln(err)
		return nil // user err, dont' retry.
	}
	if err := c.manageInitialPhase(pgbouncer); err != nil {
		log.Infoln(err)
		return err
	}
	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, pgbouncer.Namespace); err != nil {
		log.Infoln(err)
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, pgbouncer.Namespace, governingService, err)
	}
	// create or patch Service
	if err := c.manageService(pgbouncer); err != nil {
		log.Infoln(err)
		return err
	}
	// create or patch default Secret
	if err := c.manageDefaultSecret(pgbouncer); err != nil {
		log.Infoln(err)
		return err
	}
	// create or patch ConfigMap
	if err := c.manageConfigMap(pgbouncer); err != nil {
		log.Infoln(err)
		return err
	}
	// wait for certificates
	if pgbouncer.Spec.TLS != nil {
		// wait for serving certificate
		if _, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(context.TODO(), pgbouncer.Name+api.PgBouncerServingServerSuffix, metav1.GetOptions{}); kerr.IsNotFound(err) {
			return nil
		}

		// wait for serving client certificate
		if _, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(context.TODO(), pgbouncer.Name+api.PgBouncerServingClientSuffix, metav1.GetOptions{}); kerr.IsNotFound(err) {
			return nil
		}

		// wait for exporter client certificate
		if _, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(context.TODO(), pgbouncer.Name+api.PgBouncerExporterClientCertSuffix, metav1.GetOptions{}); kerr.IsNotFound(err) {
			return nil
		}
	}
	// create or patch StatefulSet
	if err := c.manageStatefulSet(pgbouncer); err != nil {
		log.Infoln(err)
		return err
	}
	// create or patch Stat service
	if err := c.manageStatService(pgbouncer); err != nil {
		log.Infoln(err)
		return err
	}

	if err := c.manageMonitor(pgbouncer); err != nil {
		c.recorder.Eventf(
			pgbouncer,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	// Add initialized or running phase
	if err := c.manageFinalPhase(pgbouncer); err != nil {
		log.Infoln(err)
		return err
	}
	return nil
}

func (c *Controller) manageValidation(pgbouncer *api.PgBouncer) error {
	if err := validator.ValidatePgBouncer(c.Client, c.ExtClient, pgbouncer, true); err != nil {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return err // user error so just record error and don't retry.
	}

	// Check if userList is absent.
	if pgbouncer.Spec.UserListSecretRef != nil && pgbouncer.Spec.UserListSecretRef.Name != "" {
		if pgbouncer.Spec.ConnectionPool != nil && pgbouncer.Spec.ConnectionPool.AuthType != "any" {
			if _, err := c.Client.CoreV1().Secrets(pgbouncer.GetNamespace()).Get(context.TODO(), pgbouncer.Spec.UserListSecretRef.Name, metav1.GetOptions{}); err != nil {
				c.recorder.Eventf(
					pgbouncer,
					core.EventTypeWarning,
					"UserListMissing",
					"user-list secret %s not found", pgbouncer.Spec.UserListSecretRef.Name)
			}
		}
	}

	return nil
}

func (c *Controller) manageInitialPhase(pgbouncer *api.PgBouncer) error {
	if pgbouncer.Status.Phase == "" {
		pg, err := util.UpdatePgBouncerStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), pgbouncer.ObjectMeta, func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		pgbouncer.Status = pg.Status
	}
	return nil
}

func (c *Controller) manageFinalPhase(pgbouncer *api.PgBouncer) error {
	if !c.isPgBouncerExist(pgbouncer) {
		return nil
	}

	if _, err := meta_util.GetString(pgbouncer.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound {
		if pgbouncer.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}
		// add to phase that PgBouncer is being initialized
		pg, err := util.UpdatePgBouncerStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), pgbouncer.ObjectMeta, func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			log.Infoln(err)
			return err
		}
		pgbouncer.Status = pg.Status
	}
	pg, err := util.UpdatePgBouncerStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), pgbouncer.ObjectMeta, func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = pgbouncer.Generation
		return in
	}, metav1.UpdateOptions{})
	if err != nil {
		log.Infoln(err)
		return err
	}
	pgbouncer.Status = pg.Status
	return nil
}

func (c *Controller) manageDefaultSecret(pgbouncer *api.PgBouncer) error {
	sVerb, err := c.CreateOrPatchDefaultSecret(pgbouncer)
	if err != nil {
		return err
	}

	if sVerb == kutil.VerbCreated {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PgBouncer Fallback Secret",
		)
	} else if sVerb == kutil.VerbPatched {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PgBouncer Fallback Secret",
		)
	}
	if sVerb != kutil.VerbUnchanged {
		log.Infoln("Default secret ", sVerb)
	}
	return nil
}

func (c *Controller) manageConfigMap(pgbouncer *api.PgBouncer) error {
	configMapVerb, err := c.ensureConfigMapFromCRD(pgbouncer)
	if err != nil {
		log.Infoln(err)
		return err
	}

	if configMapVerb == kutil.VerbCreated {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PgBouncer configMap",
		)
	} else if configMapVerb == kutil.VerbPatched {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PgBouncer configMap",
		)
	}
	if configMapVerb != kutil.VerbUnchanged {
		log.Infoln("ConfigMap ", configMapVerb)
	}

	return nil
}

func (c *Controller) manageStatefulSet(pgbouncer *api.PgBouncer) error {
	pgBouncerVersion, err := c.ExtClient.CatalogV1alpha1().PgBouncerVersions().Get(context.TODO(), pgbouncer.Spec.Version, metav1.GetOptions{})
	if err != nil {
		log.Infoln(err)
		return err
	}

	statefulSetVerb, err := c.ensureStatefulSet(pgbouncer, pgBouncerVersion, []core.EnvVar{})
	if err != nil {
		log.Infoln(err)
		return err
	}
	if statefulSetVerb == kutil.VerbCreated {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PgBouncer statefulset",
		)
	} else if statefulSetVerb == kutil.VerbPatched {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PgBouncer statefulset",
		)
	}
	if statefulSetVerb != kutil.VerbUnchanged {
		log.Infoln("Statefulset ", statefulSetVerb)
	}
	return nil
}

func (c *Controller) manageService(pgbouncer *api.PgBouncer) error {
	serviceVerb, err := c.ensureService(pgbouncer)
	if err != nil {
		return err
	}
	if serviceVerb == kutil.VerbCreated {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Service",
		)
	} else if serviceVerb == kutil.VerbPatched {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Service",
		)
	}
	if serviceVerb != kutil.VerbUnchanged {
		log.Infoln("Service ", serviceVerb)
	}
	return nil
}

func (c *Controller) manageStatService(pgbouncer *api.PgBouncer) error {
	statServiceVerb, err := c.ensureStatsService(pgbouncer)
	if err != nil {
		return err
	}
	if statServiceVerb == kutil.VerbCreated {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Stat Service",
		)
	} else if statServiceVerb == kutil.VerbPatched {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Stat Service",
		)
	}
	if statServiceVerb != kutil.VerbUnchanged {
		log.Infoln("Stat Service ", statServiceVerb)
	}
	return nil
}

func (c *Controller) isPgBouncerExist(pgbouncer *api.PgBouncer) bool {
	_, err := c.ExtClient.KubedbV1alpha1().PgBouncers(pgbouncer.Namespace).Get(context.TODO(), pgbouncer.Name, metav1.GetOptions{})
	return err == nil
}
