package controller

import (
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1"
	batch "k8s.io/api/batch/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

type Snapshotter interface {
	ValidateSnapshot(*api.Snapshot) error
	GetDatabase(metav1.ObjectMeta) (runtime.Object, error)
	GetSnapshotter(*api.Snapshot) (*batch.Job, error)
	WipeOutSnapshot(*api.Snapshot) error
	SetDatabaseStatus(metav1.ObjectMeta, api.DatabasePhase, string) error
	UpsertDatabaseAnnotation(metav1.ObjectMeta, map[string]string) error
}

type Deleter interface {
	Exists(*metav1.ObjectMeta) (bool, error)
	PauseDatabase(*api.DormantDatabase) error
	WipeOutDatabase(*api.DormantDatabase) error
	ResumeDatabase(*api.DormantDatabase) error
}
