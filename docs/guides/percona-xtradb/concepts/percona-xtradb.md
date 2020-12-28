---
title: PerconaXtraDB CRD
menu:
  docs_{{ .version }}:
    identifier: px-percona-xtradb-concepts
    name: PerconaXtraDB
    parent: px-concepts-percona-xtradb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB

## What is PerconaXtraDB

`PerconaXtraDB` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for standalone PerconaXtraDB ([Percona Server](https://www.percona.com/software/mysql-database/percona-server)) and [Percona XtraDB Cluster](https://www.percona.com/software/mysql-database/percona-xtradb-cluster) in a Kubernetes native way. You only need to describe the desired configuration in a `PerconaXtraDB` object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## PerconaXtraDB Spec

As with all other Kubernetes objects, a PerconaXtraDB needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `spec` section. Below is an example PerconaXtraDB object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: demo-px
  namespace: demo
spec:
  version: "5.7"
  replicas: 3
  authSecret:
    name: demo-px-auth
  storageType: "Durable"
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
  podTemplate:
    annotations:
      passMe: ToDatabasePod
    controller:
      annotations:
        passMe: ToStatefulSet
    spec:
      serviceAccountName: px-service-account
      schedulerName: px-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
      - name: myregistrykey
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
  serviceTemplate:
    annotations:
      passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  9200
        targetPort: http
  terminationPolicy: Halt
```

### .spec.version

`.spec.version` is a required field specifying the name of the [PerconaXtraDBVersion](/docs/guides/percona-xtradb/concepts/catalog.md) object where the docker images are specified. Currently, when you install KubeDB, it creates the following `PerconaXtraDBVersion` resources,

- `5.7`
- `5.7-cluster`

### .spec.replicas

`.spec.replicas` specifies the number of instances to deploy for PerconaXtraDB. If set to 1, KubeDB will deploy a standalone Percona server. If set to value larger than 1, deploy PerconaXtraDB cluster with specified number of masters.

To learn more about how to setup a Percona XtraDB cluster using KubeDB, please visit [here](/docs/guides/percona-xtradb/clustering/percona-xtradb-cluster.md).

### .spec.authSecret

`.spec.authSecret` is an optional field that points to a Secret used to hold credentials for `mysql` root user. If not set, KubeDB operator creates a new Secret `{percona-xtradb-object-name}-auth` for storing the password for `mysql` root user for each PerconaXtraDB object. If you want to use an existing Secret please specify that when creating the PerconaXtraDB object using `.spec.authSecret.name`.

This secret contains a `username` key and a `password` key which contains the username and password respectively for `mysql` root user. Here, the value of `username` key is fixed to be `root`.

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

Example:

```bash
$ kubectl create secret generic demo-px-auth -n demo \
    --from-literal=username=root \
    --from-literal=password=6q8u_2jMOW-OOZXk
secret/demo-px-auth created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: cm9vdA==
kind: Secret
metadata:
  ...
  name: demo-px-auth
  namespace: demo
  ...
type: Opaque
```

### .spec.storageType

`.spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create PerconaXtraDB using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. In this case, you don't have to specify `.spec.storage` field.

### .spec.storage

Since 0.9.0-rc.0, If you set `.spec.storageType:` to `Durable`, then  `.spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `.spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `.spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `.spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### .spec.init

`.spec.init` is an optional section that can be used to initialize a newly created PerconaXtraDB. PerconaXtraDB can be initialized in one of two ways:

 1. Initialize from Script
 2. Initialize from Snapshot

#### Initialize via Script

To initialize a PerconaXtraDB database (with replica 1) using a script (shell script, sql script etc.), set the `.spec.init.script` section when creating a PerconaXtraDB object. It will execute files alphabetically with extensions `.sh` , `.sql`  and `.sql.gz` that are found in the repository. The scripts inside child folders will be skipped. script must have following information:

- [VolumeSource](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes): Where your script is loaded from.

Below is an example showing how a script from a configMap can be used to initialize a PerconaXtraDB database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: init-demo-px
spec:
  version: 5.7
  init:
    script:
      configMap:
        name: my-init-script
```

In the above example, KubeDB operator will launch a Job to execute all js script of `my-init-script` in alphabetical order once StatefulSet Pods are running. For more details tutorial on how to initialize from script, please visit [here](/docs/guides/percona-xtradb/initialization/using-script.md).

#### Initialize from Stash Backup

If you have a backup of PerconaXtraDB cluster using [Stash](https://stash.run), you can initialize from it. To initialize, set the `.spec.init.stashRestoreSession` section when creating a PerconaXtraDB object. In this case, StashRestoreSession must have following information:

- `name:` Name of the RestoreSession object

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: px-new
spec:
  version: 5.7
  ...
  init:
    stashRestoreSession:
      name: "snapshot-xyz"
```

In the above example, PerconaXtraDB cluster will be initialized from Snapshot `snapshot-xyz`. Here, Stash operator will launch necessary Jobs to initialize once StatefulSet Pods are running.

When initializing from Snapshot, root user credentials must have to match with the previous one. For example, let's say, Snapshot `snapshot-xyz` is for PerconaXtraDB `px-old`. In this case, new PerconaXtraDB `px-new` should use the same credentials for root user of `px-old`. Otherwise, the restoration process will fail.

For more details tutorial on how to initialize from snapshot, please visit [here](/docs/guides/percona-xtradb/backup/stash.md).

### .spec.monitor

PerconaXtraDB managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor PerconaXtraDB with builtin Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md)
- [Monitor PerconaXtraDB with Prometheus operator](/docs/guides/percona-xtradb/monitoring/using-prometheus-operator.md)

### .spec.configSecret

`.spec.configSecret` is an optional field that allows users to provide custom configuration for PerconaXtraDB. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/percona-xtradb/configuration/using-config-file.md).

