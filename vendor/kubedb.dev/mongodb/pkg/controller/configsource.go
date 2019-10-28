package controller

import (
	"path/filepath"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

// Initially mount configmap `mongodb.conf` on initialConfigDirectoryPath "/configdb-readonly".
// But, mongodb can't write this initial mounted file. Because, configmap mounted files is not writable.
// So, This initial file is copied to configDirectoryPath "/data/configdb" by init-container.
func (c *Controller) upsertConfigSourceVolume(template core.PodTemplateSpec, configSource *core.VolumeSource) core.PodTemplateSpec {
	for i, container := range template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB {
			template.Spec.Containers[i].Args = meta_util.UpsertArgumentList(
				template.Spec.Containers[i].Args,
				[]string{"--config=" + filepath.Join(configDirectoryPath, "mongod.conf")},
			)
		}
	}

	for i, container := range template.Spec.InitContainers {
		if container.Name == InitInstallContainerName {
			template.Spec.InitContainers[i].VolumeMounts = core_util.UpsertVolumeMount(
				template.Spec.InitContainers[i].VolumeMounts,
				core.VolumeMount{
					Name:      initialConfigDirectoryName,
					MountPath: initialConfigDirectoryPath,
				})
		}
	}

	template.Spec.Volumes = core_util.UpsertVolume(
		template.Spec.Volumes,
		core.Volume{
			Name:         initialConfigDirectoryName,
			VolumeSource: *configSource,
		})

	return template
}
