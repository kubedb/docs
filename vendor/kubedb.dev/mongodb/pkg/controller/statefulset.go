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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/fatih/structs"
	"github.com/pkg/errors"
	"gomodules.xyz/envsubst"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	workDirectoryName = "workdir"
	workDirectoryPath = "/work-dir"

	certDirectoryName = "certdir"

	dataDirectoryName = "datadir"
	dataDirectoryPath = "/data/db"

	configDirectoryName = "config"
	configDirectoryPath = "/data/configdb"

	InitScriptDirectoryName = "init-scripts"
	InitScriptDirectoryPath = "/init-scripts"

	ClientCertDirectoryName = "client-cert"
	ClientCertDirectoryPath = "/client-cert"

	ServerCertDirectoryName = "server-cert"
	ServerCertDirectoryPath = "/server-cert"

	initialConfigDirectoryPath = "/configdb-readonly"

	initialKeyDirectoryName = "keydir"
	initialKeyDirectoryPath = "/keydir-readonly"
)

type workloadOptions struct {
	// App level options
	stsName   string
	labels    map[string]string
	selectors map[string]string

	// db container options
	cmd          []string      // cmd of `mongodb` container
	args         []string      // args of `mongodb` container
	envList      []core.EnvVar // envList of `mongodb` container
	volumeMount  []core.VolumeMount
	configSource *core.VolumeSource

	// pod Template level options
	replicas       *int32
	gvrSvcName     string
	podTemplate    *ofst.PodTemplateSpec
	pvcSpec        *core.PersistentVolumeClaimSpec
	initContainers []core.Container
	volumes        []core.Volume // volumes to mount on stsPodTemplate
	isMongos       bool
}

func (c *Controller) ensureMongoDBNode(mongodb *api.MongoDB) (kutil.VerbType, error) {
	// Standalone, replicaset, shard
	if mongodb.Spec.ShardTopology != nil {
		return c.ensureTopologyCluster(mongodb)
	}

	return c.ensureNonTopology(mongodb)
}

func (c *Controller) ensureTopologyCluster(mongodb *api.MongoDB) (kutil.VerbType, error) {
	st, vt1, err := c.ensureConfigNode(mongodb)
	if err != nil {
		return vt1, err
	}

	sts, vt2, err := c.ensureShardNode(mongodb)
	if err != nil {
		return vt2, err
	}

	// before running mongos, wait for config servers and shard servers to come up
	sts = append(sts, st)
	if vt1 != kutil.VerbUnchanged || vt2 != kutil.VerbUnchanged {
		for _, st := range sts {
			if err := c.checkStatefulSetPodStatus(st); err != nil {
				return kutil.VerbUnchanged, err
			}
			c.recorder.Eventf(
				mongodb,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %v StatefulSet %v/%v",
				vt2, mongodb.Namespace, st.Name,
			)
		}
	}

	mongosSts, vt3, err := c.ensureMongosNode(mongodb)
	if err != nil {
		return vt3, err
	}

	if vt3 != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(mongosSts); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			mongodb,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt3, mongodb.Namespace, mongosSts.Name,
		)
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated && vt3 == kutil.VerbCreated {
		return kutil.VerbCreated, nil
	} else if vt1 != kutil.VerbUnchanged || vt2 != kutil.VerbUnchanged || vt3 != kutil.VerbUnchanged {
		return kutil.VerbPatched, nil
	}

	return kutil.VerbUnchanged, nil
}

