/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

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
	"path/filepath"
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/pkg/errors"
	"gomodules.xyz/envsubst"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	CustomConfigMountPath         = "/elasticsearch/custom-config"
	ExporterCertDir               = "/usr/config/certs"
	ConfigMergerInitContainerName = "config-merger"
)

func (c *Controller) ensureStatefulSet(
	elasticsearch *api.Elasticsearch,
	pvcSpec *core.PersistentVolumeClaimSpec,
	resources core.ResourceRequirements,
	statefulSetName string,
	labels map[string]string,
	replicas int32,
	envList []core.EnvVar,
	nodeRole string,
	maxUnavailable *intstr.IntOrString,
) (kutil.VerbType, error) {

	esVersion, err := c.esVersionLister.Get(string(elasticsearch.Spec.Version))
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if err := c.checkStatefulSet(elasticsearch, statefulSetName); err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      statefulSetName,
		Namespace: elasticsearch.Namespace,
	}

	owner := metav1.NewControllerRef(elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	// Make a new map "labelSelector", so that it remains
	// unchanged even if the "labels" changes.
	// It contains:
	//	-	kubedb.com/kind: ResourceKindElasticsearch
	//	-	kubedb.com/name: elasticsearch.Name
	//	-	node.role.<master/data/client>: set
	labelSelector := elasticsearch.OffshootSelectors()
	labelSelector = core_util.UpsertMap(labelSelector, labels)

	initContainers := []core.Container{
		{
			Name:            "init-sysctl",
			Image:           esVersion.Spec.InitContainer.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Command:         []string{"sysctl", "-w", "vm.max_map_count=262144"},
			SecurityContext: &core.SecurityContext{
				Privileged: types.BoolP(true),
			},
			Resources: resources,
		},
	}
	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		initContainers = append(initContainers, upsertXpackInitContainer(elasticsearch, esVersion, envList))
	}

	affinity, err := parseAffinityTemplate(elasticsearch.Spec.PodTemplate.Spec.Affinity.DeepCopy(), nodeRole)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = core_util.UpsertMap(labels, elasticsearch.OffshootLabels())
			in.Annotations = elasticsearch.Spec.PodTemplate.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = types.Int32P(replicas)

			in.Spec.ServiceName = elasticsearch.GvrSvcName()
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labelSelector,
			}
			in.Spec.Template.Labels = labelSelector
			in.Spec.Template.Annotations = elasticsearch.Spec.PodTemplate.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				append(
					initContainers,
					elasticsearch.Spec.PodTemplate.Spec.InitContainers...,
				),
			)
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				in.Spec.Template.Spec.Containers,
				core.Container{
					Name:            api.ResourceSingularElasticsearch,
					Image:           esVersion.Spec.DB.Image,
					ImagePullPolicy: core.PullIfNotPresent,
					SecurityContext: &core.SecurityContext{
						Privileged: types.BoolP(false),
						Capabilities: &core.Capabilities{
							Add: []core.Capability{"IPC_LOCK", "SYS_RESOURCE"},
						},
					},
					Resources:      resources,
					LivenessProbe:  elasticsearch.Spec.PodTemplate.Spec.LivenessProbe,
					ReadinessProbe: elasticsearch.Spec.PodTemplate.Spec.ReadinessProbe,
					Lifecycle:      elasticsearch.Spec.PodTemplate.Spec.Lifecycle,
				})
			in = upsertEnv(in, elasticsearch, esVersion, envList)
			in = upsertUserEnv(in, elasticsearch)
			in = upsertPorts(in)
			in = upsertCustomConfig(in, elasticsearch, esVersion)

			in.Spec.Template.Spec.NodeSelector = elasticsearch.Spec.PodTemplate.Spec.NodeSelector
			in.Spec.Template.Spec.Affinity = affinity
			if elasticsearch.Spec.PodTemplate.Spec.SchedulerName != "" {
				in.Spec.Template.Spec.SchedulerName = elasticsearch.Spec.PodTemplate.Spec.SchedulerName
			}
			in.Spec.Template.Spec.Tolerations = elasticsearch.Spec.PodTemplate.Spec.Tolerations
			in.Spec.Template.Spec.ImagePullSecrets = elasticsearch.Spec.PodTemplate.Spec.ImagePullSecrets
			in.Spec.Template.Spec.PriorityClassName = elasticsearch.Spec.PodTemplate.Spec.PriorityClassName
			in.Spec.Template.Spec.Priority = elasticsearch.Spec.PodTemplate.Spec.Priority
			in.Spec.Template.Spec.SecurityContext = elasticsearch.Spec.PodTemplate.Spec.SecurityContext

			if nodeRole == NodeRoleClient {
				in = c.upsertMonitoringContainer(in, elasticsearch, esVersion)
				in = upsertDatabaseSecretForSG(in, esVersion, elasticsearch.Spec.DatabaseSecret.SecretName)
			}
			if !elasticsearch.Spec.DisableSecurity {
				in = upsertCertificate(in, elasticsearch.Spec.CertificateSecret.SecretName, nodeRole, esVersion)
			}

			if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack &&
				in.Spec.Template.Spec.SecurityContext == nil {
				in.Spec.Template.Spec.SecurityContext = &core.PodSecurityContext{
					FSGroup: types.Int64P(1000),
				}
			}

			in = upsertDatabaseConfigforXPack(in, elasticsearch, esVersion)

			in = upsertDataVolume(in, elasticsearch.Spec.StorageType, pvcSpec, esVersion)
			in = upsertTemporaryVolume(in)

			in.Spec.Template.Spec.ServiceAccountName = elasticsearch.Spec.PodTemplate.Spec.ServiceAccountName
			in.Spec.UpdateStrategy = elasticsearch.Spec.UpdateStrategy

			return in
		},
		metav1.PatchOptions{},
	)

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if vt == kutil.VerbCreated || vt == kutil.VerbPatched {
		// Check StatefulSet Pod status
		if err := c.CheckStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}

	// ensure pdb
	if maxUnavailable != nil {
		if err := c.createPodDisruptionBudget(statefulSet, maxUnavailable); err != nil {
			return vt, err
		}
	}

	return vt, nil
}

