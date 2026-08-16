---
title: Etcd CRD
menu:
  docs_{{ .version }}:
    identifier: etcd-etcd-concepts
    name: Etcd
    parent: etcd-concepts-etcd
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd

## What is Etcd

`Etcd` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [etcd](https://etcd.io/) in a Kubernetes native way. You only need to describe the desired database configuration in an Etcd object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

etcd is a strongly consistent, distributed key-value store that replicates data with the [Raft](https://raft.github.io/) consensus algorithm. Unlike most other databases KubeDB manages, etcd has **no standalone-vs-cluster topology switch**: a one member deployment is simply `spec.replicas: 1`, and a highly available deployment is `spec.replicas: 3` (or `5`). Every member serves the client API and forwards writes to the current Raft leader, so there is no primary/standby split and no read-replica concept. Members are added and removed through etcd's own membership API rather than by editing a config file, so KubeDB drives scaling through that API.

## Etcd Spec

As with all other Kubernetes objects, an Etcd needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Etcd object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-cluster
  namespace: demo
spec:
  version: 3.6.4
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  authSecret:
    name: etcd-cluster-auth
    externallyManaged: false
  configuration:
    tuning:
      quotaBackendBytes: 8589934592
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
      snapshotCount: 10000
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          app: kubedb
        interval: 10s
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: etcd-ca-issuer
    certificates:
      - alias: server
        secretName: etcd-cluster-server-cert
      - alias: client
        secretName: etcd-cluster-client-cert
      - alias: peer
        secretName: etcd-cluster-peer-cert
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
    controller:
      annotations:
        passMe: ToPetSet
    spec:
      serviceAccountName: my-service-account
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
        - name: myregistrykey
      containers:
        - name: etcd
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              memory: 2Gi
  serviceTemplates:
    - alias: primary
      metadata:
        annotations:
          passMe: ToService
      spec:
        type: NodePort
        ports:
          - name: client
            port: 2379
  archiver:
    pause: false
    ref:
      name: etcd-archiver
      namespace: demo
  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 1
  deletionPolicy: Halt
  halted: false
```

### spec.version

`spec.version` is a required field specifying the name of the [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `EtcdVersion` crds,

- `3.5.21`
- `3.6.4`

### spec.replicas

`spec.replicas` is the number of etcd members in the cluster. If not set, it defaults to `3`.

An **odd** number of members is strongly recommended: a cluster of `2n+1` members tolerates `n` failures, while adding an even member only grows the quorum size without improving fault tolerance. The validating webhook only rejects values below `1`; even numbers are allowed but discouraged.

KubeDB uses a `PodDisruptionBudget` so that a majority of members stays available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions), keeping quorum intact.

Changing `spec.replicas` (directly, or through a `HorizontalScaling`
[EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)) is reconciled against etcd's actual member list:

- **Scale up** — the next ordinal is first added to the cluster as a **learner**. A learner replicates the log but does not vote and does not count toward quorum, so a cold member can never stall the cluster. Once its Raft revision has caught up with the leader's, the operator **promotes** it to a voting member.
- **Scale down** — the **highest ordinal** member is removed from etcd's member list first, and only then is the PetSet shrunk, so etcd never sees the pod disappear before it has been un-registered.

The operator performs at most one membership mutation per reconcile pass, which is what etcd's membership API expects.

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for the etcd data directory. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB provisions the members with `EmptyDir` volumes instead of PVCs — all data is lost when a pod is recreated, so this is only appropriate for testing.

> `spec.storageType: Ephemeral` cannot be combined with `spec.deletionPolicy: Halt`, since there is nothing to halt onto. The webhook rejects that combination.

### spec.storage

If you set `spec.storageType` to `Durable`, then `spec.storage` is a required field that specifies the StorageClass of PVCs dynamically allocated to store the etcd data directory (`/var/lib/etcd`). This storage spec will be passed to the PetSet created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don't necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

> etcd is latency sensitive. Its backend commits every write to disk before acknowledging it, so a StorageClass backed by SSD/NVMe is strongly recommended for production clusters.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for the etcd **`root`** user — the superuser of etcd's built-in RBAC. If not set, the KubeDB operator creates a new Secret `{etcd-object-name}-auth` for storing the password of the `root` user.

We can use this field in 3 modes.

1. Using an external secret. In this case, you need to create an auth secret first with the required fields, then specify the secret name when creating the Etcd object using `spec.authSecret.name` & set `spec.authSecret.externallyManaged` to true.

```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only. In this case, you need to specify the secret name when creating the Etcd object using `spec.authSecret.name`. `externallyManaged` is by default false.

```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything for you. In this case, no work for you.

The auth secret contains a `username` key and a `password` key which contain the username and the password respectively for the etcd `root` user.

Example:

```bash
$ kubectl create secret generic etcd-auth -n demo \
  --type=kubernetes.io/basic-auth \
  --from-literal=username=root \
  --from-literal=password=6q8u_2jMOW-OOZXk
secret "etcd-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dV8yak1PVy1PT1pYaw==
  username: cm9vdA==
kind: Secret
metadata:
  name: etcd-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

Secrets provided by users are not managed by KubeDB, and therefore won't be modified or garbage collected by the KubeDB operator.

`spec.authSecret` also carries `rotateAfter`, which the Recommendation Engine uses to propose a `RotateAuth` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) once the credential reaches the given age.

### spec.init

`spec.init` is an optional section that can be used to initialize a newly created Etcd cluster. For etcd, the interesting sub-section is `spec.init.archiver`, which restores the cluster from a KubeStash snapshot taken by an [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md).

```yaml
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: etcd-full-repo
        namespace: demo
      recoveryTimestamp: "2024-05-02T10:02:45Z"
```

`spec.init.archiver` restores **at bootstrap time**, into a brand new cluster. etcd requires a seed member's data directory to be either genuinely empty or a fully restored, valid snapshot, so a snapshot can never be replayed into a *running* cluster. The operator therefore pre-creates the ordinal-0 PVC, runs the `RestoreSession` against it, and only then creates the PetSet and starts the first member. While that happens, `status.phase` stays `Provisioning`, and the restore's outcome is reported on the `SuccessfullyDataRestored` condition.

> To restore into an `Etcd` object that already exists, use an `EtcdOpsRequest` of type `Restore`. It obeys the same rule rather than breaking it: the cluster is taken apart down to a single empty volume first, the snapshot is replayed into that, and the remaining members rejoin as new members. See [In-place Restore](/docs/guides/etcd/restore/overview.md).

### spec.monitor

Etcd managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box.

etcd exposes Prometheus metrics natively, so **no exporter sidecar container is deployed**. KubeDB creates a stats Service that targets etcd's own metrics endpoint and, for `agent: prometheus.io/operator`, a matching `ServiceMonitor`.

```yaml
spec:
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

### spec.configuration

`spec.configuration` is an optional field that specifies custom configuration for the etcd cluster. For etcd, only the typed `tuning` block is supported:

```yaml
spec:
  configuration:
    tuning:
      quotaBackendBytes: 8589934592
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
      snapshotCount: 10000
```

These map directly onto etcd's own command line flags:

| Field                     | etcd flag                    | Meaning                                                                                                            |
|---------------------------|------------------------------|--------------------------------------------------------------------------------------------------------------------|
| `quotaBackendBytes`       | `--quota-backend-bytes`      | Maximum size of the backend database. etcd raises a `NOSPACE` alarm and goes read-only once the backend grows past it. |
| `autoCompactionMode`      | `--auto-compaction-mode`     | How `autoCompactionRetention` is interpreted. Either `periodic` or `revision`.                                       |
| `autoCompactionRetention` | `--auto-compaction-retention`| A duration (e.g. `"1h"`) when the mode is `periodic`, a revision count (e.g. `"1000"`) when the mode is `revision`.   |
| `snapshotCount`           | `--snapshot-count`           | Number of committed transactions that trigger a snapshot to disk.                                                    |

> **`configuration.applyConfig` and `configuration.configSecret` are not supported for etcd.** etcd's `--config-file` is mutually exclusive with the individual command line flags that KubeDB must set to bootstrap and reconcile cluster membership, so a free-form config file would silently break cluster management. The `Reconfigure` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) webhook rejects both fields. The top-level `spec.configSecret` field exists only for structural parity with the other KubeDB databases and should not be used.

Because etcd has no live-reload path for these flags, applying a new `tuning` block always requires a restart of the members. The `Reconfigure` OpsRequest does that for you with a leader-aware rolling restart.

### spec.tls

`spec.tls` specifies the TLS/SSL configuration. The KubeDB operator supports TLS management by using [cert-manager](https://cert-manager.io/). When `spec.tls` is set, `spec.tls.issuerRef` is required.

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: etcd-issuer
    certificates:
      - alias: server
        privateKey:
          encoding: PKCS8
        secretName: etcd-server-cert
        subject:
          organizations:
            - kubedb
      - alias: client
        secretName: etcd-client-cert
      - alias: peer
        secretName: etcd-peer-cert
      - alias: metrics-exporter
        secretName: etcd-metrics-exporter-cert
```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - references the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for etcd.
  - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
  - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.

- `tls.certificates` - is an `optional` field that specifies a list of certificate configurations. It has the following fields:
  - `alias` - represents the identifier of the certificate. etcd supports the following aliases:
    - `server` - served by each member on the client (gRPC/HTTP) API port.
    - `client` - used by KubeDB itself (health checker, ops-manager) to talk to etcd.
    - `peer` - used for member-to-member Raft traffic. Peer traffic is **always mutually authenticated** once TLS is enabled; there is no server-only mode for peers.
    - `metrics-exporter` - accepted by the schema, but **inert**. etcd's `--listen-metrics-urls` has no TLS flags of its own, so KubeDB keeps the metrics listener on plain HTTP and never issues or mounts a certificate for this alias. Enabling `spec.tls` does not encrypt the metrics endpoint.

  - `secretName` - ( `string` | `"<database-name>-<alias>-cert"` ) - specifies the k8s secret name that holds the certificates.

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

  - `duration` ( `string` | `""` ) - is the period during which the certificate is valid.
  - `renewBefore` ( `string` | `""` ) - is a specifiable time before expiration duration.
  - `dnsNames` ( `[]string` | `nil` ) - is a list of subject alt names.
  - `ipAddresses` ( `[]string` | `nil` ) - is a list of IP addresses.
  - `uris` ( `[]string` | `nil` ) - is a list of URI Subject Alternative Names.
  - `emailAddresses` ( `[]string` | `nil` ) - is a list of email Subject Alternative Names.

Enabling TLS switches the scheme of every advertised client and peer URL from `http` to `https`, so it is applied with a rolling restart. Use the `ReconfigureTLS` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) to add, rotate or remove TLS on a running cluster.

