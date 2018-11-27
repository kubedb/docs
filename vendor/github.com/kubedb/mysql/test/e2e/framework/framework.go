package framework

import (
	"github.com/appscode/go/crypto/rand"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

type Framework struct {
	restConfig   *rest.Config
	kubeClient   kubernetes.Interface
	extClient    cs.Interface
	kaClient     ka.Interface
	namespace    string
	name         string
	StorageClass string
}

func New(
	restConfig *rest.Config,
	kubeClient kubernetes.Interface,
	extClient cs.Interface,
	kaClient ka.Interface,
	storageClass string,
) *Framework {
	return &Framework{
		restConfig:   restConfig,
		kubeClient:   kubeClient,
		extClient:    extClient,
		kaClient:     kaClient,
		name:         "mysql-operator",
		namespace:    rand.WithUniqSuffix(api.ResourceSingularMySQL),
		StorageClass: storageClass,
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("mysql-e2e"),
	}
}

func (fi *Invocation) App() string {
	return fi.app
}

func (fi *Invocation) ExtClient() cs.Interface {
	return fi.extClient
}

type Invocation struct {
	*Framework
	app string
}
