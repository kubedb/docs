package controller

import (
	"time"

	tcs "github.com/k8sdb/apimachinery/client/clientset"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
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
