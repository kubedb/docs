---
title: MySQL CRD
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-concepts-database
    name: MySQL
    parent: guides-mysql-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL

## What is MySQL

`MySQL` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [MySQL](https://www.mysql.com/) in a Kubernetes native way. You only need to describe the desired database configuration in a MySQL object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## MySQL Spec

As with all other Kubernetes objects, a MySQL needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example MySQL object.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: m1
  namespace: demo
spec:
  version: "9.1.0"
  topology:
    mode: GroupReplication
  authSecret:
    name: m1-auth
  storageType: "Durable"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: my-init-script
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          app: kubedb
        interval: 10s
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  configSecret:
    name: my-custom-config
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
      containers:
      - name: mysql
        args:
        - --character-set-server=utf8mb4
        env:
        - name: MYSQL_DATABASE
          value: myDB
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
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
```

### spec.version

`spec.version` is a required field specifying the name of the [MySQLVersion](/docs/guides/mysql/concepts/catalog/index.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `MySQLVersion` resources,

- `9.0.1`, `9.1.0`
- `8.0.35`, `8.0.36`, `8.1.0`, `8.2.0`, `8.4.2`, `8.4.3`
- `8.0.31-innodb`
- `5.7.44`, `5.7.42-debian`

### Handling `mbind: Operation not permitted`
On certain platforms (e.g., when using specific security profiles), for some versions of `mysql`, you may see log messages like:   
`mbind: Operation not permitted`

This indicates that the `mysql` container needs the `SYS_NICE` kernel capability to perform CPU‐affinity or real-time scheduling operations. You can grant this capability by extending your Pod spec as follows:
```yaml
spec:
  podTemplate:
    spec:
      containers:
      - name: mysql
        securityContext:
          capabilities:
            add: ["SYS_NICE"]
            drop: ["ALL"]
```
This ensures that only the SYS_NICE capability is added—while all others are dropped—keeping your container’s security posture minimal.

### spec.topology

`spec.topology` is an optional field that provides a way to configure HA, fault-tolerant MySQL cluster. This field enables you to specify the clustering mode. Currently, we support only MySQL Group Replication. KubeDB uses `PodDisruptionBudget` to ensure that majority of the group replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained and no data loss has occurred.

You can specify the following fields in `spec.topology` field,

- `mode` specifies the clustering mode for MySQL. For now, the supported value is `"GroupReplication"` for MySQL Group Replication. This field is required if you want to deploy MySQL cluster.

- `group` is an optional field to configure a group replication. It contains the following fields:
  - `name` is an optional field to specify the name for the group. It must be a version 4 UUID if specified.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `mysql` root user. If not set, the KubeDB operator creates a new Secret `{mysql-object-name}-auth` for storing the password for `mysql` root user for each MySQL object. If you want to use an existing secret please specify that when creating the MySQL object using `spec.authSecret.name`.

This secret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `mysql` root user. Here, the value of `user` key is fixed to be `root`.

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

Example:

```bash
$ kubectl create secret generic m1-auth -n demo \
--from-literal=user=root \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "m1-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  user: cm9vdA==
kind: Secret
metadata:
  name: m1-auth
  namespace: demo
type: Opaque
```

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for the database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MySQL database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. In this case, you don't have to specify `spec.storage` field.

### spec.storage

Since 0.9.0-rc.0, If you set `spec.storageType:` to `Durable`, then  `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.init

`spec.init` is an optional section that can be used to initialize a newly created MySQL database. MySQL databases can be initialized in one of two ways:

- Initialize from Script

#### Initialize via Script

To initialize a MySQL database using a script (shell script, sql script, etc.), set the `spec.init.script` section when creating a MySQL object. It will execute files alphabetically with extensions `.sh` , `.sql`  and `.sql.gz` that is found in the repository. The scripts inside child folders will be skipped. script must have the following information:

- [VolumeSource](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes): Where your script is loaded from.

Below is an example showing how a script from a configMap can be used to initialize a MySQL database.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: m1
  namespace: demo
spec:
  version: 9.1.0
  init:
    script:
      configMap:
        name: mysql-init-script
```

In the above example, KubeDB operator will launch a Job to execute all js script of `mysql-init-script` in alphabetical order once PetSet pods are running. For more details tutorial on how to initialize from script, please visit [here](/docs/guides/mysql/initialization/index.md).

### spec.monitor

MySQL managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor MySQL with builtin Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md)
- [Monitor MySQL with Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md)

### spec.requireSSL

`spec.requireSSL` specifies whether the client connections require SSL. If `spec.requireSSL` is `true` then the server permits only TCP/IP connections that use SSL, or connections that use a socket file (on Unix) or shared memory (on Windows). The server rejects any non-secure connection attempt. For more details, please visit [here](https://dev.mysql.com/doc/refman/5.7/en/using-encrypted-connections.html)

### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the MySQL.

The following fields are configurable in the `spec.tls` section:

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.

  - `apiGroup` is the group name of the resource being referenced. The value for `Issuer` or   `ClusterIssuer` is "cert-manager.io"   (cert-manager v0.12.0 and later).
  - `kind` is the type of resource being referenced. KubeDB supports both `Issuer`   and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.

- `certificates` (optional) are a list of certificates used to configure the server and/or client certificate. It has the following fields:
  
  - `alias` represents the identifier of the certificate. It has the following possible value:
    - `server` is used for server certificate identification.
    - `client` is used for client certificate identification.
    - `metrics-exporter` is used for metrics exporter certificate identification.
  - `secretName` (optional) specifies the k8s secret name that holds the certificates.
    >This field is optional. If the user does not specify this field, the default secret name will be created in the following format: `<database-name>-<cert-alias>-cert`.
  - `subject` (optional) specifies an `X.509` distinguished name. It has the following possible field,
    - `organizations` (optional) are the list of different organization names to be used on the Certificate.
    - `organizationalUnits` (optional) are the list of different organization unit name to be used on the Certificate.
    - `countries` (optional) are the list of country names to be used on the Certificate.
    - `localities` (optional) are the list of locality names to be used on the Certificate.
    - `provinces` (optional) are the list of province names to be used on the Certificate.
    - `streetAddresses` (optional) are the list of a street address to be used on the Certificate.
    - `postalCodes` (optional) are the list of postal code to be used on the Certificate.
    - `serialNumber` (optional) is a serial number to be used on the Certificate.
    You can found more details from [Here](https://golang.org/pkg/crypto/x509/pkix/#Name)  

  - `duration` (optional) is the period during which the certificate is valid.
  - `renewBefore` (optional) is a specifiable time before expiration duration.
  - `dnsNames` (optional) is a list of subject alt names to be used in the Certificate.
  - `ipAddresses` (optional) is a list of IP addresses to be used in the Certificate.
  - `uriSANs` (optional) is a list of URI Subject Alternative Names to be set in the Certificate.
  - `emailSANs` (optional) is a list of email Subject Alternative Names to be set in the Certificate.

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for MySQL. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any Kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/mysql/configuration/config-file/index.md).

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for the MySQL database.

KubeDB accepts the following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (petset's annotation)
- spec:
  - containers
  - volumes
  - podPlacementPolicy
  - initContainers
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



#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the MySQL docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/_/mysql/).

Note that, KubeDB does not allow `MYSQL_ROOT_PASSWORD`, `MYSQL_ALLOW_EMPTY_PASSWORD`, `MYSQL_RANDOM_ROOT_PASSWORD`, and `MYSQL_ONETIME_PASSWORD` environment variables to set in `spec.env`. If you want to set the root password, please use `spec.authSecret` instead described earlier.

If you try to set any of the forbidden environment variables i.e. `MYSQL_ROOT_PASSWORD` in MySQL crd, Kubed operator will reject the request with the following error,

```ini
Error from server (Forbidden): error when creating "./mysql.yaml": admission webhook "mysql.validators.kubedb.com" denied the request: environment variable MYSQL_ROOT_PASSWORD is forbidden to use in MySQL spec
```

Also, note that KubeDB does not allow to update the environment variables as updating them does not have any effect once the database is created.  If you try to update environment variables, KubeDB operator will reject the request with the following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./mysql.yaml": admission webhook "mysql.validators.kubedb.com" denied the request: precondition failed for:
...At least one of the following was changed:
    apiVersion
    kind
    name
    namespace
    spec.authSecret
    spec.init
    spec.storageType
    spec.storage
    spec.podTemplate.spec.nodeSelector
    spec.podTemplate.spec.env
```

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

#### spec.podTemplate.spec.imagePullSecrets

`KubeDB` provides the flexibility of deploying MySQL database from a private Docker registry. `spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry. To learn how to deploy MySQL from a private registry, please visit [here](/docs/guides/mysql/private-registry/index.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

 `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine-tune role-based access control.

 If this field is left empty, the KubeDB operator will create a service account name matching MySQL crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

 If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

 If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/mysql/custom-rbac/index.md) to grant necessary permissions in this scenario.

### spec.serviceTemplate

You can also provide a template for the services created by KubeDB operator for MySQL database through `spec.serviceTemplate`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplate`:

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

### spec.halted

`spec.halted` is an optional field. This field will be used to halt the kubeDB operator. When you set `spec.halted` to `true`, the KubeDB operator doesn't perform any operation on `MySQL` object.

Suppose you want to delete the `MySQL` resources(`PetSet`, `Service` etc.) except `MySQL` object, `PVCs` and `Secret` then you need to set `spec.halted` to `true`. If you set `spec.halted` to `true` then the `deletionPolicy` in `MySQL` object will be set `Halt` by-default.  

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MySQL` crd or which resources KubeDB should keep or delete when you delete `MySQL` crd. KubeDB provides the following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete MySQL crd for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.

> For more details you can visit [here](https://appscode.com/blog/post/deletion-policy/)

## Next Steps

- Learn how to use KubeDB to run a MySQL database [here](/docs/guides/mysql/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
