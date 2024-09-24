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

[//]: # ()
[//]: # (## What is PgBouncer)

[//]: # ()
[//]: # (`PgBouncer` is a Kubernetes `Custom Resource Definitions` &#40;CRD&#41;. It provides declarative configuration for [PgBouncer]&#40;https://www.pgbouncer.github.io/&#41; in a Kubernetes native way. You only need to describe the desired configurations in a `PgBouncer` object, and the KubeDB operator will create Kubernetes resources in the desired state for you.)

[//]: # ()
[//]: # (## PgBouncer Spec)

[//]: # ()
[//]: # (Like any official Kubernetes resource, a `PgBouncer` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.)

[//]: # ()
[//]: # (Below is an example PgBouncer object.)

[//]: # ()
[//]: # (```yaml)

[//]: # (apiVersion: kubedb.com/v1alpha2)

[//]: # (kind: PgBouncer)

[//]: # (metadata:)

[//]: # (  name: pgbouncer-server)

[//]: # (  namespace: demo)

[//]: # (spec:)

[//]: # (  version: "1.18.0")

[//]: # (  replicas: 2)

[//]: # (  databases:)

[//]: # (  - alias: "postgres")

[//]: # (    databaseName: "postgres")

[//]: # (    databaseRef:)

[//]: # (      name: "quick-postgres")

[//]: # (      namespace: demo)

[//]: # (  connectionPool:)

[//]: # (    maxClientConnections: 20)

[//]: # (    reservePoolSize: 5)

[//]: # (  monitor:)

[//]: # (    agent: prometheus.io/operator)

[//]: # (    prometheus:)

[//]: # (      serviceMonitor:)

[//]: # (        labels:)

[//]: # (          release: prometheus)

[//]: # (        interval: 10s)

[//]: # (```)

[//]: # ()
[//]: # (### spec.version)

[//]: # ()
[//]: # (`spec.version` is a required field that specifies the name of the [PgBouncerVersion]&#40;/docs/guides/pgbouncer/concepts/catalog.md&#41; crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `PgBouncerVersion` resources,)

[//]: # ()
[//]: # (- `1.18.0`)

[//]: # ()
[//]: # (### spec.replicas)

[//]: # ()
[//]: # (`spec.replicas` specifies the total number of available pgbouncer server nodes for each crd. KubeDB uses `PodDisruptionBudget` to ensure that majority of the replicas are available during [voluntary disruptions]&#40;https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions&#41;.)

[//]: # ()
[//]: # (### spec.databases)

[//]: # ()
[//]: # (`spec.databases` specifies an array of postgres databases that pgbouncer should add to its connection pool. It contains three `required` fields and two `optional` fields for each database connection.)

[//]: # ()
[//]: # (- `spec.databases.alias`:  specifies an alias for the target database located in a postgres server specified by an appbinding.)

[//]: # (- `spec.databases.databaseName`:  specifies the name of the target database.)

[//]: # (- `spec.databases.databaseRef`:  specifies the name and namespace of the AppBinding that contains the path to a PostgreSQL server where the target database can be found.)

[//]: # ()
[//]: # (ConnectionPool is used to configure pgbouncer connection-pool. All the fields here are accompanied by default values and can be left unspecified if no customisation is required by the user.)

[//]: # ()
[//]: # (- `spec.connectionPool.port`: specifies the port on which pgbouncer should listen to connect with clients. The default is 5432.)

[//]: # ()
[//]: # (- `spec.connectionPool.poolMode`: specifies the value of pool_mode. Specifies when a server connection can be reused by other clients.)

[//]: # ()
[//]: # (  - session)

[//]: # ()
[//]: # (    Server is released back to pool after client disconnects. Default.)

[//]: # ()
[//]: # (  - transaction)

[//]: # ()
[//]: # (    Server is released back to pool after transaction finishes.)

[//]: # ()
[//]: # (  - statement)

[//]: # ()
[//]: # (    Server is released back to pool after query finishes. Long transactions spanning multiple statements are disallowed in this mode.)

[//]: # ()
[//]: # (- `spec.connectionPool.maxClientConnections`: specifies the value of max_client_conn. When increased then the file descriptor limits should also be increased. Note that actual number of file descriptors used is more than max_client_conn. Theoretical maximum used is:)

[//]: # ()
[//]: # (  ```bash)

[//]: # (  max_client_conn + &#40;max pool_size * total databases * total users&#41;)

[//]: # (  ```)

[//]: # ()
[//]: # (  if each user connects under its own username to server. If a database user is specified in connect string &#40;all users connect under same username&#41;, the theoretical maximum is:)

[//]: # ()
[//]: # (  ```bash)

[//]: # (  max_client_conn + &#40;max pool_size * total databases&#41;)

[//]: # (  ```)

[//]: # ()
[//]: # (  The theoretical maximum should be never reached, unless somebody deliberately crafts special load for it. Still, it means you should set the number of file descriptors to a safely high number.)

[//]: # ()
[//]: # (  Search for `ulimit` in your favorite shell man page. Note: `ulimit` does not apply in a Windows environment.)

[//]: # ()
[//]: # (  Default: 100)

[//]: # ()
[//]: # (- `spec.connectionPool.defaultPoolSize`: specifies the value of default_pool_size. Used to determine how many server connections to allow per user/database pair. Can be overridden in the per-database configuration.)

[//]: # ()
[//]: # (  Default: 20)

[//]: # ()
[//]: # (- `spec.connectionPool.minPoolSize`: specifies the value of min_pool_size. PgBouncer adds more server connections to pool if below this number. Improves behavior when usual load comes suddenly back after period of total inactivity.)

[//]: # ()
[//]: # (  Default: 0 &#40;disabled&#41;)

[//]: # ()
[//]: # (- `spec.connectionPool.reservePoolSize`: specifies the value of reserve_pool_size. Used to determine how many additional connections to allow to a pool. 0 disables.)

[//]: # ()
[//]: # (  Default: 0 &#40;disabled&#41;)

[//]: # ()
[//]: # (- `spec.connectionPool.reservePoolTimeout`: specifies the value of reserve_pool_timeout. If a client has not been serviced in this many seconds, pgbouncer enables use of additional connections from reserve pool. 0 disables.)

[//]: # ()
[//]: # (  Default: 5.0)

[//]: # ()
[//]: # (- `spec.connectionPool.maxDbConnections`: specifies the value of max_db_connections. PgBouncer does not allow more than this many connections per-database &#40;regardless of pool - i.e. user&#41;. It should be noted that when you hit the limit, closing a client connection to one pool will not immediately allow a server connection to be established for another pool, because the server connection for the first pool is still open. Once the server connection closes &#40;due to idle timeout&#41;, a new server connection will immediately be opened for the waiting pool.)

[//]: # ()
[//]: # (  Default: unlimited)

[//]: # ()
[//]: # (- `spec.connectionPool.maxUserConnections`: specifies the value of max_user_connections. PgBouncer does not allow more than this many connections per-user &#40;regardless of pool - i.e. user&#41;. It should be noted that when you hit the limit, closing a client connection to one pool will not immediately allow a server connection to be established for another pool, because the server connection for the first pool is still open. Once the server connection closes &#40;due to idle timeout&#41;, a new server connection will immediately be opened for the waiting pool.)

[//]: # (  Default: unlimited)

[//]: # ()
[//]: # (- `spec.connectionPool.statsPeriod`: sets how often the averages shown in various `SHOW` commands are updated and how often aggregated statistics are written to the log.)

[//]: # (  Default: 60)

[//]: # ()
[//]: # (- `spec.connectionPool.authType`: specifies how to authenticate users. PgBouncer supports several authentication methods including pam, md5, scram-sha-256, trust , or any. However hba, and cert are not supported.)

[//]: # ()
[//]: # (- `spec.connectionPool.IgnoreStartupParameters`: specifies comma-separated startup parameters that pgbouncer knows are handled by admin and it can ignore them.)

[//]: # ()
[//]: # (### spec.monitor)

[//]: # ()
[//]: # (PgBouncer managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,)

[//]: # ()
[//]: # (- [Monitor PgBouncer with builtin Prometheus]&#40;/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md&#41;)

[//]: # (- [Monitor PgBouncer with Prometheus operator]&#40;/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md&#41;)

[//]: # ()
[//]: # (### spec.podTemplate)

[//]: # ()
[//]: # (KubeDB allows providing a template for pgbouncer pods through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for PgBouncer server)

[//]: # ()
[//]: # (KubeDB accept following fields to set in `spec.podTemplate:`)

[//]: # ()
[//]: # (- metadata)

[//]: # (  - annotations &#40;pod's annotation&#41;)

[//]: # (- controller)

[//]: # (  - annotations &#40;petset's annotation&#41;)

[//]: # (- spec:)

[//]: # (  - env)

[//]: # (  - resources)

[//]: # (  - initContainers)

[//]: # (  - imagePullSecrets)

[//]: # (  - affinity)

[//]: # (  - tolerations)

[//]: # (  - priorityClassName)

[//]: # (  - priority)

[//]: # (  - lifecycle)

[//]: # ()
[//]: # (Usage of some fields in `spec.podTemplate` is described below,)

[//]: # ()
[//]: # (#### spec.podTemplate.spec.env)

[//]: # ()
[//]: # (`spec.podTemplate.spec.env` is an optional field that specifies the environment variables to pass to the PgBouncer docker image. To know about supported environment variables, please visit [here]&#40;https://hub.docker.com/kubedb/pgbouncer/&#41;.)

[//]: # ()
[//]: # (Also, note that KubeDB does not allow updates to the environment variables as updating them does not have any effect once the server is created. If you try to update environment variables, KubeDB operator will reject the request with following error,)

[//]: # ()
[//]: # (```ini)

[//]: # (Error from server &#40;BadRequest&#41;: error when applying patch:)

[//]: # (...)

[//]: # (for: "./pgbouncer.yaml": admission webhook "pgbouncer.validators.kubedb.com" denied the request: precondition failed for:)

[//]: # (...)

[//]: # (At least one of the following was changed:)

[//]: # (    apiVersion)

[//]: # (    kind)

[//]: # (    name)

[//]: # (    namespace)

[//]: # (    spec.podTemplate.spec.nodeSelector)

[//]: # (```)

[//]: # ()
[//]: # (#### spec.podTemplate.spec.imagePullSecrets)

[//]: # ()
[//]: # (`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry. For more details on how to use private docker registry, please visit [here]&#40;/docs/guides/pgbouncer/private-registry/using-private-registry.md&#41;.)

[//]: # ()
[//]: # (#### spec.podTemplate.spec.nodeSelector)

[//]: # ()
[//]: # (`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels &#40;it can have additional labels as well&#41;. To learn more, see [here]&#40;https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector&#41; .)

[//]: # ()
[//]: # (#### spec.podTemplate.spec.resources)

[//]: # ()
[//]: # (`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here]&#40;http://kubernetes.io/docs/user-guide/compute-resources/&#41;.)

[//]: # ()
[//]: # (### spec.serviceTemplate)

[//]: # ()
[//]: # (KubeDB creates a service for each PgBouncer instance. The service has the same name as the `pgbouncer.name` and points to pgbouncer pods.)

[//]: # ()
[//]: # (You can provide template for this service using `spec.serviceTemplate`. This will allow you to set the type and other properties of the service. If `spec.serviceTemplate` is not provided, KubeDB will create a service of type `ClusterIP` with minimal settings.)

[//]: # ()
[//]: # (KubeDB allows the following fields to set in `spec.serviceTemplate`:)

[//]: # ()
[//]: # (- metadata:)

[//]: # (  - annotations)

[//]: # (- spec:)

[//]: # (  - type)

[//]: # (  - ports)

[//]: # (  - clusterIP)

[//]: # (  - externalIPs)

[//]: # (  - loadBalancerIP)

[//]: # (  - loadBalancerSourceRanges)

[//]: # (  - externalTrafficPolicy)

[//]: # (  - healthCheckNodePort)

[//]: # (  - sessionAffinityConfig)

[//]: # ()
[//]: # (See [here]&#40;https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163&#41; to understand these fields in detail.)

[//]: # ()
[//]: # (## Next Steps)

[//]: # ()
[//]: # (- Learn how to use KubeDB to run a PostgreSQL database [here]&#40;/docs/guides/postgres/README.md&#41;.)

[//]: # (- Learn how to how to get started with PgBouncer [here]&#40;/docs/guides/pgbouncer/quickstart/quickstart.md&#41;.)

[//]: # (- Want to hack on KubeDB? Check our [contribution guidelines]&#40;/docs/CONTRIBUTING.md&#41;.)
