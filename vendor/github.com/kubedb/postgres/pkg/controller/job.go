package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/analytics"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storage "kmodules.xyz/objectstore-api/osm"
)

func (c *Controller) createRestoreJob(postgres *api.Postgres, snapshot *api.Snapshot) (*batch.Job, error) {
	postgresVersion, err := c.ExtClient.CatalogV1alpha1().PostgresVersions().Get(string(postgres.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := postgres.OffshootLabels()
	if jobLabel == nil {
		jobLabel = map[string]string{}
	}
	jobLabel[api.LabelDatabaseKind] = api.ResourceKindPostgres
	jobLabel[api.AnnotationJobType] = api.JobTypeRestore

	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(postgres.Spec.StorageType, postgres.Spec.Storage, jobName, postgres.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      jobLabel,
			Annotations: snapshot.Spec.PodTemplate.Controller.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: api.SchemeGroupVersion.String(),
					Kind:       api.ResourceKindPostgres,
					Name:       postgres.Name,
					UID:        postgres.UID,
				},
			},
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: snapshot.Spec.PodTemplate.Annotations,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            api.JobTypeRestore,
							Image:           postgresVersion.Spec.Tools.Image,
							ImagePullPolicy: core.PullIfNotPresent,
							Args: append([]string{
								api.JobTypeRestore,
								fmt.Sprintf(`--host=%s`, postgres.ServiceName()),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
								"--",
							}, postgres.Spec.Init.SnapshotSource.Args...),
							Env: []core.EnvVar{
								{
									Name: PostgresPassword,
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: postgres.Spec.DatabaseSecret.SecretName,
											},
											Key: PostgresPassword,
										},
									},
								},
								{
									Name:  analytics.Key,
									Value: c.AnalyticsClientID,
								},
							},
							Resources:      snapshot.Spec.PodTemplate.Spec.Resources,
							LivenessProbe:  snapshot.Spec.PodTemplate.Spec.LivenessProbe,
							ReadinessProbe: snapshot.Spec.PodTemplate.Spec.ReadinessProbe,
							Lifecycle:      snapshot.Spec.PodTemplate.Spec.Lifecycle,
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
					RestartPolicy:     core.RestartPolicyNever,
					NodeSelector:      snapshot.Spec.PodTemplate.Spec.NodeSelector,
					Affinity:          snapshot.Spec.PodTemplate.Spec.Affinity,
					SchedulerName:     snapshot.Spec.PodTemplate.Spec.SchedulerName,
					Tolerations:       snapshot.Spec.PodTemplate.Spec.Tolerations,
					PriorityClassName: snapshot.Spec.PodTemplate.Spec.PriorityClassName,
					Priority:          snapshot.Spec.PodTemplate.Spec.Priority,
					SecurityContext:   snapshot.Spec.PodTemplate.Spec.SecurityContext,
					ImagePullSecrets: core_util.MergeLocalObjectReferences(
						snapshot.Spec.PodTemplate.Spec.ImagePullSecrets,
						postgres.Spec.PodTemplate.Spec.ImagePullSecrets,
					),
				},
			},
		},
	}

	if snapshot.Spec.Backend.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.Backend.Local.MountPath,
			SubPath:   snapshot.Spec.Backend.Local.SubPath,
		})
		volume := core.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.Backend.Local.VolumeSource,
		}
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
	}
	return c.Client.BatchV1().Jobs(postgres.Namespace).Create(job)
}

func (c *Controller) GetSnapshotter(snapshot *api.Snapshot) (*batch.Job, error) {
	postgres, err := c.pgLister.Postgreses(snapshot.Namespace).Get(snapshot.Spec.DatabaseName)
	if err != nil {
		return nil, err
	}
	postgresVersion, err := c.ExtClient.CatalogV1alpha1().PostgresVersions().Get(string(postgres.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := postgres.OffshootLabels()
	if jobLabel == nil {
		jobLabel = map[string]string{}
	}
	jobLabel[api.LabelDatabaseKind] = api.ResourceKindPostgres
	jobLabel[api.AnnotationJobType] = api.JobTypeBackup

	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(postgres.Spec.StorageType, postgres.Spec.Storage, jobName, snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()
	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      jobLabel,
			Annotations: snapshot.Spec.PodTemplate.Controller.Annotations,
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
					Annotations: snapshot.Spec.PodTemplate.Annotations,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  api.JobTypeBackup,
							Image: postgresVersion.Spec.Tools.Image,
							Args: append([]string{
								api.JobTypeBackup,
								fmt.Sprintf(`--host=%s`, postgres.ServiceName()),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
								"--",
							}, snapshot.Spec.PodTemplate.Spec.Args...),
							Env: []core.EnvVar{
								{
									Name: PostgresPassword,
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: postgres.Spec.DatabaseSecret.SecretName,
											},
											Key: PostgresPassword,
										},
									},
								},
								{
									Name:  analytics.Key,
									Value: c.AnalyticsClientID,
								},
							},
							Resources:      snapshot.Spec.PodTemplate.Spec.Resources,
							LivenessProbe:  snapshot.Spec.PodTemplate.Spec.LivenessProbe,
							ReadinessProbe: snapshot.Spec.PodTemplate.Spec.ReadinessProbe,
							Lifecycle:      snapshot.Spec.PodTemplate.Spec.Lifecycle,
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
					RestartPolicy:     core.RestartPolicyNever,
					NodeSelector:      snapshot.Spec.PodTemplate.Spec.NodeSelector,
					Affinity:          snapshot.Spec.PodTemplate.Spec.Affinity,
					SchedulerName:     snapshot.Spec.PodTemplate.Spec.SchedulerName,
					Tolerations:       snapshot.Spec.PodTemplate.Spec.Tolerations,
					PriorityClassName: snapshot.Spec.PodTemplate.Spec.PriorityClassName,
					Priority:          snapshot.Spec.PodTemplate.Spec.Priority,
					SecurityContext:   snapshot.Spec.PodTemplate.Spec.SecurityContext,
					ImagePullSecrets: core_util.MergeLocalObjectReferences(
						snapshot.Spec.PodTemplate.Spec.ImagePullSecrets,
						postgres.Spec.PodTemplate.Spec.ImagePullSecrets,
					),
				},
			},
		},
	}

	if snapshot.Spec.Backend.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.Backend.Local.MountPath,
			SubPath:   snapshot.Spec.Backend.Local.SubPath,
		})
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, core.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.Backend.Local.VolumeSource,
		})
	}
	return job, nil
}

func (c *Controller) getVolumeForSnapshot(st api.StorageType, pvcSpec *core.PersistentVolumeClaimSpec, jobName, namespace string) (*core.Volume, error) {
	if st == api.StorageTypeEphemeral {
		ed := core.EmptyDirVolumeSource{}
		if pvcSpec != nil {
			if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
				ed.SizeLimit = &sz
			}
		}
		return &core.Volume{
			Name: "tools",
			VolumeSource: core.VolumeSource{
				EmptyDir: &ed,
			},
		}, nil
	}

	volume := &core.Volume{
		Name: "tools",
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
	return volume, nil
}
