package controller

import (
	"errors"
	"fmt"

	"github.com/appscode/go/crypto/rand"
	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/docker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	batch "k8s.io/client-go/pkg/apis/batch/v1"
)

const (
	SnapshotProcess_Backup  = "backup"
	snapshotType_DumpBackup = "dump-backup"
	storageSecretMountPath  = "/var/credentials/"
)

func (c *Controller) ValidateSnapshot(snapshot *tapi.Snapshot) error {
	// Database name can't empty
	databaseName := snapshot.Spec.DatabaseName
	if databaseName == "" {
		return fmt.Errorf(`Object 'DatabaseName' is missing in '%v'`, snapshot.Spec)
	}

	if err := docker.CheckDockerImageVersion(docker.ImageElasticdump, c.opt.ElasticDumpTag); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, docker.ImageElasticdump, c.opt.ElasticDumpTag)
	}

	labelMap := map[string]string{
		amc.LabelDatabaseKind:   tapi.ResourceKindElastic,
		amc.LabelDatabaseName:   snapshot.Spec.DatabaseName,
		amc.LabelSnapshotStatus: string(tapi.DatabasePhaseRunning),
	}

	snapshotList, err := c.ExtClient.Snapshots(snapshot.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	})
	if err != nil {
		return err
	}

	if len(snapshotList.Items) > 0 {
		if snapshot, err = c.ExtClient.Snapshots(snapshot.Namespace).Get(snapshot.Name); err != nil {
			return err
		}

		t := metav1.Now()
		snapshot.Status.StartTime = &t
		snapshot.Status.CompletionTime = &t
		snapshot.Status.Phase = tapi.SnapshotPhaseFailed
		snapshot.Status.Reason = "One Snapshot is already Running"
		if _, err := c.ExtClient.Snapshots(snapshot.Namespace).Update(snapshot); err != nil {
			return err
		}
		return errors.New("One Snapshot is already Running")
	}

	snapshotSpec := snapshot.Spec.SnapshotStorageSpec
	if err := c.ValidateSnapshotSpec(snapshotSpec); err != nil {
		return err
	}

	if err := c.CheckBucketAccess(snapshot.Spec.SnapshotStorageSpec, snapshot.Namespace); err != nil {
		return err
	}
	return nil
}

func (c *Controller) GetDatabase(snapshot *tapi.Snapshot) (runtime.Object, error) {
	return c.ExtClient.Elastics(snapshot.Namespace).Get(snapshot.Spec.DatabaseName)
}

func (c *Controller) GetSnapshotter(snapshot *tapi.Snapshot) (*batch.Job, error) {
	databaseName := snapshot.Spec.DatabaseName
	jobName := rand.WithUniqSuffix(databaseName)
	jobLabel := map[string]string{
		amc.LabelDatabaseName: databaseName,
		amc.LabelJobType:      SnapshotProcess_Backup,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec

	elastic, err := c.ExtClient.Elastics(snapshot.Namespace).Get(databaseName)
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(elastic.Spec.Storage, jobName, snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName := fmt.Sprintf("%v/%v/%v", amc.DatabaseNamePrefix, snapshot.Namespace, snapshot.Spec.DatabaseName)
	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: batch.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  SnapshotProcess_Backup,
							Image: docker.ImageElasticdump + ":" + c.opt.ElasticDumpTag,
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Backup),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, backupSpec.BucketName),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "cloud",
									MountPath: storageSecretMountPath,
								},
								{
									Name:      persistentVolume.Name,
									MountPath: "/var/" + snapshotType_DumpBackup + "/",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "cloud",
							VolumeSource: apiv1.VolumeSource{
								Secret: backupSpec.StorageSecret,
							},
						},
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}
	return job, nil
}

func (c *Controller) WipeOutSnapshot(snapshot *tapi.Snapshot) error {
	return c.DeleteSnapshotData(snapshot)
}

func (c *Controller) getVolumeForSnapshot(storage *tapi.StorageSpec, jobName, namespace string) (*apiv1.Volume, error) {
	volume := &apiv1.Volume{
		Name: "util-volume",
	}
	if storage != nil {
		claim := &apiv1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: namespace,
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": storage.Class,
				},
			},
			Spec: storage.PersistentVolumeClaimSpec,
		}

		if _, err := c.Client.Core().PersistentVolumeClaims(claim.Namespace).Create(claim); err != nil {
			return nil, err
		}

		volume.PersistentVolumeClaim = &apiv1.PersistentVolumeClaimVolumeSource{
			ClaimName: claim.Name,
		}
	} else {
		volume.EmptyDir = &apiv1.EmptyDirVolumeSource{}
	}
	return volume, nil
}
