/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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

type database interface {
	PeerName(i int) string
	GetAuthSecretName() string
}

var _ database = api.PerconaXtraDB{}
var _ database = api.MySQL{}

type workloadOptions struct {
	// App level options
	stsName   string
	labels    map[string]string
	selectors map[string]string

	// database container options
	conatainerName string
	image          string
	cmd            []string // cmd of `proxysql` container
	args           []string // args of `proxysql` container
	ports          []core.ContainerPort
	envList        []core.EnvVar // envList of `proxysql` container
	volumeMount    []core.VolumeMount
	configSecret   *core.LocalObjectReference

	// monitor container
	monitorContainer *core.Container

	// pod Template level options
	replicas       *int32
	gvrSvcName     string
	podTemplate    *ofst.PodTemplateSpec
	initContainers []core.Container
	volume         []core.Volume // volumes to mount on stsPodTemplate
}

func (c *Controller) ensureProxySQLNode(db *api.ProxySQL) (kutil.VerbType, error) {
	proxysqlVersion, err := c.DBClient.CatalogV1alpha1().ProxySQLVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	var ports = []core.ContainerPort{
		{
			Name:          api.ProxySQLDatabasePortName,
			ContainerPort: api.ProxySQLDatabasePort,
			Protocol:      core.ProtocolTCP,
		},
		{
			Name:          api.ProxySQLAdminPortName,
			ContainerPort: api.ProxySQLAdminPort,
			Protocol:      core.ProtocolTCP,
		},
	}

	db.Spec.PodTemplate.Spec.ServiceAccountName = db.OffshootName()

	var backendDB database
	backend := db.Spec.Backend
	gk := schema.GroupKind{Group: *backend.Ref.APIGroup, Kind: backend.Ref.Kind}

	switch gk {
	case api.Kind(api.ResourceKindPerconaXtraDB):
		backendDB, err = c.DBClient.KubedbV1alpha2().PerconaXtraDBs(db.Namespace).Get(context.TODO(), backend.Ref.Name, metav1.GetOptions{})
	case api.Kind(api.ResourceKindMySQL):
		backendDB, err = c.DBClient.KubedbV1alpha2().MySQLs(db.Namespace).Get(context.TODO(), backend.Ref.Name, metav1.GetOptions{})
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
	if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
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
					db.Spec.Monitor.Prometheus.Exporter.Port, db.StatsService().Path(), strings.Join(db.Spec.Monitor.Prometheus.Exporter.Args, " ")),
			},
			Image: proxysqlVersion.Spec.Exporter.Image,
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

	envList := []core.EnvVar{
		{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: backendDB.GetAuthSecretName(),
					},
					Key: core.BasicAuthPasswordKey,
				},
			},
		},
		{
			Name:  "PEERS",
			Value: strings.Join(peers, ","),
		},
		{
			Name:  "LOAD_BALANCE_MODE",
			Value: string(*db.Spec.Mode),
		},
	}

	opts := workloadOptions{
		stsName:          db.OffshootName(),
		labels:           db.OffshootLabels(),
		selectors:        db.OffshootSelectors(),
		conatainerName:   api.ResourceSingularProxySQL,
		image:            proxysqlVersion.Spec.Proxysql.Image,
		args:             nil,
		cmd:              nil,
		ports:            ports,
		envList:          envList,
		initContainers:   nil,
		gvrSvcName:       db.GoverningServiceName(),
		podTemplate:      &db.Spec.PodTemplate,
		configSecret:     db.Spec.ConfigSecret,
		replicas:         db.Spec.Replicas,
		volume:           nil,
		volumeMount:      nil,
		monitorContainer: monitorContainer,
	}

	return c.ensureStatefulSet(db, opts)
}

func (c *Controller) checkStatefulSet(db *api.ProxySQL, stsName string) error {
	// StatefulSet for ProxySQL database
	statefulSet, err := c.Client.AppsV1().StatefulSets(db.Namespace).Get(context.TODO(), stsName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindProxySQL ||
		statefulSet.Labels[api.LabelProxySQLName] != db.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, db.Namespace, stsName)
	}

	return nil
}

func upsertCustomConfig(template core.PodTemplateSpec, configSecret *core.LocalObjectReference) core.PodTemplateSpec {
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

func (c *Controller) ensureStatefulSet(db *api.ProxySQL, opts workloadOptions) (kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkStatefulSet(db, opts.stsName); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for ProxySQL database
	statefulSetMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindProxySQL))

	readinessProbe := pt.Spec.ReadinessProbe
	if readinessProbe != nil && structs.IsZero(*readinessProbe) {
		readinessProbe = nil
	}
	livenessProbe := pt.Spec.LivenessProbe
	if livenessProbe != nil && structs.IsZero(*livenessProbe) {
		livenessProbe = nil
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

			if opts.monitorContainer != nil && db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
					in.Spec.Template.Spec.Containers, *opts.monitorContainer)
			}

			// Set proxysql Secret as MYSQL_PROXY_USER and MYSQL_PROXY_PASSWORD env variable
			in = upsertEnv(in, db)

			in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volume...)

			if opts.configSecret != nil {
				in.Spec.Template = upsertCustomConfig(in.Spec.Template, opts.configSecret)
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
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt, db.Namespace, opts.stsName,
		)
	}

	return vt, nil
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

func upsertEnv(statefulSet *apps.StatefulSet, db *api.ProxySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularProxySQL || container.Name == "exporter" {
			envs := []core.EnvVar{
				{
					Name: "MYSQL_PROXY_USER",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: db.Spec.AuthSecret.Name,
							},
							Key: core.BasicAuthUsernameKey,
						},
					},
				},
				{
					Name: "MYSQL_PROXY_PASSWORD",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: db.Spec.AuthSecret.Name,
							},
							Key: core.BasicAuthPasswordKey,
						},
					},
				},
			}

			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return statefulSet
}
