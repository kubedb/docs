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
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/fatih/structs"
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
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type workloadOptions struct {
	// App level options
	stsName string

	// db container options
	conatainerName string
	image          string
	cmd            []string // cmd of `percona-xtradb` container
	args           []string // args of `percona-xtradb` container
	ports          []core.ContainerPort
	envList        []core.EnvVar // envList of `percona-xtradb` container
	volumeMount    []core.VolumeMount
	configSecret   *core.LocalObjectReference

	// monitor container
	monitorContainer *core.Container

	// pod Template level options
	replicas       *int32
	gvrSvcName     string
	podTemplate    *ofst.PodTemplateSpec
	pvcSpec        *core.PersistentVolumeClaimSpec
	initContainers []core.Container
	volume         []core.Volume // volumes to mount on stsPodTemplate
}

func (c *Controller) ensurePerconaXtraDB(db *api.PerconaXtraDB) (kutil.VerbType, error) {
	pxVersion, err := c.DBClient.CatalogV1alpha1().PerconaXtraDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	initContainers := []core.Container{
		{
			Name:            "remove-lost-found",
			Image:           pxVersion.Spec.InitContainer.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Command: []string{
				"rm",
				"-rf",
				"/var/lib/mysql/lost+found",
			},
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "data",
					MountPath: api.PerconaXtraDBDataMountPath,
				},
			},
			Resources: db.Spec.PodTemplate.Spec.Resources,
		},
	}

	var cmds, args []string
	ports := []core.ContainerPort{
		{
			Name:          api.MySQLDatabasePortName,
			ContainerPort: api.MySQLDatabasePort,
			Protocol:      core.ProtocolTCP,
		},
	}
	if db.IsCluster() {
		cmds = []string{
			"peer-finder",
		}
		userProvidedArgs := strings.Join(db.Spec.PodTemplate.Spec.Args, " ")
		args = []string{
			fmt.Sprintf("-service=%s", db.GoverningServiceName()),
			fmt.Sprintf("-on-start=/on-start.sh %s", userProvidedArgs),
		}
		ports = append(ports, []core.ContainerPort{
			{
				Name:          "sst",
				ContainerPort: 4567,
			},
			{
				Name:          "replication",
				ContainerPort: 4568,
			},
		}...)
	}

	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

	if !db.IsCluster() && db.Spec.Init != nil && db.Spec.Init.Script != nil {
		volumes = append(volumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: db.Spec.Init.Script.VolumeSource,
		})
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      "initial-script",
			MountPath: api.PerconaXtraDBInitDBMountPath,
		})
	}
	db.Spec.PodTemplate.Spec.ServiceAccountName = db.OffshootName()

	envList := []core.EnvVar{}
	if db.IsCluster() {
		envList = append(envList, core.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: db.OffshootName(),
		})
	}

	var monitorContainer core.Container
	if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
		monitorContainer = core.Container{
			Name: "exporter",
			Command: []string{
				"/bin/sh",
			},
			Args: []string{
				"-c",
				// DATA_SOURCE_NAME=user:password@tcp(localhost:5555)/dbname
				// ref: https://github.com/prometheus/mysqld_exporter#setting-the-mysql-servers-data-source-name
				fmt.Sprintf(`export DATA_SOURCE_NAME="${MYSQL_ROOT_USERNAME:-}:${MYSQL_ROOT_PASSWORD:-}@(127.0.0.1:3306)/"
						/bin/mysqld_exporter --web.listen-address=:%v --web.telemetry-path=%v %v`,
					db.Spec.Monitor.Prometheus.Exporter.Port, db.StatsService().Path(), strings.Join(db.Spec.Monitor.Prometheus.Exporter.Args, " ")),
			},
			Image: pxVersion.Spec.Exporter.Image,
			Ports: []core.ContainerPort{
				{
					Name:          mona.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: db.Spec.Monitor.Prometheus.Exporter.Port,
				},
			},
			Env:             db.Spec.Monitor.Prometheus.Exporter.Env,
			Resources:       db.Spec.Monitor.Prometheus.Exporter.Resources,
			SecurityContext: db.Spec.Monitor.Prometheus.Exporter.SecurityContext,
		}
	}

	opts := workloadOptions{
		stsName:          db.OffshootName(),
		conatainerName:   api.ResourceSingularPerconaXtraDB,
		image:            pxVersion.Spec.DB.Image,
		args:             args,
		cmd:              cmds,
		ports:            ports,
		envList:          envList,
		initContainers:   initContainers,
		gvrSvcName:       db.GoverningServiceName(),
		podTemplate:      &db.Spec.PodTemplate,
		configSecret:     db.Spec.ConfigSecret,
		pvcSpec:          db.Spec.Storage,
		replicas:         db.Spec.Replicas,
		volume:           volumes,
		volumeMount:      volumeMounts,
		monitorContainer: &monitorContainer,
	}

	return c.ensureStatefulSet(db, opts)
}

func (c *Controller) checkStatefulSet(db *api.PerconaXtraDB, stsName string) error {
	// StatefulSet for PerconaXtraDB database
	statefulSet, err := c.Client.AppsV1().StatefulSets(db.Namespace).Get(context.TODO(), stsName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[meta_util.NameLabelKey] != db.ResourceFQN() ||
		statefulSet.Labels[meta_util.InstanceLabelKey] != db.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, db.Namespace, stsName)
	}

	return nil
}

