package framework

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invocation) ElasticsearchVersion() *api.ElasticsearchVersion {
	return &api.ElasticsearchVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBVersion,
			Labels: map[string]string{
				"app": i.app,
			},
		},
		Spec: api.ElasticsearchVersionSpec{
			Version: DBVersion,
			DB: api.ElasticsearchVersionDatabase{
				Image: fmt.Sprintf("%s/elasticsearch:%s", DockerRegistry, DBVersion),
			},
			Exporter: api.ElasticsearchVersionExporter{
				Image: fmt.Sprintf("%s/elasticsearch_exporter:%v", DockerRegistry, ExporterTag),
			},
			Tools: api.ElasticsearchVersionTools{
				Image: fmt.Sprintf("%s/elasticsearch-tools:%s", DockerRegistry, DBVersion),
			},
		},
	}
}

func (f *Framework) CreateElasticsearchVersion(obj *api.ElasticsearchVersion) error {
	_, err := f.extClient.CatalogV1alpha1().ElasticsearchVersions().Create(obj)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (f *Framework) DeleteElasticsearchVersion(meta metav1.ObjectMeta) error {
	return f.extClient.CatalogV1alpha1().ElasticsearchVersions().Delete(meta.Name, &metav1.DeleteOptions{})
}
