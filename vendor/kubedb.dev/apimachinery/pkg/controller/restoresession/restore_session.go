/*
Copyright AppsCode Inc. and Contributors

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

package restoresession

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
)

func (c *Controller) handleRestoreSession(rs *v1beta1.RestoreSession) error {
	if rs.Status.Phase != v1beta1.RestoreSessionSucceeded && rs.Status.Phase != v1beta1.RestoreSessionFailed {
		log.Debugf("restoreSession %v/%v is not any of 'succeeded' or 'failed' ", rs.Namespace, rs.Name)
		return nil
	}

	if rs.Spec.Target == nil {
		log.Debugf("restoreSession %v/%v does not have spec.target set. ", rs.Namespace, rs.Name)
		return nil
	}

	var meta metav1.ObjectMeta

	// In case PerconaXtraDB restore using Stash, we can't refer the Appbinding object
	// in `.spec.target.ref` of RestoreSession object. As a result, the name of the
	// original PerconaXtraDB object is unknown here. So, we need to check which object
	// of kind PerconaXtraDB has specified the current RestoreSession object.
	//
	// But, for other database we don't need to do this. We can simply use the value of
	// `.spec.target.ref.name` from current RestoreSession object.
	switch rs.Labels[api.LabelDatabaseKind] {
	case api.ResourceKindPerconaXtraDB:
		if rs.Spec.Target.Replicas == nil {
			meta = metav1.ObjectMeta{
				Name: rs.Spec.Target.Ref.Name,
			}

			break
		}
		pxList, err := c.ExtClient.KubedbV1alpha1().PerconaXtraDBs(rs.Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, px := range pxList.Items {
			if px.Spec.Init != nil && px.Spec.Init.StashRestoreSession != nil &&
				px.Spec.Init.StashRestoreSession.Name == rs.Name {
				meta = metav1.ObjectMeta{
					Name: px.Name,
				}

				break
			}
		}

		if meta.Name == "" {
			log.Debugf("no existing %v objects in namespace %q has specified %v named %q",
				api.ResourceKindPerconaXtraDB, rs.Namespace, v1beta1.ResourceKindRestoreSession, rs.Name)
			return nil
		}
	default:
		meta = metav1.ObjectMeta{
			Name: rs.Spec.Target.Ref.Name,
		}
	}

	meta.Namespace = rs.Namespace

	var phase api.DatabasePhase
	var reason string
	if rs.Status.Phase == v1beta1.RestoreSessionSucceeded {
		phase = api.DatabasePhaseRunning
		if err := c.snapshotter.UpsertDatabaseAnnotation(meta, map[string]string{
			api.AnnotationInitialized: "",
		}); err != nil {
			return err
		}
	} else {
		phase = api.DatabasePhaseFailed
		reason = "Failed to complete initialization"
	}
	if err := c.snapshotter.SetDatabaseStatus(meta, phase, reason); err != nil {
		return err
	}

	runtimeObj, err := c.snapshotter.GetDatabase(meta)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	if rs.Status.Phase == v1beta1.RestoreSessionSucceeded {
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}
