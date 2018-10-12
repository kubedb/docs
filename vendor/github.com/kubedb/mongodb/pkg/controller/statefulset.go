package controller

import (
	"fmt"
	"strconv"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
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

func (c *Controller) ensureStatefulSet(mongodb *api.MongoDB) (kutil.VerbType, error) {
	if err := c.checkStatefulSet(mongodb); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for MongoDB database
	statefulSet, vt, err := c.createStatefulSet(mongodb)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			c.recorder.Eventf(
				mongodb,
				core.EventTypeWarning,
				eventer.EventReasonFailedToStart,
				`Failed to CreateOrPatch StatefulSet. Reason: %v`,
				err,
			)
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			mongodb,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkStatefulSet(mongodb *api.MongoDB) error {
	// SatatefulSet for MongoDB database
	statefulSet, err := c.Client.AppsV1().StatefulSets(mongodb.Namespace).Get(mongodb.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB ||
		statefulSet.Labels[api.LabelDatabaseName] != mongodb.Name {
		return fmt.Errorf(`intended statefulSet "%v" already exists`, mongodb.OffshootName())
	}

	return nil
}

func (c *Controller) createStatefulSet(mongodb *api.MongoDB) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      mongodb.OffshootName(),
		Namespace: mongodb.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb)
	if rerr != nil {
		return nil, kutil.VerbUnchanged, rerr
	}

	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = mongodb.OffshootLabels()
		in.Annotations = mongodb.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)

		in.Spec.Replicas = mongodb.Spec.Replicas
		in.Spec.ServiceName = c.GoverningService
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: mongodb.OffshootSelectors(),
		}
		in.Spec.Template.Labels = mongodb.OffshootSelectors()
		in.Spec.Template.Annotations = mongodb.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, mongodb.Spec.PodTemplate.Spec.InitContainers)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
			in.Spec.Template.Spec.Containers,
			core.Container{
				Name:            api.ResourceSingularMongoDB,
				Image:           mongodbVersion.Spec.DB.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Args: meta_util.UpsertArgumentList([]string{
					"--dbpath=" + dataDirectoryPath,
					"--auth",
					"--bind_ip=0.0.0.0",
					"--port=" + strconv.Itoa(MongoDBPort),
				}, mongodb.Spec.PodTemplate.Spec.Args),
				Ports: []core.ContainerPort{
					{
						Name:          "db",
						ContainerPort: MongoDBPort,
						Protocol:      core.ProtocolTCP,
					},
				},
				Resources:      mongodb.Spec.PodTemplate.Spec.Resources,
				LivenessProbe:  mongodb.Spec.PodTemplate.Spec.LivenessProbe,
				ReadinessProbe: mongodb.Spec.PodTemplate.Spec.ReadinessProbe,
				Lifecycle:      mongodb.Spec.PodTemplate.Spec.Lifecycle,
			})

		in = c.upsertInstallInitContainer(in, mongodb, mongodbVersion)
		if mongodb.Spec.ReplicaSet != nil {
			in = c.upsertRSInitContainer(in, mongodb, mongodbVersion)
			in = upsertRSArgs(in, mongodb)

		}

		if mongodb.GetMonitoringVendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: []string{
					fmt.Sprintf("--web.listen-address=:%d", mongodb.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--web.metrics-path=%v", mongodb.StatsService().Path()),
					"--mongodb.uri=mongodb://$(MONGO_INITDB_ROOT_USERNAME):$(MONGO_INITDB_ROOT_PASSWORD)@127.0.0.1:27017",
				},
				Image: mongodbVersion.Spec.Exporter.Image,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: mongodb.Spec.Monitor.Prometheus.Port,
					},
				},
			})
		}
		// Set Admin Secret as MONGO_INITDB_ROOT_PASSWORD env variable
		in = upsertEnv(in, mongodb)
		in = upsertUserEnv(in, mongodb)
		in = upsertDataVolume(in, mongodb)
		in = addContainerProbe(in, mongodb)

		if mongodb.Spec.ConfigSource != nil {
			in = c.upsertConfigSourceVolume(in, mongodb)
		}

		if mongodb.Spec.Init != nil && mongodb.Spec.Init.ScriptSource != nil {
			in = upsertInitScript(in, mongodb.Spec.Init.ScriptSource.VolumeSource)
		}

		in.Spec.Template.Spec.NodeSelector = mongodb.Spec.PodTemplate.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = mongodb.Spec.PodTemplate.Spec.Affinity
		if mongodb.Spec.PodTemplate.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = mongodb.Spec.PodTemplate.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = mongodb.Spec.PodTemplate.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = mongodb.Spec.PodTemplate.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = mongodb.Spec.PodTemplate.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = mongodb.Spec.PodTemplate.Spec.Priority
		in.Spec.Template.Spec.SecurityContext = mongodb.Spec.PodTemplate.Spec.SecurityContext

		in.Spec.UpdateStrategy = mongodb.Spec.UpdateStrategy
		return in
	})
}

