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

`EtcdOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [etcd](https://etcd.io/) administrative operations like version updating, horizontal scaling, vertical scaling, TLS reconfiguration, etcd's own maintenance RPCs (leader transfer, defragmentation, compaction), and the two disaster-recovery procedures (rebuilding a cluster after a permanent quorum loss, and restoring a snapshot into an existing cluster) in a Kubernetes native way.

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

`spec.type` specifies the kind of operation that will be applied to the database. The following fourteen types of operations are allowed in an `EtcdOpsRequest`.

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
| `RecoverFromQuorumLoss` | **Destructive.** Rebuilds a cluster that has permanently lost its Raft quorum from a single surviving member, discarding every other member's data. |
| `Restore`           | **Destructive.** Replaces the entire keyspace of an existing cluster with a KubeStash snapshot.                                                   |

The last two are break-glass, disaster-recovery procedures, not routine maintenance. Both throw data away as their *normal* outcome, both require `spec.timeout`, and both leave the cluster as a healthy single member that the ordinary membership reconciliation then regrows back to `spec.replicas`. See [spec.recoverFromQuorumLoss](#specrecoverfromquorumloss) and [spec.restore](#specrestore) below.

`MoveLeader`, `Defragment` and `Compact` have no analog in the other KubeDB databases. etcd is its own consensus layer — there is no external replication to repair, no "force failover" and no "reconnect standby" — but it does expose maintenance RPCs that a cluster operator is expected to call, and those are what these ops request types wrap. See [spec.moveLeader](#specmoveleader), [spec.defragment](#specdefragment) and [spec.compact](#speccompact) below.

> You can perform only one type of operation in a single `EtcdOpsRequest` CR. For example, if you want to update the version of your database and scale up its members, you have to create two separate `EtcdOpsRequest` objects. First create the one for updating. Once it is completed, then create the other one for scaling.

### spec.updateVersion

If you want to update your etcd version, you have to specify the `spec.updateVersion` section. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to an [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) CR that contains the etcd version information you want to update to.

The target version is validated against etcd's own upgrade rules: no downgrades, no major version jumps and at most one minor version step at a time. So going from `3.5.x` to `3.7.x` requires two ops requests, with a stop at `3.6.x`. A target whose `EtcdVersion` is marked `spec.deprecated` is refused as well.

> These rules are applied by the ops-manager when the request starts executing — they are hard-coded there, not read from the `EtcdVersion`'s `spec.updateConstraints.allowlist`. The admission webhook only checks that `spec.updateVersion.targetVersion` is non-empty, so an invalid upgrade path produces a request that is accepted and then fails. See [The Upgrade Path Is Validated](/docs/guides/etcd/update-version/overview.md#the-upgrade-path-is-validated--read-this-first).

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

> The generic `configuration.applyConfig` field is **rejected by the webhook for etcd**, and `configuration.configSecret` is accepted but inert — nothing ever mounts it. etcd's `--config-file` is mutually exclusive with the individual flags KubeDB has to set to bootstrap and reconcile membership, so a free-form config file cannot be supported. In practice, only the typed `tuning` block is available.

etcd has no live-reload path for these flags — nothing analogous to Postgres' `pg_reload_conf()` — so a `Reconfigure` ops request normally ends with a leader-aware rolling restart of the members. You can suppress that with `spec.configuration.restart: "false"`, but the new values will then not take effect until the members are restarted for some other reason.

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

### spec.recoverFromQuorumLoss

`spec.type: RecoverFromQuorumLoss` rebuilds an etcd cluster that has **permanently** lost its Raft quorum — a majority of its voting members is gone for good — from a single surviving member, using etcd's own `--force-new-cluster` procedure.

> **This is a destructive, break-glass procedure.** Every member except the confirmed survivor has its data directory discarded permanently, and any write the lost majority had committed but the survivor had not yet applied is lost with no way to get it back. It is **never** triggered automatically: the Provisioner operator only ever raises a `QuorumLost` condition on the `Etcd` object as a signal, and the recovery itself has to be asked for by a human-authored ops request. Read the [Recover from Quorum Loss Overview](/docs/guides/etcd/recover-from-quorum-loss/overview.md) before using it.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-recover-quorum
  namespace: demo
spec:
  type: RecoverFromQuorumLoss
  databaseRef:
    name: etcd-quickstart
  recoverFromQuorumLoss: {}
  timeout: 30m
  apply: Always
```

You then read the resolved survivor out of the status and patch `confirmMember` to match it:

```yaml
spec:
  recoverFromQuorumLoss:
    confirmMember: etcd-quickstart-1
```

