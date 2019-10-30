module kubedb.dev/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20191025021232-311ac347b3ef
	github.com/coreos/prometheus-operator v0.31.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190516231611-bf6753f2aa24
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190516230822-f89599b3f645
	k8s.io/client-go v11.0.0+incompatible
	kmodules.xyz/client-go v0.0.0-20191023042933-b12d1ccfaf57
	kmodules.xyz/custom-resources v0.0.0-20190927035424-65fe358bb045
	kmodules.xyz/webhook-runtime v0.0.0-20190808145328-4186c470d56b
	kubedb.dev/apimachinery v0.13.0-rc.2
	kubedb.dev/elasticsearch v0.13.0-rc.1.0.20191030000328-b7012bc5d210
	kubedb.dev/etcd v0.5.0-rc.1.0.20191030024915-cf75e74fc547
	kubedb.dev/memcached v0.6.0-rc.1.0.20191030000830-7b79fbfed004
	kubedb.dev/mongodb v0.6.0-rc.1.0.20191030003732-cc545e04a884
	kubedb.dev/mysql v0.6.0-rc.0.0.20191030004636-132b2a0eff45
	kubedb.dev/pgbouncer v0.0.0-20191030130947-6c1a78a0ea78
	kubedb.dev/postgres v0.13.0-rc.0.0.20191030004721-aa1d98d07b28
	kubedb.dev/redis v0.6.0-rc.0.0.20191030003703-6ccd651c0777
	stash.appscode.dev/stash v0.9.0-rc.2
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.34.0
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