func addContainerProbe(statefulSet *apps.StatefulSet, mongodb *api.MongoDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB {
			cmd := []string{
				"mongo",
				"--eval",
				"db.adminCommand('ping')",
			}
			statefulSet.Spec.Template.Spec.Containers[i].LivenessProbe = &core.Probe{
				Handler: core.Handler{
					Exec: &core.ExecAction{
						Command: cmd,
					},
				},
				FailureThreshold: 3,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				TimeoutSeconds:   5,
			}
			statefulSet.Spec.Template.Spec.Containers[i].ReadinessProbe = &core.Probe{
				Handler: core.Handler{
					Exec: &core.ExecAction{
						Command: cmd,
					},
				},
				FailureThreshold: 3,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				TimeoutSeconds:   1,
			}
		}
	}
	return statefulSet
}

// Init container for both ReplicaSet and Standalone instances
func (c *Controller) upsertInstallInitContainer(statefulSet *apps.StatefulSet, mongodb *api.MongoDB, mongodbVersion *catalog.MongoDBVersion) *apps.StatefulSet {
	installContainer := core.Container{
		Name:            InitInstallContainerName,
		Image:           "busybox",
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
			fi`,
		},
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
		Resources: mongodb.Spec.PodTemplate.Spec.Resources,
	}
	if mongodb.Spec.ReplicaSet != nil {
		installContainer.VolumeMounts = core_util.UpsertVolumeMount(installContainer.VolumeMounts, core.VolumeMount{
			Name:      initialKeyDirectoryName,
			MountPath: initialKeyDirectoryPath,
		})
	}

	initContainers := statefulSet.Spec.Template.Spec.InitContainers
	statefulSet.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(initContainers, installContainer)

	initVolumes := core.Volume{
		Name: workDirectoryName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, initVolumes)

	return statefulSet
}

func upsertDataVolume(statefulSet *apps.StatefulSet, mongodb *api.MongoDB) *apps.StatefulSet {
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
			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, volumes)

			pvcSpec := mongodb.Spec.Storage
			if mongodb.Spec.StorageType == api.StorageTypeEphemeral {
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
				statefulSet.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(statefulSet.Spec.VolumeClaimTemplates, claim)
			}
			break
		}
	}
	return statefulSet
}

func upsertEnv(statefulSet *apps.StatefulSet, mongodb *api.MongoDB) *apps.StatefulSet {
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
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB || container.Name == "exporter" {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
		}
	}
	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, mongodb *api.MongoDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, mongodb.Spec.PodTemplate.Spec.Env...)
			break
		}
	}
	for i, container := range statefulSet.Spec.Template.Spec.InitContainers {
		if container.Name == InitBootstrapContainerName {
			statefulSet.Spec.Template.Spec.InitContainers[i].Env = core_util.UpsertEnvVars(container.Env, mongodb.Spec.PodTemplate.Spec.Env...)
			break
		}
	}
	return statefulSet
}

func upsertInitScript(statefulSet *apps.StatefulSet, script core.VolumeSource) *apps.StatefulSet {
	volume := core.Volume{
		Name:         "initial-script",
		VolumeSource: script,
	}

	volumeMount := core.VolumeMount{
		Name:      "initial-script",
		MountPath: "/docker-entrypoint-initdb.d",
	}

	statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
		statefulSet.Spec.Template.Spec.Volumes,
		volume,
	)

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB {
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(
				container.VolumeMounts,
				volumeMount,
			)
			break
		}
	}

	for i, container := range statefulSet.Spec.Template.Spec.InitContainers {
		if container.Name == InitBootstrapContainerName {
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = core_util.UpsertVolumeMount(
				container.VolumeMounts,
				volumeMount,
			)
			break
		}
	}
	return statefulSet
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
