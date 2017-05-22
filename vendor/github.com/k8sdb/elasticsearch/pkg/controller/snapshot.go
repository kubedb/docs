package controller

import (
	"errors"
	"fmt"

	"github.com/appscode/go/crypto/rand"
	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kbatch "k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
)

const (
	ImageElasticDump        = "k8sdb/elasticdump"
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

	if err := amc.CheckDockerImageVersion(ImageElasticDump, c.elasticDumpTag); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, ImageElasticDump, c.elasticDumpTag)
	}

	labelMap := map[string]string{
		amc.LabelDatabaseKind:   tapi.ResourceKindElastic,
		amc.LabelDatabaseName:   snapshot.Spec.DatabaseName,
		amc.LabelSnapshotStatus: string(tapi.DatabasePhaseRunning),
	}

	snapshotList, err := c.ExtClient.Snapshots(snapshot.Namespace).List(kapi.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set(labelMap)),
	})
	if err != nil {
		return err
	}

	if len(snapshotList.Items) > 0 {
		if snapshot, err = c.ExtClient.Snapshots(snapshot.Namespace).Get(snapshot.Name); err != nil {
			return err
		}

		t := unversioned.Now()
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

func (c *Controller) GetSnapshotter(snapshot *tapi.Snapshot) (*kbatch.Job, error) {
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
	job := &kbatch.Job{
		ObjectMeta: kapi.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: kbatch.JobSpec{
			Template: kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: kapi.PodSpec{
					Containers: []kapi.Container{
						{
							Name:  SnapshotProcess_Backup,
							Image: ImageElasticDump + ":" + c.elasticDumpTag,
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Backup),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, backupSpec.BucketName),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							VolumeMounts: []kapi.VolumeMount{
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
					Volumes: []kapi.Volume{
						{
							Name: "cloud",
							VolumeSource: kapi.VolumeSource{
								Secret: backupSpec.StorageSecret,
							},
						},
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
					},
					RestartPolicy: kapi.RestartPolicyNever,
				},
			},
		},
	}
	return job, nil
}

func (c *Controller) WipeOutSnapshot(snapshot *tapi.Snapshot) error {
	return c.DeleteSnapshotData(snapshot)
}

func (c *Controller) getVolumeForSnapshot(storage *tapi.StorageSpec, jobName, namespace string) (*kapi.Volume, error) {
	volume := &kapi.Volume{
		Name: "util-volume",
	}
	if storage != nil {
		claim := &kapi.PersistentVolumeClaim{
			ObjectMeta: kapi.ObjectMeta{
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

		volume.PersistentVolumeClaim = &kapi.PersistentVolumeClaimVolumeSource{
			ClaimName: claim.Name,
		}
	} else {
		volume.EmptyDir = &kapi.EmptyDirVolumeSource{}
	}
	return volume, nil
}
