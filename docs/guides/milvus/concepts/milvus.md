---
title: Milvus CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-concepts-milvus
    name: Milvus
    parent: milvus-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus

## What is Milvus

`Milvus` is a KubeDB `CustomResourceDefinition` used to deploy and manage Milvus vector databases. You only need to describe the desired database configuration in a `Milvus`object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Milvus Spec

As with all other Kubernetes objects, a Milvus needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example of Milvus object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: kubedb
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  metaStorage:
    externallyManaged: true
    endpoints:
      - http://etcdcluster-sample-0.etcdcluster-sample.default.svc.cluster.local:2379
      - http://etcdcluster-sample-1.etcdcluster-sample.default.svc.cluster.local:2379
      - http://etcdcluster-sample-2.etcdcluster-sample.default.svc.cluster.local:2379
  disableSecurity: false
  authSecret:
    name: "milvus-auth"
    externallyManaged: true
  configuration:
    secretName: my-release-user-config
    inline:
      milvus.yaml: |
        log:
          level: info
          file:
            maxAge: 30
  topology:
    mode: Distributed
    distributed:
      mixcoord:
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: milvus
                resources:
                  requests:
                    cpu: 500m
                    memory: 1Gi
                  limits:
                    cpu: 600m
                    memory: 2Gi

      datanode:
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: milvus
                resources:
                  requests:
                    cpu: 600m
                    memory: 1Gi
                  limits:
                    cpu: 700m
                    memory: 3Gi

      proxy:
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: milvus
                resources:
                  requests:
                    cpu: 500m
                    memory: 2Gi
                  limits:
                    cpu: 600m
                    memory: 4Gi
      querynode:
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: milvus
                resources:
                  requests:
                    cpu: 800m
                    memory: 3Gi
                  limits:
                    cpu: 900m
                    memory: 4Gi
      streamingnode:
        replicas: 3
        podTemplate:
          spec:
            containers:
              - name: milvus
                resources:
                  requests:
                    cpu: 600m
                    memory: 2Gi
                  limits:
                    cpu: 700m
                    memory: 2Gi
        storageType: Durable
        storage:
          accessModes:
            - ReadWriteOnce
          storageClassName: local-path
          resources:
            requests:
              storage: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9091
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 600m
            memory: 256Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
              - ALL
          runAsGroup: 1000
          runAsNonRoot: true
          runAsUser: 1000
          seccompProfile:
            type: RuntimeDefault
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  tls:
    issuerRef:
      name: milvus-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    external:
      mode: mTLS
    internal:
      mode: TLS
  deletionPolicy: WipeOut
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 2
    disableWriteCheck: false
```

### spec.version

`spec.version` is a required field specifying the name of the [MilvusVersion](/docs/guides/milvus/concepts/milvusversion.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `Milvus` resources,

- `2.6.7`
- `2.6.9`
- `2.6.11`

### spec.objectStorage

`spec.objectStorage` is a required field that specifies the object storage backend used by Milvus. Milvus depends on external object storage (such as MinIO or any S3-compatible service) to store its data.

The configuration is provided via a Kubernetes `Secret` referenced in this field:

```yaml
objectStorage:
  configSecret:
    name: my-release-minio
