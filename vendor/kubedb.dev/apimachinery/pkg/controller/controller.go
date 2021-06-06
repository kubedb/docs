/*
Copyright AppsCode Inc. and Contributors

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

	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	kubedbinformers "kubedb.dev/apimachinery/client/informers/externalversions"

	cmInformers "github.com/jetstack/cert-manager/pkg/client/informers/externalversions"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	externalInformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslister "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/discovery"
	"kmodules.xyz/client-go/tools/queue"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	appcat_in "kmodules.xyz/custom-resources/client/informers/externalversions"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stashinformer "stash.appscode.dev/apimachinery/client/informers/externalversions"
	lister "stash.appscode.dev/apimachinery/client/listers/stash/v1beta1"
)

type Controller struct {
	ClientConfig *rest.Config
	// Kubernetes client
	Client kubernetes.Interface
	// CRD Client
	CRDClient crd_cs.Interface
	// KubeDB client
	DBClient cs.Interface
	// Dynamic client
	DynamicClient dynamic.Interface
	// AppCatalog client
	AppCatalogClient appcat_cs.Interface
	// Cluster topology when the operator started
	ClusterTopology *core_util.Topology
	// RESTMapper allows clients to map resources to kind, and map kind and version
	// to interfaces for manipulating those objects.
	Mapper discovery.ResourceMapper
	// Event Recorder
	Recorder record.EventRecorder
	// Audit Event Publisher
	Auditor cache.ResourceEventHandler
}

type Config struct {
	// Informer factory
	KubeInformerFactory        informers.SharedInformerFactory
	KubedbInformerFactory      kubedbinformers.SharedInformerFactory
	AppCatInformerFactory      appcat_in.SharedInformerFactory
	ExternalInformerFactory    externalInformers.SharedInformerFactory
	CertManagerInformerFactory cmInformers.SharedInformerFactory

	// External tool to initialize the database
	Initializers Initializers

	// Secret
	SecretInformer cache.SharedIndexInformer
	SecretLister   corelisters.SecretLister

	// StatefulSet Watcher
	StsQueue    *queue.Worker
	StsInformer cache.SharedIndexInformer
	StsLister   appslister.StatefulSetLister

	ResyncPeriod            time.Duration
	ReadinessProbeInterval  time.Duration
	MaxNumRequeues          int
	NumThreads              int
	EnableAnalytics         bool
	AnalyticsClientID       string
	WatchNamespace          string
	EnableValidatingWebhook bool
	EnableMutatingWebhook   bool
}

type Initializers struct {
	Stash StashInitializer
}

type StashInitializer struct {
	StashClient          scs.Interface
	StashInformerFactory stashinformer.SharedInformerFactory
	// StashInitializer RestoreSession
	RSQueue    *queue.Worker
	RSInformer cache.SharedIndexInformer
	RSLister   lister.RestoreSessionLister

	// StashInitializer RestoreBatch
	RBQueue    *queue.Worker
	RBInformer cache.SharedIndexInformer
	RBLister   lister.RestoreBatchLister
}
