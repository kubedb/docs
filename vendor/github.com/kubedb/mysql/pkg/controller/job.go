package controller

import (
	"fmt"

	"github.com/appscode/kutil/tools/analytics"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/storage"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	snapshotDumpDir        = "/var/data"
	snapshotProcessRestore = "restore"
	snapshotProcessBackup  = "backup"
)

func (c *Controller) createRestoreJob(mysql *api.MySQL, snapshot *api.Snapshot) (*batch.Job, error) {
	databaseName := mysql.Name
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindMySQL,
	}
	jobAnnotation := map[string]string{
		api.AnnotationJobType: api.JobTypeRestore,
	}

	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(mysql.Spec.Storage, jobName, mysql.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      jobLabel,
			Annotations: jobAnnotation,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: api.SchemeGroupVersion.String(),
					Kind:       api.ResourceKindMySQL,
					Name:       mysql.Name,
					UID:        mysql.UID,
				},
			},
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  snapshotProcessRestore,
							Image: c.docker.GetToolsImageWithTag(mysql),
							Args: []string{
								snapshotProcessRestore,
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--user=%s`, mysqlUser),
								fmt.Sprintf(`--data-dir=%s`, snapshotDumpDir),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
							},
							Env: []core.EnvVar{
								{
									Name:  analytics.Key,
									Value: c.AnalyticsClientID,
								},
								{
									Name: "DB_PASSWORD",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: mysql.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyMySQLPassword,
										},
									},
								},
							},
							Resources: snapshot.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      persistentVolume.Name,
									MountPath: snapshotDumpDir,
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
									ReadOnly:  true,
								},
							},
						},
					},
					ImagePullSecrets: mysql.Spec.ImagePullSecrets,
					Volumes: []core.Volume{
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
						{
							Name: "osmconfig",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: snapshot.OSMSecretName(),
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyNever,
				},
			},
		},
	}
	if snapshot.Spec.SnapshotStorageSpec.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.SnapshotStorageSpec.Local.MountPath,
			SubPath:   snapshot.Spec.SnapshotStorageSpec.Local.SubPath,
		})
		volume := core.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.SnapshotStorageSpec.Local.VolumeSource,
		}
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
	}
	return c.Client.BatchV1().Jobs(mysql.Namespace).Create(job)
}

func (c *Controller) getSnapshotterJob(snapshot *api.Snapshot) (*batch.Job, error) {
	databaseName := snapshot.Spec.DatabaseName
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindMySQL,
	}
	jobAnnotation := map[string]string{
		api.AnnotationJobType: snapshotProcessBackup,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}
	mysql, err := c.myLister.MySQLs(snapshot.Namespace).Get(databaseName)
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(mysql.Spec.Storage, jobName, snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()
	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      jobLabel,
			Annotations: jobAnnotation,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: api.SchemeGroupVersion.String(),
					Kind:       api.ResourceKindSnapshot,
					Name:       snapshot.Name,
					UID:        snapshot.UID,
				},
			},
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  snapshotProcessBackup,
							Image: c.docker.GetToolsImageWithTag(mysql),
							Args: []string{
								snapshotProcessBackup,
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--user=%s`, mysqlUser),
								fmt.Sprintf(`--data-dir=%s`, snapshotDumpDir),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
							},
							Env: []core.EnvVar{
								{
									Name:  analytics.Key,
									Value: c.AnalyticsClientID,
								},
								{
									Name: "DB_PASSWORD",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: mysql.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyMySQLPassword,
										},
									},
								},
							},
							Resources: snapshot.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      persistentVolume.Name,
									MountPath: snapshotDumpDir,
								},
								{
									Name:      "osmconfig",
									ReadOnly:  true,
									MountPath: storage.SecretMountPath,
								},
							},
						},
					},
					ImagePullSecrets: mysql.Spec.ImagePullSecrets,
					Volumes: []core.Volume{
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
						{
							Name: "osmconfig",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: snapshot.OSMSecretName(),
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyNever,
				},
			},
		},
	}
	if snapshot.Spec.SnapshotStorageSpec.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.SnapshotStorageSpec.Local.MountPath,
			SubPath:   snapshot.Spec.SnapshotStorageSpec.Local.SubPath,
		})
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, core.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.SnapshotStorageSpec.Local.VolumeSource,
		})
	}
	return job, nil
}