### spec.podTemplate

KubeDB allows providing a template for the database pod through `spec.podTemplate`. The KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for the etcd cluster.

KubeDB accepts the following fields to set in `spec.podTemplate`:

- metadata:
  - annotations (pod's annotation)
  - labels (pod's labels)
- controller:
  - annotations (petset's annotation)
  - labels (petset's labels)
- spec:
  - containers
  - initContainers
  - volumes
  - imagePullSecrets
  - nodeSelector
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - podPlacementPolicy
  - priority
  - securityContext

> `args`, `env` and `resources` are **container**-level fields, not pod-level ones. Set them on the entry in `spec.podTemplate.spec.containers` whose `name` is `etcd`, as shown below.

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go).

The etcd pod runs a single operator-managed container, named **`etcd`**; use that name when overriding container-level settings such as `resources` or `args`. KubeDB adds no init container of its own — anything you put in `spec.podTemplate.spec.initContainers` is yours and is passed through as-is.

Uses of some fields of `spec.podTemplate` are described below.

#### spec.podTemplate.spec.containers[].resources

`resources` on the `etcd` container can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

```yaml
spec:
  podTemplate:
    spec:
      containers:
        - name: etcd
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              memory: 2Gi
```

#### spec.podTemplate.spec.containers[].args

`args` on the `etcd` container is an optional list of extra etcd command-line flags. It is the escape hatch for the etcd settings `spec.configuration.tuning` does not expose. It has to go on the container entry named `etcd` — there is no pod-level `args` field:

```yaml
spec:
  podTemplate:
    spec:
      containers:
        - name: etcd
          args:
            - --heartbeat-interval=250
            - --election-timeout=2500
```

These flags are appended **after** the ones the operator renders. etcd parses its command line with Go's standard `flag` package, where a repeated flag is not an error and the last occurrence wins — so a flag you set here also overrides one the operator set. That makes it powerful and sharp: overriding the flags that carry a member's identity, its listeners or its TLS material detaches the member from the cluster KubeDB is reconciling, and `--force-new-cluster` in particular would rewrite the data directory into a fresh single-member cluster on every restart. See [Extra etcd flags](/docs/guides/etcd/custom-configuration/using-config.md#extra-etcd-flags).

#### spec.podTemplate.spec.containers[].env

`env` on the `etcd` container is an optional list of environment variables to pass to the etcd image. Like `args` and `resources`, it is a container-level field — there is no pod-level `spec.podTemplate.spec.env`:

```yaml
spec:
  podTemplate:
    spec:
      containers:
        - name: etcd
          env:
            - name: ETCD_LOG_LEVEL
              value: debug
```

> The cluster bootstrap variables (`ETCD_NAME`, `ETCD_INITIAL_CLUSTER`, `ETCD_INITIAL_CLUSTER_STATE`, `ETCD_INITIAL_ADVERTISE_PEER_URLS`, and the listen/advertise URL variables) are computed by the operator and re-projected on every membership change. Do not override them.

#### spec.podTemplate.spec.imagePullSecrets

`KubeDB` provides the flexibility of deploying etcd from a private Docker registry. `spec.podTemplate.spec.imagePullSecrets` is a list of references to Secrets in the same namespace that hold the registry credentials.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the `podPlacementPolicy`. `name` of the podPlacementPolicy is referred under this attribute. This will be used by our PetSet controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes the Kubernetes affinity & podTopologySpreadConstraints features to do so.

```yaml
spec:
  podTemplate:
    spec:
      podPlacementPolicy:
        name: default
```

Spreading etcd members across failure domains is particularly valuable, since quorum survives only a minority of members failing at once.

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector).

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account whose name matches the Etcd crd name. Role and RoleBinding that provide the necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide the necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing the necessary access permissions manually.

### spec.serviceTemplates

You can also provide a template for the services created by the KubeDB operator for etcd through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows the following fields to be set in `spec.serviceTemplates`:

- `alias` represents the identifier of the service. For etcd it has the following possible values:
  - `primary` — the load balanced client Service (`<db-name>`) on port `2379`. Because every etcd member serves the client API and forwards writes to the Raft leader, one Service over all members is correct — there is no separate `standby` service for etcd.
  - `stats` — the metrics Service used by Prometheus.

- metadata:
  - annotations
  - labels
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

Besides these, the operator always creates a headless *governing* Service named `<db-name>-pods`, which publishes not-ready addresses and exposes both the client port `2379` and the peer port `2380`. The stable per-member DNS names derived from it (`<db-name>-0.<db-name>-pods.<namespace>.svc.<cluster-domain>`) are what etcd uses as its peer and client URLs. A learner has to be reachable by its peers *before* it becomes ready, which is why not-ready addresses are published.

### spec.archiver

`spec.archiver` opts the cluster into a scheduled snapshot backup driven by an [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md) policy object.

```yaml
spec:
  archiver:
    pause: false
    ref:
      name: etcd-archiver
      namespace: demo
```

- `spec.archiver.ref` is the name/namespace reference to the `EtcdArchiver` object.
- `spec.archiver.pause` temporarily stops the archiver-driven backup for this database.

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of the `Etcd` crd or which resources KubeDB should keep or delete when you delete the `Etcd` crd. KubeDB provides the following four deletion policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of the `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement the `DoNotTerminate` feature. If the admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as `spec.deletionPolicy` is set to `DoNotTerminate`.

The following table shows what KubeDB does when you delete an Etcd crd for different deletion policies,

| Behavior                            | DoNotTerminate |   Halt   |  Delete  | WipeOut  |
|-------------------------------------|:--------------:|:--------:|:--------:|:--------:|
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet                    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshot data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy`, KubeDB uses the `Delete` deletion policy by default.

> For more details you can visit [here](https://appscode.com/blog/post/deletion-policy/)

### spec.halted

Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.

### spec.healthChecker

It defines the attributes for the health checker.

- `spec.healthChecker.periodSeconds` specifies how often to perform the health check. Defaults to `10`.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out. Defaults to `10`.
- `spec.healthChecker.failureThreshold` specifies the minimum consecutive failures for the healthChecker to be considered failed. Defaults to `1`.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the write check or not.

For etcd, the health checker dials the client port of every member and uses etcd's own `Status()` and `MemberList()` RPCs. The cluster is considered quorum-healthy while at least `N/2+1` members respond. etcd is a key-value store, not a relational database, so there is no query-language based health probe.

Know details about KubeDB health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Etcd Status

`.status` describes the observed state of the `Etcd` object. It has the following fields.

### status.phase

`status.phase` is the overall phase of the database, derived from the conditions below:

| Phase           | Meaning                                                                                     |
|-----------------|-----------------------------------------------------------------------------------------------|
| `Provisioning`  | Offshoot resources are being created.                                                          |
| `Ready`         | All members are ready, quorum is healthy and the client endpoint accepts connections.          |
| `Critical`      | The client endpoint is reachable, but not every member is ready.                               |
| `NotReady`      | The client endpoint is not reachable.                                                          |
| `Halted`        | `spec.halted` is `true`.                                                                       |
| `Unknown`       | The phase could not be computed.                                                               |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation of the `Etcd` object observed by the KubeDB operator.

### status.conditions

`status.conditions` is an array describing the state of individual steps of provisioning and of the running cluster. Each entry carries a `type`, a `status` (`True` / `False` / `Unknown`), a `reason`, a human readable `message`, a `lastTransitionTime` and an `observedGeneration`.

| Type                  | Meaning                                                                    |
|-----------------------|-----------------------------------------------------------------------------|
| `ProvisioningStarted` | The operator has accepted the object and started creating its offshoots.    |
| `ReplicaReady`        | All etcd member pods are ready.                                             |
| `AcceptingConnection` | The client endpoint is accepting connection requests.                       |
| `Ready`               | The readiness check (quorum health) succeeded.                              |
| `Provisioned`         | The cluster has been successfully provisioned.                              |
| `SuccessfullyDataRestored` | The restore from a snapshot completed (bootstrap or in-place).          |

### status.authSecret

`status.authSecret` records the age of the credential currently referenced by `spec.authSecret`. Together with `spec.authSecret.rotateAfter`, this is what lets the Recommendation Engine propose a `RotateAuth` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) when the credential gets old.

## Next Steps

- Learn how to use KubeDB to run an etcd cluster [here](/docs/guides/etcd/README.md).
- Deploy your first etcd cluster with KubeDB by following the guide [here](/docs/guides/etcd/quickstart/quickstart.md).
- Detail concepts of [EtcdVersion object](/docs/guides/etcd/concepts/etcdversion.md).
- Detail concepts of [EtcdOpsRequest object](/docs/guides/etcd/concepts/etcdopsrequest.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
