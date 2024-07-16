---
title: Pgpool CRD
menu:
  docs_{{ .version }}:
    identifier: pp-pgpool-concepts
    name: Pgpool
    parent: pp-concepts-pgpool
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Pgpool

## What is Pgpool

`Pgpool` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Pgpool](https://pgpool.net/) in a Kubernetes native way. You only need to describe the desired configuration in a `Pgpool`object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Pgpool Spec

As with all other Kubernetes objects, a Pgpool needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Pgpool object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: pool
spec:
  version: "4.5.0"
  replicas: 1
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  authSecret:
    name: pgpool-auth
    externallyManaged: false
  postgresRef:
    name: ha-postgres
    namespace: demo
  sslMode: verify-ca
  clientAuthMode: cert
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pgpool-ca-issuer
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
  configSecret:
    name: pgpool-config
  initConfig:
    pgpoolConfig:
      log_statement : on
      log_per_node_statement : on
      sr_check_period : 0
      health_check_period : 0
      backend_clustering_mode : 'streaming_replication'
      num_init_children : 5
      max_pool : 75
      child_life_time : 300
      child_max_connections : 0
      connection_life_time : 0
      client_idle_limit : 0
      connection_cache : on
      load_balance_mode : on
      ssl : on
      failover_on_backend_error : off
      log_min_messages : warning
      statement_level_load_balance: on
      memory_cache_enabled: on
  deletionPolicy: WipeOut
  syncUsers: true
  podTemplate:
    spec:
      containers:
        - name: pgpool
          resources:
            limits:
              memory: 2Gi
            requests:
              cpu: 200m
              memory: 256Mi
  serviceTemplates:
    - alias: primary
      spec:
        type: LoadBalancer
        ports:
          - name: http
            port: 9999
```
### spec.version

`spec.version` is a required field specifying the name of the [PgpoolVersion](/docs/guides/pgpool/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `PgpoolVersion` resources,

- `4.4.5`, `4.5.0`

### spec.replicas

`spec.replicas` the number of members in pgpool replicaset. `Minimum` allowed replicas is `1` and `maximum` is `9`.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `pgpool` `pcp` user. If not set, KubeDB operator creates a new Secret `{pgpool-object-name}-auth` for storing the password for `pgpool` pcp user for each Pgpool object.

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Pgpool object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Pgpool object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for `pgpool` pcp user.

Example:

```bash
$ kubectl create secret generic pgpool-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=O9xE1mZZDAdBTbrV
secret "pgpool-auth" created
```

```yaml
apiVersion: v1
data:
  password: "O9xE1mZZDAdBTbrV"
  username: "jhon-doe"
kind: Secret
metadata:
  name: pgpool-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.postgresRef

`spec.postgresRef` is a required field that points to the `appbinding` associated with the backend postgres. If the postgres is KubeDB managed an appbinding will be created automatically upon creating the postgres. If the postgres is not KubeDB managed then you need to create an appbinding yourself. `spec.postgresRef` takes the name (`spec.postgresRef.Name`) of the appbinding and the namespace (`spec.postgresRef.Namespace`) where the appbinding is created.

### spec.sslMode

Enables TLS/SSL or mixed TLS/SSL used for all network connections. The value of `sslMode` field can be one of the following:

|     Value     | Description                                                                                                                                                                   |
|:-------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|  `disabled`   | The server does not use TLS/SSL.                                                                                                                                              |
|   `require`   | The server uses and accepts only TLS/SSL encrypted connections.                                                                                                               |
|  `verify-ca`  | The server uses and accepts only TLS/SSL encrypted connections and client want to be sure that client connect to a server that client trust.                                  |
| `verify-full` | The server uses and accepts only TLS/SSL encrypted connections and client want to be sure that client connect to a server client trust, and that it's the one client specify. |

The specified ssl mode will be used by health checker and exporter of Pgpool.

### spec.clientAuthMode

The value of `clientAuthMode` field can be one of the following:

|     Value     | Description                                                                                                                                                                   |
|:-------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|    `scram`    | The server uses scram-sha-256 authentication method to authenticate the users.                                                                                                |
|     `md5`     | The server uses md5 authentication method to authenticate the users.                                                                                                          |
|    `cert`     | The server uses tls certificates to authenticate the users and for this `sslMode` must not be disabled                                                                        |

The  `pool_hba.conf` of Pgpool will have the configuration based on the specified clientAuthMode.

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

### spec.monitor

Pgpool managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor Pgpool with builtin Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md)
- [Monitor Pgpool with Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for Pgpool. You can provide the custom configuration in a secret, then you can specify the secret name `spec.configSecret.name`.

> Please note that, the secret key needs to be `pgpool.conf`.

To learn more about how to use a custom configuration file see [here](/docs/guides/pgpool/configuration/using-config-file.md).

NB. If `spec.configSecret` is set, then `spec.initConfig` needs to be empty.

### spec.initConfig

`spec.initConfig` is an optional field that allows users to provide custom configuration for Pgpool while initializing.

To learn more about how to use init configuration see [here](/docs/guides/pgpool/configuration/using-init-config.md).

NB. If `spec.initConfig` is set, then `spec.configSecret` needs to be empty.

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Pgpool` CR or which resources KubeDB should keep or delete when you delete `Pgpool` CR. KubeDB provides following four deletion policies:

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

### spec.syncUsers

`spec.syncUsers` is an optional field by default its value is false. If it is true, you can provide a secret with username and password as key and with some desired labels to sync PostgreSQL users to Pgpool in runtime.

The contains a `user` key and a `password` key which contains the `username` and `password` respectively.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  labels:
    app.kubernetes.io/instance: <Appbinding name mentioned in .spec.postgresRef.name>
    app.kubernetes.io/name: postgreses.kubedb.com
  name: pg-user
  namespace: <Namespace mentioned in .spec.postgresRef.namespace>
stringData:
  password: "12345"
  username: alice
```
In every `10 seconds` KubeDB operator will sync all the users to Pgpool. 

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).


### spec.podTemplate

KubeDB allows providing a template for pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Pgpool.

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
