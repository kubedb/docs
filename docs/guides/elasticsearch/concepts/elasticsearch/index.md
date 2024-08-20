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

`Elasticsearch` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [Elasticsearch](https://www.elastic.co/products/elasticsearch) and [OpenSearch](https://opensearch.org/) in a Kubernetes native way. You only need to describe the desired database configuration in an Elasticsearch object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Elasticsearch Spec

As with all other Kubernetes objects, an Elasticsearch needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Elasticsearch object.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: myes
  namespace: demo
spec:
  autoOps:
    disabled: true
  authSecret:
    name: es-admin-cred
    externallyManaged: false
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
        passTo: petSets
    metadata:
      annotations:
        passTo: pods
    spec:
      nodeSelector:
        kubernetes.io/os: linux
      containers:
      - name: elasticsearch
        env:
        - name: node.processors
          value: "2"
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
  deletionPolicy: WipeOut
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: es-issuer
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
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 2
    disableWriteCheck: false
  version: xpack-8.11.1
```
### spec.autoOps
AutoOps is an optional field to control the generation of versionUpdate & TLS-related recommendations.

### spec.version
`spec.version` is a `required` field that specifies the name of the [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog/index.md) CRD where the docker images are specified.

- Name format: `{Security Plugin Name}-{Application Version}-{Modification Tag}`

- Samples: `xpack-8.2.3`, `xpack-8.11.1`, `opensearch-1.3.0`, etc.

```yaml
spec:
  version: xpack-8.11.1
```

### spec.kernelSettings

`spec.kernelSettings` is an `optional` field that is used to configure the k8s-cluster node's kernel settings. It let users to perform `sysctl -w key=value` commands to the node's kernel. These commands are performed from an `initContainer`. If any of those commands require `privileged` access, you need to set the `kernelSettings.privileged` to `true` resulting in the `initContainer` running in `privileged` mode.

```yaml
spec:
  kernelSettings:
    privileged: true
    sysctls:
    - name: vm.max_map_count
      value: "262144"
```

To disable the kernetSetting `initContainer`, set the `kernelSettings.disableDefaults` to `true` .

```yaml
spec:
  kernelSettings:
    disableDefaults: true
```

> Note: Make sure that `vm.max_map_count` is greater or equal to `262144`, otherwise the Elasticsearch may fail to bootstrap.


### spec.disableSecurity

`spec.disableSecurity` is an `optional` field that allows a user to run the Elasticsearch with the security plugin `disabled`. Default to `false`.

```yaml
spec:
  disableSecurity: true
```

### spec.internalUsers

`spec.internalUsers` provides an alternative way to configure the existing internal users or create new users without using the `internal_users.yml` file. This field expects the input format to be in the `map[username]ElasticsearchUserSpec` format. The KubeDB operator creates and synchronizes secure passwords for those users and stores in k8s secrets.  The k8s secret names are formed by the following format: `{Elasticsearch Instance Name}-{Username}-cred`.

The `ElasticsearchUserSpec` contains the following fields:
-  `hash` ( `string` | `""` ) - Specifies the hash of the password.
-  `full_name` ( `string` | `""` ) - Specifies The full name of the user. Only applicable for xpack authplugin.
-  `metadata` ( `map[string]string` | `""` ) - Specifies Arbitrary metadata that you want to associate with the user. Only applicable for xpack authplugin.
-  `secretName` ( `string` | `""` ) - Specifies the k8s secret name that holds the user credentials. Defaults to "<resource-name>-<username>-cred".
-  `roles` ( `[]string` | `nil` ) - A set of roles the user has. The roles determine the user’s access permissions. To create a user without any roles, specify an empty list: []. Only applicable for xpack authplugin.
-  `email` ( `string` | `""` ) - Specifies the email of the user. Only applicable for xpack authplugin.
-  `reserved` ( `bool` | `false` ) - specifies the reserved status. The resources that have this set to `true` cannot be changed using the REST API or Kibana.
-  `hidden` ( `bool` | `false` ) - specifies the hidden status. The resources that have this set to true are not returned by the REST API and not visible in Kibana.
-  `backendRoles` (`[]string` | `nil`) - specifies a list of backend roles assigned to this user. The backend roles can come from the internal user database, LDAP groups, JSON web token claims, or SAML assertions.
-  `searchGuardRoles` ( `[]string` | `nil` ) - specifies a list of SearchGuard security plugin roles assigned to this user.
-  `opendistroSecurityRoles` ( `[]string` | `nil` ) - specifies a list of opendistro security plugin roles assigned to this user.
-  `attributes` ( `map[string]string` | `nil` )- specifies one or more custom attributes which can be used in index names and DLS queries.
-  `description` ( `string` | `""` ) - specifies the description of the user.

Here's how `.spec.internalUsers` can be configured for `searchguard` or `opendistro` auth plugins.

```yaml
spec:
  internalUsers:
    # update the attribute of default kibanaro user
    kibanaro: 
      attributes:
        attribute1: "value-a"
        attribute2: "value-b"
        attribute3: "value-c"
    # update the desciption of snapshotrestore user
    snapshotrestore: 
      description: "This is the new description"
    # Create a new  readall user 
    custom_readall_user:
      backend_roles:
      - "readall"
      description: "Custom readall user"
```

Here's how `.spec.internalUsers` can be configured for `xpack` auth plugins.

```yaml
spec:
  internalUsers:
    apm_system:
      backendRoles:
      - apm_system
      secretName: es-cluster-apm-system-cred
    beats_system:
      backendRoles:
      - beats_system
      secretName: es-cluster-beats-system-cred
    elastic:
      backendRoles:
      - superuser
      secretName: es-cluster-elastic-cred
    kibana_system:
      backendRoles:
      - kibana_system
      secretName: es-cluster-kibana-system-cred
    logstash_system:
      backendRoles:
      - logstash_system
      secretName: es-cluster-logstash-system-cred
    remote_monitoring_user:
      backendRoles:
      - remote_monitoring_collector
      - remote_monitoring_agent
      secretName: es-cluster-remote-monitoring-user-cred
```
**ElasticStack:**

Default Users: [Official Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html)

- `elastic` - Has direct read-only access to restricted indices, such as .security. This user also has the ability to manage security and create roles with unlimited privileges
- `kibana_system` -  The user Kibana uses to connect and communicate with Elasticsearch.
- `logstash_system` - The user Logstash uses when storing monitoring information in Elasticsearch.
- `beats_system` - The user the Beats use when storing monitoring information in Elasticsearch.
- `apm_system` - The user the APM server uses when storing monitoring information in Elasticsearch.
- `remote_monitoring_user` - The user Metricbeat uses when collecting and storing monitoring information in Elasticsearch. It has the remote_monitoring_agent and remote_monitoring_collector built-in roles.

**SearchGuard:**

Default Users: [Official Docs](https://docs.search-guard.com/latest/demo-users-roles)

- `admin` - Full access to the cluster and all indices.
- `kibanaserver` -  Has all permissions on the `.kibana` index.
- `kibanaro` - Has `SGS_READ` access to all indices and all permissions on the `.kibana` index.
- `logstash` - Has `SGS_CRUD` and `SGS_CREATE_INDEX` permissions on all logstash and beats indices.
- `readall` - Has read access to all indices.
- `snapshotrestore` - Has permissions to perform snapshot and restore operations.

**OpenDistro:** 

Default Users: [Official Docs](https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/users-roles/)

- `admin` - Grants full access to the cluster: all cluster-wide operations, write to all indices, write to all tenants.
- `kibanaserver` - Has all permissions on the `.kibana` index
- `kibanaro` - Grants permissions to use Kibana: cluster-wide searches, index monitoring, and write to various Kibana indices.
- `logstash` - Grants permissions for Logstash to interact with the cluster: cluster-wide searches, cluster monitoring, and write to the various Logstash indices.
- `readall` - Grants permissions for cluster-wide searches like msearch and search permissions for all indices.
- `snapshotrestore` - Grants permissions to manage snapshot repositories, take snapshots, and restore snapshots.

### spec.rolesMapping

`spec.rolesMapping` provides an alternative way to  map backend roles, hosts and users to roles without using the `roles_mapping.yml` file. Only works with `SearchGurad` and `OpenDistro` security plugins. This field expects the input format to be in the `map[rolename]RoleSpec` format.

The `RoleSpec` contains the following fields:

- `reserved` ( `bool` | `false` ) - specifies the reserved status. The resources that have this set to `true`, cannot be changed using the REST API or Kibana.
- `hidden` ( `bool` | `false` ) - specifies the hidden status. The resources that have this field set to `true` are not returned by the REST API and not visible in Kibana.
- `backendRoles` ( `[]string` | `nil` )- specifies a list of backend roles assigned to this role. The backend roles can come from the internal user database, LDAP groups, JSON web token-claims or SAML assertions.
- `hosts` ( `[]string` | `nil` ) - specifies a list of hosts assigned to this role.
- `users` ( `[]string` | `nil` ) - specifies a list of users assigned to this role.
- `

```yaml
spec:
  rolesMapping:
    # create role mapping for the custom readall user
    readall:
      users:
      - custom_readall_user
```

For the default roles visit the [SearchGurad docs](https://docs.search-guard.com/latest/roles-permissions), [OpenDistro docs](https://opendistro.github.io/for-elasticsearch-docs/docs/security/access-control/users-roles/#create-roles).

### spec.topology

`spec.topology` is an `optional` field that provides a way to configure different types of nodes for the Elasticsearch cluster. This field enables you to specify how many nodes you want to act as `master`, `data`, `ingest` or other node roles for Elasticsearch. You can also specify how much storage and resources to allocate for each type of node independently.

Currently supported node types are -
- **data**: Data nodes hold the shards that contain the documents you have indexed. Data nodes handle data related operations like CRUD, search, and aggregations
- **ingest**: Ingest nodes can execute pre-processing pipelines, composed of one or more ingest processors
- **master**: The master node is responsible for lightweight cluster-wide actions such as creating or deleting an index, tracking which nodes are part of the cluster, and deciding which shards to allocate to which nodes. It is important for cluster health to have a stable master node.
- **dataHot**: Hot data nodes are part of the hot tier. The hot tier is the Elasticsearch entry point for time series data and holds your most-recent, most-frequently-searched time series data.
- **dataWarm**: Warm data nodes are part of the warm tier. Time series data can move to the warm tier once it is being queried less frequently than the recently-indexed data in the hot tier.
- **dataCold**: Cold data nodes are part of the cold tier. When you no longer need to search time series data regularly, it can move from the warm tier to the cold tier.
- **dataFrozen**: Frozen data nodes are part of the frozen tier. Once data is no longer being queried, or being queried rarely, it may move from the cold tier to the frozen tier where it stays for the rest of its life.
- **dataContent**: Content data nodes are part of the content tier. Data stored in the content tier is generally a collection of items such as a product catalog or article archive. Unlike time series data, the value of the content remains relatively constant over time, so it doesn’t make sense to move it to a tier with different performance characteristics as it ages.
- **ml**: Machine learning nodes run jobs and handle machine learning API requests.
- **transform**: Transform nodes run transforms and handle transform API requests.
- **coordinating**: The coordinating node forwards the request to the data nodes which hold the data.

```yaml
  topology:
    data:
      maxUnavailable: 1
      replicas: 3
      podTemplate: 
        spec:
          containers:
            - name: "elasticsearch"
              resources:
                requests:
                  cpu: "500m"
                limits:
                  cpu: "600m"
                  memory: "1.5Gi"
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
      podTemplate:
        spec:
          containers:
            - name: "elasticsearch"
              resources:
                requests:
                  cpu: "500m"
                limits:
                  cpu: "600m"
                  memory: "1.5Gi"
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
      podTemplate:
        spec:
          containers:
            - name: "elasticsearch"
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
  - `suffix` (`: "master"`) - is an `optional` field that is added as the suffix of the master PetSet name. Defaults to `master`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `master` nodes.
  - `resources` (`: "cpu: 500m, memory: 1Gi" `) - is an `optional` field that specifies how much computational resources to request or to limit for each of the `master` nodes.
  - `maxUnavailable` is an `optional` field that specifies the exact number of master nodes (ie. pods) that can be safely evicted before the pod disruption budget (PDB) kicks in. KubeDB uses Pod Disruption Budget to ensure that desired number of replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that no data loss occurs.

- `topology.data`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `data` nodes. Defaults to `1`.
  - `suffix` (`: "data"`) - is an `optional` field that is added as the suffix of the data PetSet name. Defaults to `data`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `data` nodes.
  - `resources` (` cpu: 500m, memory: 1Gi `) - is an `optional` field that specifies which amount of computational resources to request or to limit for each of the `data` nodes.
  - `maxUnavailable` is an `optional` field that specifies the exact number of data nodes (ie. pods) that can be safely evicted before the pod disruption budget (PDB) kicks in. KubeDB uses Pod Disruption Budget to ensure that desired number of replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that no data loss occurs.

- `topology.ingest`:
  - `replicas` (`: "1"`) - is an `optional` field to specify the number of nodes (ie. pods ) that act as the `ingest` nodes. Defaults to `1`.
  - `suffix` (`: "ingest"`) - is an `optional` field that is added as the suffix of the data PetSet name. Defaults to `ingest`.
  - `storage` is a `required` field that specifies how much storage to claim for each of the `ingest` nodes.
  - `resources` (` cpu: 500m, memory: 1Gi `) - is an `optional` field that specifies which amount of computational resources to request or to limit for each of the `data` nodes.
  - `maxUnavailable` is an `optional` field that specifies the exact number of ingest nodes (ie. pods) that can be safely evicted before the pod disruption budget (PDB) kicks in. KubeDB uses Pod Disruption Budget to ensure that desired number of replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that no data loss is occurs.
  
> Note: Any two types of nodes can't have the same `suffix`.

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

`spec.maxUnavailable` is an `optional` field that is used to specify the exact number of cluster replicas that can be safely evicted before the pod disruption budget kicks in to prevent unwanted data loss.

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

> Note: The `transport` layer of an Elasticsearch cluster is always secured with certificates. If you want to disable it, you need to disable the security plugin by setting the `spec.disableSecurity` to `true`.

### spec.tls

`spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). Currently, the operator only supports the `PKCS#8` encoded certificates.

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: es-issuer
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
```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Elasticsearch. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA. 
  - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
  - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.

- `tls.certificates` - is an `optional` field that specifies a list of certificate configurations used to configure the  certificates. It has the following fields:
  - `alias` - represents the identifier of the certificate. It has the following possible value:
    - `transport` - is used for the transport layer certificate configuration.
    - `http` - is used for the HTTP layer certificate configuration.
    - `admin` - is used for the admin certificate configuration. Available for the `SearchGuard` and the `OpenDistro` auth-plugins.
    - `metrics-exporter` - is used for the metrics-exporter sidecar certificate configuration.
  
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

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Elasticsearch object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Elasticsearch object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `elastic` superuser.

Example:

```bash
$ kubectl create secret generic elastic-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "elastic-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: elastic-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.storageType

`spec.storageType` is an `optional` field that specifies the type of storage to use for the database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Elasticsearch database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. In this case, you don't have to specify `spec.storage` field.

```yaml
spec:
  storageType: Durable
```

### spec.storage

If the `spec.storageType`  is not set to `Ephemeral` and if the `spec.topology` field also is not set then `spec.storage` field is `required`. This field specifies the StorageClass of the PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

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

`spec.init` is an `optional` section that can be used to initialize a newly created Elasticsearch cluster from prior snapshots, taken by [Stash](/docs/guides/elasticsearch/backup/overview/index.md).

```yaml
spec:
  init:
    waitForInitialRestore: true
```

When the `waitForInitialRestore` is set to true, the Elasticsearch instance will be stack in the `Provisioning` state until the initial backup is completed. On completion of the very first restore operation, the Elasticsearch instance will go to the `Ready` state.

For detailed tutorial on how to initialize Elasticsearch from Stash backup, please visit [here](/docs/guides/elasticsearch/backup/overview/index.md).

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
- `sg_roles_mapping.yml` - map backend roles, hosts, and users to roles.
- `sg_internal_users.yml` - stores users, and hashed passwords in the internal user database.
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

- `YML`: The default configuration file pre-stored at config directories is overwritten by the operator-generated configuration file (if any). Then the resultant configuration file is overwritten by the user-provided custom configuration file (if any). The [yq](https://github.com/mikefarah/yq) tool is used to merge two YAML files.

  ```bash
  $ yq merge -i --overwrite file1.yml file2.yml
  ```

- `Non-YML`: The default configuration file is replaced by the operator-generated one (if any). Then the resultant configuration file is replaced by the user-provided custom configuration file (if any).

  ```bash
  $ cp -f file2 file1
  ```

**How to provide node-role specific configurations?**

If an Elasticsearch cluster is running in the topology mode (ie. `spec.topology` is set), a user may want to provide node-role specific configurations, say configurations that will only be merged to `master` nodes. To achieve this, users need to add the node role as a prefix to the file name.

- Format: `<node-role>-<file-name>.extension`
- Samples:
  - `data-elasticsearch.yml`: Only applied to `data` nodes.
  - `master-jvm.options`: Only applied to `master` nodes.
  - `ingest-log4j2.properties`: Only applied to `ingest` nodes.

**How to provide additional files that is referenced from the configurations?**

All these files provided via `configSecret` is stored in each Elasticsearch node (i.e. pod) at `ES_CONFIG_DIR/custom_config/` ( i.e. `/usr/share/elasticsearch/config/custom_config/`) directory. So, user can use this path while configuring the Elasticsearch.

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

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Elasticsearch database.

KubeDB accept following fields to set in `spec.podTemplate:`

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

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go#L26C1-L279C1).

Uses of some fields of `spec.podTemplate` are described below,



#### spec.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.podTemplate.spec.volumes

The `spec.podTemplate.spec.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the podPlacementPolicy. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.


#### spec.podTemplate.spec.imagePullSecrets

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image when you are using a private docker registry. For more details on how to use private docker registry, please visit [here](/docs/guides/elasticsearch/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an `optional` field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine-tune role-based access control.

If this field is left empty, the KubeDB operator will create a service account name matching the Elasticsearch instance name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/elasticsearch/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

```yaml
spec:
  podTemplate:
    spec:
      serviceAccountName: es
```

#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.env` is an `optional` field that specifies the environment variables to pass to the Elasticsearch Containers.

You are not allowed to pass the following `env`:
- `node.name`
- `node.ingest`
- `node.master`
- `node.data`


```ini
Error from server (Forbidden): error when creating "./elasticsearch.yaml": admission webhook "elasticsearch.validators.kubedb.com" denied the request: environment variable node.name is forbidden to use in Elasticsearch spec
```

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an `optional` field. then it can be used to request or limit computational resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

```yaml
spec:
  podTemplate:
    spec:
      containers:
        - name: "elasticsearch"
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
```



### spec.serviceTemplates

`spec.serviceTemplates` is an `optional` field that contains a list of the serviceTemplate. The templates are identified by the `alias`. For Elasticsearch, the configurable services' `alias` are `primary` and `stats`.

You can also provide template for the services created by KubeDB operator for Elasticsearch database through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
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

```yaml
spec:
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passTo: service
    spec:
      type: NodePort
  - alias: stats
    # stats service configurations
```

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.18.9/api/v1/types.go#L192) to understand these fields in detail.

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Elasticsearch` CRD or which resources KubeDB should keep or delete when you delete `Elasticsearch` CRD. The KubeDB operator provides the following termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes v1.9+ to provide safety from accidental deletion of the database. If admission webhook is enabled, KubeDB prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Elasticsearch CRD for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete TLS Credential Secrets    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 5. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 6. Delete User Credential Secrets   |    &#10007;    | &#10007; | &#10007; | &#10003; |


If the `spec.deletionPolicy` is not specified, the KubeDB operator defaults it to `Delete`.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run an Elasticsearch database [here](/docs/guides/elasticsearch/README.md).
- Learn how to use ElasticsearchOpsRequest [here](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