func (c *Controller) CheckStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	err := core_util.WaitUntilPodRunningBySelector(
		context.TODO(),
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

func getHeapSizeForNode(val int64) int64 {
	ret := val / 100
	return ret * 80
}

func (c *Controller) ensureClientNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()
	clientNode := elasticsearch.Spec.Topology.Client

	if clientNode.Prefix != "" {
		statefulSetName = fmt.Sprintf("%v-%v", clientNode.Prefix, statefulSetName)
	}

	labels := map[string]string{
		NodeRoleClient: NodeRoleSet,
	}

	heapSize := int64(134217728) // 128mb
	if request, found := clientNode.Resources.Requests[core.ResourceMemory]; found && request.Value() > 0 {
		heapSize = getHeapSizeForNode(request.Value())
	}

	esVersion, err := c.esVersionLister.Get(elasticsearch.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
		// following envs are used in xpack too for `config-merge` init container
		{
			Name:  "NODE_MASTER",
			Value: "false",
		},
		{
			Name:  "NODE_DATA",
			Value: "false",
		},
		{
			Name:  "NODE_INGEST",
			Value: "true",
		},
	}

	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "MODE",
				Value: "client",
			},
		}...)
	} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "node.ingest",
				Value: "true",
			},
			{
				Name:  "node.master",
				Value: "false",
			},
			{
				Name:  "node.data",
				Value: "false",
			},
		}...)
	}

	replicas := int32(1)
	if clientNode.Replicas != nil {
		replicas = types.Int32(clientNode.Replicas)
	}
	maxUnavailable := elasticsearch.Spec.Topology.Client.MaxUnavailable

	return c.ensureStatefulSet(elasticsearch, clientNode.Storage, clientNode.Resources, statefulSetName, labels, replicas, envList, NodeRoleClient, maxUnavailable)
}

