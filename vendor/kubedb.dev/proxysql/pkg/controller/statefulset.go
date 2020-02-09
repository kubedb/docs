/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/types"
	"github.com/fatih/structs"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type db interface {
	PeerName(i int) string
	GetDatabaseSecretName() string
}

var _ db = api.PerconaXtraDB{}
var _ db = api.MySQL{}

type workloadOptions struct {
	// App level options
	stsName   string
	labels    map[string]string
	selectors map[string]string

	// db container options
	conatainerName string
	image          string
	cmd            []string // cmd of `proxysql` container
	args           []string // args of `proxysql` container
	ports          []core.ContainerPort
	envList        []core.EnvVar // envList of `proxysql` container
	volumeMount    []core.VolumeMount
	configSource   *core.VolumeSource

	// monitor container
	monitorContainer *core.Container

	// pod Template level options
	replicas       *int32
	gvrSvcName     string
	podTemplate    *ofst.PodTemplateSpec
	initContainers []core.Container
	volume         []core.Volume // volumes to mount on stsPodTemplate
}

func (c *Controller) ensureProxySQLNode(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	proxysqlVersion, err := c.ExtClient.CatalogV1alpha1().ProxySQLVersions().Get(string(proxysql.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	var ports = []core.ContainerPort{
		{
			Name:          "mysql",
			ContainerPort: api.ProxySQLMySQLNodePort,
			Protocol:      core.ProtocolTCP,
		},
		{
			Name:          api.ProxySQLAdminPortName,
			ContainerPort: api.ProxySQLAdminPort,
			Protocol:      core.ProtocolTCP,
		},
	}

	proxysql.Spec.PodTemplate.Spec.ServiceAccountName = proxysql.OffshootName()

	var backendDB db
	backend := proxysql.Spec.Backend
	gk := schema.GroupKind{Group: *backend.Ref.APIGroup, Kind: backend.Ref.Kind}

	switch gk {
	case api.Kind(api.ResourceKindPerconaXtraDB):
		backendDB, err = c.ExtClient.KubedbV1alpha1().PerconaXtraDBs(proxysql.Namespace).Get(backend.Ref.Name, metav1.GetOptions{})
	case api.Kind(api.ResourceKindMySQL):
		backendDB, err = c.ExtClient.KubedbV1alpha1().MySQLs(proxysql.Namespace).Get(backend.Ref.Name, metav1.GetOptions{})
	// TODO: add other cases for MySQL and MariaDB when they will be configured
	default:
		return kutil.VerbUnchanged, fmt.Errorf("unknown group kind '%v' is specified", gk.String())
	}
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	peers := make([]string, 0, *backend.Replicas)
	for i := 0; i < int(*backend.Replicas); i += 1 {
		peers = append(peers, backendDB.PeerName(i))
	}

	var monitorContainer *core.Container
	if proxysql.GetMonitoringVendor() == mona.VendorPrometheus {
		monitorContainer = &core.Container{
			Name: "exporter",
			Command: []string{
				"/bin/sh",
			},
			Args: []string{
				"-c",
				// DATA_SOURCE_NAME=user:password@tcp(localhost:5555)/dbname
				// ref: https://github.com/go-sql-driver/mysql#dsn-data-source-name
				fmt.Sprintf(`export DATA_SOURCE_NAME="admin:admin@tcp(127.0.0.1:6032)/"
						/bin/proxysql_exporter --web.listen-address=:%v --web.telemetry-path=%v %v`,
					proxysql.Spec.Monitor.Prometheus.Port, proxysql.StatsService().Path(), strings.Join(proxysql.Spec.Monitor.Args, " ")),
			},
			Image: proxysqlVersion.Spec.Exporter.Image,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: proxysql.Spec.Monitor.Prometheus.Port,
				},
			},
			Env:             proxysql.Spec.Monitor.Env,
			Resources:       proxysql.Spec.Monitor.Resources,
			SecurityContext: proxysql.Spec.Monitor.SecurityContext,
		}
	}

	envList := []core.EnvVar{
		{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: backendDB.GetDatabaseSecretName(),
					},
					Key: api.MySQLPasswordKey,
				},
			},
		},
		{
			Name:  "PEERS",
			Value: strings.Join(peers, ","),
		},
		{
			Name:  "LOAD_BALANCE_MODE",
			Value: string(*proxysql.Spec.Mode),
		},
	}

	opts := workloadOptions{
		stsName:          proxysql.OffshootName(),
		labels:           proxysql.OffshootLabels(),
		selectors:        proxysql.OffshootSelectors(),
		conatainerName:   api.ResourceSingularProxySQL,
		image:            proxysqlVersion.Spec.Proxysql.Image,
		args:             nil,
		cmd:              nil,
		ports:            ports,
		envList:          envList,
		initContainers:   nil,
		gvrSvcName:       c.GoverningService,
		podTemplate:      &proxysql.Spec.PodTemplate,
		configSource:     proxysql.Spec.ConfigSource,
		replicas:         proxysql.Spec.Replicas,
		volume:           nil,
		volumeMount:      nil,
		monitorContainer: monitorContainer,
	}

	return c.ensureStatefulSet(proxysql, proxysql.Spec.UpdateStrategy, opts)
}

func (c *Controller) checkStatefulSet(proxysql *api.ProxySQL, stsName string) error {
	// StatefulSet for ProxySQL database
	statefulSet, err := c.Client.AppsV1().StatefulSets(proxysql.Namespace).Get(stsName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindProxySQL ||
		statefulSet.Labels[api.LabelProxySQLName] != proxysql.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, proxysql.Namespace, stsName)
	}

	return nil
}

func upsertCustomConfig(template core.PodTemplateSpec, configSource *core.VolumeSource) core.PodTemplateSpec {
	for i, container := range template.Spec.Containers {
		if container.Name == api.ResourceSingularProxySQL {
			configVolumeMount := core.VolumeMount{
				Name:      "custom-config",
				MountPath: api.ProxySQLCustomConfigMountPath,
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
	proxysql *api.ProxySQL,
	updateStrategy apps.StatefulSetUpdateStrategy,
	opts workloadOptions) (kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkStatefulSet(proxysql, opts.stsName); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for ProxySQL database
	statefulSetMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: proxysql.Namespace,
	}

	owner := metav1.NewControllerRef(proxysql, api.SchemeGroupVersion.WithKind(api.ResourceKindProxySQL))

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

		if opts.monitorContainer != nil && proxysql.GetMonitoringVendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				in.Spec.Template.Spec.Containers, *opts.monitorContainer)
		}

		// Set proxysql Secret as MYSQL_PROXY_USER and MYSQL_PROXY_PASSWORD env variable
		in = upsertEnv(in, proxysql)

		in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volume...)

		if opts.configSource != nil {
			in.Spec.Template = upsertCustomConfig(in.Spec.Template, opts.configSource)
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
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt, proxysql.Namespace, opts.stsName,
		)
	}

	return vt, nil
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

func upsertEnv(statefulSet *apps.StatefulSet, proxysql *api.ProxySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularProxySQL || container.Name == "exporter" {
			envs := []core.EnvVar{
				{
					Name: "MYSQL_PROXY_USER",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: proxysql.Spec.ProxySQLSecret.SecretName,
							},
							Key: api.ProxySQLUserKey,
						},
					},
				},
				{
					Name: "MYSQL_PROXY_PASSWORD",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: proxysql.Spec.ProxySQLSecret.SecretName,
							},
							Key: api.ProxySQLPasswordKey,
						},
					},
				},
			}

			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return statefulSet
}
