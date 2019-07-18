package controller

import (
	"fmt"
	"strconv"

	"github.com/appscode/go/log"
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
	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"
)

const (
	workDirectoryName = "workdir"
	workDirectoryPath = "/work-dir"

	dataDirectoryName = "datadir"
	dataDirectoryPath = "/data/db"

	configDirectoryName = "config"
	configDirectoryPath = "/data/configdb"

	initialConfigDirectoryName = "configdir"
	initialConfigDirectoryPath = "/configdb-readonly"

	initialKeyDirectoryName = "keydir"
	initialKeyDirectoryPath = "/keydir-readonly"

	InitInstallContainerName   = "copy-config"
	InitBootstrapContainerName = "bootstrap"
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
	volume         []core.Volume // volumes to mount on stsPodTemplate
}

func (c *Controller) ensureMongoDBNode(mongodb *api.MongoDB) (kutil.VerbType, error) {
	// Standalone, replicaset, shard
	if mongodb.Spec.ShardTopology != nil {
		return c.ensureTopologyCluster(mongodb)
	}

	return c.ensureNonTopology(mongodb)
}

func (c *Controller) ensureTopologyCluster(mongodb *api.MongoDB) (kutil.VerbType, error) {
	vt1, err := c.ensureConfigNode(mongodb)
	if err != nil {
		return vt1, err
	}

	vt2, err := c.ensureShardNode(mongodb)
	if err != nil {
		return vt2, err
	}

	vt3, err := c.ensureMongosNode(mongodb)
	if err != nil {
		return vt3, err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated && vt3 == kutil.VerbCreated {
		return kutil.VerbCreated, nil
	} else if vt1 != kutil.VerbUnchanged || vt2 != kutil.VerbUnchanged || vt3 != kutil.VerbUnchanged {
		return kutil.VerbPatched, nil
	}

	return kutil.VerbUnchanged, nil
}

func (c *Controller) ensureShardNode(mongodb *api.MongoDB) (kutil.VerbType, error) {
	shardSts := func(nodeNum int32) (kutil.VerbType, error) {
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

		args := []string{
			"--dbpath=" + dataDirectoryPath,
			"--auth",
			"--bind_ip=0.0.0.0",
			"--port=" + strconv.Itoa(MongoDBPort),
			"--shardsvr",
			"--replSet=" + mongodb.ShardRepSetName(nodeNum),
			"--clusterAuthMode=" + string(clusterAuth),
			"--sslMode=" + string(sslMode),
			"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
		}

		if sslMode != api.SSLModeDisabled {
			args = append(args, []string{
				fmt.Sprintf("--sslCAFile=/data/configdb/%v", TLSCert),
				fmt.Sprintf("--sslPEMKeyFile=/data/configdb/%v", MongoServerPem),
			}...)
		}

		initContnr, initvolumes := installInitContainer(
			mongodb,
			mongodbVersion,
			&mongodb.Spec.ShardTopology.Shard.PodTemplate,
		)

		var initContainers []core.Container
		var volumes []core.Volume
		var volumeMounts []core.VolumeMount
		cmds := []string{"mongod"}

		initContainers = append(initContainers, initContnr)
		volumes = core_util.UpsertVolume(volumes, initvolumes...)

		bootstrpContnr, bootstrpVol := topologyInitContainer(
			mongodb,
			mongodbVersion,
			&mongodb.Spec.ShardTopology.Shard.PodTemplate,
			mongodb.ShardRepSetName(nodeNum),
			mongodb.GvrSvcName(mongodb.ShardNodeName(nodeNum)),
			"sharding.sh",
		)
		initContainers = append(initContainers, bootstrpContnr)
		volumes = core_util.UpsertVolume(volumes, bootstrpVol...)

		opts := workloadOptions{
			stsName:        mongodb.ShardNodeName(nodeNum),
			labels:         mongodb.ShardLabels(nodeNum),
			selectors:      mongodb.ShardSelectors(nodeNum),
			args:           args,
			cmd:            cmds,
			envList:        nil,
			initContainers: initContainers,
			gvrSvcName:     mongodb.GvrSvcName(mongodb.ShardNodeName(nodeNum)),
			podTemplate:    &mongodb.Spec.ShardTopology.Shard.PodTemplate,
			configSource:   mongodb.Spec.ShardTopology.Shard.ConfigSource,
			pvcSpec:        mongodb.Spec.ShardTopology.Shard.Storage,
			replicas:       &mongodb.Spec.ShardTopology.Shard.Replicas,
			volume:         volumes,
			volumeMount:    volumeMounts,
		}

		return c.ensureStatefulSet(mongodb, opts)
	}

	for i := int32(0); i < mongodb.Spec.ShardTopology.Shard.Shards; i++ {
		if _, err := shardSts(i); err != nil {
			return kutil.VerbUnchanged, err
		}
	}

	return kutil.VerbUnchanged, nil
}

func (c *Controller) ensureConfigNode(mongodb *api.MongoDB) (kutil.VerbType, error) {
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

	args := []string{
		"--dbpath=" + dataDirectoryPath,
		"--auth",
		"--bind_ip=0.0.0.0",
		"--port=" + strconv.Itoa(MongoDBPort),
		"--configsvr",
		"--replSet=" + mongodb.ConfigSvrRepSetName(),
		"--clusterAuthMode=" + string(clusterAuth),
		"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
		"--sslMode=" + string(sslMode),
	}

	if sslMode != api.SSLModeDisabled {
		args = append(args, []string{
			fmt.Sprintf("--sslCAFile=/data/configdb/%v", TLSCert),
			fmt.Sprintf("--sslPEMKeyFile=/data/configdb/%v", MongoServerPem),
		}...)
	}

	initContnr, initvolumes := installInitContainer(
		mongodb,
		mongodbVersion,
		&mongodb.Spec.ShardTopology.ConfigServer.PodTemplate,
	)

	var initContainers []core.Container
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount
	cmds := []string{"mongod"}

	initContainers = append(initContainers, initContnr)
	volumes = core_util.UpsertVolume(volumes, initvolumes...)

	bootstrpContnr, bootstrpVol := topologyInitContainer(
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
		volume:         volumes,
		volumeMount:    volumeMounts,
	}

	return c.ensureStatefulSet(mongodb, opts)
}

func (c *Controller) ensureNonTopology(mongodb *api.MongoDB) (kutil.VerbType, error) {
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

	args := []string{
		"--dbpath=" + dataDirectoryPath,
		"--auth",
		"--bind_ip=0.0.0.0",
		"--port=" + strconv.Itoa(MongoDBPort),
		"--sslMode=" + string(sslMode),
	}

	if sslMode != api.SSLModeDisabled {
		args = append(args, []string{
			fmt.Sprintf("--sslCAFile=/data/configdb/%v", TLSCert),
			fmt.Sprintf("--sslPEMKeyFile=/data/configdb/%v", MongoServerPem),
		}...)
	}

	initContnr, initvolumes := installInitContainer(mongodb, mongodbVersion, mongodb.Spec.PodTemplate)

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
		bootstrpContnr, bootstrpVol := topologyInitContainer(
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

	opts := workloadOptions{
		stsName:        mongodb.OffshootName(),
		labels:         mongodb.OffshootLabels(),
		selectors:      mongodb.OffshootSelectors(),
		args:           args,
		cmd:            cmds,
		envList:        nil,
		initContainers: initContainers,
		gvrSvcName:     mongodb.GvrSvcName(mongodb.OffshootName()),
		podTemplate:    mongodb.Spec.PodTemplate,
		configSource:   mongodb.Spec.ConfigSource,
		pvcSpec:        mongodb.Spec.Storage,
		replicas:       mongodb.Spec.Replicas,
		volume:         volumes,
		volumeMount:    volumeMounts,
	}

	return c.ensureStatefulSet(mongodb, opts)
}

func (c *Controller) ensureStatefulSet(mongodb *api.MongoDB, opts workloadOptions) (kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkStatefulSet(mongodb, opts.stsName); err != nil {
		return kutil.VerbUnchanged, err
	}

	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for MongoDB database
	statefulSetMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: mongodb.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	readinessProbe := pt.Spec.ReadinessProbe
	if readinessProbe != nil && structs.IsZero(*readinessProbe) {
		readinessProbe = nil
	}
	livenessProbe := pt.Spec.LivenessProbe
	if livenessProbe != nil && structs.IsZero(*livenessProbe) {
		livenessProbe = nil
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = opts.labels
		in.Annotations = pt.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)

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
				core.Container{
					Name: "exporter",
					Args: append([]string{
						fmt.Sprintf("--web.listen-address=:%d", mongodb.Spec.Monitor.Prometheus.Port),
						fmt.Sprintf("--web.metrics-path=%v", mongodb.StatsService().Path()),
						"--mongodb.uri=mongodb://$(MONGO_INITDB_ROOT_USERNAME):$(MONGO_INITDB_ROOT_PASSWORD)@127.0.0.1:27017",
					}, mongodb.Spec.Monitor.Args...),
					Image: mongodbVersion.Spec.Exporter.Image,
					Ports: []core.ContainerPort{
						{
							Name:          api.PrometheusExporterPortName,
							Protocol:      core.ProtocolTCP,
							ContainerPort: mongodb.Spec.Monitor.Prometheus.Port,
						},
					},
					Env:             mongodb.Spec.Monitor.Env,
					Resources:       mongodb.Spec.Monitor.Resources,
					SecurityContext: mongodb.Spec.Monitor.SecurityContext,
				})
		}

		in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volume...)

		in.Spec.Template = upsertEnv(in.Spec.Template, mongodb)
		in = upsertDataVolume(in, opts.pvcSpec, mongodb.Spec.StorageType)

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

		in.Spec.UpdateStrategy = mongodb.Spec.UpdateStrategy
		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
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
	// ensure pdb
	if err := c.CreateStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		return vt, err
	}

	return vt, nil
}

