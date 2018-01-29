package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/analytics"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/storage"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	snapshotProcessBackup  = "backup"
	snapshotProcessRestore = "restore"
)

func (c *Controller) createRestoreJob(elasticsearch *api.Elasticsearch, snapshot *api.Snapshot) (*batch.Job, error) {
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindElasticsearch,
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
	persistentVolume, err := c.getVolumeForSnapshot(elasticsearch.Spec.Storage, jobName, elasticsearch.Namespace)
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
					Kind:       api.ResourceKindElasticsearch,
					Name:       elasticsearch.Name,
					UID:        elasticsearch.UID,
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
							Name:            snapshotProcessRestore,
							Image:           c.opt.Docker.GetToolsImageWithTag(elasticsearch),
							ImagePullPolicy: core.PullIfNotPresent,
							Args: []string{
								snapshotProcessRestore,
								fmt.Sprintf(`--host=%s`, elasticsearch.OffshootName()),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf("--analytics=%v", c.opt.EnableAnalytics),
							},
							Env: []core.EnvVar{
								{
									Name:  "DB_USER",
									Value: AdminUser,
								},
								{
									Name: "DB_PASSWORD",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: elasticsearch.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyAdminPassword,
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
									MountPath: "/var/data",
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
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
	elasticsearch, err := c.ExtClient.Elasticsearchs(snapshot.Namespace).Get(snapshot.Spec.DatabaseName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindElasticsearch,
	}
	jobAnnotation := map[string]string{
		api.AnnotationJobType: api.JobTypeBackup,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
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

	indices, err := c.getAllIndices(elasticsearch)
	if err != nil {
		return nil, err
	}

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
							Name:            snapshotProcessBackup,
							Image:           c.opt.Docker.GetToolsImageWithTag(elasticsearch),
							ImagePullPolicy: core.PullAlways,
							Args: []string{
								snapshotProcessBackup,
								fmt.Sprintf(`--host=%s`, elasticsearch.OffshootName()),
								fmt.Sprintf(`--indices=%s`, indices),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf("--analytics=%v", c.opt.EnableAnalytics),
							},
							Env: []core.EnvVar{
								{
									Name:  "DB_USER",
									Value: ReadAllUser,
								},
								{
									Name: "DB_PASSWORD",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: elasticsearch.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyReadAllPassword,
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
									MountPath: "/var/data",
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
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

func (c *Controller) getVolumeForSnapshot(pvcSpec *core.PersistentVolumeClaimSpec, jobName, namespace string) (*core.Volume, error) {
	volume := &core.Volume{
		Name: "tools",
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
