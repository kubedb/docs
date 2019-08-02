module kubedb.dev/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20190722173419-e454bf744023
	github.com/coreos/prometheus-operator v0.30.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	gopkg.in/olivere/elastic.v5 v5.0.80 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190516231611-bf6753f2aa24
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190516230822-f89599b3f645
	k8s.io/client-go v11.0.0+incompatible
	kmodules.xyz/client-go v0.0.0-20190715080709-7162a6c90b04
	kmodules.xyz/custom-resources v0.0.0-20190730174012-d0224972f055
	kmodules.xyz/webhook-runtime v0.0.0-20190715115250-a84fbf77dd30
	kubedb.dev/apimachinery v0.0.0-20190801152009-3ee2a59976e1
	kubedb.dev/elasticsearch v0.0.0-20190731222758-3dd46b3f441e
	kubedb.dev/etcd v0.0.0-20190718040006-2a7e758b7ad4
	kubedb.dev/memcached v0.0.0-20190718030029-dab1a2fca289
	kubedb.dev/mongodb v0.0.0-20190802112916-172be98da3bf
	kubedb.dev/mysql v0.0.0-20190731224757-086032c5b517
	kubedb.dev/postgres v0.0.0-20190731230519-af93201a725c
	kubedb.dev/redis v0.0.0-20190718022734-c046a9756237
	stash.appscode.dev/stash v0.0.0-20190730144328-4ec6caf83810
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
