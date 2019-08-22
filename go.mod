module kubedb.dev/operator

go 1.12

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/Azure/azure-pipeline-go v0.1.9 // indirect
	github.com/Azure/azure-storage-blob-go v0.6.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/appscode/go v0.0.0-20190808133642-1d4ef1f1c1e0
	github.com/coreos/prometheus-operator v0.31.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/go-ini/ini v1.40.0 // indirect
	github.com/go-openapi/swag v0.19.0 // indirect
	github.com/google/martian v2.1.1-0.20190517191504-25dcb96d9e51+incompatible // indirect
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/prometheus/client_golang v0.9.4 // indirect
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190516231611-bf6753f2aa24
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190516230822-f89599b3f645
	k8s.io/cli-runtime v0.0.0-20190516231937-17bc0b7fcef5 // indirect
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubernetes v1.14.2 // indirect
	kmodules.xyz/client-go v0.0.0-20190808141354-bbb9e14f60ab
	kmodules.xyz/custom-resources v0.0.0-20190808144301-114abf10dfe2
	kmodules.xyz/webhook-runtime v0.0.0-20190808145328-4186c470d56b
	kubedb.dev/apimachinery v0.13.0-rc.0
	kubedb.dev/elasticsearch v0.0.0-20190822064249-61e89f4b8fdf
	kubedb.dev/etcd v0.0.0-20190809234758-5b9c423205ad
	kubedb.dev/memcached v0.0.0-20190822064902-85f05095a37f
	kubedb.dev/mongodb v0.0.0-20190822064946-6966e187de49
	kubedb.dev/mysql v0.0.0-20190822065357-1bd9d987e90f
	kubedb.dev/postgres v0.0.0-20190822070154-05b7ef5aa33e
	kubedb.dev/redis v0.0.0-20190822065252-2960649447ae
	stash.appscode.dev/stash v0.9.0-rc.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.4.2+incompatible
	gomodules.xyz/envsubst => gomodules.xyz/envsubst v0.1.0
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190811223248-5a95b2df4348
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213
	k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)
