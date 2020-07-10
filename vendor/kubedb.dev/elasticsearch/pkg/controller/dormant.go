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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
)

// WaitUntilPaused is an Interface of *amc.Controller
func (c *Controller) waitUntilPaused(db *api.Elasticsearch) error {
	log.Infof("waiting for pods for Elasticsearch %v/%v to be deleted\n", db.Namespace, db.Name)
	if err := core_util.WaitUntilPodDeletedBySelector(context.TODO(), c.Client, db.Namespace, metav1.SetAsLabelSelector(db.OffshootSelectors())); err != nil {
		return err
	}

	log.Infof("waiting for services for Elasticsearch %v/%v to be deleted\n", db.Namespace, db.Name)
	if err := core_util.WaitUntilServiceDeletedBySelector(context.TODO(), c.Client, db.Namespace, metav1.SetAsLabelSelector(db.OffshootSelectors())); err != nil {
		return err
	}

	if err := c.waitUntilRBACStuffDeleted(db); err != nil {
		return err
	}

	if err := c.waitUntilStatefulSetsDeleted(db); err != nil {
		return err
	}

	if err := c.waitUntilDeploymentsDeleted(db); err != nil {
		return err
	}

	return nil
}

func (c *Controller) waitUntilRBACStuffDeleted(elasticsearch *api.Elasticsearch) error {
	// Delete ServiceAccount
	if err := core_util.WaitUntillServiceAccountDeleted(context.TODO(), c.Client, elasticsearch.ObjectMeta); err != nil {
		return err
	}
	return nil
}

func (c *Controller) waitUntilStatefulSetsDeleted(db *api.Elasticsearch) error {
	log.Infof("waiting for statefulsets for Elasticsearch %v/%v to be deleted\n", db.Namespace, db.Name)
	return wait.PollImmediate(kutil.RetryInterval, kutil.GCTimeout, func() (bool, error) {
		if sts, err := c.Client.AppsV1().StatefulSets(db.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(db.OffshootSelectors()).String()}); err != nil && kerr.IsNotFound(err) || len(sts.Items) == 0 {
			return true, nil
		}
		return false, nil
	})
}

func (c *Controller) waitUntilDeploymentsDeleted(db *api.Elasticsearch) error {
	log.Infof("waiting for deployments for Elasticsearch %v/%v to be deleted\n", db.Namespace, db.Name)
	return wait.PollImmediate(kutil.RetryInterval, kutil.GCTimeout, func() (bool, error) {
		if deploys, err := c.Client.AppsV1().Deployments(db.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(db.OffshootSelectors()).String()}); err != nil && kerr.IsNotFound(err) || len(deploys.Items) == 0 {
			return true, nil
		}
		return false, nil
	})
}

// wipeOutDatabase is a generic function to call from WipeOutDatabase and elasticsearch pause method.
func (c *Controller) wipeOutDatabase(meta metav1.ObjectMeta, secrets []string, owner *metav1.OwnerReference) error {
	secretUsed, err := c.secretsUsedByPeers(meta)
	if err != nil {
		return errors.Wrap(err, "error in getting used secret list")
	}
	unusedSecrets := sets.NewString(secrets...).Difference(secretUsed)

	for _, unusedSecret := range unusedSecrets.List() {
		secret, err := c.Client.CoreV1().Secrets(meta.Namespace).Get(context.TODO(), unusedSecret, metav1.GetOptions{})
		if kerr.IsNotFound(err) {
			unusedSecrets.Delete(secret.Name)
			continue
		}

		if err != nil {
			return errors.Wrap(err, "error in getting db secret")
		}
		genericKey, ok := secret.Labels[meta_util.ManagedByLabelKey]
		if !ok || genericKey != api.GenericKey {
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
// Similarly, isSecretUsed also checks for DomantDB of similar dbKind label.
func (c *Controller) secretsUsedByPeers(meta metav1.ObjectMeta) (sets.String, error) {
	secretUsed := sets.NewString()

	dbList, err := c.esLister.Elasticsearches(meta.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, es := range dbList {
		if es.Name != meta.Name {
			secretUsed.Insert(es.Spec.GetSecrets()...)
		}
	}
	return secretUsed, nil
}

// haltDatabase keeps PVC and secrets and deletes rest of the resources generated by kubedb
func (c *Controller) haltDatabase(db *api.Elasticsearch) error {
	labelSelector := labels.SelectorFromSet(db.OffshootSelectors()).String()

	// delete appbinding
	log.Infof("deleting AppBindings of Elasticsearch %v/%v.", db.Namespace, db.Name)
	if err := c.AppCatalogClient.
		AppcatalogV1alpha1().
		AppBindings(db.Namespace).
		DeleteCollection(
			context.TODO(),
			meta_util.DeleteInForeground(),
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}

	// delete PDB
	log.Infof("deleting PodDisruptionBudget of Elasticsearch %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		PolicyV1beta1().
		PodDisruptionBudgets(db.Namespace).
		DeleteCollection(
			context.TODO(),
			meta_util.DeleteInForeground(),
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}

	// delete sts collection offshoot labels
	log.Infof("deleting StatefulSets of Elasticsearch %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		AppsV1().
		StatefulSets(db.Namespace).
		DeleteCollection(
			context.TODO(),
			meta_util.DeleteInForeground(),
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}
	// delete rbacs: rolebinding, roles, serviceaccounts
	log.Infof("deleting RoleBindings of Elasticsearch %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		RbacV1().
		RoleBindings(db.Namespace).
		DeleteCollection(
			context.TODO(),
			meta_util.DeleteInForeground(),
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}
	log.Infof("deleting Roles of Elasticsearch %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		RbacV1().
		Roles(db.Namespace).
		DeleteCollection(
			context.TODO(),
			meta_util.DeleteInForeground(),
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}
	log.Infof("deleting ServiceAccounts of Elasticsearch %v/%v.", db.Namespace, db.Name)
	if err := c.Client.
		CoreV1().
		ServiceAccounts(db.Namespace).
		DeleteCollection(
			context.TODO(),
			meta_util.DeleteInForeground(),
			metav1.ListOptions{LabelSelector: labelSelector},
		); err != nil {
		return err
	}
	// delete services

	// service, stats service, gvr service
	log.Infof("deleting Services of Elasticsearch %v/%v.", db.Namespace, db.Name)
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
			Delete(context.TODO(), svc.Name, meta_util.DeleteInForeground()); err != nil {
			return err
		}
	}

	// Delete monitoring resources
	log.Infof("deleting Monitoring resources of Elasticsearch %v/%v.", db.Namespace, db.Name)
	if db.Spec.Monitor != nil {
		if err := c.deleteMonitor(db); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}
