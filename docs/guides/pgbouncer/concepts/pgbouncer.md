---
title: PgBouncer CRD
menu:
  docs_{{ .version }}:
    identifier: pb-pgbouncer-concepts
    name: PgBouncer
    parent: pb-concepts-pgbouncer
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PgBouncer

## What is PgBouncer

`PgBouncer` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [PgBouncer](https://www.pgbouncer.github.io/) in a Kubernetes native way. You only need to describe the desired configurations in a `PgBouncer` object, and the KubeDB operator will create Kubernetes resources in the desired state for you.

## PgBouncer Spec

Like any official Kubernetes resource, a `PgBouncer` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Below is an example PgBouncer object.

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pgbouncer-server
  namespace: demo
spec:
  version: "1.18.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
      namespace: demo
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

### spec.version

`spec.version` is a required field that specifies the name of the [PgBouncerVersion](/docs/guides/pgbouncer/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `PgBouncerVersion` resources,

- `1.18.0`

### spec.replicas

`spec.replicas` specifies the total number of available pgbouncer server nodes for each crd. KubeDB uses `PodDisruptionBudget` to ensure that majority of the replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).

### spec.database

`spec.database` specifies a single postgres database that pgbouncer should add to its connection pool. It contains two `required` fields and one `optional` field for database connection.

- `spec.databases.syncUsers`:  specifies whether the pgbouncer should collect usernames and passwords from external secrets with specified labels.
- `spec.databases.databaseName`:  specifies the name of the target database.
- `spec.databases.databaseRef`:  specifies the name and namespace of the AppBinding that contains the path to a PostgreSQL server where the target database can be found.

ConnectionPool is used to configure pgbouncer connection-pool. All the fields here are accompanied by default values and can be left unspecified if no customisation is required by the user.

- `spec.connectionPool.port`: specifies the port on which pgbouncer should listen to connect with clients. The default is 5432.

- `spec.connectionPool.poolMode`: specifies the value of pool_mode. Specifies when a server connection can be reused by other clients.

  - session

    Server is released back to pool after client disconnects. Default.

  - transaction

    Server is released back to pool after transaction finishes.

  - statement

    Server is released back to pool after query finishes. Long transactions spanning multiple statements are disallowed in this mode.

- `spec.connectionPool.maxClientConnections`: specifies the value of max_client_conn. When increased then the file descriptor limits should also be increased. Note that actual number of file descriptors used is more than max_client_conn. Theoretical maximum used is:

  ```bash
  max_client_conn + (max pool_size * total databases * total users)
  ```

  if each user connects under its own username to server. If a database user is specified in connect string (all users connect under same username), the theoretical maximum is:

  ```bash
  max_client_conn + (max pool_size * total databases)
  ```

  The theoretical maximum should be never reached, unless somebody deliberately crafts special load for it. Still, it means you should set the number of file descriptors to a safely high number.

  Search for `ulimit` in your favorite shell man page. Note: `ulimit` does not apply in a Windows environment.

  Default: 100

- `spec.connectionPool.defaultPoolSize`: specifies the value of default_pool_size. Used to determine how many server connections to allow per user/database pair. Can be overridden in the per-database configuration.

  Default: 20

- `spec.connectionPool.minPoolSize`: specifies the value of min_pool_size. PgBouncer adds more server connections to pool if below this number. Improves behavior when usual load comes suddenly back after period of total inactivity.

  Default: 0 (disabled)

- `spec.connectionPool.reservePoolSize`: specifies the value of reserve_pool_size. Used to determine how many additional connections to allow to a pool. 0 disables.

  Default: 0 (disabled)

- `spec.connectionPool.reservePoolTimeout`: specifies the value of reserve_pool_timeout. If a client has not been serviced in this many seconds, pgbouncer enables use of additional connections from reserve pool. 0 disables.

  Default: 5.0

- `spec.connectionPool.maxDbConnections`: specifies the value of max_db_connections. PgBouncer does not allow more than this many connections per-database (regardless of pool - i.e. user). It should be noted that when you hit the limit, closing a client connection to one pool will not immediately allow a server connection to be established for another pool, because the server connection for the first pool is still open. Once the server connection closes (due to idle timeout), a new server connection will immediately be opened for the waiting pool.

  Default: unlimited

- `spec.connectionPool.maxUserConnections`: specifies the value of max_user_connections. PgBouncer does not allow more than this many connections per-user (regardless of pool - i.e. user). It should be noted that when you hit the limit, closing a client connection to one pool will not immediately allow a server connection to be established for another pool, because the server connection for the first pool is still open. Once the server connection closes (due to idle timeout), a new server connection will immediately be opened for the waiting pool.
  Default: unlimited

- `spec.connectionPool.statsPeriod`: sets how often the averages shown in various `SHOW` commands are updated and how often aggregated statistics are written to the log.
  Default: 60

- `spec.connectionPool.authType`: specifies how to authenticate users. PgBouncer supports several authentication methods including pam, md5, scram-sha-256, trust , or any. However hba, and cert are not supported.

- `spec.connectionPool.IgnoreStartupParameters`: specifies comma-separated startup parameters that pgbouncer knows are handled by admin and it can ignore them.

### spec.monitor

PgBouncer managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor PgBouncer with builtin Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md)
- [Monitor PgBouncer with Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md)

### spec.podTemplate

KubeDB allows providing a template for pgbouncer pods through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for PgBouncer server

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
  - annotations (pod's annotation)
- controller
  - annotations (petset's annotation)
- spec:
  - containers
  - volumes
  - podPlacementPolicy
  - initContainers
  - imagePullSecrets
  - tolerations
  - priorityClassName
  - priority

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go#L26C1-L279C1).
Usage of some fields in `spec.podTemplate` is described below,

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

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the PgBouncer containers. To know about supported environment variables, please visit [here](https://hub.docker.com/kubedb/pgbouncer/).

Also, note that KubeDB does not allow updates to the environment variables as updating them does not have any effect once the server is created. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./pgbouncer.yaml": admission webhook "pgbouncer.validators.kubedb.com" denied the request: precondition failed for:
...
At least one of the following was changed:
    apiVersion
    kind
    name
    namespace
    spec.podTemplate.spec.nodeSelector
```

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

#### spec.podTemplate.spec.imagePullSecrets

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry. For more details on how to use private docker registry, please visit [here](/docs/guides/pgbouncer/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

### spec.serviceTemplates

KubeDB creates a service for each PgBouncer instance. The service has the same name as the `pgbouncer.name` and points to pgbouncer pods.

You can provide template for this service using `spec.serviceTemplates`. This will allow you to set the type and other properties of the service. If `spec.serviceTemplates` is not provided, KubeDB will create a service of type `ClusterIP` with minimal settings.

KubeDB allows the following fields to set in `spec.serviceTemplates`:

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

## Next Steps

- Learn how to use KubeDB to run a PostgreSQL database [here](/docs/guides/postgres/README.md).
- Learn how to how to get started with PgBouncer [here](/docs/guides/pgbouncer/quickstart/quickstart.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
