package snapshot

import (
	"fmt"
	"sync"

	hookapi "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	"github.com/appscode/kutil/meta"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	plugin "github.com/kubedb/apimachinery/pkg/admission"
	amv "github.com/kubedb/apimachinery/pkg/validator"
	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
			Group:    "admission.kubedb.com",
			Version:  "v1alpha1",
			Resource: "snapshotreviews",
		},
		"snapshotreview"
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
			return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
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
	if err := amv.ValidateSnapshotSpec(a.client, obj.(*api.Snapshot).Spec.SnapshotStorageSpec, req.Namespace); err != nil {
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

	kind, err := meta_util.GetStringValue(snapshot.Labels, api.LabelDatabaseKind)
	if err != nil {
		return fmt.Errorf("'%v:XDB' label is missing", api.LabelDatabaseKind)
	}

	// Check if DB exists
	switch kind {
	case api.ResourceKindElasticsearch:
		if _, err := a.extClient.KubedbV1alpha1().Elasticsearches(snapshot.Namespace).Get(databaseName, metav1.GetOptions{}); err != nil {
			return err
		}
	case api.ResourceKindPostgres:
		if _, err := a.extClient.KubedbV1alpha1().Postgreses(snapshot.Namespace).Get(databaseName, metav1.GetOptions{}); err != nil {
			return err
		}
	case api.ResourceKindMongoDB:
		if _, err := a.extClient.KubedbV1alpha1().MongoDBs(snapshot.Namespace).Get(databaseName, metav1.GetOptions{}); err != nil {
			return err
		}
	case api.ResourceKindMySQL:
		if _, err := a.extClient.KubedbV1alpha1().MySQLs(snapshot.Namespace).Get(databaseName, metav1.GetOptions{}); err != nil {
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
