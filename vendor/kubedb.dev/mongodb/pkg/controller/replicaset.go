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
	"fmt"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (c *Controller) topologyInitContainer(
	mongodb *api.MongoDB,
	mongodbVersion *catalog.MongoDBVersion,
	podTemplate *ofst.PodTemplateSpec,
	repSetName string,
	gvrSvc string,
	scriptName string,
) (core.Container, []core.Volume) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if podTemplate != nil {
		pt = *podTemplate
	}

	// mongodb.Spec.SSLMode & mongodb.Spec.ClusterAuthMode can be empty if upgraded operator from
	// previous version. But, eventually it will be defaulted. TODO: delete in future.
	sslMode := mongodb.Spec.SSLMode
	if sslMode == "" {
		sslMode = api.SSLModeDisabled
	}
	clusterAuth := mongodb.Spec.ClusterAuthMode
	if clusterAuth == "" {
		clusterAuth = api.ClusterAuthModeKeyFile
		if sslMode != api.SSLModeDisabled {
			clusterAuth = api.ClusterAuthModeX509
		}
	}

	bootstrapContainer := core.Container{
		Name:            api.MongoDBInitBootstrapContainerName,
		Image:           mongodbVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{fmt.Sprintf("%v/peer-finder", InitScriptDirectoryPath)},
		Args: []string{
			fmt.Sprintf("-on-start=%v/%v", InitScriptDirectoryPath, scriptName),
			"-service=" + gvrSvc,
		},
		Env: core_util.UpsertEnvVars([]core.EnvVar{
			{
				Name: "POD_NAMESPACE",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.namespace",
					},
				},
			},
			{
				Name:  "REPLICA_SET",
				Value: repSetName,
			},
			{
				Name:  "AUTH",
				Value: "true",
			},
			{
				Name:  "SSL_MODE",
				Value: string(sslMode),
			},
			{
				Name:  "CLUSTER_AUTH_MODE",
				Value: string(clusterAuth),
			},
			{
				Name: "MONGO_INITDB_ROOT_USERNAME",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: mongodb.Spec.DatabaseSecret.SecretName,
						},
						Key: core.BasicAuthUsernameKey,
					},
				},
			},
			{
				Name: "MONGO_INITDB_ROOT_PASSWORD",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: mongodb.Spec.DatabaseSecret.SecretName,
						},
						Key: core.BasicAuthPasswordKey,
					},
				},
			},
		}, pt.Spec.Env...),
		VolumeMounts: []core.VolumeMount{
			{
				Name:      workDirectoryName,
				MountPath: workDirectoryPath,
			},
			{
				Name:      configDirectoryName,
				MountPath: configDirectoryPath,
			},
			{
				Name:      certDirectoryName,
				MountPath: api.MongoCertDirectory,
			},
			{
				Name:      dataDirectoryName,
				MountPath: dataDirectoryPath,
			},
			{
				Name:      InitScriptDirectoryName,
				MountPath: InitScriptDirectoryPath,
			},
		},
		Resources: pt.Spec.Resources,
	}

	var rsVolumes []core.Volume

	if mongodb.Spec.KeyFile != nil {
		rsVolumes = append(rsVolumes, core.Volume{
			Name: initialKeyDirectoryName, // FIXIT: mounted where?
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					DefaultMode: types.Int32P(0400),
					SecretName:  mongodb.Spec.KeyFile.SecretName,
				},
			},
		})
	}

	//only on mongos in case of sharding (which is handled on 'ensureMongosNode'.
	if mongodb.Spec.ShardTopology == nil && mongodb.Spec.Init != nil && mongodb.Spec.Init.Script != nil {
		rsVolumes = append(rsVolumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: mongodb.Spec.Init.Script.VolumeSource,
		})

		bootstrapContainer.VolumeMounts = core_util.UpsertVolumeMount(
			bootstrapContainer.VolumeMounts,
			core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/docker-entrypoint-initdb.d",
			},
		)
	}

	return bootstrapContainer, rsVolumes
}
