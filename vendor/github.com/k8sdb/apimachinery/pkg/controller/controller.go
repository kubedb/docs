package controller

import (
	"time"

	cs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	// Kubernetes client
	Client kubernetes.Interface
	// ThirdPartyExtension client
	ExtClient cs.KubedbV1alpha1Interface
}

const (
	sleepDuration = time.Second * 10
)