func upsertCustomConfig(
	template core.PodTemplateSpec, configSecret *core.LocalObjectReference, replicas int32,
) core.PodTemplateSpec {
	for i, container := range template.Spec.Containers {
		if container.Name == api.ResourceSingularPerconaXtraDB {
			configVolumeMount := core.VolumeMount{
				Name:      "custom-config",
				MountPath: api.PerconaXtraDBCustomConfigMountPath,
			}
			if replicas > 1 {
				configVolumeMount.MountPath = api.PerconaXtraDBClusterCustomConfigMountPath
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			template.Spec.Containers[i].VolumeMounts = volumeMounts

			configVolume := core.Volume{
				Name: "custom-config",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: configSecret.Name,
					},
				},
			}

			volumes := template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, configVolume)
			template.Spec.Volumes = volumes
			break
		}
	}

	return template
}

func (c *Controller) ensureStatefulSet(db *api.PerconaXtraDB, opts workloadOptions) (kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkStatefulSet(db, opts.stsName); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for PerconaXtraDB database
	statefulSetMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))

	readinessProbe := pt.Spec.ReadinessProbe
	if readinessProbe != nil && structs.IsZero(*readinessProbe) {
		readinessProbe = nil
	}
	livenessProbe := pt.Spec.LivenessProbe
	if livenessProbe != nil && structs.IsZero(*livenessProbe) {
		livenessProbe = nil
	}

	if readinessProbe != nil {
		readinessProbe.InitialDelaySeconds = 60
		readinessProbe.PeriodSeconds = 10
		readinessProbe.TimeoutSeconds = 50
		readinessProbe.SuccessThreshold = 1
		readinessProbe.FailureThreshold = 3
	}
	if livenessProbe != nil {
		livenessProbe.InitialDelaySeconds = 60
		livenessProbe.PeriodSeconds = 10
		livenessProbe.TimeoutSeconds = 50
		livenessProbe.SuccessThreshold = 1
		livenessProbe.FailureThreshold = 3
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = db.PodControllerLabels()
			in.Annotations = pt.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = opts.replicas
			in.Spec.ServiceName = opts.gvrSvcName
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: db.OffshootSelectors(),
			}
			in.Spec.Template.Labels = db.PodLabels()
			in.Spec.Template.Annotations = pt.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				pt.Spec.InitContainers,
			)
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				in.Spec.Template.Spec.Containers,
				core.Container{
					Name:            opts.conatainerName,
					Image:           opts.image,
					ImagePullPolicy: core.PullIfNotPresent,
					Command:         opts.cmd,
					Args:            opts.args,
					Ports:           opts.ports,
					Env:             core_util.UpsertEnvVars(opts.envList, pt.Spec.Env...),
					Resources:       pt.Spec.Resources,
					SecurityContext: pt.Spec.ContainerSecurityContext,
					Lifecycle:       pt.Spec.Lifecycle,
					LivenessProbe:   livenessProbe,
					ReadinessProbe:  readinessProbe,
					VolumeMounts:    opts.volumeMount,
				})

			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				opts.initContainers,
			)

			if opts.monitorContainer != nil && db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
					in.Spec.Template.Spec.Containers, *opts.monitorContainer)
			}

			in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volume...)

			in = upsertEnv(in, db)
			in = upsertDataVolume(in, db)

			if opts.configSecret != nil {
				in.Spec.Template = upsertCustomConfig(in.Spec.Template, opts.configSecret, pointer.Int32(db.Spec.Replicas))
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
			in.Spec.Template.Spec.HostNetwork = pt.Spec.HostNetwork
			in.Spec.Template.Spec.HostPID = pt.Spec.HostPID
			in.Spec.Template.Spec.HostIPC = pt.Spec.HostIPC
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
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt, db.Namespace, opts.stsName,
		)
	}

	return vt, nil
}

func upsertDataVolume(statefulSet *apps.StatefulSet, db *api.PerconaXtraDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPerconaXtraDB {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: api.PerconaXtraDBDataMountPath,
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := db.Spec.Storage
			if db.Spec.StorageType == api.StorageTypeEphemeral {
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
					klog.Infof(`Using "%v" as AccessModes in .spec.storage`, core.ReadWriteOnce)
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
			break
		}
	}
	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertEnv(statefulSet *apps.StatefulSet, db *api.PerconaXtraDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPerconaXtraDB || container.Name == "exporter" {
			envs := []core.EnvVar{
				{
					Name: "MYSQL_ROOT_PASSWORD",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: db.Spec.AuthSecret.Name,
							},
							Key: core.BasicAuthPasswordKey,
						},
					},
				},
				{
					Name: "MYSQL_ROOT_USERNAME",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: db.Spec.AuthSecret.Name,
							},
							Key: core.BasicAuthUsernameKey,
						},
					},
				},
			}

			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return statefulSet
}

func (c *Controller) checkStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	err := core_util.WaitUntilPodRunningBySelector(
		context.TODO(),
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(pointer.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}