func (c *Controller) ensureMasterNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()
	masterNode := elasticsearch.Spec.Topology.Master

	if masterNode.Prefix != "" {
		statefulSetName = fmt.Sprintf("%v-%v", masterNode.Prefix, statefulSetName)
	}

	esVersion, err := c.esVersionLister.Get(elasticsearch.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	labels := map[string]string{
		NodeRoleMaster: NodeRoleSet,
	}

	heapSize := int64(134217728) // 128mb
	if request, found := masterNode.Resources.Requests[core.ResourceMemory]; found && request.Value() > 0 {
		heapSize = getHeapSizeForNode(request.Value())
	}

	replicas := int32(1)
	if masterNode.Replicas != nil {
		replicas = types.Int32(masterNode.Replicas)
	}

	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
		// following envs are used in xpack too for `config-merge` init container
		{
			Name:  "NODE_MASTER",
			Value: "true",
		},
		{
			Name:  "NODE_DATA",
			Value: "false",
		},
		{
			Name:  "NODE_INGEST",
			Value: "false",
		},
	}

	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "HTTP_ENABLE",
				Value: "false",
			},
			{
				Name:  "NUMBER_OF_MASTERS",
				Value: fmt.Sprintf("%v", (replicas/2)+1),
			},
		}...)
		if strings.HasPrefix(esVersion.Spec.Version, "7.") {
			envList = append(envList, core.EnvVar{
				Name:  "INITIAL_MASTER_NODES",
				Value: getInitialMasterNodes(elasticsearch),
			})
		}
	} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "node.master",
				Value: "true",
			},
			{
				Name:  "node.data",
				Value: "false",
			},
			{
				Name:  "node.ingest",
				Value: "false",
			},
		}...)

		if strings.HasPrefix(esVersion.Spec.Version, "7.") {
			envList = append(envList, core.EnvVar{
				Name:  "cluster.initial_master_nodes",
				Value: getInitialMasterNodes(elasticsearch),
			})
		} else {
			envList = append(envList, core.EnvVar{
				Name:  "discovery.zen.minimum_master_nodes",
				Value: fmt.Sprintf("%v", (replicas/2)+1),
			})
		}
	}

	maxUnavailable := elasticsearch.Spec.Topology.Master.MaxUnavailable

	return c.ensureStatefulSet(elasticsearch, masterNode.Storage, masterNode.Resources, statefulSetName, labels, replicas, envList, NodeRoleMaster, maxUnavailable)
}

func (c *Controller) ensureDataNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()
	dataNode := elasticsearch.Spec.Topology.Data

	if dataNode.Prefix != "" {
		statefulSetName = fmt.Sprintf("%v-%v", dataNode.Prefix, statefulSetName)
	}

	labels := map[string]string{
		NodeRoleData: NodeRoleSet,
	}

	heapSize := int64(134217728) // 128mb
	if request, found := dataNode.Resources.Requests[core.ResourceMemory]; found && request.Value() > 0 {
		heapSize = getHeapSizeForNode(request.Value())
	}

	esVersion, err := c.esVersionLister.Get(elasticsearch.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
		// following envs are used in xpack too for `config-merge` init container
		{
			Name:  "NODE_MASTER",
			Value: "false",
		},
		{
			Name:  "NODE_DATA",
			Value: "true",
		},
		{
			Name:  "NODE_INGEST",
			Value: "false",
		},
	}

	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "HTTP_ENABLE",
				Value: "false",
			},
		}...)
	} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "node.master",
				Value: "false",
			},
			{
				Name:  "node.data",
				Value: "true",
			},
			{
				Name:  "node.ingest",
				Value: "false",
			},
		}...)
	}

	replicas := int32(1)
	if dataNode.Replicas != nil {
		replicas = types.Int32(dataNode.Replicas)
	}

	maxUnavailable := elasticsearch.Spec.Topology.Data.MaxUnavailable

	return c.ensureStatefulSet(elasticsearch, dataNode.Storage, dataNode.Resources, statefulSetName, labels, replicas, envList, NodeRoleData, maxUnavailable)
}

