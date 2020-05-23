module kubedb.dev/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20200323182826-54e98e09185a
	github.com/coreos/prometheus-operator v0.39.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/apiserver v0.18.3
	k8s.io/client-go v12.0.0+incompatible
	kmodules.xyz/client-go v0.0.0-20200522120609-c6430d66212f
	kmodules.xyz/custom-resources v0.0.0-20200521070540-2221c4957ef6
	kmodules.xyz/webhook-runtime v0.0.0-20200522123600-ca70a7e28ed0
	kubedb.dev/apimachinery v0.14.0-alpha.1
	kubedb.dev/elasticsearch v0.13.0-rc.1.0.20200523034244-08c1d2a8b229
	kubedb.dev/memcached v0.6.0-rc.1.0.20200522170052-6ed07efc99e5
	kubedb.dev/mongodb v0.6.0-rc.1.0.20200523013554-d6d87e16bfd6
	kubedb.dev/mysql v0.6.0-rc.0.0.20200523003329-5df90daa5e53
	kubedb.dev/percona-xtradb v0.0.0-20200523050450-e81d2b4cbcd9
	kubedb.dev/pgbouncer v0.0.0-20200523021050-ef7fe4752358
	kubedb.dev/postgres v0.13.0-rc.0.0.20200523040826-6ce6deb18394
	kubedb.dev/proxysql v0.0.0-20200523013326-4f5bea8df303
	kubedb.dev/redis v0.6.0-rc.0.0.20200522170523-bf072134923f
	stash.appscode.dev/apimachinery v0.9.0-rc.6.0.20200522135619-e81205a3590e
)

replace (
	bitbucket.org/ww/goautoneg => gomodules.xyz/goautoneg v0.0.0-20120707110453-a547fc61f48d
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v35.0.0+incompatible
	github.com/Azure/go-ansiterm => github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0
	github.com/Azure/go-autorest/autorest/azure/auth => github.com/Azure/go-autorest/autorest/azure/auth v0.2.0
	github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.1.0
	github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.2.0
	github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.2.0
	github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.1.0
	github.com/Azure/go-autorest/logger => github.com/Azure/go-autorest/logger v0.1.0
	github.com/Azure/go-autorest/tracing => github.com/Azure/go-autorest/tracing v0.5.0
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.0.0
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200520235721-10b58e57a423
	k8s.io/apiserver => github.com/kmodules/apiserver v0.18.4-0.20200521000930-14c5f6df9625
	k8s.io/client-go => k8s.io/client-go v0.18.3
	k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200521033432-49d3646051ad
)