func (c *Controller) ensureShardNode(mongodb *api.MongoDB) ([]*apps.StatefulSet, kutil.VerbType, error) {
	shardSts := func(nodeNum int32) (*apps.StatefulSet, kutil.VerbType, error) {
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

		args := []string{
			"--dbpath=" + dataDirectoryPath,
			"--auth",
			"--bind_ip=0.0.0.0",
			"--port=" + strconv.Itoa(MongoDBPort),
			"--shardsvr",
			"--replSet=" + mongodb.ShardRepSetName(nodeNum),
			"--clusterAuthMode=" + string(clusterAuth),
			"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
		}

		sslArgs, err := c.getTLSArgs(mongodb, mongodbVersion)
		if err != nil {
			return &apps.StatefulSet{}, "", err
		}
		args = append(args, sslArgs...)

		initContnr, initvolumes := installInitContainer(
			mongodb,
			mongodbVersion,
			&mongodb.Spec.ShardTopology.Shard.PodTemplate,
			mongodb.ShardNodeName(nodeNum),
		)

		var initContainers []core.Container
		var volumes []core.Volume
		var volumeMounts []core.VolumeMount
		var opts workloadOptions
		var envs []core.EnvVar
		cmds := []string{"mongod"}

		initContainers = append(initContainers, initContnr)
		volumes = core_util.UpsertVolume(volumes, initvolumes...)

		if mongodb.Spec.StorageEngine == api.StorageEngineInMemory {
			podTemplate := &mongodb.Spec.ShardTopology.Shard.PodTemplate

			args = append(args, []string{
				"--storageEngine=inMemory",
			}...)

			envs = core_util.UpsertEnvVars([]core.EnvVar{
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
					Value: mongodb.ShardRepSetName(nodeNum),
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
			}, podTemplate.Spec.Env...)

			peerFinderLocation := fmt.Sprintf("%v/peer-finder", InitScriptDirectoryPath)
			inMemoryShardScriptName := fmt.Sprintf("%v/sharding-inmemory.sh", InitScriptDirectoryPath)
			podTemplate.Spec.Lifecycle = &core.Lifecycle{
				PostStart: &core.Handler{
					Exec: &core.ExecAction{
						Command: []string{
							"/bin/bash",
							"-c",
							peerFinderLocation + " -on-start=" + inMemoryShardScriptName + " -service=" + mongodb.GvrSvcName(mongodb.ShardNodeName(nodeNum)),
						},
					},
				},
			}

			volumeMounts = core_util.UpsertVolumeMount(
				volumeMounts,
				core.VolumeMount{
					Name:      workDirectoryName,
					MountPath: workDirectoryPath,
				})
		} else {
			bootstrpContnr, bootstrpVol := c.topologyInitContainer(
				mongodb,
				mongodbVersion,
				&mongodb.Spec.ShardTopology.Shard.PodTemplate,
				mongodb.ShardRepSetName(nodeNum),
				mongodb.GvrSvcName(mongodb.ShardNodeName(nodeNum)),
				"sharding.sh",
			)
			initContainers = append(initContainers, bootstrpContnr)
			volumes = core_util.UpsertVolume(volumes, bootstrpVol...)
		}

		podTemplate := mongodb.Spec.ShardTopology.Shard.PodTemplate.DeepCopy()
		podTemplate, err = parseAffinityTemplate(podTemplate, nodeNum)
		if err != nil {
			return nil, kutil.VerbUnchanged, errors.Wrap(err, "error while templating affinity for shard nodes")
		}

		opts = workloadOptions{
			stsName:        mongodb.ShardNodeName(nodeNum),
			labels:         mongodb.ShardLabels(nodeNum),
			selectors:      mongodb.ShardSelectors(nodeNum),
			args:           args,
			cmd:            cmds,
			envList:        envs,
			initContainers: initContainers,
			gvrSvcName:     mongodb.GvrSvcName(mongodb.ShardNodeName(nodeNum)),
			podTemplate:    podTemplate,
			configSource:   mongodb.Spec.ShardTopology.Shard.ConfigSource,
			pvcSpec:        mongodb.Spec.ShardTopology.Shard.Storage,
			replicas:       &mongodb.Spec.ShardTopology.Shard.Replicas,
			volumes:        volumes,
			volumeMount:    volumeMounts,
		}

		return c.ensureStatefulSet(mongodb, opts)
	}

	var sts []*apps.StatefulSet
	vt := kutil.VerbUnchanged
	for i := int32(0); i < mongodb.Spec.ShardTopology.Shard.Shards; i++ {
		st, vt1, err := shardSts(i)
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		sts = append(sts, st)
		if vt1 != kutil.VerbUnchanged {
			vt = vt1
		}
	}

	return sts, vt, nil
}

