/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"errors"
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/pgbouncer/pkg/admission"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Controller) managePgBouncerEvent(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.pgInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}
	if !exists {
		log.Debugf("PgBouncer %s does not exist anymore", key)
		splitKey := strings.Split(key, "/")

		if len(splitKey) != 2 || splitKey[0] == "" || splitKey[1] == "" {
			return errors.New("received unknown key")
		}
		//Now we are interested in this particular secret
		pgbouncerNamespace := splitKey[0]
		pgbouncerName := splitKey[1]
		_, err := c.Client.CoreV1().Secrets(pgbouncerNamespace).Get(pgbouncerName+"-auth", metav1.GetOptions{})
		if err == nil {
			return c.removeDefaultSecret(pgbouncerNamespace, pgbouncerName+"-auth")
		}
		log.Infoln("pgbouncer default secret not found")

	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a PgBouncer was recreated with the same name
		pgbouncer := obj.(*api.PgBouncer).DeepCopy()
		if pgbouncer.DeletionTimestamp != nil {
			if core_util.HasFinalizer(pgbouncer.ObjectMeta, api.GenericKey) {
				if err := c.terminate(pgbouncer); err != nil {
					log.Errorln(err)
					return err
				}
				_, _, err = util.PatchPgBouncer(c.ExtClient.KubedbV1alpha1(), pgbouncer, func(in *api.PgBouncer) *api.PgBouncer {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			pgbouncer, _, err = util.PatchPgBouncer(c.ExtClient.KubedbV1alpha1(), pgbouncer, func(in *api.PgBouncer) *api.PgBouncer {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			if err := c.create(pgbouncer); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(pgbouncer, err.Error())
				return err
			}
		}
	}
	return nil
}

func (c *Controller) create(pgbouncer *api.PgBouncer) error {
	if err := c.manageValidation(pgbouncer); err != nil {
		return err
	}

	if err := c.manageInitialPhase(pgbouncer); err != nil {
		return err
	}
	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, pgbouncer.Namespace); err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, pgbouncer.Namespace, governingService, err)
	}
	// create or patch Service
	if err := c.manageService(pgbouncer); err != nil {
		return err
	}
	// create or patch Fallback Secret
	if err := c.manageDefaultSecret(pgbouncer); err != nil {
		return err
	}
	// create or patch ConfigMap
	if err := c.manageConfigMap(pgbouncer); err != nil {
		return err
	}
	// create or patch Statefulset
	if err := c.manageStatefulSet(pgbouncer); err != nil {
		return err
	}
	// create or patch Stat service
	if err := c.manageStatService(pgbouncer); err != nil {
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
		return err
	}
	return nil
}

