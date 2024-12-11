---
title: Solr CRD
menu:
  docs_{{ .version }}:
    identifier: sl-solr-concepts
    name: Solr
    parent: sl-concepts-solr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr

## What is Solr

`Solr` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Solr](https://solr.apache.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a Solr object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Solr Spec

As with all other Kubernetes objects, a Solr needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Solr object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  authConfigSecret:
    name: solr-cluster-auth-config
  authSecret:
    name: solr-cluster-admin-cred
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  serviceTemplates:
    - alias: primary
      metadata:
        annotations:
          passMe: ToService
      spec:
        type: NodePort
        ports:
          - name:  http
            port:  8983
  storageType: Durable
  deletionPolicy: Delete
  topology:
    coordinator:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: coordinator
    data:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: data
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: overseer
  version: 9.4.1
  zookeeperDigestReadonlySecret:
    name: solr-cluster-zk-digest-readonly
  zookeeperDigestSecret:
    name: solr-cluster-zk-digest
  zookeeperRef:
    name: zk-com
    namespace: demo
```


### spec.version

`spec.version` is a required field specifying the name of the [SolrVersion](/docs/guides/solr/concepts/solrversion.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `SolrVersion` crds,

-  `8.11.2`
-  `9.4.1`

### spec.disableSecurity

`spec.disableSecurity` is an optional field that decides whether Solr instance will be secured by auth or no.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `Solr` superuser. If not set, KubeDB operator creates a new Secret `{Solr-object-name}-admin-cred` for storing the password for `Solr` superuser.

We can use this field in 3 mode.

1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Solr object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```
2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Solr object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `username` key and a `password` key which contains the `username` and `password` respectively for `Solr` superuser.

Example:

```bash
$ kubectl create secret generic solr-cluster-admin0-cred -n demo \
--from-literal=username=admin \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "solr-cluster-admin-cred" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: solr-cluster-admin-cred
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.topology

`spec.topology` is an `optional` field that provides a way to configure different types of nodes for the Solr cluster. This field enables you to specify how many nodes you want to act as `overseer`, `data`, `coordinator` or other node roles for Elasticsearch. You can also specify how much storage and resources to allocate for each type of node independently.

Currently supported node types are -
- **data**: Data nodes hold the shards that contain the documents you have indexed. Data nodes handle data related operations like CRUD, search, and aggregations
- **overseer**: Overseer nodes can execute shard distributions, composed of one or more overseer processors
- **coordinator**: The coordinator node can act as if it has replicas of all collections in the cluster when a query is performed.

```yaml
  topology:
    data:
      replicas: 3
      podTemplate: 
        spec:
          containers:
            - name: "solr"
              resources:
                requests:
                  cpu: "900m"
                limits:
                  cpu: "900m"
                  memory: "2.5Gi"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: data
    overseer:
      replicas: 3
      podTemplate:
        spec:
          containers:
            - name: "solr"
              resources:
                requests:
                  cpu: "900m"
                limits:
                  cpu: "900m"
                  memory: "2.5Gi"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: overseer
    coordinator:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: "solr"
              resources:
                requests:
                  cpu: "900m"
                limits:
                  cpu: "900m"
                  memory: "2.5Gi"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: coordinator
```

The `spec.topology` contains the following fields:

- `topology.overseer`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `overseer` nodes. Defaults to `1`.
  - `suffix` (`: "overseer"`) - is an `optional` field that is added as the suffix of the overseer PetSet name. Defaults to `overseer`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `overseer` nodes.
  - `resources` (`: "cpu: 900m, memory: 2Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `overseer` nodes.

- `topology.data`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `data` nodes. Defaults to `1`.
  - `suffix` (`: "data"`) - is an `optional` field that is added as the suffix of the data PetSet name. Defaults to `data`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `data` nodes.
  - `resources` (` cpu: 900m, memory: 2Gi `) - is an `optional` field that specifies which amount of computational resources to request or to limit for each of the `data` nodes.

- `topology.coordinator`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `coordinator` nodes. Defaults to `1`.
  - `suffix` (`: "coordinator"`) - is an `optional` field that is added as the suffix of the data PetSet name. Defaults to `coordinator`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `coordinator` nodes.
  - `resources` (` cpu: 900m, memory: 2Gi `) - is an `optional` field that specifies which amount of computational resources to request or to limit for each of the `data` nodes.

> Note: Any two types of nodes can't have the same `suffix`.

If you specify `spec.topology` field then you **do not need** to specify the following fields in Elasticsearch CRD.

- `spec.replicas`
- `spec.storage`
- `spec.podTemplate.spec.resources`

If you do not specify `spec.topology` field, the Elasticsearch Cluster runs in combined mode.

> Combined Mode: all nodes of the Elasticsearch cluster will work as `overseer`, `data` and `coordinator` nodes simultaneously.

### spec.replicas

`spec.replicas` is an `optional` field that can be used if `spec.topology` is not specified. This field specifies the number of nodes (ie. pods) in the Elasticsearch cluster. The default value of this field is `1`.

```yaml
spec:
  replicas: 3
```



### spec.zookeeperRef

Referenece of zookeeper cluster which will coordinate solr and save necessary credentials of solr cluster.

### spec.zookeeperDigestSecret

We have some zookeeper digest secret which will keep data in out zookeeper cluster safe. These secret do not guarantee security of zookeeper cluster. It just encodes solr data in the zookeeper cluster.

### spec.storage

If you set `spec.storageType:` to `Durable`, then  `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the Petset created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.solrModules

We have to enable certain modules to conduct the operations like backup and monitoring. Like we have to enable "prometheus-exporter" module to enable monitoring.

### spec.monitor

Solr managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. 


### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for Solr. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any Kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc.

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the Petset created for Solr server.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (petset's annotation)
- spec:
  - resources
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

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/39bf8b2/api/v2/types.go#L44-L279).
Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.imagePullSecret

`KubeDB` provides the flexibility of deploying Solr server from a private Docker registry.
#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

  `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

  If this field is left empty, the KubeDB operator will create a service account name matching Solr crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

  If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

  If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. 

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide a template for the services created by KubeDB operator for Solr server through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
  - `primary` is used for the primary service identification.
  - `standby` is used for the secondary service identification.
  - `stats` is used for the exporter service identification.

- metadata:
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

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163) to understand these fields in detail.

### spec.tls

> The ReconfigureTLS only works with the [Cert-Manager](https://cert-manager.io/docs/concepts/) managed certificates. [Installation guide](https://cert-manager.io/docs/installation/).

`spec.tls` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `ReconfigureTLS`. It specifies the necessary information required to add or remove or update the TLS configuration of the Solr cluster. It consists of the following sub-fields:

- `tls.remove` ( `bool` | `false` ) - tells the operator to remove the TLS configuration for the HTTP layer. The transport layer is always secured with certificates, so the removal process does not affect the transport layer.
- `tls.rotateCertificates` ( `bool` | `false`) - tells the operator to renew all the certificates.
- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Solr. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
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

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Solr` crd or which resources KubeDB should keep or delete when you delete `Solr` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Solr crd for different deletion policies,

| Behavior                            | DoNotTerminate |   Halt   |  Delete  | WipeOut  |
|-------------------------------------|:--------------:|:--------:|:--------:|:--------:|
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete Petset                    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshot data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |
If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` deletion policy by default.

### spec.halted
Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run a Solr server [here](/docs/guides/solr/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