func (c *Controller) ensureConfigNode(mongodb *api.MongoDB) (*apps.StatefulSet, kutil.VerbType, error) {
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

	args := []string{
		"--dbpath=" + dataDirectoryPath,
		"--auth",
		"--bind_ip=0.0.0.0",
		"--port=" + strconv.Itoa(MongoDBPort),
		"--configsvr",
		"--replSet=" + mongodb.ConfigSvrRepSetName(),
		"--clusterAuthMode=" + string(clusterAuth),
		"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
	}

	sslArgs, err := c.getTLSArgs(mongodb, mongodbVersion)
	if err != nil {
		return &apps.StatefulSet{}, "", err
	}
	args = append(args, sslArgs...)

	initContnr, initvolumes := installInitContainer(
		mongodb,
		mongodbVersion,
		&mongodb.Spec.ShardTopology.ConfigServer.PodTemplate,
		mongodb.ConfigSvrNodeName(),
	)

	var initContainers []core.Container
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount
	cmds := []string{"mongod"}

	initContainers = append(initContainers, initContnr)
	volumes = core_util.UpsertVolume(volumes, initvolumes...)

	bootstrpContnr, bootstrpVol := c.topologyInitContainer(
		mongodb,
		mongodbVersion,
		&mongodb.Spec.ShardTopology.ConfigServer.PodTemplate,
		mongodb.ConfigSvrRepSetName(),
		mongodb.GvrSvcName(mongodb.ConfigSvrNodeName()),
		"configdb.sh",
	)
	initContainers = append(initContainers, bootstrpContnr)
	volumes = core_util.UpsertVolume(volumes, bootstrpVol...)

	opts := workloadOptions{
		stsName:        mongodb.ConfigSvrNodeName(),
		labels:         mongodb.ConfigSvrLabels(),
		selectors:      mongodb.ConfigSvrSelectors(),
		args:           args,
		cmd:            cmds,
		envList:        nil,
		initContainers: initContainers,
		gvrSvcName:     mongodb.GvrSvcName(mongodb.ConfigSvrNodeName()),
		podTemplate:    &mongodb.Spec.ShardTopology.ConfigServer.PodTemplate,
		configSource:   mongodb.Spec.ShardTopology.ConfigServer.ConfigSource,
		pvcSpec:        mongodb.Spec.ShardTopology.ConfigServer.Storage,
		replicas:       &mongodb.Spec.ShardTopology.ConfigServer.Replicas,
		volumes:        volumes,
		volumeMount:    volumeMounts,
	}

	return c.ensureStatefulSet(mongodb, opts)
}