func (c *Controller) terminate(pgbouncer *api.PgBouncer) error {
	if pgbouncer.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(pgbouncer); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

//func (c *Controller) removeOwnerReferenceFromOffshoots(pgbouncer *api.PgBouncer, ref *core.ObjectReference) error {
//	// First, Get LabelSelector for Other Components
//	labelSelector := labels.SelectorFromSet(pgbouncer.OffshootSelectors())
//
//	if err := dynamic_util.RemoveOwnerReferenceForSelector(
//		c.DynamicClient,
//		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
//		pgbouncer.Namespace,
//		labelSelector,
//		ref); err != nil {
//		return err
//	}
//	if err := dynamic_util.RemoveOwnerReferenceForSelector(
//		c.DynamicClient,
//		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
//		pgbouncer.Namespace,
//		labelSelector,
//		ref); err != nil {
//		return err
//	}
//	if err := dynamic_util.RemoveOwnerReferenceForItems(
//		c.DynamicClient,
//		core.SchemeGroupVersion.WithResource("secrets"),
//		pgbouncer.Namespace,
//		nil,
//		ref); err != nil {
//		return err
//	}
//	return nil
//}

//func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
//	pgbouncer, err := c.ExtClient.KubedbV1alpha1().PgBouncers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
//	if err != nil {
//		return err
//	}
//	_, err = util.UpdatePgBouncerStatus(c.ExtClient.KubedbV1alpha1(), pgbouncer, func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
//		in.Phase = phase
//		in.Reason = reason
//		return in
//	})
//	return err
//}

//func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
//	pgbouncer, err := c.ExtClient.KubedbV1alpha1().PgBouncers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
//	if err != nil {
//		return err
//	}
//
//	_, _, err = util.PatchPgBouncer(c.ExtClient.KubedbV1alpha1(), pgbouncer, func(in *api.PgBouncer) *api.PgBouncer {
//		in.Annotations = core_util.UpsertMap(in.Annotations, annotation)
//		return in
//	})
//	return err
//}

func (c *Controller) manageValidation(pgbouncer *api.PgBouncer) error {
	if err := validator.ValidatePgBouncer(c.Client, c.ExtClient, pgbouncer, true); err != nil {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		return nil // user error so just record error and don't retry.
	}

	// Check if usrlist is absent.
	if pgbouncer.Spec.UserListSecretRef != nil && pgbouncer.Spec.UserListSecretRef.Name != "" {
		if pgbouncer.Spec.ConnectionPool != nil && pgbouncer.Spec.ConnectionPool.AuthType != "any" {
			if _, err := c.Client.CoreV1().Secrets(pgbouncer.GetNamespace()).Get(pgbouncer.Spec.UserListSecretRef.Name, metav1.GetOptions{}); err != nil {
				c.recorder.Eventf(
					pgbouncer,
					core.EventTypeWarning,
					"UserListMissing",
					"userlist secret %s not found", pgbouncer.Spec.UserListSecretRef.Name)
			}
		}
	}
	return nil //if no err
}

func (c *Controller) manageInitialPhase(pgbouncer *api.PgBouncer) error {
	if pgbouncer.Status.Phase == "" {
		pg, err := util.UpdatePgBouncerStatus(c.ExtClient.KubedbV1alpha1(), pgbouncer, func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			return err
		}
		pgbouncer.Status = pg.Status
	}
	return nil //if no err
}

func (c *Controller) manageFinalPhase(pgbouncer *api.PgBouncer) error {
	if _, err := meta_util.GetString(pgbouncer.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound {
		if pgbouncer.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}
		// add phase that database is being initialized
		pg, err := util.UpdatePgBouncerStatus(c.ExtClient.KubedbV1alpha1(), pgbouncer, func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			return err
		}
		pgbouncer.Status = pg.Status
	}
	pg, err := util.UpdatePgBouncerStatus(c.ExtClient.KubedbV1alpha1(), pgbouncer, func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(pgbouncer.Generation, meta_util.GenerationHash(pgbouncer))
		return in
	})
	if err != nil {
		return err
	}
	pgbouncer.Status = pg.Status
	return nil //if no err
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
	return nil //if no err
}

func (c *Controller) manageConfigMap(pgbouncer *api.PgBouncer) error {
	configMapVerb, err := c.ensureConfigMapFromCRD(pgbouncer)
	if err != nil {
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

	return nil //if no err
}

func (c *Controller) manageStatefulSet(pgbouncer *api.PgBouncer) error {
	pgBouncerVersion, err := c.ExtClient.CatalogV1alpha1().PgBouncerVersions().Get(string(pgbouncer.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}

	statefulsetVerb, err := c.ensureStatefulSet(pgbouncer, pgBouncerVersion, []core.EnvVar{})
	if err != nil {
		return err
	}
	if statefulsetVerb == kutil.VerbCreated {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PgBouncer statefulset",
		)
	} else if statefulsetVerb == kutil.VerbPatched {
		c.recorder.Event(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PgBouncer statefulset",
		)
	}
	if statefulsetVerb != kutil.VerbUnchanged {
		log.Infoln("Statefulset ", statefulsetVerb)
	}
	return nil //if no err
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
	return nil //if no err
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
	return nil //if no err
}

func (c *Controller) getVolumeAndVolumeMountForDefaultUserList(pgbouncer *api.PgBouncer) (*core.Volume, *core.VolumeMount, error) {
	fSecret := c.GetDefaultSecretSpec(pgbouncer)
	_, err := c.Client.CoreV1().Secrets(fSecret.Namespace).Get(fSecret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	secretVolume := &core.Volume{
		Name: "fallback-userlist",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: fSecret.Name,
			},
		},
	}
	//Add to volumeMounts to mount the volume
	secretVolumeMount := &core.VolumeMount{
		Name:      "fallback-userlist",
		MountPath: userListMountPath,
		ReadOnly:  true,
	}

	return secretVolume, secretVolumeMount, nil //if no err
}

//func (c *Controller) manageTemPlate(pgbouncer *api.PgBouncer) error {
//
//	return nil //if no err
//}
