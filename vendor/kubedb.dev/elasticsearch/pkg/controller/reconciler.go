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
	"kubedb.dev/elasticsearch/pkg/distribution"

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

func (r *Reconciler) ReconcileNodes(db *api.Elasticsearch) (*api.Elasticsearch, kutil.VerbType, error) {
	if db == nil {
		return nil, kutil.VerbUnchanged, errors.New("Elasticsearch object is empty")
	}

	elastic, err := distribution.NewElasticsearch(r.Client, r.DBClient, db)
	if err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to get elasticsearch distribution")
	}

	// Create/sync certificate secrets
	// But if  the tls.issuerRef is set, do nothing (i.e. should be handled from enterprise operator).
	if err = elastic.EnsureCertSecrets(); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to ensure certificates secret")
	}

	// Create/sync user credential (ie. username, password) secrets
	if err = elastic.EnsureAuthSecret(); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to ensure database credential secret")
	}

	// Get the cert secret names
	// List varies depending on the elasticsearch distribution & configuration.
	sNames := elastic.RequiredCertSecretNames()
	// Check whether the secrets are available or not.
	ok, err := dynamic_util.ResourcesExists(
		r.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		db.Namespace,
		sNames...,
	)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if !ok {
		// If the certificates are managed by the enterprise operator,
		// It takes some time for the secrets to get ready.
		// If any required secret is yet to get ready,
		// drop the elasticsearch object from work queue (i.e. return nil with no error).
		// When any secret owned by this elasticsearch object is created/updated,
		// this elasticsearch object will be enqueued again for processing.
		klog.Infoln(fmt.Sprintf("Required secrets for Elasticsearch: %s/%s are not ready yet", db.Namespace, db.Name))
		return nil, kutil.VerbUnchanged, nil
	}

	if err = elastic.EnsureDefaultConfig(); err != nil {
		return nil, kutil.VerbUnchanged, errors.Wrap(err, "failed to ensure default configuration for elasticsearch")
	}

	var vt kutil.VerbType
	topology := elastic.UpdatedElasticsearch().Spec.Topology
	if topology != nil {
		vt1, err := elastic.EnsureIngestNodes()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		vt = vt1

		vt2, err := elastic.EnsureMasterNodes()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched {
			vt = vt2
		}

		vt3, err := elastic.EnsureDataNodes()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.Data != nil {
			vt = vt3
		}

		vt4, err := elastic.EnsureDataContentNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.DataContent != nil {
			vt = vt4
		}

		vt5, err := elastic.EnsureDataHotNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.DataHot != nil {
			vt = vt5
		}

		vt6, err := elastic.EnsureDataWarmNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.DataWarm != nil {
			vt = vt6
		}

		vt7, err := elastic.EnsureDataColdNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.DataCold != nil {
			vt = vt7
		}

		vt8, err := elastic.EnsureDataFrozenNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.DataFrozen != nil {
			vt = vt8
		}

		vt9, err := elastic.EnsureMLNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.ML != nil {
			vt = vt9
		}

		vt10, err := elastic.EnsureTransformNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		if vt != kutil.VerbPatched && topology.Transform != nil {
			vt = vt10
		}

	} else {
		vt, err = elastic.EnsureCombinedNode()
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
	}

	return elastic.UpdatedElasticsearch(), vt, nil
}
