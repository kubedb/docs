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
	"context"
	"fmt"
	"strconv"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/types"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (c *Controller) ensureMongosNode(mongodb *api.MongoDB) (*apps.StatefulSet, kutil.VerbType, error) {
	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
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

	cmds := []string{"mongos"}
	args := []string{
		"--bind_ip=0.0.0.0",
		"--port=" + strconv.Itoa(MongoDBPort),
		"--configdb=$(CONFIGDB_REPSET)",
		"--clusterAuthMode=" + string(clusterAuth),
		"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
	}

	sslArgs, err := c.getTLSArgs(mongodb, mongodbVersion)
	if err != nil {
		return &apps.StatefulSet{}, "", err
	}
	args = append(args, sslArgs...)

	// shardDsn List, separated by space ' '
	var shardDsn string
	for i := int32(0); i < mongodb.Spec.ShardTopology.Shard.Shards; i++ {
		if i != 0 {
			shardDsn += " "
		}
		shardDsn += mongodb.ShardDSN(i)
	}

	envList := []core.EnvVar{
		{
			Name:  "CONFIGDB_REPSET",
			Value: mongodb.ConfigSvrDSN(),
		},
		{
			Name:  "SHARD_REPSETS",
			Value: shardDsn,
		},
		{
			Name:  "SERVICE_NAME",
			Value: mongodb.ServiceName(),
		},
	}

	initContnr, initvolumes := installInitContainer(
		mongodb,
		mongodbVersion,
		&mongodb.Spec.ShardTopology.Mongos.PodTemplate,
		mongodb.MongosNodeName(),
	)

	var initContainers []core.Container
	var volumes []core.Volume

	volumes = append(volumes, core.Volume{
		Name: configDirectoryName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	})

	volumeMounts := []core.VolumeMount{
		{
			Name:      configDirectoryName,
			MountPath: configDirectoryPath,
		},
	}

	initContainers = append(initContainers, initContnr)
	volumes = core_util.UpsertVolume(volumes, initvolumes...)

	if mongodb.Spec.Init != nil && mongodb.Spec.Init.Script != nil {
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: mongodb.Spec.Init.Script.VolumeSource,
		})

		volumeMounts = append(
			volumeMounts,
			core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/docker-entrypoint-initdb.d",
			},
		)
	}

	bootstrpContnr, bootstrpVol := mongosInitContainer(
		mongodb,
		mongodbVersion,
		mongodb.Spec.ShardTopology.Mongos.PodTemplate,
		envList,
		"mongos.sh",
	)
	initContainers = append(initContainers, bootstrpContnr)
	volumes = core_util.UpsertVolume(volumes, bootstrpVol...)

	opts := workloadOptions{
		stsName:        mongodb.MongosNodeName(),
		labels:         mongodb.MongosLabels(),
		selectors:      mongodb.MongosSelectors(),
		args:           args,
		cmd:            cmds,
		envList:        envList,
		initContainers: initContainers,
		gvrSvcName:     mongodb.GvrSvcName(mongodb.MongosNodeName()),
		podTemplate:    &mongodb.Spec.ShardTopology.Mongos.PodTemplate,
		configSource:   mongodb.Spec.ShardTopology.Mongos.ConfigSource,
		pvcSpec:        mongodb.Spec.Storage,
		replicas:       &mongodb.Spec.ShardTopology.Mongos.Replicas,
		volumes:        volumes,
		volumeMount:    volumeMounts,
		isMongos:       true,
	}

	return c.ensureStatefulSet(mongodb, opts)
}

func mongosInitContainer(
	mongodb *api.MongoDB,
	mongodbVersion *catalog.MongoDBVersion,
	podTemplate ofst.PodTemplateSpec,
	envList []core.EnvVar,
	scriptName string,
) (core.Container, []core.Volume) {

	envList = core_util.UpsertEnvVars(envList, podTemplate.Spec.Env...)

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
		Command:         []string{"/bin/sh"},
		Args: []string{
			"-c",
			fmt.Sprintf("%v/%v", InitScriptDirectoryPath, scriptName),
		},
		Env: core_util.UpsertEnvVars([]core.EnvVar{
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
		}, envList...),
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
				Name:      InitScriptDirectoryName,
				MountPath: InitScriptDirectoryPath,
			},
			{
				Name:      certDirectoryName,
				MountPath: api.MongoCertDirectory,
			},
		},
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

	if mongodb.Spec.Init != nil && mongodb.Spec.Init.Script != nil {
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
