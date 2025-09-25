---
title: Postgres CRD
menu:
  docs_{{ .version }}:
    identifier: pg-postgres-gitops-concepts
    name: Postgres(GitOps)
    parent: pg-concepts-postgres
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Postgres(`gitops.kubedb.com/v1alpha1`)

## What is `gitops.kubedb.com` Postgres

`Postgres` is a Kubernetes `Custom Resource Definitions` (CRD) which is under the `gitops.kubedb.com/v1alpha1` API group. It provides the same specification as the standard KubeDB [`Postgres`](/docs/guides/postgres/concepts/postgres.md) CRD, but is designed to work with GitOps workflows. This allows you to manage your PostgreSQL databases using GitOps principles, where the desired state of your database is stored in a Git repository.
 You only need to describe the desired database configuration in a `gitops` Postgres object through GitOps tools like ArgoCD or FluxCD. The GitOps tool will monitor the Git repository for changes, and when a change is detected, it will apply the changes to your Kubernetes cluster and create necessary [OpsRequest](/docs/guides/postgres/concepts/opsrequest.md). This means you don't have to manually create or update the database configuration in your cluster; instead, you just update the Git repository with the desired state of your database configuration
, and the KubeDB gitops operator will create/update standard Kubernetes database [objects](/docs/guides/postgres/concepts/postgres.md) in the desired state for you.

## Postgres Spec

As with all other Kubernetes objects, a gitops `Postgres` needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.
This object reflects the same specification as the standard KubeDB [`Postgres`](/docs/guides/postgres/concepts/postgres.md) CRD. The `Postgres` CRD is used to create and manage PostgreSQL databases in Kubernetes.
Below is an example Postgres object.

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: p1
  namespace: demo
spec:
  version: "13.13"
  replicas: 2
  standbyMode: Hot
  streamingMode: Asynchronous
  leaderElection:
    leaseDurationSeconds: 15
    renewDeadlineSeconds: 10
    retryPeriodSeconds: 2
  authSecret:
    kind: Secret
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
        passMe: ToPetSet
    spec:
      serviceAccountName: my-custom-sa
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
      - name: myregistrykey
      containers:
      - name: postgres
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
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  5432
  - alias: standby
    metadata:
      annotations:
        passMe: ToReplicaService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  5432
  deletionPolicy: "Halt"
