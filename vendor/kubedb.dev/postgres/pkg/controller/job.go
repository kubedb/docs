package controller

import (
	"fmt"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/analytics"
	storage "kmodules.xyz/objectstore-api/osm"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
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
	pvcSpec := snapshot.Spec.PodVolumeClaimSpec
	if pvcSpec == nil {
		pvcSpec = postgres.Spec.Storage
	}
	st := snapshot.Spec.StorageType
	if st == nil {
		st = &postgres.Spec.StorageType
	}
	persistentVolume, err := c.GetVolumeForSnapshot(*st, pvcSpec, jobName, snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, err := snapshot.Location()
	if err != nil {
		return nil, err
	}

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
									Name: PostgresUser,
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: postgres.Spec.DatabaseSecret.SecretName,
											},
											Key: PostgresUser,
										},
									},
								},
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

	if c.EnableRBAC {
		if snapshot.Spec.PodTemplate.Spec.ServiceAccountName == "" {
			if err := c.ensureSnapshotRBAC(postgres); err != nil {
				return nil, err
			}
			job.Spec.Template.Spec.ServiceAccountName = postgres.SnapshotSAName()
		} else {
			job.Spec.Template.Spec.ServiceAccountName = snapshot.Spec.PodTemplate.Spec.ServiceAccountName
		}
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
	pvcSpec := snapshot.Spec.PodVolumeClaimSpec
	if pvcSpec == nil {
		pvcSpec = postgres.Spec.Storage
	}
	st := snapshot.Spec.StorageType
	if st == nil {
		st = &postgres.Spec.StorageType
	}
	persistentVolume, err := c.GetVolumeForSnapshot(*st, pvcSpec, jobName, snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, err := snapshot.Location()
	if err != nil {
		return nil, err
	}

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
									Name: PostgresUser,
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: postgres.Spec.DatabaseSecret.SecretName,
											},
											Key: PostgresUser,
										},
									},
								},
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

	if c.EnableRBAC {
		if snapshot.Spec.PodTemplate.Spec.ServiceAccountName == "" {
			job.Spec.Template.Spec.ServiceAccountName = postgres.SnapshotSAName()
			if err := c.ensureSnapshotRBAC(postgres); err != nil {
				return nil, err
			}
		} else {
			job.Spec.Template.Spec.ServiceAccountName = snapshot.Spec.PodTemplate.Spec.ServiceAccountName
		}
	}

	return job, nil
}
