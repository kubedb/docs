package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	mon_api "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1beta1"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) ensureStatefulSet(
	elasticsearch *api.Elasticsearch,
	statefulSetName string,
	labels map[string]string,
	replicas int32,
	envList []core.EnvVar,
	isClient bool,
) (kutil.VerbType, error) {

	if err := c.checkStatefulSet(elasticsearch, statefulSetName); err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      statefulSetName,
		Namespace: elasticsearch.Namespace,
	}

	if replicas <= 0 {
		replicas = 1
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in = upsertObjectMeta(in, labels, elasticsearch.StatefulSetAnnotations())

		in.Spec.Replicas = types.Int32P(replicas)
		in.Spec.ServiceName = c.opt.GoverningService
		in.Spec.Template.Labels = in.Labels

		in = upsertInitContainer(in)
		in = c.upsertContainer(in, elasticsearch)
		in = upsertEnv(in, elasticsearch, envList)
		in = upsertPort(in, isClient)

		in.Spec.Template.Spec.NodeSelector = elasticsearch.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = elasticsearch.Spec.Affinity

		if elasticsearch.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = elasticsearch.Spec.SchedulerName
		}

		in.Spec.Template.Spec.Tolerations = elasticsearch.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = elasticsearch.Spec.ImagePullSecrets

		if isClient {
			in = c.upsertMonitoringContainer(in, elasticsearch)
			in = upsertDatabaseSecret(in, elasticsearch.Spec.DatabaseSecret.SecretName)
		}

		in = upsertCertificate(in, elasticsearch.Spec.CertificateSecret.SecretName, isClient, elasticsearch.Spec.EnableSSL)
		in = upsertDataVolume(in, elasticsearch)
		in.Spec.UpdateStrategy.Type = apps.RollingUpdateStatefulSetStrategyType

		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if vt == kutil.VerbCreated || vt == kutil.VerbPatched {
		// Check StatefulSet Pod status
		if err := c.CheckStatefulSetPodStatus(statefulSet); err != nil {
			c.recorder.Eventf(
				elasticsearch.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToStart,
				`Failed to be running after StatefulSet %v. Reason: %v`,
				vt,
				err,
			)
			return kutil.VerbUnchanged, err
		}

		c.recorder.Eventf(
			elasticsearch.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) CheckStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
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

func (c *Controller) ensureClientNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()
	clientNode := elasticsearch.Spec.Topology.Client

	if clientNode.Prefix != "" {
		statefulSetName = fmt.Sprintf("%v-%v", clientNode.Prefix, statefulSetName)
	}

	labels := elasticsearch.StatefulSetLabels()
	labels[NodeRoleClient] = "set"

	envList := []core.EnvVar{
		{
			Name:  "NODE_MASTER",
			Value: fmt.Sprintf("%v", false),
		},
		{
			Name:  "NODE_DATA",
			Value: fmt.Sprintf("%v", false),
		},
		{
			Name:  "MODE",
			Value: "client",
		},
	}

	return c.ensureStatefulSet(elasticsearch, statefulSetName, labels, clientNode.Replicas, envList, true)
}

func (c *Controller) ensureMasterNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()
	masterNode := elasticsearch.Spec.Topology.Master

	if masterNode.Prefix != "" {
		statefulSetName = fmt.Sprintf("%v-%v", masterNode.Prefix, statefulSetName)
	}

	labels := elasticsearch.StatefulSetLabels()
	labels[NodeRoleMaster] = "set"

	replicas := masterNode.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	envList := []core.EnvVar{
		{
			Name:  "NODE_DATA",
			Value: fmt.Sprintf("%v", false),
		},
		{
			Name:  "NODE_INGEST",
			Value: fmt.Sprintf("%v", false),
		},
		{
			Name:  "HTTP_ENABLE",
			Value: fmt.Sprintf("%v", false),
		},
		{
			Name:  "NUMBER_OF_MASTERS",
			Value: fmt.Sprintf("%v", (replicas/2)+1),
		},
	}

	return c.ensureStatefulSet(elasticsearch, statefulSetName, labels, masterNode.Replicas, envList, false)
}

