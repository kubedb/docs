package controller

import (
	"path/filepath"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

// Initially mount configmap `mongodb.conf` on initialConfigDirectoryPath "/configdb-readonly".
// But, mongodb can't write this initial mounted file. Because, configmap mounted files is not writable.
// So, This initial file is copied to configDirectoryPath "/data/configdb" by init-container.
func (c *Controller) upsertConfigSourceVolume(statefulSet *apps.StatefulSet, mongodb *api.MongoDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB {
			statefulSet.Spec.Template.Spec.Containers[i].Args = meta_util.UpsertArgumentList(
				statefulSet.Spec.Template.Spec.Containers[i].Args,
				[]string{"--config=" + filepath.Join(configDirectoryPath, "mongod.conf")},
			)
		}
	}

	for i, container := range statefulSet.Spec.Template.Spec.InitContainers {
		if container.Name == InitInstallContainerName {
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = core_util.UpsertVolumeMount(
				statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts,
				core.VolumeMount{
					Name:      initialConfigDirectoryName,
					MountPath: initialConfigDirectoryPath,
				})
		}
	}

	statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
		statefulSet.Spec.Template.Spec.Volumes,
		core.Volume{
			Name:         initialConfigDirectoryName,
			VolumeSource: *mongodb.Spec.ConfigSource,
		})

	return statefulSet
}
