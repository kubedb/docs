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

package search_guard

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	certlib "kubedb.dev/elasticsearch/pkg/lib/cert"
	"kubedb.dev/elasticsearch/pkg/lib/kernel"

	"github.com/pkg/errors"
	"gomodules.xyz/envsubst"
	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	ExporterCertDir = "/usr/config/certs"
)

var (
	defaultRestPort = core.ContainerPort{
		Name:          api.ElasticsearchRestPortName,
		ContainerPort: api.ElasticsearchRestPort,
		Protocol:      core.ProtocolTCP,
	}
	defaultTransportPort = core.ContainerPort{
		Name:          api.ElasticsearchTransportPortName,
		ContainerPort: api.ElasticsearchTransportPort,
		Protocol:      core.ProtocolTCP,
	}
)

func (es *Elasticsearch) ensureStatefulSet(
	esNode *api.ElasticsearchNode,
	stsName string,
	labels map[string]string,
	replicas *int32,
	nodeRole string,
	envList []core.EnvVar,
	initEnvList []core.EnvVar,
) (kutil.VerbType, error) {

	if esNode == nil {
		return kutil.VerbUnchanged, errors.New("ElasticsearchNode is empty")
	}

	if err := es.checkStatefulSet(stsName); err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      stsName,
		Namespace: es.db.Namespace,
	}

	owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	// Make a new map "labelSelector", so that it remains
	// unchanged even if the "labels" changes.
	// It contains:
	//	-	kubedb.com/kind: ResourceKindElasticsearch
	//	-	kubedb.com/name: elasticsearch.Name
	//	-	node.role.<master/data/ingest>: set
	labelSelector := es.db.OffshootSelectors()
	labelSelector = core_util.UpsertMap(labelSelector, labels)

	// Node affinity is added to support, multi-regional cluster.
	affinity, err := parseAffinityTemplate(es.db.Spec.PodTemplate.Spec.Affinity.DeepCopy(), nodeRole)
	if err != nil {
		return kutil.VerbUnchanged, errors.Wrap(err, "failed to parse the affinity template")
	}

	// Get default initContainers; i.e. config-merger
	initContainers, err := es.getInitContainers(esNode, initEnvList)
	if err != nil {
		return kutil.VerbUnchanged, errors.Wrap(err, "failed to get initContainers")
	}

	// Add/Overwrite user provided initContainers
	initContainers = core_util.UpsertContainers(initContainers, es.db.Spec.PodTemplate.Spec.InitContainers)

	// Get elasticsearch container.
	// Also get monitoring sidecar if any.
	containers, err := es.getContainers(esNode, nodeRole, envList)
	if err != nil {
		return kutil.VerbUnchanged, errors.Wrap(err, "failed to get containers")
	}

	volumes, pvc, err := es.getVolumes(esNode, nodeRole)
	if err != nil {
		return kutil.VerbUnchanged, errors.Wrap(err, "failed to get volumes")
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(context.TODO(), es.kClient, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = core_util.UpsertMap(labels, es.db.OffshootLabels())
		in.Annotations = es.db.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

		in.Spec.Replicas = replicas
		in.Spec.ServiceName = es.db.GoverningServiceName()

		in.Spec.Selector = &metav1.LabelSelector{MatchLabels: labelSelector}
		in.Spec.Template.Labels = labelSelector

		in.Spec.Template.Annotations = es.db.Spec.PodTemplate.Annotations

		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, initContainers)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainers(in.Spec.Template.Spec.Containers, containers)

		in.Spec.Template.Spec.NodeSelector = es.db.Spec.PodTemplate.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = affinity
		if es.db.Spec.PodTemplate.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = es.db.Spec.PodTemplate.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = es.db.Spec.PodTemplate.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = es.db.Spec.PodTemplate.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = es.db.Spec.PodTemplate.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = es.db.Spec.PodTemplate.Spec.Priority
		in.Spec.Template.Spec.HostNetwork = es.db.Spec.PodTemplate.Spec.HostNetwork
		in.Spec.Template.Spec.HostPID = es.db.Spec.PodTemplate.Spec.HostPID
		in.Spec.Template.Spec.HostIPC = es.db.Spec.PodTemplate.Spec.HostIPC
		in.Spec.Template.Spec.SecurityContext = es.db.Spec.PodTemplate.Spec.SecurityContext

		if in.Spec.Template.Spec.SecurityContext == nil {
			in.Spec.Template.Spec.SecurityContext = &core.PodSecurityContext{
				FSGroup: pointer.Int64P(1000),
			}
		}

		in.Spec.Template.Spec.ServiceAccountName = es.db.Spec.PodTemplate.Spec.ServiceAccountName

		// Upsert volumeClaimTemplates if any
		if pvc != nil {
			in.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(in.Spec.VolumeClaimTemplates, *pvc)
		}

		// No need to upsert volumes
		// Everytime the volume list is generated from the YAML file,
		// so it will contain all required volumes. As there is no support for user provided volume for now,
		// we don't need to use upsert here.
		in.Spec.Template.Spec.Volumes = volumes

		// Statefulset update strategy is set default to "OnDelete".
		// Any kind of modification on Elasticsearch will be performed via ElasticsearchModificationRequest CRD.
		// If user update the Elasticsearch object without ElasticsearchModificationRequest,
		// user will have delete the pods manually to encounter the changes.
		in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
			Type: apps.OnDeleteStatefulSetStrategyType,
		}

		return in
	}, metav1.PatchOptions{})

	if err != nil {
		return kutil.VerbUnchanged, errors.Wrap(err, "failed to create or patch statefulset")
	}

	// ensure pdb
	if esNode.MaxUnavailable != nil {
		if err := es.createPodDisruptionBudget(statefulSet, esNode.MaxUnavailable); err != nil {
			return vt, errors.Wrap(err, "failed to create PodDisruptionBudget")
		}
	}

	return vt, nil
}

