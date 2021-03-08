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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (c *Controller) ensureMongosNode(db *api.MongoDB) (*apps.StatefulSet, kutil.VerbType, error) {
	mongodbVersion, err := c.DBClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	// mongodb.Spec.SSLMode & mongodb.Spec.ClusterAuthMode can be empty if upgraded operator from
	// previous version. But, eventually it will be defaulted. TODO: delete in future.
	sslMode := db.Spec.SSLMode
	if sslMode == "" {
		sslMode = api.SSLModeDisabled
	}
	clusterAuth := db.Spec.ClusterAuthMode
	if clusterAuth == "" {
		clusterAuth = api.ClusterAuthModeKeyFile
		if sslMode != api.SSLModeDisabled {
			clusterAuth = api.ClusterAuthModeX509
		}
	}

	cmds := []string{"mongos"}
	args := []string{
		"--ipv6",
		"--bind_ip_all",
		"--port=" + strconv.Itoa(api.MongoDBDatabasePort),
		"--configdb=$(CONFIGDB_REPSET)",
		"--clusterAuthMode=" + string(clusterAuth),
		"--keyFile=" + api.MongoDBConfigDirectoryPath + "/" + api.MongoDBKeyForKeyFile,
	}

	sslArgs, err := c.getTLSArgs(db, mongodbVersion)
	if err != nil {
		return &apps.StatefulSet{}, "", err
	}
	args = append(args, sslArgs...)

	// shardDsn List, separated by space ' '
	var shardDsn string
	for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
		if i != 0 {
			shardDsn += " "
		}
		shardDsn += db.ShardDSN(i)
	}

	envList := []core.EnvVar{
		{
			Name:  "CONFIGDB_REPSET",
			Value: db.ConfigSvrDSN(),
		},
		{
			Name:  "SHARD_REPSETS",
			Value: shardDsn,
		},
		{
			Name:  "SERVICE_NAME",
			Value: db.ServiceName(),
		},
	}

	initContnr, initvolumes := installInitContainer(
		db,
		mongodbVersion,
		&db.Spec.ShardTopology.Mongos.PodTemplate,
		db.MongosNodeName(),
	)

	var initContainers []core.Container
	var volumes []core.Volume

	volumes = append(volumes, core.Volume{
		Name: api.MongoDBConfigDirectoryName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	})

	volumeMounts := []core.VolumeMount{
		{
			Name:      api.MongoDBConfigDirectoryName,
			MountPath: api.MongoDBConfigDirectoryPath,
		},
	}

	initContainers = append(initContainers, initContnr)
	volumes = core_util.UpsertVolume(volumes, initvolumes...)

	if db.Spec.Init != nil && db.Spec.Init.Script != nil {
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: db.Spec.Init.Script.VolumeSource,
		})

		volumeMounts = core_util.UpsertVolumeMount(
			volumeMounts,
			core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/docker-entrypoint-initdb.d",
			},
		)
	}

	bootstrpContnr, bootstrpVol := mongosInitContainer(
		db,
		mongodbVersion,
		db.Spec.ShardTopology.Mongos.PodTemplate,
		envList,
		"mongos.sh",
	)
	initContainers = append(initContainers, bootstrpContnr)
	volumes = core_util.UpsertVolume(volumes, bootstrpVol...)

	opts := workloadOptions{
		stsName:        db.MongosNodeName(),
		labels:         db.MongosLabels(),
		selectors:      db.MongosSelectors(),
		args:           args,
		cmd:            cmds,
		envList:        envList,
		initContainers: initContainers,
		gvrSvcName:     db.GoverningServiceName(db.MongosNodeName()),
		podTemplate:    &db.Spec.ShardTopology.Mongos.PodTemplate,
		configSecret:   db.Spec.ShardTopology.Mongos.ConfigSecret,
		pvcSpec:        db.Spec.Storage,
		replicas:       &db.Spec.ShardTopology.Mongos.Replicas,
		volumes:        volumes,
		volumeMount:    volumeMounts,
		isMongos:       true,
	}

	return c.ensureStatefulSet(db, opts)
}

func mongosInitContainer(
	db *api.MongoDB,
	mongodbVersion *catalog.MongoDBVersion,
	podTemplate ofst.PodTemplateSpec,
	envList []core.EnvVar,
	scriptName string,
) (core.Container, []core.Volume) {

	envList = core_util.UpsertEnvVars(envList, podTemplate.Spec.Env...)

	// mongodb.Spec.SSLMode & mongodb.Spec.ClusterAuthMode can be empty if upgraded operator from
	// previous version. But, eventually it will be defaulted. TODO: delete in future.
	sslMode := db.Spec.SSLMode
	if sslMode == "" {
		sslMode = api.SSLModeDisabled
	}
	clusterAuth := db.Spec.ClusterAuthMode
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
			fmt.Sprintf("%v/%v", api.MongoDBInitScriptDirectoryPath, scriptName),
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
							Name: db.Spec.AuthSecret.Name,
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
							Name: db.Spec.AuthSecret.Name,
						},
						Key: core.BasicAuthPasswordKey,
					},
				},
			},
		}, envList...),
		VolumeMounts: []core.VolumeMount{
			{
				Name:      api.MongoDBWorkDirectoryName,
				MountPath: api.MongoDBWorkDirectoryPath,
			},
			{
				Name:      api.MongoDBConfigDirectoryName,
				MountPath: api.MongoDBConfigDirectoryPath,
			},
			{
				Name:      api.MongoDBInitScriptDirectoryName,
				MountPath: api.MongoDBInitScriptDirectoryPath,
			},
			{
				Name:      api.MongoDBCertDirectoryName,
				MountPath: api.MongoCertDirectory,
			},
		},
	}

	var rsVolumes []core.Volume

	if db.Spec.KeyFileSecret != nil {
		rsVolumes = core_util.UpsertVolume(rsVolumes, core.Volume{
			Name: api.MongoDBInitialKeyDirectoryName, // FIXIT: mounted where?
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					DefaultMode: pointer.Int32P(0400),
					SecretName:  db.Spec.KeyFileSecret.Name,
				},
			},
		})
	}

	if db.Spec.Init != nil && db.Spec.Init.Script != nil {
		rsVolumes = core_util.UpsertVolume(rsVolumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: db.Spec.Init.Script.VolumeSource,
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
