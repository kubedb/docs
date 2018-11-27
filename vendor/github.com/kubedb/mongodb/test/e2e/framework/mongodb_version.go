package framework

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invocation) MongoDBVersion() *api.MongoDBVersion {
	return &api.MongoDBVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBVersion,
			Labels: map[string]string{
				"app": i.app,
			},
		},
		Spec: api.MongoDBVersionSpec{
			Version: DBVersion,
			DB: api.MongoDBVersionDatabase{
				Image: fmt.Sprintf("%s/mongo:%s", DockerRegistry, DBVersion),
			},
			Exporter: api.MongoDBVersionExporter{
				Image: fmt.Sprintf("%s/operator:%s", DockerRegistry, ExporterTag),
			},
			Tools: api.MongoDBVersionTools{
				Image: fmt.Sprintf("%s/mongo-tools:%s", DockerRegistry, DBVersion),
			},
		},
	}
}

func (f *Framework) CreateMongoDBVersion(obj *api.MongoDBVersion) error {
	_, err := f.extClient.CatalogV1alpha1().MongoDBVersions().Create(obj)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (f *Framework) DeleteMongoDBVersion(meta metav1.ObjectMeta) error {
	return f.extClient.CatalogV1alpha1().MongoDBVersions().Delete(meta.Name, &metav1.DeleteOptions{})
}
