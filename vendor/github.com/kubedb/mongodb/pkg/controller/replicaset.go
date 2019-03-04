package controller

import (
	"github.com/appscode/go/types"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func upsertRSArgs(statefulSet *apps.StatefulSet, mongodb *api.MongoDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMongoDB {
			statefulSet.Spec.Template.Spec.Containers[i].Args = meta_util.UpsertArgumentList(
				statefulSet.Spec.Template.Spec.Containers[i].Args,
				[]string{
					"--replSet=" + mongodb.Spec.ReplicaSet.Name,
					"--bind_ip=0.0.0.0",
					"--keyFile=" + configDirectoryPath + "/" + KeyForKeyFile,
				})
			statefulSet.Spec.Template.Spec.Containers[i].Command = []string{
				"mongod",
			}
		}
	}
	statefulSet.Spec.Template.Spec.SecurityContext = &core.PodSecurityContext{
		FSGroup:      types.Int64P(999),
		RunAsNonRoot: types.BoolP(true),
		RunAsUser:    types.Int64P(999),
	}
	return statefulSet
}

func (c *Controller) upsertRSInitContainer(statefulSet *apps.StatefulSet, mongodb *api.MongoDB, mongodbVersion *catalog.MongoDBVersion) *apps.StatefulSet {
	bootstrapContainer := core.Container{
		Name:            InitBootstrapContainerName,
		Image:           mongodbVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"peer-finder"},
		Args:            []string{"-on-start=/usr/local/bin/on-start.sh", "-service=" + c.GoverningService},
		Env: []core.EnvVar{
			{
				Name: "POD_NAMESPACE",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.namespace",
					},
				},
			},
			{
				Name:  "REPLICA_SET",
				Value: mongodb.Spec.ReplicaSet.Name,
			},
			{
				Name:  "AUTH",
				Value: "true",
			},
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
			{
				Name:      dataDirectoryName,
				MountPath: dataDirectoryPath,
			},
		},
	}

	initContainers := statefulSet.Spec.Template.Spec.InitContainers
	statefulSet.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(initContainers, bootstrapContainer)

	rsVolume := core.Volume{
		Name: initialKeyDirectoryName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				DefaultMode: types.Int32P(256),
				SecretName:  mongodb.Spec.ReplicaSet.KeyFile.SecretName,
			},
		},
	}
	volumes := statefulSet.Spec.Template.Spec.Volumes
	statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(volumes, rsVolume)
	return statefulSet
}