func (c *Controller) checkStatefulSet(mongodb *api.MongoDB, stsName string) error {
	// StatefulSet for MongoDB database
	statefulSet, err := c.Client.AppsV1().StatefulSets(mongodb.Namespace).Get(stsName, metav1.GetOptions{})
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
) (core.Container, []core.Volume) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if podTemplate != nil {
		pt = *podTemplate
	}

	installContainer := core.Container{
		Name:            InitInstallContainerName,
		Image:           mongodbVersion.Spec.InitContainer.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"sh"},
		Args: []string{
			"-c",
			`set -xe
			if [ -f "/configdb-readonly/mongod.conf" ]; then
  				cp /configdb-readonly/mongod.conf /data/configdb/mongod.conf
			else
				touch /data/configdb/mongod.conf
			fi
			
			if [ -f "/keydir-readonly/key.txt" ]; then
  				cp /keydir-readonly/key.txt /data/configdb/key.txt
  				chmod 600 /data/configdb/key.txt
			fi

			if [ -f "/keydir-readonly/tls.crt" ]; then
  				cp /keydir-readonly/tls.crt /data/configdb/tls.crt
  				chmod 600 /data/configdb/tls.crt
			fi

			if [ -f "/keydir-readonly/tls.key" ]; then
  				cp /keydir-readonly/tls.key /data/configdb/tls.key
  				chmod 600 /data/configdb/tls.key
			fi

			if [ -f "/keydir-readonly/mongo.pem" ]; then
  				cp /keydir-readonly/mongo.pem /data/configdb/mongo.pem
  				chmod 600 /data/configdb/mongo.pem
			fi

			if [ -f "/keydir-readonly/client.pem" ]; then
  				cp /keydir-readonly/client.pem /data/configdb/client.pem
  				chmod 600 /data/configdb/client.pem
			fi`,
		},
		VolumeMounts: []core.VolumeMount{
			{
				Name:      configDirectoryName,
				MountPath: configDirectoryPath,
			},
		},
		Resources: pt.Spec.Resources,
	}

	initVolumes := []core.Volume{{
		Name: workDirectoryName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}}

	// mongodb.Spec.SSLMode can be empty if upgraded operator from previous version.
	// But, eventually it will be defaulted. TODO: delete `mongodb.Spec.SSLMode != ""` in future.
	sslMode := mongodb.Spec.SSLMode
	if sslMode == "" {
		sslMode = api.SSLModeDisabled
	}
	if sslMode != api.SSLModeDisabled || mongodb.Spec.ReplicaSet != nil || mongodb.Spec.ShardTopology != nil {
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
					DefaultMode: types.Int32P(256),
					SecretName:  mongodb.Spec.CertificateSecret.SecretName,
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
		if container.Name == api.ResourceSingularMongoDB {
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
	}
	for i, container := range template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB || container.Name == "exporter" {
			template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
		}
	}
	return template
}

func (c *Controller) checkStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	err := core_util.WaitUntilPodRunningBySelector(
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}
