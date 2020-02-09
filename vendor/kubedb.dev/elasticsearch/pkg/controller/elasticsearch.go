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
	"fmt"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/elasticsearch/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	policy_util "kmodules.xyz/client-go/policy/v1beta1"
)

func (c *Controller) create(elasticsearch *api.Elasticsearch) error {
	if err := validator.ValidateElasticsearch(c.Client, c.ExtClient, elasticsearch, true); err != nil {
		c.recorder.Event(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil
	}

	if elasticsearch.Status.Phase == "" {
		es, err := util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			return err
		}
		elasticsearch.Status = es.Status
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
	vt2, err := c.ensureElasticsearchNode(elasticsearch)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Elasticsearch",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
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

	if _, err := meta_util.GetString(elasticsearch.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		elasticsearch.Spec.Init != nil && elasticsearch.Spec.Init.StashRestoreSession != nil {

		if elasticsearch.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		// add phase that database is being initialized
		mg, err := util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			return err
		}
		elasticsearch.Status = mg.Status

		init := elasticsearch.Spec.Init
		if init.StashRestoreSession != nil {
			log.Debugf("Elasticsearch %v/%v is waiting for restoreSession to be succeeded", elasticsearch.Namespace, elasticsearch.Name)
			return nil
		}
	}

	es, err := util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = elasticsearch.Generation
		return in
	})
	if err != nil {
		return err
	}
	elasticsearch.Status = es.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(elasticsearch); err != nil {
		c.recorder.Eventf(
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
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	return nil
}

func (c *Controller) ensureElasticsearchNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	var err error

	if err = c.ensureCertSecret(elasticsearch); err != nil {
		return kutil.VerbUnchanged, err
	}
	if err = c.ensureDatabaseSecret(elasticsearch); err != nil {
		return kutil.VerbUnchanged, err
	}
	if err = c.ensureDatabaseConfigForXPack(elasticsearch); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Ensure Service account, role, rolebinding, and PSP for database statefulsets
	if err := c.ensureDatabaseRBAC(elasticsearch); err != nil {
		return kutil.VerbUnchanged, err
	}

	vt := kutil.VerbUnchanged
	topology := elasticsearch.Spec.Topology
	if topology != nil {
		vt1, err := c.ensureClientNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}
		vt2, err := c.ensureMasterNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}
		vt3, err := c.ensureDataNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}

		if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated && vt3 == kutil.VerbCreated {
			vt = kutil.VerbCreated
		} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched || vt3 == kutil.VerbPatched {
			vt = kutil.VerbPatched
		}
	} else {
		vt, err = c.ensureCombinedNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}
	}

	// Need some time to build elasticsearch cluster. Nodes will communicate with each other
	// TODO: find better way
	time.Sleep(time.Second * 30)

	return vt, nil
}

func (c *Controller) halt(db *api.Elasticsearch) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	log.Infof("Halting Elasticsearch %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilPaused(db); err != nil {
		return err
	}
	log.Infof("update status of Elasticsearch %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), db, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = api.DatabasePhaseHalted
		in.ObservedGeneration = db.Generation
		return in
	}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(elasticsearch *api.Elasticsearch) error {
	owner := metav1.NewControllerRef(elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	// If TerminationPolicy is "halt", keep PVCs,Secrets intact.
	// TerminationPolicyPause is deprecated and will be removed in future.
	if elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyHalt || elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyPause {
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
		if err := c.wipeOutDatabase(elasticsearch.ObjectMeta, elasticsearch.Spec.GetSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			elasticsearch.Namespace,
			elasticsearch.Spec.GetSecrets(),
			elasticsearch); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
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
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		elasticsearch.Namespace,
		labelSelector,
		elasticsearch); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		elasticsearch.Namespace,
		elasticsearch.Spec.GetSecrets(),
		elasticsearch); err != nil {
		return err
	}
	return nil
}

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	elasticsearch, err := c.esLister.Elasticsearches(meta.Namespace).Get(meta.Name)
	if err != nil {
		return nil, err
	}

	return elasticsearch, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	elasticsearch, err := c.esLister.Elasticsearches(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}
	_, err = util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	})
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	elasticsearch, err := c.esLister.Elasticsearches(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}

	_, _, err = util.PatchElasticsearch(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Annotations = core_util.UpsertMap(in.Annotations, annotation)
		return in
	})
	return err
}

func (c *Controller) createPodDisruptionBudget(sts *appsv1.StatefulSet, maxUnavailable *intstr.IntOrString) error {
	owner := metav1.NewControllerRef(sts, appsv1.SchemeGroupVersion.WithKind("StatefulSet"))

	m := metav1.ObjectMeta{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}
	_, _, err := policy_util.CreateOrPatchPodDisruptionBudget(c.Client, m,
		func(in *policyv1beta1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
			in.Labels = sts.Labels
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: sts.Spec.Template.Labels,
			}

			in.Spec.MaxUnavailable = maxUnavailable

			in.Spec.MinAvailable = nil
			return in
		})
	return err
}
