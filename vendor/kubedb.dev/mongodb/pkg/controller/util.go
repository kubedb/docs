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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	core_util "kmodules.xyz/client-go/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	mongodb, err := c.mgLister.MongoDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return nil, err
	}

	return mongodb, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	mongodb, err := c.mgLister.MongoDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}
	_, err = util.UpdateMongoDBStatus(c.ExtClient.KubedbV1alpha1(), mongodb.ObjectMeta, func(in *api.MongoDBStatus) *api.MongoDBStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	})
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	mongodb, err := c.mgLister.MongoDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}

	_, _, err = util.PatchMongoDB(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDB) *api.MongoDB {
		in.Annotations = core_util.UpsertMap(in.Annotations, annotation)
		return in
	})
	return err
}

// wipeOutDatabase is a generic function to call from WipeOutDatabase and mongodb terminate method.
func (c *Controller) wipeOutDatabase(meta metav1.ObjectMeta, secrets []string, owner *metav1.OwnerReference) error {
	secretUsed, err := c.secretsUsedByPeers(meta)
	if err != nil {
		return errors.Wrap(err, "error in getting used secret list")
	}
	unusedSecrets := sets.NewString(secrets...).Difference(secretUsed)

	//Dont delete unused secrets that are not owned by kubeDB
	for _, unusedSecret := range unusedSecrets.List() {
		secret, err := c.Client.CoreV1().Secrets(meta.Namespace).Get(unusedSecret, metav1.GetOptions{})
		//Maybe user has delete this secret
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
			secretUsed.Insert(db.Spec.GetSecrets()...)
		}
	}
	return secretUsed, nil
}
