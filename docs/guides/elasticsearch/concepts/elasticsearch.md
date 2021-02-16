---
title: Elasticsearch CRD
menu:
  docs_{{ .version }}:
    identifier: es-elasticsearch-concepts
    name: Elasticsearch
    parent: es-concepts-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch

## What is Elasticsearch

`Elasticsearch` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [Elasticsearch](https://www.elastic.co/products/elasticsearch) in a Kubernetes native way. You only need to describe the desired database configuration in an Elasticsearch object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Elasticsearch Spec

As with all other Kubernetes objects, an Elasticsearch needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Elasticsearch object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
spec:
  authSecret:
    name: es-admin-cred
  configSecret:
    name: es-custom-config
  enableSSL: true
  internalUsers:
    metrics_exporter: {}
  rolesMapping:
    SGS_READALL_AND_MONITOR:
      users:
      - metrics_exporter
  kernelSettings:
    privileged: true
    sysctls:
    - name: vm.max_map_count
      value: "262144"
  maxUnavailable: 1
  monitor:
    agent: prometheus.io
    prometheus:
      exporter:
        port: 56790
  podTemplate:
    controller:
      annotations:
        passTo: statefulSets
    metadata:
      annotations:
        passTo: pods
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: es
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: elasticsearches.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: es
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: elasticsearches.kubedb.com
              namespaces:
              - demo
              topologyKey: failure-domain.beta.kubernetes.io/zone
            weight: 50
      env:
      - name: node.processors
        value: "2"
      nodeSelector:
        kubernetes.io/os: linux
      resources:
        limits:
          cpu: "1"
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 512Mi
      serviceAccountName: es
  replicas: 3
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passTo: service
    spec:
      type: NodePort
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: WipeOut
  tls:
    certificates:
    - alias: transport
      privateKey:
        encoding: PKCS8
      secretName: es-transport-cert
      subject:
        organizations:
        - kubedb
    - alias: http
      privateKey:
        encoding: PKCS8
      secretName: es-http-cert
      subject:
        organizations:
        - kubedb
    - alias: admin
      privateKey:
        encoding: PKCS8
      secretName: es-admin-cert
      subject:
        organizations:
        - kubedb
    - alias: metrics-exporter
      privateKey:
        encoding: PKCS8
      secretName: es-metrics-exporter-cert
      subject:
        organizations:
        - kubedb
  version: 7.9.3-searchguard
```

### spec.version

`spec.version` is a `required` field that specifies the name of the [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog.md) CRD where the docker images are specified.

- Name format: `{Application Version}-{Auth Plugin Name}-{Modification Tag}`

- Samples: `7.9.3-searchguard`, `7.9.1-xpack-v1`, `1.12.0-opendistro`, etc.

```yaml
spec:
  version: 7.9.3-searchguard
```

### spec.topology

`spec.topology` is an `optional` field that provides a way to configure different types of nodes for the Elasticsearch cluster. This field enables you to specify how many nodes you want to act as `master`, `data` and `ingest`. You can also specify how much storage and resources to allocate for each type of the nodes independently.

```yaml
  topology:
    data:
      maxUnavailable: 1
      replicas: 3
      resources:
        limits:
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: data
    ingest:
      maxUnavailable: 1
      replicas: 3
      resources:
        limits:
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: ingest
    master:
      maxUnavailable: 1
      replicas: 2
      resources:
        limits:
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: master
```

The `spec.topology` contains the following fields:

- `topology.master`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `master` nodes. Defaults to `1`.
  - `suffix` (`: "master"`) - is an `optional` field that is added as the suffix of the master StatefulSet name. Defaults to `master`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `master` nodes.
  - `resources` (` cpu: 500m, memory: 1Gi `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `master` nodes.
  - `maxUnavailable` is an `optional` field that specifies the exact number of master nodes (ie. pods) that can be safely evicted before the pod disruption budget (PDB) kicks in. KubeDB uses Pod Disruption Budget to ensure that desired number of replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that no data loss is occurred.

- `topology.data`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `data` nodes. Defaults to `1`.
  - `suffix` (`: "data"`) - is an `optional` field that is added as the suffix of the data StatefulSet name. Defaults to `data`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `data` nodes.
  - `resources` (` cpu: 500m, memory: 1Gi `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `data` nodes.
  - `maxUnavailable` is an `optional` field that specifies the exact number of data nodes (ie. pods) that can be safely evicted before the pod disruption budget (PDB) kicks in. KubeDB uses Pod Disruption Budget to ensure that desired number of replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that no data loss is occurred.

- `topology.ingest`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `ingest` nodes. Defaults to `1`.
  - `suffix` (`: "ingest"`) - is an `optional` field that is added as the suffix of the data StatefulSet name. Defaults to `ingest`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `ingest` nodes.
  - `resources` (` cpu: 500m, memory: 1Gi `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `data` nodes.
  - `maxUnavailable` is an `optional` field that specifies the exact number of ingest nodes (ie. pods) that can be safely evicted before the pod disruption budget (PDB) kicks in. KubeDB uses Pod Disruption Budget to ensure that desired number of replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that no data loss is occurred.

> Note: Any two types of nodes can't have the same suffix.

If you specify `spec.topology` field then you **do not need** to specify the following fields in Elasticsearch CRD.

- `spec.replicas`
- `spec.storage`
- `spec.podTemplate.spec.resources`

If you do not specify `spec.topology` field, the Elasticsearch Cluster runs in combined mode.

> Combined Mode: all nodes of the Elasticsearch cluster will work as `master`, `data` and `ingest` nodes simultaneously.

### spec.replicas

`spec.replicas` is an `optional` field that can be used if `spec.topology` is not specified. This field specifies the number of nodes (ie. pods) in the Elasticsearch cluster. The default value of this field is `1`.

```yaml
spec:
  replicas: 3
```

### spec.maxUnavailable

`spec.maxUnavailable` is an `optional` field that is used to specify the exact number of cluster replicas that can be safely evicted before pod disruption budget kicks in to prevent unwanted data loss.

```yaml
spec:
  maxUnavailable: 1
```

### spec.enableSSL

`spec.enableSSL` is an `optional` field that specifies whether to enable TLS to HTTP layer. The default value of this field is `false`.

```yaml
spec:
  enableSSL: true 
```

### spec.authSecret

`spec.authSecret` is an `optional` field that points to a k8s secret used to hold the Elasticsearch `elastic`/`admin` user credentials.

```yaml
spec:
  authSecret:
    name: es-admin-cred
```

The k8s secret must be of `type: kubernetes.io/basic-auth` with the following keys:

- `username`: Must be `elastic` for x-pack, or `admin` for searchGuard and OpenDistro.
- `password`: Password for the `elastic`/`admin` user.

If not set, the KubeDB operator creates a new Secret `{Elasticsearch name}-{UserName}-cred` with randomly generated secured credentials.

### spec.storageType

`spec.storageType` is an `optional` field that specifies the type of storage to use for the database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Elasticsearch database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. In this case, you don't have to specify `spec.storage` field.

```yaml
spec:
  storageType: Durable
```

### spec.storage

If the `spec.storageType`  is not set to `Ephemeral` and if the `spec.topology` field also is not set then `spec.storage` field is `required`. This field specifies the StorageClass of the PVCs dynamically allocated to store data for the database. This storage spec will be passed to the StatefulSet created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

```yaml
spec:
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
```

- `storage.storageClassName` - is the name of the StorageClass used to provision the PVCs. The PVCs don’t necessarily have to request a class. A PVC with the storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.init

`spec.init` is an `optional` section that can be used to initialize a newly created Elasticsearch cluster from prior snapshots, taken by [Stash](/docs/guides/elasticsearch/backup/stash.md).

```yaml
spec:
  init:
    waitForInitialRestore: true
```

When the `waitForInitialRestore` is set to true, the Elasticsearch instance will be stack in the `Provisioning` state until the initial backup is completed. On completion of the very first backup the Elasticsearch instance will go to the `Ready` state.

For detailed tutorial on how to initialize Elasticsearch from Stash backup, please visit [here](/docs/guides/elasticsearch/backup/stash.md).

### spec.monitor

Elasticsearch managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor Elasticsearch with builtin Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md)
- [Monitor Elasticsearch with Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an `optional` field that allows users to provide custom configuration for Elasticsearch. It contains a k8s secret name that holds the configuration files for both Elasticsearch and the security plugins (ie. x-pack, SearchGuard, and openDistro).

```yaml
spec:
  configSecret:
    name: es-custom-config
```

The configuration file names are used as secret keys.

**Elasticsearch:**

- `elasticsearch.yml` - for configuring Elasticsearch
- `jvm.options` - for configuring Elasticsearch JVM settings
- `log4j2.properties` - for configuring Elasticsearch logging

**X-Pack:**

- `roles.yml` - define roles and the associated permissions.
- `role_mapping.yml` - define which roles should be assigned to each user based on their username, groups, or other metadata.

**SearchGuard:**

- `sg_config.yml` - configure authenticators and authorization backends.
- `sg_roles.yml` - define roles and the associated permissions.
- `sg_roles_mapping.yml` - map backend roles, hosts and users to roles.
- `sg_internal_users.yml` - stores users,and hashed passwords in the internal user database.
- `sg_action_groups.yml` - define named permission groups.
- `sg_tenants.yml` - defines tenants for configuring the Kibana access.
- `sg_blocks.yml` -  defines blocked users and IP addresses.

**OpenDistro:**

- `internal_users.yml` - contains any initial users that you want to add to the security plugin’s internal user database.
- `roles.yml` - contains any initial roles that you want to add to the security plugin.
- `roles_mapping.yml` - maps backend roles, hosts and users to roles.
- `action_groups.yml` - contains any initial action groups that you want to add to the security plugin.
- `tenants.yml` - contains the tenant configurations.
- `nodes_dn.yml` - contains nodesDN mapping name and corresponding values.

**How the resultant configuration files are generated?**

- `YML`: The default configuration file pre-stored at config directories is overwritten by the operator generated configuration file. Then the resultant configuration file is overwritten by the user provided custom configuration file (if any). The [yq](https://github.com/mikefarah/yq) tool is used to merge two yaml files.

  ```bash
  $ yq merge -i --overwrite file1.yml file2.yml
  ```

- `Non-YML`: The default configuration file is replaced by the operator generated one (if any). Then the resultant configuration file is replaced by the user provided custom configuration file (if any).

  ```bash
  $ cp -f file2 file1
  ```

**How to provide node-role specific configurations?**

If an Elasticsearch cluster is running in the topology mode (ie. `spec.topology` is set), user may want to provide node-role specific configurations, say configurations that will only be merged to `master` nodes. To achieve this, users need to add the node role as prefix to file name.

- Format: `<node-role>-<file-name>.extension`
- Samples:
  - `data-elasticsearch.yml`: Only applied to `data` nodes.
  - `master-jvm.options`: Only applied to `master` nodes.
  - `ingest-log4j2.properties`: Only applied to `ingest` nodes.

**How to provide additional files that is referenced from the configurations?**

All these files provided via `configSecret` is stored in each Elasticsearch node (i.e. pod) at `config_dir/custom_config/` ( i.e. `/usr/share/elasticsearch/config/custom_config/`) directory. So, user can use this path while configuring the Elasticsearch.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-custom-config
  namespace: demo
stringData:
  elasticsearch.yml: |-
    logger.org.elasticsearch.discovery: DEBUG
```

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the StatefulSet created for Elasticsearch database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
  - annotations (pod's annotation)
- controller
  - annotations (statefulset's annotation)
- spec:
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

Uses of some fields of `spec.podTemplate` are described below,

#### spec.podTemplate.spec.env

`spec.podTemplate.spec.env` is an optional field that specifies the environment variables to pass to the Elasticsearch docker image. To know about supported environment variables, please visit [here](https://github.com/pires/docker-elasticsearch#environment-variables).

A list of the supported environment variables, their permission to use in KubeDB and their default value is given below.

|      Environment variables      | Permission to use |                                           Default value                                            |
| ------------------------------- | :---------------: | -------------------------------------------------------------------------------------------------- |
| CLUSTER_NAME                    |     `allowed`     | `metadata.name`                                                                                    |
| NODE_NAME                       |   `not allowed`   | Pod name                                                                                           |
| NODE_MASTER                     |   `not allowed`   | KubeDB sets it based on `Elasticsearch` CRD sepcification                                           |
| NODE_DATA                       |   `not allowed`   | KubeDB sets it based on `Elasticsearch` CRD sepcification                                           |
| NETWORK_HOST                    |     `allowed`     | `_site_`                                                                                           |
| HTTP_ENABLE                     |     `allowed`     | If `spec.topology` is not specified then `true`. Otherwise, `false` for Master node and Data node. |
| HTTP_CORS_ENABLE                |     `allowed`     | `true`                                                                                             |
| HTTP_CORS_ALLOW_ORIGIN          |     `allowed`     | `*`                                                                                                |
| NUMBER_OF_MASTERS               |     `allowed`     | `(replicas/2)+1`                                                                                   |
| MAX_LOCAL_STORAGE_NODES         |     `allowed`     | `1`                                                                                                |
| ES_JAVA_OPTS                    |     `allowed`     | `-Xms128m -Xmx128m`                                                                                |
| ES_PLUGINS_INSTALL              |     `allowed`     | Not set                                                                                            |
| SHARD_ALLOCATION_AWARENESS      |     `allowed`     | `""`                                                                                               |
| SHARD_ALLOCATION_AWARENESS_ATTR |     `allowed`     | `""`                                                                                               |
| MEMORY_LOCK                     |     `allowed`     | `true`                                                                                             |
| REPO_LOCATIONS                  |     `allowed`     | `""`                                                                                               |
| PROCESSORS                      |     `allowed`     | `1`                                                                                                |

Note that, KubeDB does not allow `NODE_NAME`, `NODE_MASTER`, and `NODE_DATA` environment variables to set in `spec.podTemplate.spec.env`. KubeDB operator set them based on Elasticsearch CRD specification.

If you try to set any these forbidden environment variable in Elasticsearch CRD, KubeDB operator will reject the request with following error,

```ini
Error from server (Forbidden): error when creating "./elasticsearch.yaml": admission webhook "elasticsearch.validators.kubedb.com" denied the request: environment variable NODE_NAME is forbidden to use in Elasticsearch spec
```

Also, note that KubeDB does not allow to update the environment variables as updating them does not have any effect once the database is created.  If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./elasticsearch.yaml": admission webhook "elasticsearch.validators.kubedb.com" denied the request: precondition failed for:
...
At least one of the following was changed:
    apiVersion
    kind
    name
    namespace
    spec.version
    spec.topology.*.prefix
    spec.topology.*.storage
    spec.enableSSL
    spec.certificateSecret
    spec.authSecret
    spec.storageType
    spec.storage
    spec.nodeSelector
    spec.init
    spec.env
```

#### spec.podTemplate.spec.imagePullSecrets

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image when you are using a private docker registry. For more details on how to use private docker registry, please visit [here](/docs/guides/elasticsearch/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

  `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

  If this field is left empty, the KubeDB operator will create a service account name matching Elasticsearch CRD name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

  If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

  If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/elasticsearch/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. If you didn't specify `spec.topology` field then this can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplate

You can also provide a template for the services created by KubeDB operator for Elasticsearch database through `spec.serviceTemplate`. This will allow you to set the type and other properties of the services.

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

### spec.terminationPolicy

`terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Elasticsearch` CRD or which resources KubeDB should keep or delete when you delete `Elasticsearch` CRD. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `terminationPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to provide safety from accidental deletion of database. If admission webhook is enabled, KubeDB prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Elasticsearch CRD for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Create Dormant Database          |    &#10007;    | &#10003; | &#10007; | &#10007; |
| 3. Delete StatefulSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 5. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 6. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 8. Delete Snapshot data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.terminationPolicy` KubeDB uses `Halt` termination policy by default.

## Next Steps

- Learn how to use KubeDB to run an Elasticsearch database [here](/docs/guides/elasticsearch/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
