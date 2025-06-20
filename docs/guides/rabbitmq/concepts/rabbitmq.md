---
title: RabbitMQ CRD
menu:
  docs_{{ .version }}:
    identifier: rm-concepts
    name: RabbitMQ
    parent: rm-concepts-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQ

## KubeDB managed RabbitMQ

`RabbitMQ` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [RabbitMQ](https://www.rabbitmq.com/) in a Kubernetes native way. You only need to describe the desired database configuration in a RabbitMQ object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## RabbitMQ Spec

As with all other Kubernetes objects, a RabbitMQ needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example RabbitMQ object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq
  namespace: rabbit
spec:
  version: "3.13.2"
  authSecret:
    name: rabbit-auth
  configSecret:
    name: rabbit-custom-config
  enableSSL: true
  replicas: 4
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
  serviceTemplates:
  - alias: primary
    spec:
      type: LoadBalancer
  - alias: stats
    spec:
      type: LoadBalancer
  podTemplate:
    spec:
      containers:
        - name: "rabbitmq"
          resources:
            requests:
              cpu: "500m"
            limits:
              cpu: "600m"
              memory: "1.5Gi"
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: rabbit-ca-issuer
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
          - dev@appscode.com
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 2
    disableWriteCheck: false
  storageType: Durable
  deletionPolicy: Halt  
```

### spec.autoOps
AutoOps is an optional field to control the generation of versionUpdate & TLS-related recommendations.

### spec.version

`spec.version` is a required field specifying the name of the [RabbitMQVersion](/docs/guides/rabbitmq/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `RabbitMQVersion` resources,

- `3.12.12`, `3.13.2`

### spec.replicas

`spec.replicas` the number of nodes in RabbitMQ cluster.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `RabbitMQ` superuser. If not set, KubeDB operator creates a new Secret `{RabbitMQ-object-name}-auth` for storing the password for `RabbitMQ` superuser for each RabbitMQ object. 

We can use this field in 3 mode. 
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the RabbitMQ object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the RabbitMQ object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `RabbitMQ` superuser.

Example:

```bash
$ kubectl create secret generic -n demo rabbit-auth \
  --from-literal=username=rabbit-admin \
  --from-literal=password=mypassword
secret/rabbit-auth created
```

```yaml
apiVersion: v1
data:
  password: bXlwYXNzd29yZA==
  username: cmFiYml0LWFkbWlu
kind: Secret
metadata:
  creationTimestamp: "2024-09-09T03:56:36Z"
  name: rabbit-auth
  namespace: demo
  resourceVersion: "263545"
  uid: 4734f693-9ff8-4f42-bcac-ab9b5ba17afd
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the RabbitMQ. KubeDB uses [cert-manager](https://cert-manager.io/) v1 api to provision and manage TLS certificates.

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

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create RabbitMQ database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. 
In this case, you don't have to specify `spec.storage` field. Specify `spec.ephemeralStorage` spec instead.

### spec.storage

Since 0.9.0-rc.0, If you set `spec.storageType:` to `Durable`, then `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.monitor

RabbitMQ managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor RabbitMQ with builtin Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md)
- [Monitor RabbitMQ with Prometheus operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for RabbitMQ. You can provide the custom configuration in a secret, then you can specify the secret name `spec.configSecret.name`.

> Please note that, the secret key needs to be `rabbitmq.conf`.

To learn more about how to use a custom configuration file see [here](/docs/guides/rabbitmq/configuration/using-config-file.md).

### spec.podTemplate

KubeDB allows providing a template for pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for RabbitMQ.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
  - labels (pod's labels)
- controller:
  - annotations (PetSet's annotation)
  - labels (PetSet's labels)
- spec:
  - volumes
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


### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `RabbitMQ` CR or which resources KubeDB should keep or delete when you delete `RabbitMQ` CR. KubeDB provides following four deletion policies:

- DoNotTerminate
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Pgpool CR for different deletion policies,

| Behavior                  | DoNotTerminate |    Delete    | WipeOut  |
|---------------------------| :------------: |:------------:| :------: |
| 1. Block Delete operation |    &#10003;    |   &#10007;   | &#10007; |
| 2. Delete PetSet          |    &#10007;    |   &#10003;   | &#10003; |
| 3. Delete Services        |    &#10007;    |   &#10003;   | &#10003; |
| 4. Delete Secrets         |    &#10007;    |   &#10007;   | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` deletion policy by default.


## spec.healthChecker
It defines the attributes for the health checker. 
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run a RabbitMQ database [here](/docs/guides/rabbitmq/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
