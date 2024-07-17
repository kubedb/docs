---
title: MongoDB CRD
menu:
  docs_{{ .version }}:
    identifier: mg-mongodb-concepts
    name: MongoDB
    parent: mg-concepts-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB

## What is MongoDB

`MongoDB` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [MongoDB](https://www.mongodb.com/) in a Kubernetes native way. You only need to describe the desired database configuration in a MongoDB object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## MongoDB Spec

As with all other Kubernetes objects, a MongoDB needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example MongoDB object.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mgo1
  namespace: demo
spec:
  autoOps:
    disabled: true
  version: "4.4.26"
  replicas: 3
  authSecret:
    name: mgo1-auth
    externallyManaged: false
  replicaSet:
    name: rs0
  shardTopology:
    configServer:
      podTemplate: {}
      replicas: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      podTemplate: {}
      replicas: 2
    shard:
      podTemplate: {}
      replicas: 3
      shards: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  sslMode: requireSSL
  tls:
    issuerRef:
      name: mongo-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - kubedb
        emailAddresses:
          - abc@appscode.com
      - alias: server
        subject:
          organizations:
            - kubedb
        emailAddresses:
          - abc@appscode.com
  clusterAuthMode: x509
  storageType: "Durable"
  storageEngine: wiredTiger
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  ephemeralStorage:
    medium: "Memory"
    sizeLimit: 500Mi
  init:
    script:
      configMap:
        name: mg-init-script
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          app: kubedb
        interval: 10s
  configSecret:
    name: mg-custom-config
  podTemplate:
    metadata:      
      annotations:
        passMe: ToDatabasePod
      labels:
        thisLabel: willGoToPod
    controller:
      annotations:
        passMe: ToPetSet
      labels:
        thisLabel: willGoToSts
    spec:
      serviceAccountName: my-service-account
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
        - name: myregistrykey
      containers:
      - name: mongo
        args:
          - --maxConns=100
        env:
          - name: MONGO_INITDB_DATABASE
            value: myDB
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
      - name: replication-mode-detector
        resources:
          requests:
            cpu: "300m"
            memory: 500Mi
        securityContext:
            runAsUser: 1001
  serviceTemplates:
  - alias: primary
    spec:
      type: NodePort
      ports:
        - name: primary
          port: 27017
          nodePort: 300006
  deletionPolicy: Halt
  halted: false
  arbiter:
    podTemplate:
      spec:
        resources:
          requests:
            cpu: "200m"
            memory: "200Mi"
    configSecret:
      name: another-config
  allowedSchemas:
    namespaces:
      from: Selector
      selector:
        matchExpressions:
          - {key: kubernetes.io/metadata.name, operator: In, values: [dev]}
    selector:
      matchLabels:
        "schema.kubedb.com": "mongo"
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 2
    disableWriteCheck: false
```

### spec.autoOps
AutoOps is an optional field to control the generation of versionUpdate & TLS-related recommendations.

### spec.version

`spec.version` is a required field specifying the name of the [MongoDBVersion](/docs/guides/mongodb/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `MongoDBVersion` resources,

- `3.4.17-v1`, `3.4.22-v1`
- `3.6.13-v1`, `4.4.26`, 
- `4.0.3-v1`, `4.4.26`, `4.0.11-v1`,
- `4.1.4-v1`, `4.1.7-v3`, `4.4.26`
- `4.4.26`, `4.4.26`
- `5.0.2`, `5.0.3`
- `percona-3.6.18`
- `percona-4.0.10`, `percona-4.2.7`, `percona-4.4.10`

### spec.replicas

`spec.replicas` the number of members(primary & secondary) in mongodb replicaset.

If `spec.shardTopology` is set, then `spec.replicas` needs to be empty. Instead use `spec.shardTopology.<shard/configServer/mongos>.replicas`

If both `spec.replicaset` and `spec.shardTopology` is not set, then `spec.replicas` can be value `1`.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `mongodb` superuser. If not set, KubeDB operator creates a new Secret `{mongodb-object-name}-auth` for storing the password for `mongodb` superuser for each MongoDB object. 

We can use this field in 3 mode. 
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the MongoDB object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the MongoDB object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `mongodb` superuser.

Example:

```bash
$ kubectl create secret generic mgo1-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "mgo1-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: mgo1-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.replicaSet

`spec.replicaSet` represents the configuration for replicaset. When `spec.replicaSet` is set, KubeDB will deploy a mongodb replicaset where number of replicaset member is spec.replicas.

- `name` denotes the name of mongodb replicaset.
NB. If `spec.shardTopology` is set, then `spec.replicaset` needs to be empty.

### spec.keyFileSecret
`keyFileSecret.name` denotes the name of the secret that contains the `key.txt`, which provides the security between replicaset members using internal authentication. See [Keyfile Authentication](https://docs.mongodb.com/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/) for more information.
It will make impact only if the ClusterAuthMode is `keyFile` or `sendKeyFile`.

### spec.shardTopology

`spec.shardTopology` represents the topology configuration for sharding.

Available configurable fields:

- shard
- configServer
- mongos

When `spec.shardTopology` is set, the following fields needs to be empty, otherwise validating webhook will throw error.

- `spec.replicas`
- `spec.podTemplate`
- `spec.configSecret`
- `spec.storage`
- `spec.ephemeralStorage`

KubeDB uses `PodDisruptionBudget` to ensure that majority of the replicas of these shard components are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum and data integrity is maintained.

#### spec.shardTopology.shard

`shard` represents configuration for Shard component of mongodb.

Available configurable fields:

- `shards` represents number of shards for a mongodb deployment. Each shard is deployed as a [replicaset](/docs/guides/mongodb/clustering/replication_concept.md).
- `replicas` represents number of replicas of each shard replicaset.
- `prefix` represents the prefix of each shard node.
- `configSecret` is an optional field to provide custom configuration file for shards (i.e. mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used. See below to know about [spec.configSecret](/docs/guides/mongodb/concepts/mongodb.md#specconfigsecret) in details.
- `podTemplate` is an optional configuration for pods. See below to know about [spec.podTemplate](/docs/guides/mongodb/concepts/mongodb.md#specpodtemplate) in details.
- `storage` to specify pvc spec for each node of sharding. You can specify any StorageClass available in your cluster with appropriate resource requests. See below to know about [spec.storage](/docs/guides/mongodb/concepts/mongodb.md#specstorage) in details.
- `ephemeralStorage` to specify the configuration of ephemeral storage type, If you want to use volatile temporary storage attached to your instances which is only present during the running lifetime of the instance.

#### spec.shardTopology.configServer

`configServer` represents configuration for ConfigServer component of mongodb.

Available configurable fields:

- `replicas` represents number of replicas for configServer replicaset. Here, configServer is deployed as a replicaset of mongodb.
- `prefix` represents the prefix of configServer nodes.
- `configSecret` is an optional field to provide custom configuration file for config server (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used. See below to know about [spec.configSecret](/docs/guides/mongodb/concepts/mongodb.md#specconfigsecret) in details.
- `podTemplate` is an optional configuration for pods. See below to know about [spec.podTemplate](/docs/guides/mongodb/concepts/mongodb.md#specpodtemplate) in details.
- `storage` to specify pvc spec for each node of configServer. You can specify any StorageClass available in your cluster with appropriate resource requests. See below to know about [spec.storage](/docs/guides/mongodb/concepts/mongodb.md#specstorage) in details.
- `ephemeralStorage` to specify the configuration of ephemeral storage type, If you want to use volatile temporary storage attached to your instances which is only present during the running lifetime of the instance.

#### spec.shardTopology.mongos

`mongos` represents configuration for Mongos component of mongodb.

Available configurable fields:

- `replicas` represents number of replicas of `Mongos` instance. Here, Mongos is deployed as stateless (deployment) instance.
- `prefix` represents the prefix of mongos nodes.
- `configSecret` is an optional field to provide custom configuration file for mongos (i.e. mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used. See below to know about [spec.configSecret](/docs/guides/mongodb/concepts/mongodb.md#specconfigsecret) in details.
- `podTemplate` is an optional configuration for pods. See below to know about [spec.podTemplate](/docs/guides/mongodb/concepts/mongodb.md#specpodtemplate) in details.

### spec.sslMode

Enables TLS/SSL or mixed TLS/SSL used for all network connections. The value of [`sslMode`](https://docs.mongodb.com/manual/reference/program/mongod/#cmdoption-mongod-sslmode) field can be one of the following:

|    Value     | Description                                                                                                                    |
| :----------: | :----------------------------------------------------------------------------------------------------------------------------- |
|  `disabled`  | The server does not use TLS/SSL.                                                                                               |
|  `allowSSL`  | Connections between servers do not use TLS/SSL. For incoming connections, the server accepts both TLS/SSL and non-TLS/non-SSL. |
| `preferSSL`  | Connections between servers use TLS/SSL. For incoming connections, the server accepts both TLS/SSL and non-TLS/non-SSL.        |
| `requireSSL` | The server uses and accepts only TLS/SSL encrypted connections.                                                                |

### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the MongoDB. KubeDB uses [cert-manager](https://cert-manager.io/) v1 api to provision and manage TLS certificates.

The following fields are configurable in the `spec.tls` section:

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.

  - `apiGroup` is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` is the type of resource that is being referenced. KubeDB supports both `Issuer` and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.

- `certificates` (optional) are a list of certificates used to configure the server and/or client certificate. It has the following fields:
  - `alias` represents the identifier of the certificate. It has the following possible value:
    - `server` is used for server certificate identification.
    - `client` is used for client certificate identification.
    - `metrics-exporter` is used for metrics exporter certificate identification.
  - `secretName` (optional) specifies the k8s secret name that holds the certificates.
    > This field is optional. If the user does not specify this field, the default secret name will be created in the following format: `<database-name>-<cert-alias>-cert`.
  
  - `subject` (optional) specifies an `X.509` distinguished name. It has the following possible field,
    - `organizations` (optional) are the list of different organization names to be used on the Certificate.
    - `organizationalUnits` (optional) are the list of different organization unit name to be used on the Certificate.
    - `countries` (optional) are the list of country names to be used on the Certificate.
    - `localities` (optional) are the list of locality names to be used on the Certificate.
    - `provinces` (optional) are the list of province names to be used on the Certificate.
    - `streetAddresses` (optional) are the list of a street address to be used on the Certificate.
    - `postalCodes` (optional) are the list of postal code to be used on the Certificate.
    - `serialNumber` (optional) is a serial number to be used on the Certificate.
      You can find more details from [Here](https://golang.org/pkg/crypto/x509/pkix/#Name)
  - `duration` (optional) is the period during which the certificate is valid.
  - `renewBefore` (optional) is a specifiable time before expiration duration.
  - `dnsNames` (optional) is a list of subject alt names to be used in the Certificate.
  - `ipAddresses` (optional) is a list of IP addresses to be used in the Certificate.
  - `uris` (optional) is a list of URI Subject Alternative Names to be set in the Certificate.
  - `emailAddresses` (optional) is a list of email Subject Alternative Names to be set in the Certificate.
  - `privateKey` (optional) specifies options to control private keys used for the Certificate.
    - `encoding` (optional) is the private key cryptography standards (PKCS) encoding for this certificate's private key to be encoded in. If provided, allowed values are "pkcs1" and "pkcs8" standing for PKCS#1 and PKCS#8, respectively. It defaults to PKCS#1 if not specified.

### spec.clusterAuthMode

The authentication mode used for cluster authentication. This option can have one of the following values:

|     Value     | Description                                                                                                                      |
| :-----------: | :------------------------------------------------------------------------------------------------------------------------------- |
|   `keyFile`   | Use a keyfile for authentication. Accept only keyfiles.                                                                          |
| `sendKeyFile` | For rolling update purposes. Send a keyfile for authentication but can accept both keyfiles and x.509 certificates.             |
|  `sendX509`   | For rolling update purposes. Send the x.509 certificate for authentication but can accept both keyfiles and x.509 certificates. |
|    `x509`     | Recommended. Send the x.509 certificate for authentication and accept only x.509 certificates.                                   |

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MongoDB database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. 
In this case, you don't have to specify `spec.storage` field. Specify `spec.ephemeralStorage` spec instead.

### spec.storageEngine

`spec.storageEngine` is an optional field that specifies the type of storage engine is going to be used by mongodb. There are two types of storage engine, `wiredTiger` and `inMemory`. Default value of storage engine is `wiredTiger`. `inMemory` storage engine is only supported by the percona variant of mongodb, i.e. the version that has the `percona-` prefix in the mongodb-version name.

### spec.storage

Since 0.9.0-rc.0, If you set `spec.storageType:` to `Durable`, then `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

NB. If `spec.shardTopology` is set, then `spec.storage` needs to be empty. Instead use `spec.shardTopology.<shard/configServer>.storage`

### spec.ephemeralStorage
Use this field to specify the configuration of ephemeral storage type, If you want to use volatile temporary storage attached to your instances which is only present during the running lifetime of the instance.
- `spec.ephemeralStorage.medium` refers to the name of the storage medium.
- `spec.ephemeralStorage.sizeLimit` to specify the sizeLimit of the emptyDir volume.

For more details of these two fields, see [EmptyDir struct](https://github.com/kubernetes/api/blob/ed22bb34e3bbae9e2fafba51d66ee3f68ee304b2/core/v1/types.go#L700-L715)

### spec.init

`spec.init` is an optional section that can be used to initialize a newly created MongoDB database. MongoDB databases can be initialized by Script.

`Initialize from Snapshot` is still not supported.

#### Initialize via Script

To initialize a MongoDB database using a script (shell script, js script), set the `spec.init.script` section when creating a MongoDB object. It will execute files alphabetically with extensions `.sh` and `.js` that are found in the repository. script must have the following information:

- [VolumeSource](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes): Where your script is loaded from.

Below is an example showing how a script from a configMap can be used to initialize a MongoDB database.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mgo1
  namespace: demo
spec:
  version: 4.4.26
  init:
    script:
      configMap:
        name: mongodb-init-script
```

In the above example, KubeDB operator will launch a Job to execute all js script of `mongodb-init-script` in alphabetical order once PetSet pods are running. For more details tutorial on how to initialize from script, please visit [here](/docs/guides/mongodb/initialization/using-script.md).

These are the fields of `spec.init` which you can make use of :
- `spec.init.initialized` indicating that this database has been initialized or not. `false` by default.
- `spec.init.script.scriptPath` to specify where all the init scripts should be mounted.
- `spec.init.script.<volumeSource>` as described in the above example. To see all the volumeSource options go to [VolumeSource](https://github.com/kubernetes/api/blob/ed22bb34e3bbae9e2fafba51d66ee3f68ee304b2/core/v1/types.go#L49).
- `spec.init.waitForInitialRestore` to tell the operator if it should wait for the initial restore process or not.

### spec.monitor

MongoDB managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor MongoDB with builtin Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md)
- [Monitor MongoDB with Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for MongoDB. You can provide the custom configuration in a secret, then you can specify the secret name `spec.configSecret.name`.

> Please note that, the secret key needs to be `mongod.conf`.

To learn more about how to use a custom configuration file see [here](/docs/guides/mongodb/configuration/using-config-file.md).

NB. If `spec.shardTopology` is set, then `spec.configSecret` needs to be empty. Instead use `spec.shardTopology.<shard/configServer/mongos>.configSecret`

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for MongoDB database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
  - labels (pod's labels)
- controller:
  - annotations (petset's annotation)
  - labels (petset's labels)
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

You can checkout the full list [here](https://github.com/kmodules/offshoot-api/blob/ea366935d5bad69d7643906c7556923271592513/api/v1/types.go#L42-L259). Uses of some field of `spec.podTemplate` is described below,

NB. If `spec.shardTopology` is set, then `spec.podTemplate` needs to be empty. Instead use `spec.shardTopology.<shard/configServer/mongos>.podTemplate`

#### spec.podTemplate.spec.args

`spec.podTemplate.spec.args` is an optional field. This can be used to provide additional arguments to database installation. To learn about available args of `mongod`, visit [here](https://docs.mongodb.com/manual/reference/program/mongod/).

#### spec.podTemplate.spec.env

`spec.podTemplate.spec.env` is an optional field that specifies the environment variables to pass to the MongoDB docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/r/_/mongo/).

Note that, KubeDB does not allow `MONGO_INITDB_ROOT_USERNAME` and `MONGO_INITDB_ROOT_PASSWORD` environment variables to set in `spec.podTemplate.spec.env`. If you want to use custom superuser and password, please use `spec.authSecret` instead described earlier.

If you try to set `MONGO_INITDB_ROOT_USERNAME` or `MONGO_INITDB_ROOT_PASSWORD` environment variable in MongoDB crd, Kubedb operator will reject the request with following error,

```ini
Error from server (Forbidden): error when creating "./mongodb.yaml": admission webhook "mongodb.validators.kubedb.com" denied the request: environment variable MONGO_INITDB_ROOT_USERNAME is forbidden to use in MongoDB spec
```

Also, note that KubeDB does not allow updating the environment variables as updating them does not have any effect once the database is created. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./mongodb.yaml": admission webhook "mongodb.validators.kubedb.com" denied the request: precondition failed for:
...At least one of the following was changed:
    apiVersion
    kind
    name
    namespace
    spec.ReplicaSet
    spec.authSecret
    spec.init
    spec.storageType
    spec.storage
    spec.podTemplate.spec.nodeSelector
    spec.podTemplate.spec.env
```

#### spec.podTemplate.spec.imagePullSecret

`KubeDB` provides the flexibility of deploying MongoDB database from a private Docker registry. `spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry. To learn how to deploy MongoDB from a private registry, please visit [here](/docs/guides/mongodb/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account name matching MongoDB crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/mongodb/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide template for the services created by KubeDB operator for MongoDB database through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
  - `primary` is used for the primary service identification.
  - `standby` is used for the secondary service identification.
  - `stats` is used for the exporter service identification.
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

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MongoDB` crd or which resources KubeDB should keep or delete when you delete `MongoDB` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete MongoDB crd for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshot data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.

### spec.halted
Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.

### spec.arbiter
If `spec.arbiter` is not null, there will be one arbiter pod on each of the replicaset structure, including shards. It has two fields. 
- `spec.arbiter.podTemplate` defines the arbiter-pod's template. See [spec.podTemplate](/docs/guides/mongodb/configuration/using-config-file.md) part for more details of this.
- `spec.arbiter.configSecret` is an optional field that allows users to provide custom configurations for MongoDB arbiter. You just need to refer the configuration secret in `spec.arbiter.configSecret.name` field.
> Please note that, the secret key needs to be `mongod.conf`.

N.B. If `spec.replicaset` & `spec.shardTopology` both is empty, `spec.arbiter` has to be empty too.

### spec.allowedSchemas
It defines which consumers may refer to a database instance. We implemented double-optIn feature between database instance and schema-manager using this field.
- `spec.allowedSchemas.namespace.from` indicates how you want to filter the namespaces, from which a schema-manager will be able to communicate with this db instance.
Possible values are : i) `All` to allow all namespaces, ii) `Same` to allow only if schema-manager & MongoDB is deployed in same namespace & iii) `Selector` to select some namespaces through labels.
- `spec.allowedSchemas.namespace.selector`. You need to set this field only if `spec.allowedSchemas.namespace.from` is set to `selector`. Here you will give the labels of the namespaces to allow.
- `spec.allowedSchemas.selctor` denotes the labels of the schema-manager instances, which you want to give allowance to use this database. 

### spec.coordinator
We use a dedicated container, named `replication-mode-detector`, to continuously select primary pod and add label as primary. By specifying `spec.coordinator.resources` & `spec.coordinator.securityContext`, you can set the resources and securityContext of that mode-detector container.


## spec.healthChecker
It defines the attributes for the health checker. 
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run a MongoDB database [here](/docs/guides/mongodb/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
