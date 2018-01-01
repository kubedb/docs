package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/docker"
	amv "github.com/kubedb/apimachinery/pkg/validator"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (c *Controller) ValidateSnapshot(snapshot *api.Snapshot) error {
	// Database name can't empty
	databaseName := snapshot.Spec.DatabaseName
	if databaseName == "" {
		return fmt.Errorf(`object 'DatabaseName' is missing in '%v'`, snapshot.Spec)
	}

	mysql, err := c.ExtClient.MySQLs(snapshot.Namespace).Get(databaseName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := docker.CheckDockerImageVersion(c.opt.Docker.GetToolsImage(mysql), string(mysql.Spec.Version)); err != nil {
		return fmt.Errorf(`image %s not found`, c.opt.Docker.GetToolsImageWithTag(mysql))
	}

	return amv.ValidateSnapshotSpec(c.Client, snapshot.Spec.SnapshotStorageSpec, snapshot.Namespace)
}

func (c *Controller) GetDatabase(snapshot *api.Snapshot) (runtime.Object, error) {
	mysql, err := c.ExtClient.MySQLs(snapshot.Namespace).Get(snapshot.Spec.DatabaseName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return mysql, nil
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

func (c *Controller) getVolumeForSnapshot(pvcSpec *core.PersistentVolumeClaimSpec, jobName, namespace string) (*core.Volume, error) {
	volume := &core.Volume{
		Name: "util-volume",
	}
	if pvcSpec != nil {
		if len(pvcSpec.AccessModes) == 0 {
			pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			}
			log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, *pvcSpec)
		}

		claim := &core.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: namespace,
			},
			Spec: *pvcSpec,
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
	} else {
		volume.EmptyDir = &core.EmptyDirVolumeSource{}
	}
	return volume, nil
}