func (c *Controller) ensureNonTopology(mongodb *api.MongoDB) (kutil.VerbType, error) {
	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// mongodb.Spec.SSLMode & mongodb.Spec.ClusterAuthMode can be empty if upgraded operator from
	// previous version. But, eventually it will be defaulted. TODO: delete in future.
	sslMode := mongodb.Spec.SSLMode
	if sslMode == "" {
		sslMode = api.SSLModeDisabled
	}

	envList := []core.EnvVar{{Name: "SSL_MODE", Value: string(sslMode)}}

	clusterAuth := mongodb.Spec.ClusterAuthMode
	if clusterAuth == "" {
		clusterAuth = api.ClusterAuthModeKeyFile
		if sslMode != api.SSLModeDisabled {
			clusterAuth = api.ClusterAuthModeX509
		}
	}

	args := []string{
		"--dbpath=" + dataDirectoryPath,
		"--auth",
		"--bind_ip=0.0.0.0",
		"--port=" + strconv.Itoa(MongoDBPort),
	}

	sslArgs, err := c.getTLSArgs(mongodb, mongodbVersion)
	if err != nil {
		return "", err
	}
	args = append(args, sslArgs...)

	initContnr, initvolumes := installInitContainer(
		mongodb,
		mongodbVersion,
		mongodb.Spec.PodTemplate,
		"")

	var initContainers []core.Container
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount
	var cmds []string

	initContainers = append(initContainers, initContnr)
	volumes = core_util.UpsertVolume(volumes, initvolumes...)

	if mongodb.Spec.Init != nil && mongodb.Spec.Init.ScriptSource != nil {
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: mongodb.Spec.Init.ScriptSource.VolumeSource,
		})

		volumeMounts = []core.VolumeMount{
			{
				Name:      "initial-script",
				MountPath: "/docker-entrypoint-initdb.d",
			},
		}
	}

	if mongodb.Spec.ReplicaSet != nil {
		cmds = []string{"mongod"}
		args = meta_util.UpsertArgumentList(args, []string{
			"--replSet=" + mongodb.RepSetName(),
			"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
			"--clusterAuthMode=" + string(clusterAuth),
		})

		if mongodb.Spec.StorageEngine == api.StorageEngineInMemory {
			podTemplate := mongodb.Spec.PodTemplate

			args = append(args, []string{
				"--storageEngine=inMemory",
			}...)

			envList = core_util.UpsertEnvVars([]core.EnvVar{
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
					Value: mongodb.RepSetName(),
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
			}, podTemplate.Spec.Env...)

			peerFinderLocation := fmt.Sprintf("%v/peer-finder", InitScriptDirectoryPath)
			inMemoryReplicaSetScriptName := fmt.Sprintf("%v/replicaset-inmemory.sh", InitScriptDirectoryPath)

			podTemplate.Spec.Lifecycle = &core.Lifecycle{
				PostStart: &core.Handler{
					Exec: &core.ExecAction{
						Command: []string{
							"/bin/bash",
							"-c",
							peerFinderLocation + " -on-start=" + inMemoryReplicaSetScriptName + " -service=" + mongodb.GvrSvcName(mongodb.OffshootName()),
						},
					},
				},
			}

			volumeMounts = core_util.UpsertVolumeMount(
				volumeMounts,
				core.VolumeMount{
					Name:      workDirectoryName,
					MountPath: workDirectoryPath,
				})

		} else {
			bootstrpContnr, bootstrpVol := c.topologyInitContainer(
				mongodb,
				mongodbVersion,
				mongodb.Spec.PodTemplate,
				mongodb.RepSetName(),
				mongodb.GvrSvcName(mongodb.OffshootName()),
				"replicaset.sh",
			)
			initContainers = append(initContainers, bootstrpContnr)
			volumes = core_util.UpsertVolume(volumes, bootstrpVol...)
		}
	}

	opts := workloadOptions{
		stsName:        mongodb.OffshootName(),
		labels:         mongodb.OffshootLabels(),
		selectors:      mongodb.OffshootSelectors(),
		args:           args,
		cmd:            cmds,
		envList:        envList,
		initContainers: initContainers,
		gvrSvcName:     mongodb.GvrSvcName(mongodb.OffshootName()),
		podTemplate:    mongodb.Spec.PodTemplate,
		configSource:   mongodb.Spec.ConfigSource,
		pvcSpec:        mongodb.Spec.Storage,
		replicas:       mongodb.Spec.Replicas,
		volumes:        volumes,
		volumeMount:    volumeMounts,
	}

	st, vt, err := c.ensureStatefulSet(mongodb, opts)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(st); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			mongodb,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt, mongodb.Namespace, opts.stsName,
		)
	}
	return vt, err
}

