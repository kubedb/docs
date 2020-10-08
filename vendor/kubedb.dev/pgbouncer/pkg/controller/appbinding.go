/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) manageAppBindingEvent(key string) error {
	//wait for pgboncer to ber ready
	log.Debugln("started processing appBindings, key:", key)
	_, _, err := c.appBindingInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching appBinding with key %s from store failed with %v", key, err)
		return err
	}
	splitKey := strings.Split(key, "/")

	if len(splitKey) != 2 || splitKey[0] == "" || splitKey[1] == "" {
		return nil
	}
	//Now we are interested in this particular appBinding
	appBindingInfo := make(map[string]string)
	appBindingInfo[namespaceKey] = splitKey[0]
	appBindingInfo[nameKey] = splitKey[1]
	if appBindingInfo[namespaceKey] == systemNamespace || appBindingInfo[namespaceKey] == publicNamespace {
		return nil
	}

	pgBouncerList, err := c.DBClient.KubedbV1alpha2().PgBouncers(core.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pgBouncer := range pgBouncerList.Items {
		if pgBouncer.GetNamespace() == appBindingInfo[namespaceKey] {
			err := c.checkAppBindingsInPgBouncerSpec(appBindingInfo, &pgBouncer)
			if err != nil {
				log.Warning(err)
			}
		}
	}
	return nil
}

func (c *Controller) checkAppBindingsInPgBouncerSpec(appBindingInfo map[string]string, pgbouncer *api.PgBouncer) error {
	if pgbouncer.Spec.Databases != nil && len(pgbouncer.Spec.Databases) > 0 {
		for _, db := range pgbouncer.Spec.Databases {
			if db.DatabaseRef.Name == appBindingInfo[nameKey] {
				err := c.manageService(pgbouncer)
				if err != nil {
					return err
				}
				err = c.manageConfigMap(pgbouncer)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Controller) getCABundlesFromAppBindingsInPgBouncerSpec(pgbouncer *api.PgBouncer) (string, error) {
	isCAForAppBindingInserted := map[string]bool{}
	var myCAStrings string
	if pgbouncer.Spec.Databases != nil && len(pgbouncer.Spec.Databases) > 0 {
		for _, db := range pgbouncer.Spec.Databases {
			appBinding, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(db.DatabaseRef.Namespace).Get(context.TODO(), db.DatabaseRef.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					log.Infoln(err)
					continue //because non blocking err
				}
				return "", err
			}
			if !isCAForAppBindingInserted[appBinding.Namespace+"-"+appBinding.Name] && len(appBinding.Spec.ClientConfig.CABundle) > 0 {
				isCAForAppBindingInserted[appBinding.Namespace+"-"+appBinding.Name] = true
				myCAStrings = myCAStrings + fmt.Sprintln(string(appBinding.Spec.ClientConfig.CABundle))
			}
		}
	}
	if len(myCAStrings) > 0 {
		return myCAStrings, nil
	}

	return "", nil
}