func (c *Controller) ensureCombinedNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	statefulSetName := elasticsearch.OffshootName()

	labels := map[string]string{
		NodeRoleClient: NodeRoleSet,
		NodeRoleMaster: NodeRoleSet,
		NodeRoleData:   NodeRoleSet,
	}

	esVersion, err := c.esVersionLister.Get(elasticsearch.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	replicas := int32(1)
	if elasticsearch.Spec.Replicas != nil {
		replicas = types.Int32(elasticsearch.Spec.Replicas)
	}

	heapSize := int64(134217728) // 128mb
	if elasticsearch.Spec.PodTemplate.Spec.Resources.Size() != 0 {
		if request, found := elasticsearch.Spec.PodTemplate.Spec.Resources.Requests[core.ResourceMemory]; found && request.Value() > 0 {
			heapSize = getHeapSizeForNode(request.Value())
		}
	}

	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
		// following envs are used in xpack too for `config-merge` init container
		{
			Name:  "NODE_MASTER",
			Value: "true",
		},
		{
			Name:  "NODE_DATA",
			Value: "true",
		},
		{
			Name:  "NODE_INGEST",
			Value: "true",
		},
	}

	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "NUMBER_OF_MASTERS",
				Value: fmt.Sprintf("%v", (replicas/2)+1),
			},
			{
				Name:  "MODE",
				Value: "client",
			},
		}...)
		if strings.HasPrefix(esVersion.Spec.Version, "7.") {
			envList = append(envList, core.EnvVar{
				Name:  "INITIAL_MASTER_NODES",
				Value: getInitialMasterNodes(elasticsearch),
			})
		}
	} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		envList = append(envList, []core.EnvVar{
			{
				Name:  "node.master",
				Value: "true",
			},
			{
				Name:  "node.data",
				Value: "true",
			},
			{
				Name:  "node.ingest",
				Value: "true",
			},
		}...)

		if strings.HasPrefix(esVersion.Spec.Version, "7.") {
			envList = append(envList, core.EnvVar{
				Name:  "cluster.initial_master_nodes",
				Value: getInitialMasterNodes(elasticsearch),
			})
		} else {
			envList = append(envList, core.EnvVar{
				Name:  "discovery.zen.minimum_master_nodes",
				Value: fmt.Sprintf("%v", (replicas/2)+1),
			})
		}
	}

	var pvcSpec core.PersistentVolumeClaimSpec
	var resources core.ResourceRequirements
	if elasticsearch.Spec.Storage != nil {
		pvcSpec = *elasticsearch.Spec.Storage
	}
	if elasticsearch.Spec.PodTemplate.Spec.Resources.Size() != 0 {
		resources = elasticsearch.Spec.PodTemplate.Spec.Resources
	}

	maxUnavailable := elasticsearch.Spec.MaxUnavailable

	return c.ensureStatefulSet(elasticsearch, &pvcSpec, resources, statefulSetName, labels, replicas, envList, NodeRoleClient, maxUnavailable)
}

func (c *Controller) checkStatefulSet(elasticsearch *api.Elasticsearch, name string) error {
	elasticsearchName := elasticsearch.OffshootName()
	// SatatefulSet for Elasticsearch database
	statefulSet, err := c.Client.AppsV1().StatefulSets(elasticsearch.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch ||
		statefulSet.Labels[api.LabelDatabaseName] != elasticsearchName {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, elasticsearch.Namespace, name)
	}

	return nil
}

