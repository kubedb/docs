---
title: Cassandra CRD
menu:
  docs_{{ .version }}:
    identifier: cas-cassandra-concepts
    name: Cassandra
    parent: cas-concepts-cassandra
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra

## What is Cassandra

`Cassandra` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Cassandra](https://cassandra.apache.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a `Cassandra` object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Cassandra Spec

As with all other Kubernetes objects, a Cassandra needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Cassandra object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra
  namespace: demo
spec:
  authSecret:
    name: cassandra-admin-cred
  configSecret:
    name: cassandra-custom-config
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  keystoreCredSecret:
    name: cassandra-keystore-cred
  deletionPolicy: DoNotTerminate
  tls:
    certificates:
      - alias: server
        secretName: cassandra-server-cert
      - alias: client
        secretName: cassandra-client-cert
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: cassandra-ca-issuer
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 4Gi
                    cpu: 2000m
                  requests:
                    memory: 1Gi
                    cpu: 500m
            securityContext:
              runAsUser: 999
              fsGroup: 999
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 56790
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  version: 5.0.3
```

### spec.version

`spec.version` is a required field specifying the name of the [CassandraVersion](/docs/guides/cassandra/concepts/cassandraversion.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `Cassandra` resources,

- `4.1.8`
- `5.0.3`

### spec.replicas

`spec.replicas` the number of members in Cassandra replicaset.

If `spec.topology` is set, then `spec.replicas` needs to be empty. Instead use `spec.topology.rack[ind].replicas`.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `cassandra` admin user. If not set, KubeDB operator creates a new Secret `{cassandra-object-name}-auth` for storing the password for `admin` user for each Cassandra object.

We can use this field in 3 mode.
1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Cassandra object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Cassandra object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `user` key and a `password` key which contains the `username` and `password` respectively for Cassandra `admin` user.

Example:

```bash
$ kubectl create secret generic cassandra-auth -n demo \
--from-literal=username=jhon-doe \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "cassandra-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: cassandra-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### spec.configSecret

`spec.configSecret` is an optional field that points to a Secret used to hold custom Cassandra configuration. If not set, KubeDB operator will use default configuration for Cassandra.

### spec.topology

`spec.topology` represents the topology configuration for Cassandra cluster.

When `spec.topology` is set, the following fields needs to be empty, otherwise validating webhook will throw error.

- `spec.replicas`
- `spec.podTemplate`
- `spec.storage`

#### spec.topology.rack

`rack` represents a logical grouping of nodes of Cassandra cluster. `spec.topology.rack[]` is an array of RackSpec.  It is a mandatory field when `spec.topology` is specified. Each RackSpec describes the configuration of a single rack — including its name, number of replicas, pod template, and storage options.

Available configurable fields:

- `name` (`: "rack-east"`) — is a `mandatory` field that specifies the unique name of the rack. Cassandra uses this name to assign and distribute replicas logically across racks.

- `replicas` (`: "3"`) — is an `optional` field to specify the number of Cassandra nodes (pods) to deploy in this rack. This field must hold a value greater than `0`.

- `podTemplate` (`: "<custom pod template>"`) — is an `optional` field that allows you to customize pod-level configurations (like affinity, tolerations, nodeSelector, container resources) for pods within this rack.

- `storage` (`: "resources.requests.storage: 10Gi"`) — is an optional field to define how persistent storage should be configured for the pods in this rack. It uses a standard PersistentVolumeClaimSpec format.

- `storageType` (`: "Durable"`) — is an `optional` field to specify whether the pods in this rack should use `Durable` (persistent disk-backed) or `Ephemeral` (temporary) storage. Defaults to `Durable`.

### spec.tls

`spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). 

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: cassandra-issuer
    certificates:
    - alias: server
      privateKey:
        encoding: PKCS8
      secretName: cassandra-client-cert
      subject:
        organizations:
        - kubedb
    - alias: http
      privateKey:
        encoding: PKCS8
      secretName: cassandra-server-cert
      subject:
        organizations:
        - kubedb
```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Cassandra. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
    - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
    - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
    - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.

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

### spec.monitor

Cassandra managed by KubeDB can be monitored with Prometheus operator out-of-the-box. To learn more,
- [Monitor Apache Cassandra with Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md)
- [Monitor Apache Cassandra with Built-in Prometheus](/docs/guides/cassandra/monitoring/using-builtin-prometheus.md)

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Cassandra cluster.

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
Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.podTemplate.spec.volumes

The `spec.podTemplate.spec.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the `podPlacementPolicy`. `name` of the podPlacementPolicy is referred under this attribute. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.
```yaml
spec:
  podPlacementPolicy:
    name: default
```

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

### spec.serviceTemplates

You can also provide template for the services created by KubeDB operator for Cassandra cluster through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
    - `stats` for is used for the `exporter` service identification.

There are two options for providing serviceTemplates:
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


#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the Cassandra containers.

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Cassandra` crd or which resources KubeDB should keep or delete when you delete `Cassandra` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- WipeOut
- Halt
- Delete

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Cassandra crd for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.


## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run Apache Cassandra cluster [here](/docs/guides/cassandra/README.md).
- Deploy [dedicated topology cluster](/docs/guides/cassandra/clustering/guide/index.md) for Apache Cassandra
- Monitor your Cassandra cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).
- Detail concepts of [CassandraVersion object](/docs/guides/cassandra/concepts/cassandraversion.md).

[//]: # (- Learn to use KubeDB managed Cassandra objects using [CLIs]&#40;/docs/guides/cassandra/cli/cli.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
