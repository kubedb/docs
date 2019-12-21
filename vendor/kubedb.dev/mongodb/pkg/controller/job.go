/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/analytics"
	storage "kmodules.xyz/objectstore-api/osm"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	snapshotDumpDir  = "/var/data"
	restoreConfigArg = "--skip-config"
)

func (c *Controller) createRestoreJob(mongodb *api.MongoDB, snapshot *api.Snapshot) (*batch.Job, error) {
	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := mongodb.OffshootLabels()
	if jobLabel == nil {
		jobLabel = map[string]string{}
	}
	jobLabel[api.LabelDatabaseKind] = api.ResourceKindMongoDB
	jobLabel[api.AnnotationJobType] = api.JobTypeRestore

	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	pvcSpec := snapshot.Spec.PodVolumeClaimSpec
	if pvcSpec == nil {
		pvcSpec = snapshotStorageSize(mongodb)
	}
	st := snapshot.Spec.StorageType
	if st == nil {
		st = &mongodb.Spec.StorageType
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

	// Take podTemplate from either Shard or spec.PodTemplate
	var dbPodTemplate ofst.PodTemplateSpec
	if mongodb.Spec.ShardTopology != nil {
		dbPodTemplate = mongodb.Spec.ShardTopology.Shard.PodTemplate
	} else if mongodb.Spec.PodTemplate != nil {
		dbPodTemplate = *mongodb.Spec.PodTemplate
	}

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      jobLabel,
			Annotations: snapshot.Spec.PodTemplate.Controller.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: api.SchemeGroupVersion.String(),
					Kind:       api.ResourceKindMongoDB,
					Name:       mongodb.Name,
					UID:        mongodb.UID,
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
							Name:  api.JobTypeRestore,
							Image: mongodbVersion.Spec.Tools.Image,
							Args: append([]string{
								api.JobTypeRestore,
								fmt.Sprintf(`--host=%s`, mongodb.HostAddress()),
								fmt.Sprintf(`--data-dir=%s`, snapshotDumpDir),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
							}, getRestoreConfigArg(mongodb)...),
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
												Name: mongodb.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyMongoDBPassword,
										},
									},
								},
								{
									Name: "DB_USER",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: mongodb.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyMongoDBUser,
										},
									},
								},
							},
							Resources:      snapshot.Spec.PodTemplate.Spec.Resources,
							LivenessProbe:  snapshot.Spec.PodTemplate.Spec.LivenessProbe,
							ReadinessProbe: snapshot.Spec.PodTemplate.Spec.ReadinessProbe,
							Lifecycle:      snapshot.Spec.PodTemplate.Spec.Lifecycle,
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
						dbPodTemplate.Spec.ImagePullSecrets,
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

	if snapshot.Spec.PodTemplate.Spec.ServiceAccountName == "" {
		if err := c.ensureSnapshotRBAC(mongodb); err != nil {
			return nil, err
		}
		job.Spec.Template.Spec.ServiceAccountName = mongodb.SnapshotSAName()
	} else {
		job.Spec.Template.Spec.ServiceAccountName = snapshot.Spec.PodTemplate.Spec.ServiceAccountName
	}

	return c.Client.BatchV1().Jobs(mongodb.Namespace).Create(job)
}

func (c *Controller) getSnapshotterJob(snapshot *api.Snapshot) (*batch.Job, error) {
	mongodb, err := c.mgLister.MongoDBs(snapshot.Namespace).Get(snapshot.Spec.DatabaseName)
	if err != nil {
		return nil, err
	}
	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := mongodb.OffshootLabels()
	if jobLabel == nil {
		jobLabel = map[string]string{}
	}
	jobLabel[api.LabelDatabaseKind] = api.ResourceKindMongoDB
	jobLabel[api.AnnotationJobType] = api.JobTypeBackup

	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	pvcSpec := snapshot.Spec.PodVolumeClaimSpec
	if pvcSpec == nil {
		pvcSpec = snapshotStorageSize(mongodb)
	}
	st := snapshot.Spec.StorageType
	if st == nil {
		st = &mongodb.Spec.StorageType
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

	// Take podTemplate from either Shard or spec.PodTemplate
	var dbPodTemplate ofst.PodTemplateSpec
	if mongodb.Spec.ShardTopology != nil {
		dbPodTemplate = mongodb.Spec.ShardTopology.Shard.PodTemplate
	} else if mongodb.Spec.PodTemplate != nil {
		dbPodTemplate = *mongodb.Spec.PodTemplate
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
							Image: mongodbVersion.Spec.Tools.Image,
							Args: append([]string{
								api.JobTypeBackup,
								fmt.Sprintf(`--host=%s`, mongodb.HostAddress()),
								fmt.Sprintf(`--data-dir=%s`, snapshotDumpDir),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
								"--",
							}, snapshot.Spec.PodTemplate.Spec.Args...),
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
												Name: mongodb.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyMongoDBPassword,
										},
									},
								},
								{
									Name: "DB_USER",
									ValueFrom: &core.EnvVarSource{
										SecretKeyRef: &core.SecretKeySelector{
											LocalObjectReference: core.LocalObjectReference{
												Name: mongodb.Spec.DatabaseSecret.SecretName,
											},
											Key: KeyMongoDBUser,
										},
									},
								},
							},
							Resources:      snapshot.Spec.PodTemplate.Spec.Resources,
							LivenessProbe:  snapshot.Spec.PodTemplate.Spec.LivenessProbe,
							ReadinessProbe: snapshot.Spec.PodTemplate.Spec.ReadinessProbe,
							Lifecycle:      snapshot.Spec.PodTemplate.Spec.Lifecycle,
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
						dbPodTemplate.Spec.ImagePullSecrets,
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

	if snapshot.Spec.PodTemplate.Spec.ServiceAccountName == "" {
		job.Spec.Template.Spec.ServiceAccountName = mongodb.SnapshotSAName()
		if err := c.ensureSnapshotRBAC(mongodb); err != nil {
			return nil, err
		}
	} else {
		job.Spec.Template.Spec.ServiceAccountName = snapshot.Spec.PodTemplate.Spec.ServiceAccountName
	}

	return job, nil
}

func snapshotStorageSize(db *api.MongoDB) *core.PersistentVolumeClaimSpec {
	topology := db.Spec.ShardTopology
	if topology != nil {
		stg := topology.Shard.Storage
		stgSizePtr := stg.Resources.Requests[core.ResourceStorage]
		for i := int32(0); i < topology.Shard.Shards; i++ {
			stgSizePtr.Add(stg.Resources.Requests[core.ResourceStorage])
		}
		stg.Resources.Requests[core.ResourceStorage] = stgSizePtr
		return stg
	}
	return db.Spec.Storage
}

func getRestoreConfigArg(db *api.MongoDB) []string {
	args := append([]string{"--"}, db.Spec.Init.SnapshotSource.Args...)
	for in, val := range args {
		if strings.HasPrefix(val, string(restoreConfigArg)) {
			// Move '--skip-config=*' to the beginning (before of '--') of argument list
			args = append(args[:in], args[in+1:]...)
			args = append([]string{val}, args...)
			break
		}
	}
	return args
}