- `spec.recoverFromQuorumLoss.member` is the **optional** pod name — or a bare ordinal, e.g. `"1"` — of the surviving member to rebuild the cluster from. If it is empty, the operator picks the reachable member with the **highest Raft applied index**, which is the survivor that has lost the fewest writes.
- `spec.recoverFromQuorumLoss.confirmMember` is a **mandatory hard confirmation gate**. The operator resolves the survivor exactly once, records it in an `EtcdQuorumLossSurvivorResolved--<pod>` condition, reports it in the `EtcdQuorumLossAwaitingConfirmation` condition, and then refuses to do anything destructive until this field names *exactly* that pod. It waits there indefinitely — the wait never times out and never fails the request, because what it is waiting for is a person. This exists so a mistyped `member` cannot silently destroy the wrong member's data, which is why the gate applies even when you set `member` yourself.

The intended workflow is therefore two steps: create the request with `confirmMember` unset, read the resolved survivor back out of the status, then patch `confirmMember` to match it.

The request is **refused outright** — terminally, without burning `spec.maxRetries` — unless the target `Etcd` currently reports the `QuorumLost` condition with status `True`. The condition is read live from the API server on every admission, and a cluster that has regained its quorum on its own while the request waited for confirmation is refused as well.

> `spec.timeout` is required for a `RecoverFromQuorumLoss` ops request. Several steps — waiting for the rebuilt member to come up, waiting for a volume to rebind — have no other deadline at all. `spec.storageType` must also be `Durable`: an ephemeral member's data died with its pod, so there is nothing to rescue.

Because the cluster it is run against is by definition not `Ready`, this ops request generally needs `spec.apply: Always`. It ends at a healthy **single-member** cluster; the members that were discarded are added back afterwards by the ordinary membership reconciliation, as learners, through the same path an ordinary scale up uses.

See the [Recover from Quorum Loss guide](/docs/guides/etcd/recover-from-quorum-loss/recover-from-quorum-loss.md) for the full walkthrough.

### spec.restore

