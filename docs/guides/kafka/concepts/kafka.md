---
title: Kafka CRD
menu:
  docs_{{ .version }}:
    identifier: kf-kafka-concepts
    name: Kafka
    parent: kf-concepts-kafka
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka

## What is Kafka

`Kafka` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Kafka](https://kafka.apache.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a `Kafka`object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Kafka Spec

As with all other Kubernetes objects, a Kafka needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Kafka object.

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka
  namespace: demo
spec:
  disableSecurity: false
  tieredStorage:
    provider: s3
    s3:
      bucket: kafka
      endpoint: <s3-endpoint>
      region: us-east-1
      secretName: aws-secret
      prefix: tiered-storage-demo/
    storageManagerClassName: io.aiven.kafka.tieredstorage.RemoteStorageManager
    storageManagerClassPath: /opt/kafka/libs/tiered-plugins/*
  authSecret:
    kind: Secret
    name: kafka-auth
  configuration:
    secretName: kafka-custom-config
    inline:
      broker.properties: |
        log.retention.hours=168
        log.segment.bytes=1073741824
      controller.properties: |
        log.retention.hours=168
        log.segment.bytes=1073741824
      server.properties: |
        log.retention.hours=168
        log.segment.bytes=1073741824
  enableSSL: true
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  keystoreCredSecret:
    kind: Secret
    name: kafka-keystore-cred
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
        thisLabel: willGoToSts
  storageType: Durable
  deletionPolicy: DoNotTerminate
  tls:
    certificates:
      - alias: server
        secretName: kafka-server-cert
      - alias: client
        secretName: kafka-client-cert
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: kafka-ca-issuer
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: standard
    controller:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 56790
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  version: 4.0.0
```

### spec.version

`spec.version` is a required field specifying the name of the [KafkaVersion](/docs/guides/kafka/concepts/kafkaversion.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `Kafka` resources,

- `3.5.2`
- `3.6.1`
- `3.7.2`
- `3.8.1`
- `3.9.0`
- `4.0.0`

### spec.replicas

`spec.replicas` the number of members in Kafka replicaset.

If `spec.topology` is set, then `spec.replicas` needs to be empty. Instead use `spec.topology.controller.replicas` and `spec.topology.broker.replicas`. You need to set both of them for topology clustering.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.disableSecurity

`spec.disableSecurity` is an optional field that specifies whether to disable security for Kafka cluster which means no authentication and authorization will be enabled. All the Kafka Brokers and controllers will be set to spin up with `security.protocol=PLAINTEXT` configuration. The default value of this field is `false`.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `kafka` admin user. If not set, KubeDB operator creates a new Secret `{kafka-object-name}-auth` for storing the password for `admin` user for each Kafka object.

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Kafka object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
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

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for Kafka `admin` user.

Example:

```bash
$ kubectl create secret generic kf-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "kf-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: kf-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.configuration.secretName

`spec.configuration.secretName` is an optional field that specifies the name of the secret containing the custom configuration for the ConnectCluster. The secret should contain a key `config.properties` which contains the custom configuration for the ConnectCluster. The default value of this field is `nil`.
```yaml
configuration:
  secretName: <custom-config-secret-name>
```

> **Note**: Use `.spec.configuration.secretName` to specify the name of the secret instead of `.spec.configSecret.name` The field `.spec.configsecret` is deprecated and will be removed in future releases. If you still use `.spec.configSecret`, KubeDB will copy `.spec.configuSecret.name` to `.spec.configuration.secretName` internally.

### spec.configuration
`spec.configuration` is an optional field that specifies custom configuration for Kafka cluster. It has the following fields:
- `configuration.secretName` is an optional field that specifies the name of the secret that holds custom configuration files for Kafka cluster.
- `configuration.inline` is an optional field that allows you to provide custom configuration directly in the Kafka object. It has the following possible keys:
    - `broker.properties` - is used to provide custom configuration for Kafka brokers.
    - `controller.properties` - is used to provide custom configuration for Kafka controllers.
    - `server.propertbies` - is used to provide custom configuration for both Kafka brokers and controllers.
    - ```yaml
      spec:
        configuration:
          inline:
            config.properties: |
              connector.class=com.mongodb.*
              tasks.max=1
              topic.prefix=mongodb-
              connection.uri=mongodb://mongo-user:mongo-password@mongo-host:27017
      ```

### spec.tieredStorage
`spec.tieredStorage` is an optional field that specifies the tiered storage configuration for Kafka. Tiered storage allows Kafka to offload older data to cheaper storage solutions like S3, Azure Blob Storage, GCS, or local storage, while keeping recent data on faster local storage. Tiered storage helps in reducing the cost of storage and improves the performance of Kafka by keeping the frequently accessed data on local storage. Following are the fields of `spec.tieredStorage`:
- `provider` ( `string` | `""` ) - is a field that specifies the tiered storage provider. Supported providers are `s3`, `azure`, `gcs`, and `local`.
- `s3` ( `object` | `nil` ) - is a field that specifies the S3 tiered storage configuration. It has the following fields:
    - `bucket` ( `string` | `""` ) - is a required field that specifies the S3 bucket name.
    - `endpoint` ( `string` | `""` ) - is an optional field that specifies the S3 endpoint.
    - `region` ( `string` | `""` ) - is a required field that specifies the S3 region.
    - `secretName` ( `string` | `""` ) - is a required field that specifies the name of the secret that contains the S3 credentials.
    - `prefix` ( `string` | `""` ) - is an optional field that specifies the prefix for the S3 objects.
- `azure` ( `object` | `nil` ) - is a field that specifies the Azure Blob Storage tiered storage configuration. It has the following fields:
    - `container` ( `string` | `""` ) - is a required field that specifies the Azure Blob Storage container name.
    - `storageAccount` ( `string` | `""` ) - is a required field that specifies the Azure storage account name.
    - `secretName` ( `string` | `""` ) - is a required field that specifies the name of the secret that contains the Azure Blob Storage credentials.
    - `prefix` ( `string` | `""` ) - is an optional field that specifies the prefix for the Azure Blob Storage objects.
- `gcs` ( `object` | `nil` ) - is a field that specifies the GCS tiered storage configuration. It has the following fields:
    - `bucket` ( `string` | `""` ) - is a required field that specifies the GCS bucket name.
    - `secretName` ( `string` | `""` ) - is a required field that specifies the name of the secret that contains the GCS credentials.
    - `prefix` ( `string` | `""` ) - is an optional field that specifies the prefix for the GCS objects.
- `local` ( `object` | `nil` ) - is a field that specifies the local tiered storage configuration. It has the following fields:
    - `mountPath` ( `string` | `""` ) - is a required field that specifies the mount path for the local storage.
    - `subPath` ( `string` | `""` ) - is an optional field that specifies the sub-path for the local storage.
- `storageManagerClassName` ( `string` | `""` ) - is an optional field that specifies the storage manager class name for the tiered storage. Default is `io.aiven.kafka.tieredstorage.RemoteStorageManager`.

- `storageManagerClassPath` ( `string` | `""` ) - is an optional field that specifies the class path where the storage manager implementation JARs are located. Example: `/opt/kafka/libs/tiered-plugins/*`.


### spec.topology

`spec.topology` represents the topology configuration for Kafka cluster in KRaft mode.

When `spec.topology` is set, the following fields needs to be empty, otherwise validating webhook will throw error.

- `spec.replicas`
- `spec.podTemplate`
- `spec.storage`

#### spec.topology.broker

`broker` represents configuration for brokers of Kafka. In KRaft Topology mode clustering each pod can act as a single dedicated Kafka broker.

Available configurable fields:

- `topology.broker`:
    - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Kafka `broker` pods. Defaults to `1`.
    - `suffix` (`: "broker"`) - is an `optional` field that is added as the suffix of the broker PetSet name. Defaults to `broker`.
    - `storage` is a `required` field that specifies how much storage to claim for each of the `broker` pods.
    - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `broker` pods.

#### spec.topology.controller

`controller` represents configuration for controllers of Kafka. In KRaft Topology mode clustering each pod can act as a single dedicated Kafka controller that preserves metadata for the whole cluster and participated in leader election.

Available configurable fields:

- `topology.controller`:
    - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Kafka `controller` pods. Defaults to `1`.
    - `suffix` (`: "controller"`) - is an `optional` field that is added as the suffix of the controller PetSet name. Defaults to `controller`.
    - `storage` is a `required` field that specifies how much storage to claim for each of the `controller` pods.
    - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `controller` pods.

### spec.enableSSL

`spec.enableSSL` is an `optional` field that specifies whether to enable TLS to HTTP layer. The default value of this field is `false`.

```yaml
spec:
  enableSSL: true 
```

### spec.tls

`spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). Currently, the operator only supports the `PKCS#8` encoded certificates.

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: kf-issuer
    certificates:
    - alias: server
      privateKey:
        encoding: PKCS8
      secretName: kf-client-cert
      subject:
        organizations:
        - kubedb
    - alias: http
      privateKey:
        encoding: PKCS8
      secretName: kf-server-cert
      subject:
        organizations:
        - kubedb
```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Kafka. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
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

      For more details, visit [here](https://golang.org/pkg/crypto/x509/pkix/#Name).

    - `duration` ( `string` | `""` ) - is the period during which the certificate is valid. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300m"`, `"1.5h"` or `"20h45m"`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    - `renewBefore` ( `string` | `""` ) - is a specifiable time before expiration duration.
    - `dnsNames` ( `[]string` | `nil` ) - is a list of subject alt names.
    - `ipAddresses` ( `[]string` | `nil` ) - is a list of IP addresses.
    - `uris` ( `[]string` | `nil` ) - is a list of URI Subject Alternative Names.
    - `emailAddresses` ( `[]string` | `nil` ) - is a list of email Subject Alternative Names.


### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Kafka cluster using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume.

### spec.storage

If you set `spec.storageType:` to `Durable`, then `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

NB. If `spec.topology` is set, then `spec.storage` needs to be empty. Instead use `spec.topology.<controller/broker>.storage`

### spec.monitor

Kafka managed by KubeDB can be monitored with Prometheus operator out-of-the-box. To learn more,
- [Monitor Apache Kafka with Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md)
- [Monitor Apache Kafka with Built-in Prometheus](/docs/guides/kafka/monitoring/using-builtin-prometheus.md)

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Kafka cluster.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
    - annotations (pod's annotation)
    - labels (pod's labels)
- controller:
    - annotations (petset's annotation)
    - labels (petset's labels)
- spec:
    - containers
    - volumes
    - podPlacementPolicy
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

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go#L26C1-L279C1).
Uses of some field of `spec.podTemplate` is described below,

NB. If `spec.topology` is set, then `spec.podTemplate` needs to be empty. Instead use `spec.topology.<controller/broker>.podTemplate`

#### spec.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.podTemplate.spec.volumes

The `spec.podTemplate.spec.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the `podPlacementPolicy`. `name` of the podPlacementPolicy is referred under this attribute. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.
```yaml
spec:
  podPlacementPolicy:
    name: default
```

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

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


#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the Redis containers.

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Kafka` crd or which resources KubeDB should keep or delete when you delete `Kafka` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- WipeOut
- Halt
- Delete

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

> For more details you can visit [here](https://appscode.com/blog/post/deletion-policy/)

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run Apache Kafka cluster [here](/docs/guides/kafka/README.md).
- Deploy [dedicated topology cluster](/docs/guides/kafka/clustering/topology-cluster/index.md) for Apache Kafka
- Deploy [combined cluster](/docs/guides/kafka/clustering/combined-cluster/index.md) for Apache Kafka
- Monitor your Kafka cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Detail concepts of [KafkaVersion object](/docs/guides/kafka/concepts/kafkaversion.md).
- Learn to use KubeDB managed Kafka objects using [CLIs](/docs/guides/kafka/cli/cli.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
