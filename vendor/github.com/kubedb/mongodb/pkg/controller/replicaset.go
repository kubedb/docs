package controller

import (
	"fmt"

	"github.com/appscode/go/types"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func topologyInitContainer(
	mongodb *api.MongoDB,
	mongodbVersion *catalog.MongoDBVersion,
	podTemplate *ofst.PodTemplateSpec,
	repSetName string,
	gvrSvc string,
	scriptName string,
) (core.Container, []core.Volume) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if podTemplate != nil {
		pt = *podTemplate
	}

	bootstrapContainer := core.Container{
		Name:            InitBootstrapContainerName,
		Image:           mongodbVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"peer-finder"},
		Args: []string{
			fmt.Sprintf("-on-start=/usr/local/bin/%v", scriptName),
			"-service=" + gvrSvc,
		},
		Env: core_util.UpsertEnvVars([]core.EnvVar{
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
				Value: repSetName,
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
		}, pt.Spec.Env...),
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
		Resources: pt.Spec.Resources,
	}

	rsVolume := []core.Volume{
		{
			Name: initialKeyDirectoryName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					DefaultMode: types.Int32P(256),
					SecretName:  mongodb.Spec.CertificateSecret.SecretName,
				},
			},
		},
	}

	//only on mongos in case of sharding (which is handled on 'ensureMongosNode'.
	if mongodb.Spec.ShardTopology == nil && mongodb.Spec.Init != nil && mongodb.Spec.Init.ScriptSource != nil {
		rsVolume = append(rsVolume, core.Volume{
			Name:         "initial-script",
			VolumeSource: mongodb.Spec.Init.ScriptSource.VolumeSource,
		})

		bootstrapContainer.VolumeMounts = core_util.UpsertVolumeMount(
			bootstrapContainer.VolumeMounts,
			core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/docker-entrypoint-initdb.d",
			},
		)
	}

	return bootstrapContainer, rsVolume
}
