package controller

import (
	"time"

	"github.com/appscode/go/log/golog"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	kubedbinformers "github.com/kubedb/apimachinery/client/informers/externalversions"
	batch "k8s.io/api/batch/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Controller struct {
	// Kubernetes client
	Client kubernetes.Interface
	// Api Extension Client
	ApiExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	// ThirdPartyExtension client
	ExtClient cs.Interface
	// Dynamic client
	DynamicClient dynamic.Interface
}

type Config struct {
	// Informer factory
	KubeInformerFactory   informers.SharedInformerFactory
	KubedbInformerFactory kubedbinformers.SharedInformerFactory

	// DormantDb queue
	DrmnQueue    *queue.Worker
	DrmnInformer cache.SharedIndexInformer
	// job queue
	JobQueue    *queue.Worker
	JobInformer cache.SharedIndexInformer
	// snapshot queue
	SnapQueue    *queue.Worker
	SnapInformer cache.SharedIndexInformer

	EnableRBAC        bool
	OperatorNamespace string
	GoverningService  string
	ResyncPeriod      time.Duration
	MaxNumRequeues    int
	NumThreads        int
	LoggerOptions     golog.Options
	EnableAnalytics   bool
	AnalyticsClientID string
	WatchNamespace    string
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
	// WaitUntilPaused will block until db pods and service are deleted. PV/PVC will remain intact.
	WaitUntilPaused(*api.DormantDatabase) error
	// WipeOutDatabase won't need to handle snapshots and PVCs.
	// All other elements of database will be Wipedout on WipeOutDatabase function.
	// Ex: secrets, wal-g data and other staff that is required.
	WipeOutDatabase(*api.DormantDatabase) error
}