func (es *Elasticsearch) getVolumes(esNode *api.ElasticsearchNode, nodeRole string) ([]core.Volume, *core.PersistentVolumeClaim, error) {
	if esNode == nil {
		return nil, nil, errors.New("elasticsearchNode is empty")
	}

	var volumes []core.Volume
	var pvc *core.PersistentVolumeClaim

	// Upsert Volume for config directory
	volumes = core_util.UpsertVolume(volumes, core.Volume{
		Name: "esconfig",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	})

	// Default configuration files, will be stored in a temporary directory.
	// i.e. "/elasticsearch/temp-config"
	secretName := es.db.ConfigSecretName()
	volumes = core_util.UpsertVolume(volumes, core.Volume{
		Name: "temp-esconfig",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	})

	// Default security-config files will be merged
	// with user provided security-config files(if any) from "config-merger" init container
	// and later those files will be mounted on this shared volume so that elaticsearch container
	// can use them.
	if !es.db.Spec.DisableSecurity {
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name: "security-config",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		})
	}

	// Upsert Volume for user provided custom configuration.
	// These configuration will be merged to default config yaml (ie. elasticsearch.yaml)
	// from config-merger initContainer.
	if es.db.Spec.ConfigSecret != nil {
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name: "custom-config",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: es.db.Spec.ConfigSecret.Name,
				},
			},
		})
	}

	// Upsert Volume for data directory.
	// If storageType is "Ephemeral", add volume of "EmptyDir" type.
	// The storageType is default to "Durable".
	if es.db.Spec.StorageType == api.StorageTypeEphemeral {
		ed := core.EmptyDirVolumeSource{}
		if esNode.Storage != nil {
			if sz, found := esNode.Storage.Resources.Requests[core.ResourceStorage]; found {
				ed.SizeLimit = &sz
			}
		}
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name: "data",
			VolumeSource: core.VolumeSource{
				EmptyDir: &ed,
			},
		})
	} else {
		if len(esNode.Storage.AccessModes) == 0 {
			esNode.Storage.AccessModes = []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			}
			klog.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, esNode.Storage)
		}

		pvc = &core.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "data",
			},
			Spec: *esNode.Storage,
		}
		if esNode.Storage.StorageClassName != nil {
			pvc.Annotations = map[string]string{
				"volume.beta.kubernetes.io/storage-class": *esNode.Storage.StorageClassName,
			}
		}
	}

	// Upsert Volume for certificates
	if es.db.Spec.TLS == nil && !es.db.Spec.DisableSecurity {
		return nil, nil, errors.New("Certificate secret is missing")
	}
	if !es.db.Spec.DisableSecurity {
		// transport layer is always secured
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name: es.db.CertSecretVolumeName(api.ElasticsearchTransportCert),
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: es.db.GetCertSecretName(api.ElasticsearchTransportCert),
					Items: []core.KeyToPath{
						{
							Key:  certlib.CACert,
							Path: certlib.CACert,
						},
						{
							Key:  certlib.TLSCert,
							Path: certlib.TLSCert,
						},
						{
							Key:  certlib.TLSKey,
							Path: certlib.TLSKey,
						},
					},
				},
			},
		})

		// if security is enabled at rest layer
		if es.db.Spec.EnableSSL {
			volumes = core_util.UpsertVolume(volumes, core.Volume{
				Name: es.db.CertSecretVolumeName(api.ElasticsearchHTTPCert),
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: es.db.GetCertSecretName(api.ElasticsearchHTTPCert),
						Items: []core.KeyToPath{
							{
								Key:  certlib.CACert,
								Path: certlib.CACert,
							},
							{
								Key:  certlib.TLSCert,
								Path: certlib.TLSCert,
							},
							{
								Key:  certlib.TLSKey,
								Path: certlib.TLSKey,
							},
						},
					},
				},
			})
		}

	}

	// Upsert Volume for monitoring sidecar.
	// This volume is used only for ingest nodes.
	if es.db.Spec.Monitor != nil &&
		es.db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus &&
		es.db.Spec.EnableSSL &&
		nodeRole == string(api.ElasticsearchNodeRoleTypeIngest) {
		volumes = core_util.UpsertVolume(volumes, core.Volume{
			Name: es.db.CertSecretVolumeName(api.ElasticsearchMetricsExporterCert),
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: es.db.GetCertSecretName(api.ElasticsearchMetricsExporterCert),
					Items: []core.KeyToPath{
						{
							Key:  certlib.CACert,
							Path: certlib.CACert,
						},
						{
							Key:  certlib.TLSCert,
							Path: certlib.TLSCert,
						},
						{
							Key:  certlib.TLSKey,
							Path: certlib.TLSKey,
						},
					},
				},
			},
		})
	}

	return volumes, pvc, nil
}

