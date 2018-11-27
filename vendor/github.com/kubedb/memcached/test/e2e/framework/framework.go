package framework

import (
	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/kutil/tools/portforward"
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
	tunnel       *portforward.Tunnel
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
		name:         "memcached-operator",
		namespace:    rand.WithUniqSuffix(api.ResourceSingularMemcached),
		StorageClass: storageClass,
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("memcached-e2e"),
	}
}

func (fi *Invocation) ExtClient() cs.Interface {
	return fi.extClient
}

func (fi *Invocation) RestConfig() *rest.Config {
	return fi.restConfig
}

type Invocation struct {
	*Framework
	app string
}
