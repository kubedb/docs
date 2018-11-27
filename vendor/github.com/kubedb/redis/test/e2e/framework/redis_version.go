package framework

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invocation) RedisVersion() *api.RedisVersion {
	return &api.RedisVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBVersion,
			Labels: map[string]string{
				"app": i.app,
			},
		},
		Spec: api.RedisVersionSpec{
			Version: DBVersion,
			DB: api.RedisVersionDatabase{
				Image: fmt.Sprintf("%s/redis:%s", DockerRegistry, DBVersion),
			},
			Exporter: api.RedisVersionExporter{
				Image: fmt.Sprintf("%s/operator:%s", DockerRegistry, ExporterTag),
			},
		},
	}
}

func (f *Framework) CreateRedisVersion(obj *api.RedisVersion) error {
	_, err := f.extClient.CatalogV1alpha1().RedisVersions().Create(obj)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (f *Framework) DeleteRedisVersion(meta metav1.ObjectMeta) error {
	return f.extClient.CatalogV1alpha1().RedisVersions().Delete(meta.Name, &metav1.DeleteOptions{})
}
