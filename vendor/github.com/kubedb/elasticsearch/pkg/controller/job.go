package controller

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/analytics"
	storage "kmodules.xyz/objectstore-api/osm"
)

func (c *Controller) createRestoreJob(elasticsearch *api.Elasticsearch, snapshot *api.Snapshot) (*batch.Job, error) {
	elasticsearchVersion, err := c.ExtClient.CatalogV1alpha1().ElasticsearchVersions().Get(string(elasticsearch.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := elasticsearch.OffshootLabels()
	if jobLabel == nil {
		jobLabel = map[string]string{}
	}
	jobLabel[api.LabelDatabaseKind] = api.ResourceKindElasticsearch
	jobLabel[api.AnnotationJobType] = api.JobTypeRestore

	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	pvcSpec := snapshot.Spec.PodVolumeClaimSpec
	if pvcSpec == nil {
		if elasticsearch.Spec.Topology != nil {
			pvcSpec = elasticsearch.Spec.Topology.Data.Storage
		} else {
			pvcSpec = elasticsearch.Spec.Storage
		}
	}
	st := snapshot.Spec.StorageType
	if st == nil {
		st = &elasticsearch.Spec.StorageType
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
					Kind:       api.ResourceKindElasticsearch,
					Name:       elasticsearch.Name,
					UID:        elasticsearch.UID,
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
							Image:           elasticsearchVersion.Spec.Tools.Image,
							ImagePullPolicy: core.PullIfNotPresent,
							Args: append([]string{
								api.JobTypeRestore,
								fmt.Sprintf(`--host=%s`, elasticsearch.OffshootName()),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
								"--",
							}, elasticsearch.Spec.Init.SnapshotSource.Args...),
							Env: []core.EnvVar{
								{
									Name:  "DB_SCHEME",
									Value: elasticsearch.GetConnectionScheme(),
								},
								{
									Name: "DB_USER",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: elasticsearch.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyAdminUserName,
										},
									},
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
									Name:  "AUTH_PLUGIN",
									Value: fmt.Sprintf("%s", elasticsearch.Spec.AuthPlugin),
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
						elasticsearch.Spec.PodTemplate.Spec.ImagePullSecrets,
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
		job.Spec.Template.Spec.ServiceAccountName = elasticsearch.SnapshotSAName()
	}

	return c.Client.BatchV1().Jobs(elasticsearch.Namespace).Create(job)
}

func (c *Controller) GetSnapshotter(snapshot *api.Snapshot) (*batch.Job, error) {
	elasticsearch, err := c.ExtClient.KubedbV1alpha1().Elasticsearches(snapshot.Namespace).Get(snapshot.Spec.DatabaseName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	elasticsearchVersion, err := c.ExtClient.CatalogV1alpha1().ElasticsearchVersions().Get(string(elasticsearch.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := elasticsearch.OffshootLabels()
	if jobLabel == nil {
		jobLabel = map[string]string{}
	}
	jobLabel[api.LabelDatabaseKind] = api.ResourceKindElasticsearch
	jobLabel[api.AnnotationJobType] = api.JobTypeBackup

	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	pvcSpec := snapshot.Spec.PodVolumeClaimSpec
	if pvcSpec == nil {
		if elasticsearch.Spec.Topology != nil {
			pvcSpec = elasticsearch.Spec.Topology.Data.Storage
		} else {
			pvcSpec = elasticsearch.Spec.Storage
		}
	}
	st := snapshot.Spec.StorageType
	if st == nil {
		st = &elasticsearch.Spec.StorageType
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

	indices, err := c.getAllIndices(elasticsearch)
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
							Name:            api.JobTypeBackup,
							Image:           elasticsearchVersion.Spec.Tools.Image,
							ImagePullPolicy: core.PullIfNotPresent,
							Args: append([]string{
								api.JobTypeBackup,
								fmt.Sprintf(`--host=%s`, elasticsearch.OffshootName()),
								fmt.Sprintf(`--indices=%s`, indices),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
								"--",
							}, snapshot.Spec.PodTemplate.Spec.Args...),
							Env: []core.EnvVar{
								{
									Name:  "DB_SCHEME",
									Value: elasticsearch.GetConnectionScheme(),
								},
								{
									Name: "DB_USER",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: elasticsearch.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyReadAllUserName,
										},
									},
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
									Name:  "AUTH_PLUGIN",
									Value: fmt.Sprintf("%s", elasticsearch.Spec.AuthPlugin),
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
						elasticsearch.Spec.PodTemplate.Spec.ImagePullSecrets,
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
		job.Spec.Template.Spec.ServiceAccountName = elasticsearch.SnapshotSAName()
	}

	return job, nil
}
