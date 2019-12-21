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

package snapshot

import (
	"fmt"
	"sync"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	plugin "kubedb.dev/apimachinery/pkg/admission"
	amv "kubedb.dev/apimachinery/pkg/validator"

	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kmodules.xyz/client-go/meta"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type SnapshotValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &SnapshotValidator{}

func (a *SnapshotValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "validators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "snapshotvalidators",
		},
		"snapshotvalidator"
}

func (a *SnapshotValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	if a.client, err = kubernetes.NewForConfig(config); err != nil {
		return err
	}
	if a.extClient, err = cs.NewForConfig(config); err != nil {
		return err
	}
	return err
}

func (a *SnapshotValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindSnapshot {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return hookapi.StatusUninitialized()
	}

	obj, err := meta.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
	if err != nil {
		return hookapi.StatusBadRequest(err)
	}
	if req.Operation == admission.Update {
		oldObject, err := meta.UnmarshalFromJSON(req.OldObject.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}
		if err := plugin.ValidateUpdate(obj, oldObject, req.Kind.Kind); err != nil {
			return hookapi.StatusBadRequest(err)
		}
		// Skip checking validation if Spec is not changed
		if meta_util.Equal(obj.(*api.Snapshot).Spec, oldObject.(*api.Snapshot).Spec) {
			status.Allowed = true
			return status
		}
	}
	// validates if database of particular kind exists
	if err := a.validateSnapshot(obj.(*api.Snapshot)); err != nil {
		return hookapi.StatusForbidden(err)
	}
	// validates Snapshot Spec
	if err := amv.ValidateSnapshotSpec(obj.(*api.Snapshot).Spec.Backend); err != nil {
		return hookapi.StatusForbidden(err)
	}
	if req.Operation == admission.Create {
		// isSnapshotRunning checks if a snapshot is already running. Check this only when creating snapshot,
		// because Snapshot.Status will be needed to edit later and this method will give error for that update.
		if err := a.isSnapshotRunning(obj.(*api.Snapshot)); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}

	status.Allowed = true
	return status
}

// validateSnapshot checks if the database of the particular kind actually exists.
func (a *SnapshotValidator) validateSnapshot(snapshot *api.Snapshot) error {
	// Database name can't empty
	databaseName := snapshot.Spec.DatabaseName
	if databaseName == "" {
		return fmt.Errorf(`object 'DatabaseName' is missing in '%v'`, snapshot.Spec)
	}

	//

	kind, err := meta_util.GetStringValue(snapshot.Labels, api.LabelDatabaseKind)
	if err != nil {
		return fmt.Errorf("'%v:XDB' label is missing", api.LabelDatabaseKind)
	}

	// Check if DB exists
	switch kind {
	case api.ResourceKindElasticsearch:
		es, err := a.extClient.KubedbV1alpha1().Elasticsearches(snapshot.Namespace).Get(databaseName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		storage := es.Spec.Storage
		if es.Spec.Topology != nil {
			storage = es.Spec.Topology.Data.Storage
		}
		if err := verifyStorageType(snapshot, storage); err != nil {
			return err
		}
	case api.ResourceKindPostgres:
		pg, err := a.extClient.KubedbV1alpha1().Postgreses(snapshot.Namespace).Get(databaseName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if err := verifyStorageType(snapshot, pg.Spec.Storage); err != nil {
			return err
		}
	case api.ResourceKindMongoDB:
		mg, err := a.extClient.KubedbV1alpha1().MongoDBs(snapshot.Namespace).Get(databaseName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if err := verifyStorageType(snapshot, mg.Spec.Storage); err != nil {
			return err
		}

	case api.ResourceKindMySQL:
		my, err := a.extClient.KubedbV1alpha1().MySQLs(snapshot.Namespace).Get(databaseName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if err := verifyStorageType(snapshot, my.Spec.Storage); err != nil {
			return err
		}
	case api.ResourceKindRedis:
		if _, err := a.extClient.KubedbV1alpha1().Redises(snapshot.Namespace).Get(databaseName, metav1.GetOptions{}); err != nil {
			return err
		}
	case api.ResourceKindMemcached:
		if _, err := a.extClient.KubedbV1alpha1().Memcacheds(snapshot.Namespace).Get(databaseName, metav1.GetOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func verifyStorageType(snapshot *api.Snapshot, dbPvcSpec *core.PersistentVolumeClaimSpec) error {
	if snapshot.Spec.StorageType != nil &&
		*snapshot.Spec.StorageType == api.StorageTypeDurable &&
		snapshot.Spec.PodVolumeClaimSpec == nil &&
		dbPvcSpec == nil {
		return fmt.Errorf("snapshot storagetype is durable but, " +
			"pvc Spec is not specified in either PodVolumeClaimSpec or db.Spec.Storage")
	}
	return nil
}

func (a *SnapshotValidator) isSnapshotRunning(snapshot *api.Snapshot) error {
	labelMap := map[string]string{
		api.LabelDatabaseKind:   snapshot.Labels[api.LabelDatabaseKind],
		api.LabelDatabaseName:   snapshot.Spec.DatabaseName,
		api.LabelSnapshotStatus: string(api.SnapshotPhaseRunning),
	}

	snapshotList, err := a.extClient.KubedbV1alpha1().Snapshots(snapshot.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	})
	if err != nil {
		return err
	}

	if len(snapshotList.Items) > 0 {
		return fmt.Errorf("one Snapshot is already running")
	}

	return nil
}
