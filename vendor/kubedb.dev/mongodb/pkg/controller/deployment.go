package controller

import (
	"fmt"
	"strconv"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/types"
	"github.com/fatih/structs"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (c *Controller) checkDeployment(mongodb *api.MongoDB, deployName string) error {
	// Deployment for Mongos
	deployment, err := c.Client.AppsV1().Deployments(mongodb.Namespace).Get(deployName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if deployment.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB ||
		deployment.Labels[api.LabelDatabaseName] != mongodb.Name {
		return fmt.Errorf(`intended deployment "%v/%v" already exists`, mongodb.Namespace, deployName)
	}
	return nil
}

func (c *Controller) ensureDeployment(
	mongodb *api.MongoDB,
	strategy apps.DeploymentStrategy,
	opts workloadOptions,
) (kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkDeployment(mongodb, opts.stsName); err != nil {
		return kutil.VerbUnchanged, err
	}
	deploymentMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: mongodb.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	readinessProbe := pt.Spec.ReadinessProbe
	if readinessProbe != nil && structs.IsZero(*readinessProbe) {
		readinessProbe = nil
	}
	livenessProbe := pt.Spec.LivenessProbe
	if livenessProbe != nil && structs.IsZero(*livenessProbe) {
		livenessProbe = nil
	}

	deployment, vt, err := app_util.CreateOrPatchDeployment(c.Client, deploymentMeta, func(in *apps.Deployment) *apps.Deployment {
		in.Labels = opts.labels
		in.Annotations = pt.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)

		in.Spec.Replicas = opts.replicas
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: opts.selectors,
		}
		in.Spec.Template.Labels = opts.selectors
		in.Spec.Template.Annotations = pt.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
			in.Spec.Template.Spec.InitContainers, pt.Spec.InitContainers,
		)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
			in.Spec.Template.Spec.Containers,
			core.Container{
				Name:            api.ResourceSingularMongoDB,
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

		in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
			in.Spec.Template.Spec.Volumes,
			opts.volume...,
		)

		in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, core.Volume{
			Name: configDirectoryName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		})
		in.Spec.Template = upsertEnv(in.Spec.Template, mongodb)

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

		if c.EnableRBAC {
			in.Spec.Template.Spec.ServiceAccountName = pt.Spec.ServiceAccountName
		}

		in.Spec.Strategy = strategy

		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if err := c.CreateDeploymentPodDisruptionBudget(deployment); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := app_util.WaitUntilDeploymentReady(c.Client, deployment.ObjectMeta); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			mongodb,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v Deployment %v/%v",
			vt, mongodb.Namespace, opts.stsName,
		)
	}
	return vt, nil
}

func (c *Controller) ensureMongosNode(mongodb *api.MongoDB) (kutil.VerbType, error) {
	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
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
		"--sslMode=" + string(sslMode),
		"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
	}

	if sslMode != api.SSLModeDisabled {
		args = append(args, []string{
			fmt.Sprintf("--sslCAFile=/data/configdb/%v", api.MongoTLSCertFileName),
			fmt.Sprintf("--sslPEMKeyFile=/data/configdb/%v", api.MongoServerPemFileName),
		}...)
	}

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
	)

	var initContainers []core.Container
	var volumes []core.Volume

	volumeMounts := []core.VolumeMount{
		{
			Name:      configDirectoryName,
			MountPath: configDirectoryPath,
		},
	}

	initContainers = append(initContainers, initContnr)
	volumes = core_util.UpsertVolume(volumes, initvolumes...)

	if mongodb.Spec.Init != nil && mongodb.Spec.Init.ScriptSource != nil {
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: mongodb.Spec.Init.ScriptSource.VolumeSource,
		})

		volumeMounts = append(
			volumeMounts,
			[]core.VolumeMount{
				{
					Name:      "initial-script",
					MountPath: "/docker-entrypoint-initdb.d",
				},
			}...,
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
		gvrSvcName:     mongodb.GvrSvcName(mongodb.OffshootName()),
		podTemplate:    &mongodb.Spec.ShardTopology.Mongos.PodTemplate,
		configSource:   mongodb.Spec.ShardTopology.Mongos.ConfigSource,
		pvcSpec:        mongodb.Spec.Storage,
		replicas:       &mongodb.Spec.ShardTopology.Mongos.Replicas,
		volume:         volumes,
		volumeMount:    volumeMounts,
	}

	return c.ensureDeployment(
		mongodb,
		mongodb.Spec.ShardTopology.Mongos.Strategy,
		opts,
	)
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
		Name:            InitBootstrapContainerName,
		Image:           mongodbVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"/bin/sh"},
		Args: []string{
			"-c",
			fmt.Sprintf("/usr/local/bin/%v", scriptName),
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
						Key: KeyMongoDBUser,
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
						Key: KeyMongoDBPassword,
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
		},
	}

	rsVolume := []core.Volume{
		{
			Name: initialKeyDirectoryName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					DefaultMode: types.Int32P(256),
					SecretName:  mongodb.Spec.CertificateSecret.SecretName,
				},
			},
		},
	}

	if mongodb.Spec.Init != nil && mongodb.Spec.Init.ScriptSource != nil {
		rsVolume = append(rsVolume, core.Volume{
			Name:         "initial-script",
			VolumeSource: mongodb.Spec.Init.ScriptSource.VolumeSource,
		})

		bootstrapContainer.VolumeMounts = core_util.UpsertVolumeMount(
			bootstrapContainer.VolumeMounts,
			core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/docker-entrypoint-initdb.d",
			},
		)
	}

	return bootstrapContainer, rsVolume
}
