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
	"strconv"
	"strings"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

const (
	routerIPv6Config = `
[routing:mycluster_rw]
bind_address = ::0

[routing:mycluster_ro]
bind_address = ::0

[routing:mycluster_x_rw]
bind_address = ::0

[routing:mycluster_x_ro]
bind_address = ::0

[http_server]
bind_address = ::0
`
	routerTLSConfig = `
[DEFAULT]
client_ssl_mode=REQUIRED
server_ssl_mode=REQUIRED
server_ssl_verify=VERIFY_CA
server_ssl_ca=/etc/mysql/certs/ca.crt
server_ssl_capath=/etc/mysql/certs
client_ssl_cert=/etc/mysql/certs/server.crt
client_ssl_key=/etc/mysql/certs/server.key
`
)

func (c *Controller) ensureRouter(db *api.MySQL) error {
	//router will be created after the server is ready
	if !kmapi.IsConditionTrue(db.Status.Conditions, api.ServerReady) {
		return nil
	}

	_, vt, err := c.createOrPatchRouterStatefulSet(db)
	if err != nil {
		return err
	}
	if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"successfully patched %v StatefulSet %s",
			vt,
			db.GetRouterName(),
		)
	}
	return nil
}

func (c *Controller) ensureRouterConfigSecret(db *api.MySQL) error {
	secretMeta := metav1.ObjectMeta{
		Name:      db.GetRouterName() + "-config",
		Namespace: db.Namespace,
	}
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, _, err := core_util.CreateOrPatchSecret(context.TODO(), c.Client, secretMeta, func(in *core.Secret) *core.Secret {
		in.Labels = db.OffshootLabels()
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

		in.Type = core.SecretTypeOpaque
		routerConfig := ""
		if db.Spec.UseAddressType == api.AddressTypeIPv6 {
			routerConfig = routerIPv6Config
		}
		if db.Spec.RequireSSL {
			routerConfig = routerConfig + "\n" + routerTLSConfig
		}

		in.Data = map[string][]byte{
			"custom.conf": []byte(routerConfig),
		}
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) createOrPatchRouterStatefulSet(db *api.MySQL) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      db.GetRouterName(),
		Namespace: db.Namespace,
	}
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	mysqlVersion, err := c.DBClient.CatalogV1alpha1().MySQLVersions().Get(context.TODO(), db.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return app_util.CreateOrPatchStatefulSet(context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = db.RouterPodControllerLabels()
			in.Annotations = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = db.Spec.Topology.InnoDBCluster.Router.Replicas
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: db.RouterOffshootSelectors(),
			}
			in.Spec.Template.Labels = db.RouterPodLabels()
			in.Spec.Template.Annotations = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				append(getRouterInitContainer(in, mysqlVersion), db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.InitContainers...))

			container := core.Container{
				Name:            api.MySQLRouterContainerName,
				Image:           mysqlVersion.Spec.Router.Image,
				Command:         []string{"/scripts/mysql-router-init"},
				Args:            db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.Args,
				Resources:       db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.Resources,
				LivenessProbe:   db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.LivenessProbe,
				ReadinessProbe:  db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.ReadinessProbe,
				Lifecycle:       db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.Lifecycle,
				SecurityContext: db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.ContainerSecurityContext,
				Ports: []core.ContainerPort{
					{
						Name:          api.MySQLRouterReadWritePortName,
						ContainerPort: api.MySQLRouterReadWritePort,
						Protocol:      core.ProtocolTCP,
					},
					{
						Name:          api.MySQLRouterReadOnlyPortName,
						ContainerPort: api.MySQLRouterReadOnlyPort,
						Protocol:      core.ProtocolTCP,
					},
				},
				VolumeMounts: []core.VolumeMount{
					{
						Name:      api.MySQLRouterInitScriptDirectoryName,
						MountPath: api.MySQLRouterInitScriptDirectoryPath,
					},
					{
						Name:      api.MySQLRouterConfigDirectoryName,
						MountPath: api.MySQLRouterConfigDirectoryPath,
					},
				},
			}

			container.Command = []string{
				"/scripts/mysql-router-init",
			}
			userArgs := meta_util.ParseArgumentListToMap(db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.Args)

			args := meta_util.BuildArgumentListFromMap(userArgs, nil)

			initScriptArgs := []string{
				fmt.Sprintf("-address-type=%s", db.Spec.UseAddressType),
			}

			if db.Spec.UseAddressType.IsIP() {
				initScriptArgs = append(initScriptArgs, fmt.Sprintf("-selector=%s", labels.Set(db.OffshootSelectors()).String()))
			}

			initScriptArgs = append(initScriptArgs, fmt.Sprintf("-ns=%s", db.Namespace))

			container.Args = append(initScriptArgs,
				"-on-start",
				strings.Join(append([]string{"/scripts/router_run.sh"}, args...), " "))

			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, container)

			in.Spec.Template.Spec.NodeSelector = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.NodeSelector
			in.Spec.Template.Spec.Affinity = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.Affinity
			if db.Spec.PodTemplate.Spec.SchedulerName != "" {
				in.Spec.Template.Spec.SchedulerName = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.SchedulerName
			}
			in.Spec.Template.Spec.Tolerations = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.Tolerations
			in.Spec.Template.Spec.ImagePullSecrets = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.ImagePullSecrets
			in.Spec.Template.Spec.PriorityClassName = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.PriorityClassName
			in.Spec.Template.Spec.Priority = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.Priority
			in.Spec.Template.Spec.HostNetwork = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.HostNetwork
			in.Spec.Template.Spec.DNSPolicy = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.DNSPolicy
			in.Spec.Template.Spec.HostPID = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.HostPID
			in.Spec.Template.Spec.HostIPC = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.HostIPC
			in.Spec.Template.Spec.SecurityContext = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.SecurityContext
			in.Spec.Template.Spec.ServiceAccountName = db.Spec.Topology.InnoDBCluster.Router.PodTemplate.Spec.ServiceAccountName
			in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
				Type: apps.OnDeleteStatefulSetStrategyType,
			}

			in = upsertRouterEnv(in, db)

			in.Spec.Template.Spec.Volumes = []core.Volume{
				{
					Name: api.MySQLRouterInitScriptDirectoryName,
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				},
				{
					Name: api.MySQLRouterConfigDirectoryName,
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: db.GetRouterName() + "-config",
						},
					},
				},
			}

			in = upsertRouterTLSVolume(in, db)

			return in
		},
		metav1.PatchOptions{},
	)

}

