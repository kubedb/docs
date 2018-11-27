package framework

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invocation) MemcachedVersion() *api.MemcachedVersion {
	return &api.MemcachedVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBVersion,
			Labels: map[string]string{
				"app": i.app,
			},
		},
		Spec: api.MemcachedVersionSpec{
			Version: DBVersion,
			DB: api.MemcachedVersionDatabase{
				Image: fmt.Sprintf("%s/memcached:%s", DockerRegistry, DBVersion),
			},
			Exporter: api.MemcachedVersionExporter{
				Image: fmt.Sprintf("%s/operator:%s", DockerRegistry, ExporterTag),
			},
		},
	}
}
func (f *Framework) CreateMemcachedVersion(obj *api.MemcachedVersion) error {
	_, err := f.extClient.CatalogV1alpha1().MemcachedVersions().Create(obj)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}
func (f *Framework) DeleteMemcachedVersion(meta metav1.ObjectMeta) error {
	return f.extClient.CatalogV1alpha1().MemcachedVersions().Delete(meta.Name, &metav1.DeleteOptions{})
}