func upsertEnv(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion, envs []core.EnvVar) *apps.StatefulSet {
	var envList []core.EnvVar

	if !elasticsearch.Spec.DisableSecurity {
		envList = append(envList, core.EnvVar{
			Name: "KEY_PASS",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: elasticsearch.Spec.CertificateSecret.SecretName,
					},
					Key: "key_pass",
				},
			},
		})
	} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
		// Older versions of Elasticsearch (ie, 5.6.4, 6.2.4, 6.3.0, 6.4.0) requires KEY_PASS to be set.
		// So set a empty value in KEY_PASS
		envList = append(envList, core.EnvVar{
			Name:  "KEY_PASS",
			Value: "",
		})
	}
	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
		envList = append(envList, []core.EnvVar{
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
				Name:  "DISCOVERY_SERVICE",
				Value: elasticsearch.MasterServiceName(),
			},
			{
				Name:  "SSL_ENABLE",
				Value: fmt.Sprintf("%v", elasticsearch.Spec.EnableSSL),
			},
		}...)
		if elasticsearch.Spec.DisableSecurity {
			envList = append(envList, core.EnvVar{

				Name:  "SEARCHGUARD_DISABLED",
				Value: "true",
			})
		} else {
			envList = append(envList, core.EnvVar{

				Name:  "SEARCHGUARD_DISABLED",
				Value: "false",
			})
		}
	} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		envList = append(envList, []core.EnvVar{
			{
				Name: "node.name",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name:  "cluster.name",
				Value: elasticsearch.Name,
			},
			{
				Name:  "network.host",
				Value: "0.0.0.0",
			},
			{
				Name: "ELASTIC_USER",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: elasticsearch.Spec.DatabaseSecret.SecretName,
						},
						Key: KeyAdminUserName,
					},
				},
			},
			{
				Name: "ELASTIC_PASSWORD",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: elasticsearch.Spec.DatabaseSecret.SecretName,
						},
						Key: KeyAdminPassword,
					},
				},
			},
		}...)

		if !elasticsearch.Spec.DisableSecurity {
			envList = append(envList, core.EnvVar{
				Name:  "xpack.security.http.ssl.enabled",
				Value: fmt.Sprintf("%v", elasticsearch.Spec.EnableSSL),
			})
		}

		if strings.HasPrefix(esVersion.Spec.Version, "7.") {
			envList = append(envList, core.EnvVar{
				Name:  "discovery.seed_hosts",
				Value: elasticsearch.MasterServiceName(),
			})
		} else {
			envList = append(envList, core.EnvVar{
				Name:  "discovery.zen.ping.unicast.hosts",
				Value: elasticsearch.MasterServiceName(),
			})
		}
	}

	envList = append(envList, envs...)

	// To do this, Upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
			return statefulSet
		}
	}

	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, elasticsearch.Spec.PodTemplate.Spec.Env...)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertPorts(statefulSet *apps.StatefulSet) *apps.StatefulSet {
	getPorts := func() []core.ContainerPort {
		portList := []core.ContainerPort{
			{
				Name:          api.ElasticsearchNodePortName,
				ContainerPort: api.ElasticsearchNodePort,
				Protocol:      core.ProtocolTCP,
			},
		}

		portList = append(portList, core.ContainerPort{
			Name:          api.ElasticsearchRestPortName,
			ContainerPort: api.ElasticsearchRestPort,
			Protocol:      core.ProtocolTCP,
		})

		return portList
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			statefulSet.Spec.Template.Spec.Containers[i].Ports = getPorts()
			return statefulSet
		}
	}

	return statefulSet
}

