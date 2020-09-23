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

package stash

import (
	"context"
	"fmt"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/discovery"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog"
	ab "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	sapis "stash.appscode.dev/apimachinery/apis"
	"stash.appscode.dev/apimachinery/apis/stash"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
)

func (c *Controller) extractRestoreInfo(invoker interface{}) (*restoreInfo, error) {
	ri := &restoreInfo{
		invoker: core.TypedLocalObjectReference{
			APIGroup: types.StringP(stash.GroupName),
		},
	}
	var err error
	switch invoker := invoker.(type) {
	case *v1beta1.RestoreSession:
		ri.invoker.Kind = invoker.Kind
		ri.invoker.Name = invoker.Name
		ri.namespace = invoker.Namespace
		ri.target = invoker.Spec.Target
		ri.phase = invoker.Status.Phase
		ri.targetDBKind = invoker.Labels[api.LabelDatabaseKind]
	case *v1beta1.RestoreBatch:
		ri.invoker.Kind = invoker.Kind
		ri.invoker.Name = invoker.Name
		ri.namespace = invoker.Namespace
		// RestoreBatch can have multiple targets. In this case, finding the appropriate target is necessary.
		ri.target, err = c.identifyTarget(invoker.Spec.Members, ri.namespace)
		if err != nil {
			return ri, err
		}
		// RestoreBatch can have multiple targets. In this case, only the database related target'c phase does matter.
		ri.phase = getTargetPhase(invoker.Status, ri.target)
		ri.targetDBKind = invoker.Labels[api.LabelDatabaseKind]
	default:
		return ri, fmt.Errorf("unknown restore invoker type")
	}
	return ri, nil
}