### .spec.podTemplate

KubeDB allows providing a template for database Pod through `.spec.podTemplate`. KubeDB operator will pass the information provided in `.spec.podTemplate` to the StatefulSet created for PerconaXtraDB.

KubeDB accept following fields to set in `.spec.podTemplate:`

- metadata:
  - annotations (Pod's annotation)
- controller:
  - annotations (StatefulSet's annotation)
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

Usage of some field of `.spec.podTemplate` is described below,

#### .spec.podTemplate.spec.args

`.spec.podTemplate.spec.args` is an optional field. This can be used to provide additional arguments to database installation. To learn about available args of `mysqld`, visit [here](https://dev.mysql.com/doc/refman/5.7/en/server-options.html).

#### .spec.podTemplate.spec.env

`.spec.podTemplate.spec.env` is an optional field that specifies the environment variables to pass to the PerconaXtraDB Docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/_/mysql/).

Note that, KubeDB does not allow `MYSQL_ROOT_PASSWORD`, `MYSQL_ALLOW_EMPTY_PASSWORD`, `MYSQL_RANDOM_ROOT_PASSWORD`, and `MYSQL_ONETIME_PASSWORD` environment variables to set in `.spec.env`. If you want to set the root password, please use `.spec.authSecret` instead described earlier.

If you try to set any of the forbidden environment variables i.e. `MYSQL_ROOT_PASSWORD` in PerconaXtraDB object, KubeDB operator will reject the request with following error,

```ini
Error from server (Forbidden): error when creating "./percona-xtradb.yaml": admission webhook "perconaxtradb.validators.kubedb.com" denied the request: environment variable MYSQL_ROOT_PASSWORD is forbidden to use in PerconaXtraDB spec
```

Also note that KubeDB does not allow to update the environment variables as updating them does not have any effect once the database is created. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./percona-xtradb.yaml": admission webhook "perconaxtradb.validators.kubedb.com" denied the request: precondition failed for:
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
```

#### .spec.podTemplate.spec.imagePullSecrets

`KubeDB` provides the flexibility of deploying PerconaXtraDB from a private Docker registry. `.spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling Docker image if you are using a private docker registry. To learn how to deploy PerconaXtraDB from a private registry, please visit [here](/docs/guides/percona-xtradb/private-registry/using-private-registry.md).

#### .spec.podTemplate.spec.nodeSelector

`.spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the Pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### .spec.podTemplate.spec.serviceAccountName

 `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

 If this field is left empty, the KubeDB operator will create a service account name matching PerconaXtraDB object name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

 If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

 If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/percona-xtradb/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### .spec.podTemplate.spec.resources

`.spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database Pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### .spec.serviceTemplate

You can also provide a template for the Services created by KubeDB operator for PerconaXtraDB through `.spec.serviceTemplate`. This will allow you to set the type and other properties of the Services.

KubeDB allows following fields to set in `.spec.serviceTemplate`:

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

### .spec.terminationPolicy

`terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `PerconaXtraDB` object or which resources KubeDB should keep or delete when you delete `PerconaXtraDB` object. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `terminationPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `.spec.terminationPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete PerconaXtraDB object for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Create Dormant Database          |    &#10007;    | &#10003; | &#10007; | &#10007; |
| 3. Delete StatefulSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 5. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 6. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `.spec.terminationPolicy` KubeDB uses `Halt` termination policy by default.

## Next Steps

- Learn how to use KubeDB to run a PerconaXtraDB [here](/docs/guides/percona-xtradb/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
