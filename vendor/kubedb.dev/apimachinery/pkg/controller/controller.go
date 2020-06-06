/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	kubedbinformers "kubedb.dev/apimachinery/client/informers/externalversions"

	"github.com/appscode/go/log/golog"
	cm "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	cmInformers "github.com/jetstack/cert-manager/pkg/client/informers/externalversions"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	externalInformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	appcat_in "kmodules.xyz/custom-resources/client/informers/externalversions"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stashInformers "stash.appscode.dev/apimachinery/client/informers/externalversions"
)

type Controller struct {
	ClientConfig *rest.Config
	// Kubernetes client
	Client kubernetes.Interface
	// CRD Client
	CRDClient crd_cs.Interface
	// ThirdPartyExtension client
	ExtClient cs.Interface //#TODO: rename to DBClient
	// Dynamic client
	DynamicClient dynamic.Interface
	// AppCatalog client
	AppCatalogClient appcat_cs.Interface
	// StashClient for stash
	StashClient scs.Interface
	//CertManagerClient for cert-manger
	CertManagerClient cm.Interface
	// Cluster topology when the operator started
	ClusterTopology *core_util.Topology
}

type Config struct {
	// Informer factory
	KubeInformerFactory        informers.SharedInformerFactory
	KubedbInformerFactory      kubedbinformers.SharedInformerFactory
	StashInformerFactory       stashInformers.SharedInformerFactory
	AppCatInformerFactory      appcat_in.SharedInformerFactory
	ExternalInformerFactory    externalInformers.SharedInformerFactory
	CertManagerInformerFactory cmInformers.SharedInformerFactory

	// restoreSession queue
	RSQueue    *queue.Worker
	RSInformer cache.SharedIndexInformer

	// Secret
	SecretInformer cache.SharedIndexInformer
	SecretLister   corelisters.SecretLister

	OperatorNamespace       string
	GoverningService        string
	ResyncPeriod            time.Duration
	MaxNumRequeues          int
	NumThreads              int
	LoggerOptions           golog.Options
	EnableAnalytics         bool
	AnalyticsClientID       string
	WatchNamespace          string
	EnableValidatingWebhook bool
	EnableMutatingWebhook   bool
}

type DBHelper interface {
	GetDatabase(metav1.ObjectMeta) (runtime.Object, error)
	SetDatabaseStatus(metav1.ObjectMeta, api.DatabasePhase, string) error
	UpsertDatabaseAnnotation(metav1.ObjectMeta, map[string]string) error
}
