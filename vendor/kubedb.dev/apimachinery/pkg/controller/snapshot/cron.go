package snapshot

import (
	"errors"
	"fmt"
	"sync"
	"time"

	apiCatalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	cmap "github.com/orcaman/concurrent-map"
	cron "github.com/robfig/cron/v3"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	discovery_util "kmodules.xyz/client-go/discovery"
	meta_util "kmodules.xyz/client-go/meta"
)

type CronControllerInterface interface {
	StartCron()
	// ScheduleBackup takes parameter DB-runtime object, DB.scheduleSpec.BackupSchedule and DB-Version-Catalog
	ScheduleBackup(db runtime.Object, scheduleSpec *api.BackupScheduleSpec, catalog runtime.Object) error
	StopBackupScheduling(metav1.ObjectMeta)
	StopCron()
}

type cronController struct {
	// kube client
	kubeClient kubernetes.Interface
	// ThirdPartyExtension client
	extClient cs.Interface
	// dynamic client
	dynamicClient dynamic.Interface
	// For Internal Cron Job
	cron *cron.Cron
	// Store Cron Job EntryID for further use
	cronEntryIDs cmap.ConcurrentMap
	// Event Recorder
	eventRecorder record.EventRecorder
	// To perform start operation once
	once sync.Once
}

/*
 NewCronController returns CronControllerInterface.
 Need to call StartCron() method to start Cron.
*/
func NewCronController(client kubernetes.Interface, extClient cs.Interface, dc dynamic.Interface) CronControllerInterface {
	return &cronController{
		kubeClient:    client,
		extClient:     extClient,
		dynamicClient: dc,
		cron:          cron.New(),
		cronEntryIDs:  cmap.New(),
		eventRecorder: eventer.NewEventRecorder(client, "Cron controller"),
	}
}

func (c *cronController) StartCron() {
	c.once.Do(func() {
		c.cron.Start()
	})
}

func (c *cronController) ScheduleBackup(
	// Runtime Object to push event
	db runtime.Object,
	// BackupScheduleSpec
	scheduleSpec *api.BackupScheduleSpec,
	// DBVersion catalog
	catalog runtime.Object,
) error {
	dbObjectMeta, err := meta.Accessor(db)
	if err != nil {
		return err
	}
	// cronEntry name
	cronEntryName := fmt.Sprintf("%v@%v", dbObjectMeta.GetName(), dbObjectMeta.GetNamespace())

	invoker := &snapshotInvoker{
		kubeClient:    c.kubeClient,
		extClient:     c.extClient,
		dynamicClient: c.dynamicClient,
		db:            db,
		dbMetaObject:  dbObjectMeta,
		scheduleSpec:  scheduleSpec,
		catalog:       catalog,
		eventRecorder: c.eventRecorder,
	}

	// Remove previous cron job if exist
	if id, exists := c.cronEntryIDs.Pop(cronEntryName); exists {
		c.cron.Remove(id.(cron.EntryID))
	}

	// Set cron job
	entryID, err := c.cron.AddFunc(scheduleSpec.CronExpression, func() {
		if err := invoker.createScheduledSnapshot(); err != nil {
			invoker.eventRecorder.Eventf(
				invoker.db,
				core.EventTypeWarning,
				eventer.EventReasonFailedToList,
				err.Error(),
			)
			log.Errorf(err.Error())
		}
	})
	if err != nil {
		return err
	}

	// Add job entryID
	c.cronEntryIDs.Set(cronEntryName, entryID)

	return nil
}

func (c *cronController) StopBackupScheduling(om metav1.ObjectMeta) {
	// cronEntry name
	cronEntryName := fmt.Sprintf("%v@%v", om.Name, om.Namespace)

	if id, exists := c.cronEntryIDs.Pop(cronEntryName); exists {
		c.cron.Remove(id.(cron.EntryID))
	}
}

func (c *cronController) StopCron() {
	c.cron.Stop()
}

type snapshotInvoker struct {
	kubeClient    kubernetes.Interface
	extClient     cs.Interface
	dynamicClient dynamic.Interface
	db            runtime.Object
	dbMetaObject  metav1.Object
	scheduleSpec  *api.BackupScheduleSpec
	catalog       runtime.Object
	eventRecorder record.EventRecorder
}

func (s *snapshotInvoker) createScheduledSnapshot() error {
	dbKind := meta_util.GetKind(s.db)
	catalogKind := meta_util.GetKind(s.catalog)
	catalogMetaObject, err := meta.Accessor(s.catalog)
	if err != nil {
		return err
	}

	gvkCatalog := apiCatalog.SchemeGroupVersion.WithKind(catalogKind)
	gvrCatalog, err := discovery_util.ResourceForGVK(s.kubeClient.Discovery(), gvkCatalog)
	if err != nil {
		return fmt.Errorf("failed to get 'gvrCatalog' for %v/%v. Reason: %v", catalogKind, catalogMetaObject, err)
	}

	updatedCatalog, err := s.dynamicClient.Resource(gvrCatalog).Get(catalogMetaObject.GetName(), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get DB Catalog %v/%v. Reason: %v", catalogKind, catalogMetaObject, err)
	}

	if deprecated, found, err := unstructured.NestedBool(updatedCatalog.UnstructuredContent(), "scheduleSpec", "deprecated"); err != nil {
		return fmt.Errorf("failed to get scheduleSpec.Deprecated value. Reason: %v", err)
	} else if found && deprecated {
		return fmt.Errorf("%v %s/%s is using deprecated version %v. Skipped processing scheduler",
			dbKind, s.dbMetaObject.GetNamespace(), s.dbMetaObject.GetName(), catalogMetaObject)
	}

	labelMap := map[string]string{
		api.LabelDatabaseKind:   dbKind,
		api.LabelDatabaseName:   s.dbMetaObject.GetName(),
		api.LabelSnapshotStatus: string(api.SnapshotPhaseRunning),
	}

	snapshotList, err := s.extClient.KubedbV1alpha1().Snapshots(s.dbMetaObject.GetNamespace()).List(metav1.ListOptions{
		LabelSelector: labels.Set(labelMap).AsSelector().String(),
	})
	if err != nil {
		return fmt.Errorf("failed to list Snapshots. Reason: %v", err)
	}

	if len(snapshotList.Items) > 0 {
		return errors.New("skipping scheduled Backup. One is still active")
	}

	now := time.Now().UTC()
	snapshotName := fmt.Sprintf("%v-%v", s.dbMetaObject.GetName(), now.Format("20060102-150405"))

	if _, err = s.createSnapshot(snapshotName); err != nil {
		return err
	}
	return nil
}

func (s *snapshotInvoker) createSnapshot(snapshotName string) (*api.Snapshot, error) {
	labelMap := map[string]string{
		api.LabelDatabaseKind: meta_util.GetKind(s.db),
		api.LabelDatabaseName: s.dbMetaObject.GetName(),
	}

	snapshot := &api.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotName,
			Namespace: s.dbMetaObject.GetNamespace(),
			Labels:    labelMap,
		},
		Spec: api.SnapshotSpec{
			DatabaseName:       s.dbMetaObject.GetName(),
			Backend:            s.scheduleSpec.Backend,
			StorageType:        s.scheduleSpec.StorageType,
			PodTemplate:        s.scheduleSpec.PodTemplate,
			PodVolumeClaimSpec: s.scheduleSpec.PodVolumeClaimSpec,
		},
	}

	snapshot, err := s.extClient.KubedbV1alpha1().Snapshots(snapshot.Namespace).Create(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to create Snapshot. Reason: %v", err)
	}

	return snapshot, nil
}