func (c *Controller) upsertMonitoringContainer(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion) *apps.StatefulSet {
	if elasticsearch.GetMonitoringVendor() == mona.VendorPrometheus {
		container := core.Container{
			Name: "exporter",
			Args: append([]string{
				fmt.Sprintf("--es.uri=%s", getURI(elasticsearch, esVersion)),
				fmt.Sprintf("--web.listen-address=:%d", api.PrometheusExporterPortNumber),
				fmt.Sprintf("--web.telemetry-path=%s", elasticsearch.StatsService().Path()),
			}, elasticsearch.Spec.Monitor.Prometheus.Exporter.Args...),
			Image:           esVersion.Spec.Exporter.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(api.PrometheusExporterPortNumber),
				},
			},
			Env:             elasticsearch.Spec.Monitor.Prometheus.Exporter.Env,
			Resources:       elasticsearch.Spec.Monitor.Prometheus.Exporter.Resources,
			SecurityContext: elasticsearch.Spec.Monitor.Prometheus.Exporter.SecurityContext,
		}
		envList := []core.EnvVar{
			{
				Name: "DB_USER",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: elasticsearch.Spec.DatabaseSecret.SecretName,
						},
						Key: KeyAdminUserName,
					},
				},
			},
			{
				Name: "DB_PASSWORD",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: elasticsearch.Spec.DatabaseSecret.SecretName,
						},
						Key: KeyAdminPassword,
					},
				},
			},
		}
		container.Env = core_util.UpsertEnvVars(container.Env, envList...)

		if elasticsearch.Spec.EnableSSL {
			certVolumeMount := core.VolumeMount{
				Name:      "exporter-certs",
				MountPath: ExporterCertDir,
			}
			container.VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, certVolumeMount)

			volume := core.Volume{
				Name: "exporter-certs",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: elasticsearch.Spec.CertificateSecret.SecretName,
						Items: []core.KeyToPath{
							{
								Key:  "root.pem",
								Path: "root.pem",
							},
						},
					},
				},
			}

			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, volume)
			esCaFlag := "--es.ca=" + filepath.Join(ExporterCertDir, "root.pem")

			if len(container.Args) == 0 || container.Args[len(container.Args)-1] != esCaFlag {
				container.Args = append(container.Args, esCaFlag)
			}
		}
		statefulSet.Spec.Template.Spec.Containers = core_util.UpsertContainer(statefulSet.Spec.Template.Spec.Containers, container)
	}
	return statefulSet
}

func upsertCertificate(statefulSet *apps.StatefulSet, secretName string, nodeRole string, esVersion *catalog.ElasticsearchVersion) *apps.StatefulSet {
	addCertVolume := func() *core.SecretVolumeSource {
		svs := &core.SecretVolumeSource{
			SecretName: secretName,
			Items: []core.KeyToPath{
				{
					Key:  rootKeyStore,
					Path: rootKeyStore,
				},
				{
					Key:  nodeKeyStore,
					Path: nodeKeyStore,
				},
			},
		}

		svs.Items = append(svs.Items, core.KeyToPath{
			Key:  clientKeyStore,
			Path: clientKeyStore,
		})

		if nodeRole == NodeRoleClient && esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
			svs.Items = append(svs.Items, core.KeyToPath{
				Key:  sgAdminKeyStore,
				Path: sgAdminKeyStore,
			})
		}
		return svs
	}

	mountPath := ConfigFileMountPathSG
	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		mountPath = ConfigFileMountPath
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "certs",
				MountPath: filepath.Join(mountPath, "certs"),
			}

			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, volumeMount)

			volume := core.Volume{
				Name: "certs",
				VolumeSource: core.VolumeSource{
					Secret: addCertVolume(),
				},
			}

			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, volume)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertDatabaseSecretForSG(statefulSet *apps.StatefulSet, esVersion *catalog.ElasticsearchVersion, secretName string) *apps.StatefulSet {
	// currently only searchguard requires upserting database secret volume
	if esVersion.Spec.AuthPlugin != catalog.ElasticsearchAuthPluginSearchGuard {
		return statefulSet
	}
	searchGuard := string(esVersion.Spec.Version[0])
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "sgconfig",
				MountPath: fmt.Sprintf("/elasticsearch/plugins/search-guard-%v/sgconfig", searchGuard),
			}
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, volumeMount)

			volume := core.Volume{
				Name: "sgconfig",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, volume)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertDatabaseConfigforXPack(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion) *apps.StatefulSet {
	cmName := fmt.Sprintf("%v-%v", elasticsearch.OffshootName(), DatabaseConfigMapSuffix)
	// currently only xpack requires upserting database configuration
	if esVersion.Spec.AuthPlugin != catalog.ElasticsearchAuthPluginXpack {
		return statefulSet
	}
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "esconfig",
				MountPath: filepath.Join(ConfigFileMountPath, ConfigFileName),
				SubPath:   ConfigFileName,
			}
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, volumeMount)

			ed := core.EmptyDirVolumeSource{}
			volumeList := []core.Volume{
				{
					Name: "esconfig",
					VolumeSource: core.VolumeSource{
						EmptyDir: &ed,
					},
				},
			}
			if !elasticsearch.Spec.DisableSecurity {
				volumeList = append(volumeList, core.Volume{
					Name: "temp-esconfig",
					VolumeSource: core.VolumeSource{
						ConfigMap: &core.ConfigMapVolumeSource{
							LocalObjectReference: core.LocalObjectReference{
								Name: cmName,
							},
						},
					},
				})
			}
			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, volumeList...)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertDataVolume(statefulSet *apps.StatefulSet, st api.StorageType, pvcSpec *core.PersistentVolumeClaimSpec, esVersion *catalog.ElasticsearchVersion) *apps.StatefulSet {
	dataPath := "/data"
	if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		dataPath = "/usr/share/elasticsearch/data"
	}
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: dataPath,
			}
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, volumeMount)

			if st == api.StorageTypeEphemeral {
				ed := core.EmptyDirVolumeSource{}
				if pvcSpec != nil {
					if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
						ed.SizeLimit = &sz
					}
				}
				statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
					statefulSet.Spec.Template.Spec.Volumes,
					core.Volume{
						Name: "data",
						VolumeSource: core.VolumeSource{
							EmptyDir: &ed,
						},
					})
			} else {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, pvcSpec)
				}

				claim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					claim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				statefulSet.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(statefulSet.Spec.VolumeClaimTemplates, claim)
			}

			return statefulSet
		}
	}
	return statefulSet
}

