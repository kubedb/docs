package controller

import (
	cs "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	// Kubernetes client
	Client kubernetes.Interface
	// Api Extension Client
	ApiExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	// ThirdPartyExtension client
	ExtClient cs.KubedbV1alpha1Interface
}
