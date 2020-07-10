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
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/fatih/structs"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type workloadOptions struct {
	// App level options
	stsName   string
	labels    map[string]string
	selectors map[string]string

	// db container options
	conatainerName string
	image          string
	cmd            []string // cmd of `percona-xtradb` container
	args           []string // args of `percona-xtradb` container
	ports          []core.ContainerPort
	envList        []core.EnvVar // envList of `percona-xtradb` container
	volumeMount    []core.VolumeMount
	configSource   *core.VolumeSource

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

func (c *Controller) ensurePerconaXtraDB(px *api.PerconaXtraDB) (kutil.VerbType, error) {
	pxVersion, err := c.ExtClient.CatalogV1alpha1().PerconaXtraDBVersions().Get(context.TODO(), string(px.Spec.Version), metav1.GetOptions{})
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
			Resources: px.Spec.PodTemplate.Spec.Resources,
		},
	}

	var cmds, args []string
	var ports = []core.ContainerPort{
		{
			Name:          "mysql",
			ContainerPort: api.MySQLNodePort,
			Protocol:      core.ProtocolTCP,
		},
	}
	if px.IsCluster() {
		cmds = []string{
			"peer-finder",
		}
		userProvidedArgs := strings.Join(px.Spec.PodTemplate.Spec.Args, " ")
		args = []string{
			fmt.Sprintf("-service=%s", c.GoverningService),
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

	if !px.IsCluster() && px.Spec.Init != nil && px.Spec.Init.ScriptSource != nil {
		volumes = append(volumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: px.Spec.Init.ScriptSource.VolumeSource,
		})
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      "initial-script",
			MountPath: api.PerconaXtraDBInitDBMountPath,
		})
	}
	px.Spec.PodTemplate.Spec.ServiceAccountName = px.OffshootName()

	envList := []core.EnvVar{}
	if px.IsCluster() {
		envList = append(envList, core.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: px.OffshootName(),
		})
	}

	var monitorContainer core.Container
	if px.GetMonitoringVendor() == mona.VendorPrometheus {
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
					px.Spec.Monitor.Prometheus.Port, px.StatsService().Path(), strings.Join(px.Spec.Monitor.Args, " ")),
			},
			Image: pxVersion.Spec.Exporter.Image,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: px.Spec.Monitor.Prometheus.Port,
				},
			},
			Env:             px.Spec.Monitor.Env,
			Resources:       px.Spec.Monitor.Resources,
			SecurityContext: px.Spec.Monitor.SecurityContext,
		}
	}

	opts := workloadOptions{
		stsName:          px.OffshootName(),
		labels:           px.OffshootLabels(),
		selectors:        px.OffshootSelectors(),
		conatainerName:   api.ResourceSingularPerconaXtraDB,
		image:            pxVersion.Spec.DB.Image,
		args:             args,
		cmd:              cmds,
		ports:            ports,
		envList:          envList,
		initContainers:   initContainers,
		gvrSvcName:       c.GoverningService,
		podTemplate:      &px.Spec.PodTemplate,
		configSource:     px.Spec.ConfigSource,
		pvcSpec:          px.Spec.Storage,
		replicas:         px.Spec.Replicas,
		volume:           volumes,
		volumeMount:      volumeMounts,
		monitorContainer: &monitorContainer,
	}

	return c.ensureStatefulSet(px, px.Spec.UpdateStrategy, opts)
}

func (c *Controller) checkStatefulSet(px *api.PerconaXtraDB, stsName string) error {
	// StatefulSet for PerconaXtraDB database
	statefulSet, err := c.Client.AppsV1().StatefulSets(px.Namespace).Get(context.TODO(), stsName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindPerconaXtraDB ||
		statefulSet.Labels[api.LabelDatabaseName] != px.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, px.Namespace, stsName)
	}

	return nil
}

func upsertCustomConfig(
	template core.PodTemplateSpec, configSource *core.VolumeSource, replicas int32) core.PodTemplateSpec {
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
				Name:         "custom-config",
				VolumeSource: *configSource,
			}

			volumes := template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, configVolume)
			template.Spec.Volumes = volumes
			break
		}
	}

	return template
}

func (c *Controller) ensureStatefulSet(
	px *api.PerconaXtraDB,
	updateStrategy apps.StatefulSetUpdateStrategy,
	opts workloadOptions) (kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkStatefulSet(px, opts.stsName); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for PerconaXtraDB database
	statefulSetMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: px.Namespace,
	}

	owner := metav1.NewControllerRef(px, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))

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
					Name:            opts.conatainerName,
					Image:           opts.image,
					ImagePullPolicy: core.PullIfNotPresent,
					Command:         opts.cmd,
					Args:            opts.args,
					Ports:           opts.ports,
					Env:             core_util.UpsertEnvVars(opts.envList, pt.Spec.Env...),
					Resources:       pt.Spec.Resources,
					Lifecycle:       pt.Spec.Lifecycle,
					LivenessProbe:   livenessProbe,
					ReadinessProbe:  readinessProbe,
					VolumeMounts:    opts.volumeMount,
				})

			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				opts.initContainers,
			)

			if opts.monitorContainer != nil && px.GetMonitoringVendor() == mona.VendorPrometheus {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
					in.Spec.Template.Spec.Containers, *opts.monitorContainer)
			}

			in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volume...)

			in = upsertEnv(in, px)
			in = upsertDataVolume(in, px)

			if opts.configSource != nil {
				in.Spec.Template = upsertCustomConfig(in.Spec.Template, opts.configSource, types.Int32(px.Spec.Replicas))
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
			in.Spec.UpdateStrategy = updateStrategy
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
		c.recorder.Eventf(
			px,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt, px.Namespace, opts.stsName,
		)
	}

	return vt, nil
}

func upsertDataVolume(statefulSet *apps.StatefulSet, px *api.PerconaXtraDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPerconaXtraDB {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: api.PerconaXtraDBDataMountPath,
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := px.Spec.Storage
			if px.Spec.StorageType == api.StorageTypeEphemeral {
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
					log.Infof(`Using "%v" as AccessModes in .spec.storage`, core.ReadWriteOnce)
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
func upsertEnv(statefulSet *apps.StatefulSet, px *api.PerconaXtraDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPerconaXtraDB || container.Name == "exporter" {
			envs := []core.EnvVar{
				{
					Name: "MYSQL_ROOT_PASSWORD",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: px.Spec.DatabaseSecret.SecretName,
							},
							Key: api.MySQLPasswordKey,
						},
					},
				},
				{
					Name: "MYSQL_ROOT_USERNAME",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: px.Spec.DatabaseSecret.SecretName,
							},
							Key: api.MySQLUserKey,
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
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}
