module github.com/kubedb/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20190523031839-1468ee3a76e8
	github.com/coreos/prometheus-operator v0.29.0
	github.com/kubedb/apimachinery v0.0.0-20190526014453-48e4bab67179
	github.com/kubedb/elasticsearch v0.0.0-20190508234318-e8fbea4bd0cc
	github.com/kubedb/etcd v0.0.0-20190509011751-2cb55f3a1759
	github.com/kubedb/memcached v0.0.0-20190509011145-4a80b9afbbb3
	github.com/kubedb/mongodb v0.0.0-20190508235556-f4167b84b5fa
	github.com/kubedb/mysql v0.0.0-20190507122034-73ad7c30b884
	github.com/kubedb/postgres v0.0.0-20190508232535-7e69d665c1ad
	github.com/kubedb/redis v0.0.0-20190509010457-3699dfb2e19d
	github.com/ncw/swift v1.0.47 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/pflag v1.0.3
	gopkg.in/olivere/elastic.v5 v5.0.80 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190508224317-421cff06bf05
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190508223931-4756b09d7af2
	k8s.io/client-go v11.0.0+incompatible
	kmodules.xyz/client-go v0.0.0-20190524133821-9c8a87771aea
	kmodules.xyz/custom-resources v0.0.0-20190508103408-464e8324c3ec
	kmodules.xyz/webhook-runtime v0.0.0-20190508094945-962d01212c5b
)

replace (
	github.com/graymeta/stow => github.com/appscode/stow v0.0.0-20190506085026-ca5baa008ea3
	gopkg.in/robfig/cron.v2 => github.com/appscode/cron v0.0.0-20170717094345-ca60c6d796d4
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