func (c *Controller) ensureStatefulSet(mongodb *api.MongoDB, opts workloadOptions) (*apps.StatefulSet, kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkStatefulSet(mongodb, opts.stsName); err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	// Create statefulSet for MongoDB database
	statefulSetMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: mongodb.Namespace,
	}

	owner := metav1.NewControllerRef(mongodb, api.SchemeGroupVersion.WithKind(api.ResourceKindMongoDB))

	readinessProbe := pt.Spec.ReadinessProbe
	if readinessProbe != nil && structs.IsZero(*readinessProbe) {
		readinessProbe = nil
	}
	livenessProbe := pt.Spec.LivenessProbe
	if livenessProbe != nil && structs.IsZero(*livenessProbe) {
		livenessProbe = nil
	}

	if mongodb.Spec.SSLMode != api.SSLModeDisabled && mongodb.Spec.TLS != nil {
		opts.volumeMount = append(opts.volumeMount, core.VolumeMount{
			Name:      certDirectoryName,
			MountPath: api.MongoCertDirectory,
		})
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = opts.labels
			in.Annotations = pt.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = opts.replicas
			in.Spec.ServiceName = opts.gvrSvcName
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: opts.selectors,
			}
			in.Spec.Template.Labels = opts.selectors
			in.Spec.Template.Annotations = pt.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				pt.Spec.InitContainers,
			)
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				in.Spec.Template.Spec.Containers,
				core.Container{
					Name:            api.MongoDBContainerName,
					Image:           mongodbVersion.Spec.DB.Image,
					ImagePullPolicy: core.PullIfNotPresent,
					Command:         opts.cmd,
					Args: meta_util.UpsertArgumentList(
						opts.args, pt.Spec.Args),
					Ports: []core.ContainerPort{
						{
							Name:          "db",
							ContainerPort: MongoDBPort,
							Protocol:      core.ProtocolTCP,
						},
					},
					Env:            core_util.UpsertEnvVars(opts.envList, pt.Spec.Env...),
					Resources:      pt.Spec.Resources,
					Lifecycle:      pt.Spec.Lifecycle,
					LivenessProbe:  livenessProbe,
					ReadinessProbe: readinessProbe,
					VolumeMounts:   opts.volumeMount,
				})

			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				opts.initContainers,
			)

			if mongodb.GetMonitoringVendor() == mona.VendorPrometheus {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
					in.Spec.Template.Spec.Containers,
					getExporterContainer(mongodb, mongodbVersion),
				)
			}

			in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volumes...)

			in.Spec.Template = upsertEnv(in.Spec.Template, mongodb)
			if !opts.isMongos {
				//Mongos doesn't have any data
				in = upsertDataVolume(in, opts.pvcSpec, mongodb.Spec.StorageType)
			}

			if opts.configSource != nil {
				in.Spec.Template = c.upsertConfigSourceVolume(in.Spec.Template, opts.configSource)
			}

			in.Spec.Template.Spec.NodeSelector = pt.Spec.NodeSelector
			in.Spec.Template.Spec.Affinity = pt.Spec.Affinity
			if pt.Spec.SchedulerName != "" {
				in.Spec.Template.Spec.SchedulerName = pt.Spec.SchedulerName
			}
			in.Spec.Template.Spec.Tolerations = pt.Spec.Tolerations
			in.Spec.Template.Spec.ImagePullSecrets = pt.Spec.ImagePullSecrets
			in.Spec.Template.Spec.PriorityClassName = pt.Spec.PriorityClassName
			in.Spec.Template.Spec.Priority = pt.Spec.Priority
			in.Spec.Template.Spec.SecurityContext = pt.Spec.SecurityContext
			in.Spec.Template.Spec.ServiceAccountName = pt.Spec.ServiceAccountName
			in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
				Type: apps.OnDeleteStatefulSetStrategyType,
			}
			return in
		},
		metav1.PatchOptions{},
	)

	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	// ensure pdb
	if err := c.CreateStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		return nil, vt, err
	}

	return statefulSet, vt, nil
}

func (c *Controller) checkStatefulSet(mongodb *api.MongoDB, stsName string) error {
	// StatefulSet for MongoDB database
	statefulSet, err := c.Client.AppsV1().StatefulSets(mongodb.Namespace).Get(context.TODO(), stsName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB ||
		statefulSet.Labels[api.LabelDatabaseName] != mongodb.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, mongodb.Namespace, stsName)
	}

	return nil
}