func (es *Elasticsearch) getContainers(esNode *api.ElasticsearchNode, nodeRole string, envList []core.EnvVar) ([]core.Container, error) {
	if esNode == nil {
		return nil, errors.New("ElasticsearchNode is empty")
	}

	// Add volumeMounts for elasticsearch container
	// 		- data directory
	volumeMount := []core.VolumeMount{
		{
			Name:      "data",
			MountPath: api.ElasticsearchDataDir,
		},
		{
			Name:      "esconfig",
			MountPath: api.ElasticsearchConfigDir,
		},
	}

	// Mount user provided custom configuration to "elasticsearch_config_directory/custom_config"
	if es.db.Spec.ConfigSecret != nil {
		volumeMount = core_util.UpsertVolumeMount(volumeMount, core.VolumeMount{
			Name:      "custom-config",
			MountPath: path.Join(api.ElasticsearchConfigDir, api.DBCustomConfigName),
		})
	}

	// Add volumeMounts for elasticsearch container
	// 		- security config directory
	//		- certificates directory
	if !es.db.Spec.DisableSecurity {
		volumeMount = core_util.UpsertVolumeMount(volumeMount, []core.VolumeMount{
			{
				Name:      "security-config",
				MountPath: fmt.Sprintf(api.ElasticsearchSearchGuardSecurityConfigDir, string(es.esVersion.Spec.Version[0])),
			},
			{
				Name:      es.db.CertSecretVolumeName(api.ElasticsearchTransportCert),
				MountPath: es.db.CertSecretVolumeMountPath(api.ElasticsearchConfigDir, api.ElasticsearchTransportCert),
			},
		}...)

		if es.db.Spec.EnableSSL {
			volumeMount = core_util.UpsertVolumeMount(volumeMount, core.VolumeMount{
				Name:      es.db.CertSecretVolumeName(api.ElasticsearchHTTPCert),
				MountPath: es.db.CertSecretVolumeMountPath(api.ElasticsearchConfigDir, api.ElasticsearchHTTPCert),
			})
		}
	}

	containers := []core.Container{
		{
			Name:            api.ElasticsearchContainerName,
			Image:           es.esVersion.Spec.DB.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Env:             envList,

			// The restPort is only necessary for Ingest nodes.
			// But it is set for all type of nodes, so that our controller can
			// communicate with each nodes specifically.
			// The DBA controller uses the restPort to check health of a node.
			Ports:           []core.ContainerPort{defaultRestPort, defaultTransportPort},
			SecurityContext: es.db.Spec.PodTemplate.Spec.ContainerSecurityContext,
			Resources:       esNode.Resources,
			VolumeMounts:    volumeMount,
			LivenessProbe:   es.db.Spec.PodTemplate.Spec.LivenessProbe,
			ReadinessProbe:  es.db.Spec.PodTemplate.Spec.ReadinessProbe,
			Lifecycle:       es.db.Spec.PodTemplate.Spec.Lifecycle,
		},
	}

	// upsert metrics exporter sidecar for monitoring purpose.
	// add monitoring sidecar only for ingest nodes.
	var err error
	if es.db.Spec.Monitor != nil && nodeRole == string(api.ElasticsearchNodeRoleTypeIngest) {
		containers, err = es.upsertMonitoringContainer(containers)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get monitoring container")
		}
	}

	return containers, nil
}

