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
package server

import (
	"fmt"
	"os"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/admission/dormantdatabase"
	"kubedb.dev/apimachinery/pkg/admission/namespace"
	"kubedb.dev/apimachinery/pkg/admission/snapshot"
	"kubedb.dev/apimachinery/pkg/eventer"
	esAdmsn "kubedb.dev/elasticsearch/pkg/admission"
	edAdmsn "kubedb.dev/etcd/pkg/admission"
	mcAdmsn "kubedb.dev/memcached/pkg/admission"
	mgAdmsn "kubedb.dev/mongodb/pkg/admission"
	myAdmsn "kubedb.dev/mysql/pkg/admission"
	"kubedb.dev/operator/pkg/controller"
	pgAdmsn "kubedb.dev/postgres/pkg/admission"
	rdAdmsn "kubedb.dev/redis/pkg/admission"

	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	hooks "kmodules.xyz/webhook-runtime/admission/v1beta1"
	admissionreview "kmodules.xyz/webhook-runtime/registry/admissionreview/v1beta1"
)

const (
	apiserviceName = "v1alpha1.validators.kubedb.com"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	utilruntime.Must(admission.AddToScheme(Scheme))

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

type KubeDBServerConfig struct {
	GenericConfig  *genericapiserver.RecommendedConfig
	ExtraConfig    ExtraConfig
	OperatorConfig *controller.OperatorConfig
}

type ExtraConfig struct {
	AdmissionHooks []hooks.AdmissionHook
}

// KubeDBServer contains state for a Kubernetes cluster master/api server.
type KubeDBServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
	Operator         *controller.Controller
}

func (op *KubeDBServer) Run(stopCh <-chan struct{}) error {
	if err := op.Operator.MigrateObservedGeneration(); err != nil {
		return fmt.Errorf("failed  to migrate observedGeneration to int64 for existing objects. Reason: %v", err)
	}
	go op.Operator.Run(stopCh)
	return op.GenericAPIServer.PrepareRun().Run(stopCh)
}

type completedConfig struct {
	GenericConfig  genericapiserver.CompletedConfig
	ExtraConfig    ExtraConfig
	OperatorConfig *controller.OperatorConfig
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *KubeDBServerConfig) Complete() CompletedConfig {
	completedCfg := completedConfig{
		c.GenericConfig.Complete(),
		c.ExtraConfig,
		c.OperatorConfig,
	}

	completedCfg.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "1",
	}

	return CompletedConfig{&completedCfg}
}

