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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	core_util "kmodules.xyz/client-go/core/v1"
)

func defaultEnv(db *api.MySQL) []core.EnvVar {
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
	return envs
}

func replicatedEnv(db *api.MySQL) []core.EnvVar {
	envs := []core.EnvVar{
		{
			Name:  "BASE_NAME",
			Value: db.OffshootName(),
		},
		{
			Name:  "GOV_SVC",
			Value: db.GoverningServiceName(),
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "DB_NAME",
			Value: db.GetName(),
		},
		{
			Name: "POD_IP",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
	}

	if db.UsesGroupReplication() {
		envs = append(envs, []core.EnvVar{
			{
				Name:  "GROUP_NAME",
				Value: db.Spec.Topology.Group.Name,
			},
		}...)
	}

	return envs
}

func (c *Reconciler) readReplicaEnv(db *api.MySQL, source *api.MySQL) []core.EnvVar {
	if db.IsReadReplica() {
		sourceName := db.Spec.Topology.ReadReplica.SourceRef.Name
		sourceNameSpace := db.Spec.Topology.ReadReplica.SourceRef.Namespace

		hostToConnect := source.PrimaryServiceDNS()
		primaryHost := source.PrimaryServiceDNS()
		if source.UsesGroupReplication() {
			hostToConnect = source.StandbyServiceDNS()
		}
		envs := []core.EnvVar{
			{
				Name:  "hostToConnect",
				Value: hostToConnect,
			},
			{
				Name:  "primaryHost",
				Value: primaryHost,
			},
		}
		dbObj, err := c.DBClient.KubedbV1alpha2().MySQLs(sourceNameSpace).Get(context.TODO(), sourceName, metav1.GetOptions{})
		if err != nil {
			klog.Error(err)
			return envs
		}
		if dbObj.Spec.RequireSSL {
			envs = append(envs, []core.EnvVar{
				{
					Name:  "source_ssl",
					Value: "true",
				},
			}...)
		}
		return envs
	}
	return nil
}

func (c *Reconciler) updateStatefulSetEnv(statefulSet *apps.StatefulSet, db *api.MySQL, source *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL || container.Name == api.ContainerExporterName || container.Name == api.ReplicationModeDetectorContainerName {

			envs := defaultEnv(db)

			if (db.UsesGroupReplication() || db.IsInnoDBCluster() || db.IsReadReplica()) && container.Name == api.ResourceSingularMySQL {
				envs = append(envs, replicatedEnv(db)...)
			}

			if db.IsReadReplica() {
				envs = append(envs, c.readReplicaEnv(db, source)...)
			}
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, db.Spec.PodTemplate.Spec.Env...)
			return statefulSet
		}
	}
	return statefulSet
}
