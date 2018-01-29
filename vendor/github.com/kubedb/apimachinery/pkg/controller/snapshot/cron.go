package snapshot

import (
	"fmt"
	"sync"
	"time"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/orcaman/concurrent-map"
	"gopkg.in/robfig/cron.v2"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
)

type CronControllerInterface interface {
	StartCron()
	ScheduleBackup(runtime.Object, metav1.ObjectMeta, *api.BackupScheduleSpec) error
	StopBackupScheduling(metav1.ObjectMeta)
	StopCron()
}

type cronController struct {
	// ThirdPartyExtension client
	extClient cs.KubedbV1alpha1Interface
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
func NewCronController(client kubernetes.Interface, extClient cs.KubedbV1alpha1Interface) CronControllerInterface {
	return &cronController{
		extClient:     extClient,
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
	runtimeObj runtime.Object,
	// ObjectMeta of Database TPR object
	om metav1.ObjectMeta,
	// BackupScheduleSpec
	spec *api.BackupScheduleSpec,
) error {
	// cronEntry name
	cronEntryName := fmt.Sprintf("%v@%v", om.Name, om.Namespace)

	invoker := &snapshotInvoker{
		extClient:     c.extClient,
		runtimeObject: runtimeObj,
		om:            om,
		spec:          spec,
		eventRecorder: c.eventRecorder,
	}

	// Remove previous cron job if exist
	if id, exists := c.cronEntryIDs.Pop(cronEntryName); exists {
		c.cron.Remove(id.(cron.EntryID))
	} else {
		invoker.createScheduledSnapshot()
	}

	// Set cron job
	entryID, err := c.cron.AddFunc(spec.CronExpression, invoker.createScheduledSnapshot)
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
	extClient     cs.KubedbV1alpha1Interface
	runtimeObject runtime.Object
	om            metav1.ObjectMeta
	spec          *api.BackupScheduleSpec
	eventRecorder record.EventRecorder
}

func (s *snapshotInvoker) createScheduledSnapshot() {
	kind := s.runtimeObject.GetObjectKind().GroupVersionKind().Kind
	name := s.om.Name

	labelMap := map[string]string{
		api.LabelDatabaseKind:   kind,
		api.LabelDatabaseName:   name,
		api.LabelSnapshotStatus: string(api.SnapshotPhaseRunning),
	}

	snapshotList, err := s.extClient.Snapshots(s.om.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.Set(labelMap).AsSelector().String(),
	})
	if err != nil {
		s.eventRecorder.Eventf(
			api.ObjectReferenceFor(s.runtimeObject),
			core.EventTypeWarning,
			eventer.EventReasonFailedToList,
			"Failed to list Snapshots. Reason: %v",
			err,
		)
		log.Errorln(err)
		return
	}

	if len(snapshotList.Items) > 0 {
		s.eventRecorder.Event(
			api.ObjectReferenceFor(s.runtimeObject),
			core.EventTypeNormal,
			eventer.EventReasonIgnoredSnapshot,
			"Skipping scheduled Backup. One is still active.",
		)
		log.Debugln("Skipping scheduled Backup. One is still active.")
		return
	}

	// Set label. Elastic controller will detect this using label selector
	labelMap = map[string]string{
		api.LabelDatabaseKind: kind,
		api.LabelDatabaseName: name,
	}

	now := time.Now().UTC()
	snapshotName := fmt.Sprintf("%v-%v", s.om.Name, now.Format("20060102-150405"))

	if _, err = s.createSnapshot(snapshotName); err != nil {
		log.Errorln(err)
	}
}

func (s *snapshotInvoker) createSnapshot(snapshotName string) (*api.Snapshot, error) {
	labelMap := map[string]string{
		api.LabelDatabaseKind: s.runtimeObject.GetObjectKind().GroupVersionKind().Kind,
		api.LabelDatabaseName: s.om.Name,
	}

	snapshot := &api.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotName,
			Namespace: s.om.Namespace,
			Labels:    labelMap,
		},
		Spec: api.SnapshotSpec{
			DatabaseName:        s.om.Name,
			SnapshotStorageSpec: s.spec.SnapshotStorageSpec,
			Resources:           s.spec.Resources,
		},
	}

	snapshot, err := s.extClient.Snapshots(snapshot.Namespace).Create(snapshot)
	if err != nil {
		s.eventRecorder.Eventf(
			s.runtimeObject,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create Snapshot. Reason: %v",
			err,
		)
		return nil, err
	}

	return snapshot, nil
}
