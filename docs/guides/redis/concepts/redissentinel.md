---
title: RedisSentinel CRD
menu:
  docs_{{ .version }}:
    identifier: rd-redissentinel-concepts
    name: RedisSentinel
    parent: rd-concepts-redis
    weight: 12
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RedisSentinel

## What is RedisSentinel

`RedisSentinel` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Redis](https://redis.io/) in a Kubernetes native way. You only need to describe the desired database configuration in a Redis Sentinel object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## RedisSentinel Spec

As with all other Kubernetes objects, a Redis needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Redis object.

```yaml
apiVersion: kubedb.com/v1
kind: RedisSentinel
metadata:
  name: sentinel1
  namespace: demo
spec:
  autoOps:
    disabled: true
  version: 6.2.14
  replicas: 3
  disableAuth: false
  authSecret:
    name: sentinel1-auth
    externallyManaged: false
  tls:
    issuerRef:
      name: redis-ca-issuer
      kind: ClusterIssuer
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
    metadata:
      annotations:
        passMe: ToDatabasePod
    controller:
      annotations:
        passMe: ToPetSet
    spec:
      serviceAccountName: my-service-account
      imagePullSecrets:
        - name: regcred
      containers:
      - name: redissentinel
        args:
          - "--loglevel verbose"
        env:
          - name: ENV_VARIABLE
            value: "value"
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
  halted: false
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 2
    disableWriteCheck: false
```

### spec.autoOps
AutoOps is an optional field to control the generation of version update & TLS-related recommendations.

### spec.version

`spec.version` is a required field specifying the name of the [RedisVersion](/docs/guides/redis/concepts/catalog.md) crd where the docker images are specified. RedisSentinel is supported in following Redis Versions.

- `6.2.14`, `6.2.14` `6.2.14`
- `7.0.4`, `7.0.14`, `7.0.6`

### spec.disableAuth

`spec.disableAuth` is an optional field that decides whether RedisSentinel instance will be secured by auth or no.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `redis` superuser. If not set, KubeDB operator creates a new Secret `{redissentinel-object-name}-auth` for storing the password for `redis` superuser.

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the RedisSentinel object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the RedisSentinel object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `redis` superuser.

Example:

```bash
$ kubectl create secret generic sentinel1-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "sentinel1-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: sentinel1-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).


### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the RedisSentinel. KubeDB uses [cert-manager](https://cert-manager.io/) v1 api to provision and manage TLS certificates.

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

The Redis object we construct will be watched over by the Redis Sentinel object, therefore in order for them to connect in TLS enabled mode, 
both objects must have the same issuer. On the other side, the Redis object must likewise be TLS off if the RedisSentinel object is.
Set `spec.tls.issuerRef.kind` to `ClusterIssuer` if you want your RedisSentinel object and Redis object to be in different namespaces.
Both "Issuer" and "ClusterIssuer" can be used if both instances are in the same namespace.

### spec.storage

Since 0.10.0-rc.0, If you set `spec.storageType:` to `Durable`, then  `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.monitor

RedisSentinel managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor Redis with builtin Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md)
- [Monitor Redis with Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md)


### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for RedisSentinel server.

KubeDB accept following fields to set in `spec.podTemplate:`

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

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the podPlacementPolicy. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.




#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the Redis containers.

Note that, KubeDB does not allow to update the environment variables. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./redis.yaml": admission webhook "redis.validators.kubedb.com" denied the request: precondition failed for:
...
At least one of the following was changed:
apiVersion
kind
name
namespace
spec.storage
spec.podTemplate.spec.nodeSelector
spec.podTemplate.spec.containers[].env
```

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

#### spec.podTemplate.spec.imagePullSecret

`KubeDB` provides the flexibility of deploying Redis server from a private Docker registry. To learn how to deploy Redis from a private registry, please visit [here](/docs/guides/redis/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

  `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

  If this field is left empty, the KubeDB operator will create a service account name matching RedisSentinel crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

  If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

  If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/redis/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

### spec.serviceTemplates

You can also provide a template for the services created by KubeDB operator for Redis server through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:

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

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `RedisSentinel` crd or which resources KubeDB should keep or delete when you delete `RedisSentinel` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Redis crd for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.

### spec.halted
Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Sample Redis instace
A yaml for a sample Redis instance that can be monitored by this RedisSentinel instance is given below
```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis1
  namespace: demo
spec:
  version: 6.2.14
  replicas: 3
  sentinelRef:
    name: sentinel1
    namespace: demo
  mode: Sentinel
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: redis-ca-issuer
      kind: ClusterIssuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

## Next Steps

- Learn how to use KubeDB to run a Redis server [here](/docs/guides/redis/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