// Init container for both ReplicaSet and Standalone instances
func installInitContainer(
	mongodb *api.MongoDB,
	mongodbVersion *v1alpha1.MongoDBVersion,
	podTemplate *ofst.PodTemplateSpec,
	stsName string,
) (core.Container, []core.Volume) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	var installContainer core.Container

	if podTemplate != nil {
		pt = *podTemplate
	}

	envList := make([]core.EnvVar, 0)

	if mongodb.Spec.SSLMode == api.SSLModeDisabled || mongodb.Spec.TLS == nil {
		envList = append(envList, core.EnvVar{
			Name:  "SSL_MODE",
			Value: string(api.SSLModeDisabled),
		})
	}

	installContainer = core.Container{
		Name:            api.MongoDBInitInstallContainerName,
		Image:           mongodbVersion.Spec.InitContainer.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"/bin/sh"},
		Env:             envList,
		Args: []string{
			"-c", `
			echo "running install.sh"
			/scripts/install.sh`,
		},
		VolumeMounts: []core.VolumeMount{
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
		Resources: pt.Spec.Resources,
	}

	initVolumes := []core.Volume{
		{
			Name: workDirectoryName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
		{
			Name: InitScriptDirectoryName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
		{
			Name: certDirectoryName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
	}

	if mongodb.Spec.TLS != nil {
		installContainer.VolumeMounts = core_util.UpsertVolumeMount(
			installContainer.VolumeMounts,
			[]core.VolumeMount{
				{
					Name:      ClientCertDirectoryName,
					MountPath: ClientCertDirectoryPath,
				},
				{
					Name:      ServerCertDirectoryName,
					MountPath: ServerCertDirectoryPath,
				},
			}...)

		initVolumes = append(initVolumes, []core.Volume{
			{
				Name: ClientCertDirectoryName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						DefaultMode: types.Int32P(0400),
						SecretName:  mongodb.MustCertSecretName(api.MongoDBClientCert, ""),
					},
				},
			},
			{
				Name: ServerCertDirectoryName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						DefaultMode: types.Int32P(0400),
						SecretName:  mongodb.MustCertSecretName(api.MongoDBServerCert, stsName),
					},
				},
			},
		}...)
	}

	// mongodb.Spec.SSLMode can be empty if upgraded operator from previous version.
	// But, eventually it will be defaulted. TODO: delete `mongodb.Spec.SSLMode != ""` in future.
	//sslMode := mongodb.Spec.SSLMode
	//if sslMode == "" {
	//	sslMode = api.SSLModeDisabled
	//}
	if mongodb.Spec.KeyFile != nil {
		installContainer.VolumeMounts = core_util.UpsertVolumeMount(
			installContainer.VolumeMounts,
			core.VolumeMount{
				Name:      initialKeyDirectoryName,
				MountPath: initialKeyDirectoryPath,
			})

		initVolumes = append(initVolumes, core.Volume{
			Name: initialKeyDirectoryName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					DefaultMode: types.Int32P(0400),
					SecretName:  mongodb.Spec.KeyFile.SecretName,
				},
			},
		})
	}

	return installContainer, initVolumes
}