func upsertTemporaryVolume(statefulSet *apps.StatefulSet) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularElasticsearch {
			volumeMount := core.VolumeMount{
				Name:      "temp",
				MountPath: "/tmp",
			}
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, volumeMount)

			volume := core.Volume{
				Name: "temp",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			}
			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, volume)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertCustomConfig(statefulSet *apps.StatefulSet, elasticsearch *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion) *apps.StatefulSet {
	if elasticsearch.Spec.ConfigSource != nil {
		if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard {
			for i, container := range statefulSet.Spec.Template.Spec.Containers {
				if container.Name == api.ResourceSingularElasticsearch {
					configVolumeMount := core.VolumeMount{
						Name:      "custom-config",
						MountPath: CustomConfigMountPath,
					}
					statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, configVolumeMount)

					configVolume := core.Volume{
						Name:         "custom-config",
						VolumeSource: *elasticsearch.Spec.ConfigSource,
					}
					statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, configVolume)
					break
				}
			}
		} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
			for i, container := range statefulSet.Spec.Template.Spec.InitContainers {
				if container.Name == ConfigMergerInitContainerName {
					configVolumeMount := core.VolumeMount{
						Name:      "custom-config",
						MountPath: CustomConfigMountPath,
					}
					statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = core_util.UpsertVolumeMount(container.VolumeMounts, configVolumeMount)

					configVolume := core.Volume{
						Name:         "custom-config",
						VolumeSource: *elasticsearch.Spec.ConfigSource,
					}
					statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, configVolume)
					break
				}
			}
		}
	}
	return statefulSet
}

func getURI(es *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion) string {
	if es.Spec.DisableSecurity {
		return fmt.Sprintf("%s://localhost:%d", es.GetConnectionScheme(), api.ElasticsearchRestPort)
	} else if esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginSearchGuard || esVersion.Spec.AuthPlugin == catalog.ElasticsearchAuthPluginXpack {
		return fmt.Sprintf("%s://$(DB_USER):$(DB_PASSWORD)@localhost:%d", es.GetConnectionScheme(), api.ElasticsearchRestPort)
	} else {
		log.Infoln("Invalid Auth Plugin")
	}
	return ""
}

// INITIAL_MASTER_NODES value for >= ES7
func getInitialMasterNodes(es *api.Elasticsearch) string {
	var value string
	stsName := getMasterNodeStatefulsetName(es)
	replicas := types.Int32(es.Spec.Replicas)
	if es.Spec.Topology != nil {
		replicas = types.Int32(es.Spec.Topology.Master.Replicas)
	}

	for i := int32(0); i < replicas; i++ {
		if i != 0 {
			value += ","
		}
		value += fmt.Sprintf("%v-%v", stsName, i)
	}

	return value
}

