module kubedb.dev/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20200323182826-54e98e09185a
	github.com/prometheus-operator/prometheus-operator v0.42.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.bytebuilders.dev/license-verifier v0.3.0
	go.bytebuilders.dev/license-verifier/kubernetes v0.3.0
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/apiserver v0.18.5
	k8s.io/client-go v12.0.0+incompatible
	kmodules.xyz/client-go v0.0.0-20200917200341-3f5fe7b6c182
	kmodules.xyz/custom-resources v0.0.0-20200604135349-9e9f5c4fdba9
	kmodules.xyz/webhook-runtime v0.0.0-20200522123600-ca70a7e28ed0
	kubedb.dev/apimachinery v0.14.0-beta.2.0.20200915201356-5ddfd53ad058
	kubedb.dev/elasticsearch v0.14.0-beta.2.0.20200916003206-98c1ad832ab4
	kubedb.dev/memcached v0.7.0-beta.2.0.20200916003254-fc482bc2d868
	kubedb.dev/mongodb v0.7.0-beta.2.0.20200916003509-3c626235e51f
	kubedb.dev/mysql v0.7.0-beta.2.0.20200916004052-5162a530f835
	kubedb.dev/percona-xtradb v0.1.0-beta.2.0.20200916004222-299013484898
	kubedb.dev/pgbouncer v0.1.0-beta.2.0.20200916005131-c5fb3b0ed048
	kubedb.dev/postgres v0.14.0-beta.2.0.20200916004433-66f45a55e2fa
	kubedb.dev/proxysql v0.1.0-beta.2.0.20200916004609-4759525bce34
	kubedb.dev/redis v0.7.0-beta.2.0.20200916004859-d46d0dbdabd6
	stash.appscode.dev/apimachinery v0.11.0
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

replace github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.42.0

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.7.1

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738

replace google.golang.org/api => google.golang.org/api v0.14.0

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191115194625-c23dd37a84c9

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace k8s.io/api => github.com/kmodules/api v0.18.4-0.20200524125823-c8bc107809b9

replace k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200520235721-10b58e57a423

replace k8s.io/apiserver => github.com/kmodules/apiserver v0.18.4-0.20200521000930-14c5f6df9625

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.3

replace k8s.io/client-go => k8s.io/client-go v0.18.3

replace k8s.io/component-base => k8s.io/component-base v0.18.3

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6

replace k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200521033432-49d3646051ad
