/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"path/filepath"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

// Initially mount configmap `mongodb.conf` on initialConfigDirectoryPath "/configdb-readonly".
// But, mongodb can't write this initial mounted file. Because, configmap mounted files is not writable.
// So, This initial file is copied to configDirectoryPath "/data/configdb" by init-container.
func (c *Reconciler) upsertConfigSecretVolume(template core.PodTemplateSpec, configSecret *core.LocalObjectReference) core.PodTemplateSpec {
	for i, container := range template.Spec.Containers {
		if container.Name == api.MongoDBContainerName {
			template.Spec.Containers[i].Args = meta_util.UpsertArgumentList(
				template.Spec.Containers[i].Args,
				[]string{"--config=" + filepath.Join(api.MongoDBConfigDirectoryPath, api.MongoDBCustomConfigFile)},
			)
		}
	}

	for i, container := range template.Spec.InitContainers {
		if container.Name == api.MongoDBInitInstallContainerName {
			template.Spec.InitContainers[i].VolumeMounts = core_util.UpsertVolumeMount(
				template.Spec.InitContainers[i].VolumeMounts,
				core.VolumeMount{
					Name:      api.MongoDBInitialConfigDirectoryName,
					MountPath: api.MongoDBInitialConfigDirectoryPath,
				})
		}
	}

	template.Spec.Volumes = core_util.UpsertVolume(
		template.Spec.Volumes,
		core.Volume{
			Name: api.MongoDBInitialConfigDirectoryName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: configSecret.Name,
				},
			},
		})

	return template
}