func (c *Controller) syncDatabasePhase(ri *restoreInfo) error {
	var err error
	if ri == nil {
		return fmt.Errorf("invalid restore information. it must not be nil")
	}
	if ri.phase != v1beta1.RestoreSucceeded && ri.phase != v1beta1.RestoreFailed && ri.phase != v1beta1.RestorePhaseUnknown {
		return fmt.Errorf("restore process hasn't completed yet. Current restore phase: %s", ri.phase)
	}

	if ri.target == nil {
		return fmt.Errorf("restore invoker does not have any target specified")
	}

	targetDBMeta := metav1.ObjectMeta{
		Namespace: ri.namespace,
	}
	targetDBMeta.Name, err = c.getDatabaseName(ri)
	if err != nil {
		return err
	}

	var phase api.DatabasePhase
	var reason string
	if ri.phase == v1beta1.RestoreSucceeded {
		phase = api.DatabasePhaseRunning
		if err := c.snapshotter.UpsertDatabaseAnnotation(targetDBMeta, map[string]string{
			api.AnnotationInitialized: "",
		}); err != nil {
			return err
		}
	} else {
		phase = api.DatabasePhaseFailed
		reason = "Failed to complete initialization"
	}
	if err := c.snapshotter.SetDatabaseStatus(targetDBMeta, phase, reason); err != nil {
		return err
	}

	runtimeObj, err := c.snapshotter.GetDatabase(targetDBMeta)
	if err != nil {
		return err
	}
	if ri.phase == v1beta1.RestoreSucceeded {
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

func (c *Controller) identifyTarget(members []v1beta1.RestoreTargetSpec, namespace string) (*v1beta1.RestoreTarget, error) {
	// check if there is any AppBinding as target. if there any, this is the desired target.
	for i, m := range members {
		if m.Target != nil {
			ok, err := targetOfGroupKind(m.Target.Ref, appcat.GroupName, ab.ResourceKindApp)
			if err != nil {
				return nil, err
			}
			if ok {
				return members[i].Target, nil
			}
		}
	}
	// no AppBinding has found as target. the target might be resulting workload (i.e. StatefulSet or Deployment(for memcached)).
	// we should check the workload's owner reference to be sure.
	for i, m := range members {
		if m.Target != nil {
			ok, err := targetOfGroupKind(m.Target.Ref, apps.GroupName, sapis.KindStatefulSet)
			if err != nil {
				return nil, err
			}
			if ok {
				sts, err := c.Client.AppsV1().StatefulSets(namespace).Get(context.Background(), m.Target.Ref.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				// if the controller owner is a KubeDB resource, then this StatefulSet must be the desired target
				ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(sts), kubedb.GroupName)
				if err != nil {
					return nil, err
				}
				if ok {
					return members[i].Target, nil
				}
			}
		}
	}
	return nil, nil
}

func getTargetPhase(status v1beta1.RestoreBatchStatus, target *v1beta1.RestoreTarget) v1beta1.RestorePhase {
	if target != nil {
		for _, m := range status.Members {
			if sapis.TargetMatched(m.Ref, target.Ref) {
				return v1beta1.RestorePhase(m.Phase)
			}
		}
	}
	return status.Phase
}

func (c *Controller) getDatabaseName(ri *restoreInfo) (string, error) {
	switch ri.targetDBKind {
	// In case of clustered PerconaXtraDB, Controller restores the volumes. Hence, we don't specify the AppBinding object
	// in `.target.ref` field of the respective restore invoker. As a result, the name of the original PerconaXtraDB object is unknown here.
	// So, we need to check which PerconaXtraDB has specified this invoker in the init section.
	case api.ResourceKindPerconaXtraDB:
		if ri.target.Replicas == nil {
			// might be stand-alone percona-xtradb. in this case, the AppBinding reference is present in `*.target.ref` section.
			return ri.target.Ref.Name, nil
		}
		pxList, err := c.ExtClient.KubedbV1alpha1().PerconaXtraDBs(ri.namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return "", err
		}

		for _, px := range pxList.Items {
			if px.Spec.Init != nil && px.Spec.Init.Initializer != nil && matchRef(*px.Spec.Init.Initializer, ri.invoker) {
				return px.Name, nil
			}
		}
		return "", fmt.Errorf("no PerconaXtraDB CR found for %s  %s/%s", ri.invoker.Kind, ri.namespace, ri.invoker.Name)
	// For Redis, Controller can restore in two models.
	// 1. For RDB restore, Controller uses sidecar model. In this case, targets are the respective StatefulSets.
	// 2. For restoring from dump, Controller uses job model. In this case, the target is the respective AppBinding.
	case api.ResourceKindRedis:
		switch ri.target.Ref.Kind {
		case ab.ResourceKindApp:
			return ri.target.Ref.Name, nil
		case sapis.KindStatefulSet:
			sts, err := c.Client.AppsV1().StatefulSets(ri.namespace).Get(context.TODO(), ri.target.Ref.Name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			owner := metav1.GetControllerOf(sts)
			ok, err := core_util.IsOwnerOfGroupKind(owner, kubedb.GroupName, api.ResourceKindRedis)
			if err != nil {
				return "", err
			}
			if !ok {
				return "", fmt.Errorf("respective Redis CR is not found for StatefulSet %s/%s", sts.Namespace, sts.Name)
			}
			return owner.Name, nil
		default:
			return "", fmt.Errorf("unknown target reference in %s %s/%s", ri.invoker.Kind, ri.namespace, ri.invoker.Name)
		}
	default:
		// For other databases, `*.target.ref` refers to the respective AppBinding object which is also the respective database
		// CR name. In this case, we can just take the `*.target.ref.name` as the database CR name.
		return ri.target.Ref.Name, nil
	}
}

// waitUntilStashInstalled waits for Controller to be installed. It check whether Controller has been installed or not by querying RestoreSession crd.
// It either waits until RestoreSession crd exists or throws error otherwise
func (c *Controller) waitUntilStashInstalled(stopCh <-chan struct{}) error {
	log.Infoln("Looking for the Stash operator.......")
	return wait.PollImmediateUntil(time.Second*10, func() (bool, error) {
		return discovery.ExistsGroupKind(c.Client.Discovery(), stash.GroupName, v1beta1.ResourceKindRestoreSession) ||
			discovery.ExistsGroupKind(c.Client.Discovery(), stash.GroupName, v1beta1.ResourceKindRestoreBatch), nil
	}, stopCh)
}

func targetOfGroupKind(target v1beta1.TargetRef, group, kind string) (bool, error) {
	gv, err := schema.ParseGroupVersion(target.APIVersion)
	if err != nil {
		return false, err
	}
	return gv.Group == group && target.Kind == kind, nil
}

func matchRef(r1, r2 core.TypedLocalObjectReference) bool {
	return r1.APIGroup == r2.APIGroup && r1.Kind == r2.Kind && r1.Name == r2.Name
}
