package controller

import (
	"fmt"
	"path/filepath"
	"strings"

	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/analytics"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	etcdutil "github.com/kubedb/etcd/pkg/util"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	storage "kmodules.xyz/objectstore-api/osm"
)

const (
	snapshotDumpDir = etcdVolumeMountDir + "/snapshot"
)

func (c *Controller) getRestoreContainer(etcd *api.Etcd, snapshot *api.Snapshot, m *etcdutil.Member, ms etcdutil.MemberSet) ([]core.Container, error) {
	etcdVersion, err := c.ExtClient.EtcdVersions().Get(string(etcd.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	snapshotSource := etcd.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
		c.recorder.Eventf(
			ref,
			core.EventTypeNormal,
			eventer.EventReasonInitializing,
			`Initializing from Snapshot: "%v"`,
			snapshotSource.Name,
		)
	}
	var containers []core.Container

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = etcd.Namespace
	}

	endpoints := fmt.Sprintf("%s.%s", etcd.ClientServiceName(), etcd.Namespace)
	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}
	folderName, _ := snapshot.Location()

	containers = append(containers, core.Container{
		Name:  api.JobTypeRestore,
		Image: etcdVersion.Spec.DB.Image,
		Args: meta_util.UpsertArgumentList([]string{
			api.JobTypeRestore,
			fmt.Sprintf(`--host=%s`, endpoints),
			fmt.Sprintf(`--data-dir=%s`, snapshotDumpDir),
			fmt.Sprintf(`--bucket=%s`, bucket),
			fmt.Sprintf(`--folder=%s`, folderName),
			fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
			fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
		}, snapshot.Spec.PodTemplate.Spec.Args, "--enable-analytics"),
		Env: []core.EnvVar{
			{
				Name:  analytics.Key,
				Value: c.AnalyticsClientID,
			},
		},
		Resources: snapshot.Spec.PodTemplate.Spec.Resources,
		VolumeMounts: []core.VolumeMount{
			{
				Name:      "data",
				MountPath: etcdVolumeMountDir,
			},
			{
				Name:      "osmconfig",
				MountPath: storage.SecretMountPath,
				ReadOnly:  true,
			},
		},
	})

	containers = append(containers, core.Container{
		Name:  "restore-datadir",
		Image: etcdVersion.Spec.DB.Image,
		Command: []string{
			"/bin/sh", "-ec",
			fmt.Sprintf("ETCDCTL_API=3 etcdctl snapshot restore %[1]s"+
				" --name %[2]s"+
				" --initial-cluster %[3]s"+
				" --initial-cluster-token %[5]s"+
				" --initial-advertise-peer-urls %[4]s"+
				" --data-dir %[6]s 2>/dev/termination-log", filepath.Join(snapshotDumpDir, snapshot.Name), m.Name, strings.Join(ms.PeerURLPairs(), ","), m.PeerURL(), etcd.Name, dataDir),
		},
		Resources: snapshot.Spec.PodTemplate.Spec.Resources,
		VolumeMounts: []core.VolumeMount{
			{
				Name:      "data",
				MountPath: etcdVolumeMountDir,
			},
		},
	})

	return containers, nil
}

func (c *Controller) getSnapshotterJob(snapshot *api.Snapshot) (*batch.Job, error) {
	etcd, err := c.etcdLister.Etcds(snapshot.Namespace).Get(snapshot.Spec.DatabaseName)
	if err != nil {
		return nil, err
	}
	etcdVersion, err := c.ExtClient.EtcdVersions().Get(string(etcd.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshot.OffshootName())
	jobLabel := etcd.OffshootLabels()
	if jobLabel == nil {
		jobLabel = map[string]string{}
	}
	jobLabel[api.LabelDatabaseKind] = api.ResourceKindEtcd
	jobLabel[api.AnnotationJobType] = api.JobTypeBackup

	backupSpec := snapshot.Spec.Backend
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(etcd.Spec.StorageType, etcd.Spec.Storage, jobName, snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	endpoints := fmt.Sprintf("%s.%s", etcd.ClientServiceName(), etcd.Namespace)

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
							Image: etcdVersion.Spec.Tools.Image,
							Args: meta_util.UpsertArgumentList([]string{
								api.JobTypeBackup,
								fmt.Sprintf(`--host=%s`, endpoints),
								// fmt.Sprintf(`--user=%s`, etcdUser),
								fmt.Sprintf(`--data-dir=%s`, snapshotDumpDir),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
								fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
							}, snapshot.Spec.PodTemplate.Spec.Args, "--enable-analytics"),
							Env: []core.EnvVar{
								{
									Name:  analytics.Key,
									Value: c.AnalyticsClientID,
								},
							},
							Resources: snapshot.Spec.PodTemplate.Spec.Resources,
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
						etcd.Spec.PodTemplate.Spec.ImagePullSecrets,
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
