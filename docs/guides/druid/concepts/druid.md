---
title: Druid CRD
menu:
  docs_{{ .version }}:
    identifier: guides-druid-concepts-druid
    name: Druid
    parent: guides-druid-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid

## What is Druid

`Druid` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Druid](https://druid.apache.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a `Druid`object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Druid Spec

As with all other Kubernetes objects, a Druid needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Druid object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid
  namespace: demo
spec:
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  metadataStorage:
    type: PostgreSQL
    name: pg-demo
    namespace: demo
    externallyManaged: true
  zookeeperRef:
    name: zk-demo
    namespace: demo
    externallyManaged: true
  authSecret:
    name: druid-admin-cred
  configSecret:
    name: druid-custom-config
  enableSSL: true
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  keystoreCredSecret:
    name: druid-keystore-cred
  deletionPolicy: DoNotTerminate
  tls:
    certificates:
      - alias: server
        secretName: druid-server-cert
      - alias: client
        secretName: druid-client-cert
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: druid-ca-issuer
  topology:
    coordinators:
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
    overlords:
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
    brokers:
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
    routers:
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
    middleManagers:
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: standard
    historicals:
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                requests:
                  cpu: 500m
                  memory: 1024Mi
                limits:
                  cpu: 700m
                  memory: 2Gi
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: standard
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 56790
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  version: 30.0.0
```

### spec.version

`spec.version` is a required field specifying the name of the [DruidVersion](/docs/guides/druid/concepts/druidversion.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `Druid` resources,

- `28.0.1`
- `30.0.0`

### spec.replicas

`spec.replicas` the number of members in Druid replicaset.

If `spec.topology` is set, then `spec.replicas` needs to be empty. Instead use `spec.topology.controller.replicas` and `spec.topology.broker.replicas`. You need to set both of them for topology clustering.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `druid` admin user. If not set, KubeDB operator creates a new Secret `{druid-object-name}-auth` for storing the password for `admin` user for each Druid object.

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Druid object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Druid object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for Druid `admin` user.

Example:

```bash
$ kubectl create secret generic druid-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "druid-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: druid-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.configSecret

`spec.configSecret` is an optional field that points to a Secret used to hold custom Druid configuration. If not set, KubeDB operator will use default configuration for Druid.

### spec.topology

`spec.topology` represents the topology configuration for Druid cluster in KRaft mode.

When `spec.topology` is set, the following fields needs to be empty, otherwise validating webhook will throw error.

- `spec.replicas`
- `spec.podTemplate`
- `spec.storage`

#### spec.topology.coordinators

`coordinators` represents configuration for coordinators node of Druid. It is a mandatory node. So, if not mentioned in the `YAML`, this node will be initialized by `KubeDB` operator.  

Available configurable fields:

- `topology.coordinators`:
    - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Druid `coordinators` pods. Defaults to `1`.
    - `suffix` (`: "coordinators"`) - is an `optional` field that is added as the suffix of the coordinators PetSet name. Defaults to `coordinators`.
    - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `coordinators` pods.

#### spec.topology.overlords

`overlords` represents configuration for overlords node of Druid. It is an optional node. So, it is only going to be deployed by the `KubeDB` operator if explicitly mentioned in the `YAML`. Otherwise, `coordinators` node will act as `overlords`. 

Available configurable fields:

- `topology.overlords`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Druid `overlords` pods. Defaults to `1`.
  - `suffix` (`: "overlords"`) - is an `optional` field that is added as the suffix of the overlords PetSet name. Defaults to `overlords`.
  - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `overlords` pods.

#### spec.topology.brokers

`brokers` represents configuration for brokers node of Druid. It is a mandatory node. So, if not mentioned in the `YAML`, this node will be initialized by `KubeDB` operator.

Available configurable fields:

- `topology.brokers`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Druid `brokers` pods. Defaults to `1`.
  - `suffix` (`: "brokers"`) - is an `optional` field that is added as the suffix of the brokers PetSet name. Defaults to `brokers`.
  - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `brokers` pods.

#### spec.topology.routers

`routers` represents configuration for routers node of Druid. It is an optional node. So, it is only going to be deployed by the `KubeDB` operator if explicitly mentioned in the `YAML`. Otherwise, `coordinators` node will act as `routers`.

Available configurable fields:

- `topology.routers`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Druid `routers` pods. Defaults to `1`.
  - `suffix` (`: "routers"`) - is an `optional` field that is added as the suffix of the routers PetSet name. Defaults to `routers`.
  - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `routers` pods.

#### spec.topology.historicals

`historicals` represents configuration for historicals node of Druid. It is a mandatory node. So, if not mentioned in the `YAML`, this node will be initialized by `KubeDB` operator.  

Available configurable fields:

- `topology.historicals`:
    - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Druid `historicals` pods. Defaults to `1`.
    - `suffix` (`: "historicals"`) - is an `optional` field that is added as the suffix of the controller PetSet name. Defaults to `historicals`.
    - `storage` is a `required` field that specifies how much storage to claim for each of the `historicals` pods.
    - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `historicals` pods.

#### spec.topology.middleManagers

`middleManagers` represents configuration for middleManagers node of Druid. It is a mandatory node. So, if not mentioned in the `YAML`, this node will be initialized by `KubeDB` operator.

Available configurable fields:

- `topology.middleManagers`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the dedicated Druid `middleManagers` pods. Defaults to `1`.
  - `suffix` (`: "middleManagers"`) - is an `optional` field that is added as the suffix of the controller PetSet name. Defaults to `middleManagers`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `middleManagers` pods.
  - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `middleManagers` pods.


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
      name: druid-issuer
    certificates:
    - alias: server
      privateKey:
        encoding: PKCS8
      secretName: druid-client-cert
      subject:
        organizations:
        - kubedb
    - alias: http
      privateKey:
        encoding: PKCS8
      secretName: druid-server-cert
      subject:
        organizations:
        - kubedb
```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Druid. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
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


### spec.<historicals/middleManagers>.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Druid cluster using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume.

### spec.<historicals/middleManagers>.storage

If you set `spec.<historicals/middleManagers>.storageType:` to `Durable`, then `spec.<historicals/middleManagers>.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.<historicals/middleManagers>.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.<historicals/middleManagers>.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.<historicals/middleManagers>.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.<historicals/middleManagers>.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.monitor

Druid managed by KubeDB can be monitored with Prometheus operator out-of-the-box. To learn more,
- [Monitor Apache Druid with Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md)
- [Monitor Apache Druid with Built-in Prometheus](/docs/guides/druid/monitoring/using-builtin-prometheus.md)

### spec.<node-name>.podTemplate

KubeDB allows providing a template for database pod through `spec.<node-name>.podTemplate`. KubeDB operator will pass the information provided in `spec.<node-name>.podTemplate` to the PetSet created for Druid cluster.

KubeDB accept following fields to set in `spec.<node-name>.podTemplate:`

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
Uses of some field of `spec.<node-name>.podTemplate` is described below,

#### spec.<node-name>.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.<node-name>.podTemplate.spec.volumes

The `spec.<node-name>.podTemplate.<node-name>.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.<node-name>.podTemplate.spec.podPlacementPolicy

`spec.<node-name>.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the podPlacementPolicy. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.

#### spec.<node-name>.podTemplate.spec.nodeSelector

`spec.<node-name>.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

### spec.serviceTemplates

You can also provide template for the services created by KubeDB operator for Druid cluster through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
    - `stats` for is used for the `exporter` service identification.

Druid comes with four services for `coordinators`, `overlords`, `routers` and `brokers`. There are two options for providing serviceTemplates:
  - To provide `serviceTemplates` for a specific service, the `serviceTemplates.ports.port` should be equal to the port of that service and `serviceTemplate` will be used for that particular service only.
  - However, to provide a common `serviceTemplates`, `serviceTemplates.ports.port` should be empty.

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


#### spec.<node-name>.podTemplate.spec.containers

The `spec.<node-name>.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.<node-name>.podTemplate.spec.containers[].name
The `spec.<node-name>.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.<node-name>.podTemplate.spec.containers[].args
`spec.<node-name>.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.<node-name>.podTemplate.spec.containers[].env

`spec.<node-name>.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the Redis containers.

##### spec.<node-name>.podTemplate.spec.containers[].resources

`spec.<node-name>.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Druid` crd or which resources KubeDB should keep or delete when you delete `Druid` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- WipeOut
- Halt
- Delete

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run Apache Druid cluster [here](/docs/guides/druid/README.md).
- Deploy [dedicated topology cluster](/docs/guides/druid/clustering/guide/index.md) for Apache Druid
- Monitor your Druid cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- Detail concepts of [DruidVersion object](/docs/guides/druid/concepts/druidversion.md).

[//]: # (- Learn to use KubeDB managed Druid objects using [CLIs]&#40;/docs/guides/druid/cli/cli.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
