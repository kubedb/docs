---
title: ConnectCluster CRD
menu:
  docs_{{ .version }}:
    identifier: kf-connectcluster-concepts
    name: ConnectCluster
    parent: kf-concepts-kafka
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ConnectCluster

## What is ConnectCluster

`ConnectCluster` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [ConnectCluster](https://kafka.apache.org/) in a Kubernetes native way. You only need to describe the desired configuration in a `ConnectCluster` object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## ConnectCluster Spec

As with all other Kubernetes objects, a ConnectCluster needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example ConnectCluster object.

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: ConnectCluster
metadata:
  name: connectcluster
  namespace: demo
spec:
  version: 3.9.0
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  disableSecurity: false
  authSecret:
    name: connectcluster-auth
  enableSSL: true
  keystoreCredSecret:
    name: connectcluster-keystore-cred
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: connectcluster-ca-issuer
    certificates:
      - alias: server
        secretName: connectcluster-server-cert
      - alias: client
        secretName: connectcluster-client-cert
  configSecret:
    name: custom-connectcluster-config
  replicas: 3
  connectorPlugins:
    - gcs-0.13.0
    - mongodb-1.14.1
    - mysql-3.0.5.final
    - postgres-3.0.5.final
    - s3-2.15.0
    - jdbc-3.0.5.final
  kafkaRef:
    name: kafka
    namespace: demo
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
      labels:
        thisLabel: willGoToPod
    controller:
      annotations:
        passMe: ToPetSet
      labels:
        thisLabel: willGoToPetSet
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 56790
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

### spec.version

`spec.version` is a required field specifying the name of the [KafkaVersion](/docs/guides/kafka/concepts/kafkaversion.md) CR where the docker images are specified. Currently, when you install KubeDB, it creates the following `KafkaVersion` resources,

- `3.5.2`
- `3.6.1`
- `3.7.2`
- `3.8.1`
- `3.9.0`

### spec.replicas

`spec.replicas` the number of worker nodes in ConnectCluster.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.disableSecurity

`spec.disableSecurity` is an optional field that specifies whether to disable all kind of security features like basic authentication and tls. The default value of this field is `false`.

### spec.connectorPlugins

`spec.connectorPlugins` is an optional field that specifies the list of connector plugins to be installed in the ConnectCluster worker node. The field takes a list of strings where each string represents the name of the KafkaConnectorVersion CR. To learn more about KafkaConnectorVersion CR, visit [here](/docs/guides/kafka/concepts/kafkaconnectorversion.md).
```yaml
connectorPlugins:
  - <connector-plugin-name-1>
  - <connector-plugin-name-2>
```

### spec.kafkaRef

`spec.kafkaRef` is a required field that specifies the name and namespace of the appbinding for `Kafka` object that the `ConnectCluster` object is associated with.
```yaml
kafkaRef:
  name: <kafka-object-appbinding-name>
  namespace: <kafka-object-appbinding-namespace>
```

### spec.configSecret

`spec.configSecret` is an optional field that specifies the name of the secret containing the custom configuration for the ConnectCluster. The secret should contain a key `config.properties` which contains the custom configuration for the ConnectCluster. The default value of this field is `nil`.
```yaml
configSecret:
  name: <custom-config-secret-name>
```

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `ConnectCluster` username and password. If not set, KubeDB operator creates a new Secret `{connectcluster-object-name}-connect-cred` for storing the username and password for each ConnectCluster object.

We can use this field in 3 mode.

1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the ConnectCluster object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Kafka object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for ConnectCluster user.

Example:

```bash
$ kubectl create secret generic kcc-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "kcc-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: kcc-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.enableSSL

`spec.enableSSL` is an `optional` field that specifies whether to enable TLS to HTTP layer. The default value of this field is `false`.

```yaml
spec:
  enableSSL: true 
```

### spec.keystoreCredSecret

`spec.keystoreCredSecret` is an `optional` field that specifies the name of the secret containing the keystore credentials for the ConnectCluster. The secret should contain three keys `ssl.keystore.password`, `ssl.key.password` and `ssl.keystore.password`. The default value of this field is `nil`.

```yaml
spec:
  keystoreCredSecret:
    name: <keystore-cred-secret-name>
```

### spec.tls

`spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). Currently, the operator only supports the `PKCS#8` encoded certificates.

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: kcc-issuer
    certificates:
    - alias: server
      privateKey:
        encoding: PKCS8
      secretName: kcc-client-cert
      subject:
        organizations:
        - kubedb
    - alias: http
      privateKey:
        encoding: PKCS8
      secretName: kcc-server-cert
      subject:
        organizations:
        - kubedb