func (c *Controller) ensureDataNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()
	dataNode := elasticsearch.Spec.Topology.Data

	if dataNode.Prefix != "" {
		statefulSetName = fmt.Sprintf("%v-%v", dataNode.Prefix, statefulSetName)
	}

	labels := elasticsearch.StatefulSetLabels()
	labels[NodeRoleData] = "set"

	envList := []core.EnvVar{
		{
			Name:  "NODE_MASTER",
			Value: fmt.Sprintf("%v", false),
		},
		{
			Name:  "NODE_INGEST",
			Value: fmt.Sprintf("%v", false),
		},
		{
			Name:  "HTTP_ENABLE",
			Value: fmt.Sprintf("%v", false),
		},
	}

	return c.ensureStatefulSet(elasticsearch, statefulSetName, labels, dataNode.Replicas, envList, false)
}

func (c *Controller) ensureCombinedNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()
	labels := elasticsearch.StatefulSetLabels()
	labels[NodeRoleClient] = "set"
	labels[NodeRoleMaster] = "set"
	labels[NodeRoleData] = "set"

	replicas := elasticsearch.Spec.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	envList := []core.EnvVar{
		{
			Name:  "NUMBER_OF_MASTERS",
			Value: fmt.Sprintf("%v", (replicas/2)+1),
		},
		{
			Name:  "MODE",
			Value: "client",
		},
	}

	return c.ensureStatefulSet(elasticsearch, statefulSetName, labels, replicas, envList, true)
}

func (c *Controller) checkStatefulSet(elasticsearch *api.Elasticsearch, name string) error {
	elasticsearchName := elasticsearch.OffshootName()
	// SatatefulSet for Elasticsearch database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(elasticsearch.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch ||
		statefulSet.Labels[api.LabelDatabaseName] != elasticsearchName {
		return fmt.Errorf(`intended statefulSet "%v" already exists`, name)
	}

	return nil
}

func upsertObjectMeta(statefulSet *apps.StatefulSet, labels, annotations map[string]string) *apps.StatefulSet {
	statefulSet.Labels = core_util.UpsertMap(statefulSet.Labels, labels)
	statefulSet.Annotations = core_util.UpsertMap(statefulSet.Annotations, annotations)
	return statefulSet
}

func upsertInitContainer(statefulSet *apps.StatefulSet) *apps.StatefulSet {
	container := core.Container{
		Name:            "init-sysctl",
		Image:           "busybox",
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"sysctl", "-w", "vm.max_map_count=262144"},
		SecurityContext: &core.SecurityContext{
			Privileged: types.BoolP(true),
		},
	}
	initContainers := statefulSet.Spec.Template.Spec.InitContainers
	initContainers = core_util.UpsertContainer(initContainers, container)
	statefulSet.Spec.Template.Spec.InitContainers = initContainers
	return statefulSet
}

func (c *Controller) upsertContainer(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch) *apps.StatefulSet {
	container := core.Container{
		Name:  api.ResourceNameElasticsearch,
		Image: c.opt.Docker.GetImageWithTag(elasticsearch),
		SecurityContext: &core.SecurityContext{
			Privileged: types.BoolP(false),
			Capabilities: &core.Capabilities{
				Add: []core.Capability{"IPC_LOCK", "SYS_RESOURCE"},
			},
		},
	}
	containers := statefulSet.Spec.Template.Spec.Containers
	containers = core_util.UpsertContainer(containers, container)
	statefulSet.Spec.Template.Spec.Containers = containers
	return statefulSet
}

func upsertEnv(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch, envs []core.EnvVar) *apps.StatefulSet {

	envList := []core.EnvVar{
		{
			Name:  "CLUSTER_NAME",
			Value: elasticsearch.Name,
		},
		{
			Name: "NODE_NAME",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "ES_JAVA_OPTS",
			Value: "-Xms512m -Xmx512m",
		},
		{
			Name:  "DISCOVERY_SERVICE",
			Value: elasticsearch.MasterServiceName(),
		},
		{
			Name:  "SSL_ENABLE",
			Value: fmt.Sprintf("%v", elasticsearch.Spec.EnableSSL),
		},
		{
			Name: "KEY_PASS",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: elasticsearch.Spec.CertificateSecret.SecretName,
					},
					Key: "key_pass",
				},
			},
		},
	}

	envList = append(envList, envs...)

	// To do this, Upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameElasticsearch {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
			return statefulSet
		}
	}

	return statefulSet
}

