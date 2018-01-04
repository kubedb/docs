package controller

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/kutil/tools/analytics"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/storage"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	snapshotProcessRestore  = "restore"
	snapshotTypeDumpRestore = "dump-restore"
)

func (c *Controller) createRestoreJob(elasticsearch *api.Elasticsearch, snapshot *api.Snapshot) (*batch.Job, error) {
	databaseName := elasticsearch.Name
	jobName := rand.WithUniqSuffix(snapshot.OffshootName())
	jobLabel := map[string]string{
		api.LabelDatabaseName: databaseName,
		api.LabelJobType:      snapshotProcessRestore,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(elasticsearch.Spec.Storage, jobName, elasticsearch.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            snapshotProcessRestore,
							Image:           c.opt.Docker.GetToolsImageWithTag(elasticsearch),
							ImagePullPolicy: core.PullIfNotPresent,
							Args: []string{
								fmt.Sprintf(`--process=%s`, snapshotProcessRestore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							Env: []core.EnvVar{
								{
									Name:  "USERNAME",
									Value: "admin",
								},
								{
									Name: "PASSWORD",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: elasticsearch.Spec.DatabaseSecret.SecretName,
											},
											Key: "ADMIN_PASSWORD",
										},
									},
								},
								{
									Name:  analytics.Key,
									Value: c.opt.AnalyticsClientID,
								},
							},
							Resources: snapshot.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      persistentVolume.Name,
									MountPath: "/var/" + snapshotTypeDumpRestore + "/",
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
									ReadOnly:  true,
								},
								{
									Name:      "certs",
									MountPath: "/certs",
									ReadOnly:  true,
								},
							},
						},
					},
					ImagePullSecrets: elasticsearch.Spec.ImagePullSecrets,
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
						{
							Name: "certs",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: elasticsearch.Spec.CertificateSecret.SecretName,
									Items: []core.KeyToPath{
										{
											Key:  "ca.pem",
											Path: "ca.pem",
										},
										{
											Key:  "client-key.pem",
											Path: "client-key.pem",
										},
										{
											Key:  "client.pem",
											Path: "client.pem",
										},
									},
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
	return c.Client.BatchV1().Jobs(elasticsearch.Namespace).Create(job)
}

func (c *Controller) GetSnapshotter(snapshot *api.Snapshot) (*batch.Job, error) {
	databaseName := snapshot.Spec.DatabaseName
	jobName := rand.WithUniqSuffix(snapshot.OffshootName())
	jobLabel := map[string]string{
		api.LabelDatabaseName: databaseName,
		api.LabelJobType:      snapshotProcessBackup,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}
	elasticsearch, err := c.ExtClient.Elasticsearchs(snapshot.Namespace).Get(databaseName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(elasticsearch.Spec.Storage, jobName, snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()
	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
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
							Image: c.opt.Docker.GetToolsImageWithTag(elasticsearch),
							Args: []string{
								fmt.Sprintf(`--process=%s`, snapshotProcessBackup),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							Env: []core.EnvVar{
								{
									Name:  "USERNAME",
									Value: "readall",
								},
								{
									Name: "PASSWORD",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: elasticsearch.Spec.DatabaseSecret.SecretName,
											},
											Key: "READALL_PASSWORD",
										},
									},
								},
								{
									Name:  analytics.Key,
									Value: c.opt.AnalyticsClientID,
								},
							},
							Resources: snapshot.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      persistentVolume.Name,
									MountPath: "/var/" + snapshotTypeDumpBackup + "/",
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
									ReadOnly:  true,
								},
								{
									Name:      "certs",
									MountPath: "/certs",
									ReadOnly:  true,
								},
							},
						},
					},
					ImagePullSecrets: elasticsearch.Spec.ImagePullSecrets,
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
						{
							Name: "certs",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: elasticsearch.Spec.CertificateSecret.SecretName,
									Items: []core.KeyToPath{
										{
											Key:  "ca.pem",
											Path: "ca.pem",
										},
										{
											Key:  "client-key.pem",
											Path: "client-key.pem",
										},
										{
											Key:  "client.pem",
											Path: "client.pem",
										},
									},
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
