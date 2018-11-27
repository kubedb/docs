package framework

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invocation) PostgresVersion() *api.PostgresVersion {
	return &api.PostgresVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBVersion,
			Labels: map[string]string{
				"app": i.app,
			},
		},
		Spec: api.PostgresVersionSpec{
			Version: DBVersion,
			DB: api.PostgresVersionDatabase{
				Image: fmt.Sprintf("%s/postgres:%s", DockerRegistry, DBVersion),
			},
			Exporter: api.PostgresVersionExporter{
				Image: fmt.Sprintf("%s/postgres_exporter:%s", DockerRegistry, ExporterTag),
			},
			Tools: api.PostgresVersionTools{
				Image: fmt.Sprintf("%s/postgres-tools:%s", DockerRegistry, DBVersion),
			},
		},
	}
}

func (f *Framework) CreatePostgresVersion(obj *api.PostgresVersion) error {
	_, err := f.extClient.CatalogV1alpha1().PostgresVersions().Create(obj)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (f *Framework) DeletePostgresVersion(meta metav1.ObjectMeta) error {
	return f.extClient.CatalogV1alpha1().PostgresVersions().Delete(meta.Name, &metav1.DeleteOptions{})
}