func upsertPort(statefulSet *apps.StatefulSet, isClient bool) *apps.StatefulSet {

	getPorts := func() []core.ContainerPort {
		portList := []core.ContainerPort{
			{
				Name:          ElasticsearchNodePortName,
				ContainerPort: ElasticsearchNodePort,
				Protocol:      core.ProtocolTCP,
			},
		}
		if isClient {
			portList = append(portList, core.ContainerPort{
				Name:          ElasticsearchRestPortName,
				ContainerPort: ElasticsearchRestPort,
				Protocol:      core.ProtocolTCP,
			})
		}

		return portList
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameElasticsearch {
			statefulSet.Spec.Template.Spec.Containers[i].Ports = getPorts()
			return statefulSet
		}
	}

	return statefulSet
}

func (c *Controller) upsertMonitoringContainer(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch) *apps.StatefulSet {
	if elasticsearch.GetMonitoringVendor() == mon_api.VendorPrometheus {
		container := core.Container{
			Name: "exporter",
			Args: append([]string{
				"export",
				fmt.Sprintf("--address=:%d", api.PrometheusExporterPortNumber),
				fmt.Sprintf("--analytics=%v", c.opt.EnableAnalytics),
			}, c.opt.LoggerOptions.ToFlags()...),
			Image:           c.opt.Docker.GetOperatorImageWithTag(elasticsearch),
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(api.PrometheusExporterPortNumber),
				},
			},
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "secret",
					MountPath: ExporterSecretPath,
				},
			},
		}
		containers := statefulSet.Spec.Template.Spec.Containers
		containers = core_util.UpsertContainer(containers, container)
		statefulSet.Spec.Template.Spec.Containers = containers

		volume := core.Volume{
			Name: "secret",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: elasticsearch.Spec.DatabaseSecret.SecretName,
				},
			},
		}
		volumes := statefulSet.Spec.Template.Spec.Volumes
		volumes = core_util.UpsertVolume(volumes, volume)
		statefulSet.Spec.Template.Spec.Volumes = volumes
	}
	return statefulSet
}

func upsertCertificate(statefulSet *apps.StatefulSet, secretName string, isClientNode, isEnalbeSSL bool) *apps.StatefulSet {
	addCertVolume := func() *core.SecretVolumeSource {
		svs := &core.SecretVolumeSource{
			SecretName: secretName,
			Items: []core.KeyToPath{
				{
					Key:  "root.jks",
					Path: "root.jks",
				},
				{
					Key:  "node.jks",
					Path: "node.jks",
				},
			},
		}

		if isEnalbeSSL {
			svs.Items = append(svs.Items, core.KeyToPath{
				Key:  "client.jks",
				Path: "client.jks",
			})
		}

		if isClientNode {
			svs.Items = append(svs.Items, core.KeyToPath{
				Key:  "sgadmin.jks",
				Path: "sgadmin.jks",
			})
		}
		return svs
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "certs",
				MountPath: "/elasticsearch/config/certs",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name: "certs",
				VolumeSource: core.VolumeSource{
					Secret: addCertVolume(),
				},
			}
			volumes := statefulSet.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulSet.Spec.Template.Spec.Volumes = volumes
			return statefulSet
		}
	}
	return statefulSet
}

func upsertDatabaseSecret(statefulSet *apps.StatefulSet, secretName string) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "sgconfig",
				MountPath: "/elasticsearch/plugins/search-guard-5/sgconfig",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name: "sgconfig",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			volumes := statefulSet.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulSet.Spec.Template.Spec.Volumes = volumes
			return statefulSet
		}
	}
	return statefulSet
}

func upsertDataVolume(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/data",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := elasticsearch.Spec.Storage
			if pvcSpec != nil {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, *pvcSpec)
				}

				volumeClaim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					volumeClaim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				volumeClaims := statefulSet.Spec.VolumeClaimTemplates
				volumeClaims = core_util.UpsertVolumeClaim(volumeClaims, volumeClaim)
				statefulSet.Spec.VolumeClaimTemplates = volumeClaims
			} else {
				// Attach Empty directory
				volume := core.Volume{
					Name: "data",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				}
				volumes := statefulSet.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, volume)
				statefulSet.Spec.Template.Spec.Volumes = volumes
				return statefulSet

			}
			break
		}
	}
	return statefulSet
}
