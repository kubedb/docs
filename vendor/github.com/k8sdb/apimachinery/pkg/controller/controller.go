package controller

import (
	"time"

	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	clientset "k8s.io/client-go/kubernetes"
)

type Controller struct {
	// Kubernetes client
	Client clientset.Interface
	// ThirdPartyExtension client
	ExtClient tcs.KubedbV1alpha1Interface
}

const (
	sleepDuration = time.Second * 10
)
