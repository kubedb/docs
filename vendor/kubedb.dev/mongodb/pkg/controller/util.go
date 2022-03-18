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

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
)

// wipeOutDatabase is a generic function to call from WipeOutDatabase and mongodb terminate method.
func (c *Controller) wipeOutDatabase(meta metav1.ObjectMeta, secrets []string, owner *metav1.OwnerReference) error {
	secretUsed, err := c.secretsUsedByPeers(meta)
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
func (c *Controller) secretsUsedByPeers(meta metav1.ObjectMeta) (sets.String, error) {
	secretUsed := sets.NewString()
	dbList, err := c.mgLister.MongoDBs(meta.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, db := range dbList {
		if db.Name != meta.Name {
			secretUsed.Insert(db.Spec.GetPersistentSecrets()...)
		}
	}
	return secretUsed, nil
}
