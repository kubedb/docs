---
title: Postgres CRD
menu:
  docs_{{ .version }}:
    identifier: pg-postgres-concepts
    name: Postgres
    parent: pg-concepts-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Postgres

## What is Postgres

`Postgres` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [PostgreSQL](https://www.postgresql.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a Postgres object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Postgres Spec

As with all other Kubernetes objects, a Postgres needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

Below is an example Postgres object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: p1
  namespace: demo
spec:
  version: "10.2-v5"
  replicas: 2
  standbyMode: Hot
  streamingMode: asynchronous
  leaderElection:
    leaseDurationSeconds: 15
    renewDeadlineSeconds: 10
    retryPeriodSeconds: 2
  archiver:
    storage:
      storageSecretName: s3-secret
      s3:
        bucket: kubedb
  authSecret:
    name: p1-auth
  storageType: "Durable"
  storage:
    storageClassName: standard
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: pg-init-script
  backupSchedule:
    cronExpression: "@every 2m"
    storageSecretName: gcs-secret
    gcs:
      bucket: kubedb-qa
      prefix: demo
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          app: kubedb
        interval: 10s
  configSecret:
    name: pg-custom-config
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
    controller:
      annotations:
        passMe: ToStatefulSet
    spec:
      serviceAccountName: my-custom-sa
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
      - name: myregistrykey
      env:
      - name: POSTGRES_DB
        value: pgdb
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
        limits:
          memory: "128Mi"
          cpu: "500m"
  serviceTemplate:
    metadata:
      annotations:
        passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  5432
        targetPort: http
  replicaServiceTemplate:
    annotations:
      passMe: ToReplicaService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  5432
        targetPort: http
  terminationPolicy: "Halt"
```

### spec.version

`spec.version` is a required field that specifies the name of the [PostgresVersion](/docs/guides/postgres/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `PostgresVersion` resources,

```bash
$ kubectl get pgversion
NAME       VERSION   DB_IMAGE                   DEPRECATED   AGE
10.2       10.2      kubedb/postgres:10.2       true         44m
10.2-v1    10.2      kubedb/postgres:10.2-v2    true         44m
10.2-v2    10.2      kubedb/postgres:10.2-v3                 44m
10.2-v3    10.2      kubedb/postgres:10.2-v4                 44m
10.2-v4    10.2      kubedb/postgres:10.2-v5                 44m
10.2-v5    10.2      kubedb/postgres:10.2-v6                 44m
10.6       10.6      kubedb/postgres:10.6                    44m
10.6-v1    10.6      kubedb/postgres:10.6-v1                 44m
10.6-v2    10.6      kubedb/postgres:10.6-v2                 44m
10.6-v3    10.6      kubedb/postgres:10.6-v3                 44m
11.1       11.1      kubedb/postgres:11.1                    44m
11.1-v1    11.1      kubedb/postgres:11.1-v1                 44m
11.1-v2    11.1      kubedb/postgres:11.1-v2                 44m
11.1-v3    11.1      kubedb/postgres:11.1-v3                 44m
11.2       11.2      kubedb/postgres:11.2                    44m
11.2-v1    11.2      kubedb/postgres:11.2-v1                 44m
9.6        9.6       kubedb/postgres:9.6        true         44m
9.6-v1     9.6       kubedb/postgres:9.6-v2     true         44m
9.6-v2     9.6       kubedb/postgres:9.6-v3                  44m
9.6-v3     9.6       kubedb/postgres:9.6-v4                  44m
9.6-v4     9.6       kubedb/postgres:9.6-v5                  44m
9.6-v5     9.6       kubedb/postgres:9.6-v6                  44m
9.6.7      9.6.7     kubedb/postgres:9.6.7      true         44m
9.6.7-v1   9.6.7     kubedb/postgres:9.6.7-v2   true         44m
9.6.7-v2   9.6.7     kubedb/postgres:9.6.7-v3                44m
9.6.7-v3   9.6.7     kubedb/postgres:9.6.7-v4                44m
9.6.7-v4   9.6.7     kubedb/postgres:9.6.7-v5                44m
9.6.7-v5   9.6.7     kubedb/postgres:9.6.7-v6                44m
```
### spec.replicas

`spec.replicas` specifies the total number of primary and standby nodes in Postgres database cluster configuration. One pod is selected as Primary and others act as standby replicas. KubeDB uses `PodDisruptionBudget` to ensure that majority of the replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).

To learn more about how to setup a HA PostgreSQL cluster in KubeDB, please visit [here](/docs/guides/postgres/clustering/ha_cluster.md).

### spec.standbyMode

`spec.standby` is an optional field that specifies the standby mode (_Warm / Hot_) to use for standby replicas. In **hot standby** mode, standby replicas can accept connection and run read-only queries. In **warm standby** mode, standby replicas can't accept connection and only used for replication purpose.

### spec.streamingMode

`spec.streamingMode` is an optional field that specifies the streaming mode (_synchronous / asynchronous_) of the standby replicas. KubeDB currently supports only **asynchronous** streaming mode.

### spec.leaderElection

There are three fields under Postgres CRD's `spec.leaderElection`. These values defines how fast the leader election can happen.

- `leaseDurationSeconds`: This is the duration in seconds that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack. Default 15 sec.
- `renewDeadlineSeconds`: This is the duration in seconds that the acting master will retry refreshing leadership before giving up. Normally, LeaseDuration \* 2 / 3. Default 10 sec.
- `retryPeriodSeconds`: This is the duration in seconds the LeaderElector clients should wait between tries of actions. Normally, LeaseDuration / 3. Default 2 sec.

If the Cluster machine is powerful, user can reduce the times. But, Do not make it so little, in that case Postgres will restarts very often.

### spec.archiver

`spec.archiver` is an optional field which specifies storage information that will be used by `wal-g`. User can use either s3 or gcs.

- `storage.storageSecretName` points to the Secret containing the credentials for cloud storage destination.
- `storage.s3` points to s3 storage configuration.
- `storage.s3.bucket` points to the bucket name used to store continuous archiving data.
- `storage.gcs` points to GCS storage configuration.
- `storage.gcs.bucket` points to the bucket name used to store continuous archiving data.

Continuous archiving data will be stored in a folder called `{bucket}/{prefix}/kubedb/{namespace}/{postgres-name}/archive/`.

To learn more about how to configure Postgres to archive WAL data continuously in AWS S3 bucket, please visit [here](/docs/guides/postgres/backup/wal/continuous_archiving.md).

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `postgres` database. If not set, KubeDB operator creates a new Secret with name `{postgres-name}-auth` that hold _username_ and _password_ for `postgres` database.

If you want to use an existing or custom secret, please specify that when creating the Postgres object using `spec.authSecret.name`. This Secret should contain superuser _username_ as `POSTGRES_USER` key and superuser _password_ as `POSTGRES_PASSWORD` key. Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version >= 0.13.0).

Example:

```bash
$ kubectl create secret generic p1-auth -n demo \
--from-literal=POSTGRES_USER=not@user \
--from-literal=POSTGRES_PASSWORD=not@secret
secret "p1-auth" created
```

```bash
$ kubectl get secret -n demo p1-auth -o yaml
apiVersion: v1
data:
  POSTGRES_PASSWORD: bm90QHNlY3JldA==
  POSTGRES_USER: bm90QHVzZXI=
kind: Secret
metadata:
  creationTimestamp: 2018-09-03T11:25:39Z
  name: p1-auth
  namespace: demo
  resourceVersion: "1677"
  selfLink: /api/v1/namespaces/demo/secrets/p1-auth
  uid: 15b3e8a1-af6c-11e8-996d-0800270d7bae
type: Opaque
```

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Postgres database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. In this case, you don't have to specify `spec.storage` field.

### spec.storage

If you don't set `spec.storageType:` to `Ephemeral` then `spec.storage` field is required. This field specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.init

`spec.init` is an optional section that can be used to initialize a newly created Postgres database. PostgreSQL databases can be initialized from these three ways:

1. Initialize from Script
2. Initialize from Snapshot
3. Initialize from WAL archive

#### Initialize via Script

To initialize a PostgreSQL database using a script (shell script, db migrator, etc.), set the `spec.init.script` section when creating a Postgres object. `script` must have the following information:

- [VolumeSource](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes): Where your script is loaded from.

Below is an example showing how a script from a configMap can be used to initialize a PostgreSQL database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: postgres-db
spec:
  version: "10.2-v5"
  init:
    script:
      configMap:
        name: pg-init-script
```

In the above example, Postgres will execute provided script once the database is running. For more details tutorial on how to initialize from script, please visit [here](/docs/guides/postgres/initialization/script_source.md).

#### Initialize from WAL archive

To initialize from WAL archive, set the `spec.init.postgresWAL` section when creating a Postgres object.

Below is an example showing how to initialize a PostgreSQL database from WAL archive.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: postgres-db
spec:
  version: "10.2-v5"
  authSecret:
    name: postgres-old
  init:
    postgresWAL:
      storageSecretName: s3-secret
      s3:
        endpoint: "s3.amazonaws.com"
        bucket: kubedb
        prefix: "kubedb/demo/old-pg/archive"
```

In the above example, PostgreSQL database will be initialized from WAL archive.

When initializing from WAL archive, superuser credentials must have to match with the previous one. For example, let's say, we want to initialize this database from `postgres-old` WAL archive. In this case, superuser credentials of new Postgres should be the same as `postgres-old`. Otherwise, the restoration process will be failed.

For more details tutorial on how to initialize from wal archive, please visit [here](/docs/guides/postgres/initialization/wal/wal_source.md).

### spec.monitor

PostgreSQL managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor PostgreSQL with builtin Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md)
- [Monitor PostgreSQL with Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for PostgreSQL. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). You can use any kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/postgres/configuration/using-config-file.md).

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the StatefulSet created for Postgres database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
  - annotations (pod's annotation)
- controller
  - annotations (statefulset's annotation)
- spec:
  - serviceAccountName
  - env
  - resources
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - livenessProbe
  - readinessProbe
  - lifecycle

Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account name matching Postgres crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/postgres/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### spec.podTemplate.spec.env

`spec.podTemplate.spec.env` is an optional field that specifies the environment variables to pass to the Postgres docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/_/postgres/).

Note that, the KubeDB operator does not allow `POSTGRES_USER` and `POSTGRES_PASSWORD` environment variable to set in `spec.podTemplate.spec.env`. If you want to set the superuser _username_ and _password_, please use `spec.authSecret` instead described earlier.

If you try to set `POSTGRES_USER` or `POSTGRES_PASSWORD` environment variable in Postgres crd, KubeDB operator will reject the request with following error,

```ini
Error from server (Forbidden): error when creating "./postgres.yaml": admission webhook "postgres.validators.kubedb.com" denied the request: environment variable POSTGRES_PASSWORD is forbidden to use in Postgres spec
```

Also, note that KubeDB does not allow to update the environment variables as updating them does not have any effect once the database is created. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./postgres.yaml": admission webhook "postgres.validators.kubedb.com" denied the request: precondition failed for:
...
At least one of the following was changed:
    apiVersion
    kind
    name
    namespace
    spec.standby
    spec.streaming
    spec.archiver
    spec.authSecret
    spec.storageType
    spec.storage
    spec.podTemplate.spec.nodeSelector
    spec.init
```

#### spec.podTemplate.spec.imagePullSecrets

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry. For more details on how to use private docker registry, please visit [here](/docs/guides/postgres/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplate

KubeDB creates two different services for each Postgres instance. One of them is a master service named `<postgres-name>` and points to the Postgres `Primary` pod/node. Another one is a replica service named `<postgres-name>-replicas` and points to Postgres `replica` pods/nodes.

These `master` and `replica` services can be customized using [spec.serviceTemplate](#spec.servicetemplate) and [spec.replicaServiceTemplate](#specreplicaservicetemplate) respectively.

You can provide template for the `master` service using `spec.serviceTemplate`. This will allow you to set the type and other properties of the service. If `spec.serviceTemplate` is not provided, KubeDB will create a `master` service of type `ClusterIP` with minimal settings.

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

### spec.replicaServiceTemplate

You can provide template for the `replica` service using `spec.replicaServiceTemplate`. If `spec.replicaServiceTemplate` is not provided, KubeDB will create a `replica` service of type `ClusterIP` with minimal settings.

The fileds of `spec.replicaServiceTemplate` is similar to `spec.serviceTemplate`, that is:

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

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163) to understand these fields in detail.

### spec.terminationPolicy

`terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Postgres` crd or which resources KubeDB should keep or delete when you delete `Postgres` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `terminationPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to provide safety from accidental deletion of database. If admission webhook is enabled, KubeDB prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Postgres crd for different termination policies,

| Behavior                                 | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ---------------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation                |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Create Dormant Database               |    &#10007;    | &#10003; | &#10007; | &#10007; |
| 3. Delete StatefulSet                    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete Services                       |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 5. Delete PVCs                           |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 6. Delete Secrets                        |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshots                      |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 8. Delete Snapshot data from bucket      |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 9. Delete archieved WAL data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.terminationPolicy` KubeDB uses `Halt` termination policy by default.

## Next Steps

- Learn how to use KubeDB to run a PostgreSQL database [here](/docs/guides/postgres/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