func upsertRouterEnv(sts *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	var host string
	if db.Spec.UseAddressType == api.AddressTypeDNS {
		host = fmt.Sprintf("%s-0.%s-pods.%s.svc", db.Name, db.Name, db.Namespace)
	} else {
		host = fmt.Sprintf("%s-0", db.Name)
	}
	envs := []core.EnvVar{
		{
			Name:  "MYSQL_HOST",
			Value: host,
		},
		{
			Name:  "MYSQL_PORT",
			Value: strconv.Itoa(api.MySQLDatabasePort),
		},
		{
			Name: "MYSQL_PASSWORD",
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
			Name:  "MYSQL_USER",
			Value: api.MySQLReplicationUser,
		},
	}

	for i, container := range sts.Spec.Template.Spec.Containers {
		if container.Name == api.MySQLRouterContainerName {
			sts.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return sts
}

func getRouterInitContainer(statefulSet *apps.StatefulSet, mysqlVersion *v1alpha1.MySQLVersion) []core.Container {
	statefulSet.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.InitContainers,
		core.Container{
			Name:  "mysql-router-init",
			Image: mysqlVersion.Spec.RouterInitContainer.Image,
			VolumeMounts: []core.VolumeMount{
				{
					Name:      api.MySQLRouterInitScriptDirectoryName,
					MountPath: api.MySQLRouterInitScriptDirectoryPath,
				},
			},
		})
	return statefulSet.Spec.Template.Spec.InitContainers
}

func upsertRouterTLSVolume(sts *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	if db.Spec.TLS != nil {
		volume := core.Volume{
			Name: api.MySQLRouterTLSDirectoryName,
			VolumeSource: core.VolumeSource{
				Projected: &core.ProjectedVolumeSource{
					Sources: []core.VolumeProjection{
						{
							Secret: &core.SecretProjection{
								LocalObjectReference: core.LocalObjectReference{
									Name: db.MustCertSecretName(api.MySQLServerCert),
								},
								Items: []core.KeyToPath{
									{
										Key:  "ca.crt",
										Path: "ca.crt",
									},
									{
										Key:  "tls.crt",
										Path: "server.crt",
									},
									{
										Key:  "tls.key",
										Path: "server.key",
									},
								},
							},
						},
						{
							Secret: &core.SecretProjection{
								LocalObjectReference: core.LocalObjectReference{
									Name: db.MustCertSecretName(api.MySQLClientCert),
								},
								Items: []core.KeyToPath{
									{
										Key:  "tls.crt",
										Path: "client.crt",
									},
									{
										Key:  "tls.key",
										Path: "client.key",
									},
								},
							},
						},
					},
				},
			},
		}

		for i, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == api.MySQLRouterContainerName {
				volumeMount := core.VolumeMount{
					Name:      api.MySQLRouterTLSDirectoryName,
					MountPath: api.MySQLRouterTLSDirectoryPath,
				}
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts
			}
		}

		sts.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
			sts.Spec.Template.Spec.Volumes,
			volume,
		)
	} else {
		//clean up volume
		for i, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == api.MySQLRouterContainerName {
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.EnsureVolumeMountDeleted(sts.Spec.Template.Spec.Containers[i].VolumeMounts, api.MySQLRouterTLSDirectoryName)
			}
		}
		sts.Spec.Template.Spec.Volumes = core_util.EnsureVolumeDeleted(sts.Spec.Template.Spec.Volumes, api.MySQLRouterTLSDirectoryName)
	}
	return sts
}
