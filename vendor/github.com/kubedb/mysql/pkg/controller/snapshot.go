package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	amv "github.com/kubedb/apimachinery/pkg/validator"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	mysql, err := c.myLister.MySQLs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return nil, err
	}

	return mysql, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	mysql, err := c.myLister.MySQLs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}
	_, err = util.UpdateMySQLStatus(c.ExtClient, mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	}, api.EnableStatusSubresource)
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	mysql, err := c.myLister.MySQLs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}

	_, _, err = util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
		in.Annotations = core_util.UpsertMap(mysql.Annotations, annotation)
		return in
	})
	return err
}

func (c *Controller) ValidateSnapshot(snapshot *api.Snapshot) error {
	// Database name can't empty
	databaseName := snapshot.Spec.DatabaseName
	if databaseName == "" {
		return fmt.Errorf(`object 'DatabaseName' is missing in '%v'`, snapshot.Spec)
	}

	if _, err := c.myLister.MySQLs(snapshot.Namespace).Get(databaseName); err != nil {
		return err
	}

	return amv.ValidateSnapshotSpec(c.Client, snapshot.Spec.SnapshotStorageSpec, snapshot.Namespace)
}

func (c *Controller) GetSnapshotter(snapshot *api.Snapshot) (*batch.Job, error) {
	return c.getSnapshotterJob(snapshot)
}

func (c *Controller) WipeOutSnapshot(snapshot *api.Snapshot) error {
	if snapshot.Spec.Local != nil {
		local := snapshot.Spec.Local
		if local.VolumeSource.EmptyDir != nil {
			return nil
		}
	}
	return c.DeleteSnapshotData(snapshot)
}

func (c *Controller) getVolumeForSnapshot(pvcSpec core.PersistentVolumeClaimSpec, jobName, namespace string) (*core.Volume, error) {
	volume := &core.Volume{
		Name: "util-volume",
	}
	if len(pvcSpec.AccessModes) == 0 {
		pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
			core.ReadWriteOnce,
		}
		log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, pvcSpec)
	}

	claim := &core.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: pvcSpec,
	}
	if pvcSpec.StorageClassName != nil {
		claim.Annotations = map[string]string{
			"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
		}
	}

	if _, err := c.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Create(claim); err != nil {
		return nil, err
	}

	volume.PersistentVolumeClaim = &core.PersistentVolumeClaimVolumeSource{
		ClaimName: claim.Name,
	}

	return volume, nil
}