```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for ConnectCluster. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
    - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
    - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
    - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.

- `tls.certificates` - is an `optional` field that specifies a list of certificate configurations used to configure the  certificates. It has the following fields:
    - `alias` - represents the identifier of the certificate. It has the following possible value:
        - `server` - is used for the server certificate configuration.
        - `client` - is used for the client certificate configuration.

    - `secretName` - ( `string` | `"<database-name>-alias-cert"` ) - specifies the k8s secret name that holds the certificates.

    - `subject` - specifies an `X.509` distinguished name (DN). It has the following configurable fields:
        - `organizations` ( `[]string` | `nil` ) - is a list of organization names.
        - `organizationalUnits` ( `[]string` | `nil` ) - is a list of organization unit names.
        - `countries` ( `[]string` | `nil` ) -  is a list of country names (ie. Country Codes).
        - `localities` ( `[]string` | `nil` ) - is a list of locality names.
        - `provinces` ( `[]string` | `nil` ) - is a list of province names.
        - `streetAddresses` ( `[]string` | `nil` ) - is a list of street addresses.
        - `postalCodes` ( `[]string` | `nil` ) - is a list of postal codes.
        - `serialNumber` ( `string` | `""` ) is a serial number.

      For more details, visit [here](https://pkg.go.dev/crypto/x509/pkix#Name).

    - `duration` ( `string` | `""` ) - is the period during which the certificate is valid. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300m"`, `"1.5h"` or `"20h45m"`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    - `renewBefore` ( `string` | `""` ) - is a specifiable time before expiration duration.
    - `dnsNames` ( `[]string` | `nil` ) - is a list of subject alt names.
    - `ipAddresses` ( `[]string` | `nil` ) - is a list of IP addresses.
    - `uris` ( `[]string` | `nil` ) - is a list of URI Subject Alternative Names.
    - `emailAddresses` ( `[]string` | `nil` ) - is a list of email Subject Alternative Names.

    

### spec.monitor

ConnectCluster managed by KubeDB can be monitored with Prometheus operator out-of-the-box. To learn more,
- [Monitor Apache with Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md)

### spec.podTemplate

KubeDB allows providing a template for pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for ConnectCluster.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
    - annotations (pod's annotation)
    - labels (pod's labels)
- controller:
    - annotations (petset's annotation)
    - labels (petset's labels)
- spec:
    - volumes
    - initContainers
    - containers
    - imagePullSecrets
    - nodeSelector
    - serviceAccountName
    - schedulerName
    - tolerations
    - priorityClassName
    - priority
    - securityContext

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/39bf8b2/api/v2/types.go#L44-L279). Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide template for the services created by KubeDB operator for Kafka cluster through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
    - `stats` is used for the exporter service identification.
- metadata:
    - labels
    - annotations
- spec:
    - type
    - ports
    - clusterIP
    - externalIPs
    - loadBalancerIP
    - loadBalancerSourceRanges
    - externalTrafficPolicy
    - healthCheckNodePort
    - sessionAffinityConfig

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.21.1/api/v1/types.go#L237) to understand these fields in detail.

### spec.deletionPolicy

`spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `ConnectCluster` crd or which resources KubeDB should keep or delete when you delete `ConnectCluster` crd. KubeDB provides following four deletion policies:

- Delete
- DoNotTerminate
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run a Apache Kafka Connect cluster [here](/docs/guides/kafka/README.md).
- Monitor your ConnectCluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Detail concepts of [KafkaConnectorVersion object](/docs/guides/kafka/concepts/kafkaconnectorversion.md).
- Learn to use KubeDB managed Kafka objects using [CLIs](/docs/guides/kafka/cli/cli.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
