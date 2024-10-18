---
title: ZooKeeper CRD
menu:
  docs_{{ .version }}:
    identifier: zk-zookeeper-concepts
    name: ZooKeeper
    parent: zk-concepts-zookeeper
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ZooKeeper

## What is ZooKeeper

`ZooKeeper` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [ZooKeeper](https://zookeeper.apache.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a ZooKeeper object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## ZooKeeper Spec

As with all other Kubernetes objects, a ZooKeeper needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example ZooKeeper object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-ensemble
  namespace: demo
spec:
  version: 3.9.1
  replicas: 3
  disableAuth: false
  adminServerPort: 8080
  clientSecurePort: 2182
  authSecret:
    name: zk-auth
    externallyManaged: false
  keystoreCredSecret:
    name: zk-quickstart-keystore-cred
  enableSSL: true
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          app: kubedb
        interval: 10s
  configSecret:
    name: zk-custom-config
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
    controller:
      annotations:
        passMe: ToPetSet
    spec:
      serviceAccountName: my-service-account
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
        - name: myregistrykey
  serviceTemplates:
    - alias: primary
      metadata:
        annotations:
          passMe: ToService
      spec:
        type: NodePort
        ports:
          - name:  http
            port:  9200
  deletionPolicy: Halt
  halted: false
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 2
    disableWriteCheck: false
  tls:
    certificates:
      - alias: server
        secretName: zk-quickstart-server-cert
      - alias: client
        secretName: zk-quickstart-client-cert
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: zookeeper-ca-issuer
```


### spec.version

`spec.version` is a required field specifying the name of the [ZooKeeperVersion](/docs/guides/zookeeper/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `ZooKeeperVersion` crds,

-  `3.7.2`
-  `3.8.3`
-  `3.9.1`

### spec.replicas

`spec.replicas` the number of nodes in ZooKeeper cluster.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.disableAuth

`spec.disableAuth` is an optional field that decides whether ZooKeeper instance will be secured by auth or no.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `zookeeper` superuser. If not set, KubeDB operator creates a new Secret `{zookeeper-object-name}-auth` for storing the password for `zookeeper` superuser.

We can use this field in 3 mode.

1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the ZooKeeper object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```
2. Specifying the secret name only. In this case, You need to specify the secret name when creating the ZooKeeper object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `username` key and a `password` key which contains the `username` and `password` respectively for `zookeeper` superuser.

Example:

```bash
$ kubectl create secret generic zk-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "zk-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: zk-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).


### spec.storage

If you set `spec.storageType:` to `Durable`, then  `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.monitor

ZooKeeper managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. 


### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for ZooKeeper. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any Kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc.

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
      name: zk-issuer
    certificates:
    - alias: server
      privateKey:
        encoding: PKCS8
      secretName: zk-client-cert
      subject:
        organizations:
        - kubedb
    - alias: http
      privateKey:
        encoding: PKCS8
      secretName: zk-server-cert
      subject:
        organizations:
        - kubedb
```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for ZooKeeper. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
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


### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for ZooKeeper server.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (petset's annotation)
- spec:
  - args
  - env
  - resources
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - livenessProbe
  - readinessProbe
  - lifecycle

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/ea366935d5bad69d7643906c7556923271592513/api/v1/types.go#L42-L259).
Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.args
 `spec.podTemplate.spec.args` is an optional field. This can be used to provide additional arguments to database installation.

### spec.podTemplate.spec.env

`spec.podTemplate.spec.env` is an optional field that specifies the environment variables to pass to the ZooKeeper docker image.


#### spec.podTemplate.spec.imagePullSecret

`KubeDB` provides the flexibility of deploying ZooKeeper server from a private Docker registry.
#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

  `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

  If this field is left empty, the KubeDB operator will create a service account name matching ZooKeeper crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

  If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

  If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. 

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide a template for the services created by KubeDB operator for ZooKeeper server through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

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

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `ZooKeeper` crd or which resources KubeDB should keep or delete when you delete `ZooKeeper` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete ZooKeeper crd for different deletion policies,

| Behavior                            | DoNotTerminate |   Halt   |  Delete  | WipeOut  |
|-------------------------------------|:--------------:|:--------:|:--------:|:--------:|
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
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
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run a ZooKeeper server [here](/docs/guides/zookeeper/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
