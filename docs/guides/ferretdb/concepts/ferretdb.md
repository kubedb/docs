---
title: FerretDB CRD
menu:
  docs_{{ .version }}:
    identifier: fr-ferretdb-concepts
    name: FerretDB
    parent: fr-concepts-ferretdb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# FerretDB

## What is FerretDB

`FerretDB` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [FerretDB](https://www.ferretdb.com/) in a Kubernetes native way. You only need to describe the desired configuration in a `FerretDB`object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## FerretDB Spec

As with all other Kubernetes objects, a FerretDB needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example FerretDB object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb
  namespace: demo
spec:
  version: "2.0.0"
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  backend:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 500Mi
  authSecret:
    kind: Secret
    name: ferretdb-auth
    externallyManaged: false
  sslMode: requireSSL
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: ferretdb-ca-issuer
      kind: Issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
  server:
    primary:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: ferretdb
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 200m
                  memory: 256Mi
    secondary:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: ferretdb
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 200m
                  memory: 256Mi                  
  serviceTemplates:
    - alias: primary
      spec:
        type: ClusterIP
        ports:
          - name: http
            port: 9999
```

### spec.version

`spec.version` is a required field specifying the name of the [FerretDBVersion](/docs/guides/ferretdb/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `FerretDBVersion` resources,

- `1.18.0`, `1.23.0`, `1.24.0`, `2.0.0`

### spec.storage

`spec.storage` is a required field specifying the storage specification of backend Postgres. KubeDB will create backend Postgres according to this field. 

If you don't set `spec.storageType:` to `Ephemeral` then `spec.storage` field is required. This field specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `ferretdb`. If not set, KubeDB operator creates a new Secret `{ferretdb-object-name}-auth` for storing the password for `ferretdb` user for each FerretDB object.
As FerretDB use backend's authentication mechanisms till now, this secret is basically a copy of backend postgres.

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the FerretDB object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the FerretDB object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `ferretdb` user.

Example:

```bash
$ kubectl create secret generic ferretdb-auth -n demo \
--from-literal=username=jhon \
--from-literal=password=O9xE1mZZDAdBTbrV
secret "ferretdb-auth" created
```

```yaml
apiVersion: v1
data:
  password: "O9xE1mZZDAdBTbrV"
  username: "jhon"
kind: Secret
metadata:
  name: ferretdb-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.sslMode

Enables TLS/SSL or mixed TLS/SSL used for all network connections. The value of [`sslMode`](https://docs.ferretdb.io/security/tls-connections/) field can be one of the following:

|    Value     | Description                                                                                                                    |
| :----------: | :----------------------------------------------------------------------------------------------------------------------------- |
|  `disabled`  | The server does not use TLS/SSL.                                                                                               |
| `requireSSL` | The server uses and accepts only TLS/SSL encrypted connections.                                                                |

### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the FerretDB. KubeDB uses [cert-manager](https://cert-manager.io/) v1 api to provision and manage TLS certificates.

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

### spec.monitor

FerretDB managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor FerretDB with builtin Prometheus](/docs/guides/ferretdb/monitoring/using-builtin-prometheus.md)
- [Monitor FerretDB with Prometheus operator](/docs/guides/ferretdb/monitoring/using-prometheus-operator.md)

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `FerretDB` CR or which resources KubeDB should keep or delete when you delete `FerretDB` CR. KubeDB provides following four deletion policies:

- DoNotTerminate
- Delete
- WipeOut (`Default`)

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete FerretDB CR for different deletion policies,

| Behavior                  | DoNotTerminate |    Delete    | WipeOut  |
|---------------------------| :------------: |:------------:| :------: |
| 1. Block Delete operation |    &#10003;    |   &#10007;   | &#10007; |
| 2. Delete PetSet          |    &#10007;    |   &#10003;   | &#10003; |
| 3. Delete Services        |    &#10007;    |   &#10003;   | &#10003; |
| 4. Delete Secrets         |    &#10007;    |   &#10007;   | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` deletion policy by default.

> For more details you can visit [here](https://appscode.com/blog/post/deletion-policy/)

### spec.server

After FerretDB version 2.0.0, FerretDB uses PostgreSQL + DocumentDB extension as the database storage and currently supports replication using the Write-Ahead Logging (WAL) streaming method.

This field holds the necessary information about FerretDB Primary and FerretDB Secondary server. It accepts the following fields,

- `spec.server.primary` : Holds the Primary sever information.
- `spec.server.secondary` : Holds Secondary server information.

Both `spec.server.primary` and `spec.server.secondary` has the following fields,

- `replicas` for the number of members in ferretdb primary/secondary replicaset.
- `podTemplate` for pod template of ferretdb primary/secondary server.

#### spec.server.primary.replicas

`spec.server.primary.replicas` the number of members in ferretdb primary replicaset.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

#### spec.server.primary.podTemplate

KubeDB allows providing a template for pod through `spec.server.primary.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the Primary server PetSet created for FerretDB.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
    - annotations (pod's annotation)
    - labels (pod's labels)
- controller:
    - annotations (statefulset's annotation)
    - labels (statefulset's labels)
- spec:
    - volumes
    - initContainers
    - containers
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

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/39bf8b2/api/v2/types.go#L44-L279). Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.podPlacementPolicy
`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the `podPlacementPolicy`. `name` of the podPlacementPolicy is referred under this attribute. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.
```yaml
spec:
  podPlacementPolicy:
    name: default
```

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide template for the services created by KubeDB operator for Kafka cluster through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
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