`spec.type: Restore` restores a [KubeStash](https://kubestash.com/) snapshot into an `Etcd` database that **already exists** and may already have running members.

> **This replaces the entire keyspace of the database.** Everything currently stored in it is discarded and replaced by the contents of the snapshot. Full data loss is not a failure mode here — it is the normal, successful outcome of the operation, and there is no undo.

This is the in-place counterpart of the **bootstrap-time** restore configured through `spec.init.archiver` on the [Etcd](/docs/guides/etcd/concepts/etcd.md#specinit) object itself, which only works while a brand-new `Etcd` object is being created. If you are standing up a new cluster from a backup, use that one instead — see [Snapshot Backup & Restore](/docs/guides/etcd/backup/kubestash/snapshot/index.md#restore).

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-restore
  namespace: demo
spec:
  type: Restore
  databaseRef:
    name: etcd-quickstart
  restore:
    fullDBRepository:
      name: etcd-full-repo
      namespace: demo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
    recoveryTimestamp: "2026-08-15T08:30:00Z"
  timeout: 30m
  apply: Always
```

- `spec.restore.fullDBRepository` names the KubeStash `Repository` to restore the etcd snapshot from. It is **required and has no default** — unlike the bootstrap-time restore this destroys data that is already there, so guessing which repository was meant is not an option.
- `spec.restore.recoveryTimestamp` is optional. If it is omitted, the **latest** available snapshot is used; otherwise the operator picks the newest successful full snapshot taken *at or before* that instant. As with the bootstrap restore, this is snapshot selection rather than point-in-time recovery — etcd has no log stream to replay forward.
- `spec.restore.encryptionSecret` refers to the Secret holding the encryption key the snapshot was backed up with, if any.

Mechanically, **only the seed member (ordinal 0) is ever restored into**. The cluster is taken apart down to a single empty volume, the snapshot is written into it by the same `RestoreSession` the bootstrap path uses, and the seed is started on the restored data directory. The other members are discarded and rejoin as fresh **learners**, streaming the restored keyspace from the seed through the same learner-add/promote path an ordinary [horizontal scale up](#spechorizontalscaling) uses. Their old data directories describe a keyspace the restored one no longer relates to, so they could not have rejoined anyway.

> `spec.timeout` is required for a `Restore` ops request — the `RestoreSession` has no other deadline. `spec.storageType` must be `Durable` with `spec.storage` set: there is no PVC to restore into before the seed member starts otherwise.

Because the cluster it is run against is often degraded, this ops request generally needs `spec.apply: Always`.

See the [In-place Restore guide](/docs/guides/etcd/restore/restore.md) for the full walkthrough.

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
| `EtcdClusterHealthy`       | The etcd cluster is quorum-healthy. Written per pod as `EtcdClusterHealthy--<pod-name>` by the operations that walk the members one at a time (`Restart`, `Reconfigure`, `VerticalScaling`, `UpdateVersion`, `StorageMigration`, `Defragment`), and bare by `HorizontalScaling`. |
| `EtcdMemberListReady`      | etcd's member list matches the desired membership.                               |
| `EtcdMemberAdded`          | Recorded at the start of a scale-up, before `spec.replicas` is patched.          |
| `EtcdMemberRemoved`        | Recorded at the start of a scale-down, before `spec.replicas` is patched.        |
| `EtcdLearnerPromoted`      | A learner has caught up and has been promoted to a voting member. Also set on a scale-down, where the check simply confirms no member is still a learner. |
| `EtcdLeaderMoved`          | Raft leadership has been transferred to another member. Only set by the standalone `MoveLeader` ops type; a `Restart` that has to move leadership off the pod being evicted instead records a `MoveLeader--<pod-name>` condition. |
| `EtcdDefragmented`         | The backend of the members has been defragmented.                                |
| `EtcdCompacted`            | The keyspace history has been compacted.                                         |
| `EtcdAlarmCleared`         | An etcd alarm (for example `NOSPACE`) has been cleared.                          |
| `IssueCertificatesSucceeded` | The TLS certificate issuing is successful.                                     |
| `UpdateDatabase`           | The `Etcd` CR has been updated.                                                  |

A `RecoverFromQuorumLoss` request additionally reports its own step conditions, in this order:

| Type                                      | Meaning                                                                    |
|-------------------------------------------|------------------------------------------------------------------------------|
| `EtcdQuorumLossAdmitted`                  | The `Etcd` was observed reporting `QuorumLost=True`, so the request is allowed to run. |
| `EtcdQuorumLossSurvivorResolved--<pod>`   | The member the cluster will be rebuilt from has been chosen. Resolved **once** and never re-picked. |
| `EtcdQuorumLossAwaitingConfirmation`      | `False` while `spec.recoverFromQuorumLoss.confirmMember` does not name the resolved survivor — its message tells you what to set it to. `True` once it does. |
| `EtcdQuorumLossPreflight`                 | The last checks before anything is destroyed: the quorum is still lost, and ordinal 0's pod exists. |
| `EtcdQuorumLossPVCRelocated`              | The survivor's volume now answers to ordinal 0's claim name.                 |
| `EtcdQuorumLossStaleMembersRemoved`       | Every member outside the survivor has been discarded.                        |
| `EtcdQuorumLossClusterStateSeeded`        | The cluster state ConfigMap has been rewritten for a single-member cluster.  |
| `EtcdQuorumLossForceNewClusterBoot`       | Ordinal 0 has been recreated with `--force-new-cluster`.                     |
| `EtcdQuorumLossSingleMemberHealthy`       | The rebuilt member is Ready and reports exactly one member.                  |
| `EtcdQuorumLossFlagRemoved`               | Ordinal 0 has been recreated without `--force-new-cluster`, so a later restart cannot re-trigger it. |
| `EtcdQuorumLossVolumesReclaimed`          | The parked reclaim policies have been restored — this is what finally destroys the discarded members' data. |
| `EtcdQuorumLossPetSetRestored`            | The PetSet has been recreated around the rebuilt member.                     |
| `RecoverEtcdFromQuorumLoss`               | The whole recovery is done.                                                  |

A `Restore` request reports these:

| Type                              | Meaning                                                                            |
|-----------------------------------|--------------------------------------------------------------------------------------|
| `EtcdRestorePetSetDeleted`        | The PetSet has been orphaned so its pods and claims can be replaced.                 |
| `EtcdRestoreMembersDiscarded`     | Every member outside the seed has been discarded. Completed per member first, as `EtcdRestoreMembersDiscarded--<pod>`. |
| `EtcdRestoreSeedPVCWiped`         | The seed member's data volume has been replaced with an empty PVC.                   |
| `EtcdRestoreSessionCreated`       | The KubeStash `RestoreSession` has been created — this is what pins the chosen snapshot. |
| `EtcdRestoreSnapshotApplied`      | The session finished writing the snapshot into the seed volume.                      |
| `EtcdRestoreClusterStateSeeded`   | The cluster state ConfigMap has been rewritten for a single-member cluster.          |
| `EtcdRestorePetSetRestored`       | The PetSet has been recreated, bringing the seed back on the restored volume.        |
| `EtcdRestoreSingleMemberHealthy`  | The seed member is Ready and serving the restored keyspace.                          |
| `EtcdRestoreVolumesReclaimed`     | The parked reclaim policies have been restored — this is what finally destroys the replaced data. |
| `RestoreEtcdSnapshot`             | The whole restore is done.                                                           |

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