// New returns a new instance of KubeDBServer from the given config.
func (c completedConfig) New() (*KubeDBServer, error) {
	genericServer, err := c.GenericConfig.New("pack-server", genericapiserver.NewEmptyDelegate()) // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}

	if c.OperatorConfig.EnableMutatingWebhook {
		c.ExtraConfig.AdmissionHooks = []hooks.AdmissionHook{
			&mgAdmsn.MongoDBMutator{},
			&myAdmsn.MySQLMutator{},
			&pgAdmsn.PostgresMutator{},
			&esAdmsn.ElasticsearchMutator{},
			&edAdmsn.EtcdMutator{},
			&rdAdmsn.RedisMutator{},
			&mcAdmsn.MemcachedMutator{},
		}
	}
	if c.OperatorConfig.EnableValidatingWebhook {
		c.ExtraConfig.AdmissionHooks = append(c.ExtraConfig.AdmissionHooks,
			&mgAdmsn.MongoDBValidator{},
			&snapshot.SnapshotValidator{},
			&dormantdatabase.DormantDatabaseValidator{},
			&myAdmsn.MySQLValidator{},
			&pgAdmsn.PostgresValidator{},
			&esAdmsn.ElasticsearchValidator{},
			&edAdmsn.EtcdValidator{},
			&rdAdmsn.RedisValidator{},
			&mcAdmsn.MemcachedValidator{},
			&namespace.NamespaceValidator{
				Resources: []string{
					api.ResourcePluralElasticsearch,
					api.ResourcePluralEtcd,
					api.ResourcePluralMemcached,
					api.ResourcePluralMongoDB,
					api.ResourcePluralMySQL,
					api.ResourcePluralPostgres,
					api.ResourcePluralRedis,
				},
			})
	}

	ctrl, err := c.OperatorConfig.New()
	if err != nil {
		return nil, err
	}

	s := &KubeDBServer{
		GenericAPIServer: genericServer,
		Operator:         ctrl,
	}

	for _, versionMap := range admissionHooksByGroupThenVersion(c.ExtraConfig.AdmissionHooks...) {
		// TODO we're going to need a later k8s.io/apiserver so that we can get discovery to list a different group version for
		// our endpoint which we'll use to back some custom storage which will consume the AdmissionReview type and give back the correct response
		apiGroupInfo := genericapiserver.APIGroupInfo{
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{},
			// TODO unhardcode this.  It was hardcoded before, but we need to re-evaluate
			OptionsExternalVersion: &schema.GroupVersion{Version: "v1"},
			Scheme:                 Scheme,
			ParameterCodec:         metav1.ParameterCodec,
			NegotiatedSerializer:   Codecs,
		}

		for _, admissionHooks := range versionMap {
			for i := range admissionHooks {
				admissionHook := admissionHooks[i]
				admissionResource, _ := admissionHook.Resource()
				admissionVersion := admissionResource.GroupVersion()

				// just overwrite the groupversion with a random one.  We don't really care or know.
				apiGroupInfo.PrioritizedVersions = appendUniqueGroupVersion(apiGroupInfo.PrioritizedVersions, admissionVersion)

				admissionReview := admissionreview.NewREST(admissionHook.Admit)
				v1alpha1storage, ok := apiGroupInfo.VersionedResourcesStorageMap[admissionVersion.Version]
				if !ok {
					v1alpha1storage = map[string]rest.Storage{}
				}
				v1alpha1storage[admissionResource.Resource] = admissionReview
				apiGroupInfo.VersionedResourcesStorageMap[admissionVersion.Version] = v1alpha1storage
			}
		}

		if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
			return nil, err
		}
	}

	for i := range c.ExtraConfig.AdmissionHooks {
		admissionHook := c.ExtraConfig.AdmissionHooks[i]
		postStartName := postStartHookName(admissionHook)
		if len(postStartName) == 0 {
			continue
		}
		s.GenericAPIServer.AddPostStartHookOrDie(postStartName,
			func(context genericapiserver.PostStartHookContext) error {
				return admissionHook.Initialize(c.OperatorConfig.ClientConfig, context.StopCh)
			},
		)
	}

	if c.OperatorConfig.EnableValidatingWebhook {
		s.GenericAPIServer.AddPostStartHookOrDie("validating-webhook-xray",
			func(context genericapiserver.PostStartHookContext) error {
				go func() {
					xray := reg_util.NewCreateValidatingWebhookXray(c.OperatorConfig.ClientConfig, apiserviceName, &api.Redis{
						TypeMeta: metav1.TypeMeta{
							APIVersion: api.SchemeGroupVersion.String(),
							Kind:       api.ResourceKindRedis,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-redis-for-webhook-xray",
							Namespace: "default",
						},
						Spec: api.RedisSpec{
							StorageType: api.StorageType("Invalid"),
						},
					}, context.StopCh)
					if err := xray.IsActive(); err != nil {
						w, _, e2 := dynamic_util.DetectWorkload(
							c.OperatorConfig.ClientConfig,
							core.SchemeGroupVersion.WithResource("pods"),
							os.Getenv("MY_POD_NAMESPACE"),
							os.Getenv("MY_POD_NAME"))
						if e2 == nil {
							eventer.CreateEventWithLog(
								kubernetes.NewForConfigOrDie(c.OperatorConfig.ClientConfig),
								"kubedb-operator",
								w,
								core.EventTypeWarning,
								eventer.EventReasonAdmissionWebhookNotActivated,
								err.Error())
						}
						panic(err)
					}
				}()
				return nil
			},
		)
	}
	return s, nil
}

func appendUniqueGroupVersion(slice []schema.GroupVersion, elems ...schema.GroupVersion) []schema.GroupVersion {
	m := map[schema.GroupVersion]bool{}
	for _, gv := range slice {
		m[gv] = true
	}
	for _, e := range elems {
		m[e] = true
	}
	out := make([]schema.GroupVersion, 0, len(m))
	for gv := range m {
		out = append(out, gv)
	}
	return out
}

func postStartHookName(hook hooks.AdmissionHook) string {
	var ns []string
	gvr, _ := hook.Resource()
	ns = append(ns, fmt.Sprintf("admit-%s.%s.%s", gvr.Resource, gvr.Version, gvr.Group))
	if len(ns) == 0 {
		return ""
	}
	return strings.Join(append(ns, "init"), "-")
}

func admissionHooksByGroupThenVersion(admissionHooks ...hooks.AdmissionHook) map[string]map[string][]hooks.AdmissionHook {
	ret := map[string]map[string][]hooks.AdmissionHook{}
	for i := range admissionHooks {
		hook := admissionHooks[i]
		gvr, _ := hook.Resource()
		group, ok := ret[gvr.Group]
		if !ok {
			group = map[string][]hooks.AdmissionHook{}
			ret[gvr.Group] = group
		}
		group[gvr.Version] = append(group[gvr.Version], hook)
	}
	return ret
}
