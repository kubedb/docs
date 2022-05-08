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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

// getVolumes return all the volumes for the statefulset
func getVolumes(db *api.PgBouncer) []core.Volume {
	var volumes []core.Volume
	// secretVolume := getVolumeForConfigSecret(db)
	volumes = append(volumes, getVolumeForConfigSecret(db))
	// cfgVolume := getVolumeForAuthSecret(db)
	volumes = append(volumes, getVolumeForAuthSecret(db))
	if db.Spec.TLS != nil {
		if db.Spec.TLS.IssuerRef != nil {
			volumes = getTlsVolumes(volumes, db)
			volumes = getSharedTlsVolumes(volumes)
		}
	}
	return volumes
}

// getSharedTlsVolumes return shared volume for tls
func getSharedTlsVolumes(volumes []core.Volume) []core.Volume {
	configVolume := core.Volume{
		Name: sharedTlsVolumeName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	volumes = core_util.UpsertVolume(volumes, configVolume)
	return volumes
}

// getTlsVolumes return volume for server, client and exporter cert
func getTlsVolumes(volumes []core.Volume, db *api.PgBouncer) []core.Volume {
	servingServerSecretVolume := getVolumeForCertificate(db, api.PgBouncerServerCert)
	volumes = append(volumes, servingServerSecretVolume)

	servingClientSecretVolume := getVolumeForCertificate(db, api.PgBouncerClientCert)
	volumes = append(volumes, servingClientSecretVolume)

	exporterSecretVolume := getVolumeForCertificate(db, api.PgBouncerMetricsExporterCert)
	volumes = append(volumes, exporterSecretVolume)
	return volumes
}

// getVolumeForCertificate return volume for type of cert defined via alias
func getVolumeForCertificate(db *api.PgBouncer, alias api.PgBouncerCertificateAlias) core.Volume {
	secretName := db.GetCertSecretName(alias)
	secretVolume := core.Volume{
		Name: secretName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: secretName,
				Items: []core.KeyToPath{
					{
						Key:  api.TLSCACertFileName,
						Path: api.TLSCACertFileName,
					},
					{
						Key:  api.PgBouncerTLSCrt,
						Path: api.PgBouncerTLSCrt,
					},
					{
						Key:  api.PgBouncerTLSKey,
						Path: api.PgBouncerTLSKey,
					},
				},
			},
		},
	}
	return secretVolume
}

// getVolumeForAuthSecret return volume for auth secret
func getVolumeForAuthSecret(db *api.PgBouncer) core.Volume {
	secretVolume := core.Volume{
		Name: "fallback-userlist",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: db.AuthSecretName(),
			},
		},
	}
	return secretVolume
}

// getVolumeForConfigSecret return volume for config secret
func getVolumeForConfigSecret(db *api.PgBouncer) core.Volume {
	secretVolume := core.Volume{
		Name: db.ConfigSecretName(),
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: db.ConfigSecretName(),
			},
		},
	}
	return secretVolume
}

// getVolumeMountForAuthSecret return volumeMount for auth secret
func getVolumeMountForAuthSecret() *core.VolumeMount {
	secretVolumeMount := &core.VolumeMount{
		Name:      "fallback-userlist",
		MountPath: UserListMountPath,
		ReadOnly:  true,
	}
	return secretVolumeMount
}

// getVolumeMountForConfigSecret return volume for config secret
func getVolumeMountForConfigSecret(db *api.PgBouncer) *core.VolumeMount {
	secretVolumeMount := &core.VolumeMount{
		Name:      db.ConfigSecretName(),
		MountPath: configMountPath,
	}
	return secretVolumeMount
}

// getVolumeMountForSharedTls return volumemount for shared tls volune
func getVolumeMountForSharedTls() *core.VolumeMount {
	tlsVolumeMounts := &core.VolumeMount{
		Name:      sharedTlsVolumeName,
		MountPath: ServingCertMountPath,
	}
	return tlsVolumeMounts
}

// getVolumeMountForPBContainer return volunemount for pgbouncer container
func getVolumeMountForPBContainer(db *api.PgBouncer) []core.VolumeMount {
	var volumeMounts []core.VolumeMount
	// volumemount for pgbouncer config secret
	volumeMounts = append(volumeMounts, *getVolumeMountForConfigSecret(db))
	// volumemount for pgbouncer auth secret to postgres
	volumeMounts = append(volumeMounts, *getVolumeMountForAuthSecret())
	// volumemont for tls
	if db.Spec.TLS != nil {
		if db.Spec.TLS.IssuerRef != nil {
			tlsVolumeMounts := getVolumeMountForSharedTls()
			volumeMounts = append(volumeMounts, *tlsVolumeMounts)
		}
	}
	return volumeMounts
}
