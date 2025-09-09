---
title: Hazelcast CRD
menu:
  docs_{{ .version }}:
    identifier: hz-hazelcast-concepts
    name: Hazelcast
    parent: hz-concepts-hazelcast
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Hazelcast

## What is Hazelcast

`Hazelcast` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Hazelcast](https://hazelcast.com/) in a Kubernetes native way. You only need to describe the desired database configuration in a Hazelcast object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Hazelcast Spec

As with all other Kubernetes objects, a Hazelcast needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Hazelcast object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Hazelcast","metadata":{"annotations":{},"name":"hazelcast-sample","namespace":"demo"},"spec":{"deletionPolicy":"Halt","enableSSL":true,"licenseSecret":{"name":"hz-license-key"},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"2Gi"}},"storageClassName":"longhorn"},"tls":{"certificates":[{"alias":"server","dnsNames":["localhost"],"ipAddresses":["127.0.0.1"],"subject":{"organizations":["kubedb"]}},{"alias":"client","dnsNames":["localhost"],"ipAddresses":["127.0.0.1"],"subject":{"organizations":["kubedb"]}}],"issuerRef":{"apiGroup":"cert-manager.io","kind":"ClusterIssuer","name":"self-signed-issuer"}},"version":"5.5.2"}}
  creationTimestamp: "2025-06-11T07:35:38Z"
  finalizers:
    - kubedb.com
  generation: 2
  name: hazelcast-sample
  namespace: demo
  resourceVersion: "1180125"
  uid: c86fe3d3-276a-4124-a1cf-d7f5409ee61f
spec:
  deletionPolicy: Halt
  enableSSL: true
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  keystoreSecret:
    name: hazelcast-sample-keystore-cred
  licenseSecret:
    name: hz-license-key
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
        - livenessProbe:
            failureThreshold: 10
            httpGet:
              path: /hazelcast/health/node-state
              port: 5701
              scheme: HTTPS
            initialDelaySeconds: 30
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 10
          name: hazelcast
          readinessProbe:
            failureThreshold: 10
            httpGet:
              path: /hazelcast/health/ready
              port: 5701
              scheme: HTTPS
            initialDelaySeconds: 30
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 10
          resources:
            limits:
              memory: 1536Mi
            requests:
              cpu: 500m
              memory: 1536Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsNonRoot: true
            runAsUser: 65534
            seccompProfile:
              type: RuntimeDefault
      initContainers:
        - name: hazelcast-init
          resources:
            limits:
              memory: 512Mi
            requests:
              cpu: 200m
              memory: 256Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsNonRoot: true
            runAsUser: 65534
            seccompProfile:
              type: RuntimeDefault
      podPlacementPolicy:
        name: default
      securityContext:
        fsGroup: 65534
      terminationGracePeriodSeconds: 600
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
    storageClassName: longhorn
  storageType: Durable
  tls:
    certificates:
      - alias: server
        dnsNames:
          - localhost
        ipAddresses:
          - 127.0.0.1
        subject:
          organizations:
            - kubedb
      - alias: client
        dnsNames:
          - localhost
        ipAddresses:
          - 127.0.0.1
        subject:
          organizations:
            - kubedb
    issuerRef:
      apiGroup: cert-manager.io
      kind: ClusterIssuer
      name: self-signed-issuer
  version: 5.5.2
```


### spec.version

`spec.version` is a required field specifying the name of the [HazelcastVersion](/docs/guides/hazelcast/concepts/hazelcastversion.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `HazelcastVersion` crds,

-  `5.5.2`

### spec.disableSecurity

`spec.disableSecurity` is an optional field that decides whether Hazelcast instance will be secured by auth or no.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `Hazelcast` superuser. If not set, KubeDB operator creates a new Secret `{Hazelcast-object-name}-auth` for storing the password for `admin` superuser.

We can use this field in 3 mode.

1. Using an external secret. In this case, You need to create an auth secret first with required fields, then specify the secret name when creating the Hazelcast object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```
2. Specifying the secret name only. In this case, You need to specify the secret name when creating the Hazelcast object using `spec.authSecret.name`. `externallyManaged` is by default false.
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

AuthSecret contains a `username` key and a `password` key which contains the `username` and `password` respectively for `Hazelcast` superuser.

Example:

```bash
$ kubectl create secret generic hazelcast-sample-auth -n demo \
--from-literal=username=admin \
--from-literal=password=6q8u_2jMOW-OOZXk
secret "hazelcast-sample-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: amhvbi1kb2U=
kind: Secret
metadata:
  name: hazelcast-sample-auth
  namespace: demo
type: Opaque
```

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

> Clsuter Mode: all peers are equal in the cluster.

### spec.replicas

`spec.replicas`  specifies the number of nodes (ie. pods) in the Hazelcast cluster. The default value of this field is `1`.

```yaml
spec:
  replicas: 3
```

### spec.storage

If you set `spec.storageType:` to `Durable`, then  `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the Petset created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.javaOpts

We can add java environment variables using this attribute.

### spec.monitor

Hazelcast managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box.


### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for Hazelcast. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any Kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc.

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the Petset created for Hazelcast server.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
    - annotations (pod's annotation)
- controller:
    - annotations (petset's annotation)
- spec:
    - resources
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

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/39bf8b2/api/v2/types.go#L44-L279).
Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.imagePullSecret

`KubeDB` provides the flexibility of deploying Hazelcast server from a private Docker registry.
#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account name matching Hazelcast crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide a template for the services created by KubeDB operator for Hazelcast server through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
    - `primary` is used for the primary service identification.
    - `standby` is used for the secondary service identification.
    - `stats` is used for the exporter service identification.

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

### spec.tls

> The ReconfigureTLS only works with the [Cert-Manager](https://cert-manager.io/docs/concepts/) managed certificates. [Installation guide](https://cert-manager.io/docs/installation/).

`spec.tls` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `ReconfigureTLS`. It specifies the necessary information required to add or remove or update the TLS configuration of the Hazelcast cluster. It consists of the following sub-fields:

- `tls.remove` ( `bool` | `false` ) - tells the operator to remove the TLS configuration for the HTTP layer. The transport layer is always secured with certificates, so the removal process does not affect the transport layer.
- `tls.rotateCertificates` ( `bool` | `false`) - tells the operator to renew all the certificates.
- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Hazelcast. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
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

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Hazelcast` crd or which resources KubeDB should keep or delete when you delete `Hazelcast` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Hazelcast crd for different deletion policies,

| Behavior                            | DoNotTerminate |   Halt   |  Delete  | WipeOut  |
|-------------------------------------|:--------------:|:--------:|:--------:|:--------:|
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete Petset                    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshot data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |
If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` deletion policy by default.

### spec.halted
Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run a Hazelcast server [here](/docs/guides/hazelcast/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