func getMasterNodeStatefulsetName(elasticsearch *api.Elasticsearch) string {
	statefulSetName := elasticsearch.OffshootName()
	topology := elasticsearch.Spec.Topology

	if topology != nil && topology.Master.Prefix != "" {
		statefulSetName = fmt.Sprintf("%v-%v", topology.Master.Prefix, statefulSetName)
	}
	return statefulSetName
}

func upsertXpackInitContainer(elasticsearch *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion, envList []core.EnvVar) core.Container {
	volumeMounts := []core.VolumeMount{
		{
			Name:      "esconfig",
			MountPath: ConfigFileMountPath,
		},
		{
			Name:      "data",
			MountPath: "/usr/share/elasticsearch/data",
		},
	}
	if !elasticsearch.Spec.DisableSecurity {
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      "temp-esconfig",
			MountPath: TempConfigFileMountPath,
		})
	}

	return core.Container{
		Name:            ConfigMergerInitContainerName,
		Image:           esVersion.Spec.InitContainer.YQImage,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"sh"},
		Env:             envList,
		Args: []string{
			"-c",
			`set -x
echo "changing ownership of data folder: /usr/share/elasticsearch/data"
chown -R 1000:1000 /usr/share/elasticsearch/data

TEMP_CONFIG_FILE=/elasticsearch/temp-config/elasticsearch.yml
CUSTOM_CONFIG_DIR="/elasticsearch/custom-config"
CONFIG_FILE=/usr/share/elasticsearch/config/elasticsearch.yml

if [ -f $TEMP_CONFIG_FILE ]; then
  cp $TEMP_CONFIG_FILE $CONFIG_FILE
else
  touch $CONFIG_FILE
fi

# yq changes the file permissions after merging custom configuration.
# we need to restore the original permissions after merging done.
ORIGINAL_PERMISSION=$(stat -c '%a' $CONFIG_FILE)

# if common-config file exist then apply it
if [ -f $CUSTOM_CONFIG_DIR/common-config.yml ]; then
  yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/common-config.yml
elif [ -f $CUSTOM_CONFIG_DIR/common-config.yaml ]; then
  yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/common-config.yaml
fi

# if it is data node and data-config file exist then apply it
if [[ "$NODE_DATA" == true ]]; then
  if [ -f $CUSTOM_CONFIG_DIR/data-config.yml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/data-config.yml
  elif [ -f $CUSTOM_CONFIG_DIR/data-config.yaml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/data-config.yaml
  fi
fi

# if it is client node and client-config file exist then apply it
if [[ "$NODE_INGEST" == true ]]; then
  if [ -f $CUSTOM_CONFIG_DIR/client-config.yml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/client-config.yml
  elif [ -f $CUSTOM_CONFIG_DIR/client-config.yaml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/client-config.yaml
  fi
fi

# if it is master node and mater-config file exist then apply it
if [[ "$NODE_MASTER" == true ]]; then
  if [ -f $CUSTOM_CONFIG_DIR/master-config.yml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/master-config.yml
  elif [ -f $CUSTOM_CONFIG_DIR/master-config.yaml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/master-config.yaml
  fi
fi

# restore original permission of elasticsearh.yml file
if [[ "$ORIGINAL_PERMISSION" != "" ]]; then
  chmod $ORIGINAL_PERMISSION $CONFIG_FILE
fi
`,
		},
		VolumeMounts: volumeMounts,
	}
}

func parseAffinityTemplate(affinity *core.Affinity, nodeRole string) (*core.Affinity, error) {
	if affinity == nil {
		return nil, errors.New("affinity is nil")
	}

	templateMap := map[string]string{
		api.ElasticsearchNodeAffinityTemplateVar: nodeRole,
	}

	jsonObj, err := json.Marshal(affinity)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal affinity")
	}

	resolved, err := envsubst.EvalMap(string(jsonObj), templateMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(resolved), affinity)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the affinity")
	}

	return affinity, nil
}
