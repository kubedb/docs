module kubedb.dev/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20190808133642-1d4ef1f1c1e0
	github.com/aws/aws-sdk-go v1.20.21 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/coreos/prometheus-operator v0.31.1
	github.com/go-openapi/spec v0.19.2 // indirect
	github.com/go-openapi/swag v0.19.4 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/ncw/swift v1.0.49 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7 // indirect
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 // indirect
	google.golang.org/grpc v1.21.2 // indirect
	gopkg.in/olivere/elastic.v5 v5.0.80 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190516231611-bf6753f2aa24
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190516230822-f89599b3f645
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.3 // indirect
	k8s.io/kubernetes v1.14.4 // indirect
	kmodules.xyz/client-go v0.0.0-20190808141354-bbb9e14f60ab
	kmodules.xyz/custom-resources v0.0.0-20190808144301-114abf10dfe2
	kmodules.xyz/webhook-runtime v0.0.0-20190808145328-4186c470d56b
	kubedb.dev/apimachinery v0.13.0-rc.0
	kubedb.dev/elasticsearch v0.0.0-20190809233624-7e56a7c023c9
	kubedb.dev/etcd v0.0.0-20190809234758-5b9c423205ad
	kubedb.dev/memcached v0.0.0-20190809235704-47444708b217
	kubedb.dev/mongodb v0.0.0-20190810000309-674e7504a93a
	kubedb.dev/mysql v0.0.0-20190810000740-0b116c16a08c
	kubedb.dev/postgres v0.0.0-20190731230519-af93201a725c
	kubedb.dev/redis v0.0.0-20190810001400-2428be41b98c
	stash.appscode.dev/stash v0.9.0-rc.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190508082252-8397d761d4b5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213
	k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
)