```
The referenced secret must contain the following keys:

- address – endpoint of the object storage service
- accesskey – username for authentication
- secretkey – password for authentication

All values must be base64-encoded.

In this setup, MinIO is deployed separately (for example, via Helm) and acts as a dependency for Milvus. KubeDB does not manage MinIO directly; it only uses the credentials provided through the secret to connect to the object storage.

### spec.metaStorage

`spec.metaStorage` defines how the metadata store (etcd) used by Milvus is configured. Milvus relies on etcd to manage cluster metadata and coordination. The etcd operator must be installed and running in the user cluster.

- `externallyManaged` indicates whether the etcd cluster is managed outside of KubeDB. 
   - If true, users must provide the etcd endpoints. 
   - If false, KubeDB will create and manage an EtcdCluster resource.

- `endpoints` is the list of etcd client endpoints and required when externallyManaged: true.
- `size (optional)` is the number of etcd nodes to provision and used only when externallyManaged: false.
- `storageType (optional)` defines storage behavior (e.g., durable or ephemeral).
- `storage (optional)` specifies the PersistentVolume configuration for etcd when managed by KubeDB.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `milvus` root user. If not set, KubeDB operator creates a new Secret `{milvus-object-name}-auth` for storing the password for `root` user for each Milvus object.

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Milvus object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Milvus object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for Milvus `root` user.

Example:

```bash
$ kubectl create secret generic milvus-auth -n demo \
--from-literal=username=root \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "milvus-auth" created
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: milvus-auth
  namespace: kubedb
type: kubernetes.io/basic-auth
stringData:
  username: "root"
  password: "Milvus"
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.configuration
`spec.configuration` is an optional field that specifies custom configuration for Milvus cluster. It has the following fields:
- `configuration.secretName` is an optional field that specifies the name of the secret that holds custom configuration files for Milvus cluster.
- `configuration.inline` is an optional field that allows you to provide custom configuration directly in the Milvus object.

```yaml
configuration:
  secretName: my-release-user-config
  inline:
    milvus.yaml: |
      log:
        level: info
        file:
          maxAge: 30
```

### spec.topology

`spec.topology` defines the deployment topology for Milvus. It supports both **Standalone** and **Distributed** (cluster) modes.

### spec.topology.mode

Specifies the deployment mode of Milvus.

- **`Standalone`**: Runs Milvus as a single-node deployment. All components run inside a single pod.
- **`Distributed`**: Runs Milvus as a multi-component distributed cluster.

### spec.topology.standalone

```yaml
topology:
  mode: Standalone
```
When `mode: Standalone` is used:
- All Milvus components run in a single unified deployment.
- No separate component configuration (like datanode, proxy, etc.) is required.
- KubeDB manages all internal services automatically

### spec.topology.distributed

`distributed` contains the configuration for all Milvus components in distributed mode.

#### spec.topology.mixcoord

`mixcoord` is responsible for coordinating metadata and internal cluster orchestration.

- **replicas**: Number of mixcoord pods (default: `1`)
- **podTemplate**: Custom resource requests/limits and pod-level configuration


#### spec.topology.distributed.datanode

`datanode` handles data ingestion and persistence into storage.

- **replicas**: Number of datanode pods (default: `1`)
- **podTemplate**: Resource configuration for each datanode pod

#### spec.topology.distributed.proxy

`proxy` is the entry point for client requests.

- **replicas**: Number of proxy pods (default: `1`)
- **podTemplate**: Resource configuration for proxy pods

#### spec.topology.distributed.querynode

`querynode` executes search and query operations on vector data.

- **replicas**: Number of querynode pods (default: `1`)
- **podTemplate**: Resource configuration for query execution workload


#### spec.topology.distributed.streamingnode

`streamingnode` handles real-time streaming ingestion.

- **replicas**: Number of streamingnode pods (default: `1`)
- **podTemplate**: Resource configuration for streaming workloads

Additional storage configuration for `streamingnode`:

- **storageType**: Defines storage behavior (`Durable` or `Ephemeral`)
- **storage**: PVC specification used when `storageType` is `Durable`

### spec.<node-name>.podTemplate

KubeDB allows providing a template for database pod through `spec.<node-name>.podTemplate`. KubeDB operator will pass the information provided in `spec.<node-name>.podTemplate` to the PetSet created for Milvus cluster.

KubeDB accept following fields to set in `spec.<node-name>.podTemplate:`

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
  - containers
  - imagePullSecrets
  - nodeSelector
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go#L26C1-L279C1).
Uses of some field of `spec.<node-name>.podTemplate` is described below,

