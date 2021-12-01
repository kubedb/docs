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

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Controller) waitUntilSentinelHalted(db *api.RedisSentinel) error {
	klog.Infof("waiting for pods for Redis Sentinel %v/%v to be deleted\n", db.Namespace, db.Name)
	if err := core_util.WaitUntilPodDeletedBySelector(context.TODO(), c.Client, db.Namespace, metav1.SetAsLabelSelector(db.OffshootSelectors())); err != nil {
		return err
	}

	klog.Infof("waiting for services for Redis Sentinel %v/%v to be deleted\n", db.Namespace, db.Name)
	if err := core_util.WaitUntilServiceDeletedBySelector(context.TODO(), c.Client, db.Namespace, metav1.SetAsLabelSelector(db.OffshootSelectors())); err != nil {
		return err
	}

	if err := c.waitUntilRBACStuffDeleted(db.ObjectMeta); err != nil {
		return err
	}

	if err := c.waitUntilSentinelStatefulSetsDeleted(db); err != nil {
		return err
	}

	return nil
}

func (c *Controller) waitUntilSentinelStatefulSetsDeleted(db *api.RedisSentinel) error {
	klog.Infof("waiting for statefulsets for Redis Sentinel %v/%v to be deleted\n", db.Namespace, db.Name)
	return wait.PollImmediate(kutil.RetryInterval, kutil.GCTimeout, func() (bool, error) {
		if sts, err := c.Client.AppsV1().StatefulSets(db.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(db.OffshootSelectors()).String()}); err != nil && kerr.IsNotFound(err) || len(sts.Items) == 0 {
			return true, nil
		}
		return false, nil
	})
}

// wipeOutSentinel is a generic function to call from WipeOutDatabase and Redis Sentinel  pause method.
func (c *Controller) wipeOutSentinel(meta metav1.ObjectMeta, secrets []string, owner *metav1.OwnerReference) error {
	secretUsed, err := c.sentinelSecretsUsedByPeers(meta)
	if err != nil {
		return errors.Wrap(err, "error in getting used secret list")
	}
	unusedSecrets := sets.NewString(secrets...).Difference(secretUsed)

	//Dont delete unused secrets that are not owned by kubeDB
	for _, unusedSecret := range unusedSecrets.List() {
		secret, err := c.Client.CoreV1().Secrets(meta.Namespace).Get(context.TODO(), unusedSecret, metav1.GetOptions{})
		//Maybe user has delete this secret
		if kerr.IsNotFound(err) {
			unusedSecrets.Delete(secret.Name)
			continue
		}
		if err != nil {
			return errors.Wrap(err, "error in getting db secret")
		}
		genericKey, ok := secret.Labels[meta_util.ManagedByLabelKey]
		if !ok || genericKey != kubedb.GroupName {
			unusedSecrets.Delete(secret.Name)
		}
	}

	return dynamic_util.EnsureOwnerReferenceForItems(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		meta.Namespace,
		unusedSecrets.List(),
		owner)
}

// isSecretUsed gets the DBList of same kind, then checks if our required secret is used by those.
// Similarly, isSecretUsed also checks for DotmantDB of similar dbKind label.
func (c *Controller) sentinelSecretsUsedByPeers(meta metav1.ObjectMeta) (sets.String, error) {
	secretUsed := sets.NewString()

	rsList, err := c.rsLister.RedisSentinels(meta.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, rs := range rsList {
		if rs.Name != meta.Name {
			secretUsed.Insert(rs.GetPersistentSecrets()...)
			secretUsed.Insert(c.GetRedisSentinelSecrets(rs)...)
		}
	}
	return secretUsed, nil
}

// haltSentinelDatabase keeps PVC and secrets and deletes rest of the resources generated by kubedb
func (c *Controller) haltSentinelDatabase(db *api.RedisSentinel) error {
	labelSelector := labels.SelectorFromSet(db.OffshootSelectors()).String()
	policy := metav1.DeletePropagationBackground

	// delete appbinding
	klog.Infof("deleting AppBindings of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if err := c.AppCatalogClient.
		AppcatalogV1alpha1().
		AppBindings(db.Namespace).
		DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{PropagationPolicy: &policy},
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}

	// delete PDB
	klog.Infof("deleting PodDisruptionBudget of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		PolicyV1beta1().
		PodDisruptionBudgets(db.Namespace).
		DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{PropagationPolicy: &policy},
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}

	// delete sts collection offshoot labels
	klog.Infof("deleting StatefulSets of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		AppsV1().
		StatefulSets(db.Namespace).
		DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{PropagationPolicy: &policy},
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}

	// delete deployment collection offshoot labels
	klog.Infof("deleting Deployments of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		AppsV1().
		Deployments(db.Namespace).
		DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{PropagationPolicy: &policy},
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}

	// delete rbacs: rolebinding, roles, serviceaccounts
	klog.Infof("deleting RoleBindings of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		RbacV1().
		RoleBindings(db.Namespace).
		DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{PropagationPolicy: &policy},
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}
	klog.Infof("deleting Roles of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		RbacV1().
		Roles(db.Namespace).
		DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{PropagationPolicy: &policy},
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}
	klog.Infof("deleting ServiceAccounts of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		CoreV1().
		ServiceAccounts(db.Namespace).
		DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{PropagationPolicy: &policy},
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}
	// delete services

	// service, stats service, gvr service
	klog.Infof("deleting Services of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	svcs, err := c.Client.
		CoreV1().
		Services(db.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	for _, svc := range svcs.Items {
		if err := c.Client.
			CoreV1().
			Services(db.Namespace).
			Delete(context.TODO(), svc.Name, metav1.DeleteOptions{PropagationPolicy: &policy}); err != nil {
			return err
		}
	}

	// Delete monitoring resources
	klog.Infof("deleting Monitoring resources of Redis Sentinel %v/%v.", db.Namespace, db.Name)
	if db.Spec.Monitor != nil {
		if err := c.deleteSentinelMonitor(db); err != nil {
			klog.Errorln(err)
			return nil
		}
	}
	return nil
}