```

### spec.version

`spec.version` is a required field that specifies the name of the [PostgresVersion](/docs/guides/postgres/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `PostgresVersion` resources,

```bash
~ $ kubectl get pgversion
NAME                      VERSION   DISTRIBUTION   DB_IMAGE                                                                DEPRECATED   AGE
10.23                     10.23     Official       ghcr.io/appscode-images/postgres:10.23-alpine                                        35d
10.23-bullseye            10.23     Official       ghcr.io/appscode-images/postgres:10.23-bullseye                                      35d
11-bullseye-postgis       11.22     PostGIS        postgis/postgis:11-3.3                                                               35d
11.22                     11.22     Official       ghcr.io/appscode-images/postgres:11.22-alpine                                        35d
11.22-bookworm            11.22     Official       ghcr.io/appscode-images/postgres:11.22-bookworm                                      35d
12-bullseye-postgis       12.18     PostGIS        postgis/postgis:12-3.4                                                               35d
12.17                     12.17     Official       ghcr.io/appscode-images/postgres:12.17-alpine                                        35d
12.17-bookworm            12.17     Official       ghcr.io/appscode-images/postgres:12.17-bookworm                                      35d
12.22                     12.22     Official       ghcr.io/appscode-images/postgres:12.22-alpine                                        35d
12.22-bookworm            12.22     Official       ghcr.io/appscode-images/postgres:12.22-bookworm                                      35d
13-bullseye-postgis       13.14     PostGIS        postgis/postgis:13-3.4                                                               35d
13.13                     13.13     Official       ghcr.io/appscode-images/postgres:13.13-alpine                                        35d
13.13-bookworm            13.13     Official       ghcr.io/appscode-images/postgres:13.13-bookworm                                      35d
13.18                     13.18     Official       ghcr.io/appscode-images/postgres:13.18-alpine                                        35d
13.18-bookworm            13.18     Official       ghcr.io/appscode-images/postgres:13.18-bookworm                                      35d
13.20                     13.20     Official       ghcr.io/appscode-images/postgres:13.20-alpine                                        35d
13.20-bookworm            13.20     Official       ghcr.io/appscode-images/postgres:13.20-bookworm                                      35d
14-bullseye-postgis       14.11     PostGIS        postgis/postgis:14-3.4                                                               35d
14.10                     14.10     Official       ghcr.io/appscode-images/postgres:14.10-alpine                                        35d
14.10-bookworm            14.10     Official       ghcr.io/appscode-images/postgres:14.10-bookworm                                      35d
14.13                     14.13     Official       ghcr.io/appscode-images/postgres:14.13-alpine                                        35d
14.13-bookworm            14.13     Official       ghcr.io/appscode-images/postgres:14.13-bookworm                                      35d
14.15                     14.15     Official       ghcr.io/appscode-images/postgres:14.15-alpine                                        35d
14.15-bookworm            14.15     Official       ghcr.io/appscode-images/postgres:14.15-bookworm                                      35d
14.17                     14.17     Official       ghcr.io/appscode-images/postgres:14.17-alpine                                        35d
14.17-bookworm            14.17     Official       ghcr.io/appscode-images/postgres:14.17-bookworm                                      35d
15-bullseye-postgis       15.6      PostGIS        postgis/postgis:15-3.4                                                               35d
15.10                     15.10     Official       ghcr.io/appscode-images/postgres:15.10-alpine                                        35d
15.10-bookworm            15.10     Official       ghcr.io/appscode-images/postgres:15.10-bookworm                                      35d
15.12                     15.12     Official       ghcr.io/appscode-images/postgres:15.12-alpine                                        35d
15.12-bookworm            15.12     Official       ghcr.io/appscode-images/postgres:15.12-bookworm                                      35d
15.12-documentdb          15.12     DocumentDB     ghcr.io/appscode-images/postgres-documentdb:15-0.102.0-ferretdb-2.0.0                35d
15.5                      15.5      Official       ghcr.io/appscode-images/postgres:15.5-alpine                                         35d
15.5-bookworm             15.5      Official       ghcr.io/appscode-images/postgres:15.5-bookworm                                       35d
15.8                      15.8      Official       ghcr.io/appscode-images/postgres:15.8-alpine                                         35d
15.8-bookworm             15.8      Official       ghcr.io/appscode-images/postgres:15.8-bookworm                                       35d
16.1                      16.1      Official       ghcr.io/appscode-images/postgres:16.1-alpine                                         35d
16.1-bookworm             16.1      Official       ghcr.io/appscode-images/postgres:16.1-bookworm                                       35d
16.2-bullseye-postgis     16.2      PostGIS        postgis/postgis:16-3.4                                                               35d
16.4                      16.4      Official       ghcr.io/appscode-images/postgres:16.4-alpine                                         35d
16.4-bookworm             16.4      Official       ghcr.io/appscode-images/postgres:16.4-bookworm                                       35d
16.6                      16.6      Official       ghcr.io/appscode-images/postgres:16.6-alpine                                         35d
16.6-bookworm             16.6      Official       ghcr.io/appscode-images/postgres:16.6-bookworm                                       35d
16.8                      16.8      Official       ghcr.io/appscode-images/postgres:16.8-alpine                                         35d
16.8-bookworm             16.8      Official       ghcr.io/appscode-images/postgres:16.8-bookworm                                       35d
16.8-documentdb           16.8      DocumentDB     ghcr.io/appscode-images/postgres-documentdb:16-0.102.0-ferretdb-2.0.0                35d
17.2                      17.2      Official       ghcr.io/appscode-images/postgres:17.2-alpine                                         35d
17.2-bookworm             17.2      Official       ghcr.io/appscode-images/postgres:17.2-bookworm                                       35d
17.4                      17.4      Official       ghcr.io/appscode-images/postgres:17.4-alpine                                         35d
17.4-bookworm             17.4      Official       ghcr.io/appscode-images/postgres:17.4-bookworm                                       35d
17.4-documentdb           17.4      DocumentDB     ghcr.io/appscode-images/postgres-documentdb:17-0.102.0-ferretdb-2.0.0                35d
timescaledb-2.14.2-pg13   13.14     TimescaleDB    timescale/timescaledb:2.14.2-pg13-oss                                                35d
timescaledb-2.14.2-pg14   14.11     TimescaleDB    timescale/timescaledb:2.14.2-pg14-oss                                                35d
timescaledb-2.14.2-pg15   15.6      Official       timescale/timescaledb:2.14.2-pg15-oss                                                35d
timescaledb-2.14.2-pg16   16.2      Official       timescale/timescaledb:2.14.2-pg16-oss                                                35d
```

> Updating this field creates a Postgres [`UpdateVersion`](/docs/guides/postgres/update-version/overview/index.md) OpsRequest by GitOps operator.

### spec.replicas

`spec.replicas` specifies the total number of primary and standby nodes in Postgres database cluster configuration. One pod is selected as Primary and others act as standby replicas. KubeDB uses `PodDisruptionBudget` to ensure that majority of the replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).

To learn more about how to setup a HA PostgreSQL cluster in KubeDB, please visit [here](/docs/guides/postgres/clustering/ha_cluster.md).

> Updating this field creates a [`HorizontalScaling`](/docs/guides/postgres/scaling/horizontal-scaling/overview/index.md) OpsRequest by GitOps operator.

### spec.standbyMode

`spec.standby` is an optional field that specifies the standby mode (_Warm / Hot_) to use for standby replicas. In **hot standby** mode, standby replicas can accept connection and run read-only queries. In **warm standby** mode, standby replicas can't accept connection and only used for replication purpose.

### spec.streamingMode

`spec.streamingMode` is an optional field that specifies the streaming mode (_Synchronous / Asynchronous_) of the standby replicas. KubeDB currently supports only **Asynchronous** streaming mode.

### spec.leaderElection

There are three fields under Postgres CRD's `spec.leaderElection`. These values defines how fast the leader election can happen.

- `leaseDurationSeconds`: This is the duration in seconds that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack. Default 15 sec.
- `renewDeadlineSeconds`: This is the duration in seconds that the acting master will retry refreshing leadership before giving up. Normally, LeaseDuration \* 2 / 3. Default 10 sec.
- `retryPeriodSeconds`: This is the duration in seconds the LeaderElector clients should wait between tries of actions. Normally, LeaseDuration / 3. Default 2 sec.

If the Cluster machine is powerful, user can reduce the times. But, Do not make it so little, in that case Postgres will restarts very often.

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

> Updating this field create a `RotateAuth` OpsRequest by GitOps operator.

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Postgres database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. In this case, you don't have to specify `spec.storage` field.

### spec.storage

If you don't set `spec.storageType:` to `Ephemeral` then `spec.storage` field is required. This field specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

> Updating this field creates a [`VolumeExpansion`](/docs/guides/postgres/volume-expansion/Overview/overview.md) OpsRequest by GitOps operator.

### spec.init

`spec.init` is an optional section that can be used to initialize a newly created Postgres database. PostgreSQL databases can be initialized from these three ways:

1. Initialize from Script
2. Initialize from Snapshot

#### Initialize via Script

To initialize a PostgreSQL database using a script (shell script, db migrator, etc.), set the `spec.init.script` section when creating a Postgres object. `script` must have the following information:

- [VolumeSource](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes): Where your script is loaded from.

Below is an example showing how a script from a configMap can be used to initialize a PostgreSQL database.

```yaml
apiVersion: gitops/kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: postgres-db
  namespace: demo