#### spec.<node-name>.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.<node-name>.podTemplate.spec.volumes

The `spec.<node-name>.podTemplate.<node-name>.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.<node-name>.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the `podPlacementPolicy`. `name` of the podPlacementPolicy is referred under this attribute. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.
```yaml
spec:
  podPlacementPolicy:
    name: default
```

#### spec.<node-name>.podTemplate.spec.nodeSelector

`spec.<node-name>.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.<node-name>.podTemplate.spec.containers

The `spec.<node-name>.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.<node-name>.podTemplate.spec.containers[].name
The `spec.<node-name>.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.<node-name>.podTemplate.spec.containers[].args
`spec.<node-name>.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.<node-name>.podTemplate.spec.containers[].env

`spec.<node-name>.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the Redis containers.

##### spec.<node-name>.podTemplate.spec.containers[].resources

`spec.<node-name>.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide template for the services created by KubeDB operator for Milvus cluster through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
  - `stats` for is used for the `exporter` service identification.

Milvus comes with one primary services used for client access (Standalone or Proxy in Distributed mode) and four component services for distributed mode (`mixcoord`, `datanode`, `querynode` and `streamingnode`). There are two options for providing serviceTemplates:
- To provide `serviceTemplates` for a specific service, the `serviceTemplates.ports.port` should be equal to the port of that service and `serviceTemplate` will be used for that particular service only.
- However, to provide a common `serviceTemplates`, `serviceTemplates.ports.port` should be empty.

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

### spec.monitor

Milvus managed by KubeDB can be monitored with Prometheus operator out-of-the-box. To learn more,
- [Monitor Apache Milvus with Prometheus operator](/docs/guides/milvus/monitoring/using-prometheus-operator.md)

### spec.tls

`spec.tls` defines the TLS configuration for securing Milvus communication. The KubeDB operator uses [cert-manager](https://cert-manager.io/) to issue and manage certificates. Currently, only **PKCS#8 encoded certificates** are supported.

TLS in Milvus can be configured for:
- **External traffic** (client → Milvus)
- **Internal traffic** (inter-component communication)

---

### Example

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: milvus-issuer
    external:
      mode: mTLS
    internal:
      mode: TLS
```
The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Milvus. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
  - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
  - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.


- `tls.external` - `external` controls TLS for client-facing traffic (gRPC / REST).
  - `TLS` - requires only the server certificate to encrypt communication between client and Milvus proxy.
  - `mTLS` - requires both server and client certificates to enable mutual authentication between client and Milvus proxy.


- `tls.internal` - `internal` enables TLS for inter-component communication within the Milvus cluster.
  - `TLS` - uses only server-side certificates to encrypt communication between internal Milvus components (e.g., proxy, querynode, datanode).
  - `mTLS` - not supported for internal communication; internal traffic is generally secured using one-way TLS only.


- `tls.certificates` - is an `optional` field that specifies a list of certificate configurations used to configure the  certificates. It has the following fields:
  - `alias` - represents the identifier of the certificate. It has the following possible value:
    - `server` - is used for the server certificate configuration.
    - `client` - is used for the client certificate configuration.


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

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Milvus` crd or which resources KubeDB should keep or delete when you delete `Milvus` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- WipeOut
- Halt
- Delete

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

> For more details you can visit [here](https://appscode.com/blog/post/deletion-policy/)

### spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run Apache Milvus cluster [here](/docs/guides/milvus/README.md).
- Deploy [dedicated topology cluster](/docs/guides/milvus/clustering/guide/index.md) for Apache Milvus
- Monitor your Milvus cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/milvus/monitoring/using-prometheus-operator.md).
- Detail concepts of [MilvusVersion object](/docs/guides/milvus/concepts/milvusversion.md).

[//]: # (- Learn to use KubeDB managed Milvus objects using [CLIs]&#40;/docs/guides/milvus/cli/cli.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).