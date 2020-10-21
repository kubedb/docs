module kubedb.dev/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20201006035845-a0302ac8e3d3
	github.com/prometheus-operator/prometheus-operator v0.42.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.bytebuilders.dev/license-verifier v0.3.0
	go.bytebuilders.dev/license-verifier/kubernetes v0.3.0
	k8s.io/api v0.18.9
	k8s.io/apiextensions-apiserver v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/apiserver v0.18.9
	k8s.io/client-go v12.0.0+incompatible
	kmodules.xyz/client-go v0.0.0-20201021051118-03dac1aea508
	kmodules.xyz/custom-resources v0.0.0-20201008012351-6d8090f759d4
	kmodules.xyz/webhook-runtime v0.0.0-20200922211931-8337935590de
	kubedb.dev/apimachinery v0.14.0-beta.3.0.20201021115037-028d939d696f
	kubedb.dev/elasticsearch v0.14.0-beta.3.0.20201019183940-c22b7f3193c0
	kubedb.dev/memcached v0.7.0-beta.3.0.20201019123808-40afd78dc5cc
	kubedb.dev/mongodb v0.7.0-beta.3.0.20201019184118-7e7a960e5557
	kubedb.dev/mysql v0.7.0-beta.3.0.20201019175223-09d4743dd9e4
	kubedb.dev/percona-xtradb v0.1.0-beta.3.0.20201019181251-60f7e5a915ef
	kubedb.dev/pgbouncer v0.1.0-beta.3.0.20201019173229-d5fa2ce7d09f
	kubedb.dev/postgres v0.14.0-beta.3.0.20201019130757-0df8a375f9a2
	kubedb.dev/proxysql v0.1.0-beta.3.0.20201019174009-d2f326c74b92
	kubedb.dev/redis v0.7.0-beta.3.0.20201019181236-2e2f2d7b154d
)

replace bitbucket.org/ww/goautoneg => gomodules.xyz/goautoneg v0.0.0-20120707110453-a547fc61f48d

replace cloud.google.com/go => cloud.google.com/go v0.49.0

replace git.apache.org/thrift.git => github.com/apache/thrift v0.13.0

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v35.0.0+incompatible

replace github.com/Azure/go-ansiterm => github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible

replace github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0

replace github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0

replace github.com/Azure/go-autorest/autorest/azure/auth => github.com/Azure/go-autorest/autorest/azure/auth v0.2.0

replace github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.1.0

replace github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.2.0

replace github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.2.0

replace github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.1.0

replace github.com/Azure/go-autorest/logger => github.com/Azure/go-autorest/logger v0.1.0

replace github.com/Azure/go-autorest/tracing => github.com/Azure/go-autorest/tracing v0.5.0

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.1

replace github.com/golang/protobuf => github.com/golang/protobuf v1.3.2

replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1

replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.5

replace github.com/prometheus-operator/prometheus-operator => github.com/prometheus-operator/prometheus-operator v0.42.0

replace github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.42.0

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.7.1

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738

replace google.golang.org/api => google.golang.org/api v0.14.0

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191115194625-c23dd37a84c9

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace k8s.io/api => github.com/kmodules/api v0.18.10-0.20200922195318-d60fe725dea0

replace k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200922195535-0c9a1b86beec

replace k8s.io/apiserver => github.com/kmodules/apiserver v0.18.10-0.20200922195747-1bd1cc8f00d1

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.9

replace k8s.io/client-go => github.com/kmodules/k8s-client-go v0.18.10-0.20200922201634-73fedf3d677e

replace k8s.io/component-base => k8s.io/component-base v0.18.9

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6

replace k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200922200158-8b13196d8dc4

replace k8s.io/utils => k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