func (es *Elasticsearch) getInitContainers(esNode *api.ElasticsearchNode, envList []core.EnvVar) ([]core.Container, error) {
	if esNode == nil {
		return nil, errors.New("ElasticsearchNode is empty")
	}

	var privileged bool
	var command string
	var initContainers []core.Container
	if es.db.Spec.KernelSettings != nil {
		if es.db.Spec.KernelSettings.Privileged {
			privileged = true
		}
		if es.db.Spec.KernelSettings.Sysctls != nil {
			// Use separator `;` for sh
			command = kernel.GetSysctlCommandString(es.db.Spec.KernelSettings.Sysctls, ';')
		}
	}

	// If commands exist, add sysctl init container.
	// Otherwise skip it.
	if command != "" {
		initContainers = []core.Container{
			{
				Name:            api.ElasticsearchInitSysctlContainerName,
				Image:           es.esVersion.Spec.InitContainer.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Command:         []string{"sh", "-c", command},
				SecurityContext: &core.SecurityContext{
					Privileged: pointer.BoolP(privileged),
				},
			},
		}
	}

	initContainers = es.upsertConfigMergerInitContainer(initContainers, envList)
	return initContainers, nil
}

func (es *Elasticsearch) upsertConfigMergerInitContainer(initCon []core.Container, envList []core.EnvVar) []core.Container {
	volumeMounts := []core.VolumeMount{
		{
			Name:      "temp-esconfig",
			MountPath: api.ElasticsearchTempConfigDir,
		},
		{
			Name:      "data",
			MountPath: api.ElasticsearchDataDir,
		},
		{
			Name:      "esconfig",
			MountPath: api.ElasticsearchConfigDir,
		},
	}

	if !es.db.Spec.DisableSecurity {
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts, core.VolumeMount{
			Name:      "security-config",
			MountPath: fmt.Sprintf(api.ElasticsearchSearchGuardSecurityConfigDir, string(es.esVersion.Spec.Version[0])),
		})
	}

	// mount path for custom configuration
	if es.db.Spec.ConfigSecret != nil {
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts, core.VolumeMount{
			Name:      "custom-config",
			MountPath: api.ElasticsearchCustomConfigDir,
		})
	}

	configMerger := core.Container{
		Name:            api.ElasticsearchInitConfigMergerContainerName,
		Image:           es.esVersion.Spec.InitContainer.YQImage,
		ImagePullPolicy: core.PullIfNotPresent,
		Env:             envList,
		VolumeMounts:    volumeMounts,
	}

	return append(initCon, configMerger)
}

func (es *Elasticsearch) checkStatefulSet(sName string) error {
	elasticsearchName := es.db.OffshootName()

	// StatefulSet for Elasticsearch database
	statefulSet, err := es.kClient.AppsV1().StatefulSets(es.db.Namespace).Get(context.TODO(), sName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[meta_util.NameLabelKey] != es.db.ResourceFQN() ||
		statefulSet.Labels[meta_util.InstanceLabelKey] != elasticsearchName {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, es.db.Namespace, sName)
	}

	return nil
}

func (es *Elasticsearch) upsertContainerEnv(envList []core.EnvVar) []core.EnvVar {

	envList = core_util.UpsertEnvVars(envList, []core.EnvVar{
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
			Value: es.db.Name,
		},
		{
			Name:  "network.host",
			Value: "0.0.0.0",
		},
	}...)

	if strings.HasPrefix(es.esVersion.Spec.Version, "7.") {
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "discovery.seed_hosts",
			Value: es.db.MasterDiscoveryServiceName(),
		})
	} else {
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "discovery.zen.ping.unicast.hosts",
			Value: es.db.MasterDiscoveryServiceName(),
		})
	}

	return envList
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
