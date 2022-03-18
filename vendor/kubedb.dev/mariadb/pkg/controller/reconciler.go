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
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amc "kubedb.dev/apimachinery/pkg/controller"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

type Reconciler struct {
	amc.Config
	*amc.Controller
}

func NewReconciler(config amc.Config, controller *amc.Controller) *Reconciler {
	return &Reconciler{
		Controller: controller,
		Config:     config,
	}
}

func (c *Reconciler) ReconcileNodes(db *api.MariaDB) (*api.MariaDB, kutil.VerbType, error) {
	if db == nil {
		return nil, kutil.VerbUnchanged, errors.New("MariaDB object is empty")
	}

	if err := c.ensureAuthSecret(db); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to create AuthSecret")
	}

	// wait for certificates
	if db.Spec.TLS != nil && db.Spec.TLS.IssuerRef != nil {
		ok, err := dynamic_util.ResourcesExists(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			db.Namespace,
			requiredSecretList(db)...,
		)
		if err != nil {
			return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to get TLS secrets")
		}
		if !ok {
			// If the certificates are managed by the enterprise operator,
			// It takes some time for the secrets to get ready.
			// If any required secret is yet to get ready,
			// drop the mariadb object from work queue (i.e. return nil with no error).
			// When any secret owned by this MariaDB object is created/updated,
			// this MariaDB object will be enqueued again for processing.
			klog.Infoln(fmt.Sprintf("Required secrets for MariaDB: %s/%s are not ready yet", db.Namespace, db.Name))
			return nil, kutil.VerbUnchanged, nil
		}
	}

	vt, err := c.ensureStatefulSet(db)
	if err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to create stateful set")
	}

	return db, vt, nil
}
