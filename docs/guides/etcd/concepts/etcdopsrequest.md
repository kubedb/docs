---
title: EtcdOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: etcd-opsrequest-concepts
    name: EtcdOpsRequest
    parent: etcd-concepts-etcd
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# EtcdOpsRequest

## What is EtcdOpsRequest

`EtcdOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [etcd](https://etcd.io/) administrative operations like version updating, horizontal scaling, vertical scaling, TLS reconfiguration, and etcd's own maintenance RPCs (leader transfer, defragmentation, compaction) in a Kubernetes native way.

## EtcdOpsRequest CRD Specifications

Like any official Kubernetes resource, an `EtcdOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `EtcdOpsRequest` CRs for different administrative operations are given below:

**Sample `EtcdOpsRequest` for updating the database version:**

Let's assume that you have a KubeDB managed etcd cluster named `etcd-quickstart` running on your Kubernetes cluster with version `3.5.21`. Now, you can update its version to `3.6.4` using the following manifest.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-update-version
  namespace: demo
spec:
  databaseRef:
    name: etcd-quickstart
  type: UpdateVersion
  updateVersion:
    targetVersion: 3.6.4
```

**Sample `EtcdOpsRequest` for horizontal scaling of the cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: etcd-quickstart
  horizontalScaling:
    replicas: 5
```

**Sample `EtcdOpsRequest` for vertical scaling of the cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: etcd-quickstart
  verticalScaling:
    mode: Restart
    etcd:
      resources:
        requests:
          cpu: 600m
          memory: 1.2Gi
        limits:
          cpu: 1
          memory: 2Gi
```

**Sample `EtcdOpsRequest` for reconfiguring the cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-reconfigure
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: etcd-quickstart
  type: Reconfigure
  configuration:
    tuning:
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
      quotaBackendBytes: 8589934592
```

**Sample `EtcdOpsRequest` for volume expansion:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-online-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: etcd-quickstart
  volumeExpansion:
    mode: "Online"
    etcd: 3Gi
```

**Sample `EtcdOpsRequest` objects for reconfiguring TLS:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: etcd-quickstart
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: etcd-quickstart
  tls:
    issuerRef:
      name: etcd-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: server
        secretName: etcd-quickstart-server-cert
      - alias: client
        secretName: etcd-quickstart-client-cert
      - alias: peer
        secretName: etcd-quickstart-peer-cert
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: etcd-quickstart
  tls:
    remove: true
```

**Sample `EtcdOpsRequest` objects for the etcd-native maintenance operations:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-move-leader
  namespace: demo
spec:
  type: MoveLeader
  databaseRef:
    name: etcd-quickstart
  moveLeader:
    newLeader: etcd-quickstart-2
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-compact
  namespace: demo
spec:
  type: Compact
  databaseRef:
    name: etcd-quickstart
  compact: {}
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-defragment
  namespace: demo
spec:
  type: Defragment
  databaseRef:
    name: etcd-quickstart
  defragment: {}