spec:
  version: "13.13"
  init:
    script:
      configMap:
        name: pg-init-script
```

In the above example, Postgres will execute provided script once the database is running. For more details tutorial on how to initialize from script, please visit [here](/docs/guides/postgres/initialization/script_source.md).

### spec.monitor

PostgreSQL managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor PostgreSQL with builtin Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md)
- [Monitor PostgreSQL with Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md)

> Enabling monitoring creates a [`Restart`](/docs/guides/postgres/restart/restart.md) OpsRequest by GitOps operator.

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for PostgreSQL. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). You can use any Kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/postgres/configuration/using-config-file.md).

> Updating this field will create a [`Reconfigure`](/docs/guides/postgres/reconfigure/overview.md) OpsRequest by GitOps operator.

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Postgres database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
    - annotations (pod's annotation)
- controller
    - annotations (petset's annotation)
- spec:
    - containers
    - volumes
    - podPlacementPolicy
    - serviceAccountName
    - initContainers
    - imagePullSecrets
    - nodeSelector
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

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the podPlacementPolicy. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.




#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the Postgres docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/_/postgres/).

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
    spec.authSecret
    spec.storageType
    spec.storage
    spec.podTemplate.spec.nodeSelector
    spec.init
```

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account name matching Postgres crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/postgres/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### spec.podTemplate.spec.imagePullSecrets

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry. For more details on how to use private docker registry, please visit [here](/docs/guides/postgres/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

> Updating `postgres`/`pg-coordinator` containers resources will create a [`VerticalScaling`](/docs/guides/postgres/scaling/vertical-scaling/overview/index.md) OpsRequest by GitOps operator.


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

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Postgres` crd or which resources KubeDB should keep or delete when you delete `Postgres` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to provide safety from accidental deletion of database. If admission webhook is enabled, KubeDB prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Postgres crd for different termination policies,

| Behavior                                 | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ---------------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation                |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Create Dormant Database               |    &#10007;    | &#10003; | &#10007; | &#10007; |
| 3. Delete PetSet                    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete Services                       |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 5. Delete PVCs                           |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 6. Delete Secrets                        |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshots                      |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 8. Delete Snapshot data from bucket      |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Halt` termination policy by default.

> Lastly, as `kubedb.com/v1` and `gitops.kubedb.com/v1alpha1` are two different API groups, you can use both of them in the same cluster. They share exactly the same spec, if you update some fields that might need changes using `OpsRequest`s, GitOps operator will create and reflect those changes in the database. You don't need to create ops request manually. If updating some fields don't require any ops request, GitOps operator will simply patch the actual `Postgres` database CRO with these changes.

## Next Steps

- Learn how to use KubeDB to run a PostgreSQL database [here](/docs/guides/postgres/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
