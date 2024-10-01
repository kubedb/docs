---
title: SingleStore CRD
menu:
  docs_{{ .version }}:
    identifier: sdb-singlestore-concepts
    name: SingleStore
    parent: sdb-concepts-singlestore
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# SingleStore

## What is SingleStore

`SingleStore` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [SingleStore](https://www.singlestore.com/) in a Kubernetes native way. You only need to describe the desired database configuration in a MongoDB object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## SingleStore Spec

As with all other Kubernetes objects, a SingleStore needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example SingleStore object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      configSecret:
        name: sdb-configuration
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "4Gi"
                cpu: "1000m"
              requests:
                memory: "2Gi"
                cpu: "500m"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 3
      configSecret:
        name: sdb-configuration
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "5Gi"
                  cpu: "1100m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                     
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 40Gi
  storageType: Durable
  licenseSecret:
    name: license-secret
  authSecret:
    name: given-secret
  init:
    script:
      configMap:
        name: sdb-init-script
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9104
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: sdb-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
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
```

### spec.version

`spec.version` is a required field specifying the name of the [SinglestoreVersion](/docs/guides/mongodb/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `SinglestoreVersion` resources,

- `8.1.32`
- `8.5.7`, `8.5.30`
- `8.7.10`

### spec.topology

`spec.topology` is an optional field that enables you to specify the clustering mode.

- `aggregator` or `leaf` are optional field that configure cluster mode that contains the following fields:
  - `replicas` the number of nodes of `aggregator` and `leaf` in cluster mode.
  - `configSecret` is an optional field that points to a Secret used to hold custom SingleStore configuration.
  - `podTemplate` providing a template for database. KubeDB operator will pass the information provided in `podTemplate` to the PetSet created for the SingleStore database. KubeDB accepts the following fields to set in `podTemplate:`
    - metadata:
      - annotations (pod's annotation)
    - controller:
      - annotations (petset's annotation)
    - spec:
      - initContainers
      - imagePullSecrets
      - resources
      - containers
      - nodeSelector
      - serviceAccountName
      - securityContext
      - tolerations
      - imagePullSecrets
      - podPlacementPolicy
      - volumes
  - If you set `spec.storageType` to `Durable`, then `storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
    - `storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
    - `storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
    - `storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.
  To learn how to configure `storage`, please visit the links below:
  - https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Kafka cluster using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume.

### spec.licenseSecret

`spec.licenseSecret` is a mandatory fields points to a secret used to pass SingleStore license.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `singlestore` root user. If not set, the KubeDB operator creates a new Secret `{singlestore-object-name}-cred` for storing the password for `singlestore` root user for each SingleStore object. If you want to use an existing secret please specify that when creating the SingleStore object using `spec.authSecret.name`.

This secret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `singlestore` root user. Here, the value of `user` key is fixed to be `root`.

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

Example:

```bash
$ kubectl create secret generic sdb-cred -n demo \
--from-literal=user=root \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "sdb-cred" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  user: cm9vdA==
kind: Secret
metadata:
  name: sdb-cred
  namespace: demo
type: Opaque
```

#### Initialize via Script

To initialize a SingleStore database using a script (shell script, sql script, etc.), set the `spec.init.script` section when creating a SingleStore object. It will execute files alphabetically with extensions `.sh` , `.sql`  and `.sql.gz` that is found in the repository. The scripts inside child folders will be skipped. script must have the following information:

- [VolumeSource](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes): Where your script is loaded from.

Below is an example showing how a script from a configMap can be used to initialize a SingleStore database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb
  namespace: demo
spec:
  version: 8.7.10
  init:
    script:
      configMap:
        name: sdb-init-script
  licenseSecret:
    name: license-secret
```

In the above example, KubeDB operator will launch a Job to execute all js script of `sdb-init-script` in alphabetical order once PetSet pods are running. For more details tutorial on how to initialize from script, please visit [here](/docs/guides/mysql/initialization/index.md).

### spec.monitor

SingleStore managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator. To learn more,

- [Monitor SingleStore with builtin Prometheus](/docs/guides/singlestore/monitoring/builtin-prometheus/index.md)
- [Monitor SingleStore with Prometheus operator](/docs/guides/singlestore/monitoring/prometheus-operator/index.md)

### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the SingleStore.

The following fields are configurable in the `spec.tls` section:

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.

  - `apiGroup` is the group name of the resource being referenced. The value for `Issuer` or   `ClusterIssuer` is "cert-manager.io"   (cert-manager v0.12.0 and later).
  - `kind` is the type of resource being referenced. KubeDB supports both `Issuer`   and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.

- `certificates` (optional) are a list of certificates used to configure the server and/or client certificate. It has the following fields:
  
  - `alias` represents the identifier of the certificate. It has the following possible value:
    - `server` is used for server certificate identification.
    - `client` is used for client certificate identification.
    - `metrics-exporter` is used for metrics exporter certificate identification.
  - `secretName` (optional) specifies the k8s secret name that holds the certificates.
    >This field is optional. If the user does not specify this field, the default secret name will be created in the following format: `<database-name>-<cert-alias>-cert`.
  - `subject` (optional) specifies an `X.509` distinguished name. It has the following possible field,
    - `organizations` (optional) are the list of different organization names to be used on the Certificate.
    - `organizationalUnits` (optional) are the list of different organization unit name to be used on the Certificate.
    - `countries` (optional) are the list of country names to be used on the Certificate.
    - `localities` (optional) are the list of locality names to be used on the Certificate.
    - `provinces` (optional) are the list of province names to be used on the Certificate.
    - `streetAddresses` (optional) are the list of a street address to be used on the Certificate.
    - `postalCodes` (optional) are the list of postal code to be used on the Certificate.
    - `serialNumber` (optional) is a serial number to be used on the Certificate.
    You can found more details from [Here](https://golang.org/pkg/crypto/x509/pkix/#Name)  

  - `duration` (optional) is the period during which the certificate is valid.
  - `renewBefore` (optional) is a specifiable time before expiration duration.
  - `dnsNames` (optional) is a list of subject alt names to be used in the Certificate.
  - `ipAddresses` (optional) is a list of IP addresses to be used in the Certificate.
  - `uriSANs` (optional) is a list of URI Subject Alternative Names to be set in the Certificate.
  - `emailSANs` (optional) is a list of email Subject Alternative Names to be set in the Certificate.

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

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `singlestore` crd or which resources KubeDB should keep or delete when you delete `singlestore` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- WipeOut
- Halt
- Delete

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete MySQL crd for different termination policies,

| Behavior                  | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
|---------------------------| :------------: | :------: | :------: | :------: |
| 1. Block Delete operation |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet          |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services        |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs            |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets         |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots       |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

## Next Steps

- Learn how to use KubeDB to run a SingleStore database [here](/docs/guides/singlestore/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).