```

Here, we are going to describe the various sections of an `EtcdOpsRequest` crd.

An `EtcdOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that points to the [Etcd](/docs/guides/etcd/concepts/etcd.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object. The referenced object must live in the same namespace as the `EtcdOpsRequest`.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. The following twelve types of operations are allowed in an `EtcdOpsRequest`.

| Type                | What it does                                                                                                                                 |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| `UpdateVersion`     | Updates the cluster to a different [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md), member by member.                                    |
| `HorizontalScaling` | Changes the number of etcd members, adding new members as learners and removing members through etcd's membership API.                            |
| `VerticalScaling`   | Changes the CPU/memory resources of the etcd container.                                                                                          |
| `VolumeExpansion`   | Grows the PVCs holding the etcd data directory.                                                                                                  |
| `Restart`           | Performs a leader-aware rolling restart of the members.                                                                                          |
| `Reconfigure`       | Applies a new `spec.configuration.tuning` block.                                                                                                 |
| `ReconfigureTLS`    | Adds, rotates or removes TLS for the client, peer and metrics endpoints.                                                                         |
| `RotateAuth`        | Rotates the credential of the etcd `root` user.                                                                                                  |
| `StorageMigration`  | Migrates the data PVCs onto a different StorageClass.                                                                                            |
| `MoveLeader`        | Hands the Raft leadership over to another member on purpose.                                                                                     |
| `Defragment`        | Runs etcd's `Defragment` RPC on every member, one at a time, to reclaim backend file space.                                                       |
| `Compact`           | Runs etcd's `Compact` RPC to drop superseded MVCC key revisions from the keyspace history.                                                        |

The last three have no analog in the other KubeDB databases. etcd is its own consensus layer — there is no external replication to repair, no "force failover" and no "reconnect standby" — but it does expose maintenance RPCs that a cluster operator is expected to call, and those are what these ops request types wrap. See [spec.moveLeader](#specmoveleader), [spec.defragment](#specdefragment) and [spec.compact](#speccompact) below.

> You can perform only one type of operation in a single `EtcdOpsRequest` CR. For example, if you want to update the version of your database and scale up its members, you have to create two separate `EtcdOpsRequest` objects. First create the one for updating. Once it is completed, then create the other one for scaling.

### spec.updateVersion

If you want to update your etcd version, you have to specify the `spec.updateVersion` section. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to an [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) CR that contains the etcd version information you want to update to.

The target version is validated against the current version's `spec.updateConstraints`, which encode etcd's own upgrade rules: no downgrades, no major version jumps and at most one minor version step at a time. So going from `3.5.x` to `3.7.x` requires two ops requests, with a stop at `3.6.x`.

### spec.horizontalScaling

If you want to scale up or scale down the number of members in your etcd cluster, you have to specify the `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.replicas` indicates the desired number of etcd members after scaling.

This ops request is deliberately a thin wrapper: it patches `spec.replicas` on the `Etcd` object and then waits for the provisioner to converge. The provisioner is the component that continuously reconciles the PetSet replica count against etcd's actual member list — adding the next ordinal as a **learner**, waiting for it to catch up with the leader's Raft revision, then **promoting** it, and on the way down removing the **highest ordinal** member through `MemberRemove` before shrinking the PetSet.

> Prefer odd member counts. A `2n+1` member cluster tolerates `n` failures; growing to an even count only raises the quorum size without improving fault tolerance.

### spec.verticalScaling

`spec.verticalScaling` specifies the information about the `Etcd` resources like `cpu` and `memory` that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.etcd` indicates the desired resources for the `etcd` container. It only has a `resources` sub-field — etcd has a single container role, so there is no node-selection or topology sub-field here.
- `spec.verticalScaling.exporter` exists for structural parity with the other KubeDB databases. **For etcd there is no exporter sidecar container to resize** — etcd serves its Prometheus metrics natively — so setting this field has no container to act on.
- `spec.verticalScaling.mode` specifies how the scaling is actuated. `Restart` (the default) applies the new resources by restarting the pods, while `InPlace` resizes the running pods in place via the Kubernetes `pods/resize` subresource (no restart), automatically falling back to `Restart` for any pod whose node cannot fit the new resources.

The `resources` field has the below structure:

```yaml
requests:
  memory: "600Mi"
  cpu: "0.5"
limits:
  memory: "800Mi"
  cpu: "0.8"
```

Here, when you specify the resource request, the scheduler uses this information to decide which node to place the container of the pod on, and when you specify a resource limit for the container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can find more details [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

### spec.volumeExpansion

> To use the volume expansion feature the storage class must support volume expansion.

If you want to expand the volume of your etcd cluster, you have to specify the `spec.volumeExpansion` section. This field consists of the following sub-fields:

- `spec.volumeExpansion.etcd` indicates the desired size for the persistent volume holding the etcd data directory. It refers to the Kubernetes `Quantity` type.
- `spec.volumeExpansion.mode` indicates the mode of the volume expansion. It can be `Online` or `Offline`, based on what the storage class supports.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    mode: "Online"
    etcd: "2Gi"
```

This will expand the volume size of all the etcd members to 2 GB.

### spec.configuration

If you want to reconfigure a running etcd cluster with a new tuning configuration, you have to specify the `spec.configuration` section:

```yaml
spec:
  configuration:
    tuning:
      quotaBackendBytes: 8589934592
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
      snapshotCount: 10000
```

- `spec.configuration.tuning` holds the KubeDB-managed etcd tuning knobs described in [Etcd CRD](/docs/guides/etcd/concepts/etcd.md#specconfiguration). They end up on the etcd command line.

> The generic `configuration.configSecret` and `configuration.applyConfig` fields are **rejected by the webhook for etcd**. etcd's `--config-file` is mutually exclusive with the individual flags KubeDB has to set to bootstrap and reconcile membership, so a free-form config file cannot be supported. Only the typed `tuning` block is available.

etcd has no live-reload path for these flags — nothing analogous to Postgres' `pg_reload_conf()` — so a `Reconfigure` ops request always ends with a leader-aware rolling restart of the members.

### spec.tls

If you want to reconfigure the TLS configuration of your etcd cluster — add TLS, remove TLS, update the issuer or the certificates, or rotate the certificates — you have to specify the `spec.tls` section. It consists of the following sub-fields:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/etcd/concepts/etcd.md#spectls). The valid aliases are `server`, `client`, `peer` and `metrics-exporter`.
- `spec.tls.rotateCertificates` tells the operator to rotate the certificates of this database.
- `spec.tls.remove` tells the operator to remove TLS from this database.

Exactly one of the three operations — issue/update (`issuerRef`/`certificates`), `rotateCertificates`, or `remove` — should be requested per ops request. Each of them changes the scheme of the advertised peer and client URLs or the certificate material on disk, so the operator follows up with the same leader-aware rolling restart used by `Restart`.

### spec.authentication

If you want to rotate the credential of the etcd `root` user, use `spec.type: RotateAuth`. This field consists of the following sub-field:

- `spec.authentication.secretRef` optionally points to a Secret holding the new credential. If it is omitted, ops-manager generates a random password.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: etcd-quickstart
  authentication:
    secretRef:
      kind: Secret
      name: etcd-new-auth
```

Unlike most KubeDB databases, `RotateAuth` on etcd **does not restart the pods**. etcd applies an RBAC password change live through its `UserChangePassword` RPC, so the operator stages the new credential on the auth Secret (using `.prev` / `.next` keys) and promotes it in place.

### spec.restart

`spec.type: Restart` performs a rolling restart of the members. `spec.restart` itself carries no fields:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: etcd-quickstart
  restart: {}
```

The restart is **leader-aware**: if the pod about to be restarted is the current Raft leader, the operator first moves leadership off it, then evicts and recreates pods one at a time, waiting for each pod to become Ready and for the cluster to be quorum-healthy again before moving on to the next one.

### spec.migration

If you want to move the data PVCs onto a different StorageClass, use `spec.type: StorageMigration` together with the `spec.migration` section:

- `spec.migration.storageClassName` is the desired StorageClass to migrate the PVCs to.
- `spec.migration.oldPVReclaimPolicy` controls the reclaim policy applied to the previous PersistentVolume after the PVC has been re-pointed. Set it to `Retain` to keep the old PV around.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: etcd-quickstart
  migration:
    storageClassName: fast-ssd
  timeout: 30m
```

> `spec.timeout` is required for a `StorageMigration` ops request, since copying the data directory of each member can take an arbitrary amount of time.

### spec.moveLeader

`spec.type: MoveLeader` wraps etcd's `MoveLeader` RPC, which hands Raft leadership from the current leader to another voting member.

- `spec.moveLeader.newLeader` is the optional name of the member (pod) that should become the new leader. If it is empty, the operator picks a healthy member that is caught up with the leader.

```yaml
spec:
  type: MoveLeader
  databaseRef:
    name: etcd-quickstart
  moveLeader:
    newLeader: etcd-quickstart-2
```

Why it exists: etcd elects its own leader, and an unplanned leader loss costs an election round during which writes stall. Moving leadership deliberately — before draining the node that hosts the leader, before taking that member down for maintenance, or before a rolling operation — turns an unplanned election into a fast, controlled handover. The same primitive is what makes the `Restart`, `Reconfigure` and `UpdateVersion` rollouts leader-aware.

### spec.defragment

`spec.type: Defragment` wraps etcd's `Defragment` RPC. `spec.defragment` carries no fields:

```yaml
spec:
  type: Defragment
  databaseRef:
    name: etcd-quickstart
  defragment: {}
```

Why it exists: etcd's backend never shrinks its database file on its own. Deleting keys, or compacting away old revisions, frees space *inside* the file but leaves the file itself as large as it ever was — and that size is what counts against `quotaBackendBytes`. Defragmentation rewrites the backend to release that space back to the filesystem.

Defragmenting briefly blocks the member it runs on, so the operator walks the members **one at a time and leaves the leader for last**, checking between each step that the cluster still has quorum.

### spec.compact

`spec.type: Compact` wraps etcd's `Compact` RPC, which discards superseded revisions from the MVCC keyspace history.

- `spec.compact.revision` is the optional revision to compact the history up to. If it is omitted, the operator compacts up to the current revision at execution time.

```yaml
spec:
  type: Compact
  databaseRef:
    name: etcd-quickstart
  compact:
    revision: 1000000
```

Why it exists: etcd keeps every past version of every key so that watchers and historical reads work. Left alone, that history grows without bound and eventually trips the backend quota. Compaction drops the old versions; **defragmentation is what actually returns the freed space to the filesystem**, so the two are usually run as a pair — `Compact` first, then `Defragment`.

If you would rather have this happen automatically, set `spec.configuration.tuning.autoCompactionMode` and `autoCompactionRetention` on the [Etcd](/docs/guides/etcd/concepts/etcd.md#specconfiguration) object instead, and use this ops request only for one-off cleanups.

### spec.timeout

As we internally retry the ops request steps multiple times, this `timeout` field helps the users specify the timeout for those steps of the ops request. If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply

This field controls the execution of the ops request depending on the database state. It has two supported values: `Always` & `IfReady`. Use `IfReady` if you want to process the ops request only when the database is Ready, and use `Always` if you want to process the ops request irrespective of the database state. It defaults to `IfReady`.

### spec.maxRetries

`spec.maxRetries` is the number of times a failed ops request will be retried before it is finally marked as `Failed`. It defaults to `1`.

## EtcdOpsRequest `Status`

`.status` describes the current state and progress of an `EtcdOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `EtcdOpsRequest`. It can have the following values:

| Phase       | Meaning                                                                        |
|-------------|----------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the EtcdOpsRequest  |
| Progressing | KubeDB has started the execution of the applied EtcdOpsRequest                   |
| Failed      | KubeDB has failed the operation requested in the EtcdOpsRequest                  |
| Denied      | KubeDB has denied the operation requested in the EtcdOpsRequest                  |
| Skipped     | KubeDB has skipped the operation requested in the EtcdOpsRequest                 |

Important: The ops-manager operator can skip an ops request only if its execution has not been started yet & there is a newer ops request applied in the cluster. `spec.type` has to be the same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `EtcdOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `EtcdOpsRequest` processing. Each condition entry has the following fields:

- `type` specifies the type of the condition. `EtcdOpsRequest` uses the shared KubeDB condition types plus a handful of etcd-specific ones:

| Type                       | Meaning                                                                        |
|----------------------------|----------------------------------------------------------------------------------|
| `Running`                  | Set once, when the operation starts being executed (`status.phase` moves to `Progressing`, but the condition `type` written for this step is `Running`, not `Progressing`). |
| `Successful`               | The operation on the database was successful.                                    |
| `Failed`                   | The operation on the database failed.                                            |
| `DatabasePauseSucceeded`   | The database is paused by the operator.                                          |
| `ResumeDatabase`           | The database is resumed by the operator.                                         |
| `UpdateEtcdPetSet`         | The PetSet of the etcd cluster has been updated.                                 |
| `UpdateEtcdNodePVCs`       | The PVCs of the etcd members have been updated.                                  |
| `MigrateEtcdStorage`       | The etcd data has been migrated onto the new StorageClass.                       |
| `RestartEtcdPods`          | The etcd member pods have been restarted.                                        |
| `ReadyEtcdPod`             | An etcd member pod has become ready.                                             |
| `EtcdClusterHealthy`       | The etcd cluster is quorum-healthy.                                              |
| `EtcdMemberListReady`      | etcd's member list matches the desired membership.                               |
| `EtcdLearnerPromoted`      | A learner has caught up and has been promoted to a voting member. (`EtcdMemberAdded`/`EtcdMemberRemoved` are declared in the API but not currently set by the operator — the learner-add and member-remove steps are only observable indirectly, via `EtcdMemberListReady` and the provisioner's own logs, not a dedicated condition.) |
| `EtcdLeaderMoved`          | Raft leadership has been transferred to another member. Only set by the standalone `MoveLeader` ops type; a `Restart` that has to move leadership off the pod being evicted instead records a `MoveLeader--<pod-name>` condition. |
| `EtcdDefragmented`         | The backend of the members has been defragmented.                                |
| `EtcdCompacted`            | The keyspace history has been compacted.                                         |
| `EtcdAlarmCleared`         | An etcd alarm (for example `NOSPACE`) has been cleared.                          |
| `IssueCertificatesSucceeded` | The TLS certificate issuing is successful.                                     |
| `UpdateDatabase`           | The `Etcd` CR has been updated.                                                  |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
  - `status` will be `True` if the current transition succeeded.
  - `status` will be `False` if the current transition failed.
  - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.

## Next Steps

- Learn about the [Etcd](/docs/guides/etcd/concepts/etcd.md) crd.
- Learn about the [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) crd.
- Deploy your first etcd cluster with KubeDB by following the guide [here](/docs/guides/etcd/quickstart/quickstart.md).