func upsertDataVolume(
	statefulSet *apps.StatefulSet,
	pvcSpec *core.PersistentVolumeClaimSpec,
	storageType api.StorageType,
) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.MongoDBContainerName {
			volumeMount := []core.VolumeMount{
				{
					Name:      dataDirectoryName,
					MountPath: dataDirectoryPath,
				},
				// Mount volume for config source
				{
					Name:      configDirectoryName,
					MountPath: configDirectoryPath,
				},
				{
					Name:      InitScriptDirectoryName,
					MountPath: InitScriptDirectoryPath,
				},
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount...)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			// Volume for config source
			volumes := core.Volume{
				Name: configDirectoryName,
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			}
			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
				statefulSet.Spec.Template.Spec.Volumes,
				volumes,
			)

			if storageType == api.StorageTypeEphemeral {
				ed := core.EmptyDirVolumeSource{}
				if pvcSpec != nil {
					if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
						ed.SizeLimit = &sz
					}
				}
				statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
					statefulSet.Spec.Template.Spec.Volumes,
					core.Volume{
						Name: dataDirectoryName,
						VolumeSource: core.VolumeSource{
							EmptyDir: &ed,
						},
					})
			} else {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in mongodb.Spec.Storage`, core.ReadWriteOnce)
				}

				claim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: dataDirectoryName,
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					claim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				statefulSet.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(
					statefulSet.Spec.VolumeClaimTemplates,
					claim,
				)
			}
			break
		}
	}
	return statefulSet
}

func upsertEnv(template core.PodTemplateSpec, mongodb *api.MongoDB) core.PodTemplateSpec {
	envList := []core.EnvVar{
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
	}
	for i, container := range template.Spec.Containers {
		if container.Name == api.MongoDBContainerName || container.Name == api.ContainerExporterName {
			template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
		}
	}
	return template
}

func (c *Controller) checkStatefulSetPodStatus(sts *apps.StatefulSet) error {
	log.Infof("Waiting for running phase for statefulset %v/%v.", sts.Namespace, sts.Name)
	err := core_util.WaitUntilPodRunningBySelector(
		context.TODO(),
		c.Client,
		sts.Namespace,
		sts.Spec.Selector,
		int(types.Int32(sts.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}

func getExporterContainer(mongodb *api.MongoDB, mongodbVersion *v1alpha1.MongoDBVersion) core.Container {
	metricsPath := fmt.Sprintf("--web.metrics-path=%v", mongodb.StatsService().Path())
	// change metric path for percona-mongodb-exporter
	if strings.Contains(mongodbVersion.Spec.Exporter.Image, "percona") {
		metricsPath = fmt.Sprintf("--web.telemetry-path=%v", mongodb.StatsService().Path())
	}

	args := append([]string{
		"--mongodb.uri=mongodb://$(MONGO_INITDB_ROOT_USERNAME):$(MONGO_INITDB_ROOT_PASSWORD)@localhost:27017/admin",
		fmt.Sprintf("--web.listen-address=:%d", mongodb.Spec.Monitor.Prometheus.Exporter.Port),
		metricsPath,
	}, mongodb.Spec.Monitor.Args...)

	if mongodb.Spec.SSLMode != api.SSLModeDisabled && mongodb.Spec.TLS != nil {
		clientPEM := fmt.Sprintf("%s/%s", api.MongoCertDirectory, api.MongoClientFileName)
		clientCA := fmt.Sprintf("%s/%s", api.MongoCertDirectory, api.TLSCACertFileName)
		args = append(args, "--mongodb.tls")
		args = append(args, "--mongodb.tls-ca")
		args = append(args, clientCA)
		args = append(args, "--mongodb.tls-cert")
		args = append(args, clientPEM)
	}

	return core.Container{
		Name:  api.ContainerExporterName,
		Args:  args,
		Image: mongodbVersion.Spec.Exporter.Image,
		Ports: []core.ContainerPort{
			{
				Name:          api.PrometheusExporterPortName,
				Protocol:      core.ProtocolTCP,
				ContainerPort: mongodb.Spec.Monitor.Prometheus.Exporter.Port,
			},
		},
		Env:             mongodb.Spec.Monitor.Prometheus.Exporter.Env,
		Resources:       mongodb.Spec.Monitor.Prometheus.Exporter.Resources,
		SecurityContext: mongodb.Spec.Monitor.Prometheus.Exporter.SecurityContext,
		VolumeMounts: []core.VolumeMount{
			{
				Name:      certDirectoryName,
				MountPath: api.MongoCertDirectory, //TODO: use exporter certs by adding a exporter volume and mounting that here
			},
		},
	}
}

func parseAffinityTemplate(podTemplate *ofst.PodTemplateSpec, nodeNum int32) (*ofst.PodTemplateSpec, error) {
	if podTemplate == nil || podTemplate.Spec.Affinity == nil {
		return podTemplate, nil
	}

	templateMap := map[string]string{
		api.MongoDBShardAffinityTemplateVar: strconv.Itoa(int(nodeNum)),
	}

	jsonObj, err := json.Marshal(podTemplate)
	if err != nil {
		return podTemplate, err
	}

	resolved, err := envsubst.EvalMap(string(jsonObj), templateMap)
	if err != nil {
		return podTemplate, err
	}

	err = json.Unmarshal([]byte(resolved), podTemplate)
	return podTemplate, err
}
