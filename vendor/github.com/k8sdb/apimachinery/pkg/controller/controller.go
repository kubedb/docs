package controller

import (
	"time"

	tcs "github.com/k8sdb/apimachinery/client/clientset"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	rest "k8s.io/kubernetes/pkg/client/restclient"
)

type Controller struct {
	// Kubernetes client
	Client clientset.Interface
	// ThirdPartyExtension client
	ExtClient tcs.ExtensionInterface
}

const (
	DatabaseNamePrefix = "k8sdb"
	LabelDatabaseKind  = "k8sdb.com/kind"
	LabelDatabaseName  = "k8sdb.com/name"
	sleepDuration      = time.Second * 10
)

func NewController(c *rest.Config) *Controller {
	client := clientset.NewForConfigOrDie(c)
	extClient := tcs.NewExtensionsForConfigOrDie(c)
	return &Controller{
		Client:    client,
		ExtClient: extClient,
	}
}
