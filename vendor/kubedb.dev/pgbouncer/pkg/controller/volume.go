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
	"path/filepath"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
)

func (c *Controller) getVolumeAndVolumeMountForAuthSecret(db *api.PgBouncer) (*core.Volume, *core.VolumeMount) {
	secretVolume := &core.Volume{
		Name: "fallback-userlist",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: db.AuthSecretName(),
			},
		},
	}
	//Add to volumeMounts to mount the volume
	secretVolumeMount := &core.VolumeMount{
		Name:      "fallback-userlist",
		MountPath: UserListMountPath,
		ReadOnly:  true,
	}
	return secretVolume, secretVolumeMount
}

func (c *Controller) getVolumeAndVolumeMountForCertificate(db *api.PgBouncer, alias api.PgBouncerCertificateAlias) (*core.Volume, *core.VolumeMount) {
	//TODO: this is for issuer only, I'm not sure about clusterIssuer yet
	secretName := db.MustCertSecretName(api.PgBouncerServerCert)
	secretVolume := &core.Volume{
		Name: secretName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName:  secretName,
				DefaultMode: pointer.Int32P(0600),
			},
		},
	}
	//Add to volumeMounts to mount the volume
	secretVolumeMount := &core.VolumeMount{
		Name:      secretName,
		MountPath: filepath.Join(ServingCertMountPath, string(alias)),
		ReadOnly:  true,
	}
	return secretVolume, secretVolumeMount
}
