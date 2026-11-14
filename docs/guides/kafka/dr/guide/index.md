---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: kf-dr-guide-kafka
    name: User Guide
    parent: kf-dr-kafka
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running Kafka in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed Kafka in cross data center
disaster recovery (DC-DR) mode: the components, the naming contract, deployment, what
the operator creates, producing to the active endpoint, consumers resuming after a
flip, monitoring MM2 lag, the produce fence, switchover, failback, scaling, and day-2
operations.

Read the [DC-DR Overview](/docs/guides/kafka/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/kafka/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the active write DC. This is the single cross-DC failover authority. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease for its DC and projects the Lease decision into the local spoke as the primary-dc marker the produce fence reads. |
| **KubeDB Kafka operator (hub)** | the OCM hub | Expands the `Kafka` CR into per-DC Kafka clusters, per-DC `ConnectCluster` objects, and the MM2 `Connector` objects on the standby. On a Lease change it flips the bootstrap endpoint, reverses the MM2 direction, and moves the produce fence, then writes `status.disasterRecovery`. |
| **Per-DC Kafka clusters** | each Member DC | Each is a self-contained Kafka with its own intra-DC KRaft quorum, brokers, and per-partition ISR. KRaft never crosses the DC boundary. |
| **Per-DC `ConnectCluster` + MM2 connectors** | each Member DC | One `ConnectCluster` per DC (`kafkaRef` to its local Kafka). The three MM2 connectors run only on the current standby's `ConnectCluster`, mirroring active to standby. |
| **The produce fence** | each Member DC | Denies producer writes on a non-active cluster, fail closed, so only the Lease holder takes writes. |
| **KubeSlice (or external listeners)** | each spoke | Provides the flat cross-DC pod network so MM2 can reach the remote cluster and the bootstrap endpoint resolves across DCs. |

## The DC-name contract

One string identifies a data center everywhere. **Keep these identical:**

- the OCM spoke cluster name
- the agent `--dc-name`
- the primary-DC Lease `holderIdentity`
- the marker `data.activeDC`
- the pod label `open-cluster-management.io/cluster-name`
- the `PlacementPolicy` `distributionRule.clusterName`

## Deploying

### PlacementPolicy

Map the global pod ordinals to data centers and tag each DC with its role:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: kf-dcdr
spec:
  clusterSpreadConstraint:
    slice:
      projectNamespace: kubeslice-demo
      sliceName: demo-slice
    failoverPolicy:
      mode: TwoDC          # two Member DCs plus an Arbiter DC
      trigger:
        scope: Global      # one cluster-wide failover scope
    distributionRules:
    - clusterName: dc-a
      role: Member
      replicaIndices: [0, 1, 2]
    - clusterName: dc-b
      role: Member
      replicaIndices: [3, 4, 5]
    - clusterName: dc-c
      role: Arbiter
      replicaIndices: []
```

- A data-bearing **Member** rule carries `replicaIndices` that map to a self-contained
  Kafka cluster with its own KRaft quorum. The **Arbiter** DC carries an empty
  `replicaIndices` and holds only the `dr-controlplane` etcd member, never Kafka.
- `mode: TwoDC` expects exactly two Member DCs plus the Arbiter DC. Three or more data
  DCs is a separate design and out of scope.
- Roles are `Member` and `Arbiter` only.

### Kafka

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kf-dcdr
  namespace: demo
spec:
  version: 4.0.0
  distributed: true
  replicas: 6
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: kf-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

### What the operator creates

- **One self-contained Kafka cluster per Member DC** (`kf-dcdr` materialized into a
  cluster in `dc-a` and a cluster in `dc-b`), each with its own KRaft controller
  quorum and its own per-partition ISR. The two clusters never share a KRaft quorum.
  `spec.replicas: 6` is the total broker count across both Member DCs; the
  `replicaIndices` split it into a three-node cluster per DC (ordinals 0, 1, 2 in `dc-a`
  and 3, 4, 5 in `dc-b`).
- **One `ConnectCluster` per DC**, each `kafkaRef` pointing at that DC's local Kafka,
  because Connect's internal config, offset, and status topics must live on the
  cluster that is the mirror target. The reverse-direction `ConnectCluster` stays
  provisioned with its connectors disabled until a failover needs it.
- **The three MM2 `Connector` objects on the current standby's `ConnectCluster`**: a
  `MirrorSourceConnector`, a `MirrorCheckpointConnector`, and a
  `MirrorHeartbeatConnector`, mirroring the active cluster into the standby with
  `IdentityReplicationPolicy`.
- **The Lease-gated bootstrap Service** that resolves to the active cluster's brokers.

The MM2 connectors are ordinary KubeDB `Connector` objects. Conceptually the
`ConnectCluster` on the standby DC (`dc-b` while `dc-a` is active) looks like:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: ConnectCluster
metadata:
  name: kf-dcdr-dc-b-mm2
  namespace: demo
spec:
  version: 4.0.0
  replicas: 3
  kafkaRef:
    name: kf-dcdr-dc-b     # the local (standby) Kafka: consume from remote, produce to local
    namespace: demo
  deletionPolicy: WipeOut
```

and the `MirrorSourceConnector` carries the mirror direction and the DR-critical
properties in its config secret:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: Connector
metadata:
  name: kf-dcdr-mirror-source
  namespace: demo
spec:
  configuration:
    secretName: kf-dcdr-mirror-source-config
  connectClusterRef:
    name: kf-dcdr-dc-b-mm2
    namespace: demo
  deletionPolicy: WipeOut
```

with the connector's `config.properties` (in the referenced secret) setting, among
others:

```properties
connector.class=org.apache.kafka.connect.mirror.MirrorSourceConnector
source.cluster.alias=dc-a
target.cluster.alias=dc-b
source.cluster.bootstrap.servers=kf-dcdr-dc-a-pods.demo.svc:9092
target.cluster.bootstrap.servers=kf-dcdr-dc-b-pods.demo.svc:9092
replication.policy.class=org.apache.kafka.connect.mirror.IdentityReplicationPolicy
sync.topic.acls.enabled=false
```

The operator owns these connector configs so the mirror direction stays
operator-controlled and the two directions never overlap.

## Connecting and producing

A DC-DR Kafka exposes one user-facing **bootstrap Service** that resolves to the
active cluster's brokers. Producers and consumers always connect to that single
endpoint and reach the write cluster without reconfiguration. Because MM2 uses
`IdentityReplicationPolicy`, topic names are identical on both clusters, so after the
endpoint flips clients keep using the same topics.

```bash
# Produce to the active cluster through the single bootstrap endpoint:
$ kubectl exec -n demo kf-dcdr-dc-a-0 -- \
    kafka-console-producer.sh --bootstrap-server kf-dcdr-pods.demo.svc:9092 --topic orders
```

Only the active cluster accepts producer writes. If clients somehow reach a standby
cluster, its produce fence rejects the write (see below), which is the split-brain
guard.

### Consumers resume after a flip

The `MirrorCheckpointConnector` runs with `sync.group.offsets.enabled=true`, so it
translates consumer-group offsets from the active cluster into the standby cluster's
offset space. After a failover or switchover, a consumer group reconnecting through the
flipped endpoint resumes from the translated offset on the new active cluster rather
than re-reading from the beginning or skipping ahead.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC that holds the Lease and takes producer writes. |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | `Member` or `Arbiter`. |
| `dataCenters[].writable` | True only for the active cluster. |
| `dataCenters[].brokersReady` | Ready broker count in that DC's cluster. |
| `dataCenters[].mirrorLagMillis` | The standby's MM2 replication latency behind the active, in milliseconds. |
| `dataCenters[].healthy` | DC health: a Member DC is healthy when its brokers are ready; the Arbiter DC is healthy when its `dr-controlplane` etcd member is reachable (so an Arbiter reports `healthy: true` with `brokersReady: 0`). |

### MM2 lag

Cross-DC lag comes from MM2's own metrics: `replication-latency-ms` and
`record-age-ms` in the `kafka.connect.mirror` metric group, or the offset gap between
source and target derived from the heartbeat and checkpoint topics. The hub surfaces
this into `mirrorLagMillis`; there is no lag field in the base `KafkaStatus`.

### Useful checks

```bash
# Which DC the Lease intends as active (from the coordination plane):
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# Per-DC brokers and roles:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=kf-dcdr \
    -L kubedb.com/role,open-cluster-management.io/cluster-name

# MM2 connector status on the standby's ConnectCluster:
$ kubectl get connector -n demo -l app.kubernetes.io/instance=kf-dcdr
```

## The produce fence

A non-active cluster must refuse producer writes, fail closed: a cluster that cannot
confirm it holds the Lease denies produce. There are two fence mechanisms.

- **ACL fence (authorization on):** revoke produce ACLs for the CLIENT principals on
  the non-active cluster. This requires Kafka authorization enabled (the
  `StandardAuthorizer`, set only when security is on).
- **Listener-gate fence (default posture):** gate the client listener on the non-active
  cluster so producers cannot connect. The shipped examples default to
  `disableSecurity: true`, and a cluster with no authorization has no ACLs to revoke,
  so the listener gate is the default-posture fence.

DC-DR therefore requires one of the two: auth-on for the ACL fence, or the listener
gate. Two rules keep the fence from breaking replication:

- **Never fence the MM2 connector principal or `super.users`.** The fence must target
  only client principals. Fencing the connector principal (or the operator's auth-secret
  `super.users` user) would break mirroring and consumer-offset sync along with client
  produce.
- **Do not blanket-mirror ACLs.** `MirrorSourceConnector` defaults
  `sync.topic.acls.enabled=true`, which would copy the active cluster's client produce
  ACLs onto the standby and re-grant produce there, undoing the fence. The operator sets
  `sync.topic.acls.enabled=false` and manages client ACLs per cluster.

## Planned switchover (drained, zero record loss)

Move the active DC on purpose by annotating the Kafka:

```bash
$ kubectl annotate kafka -n demo kf-dcdr dr.kubedb.com/switchover-to=dc-b
```

The hub then:

1. checks the target is a known, healthy DC within the MM2 lag budget;
2. sets `phase: FailingOver` and quiesces producers by closing the active cluster's
   produce fence;
3. waits for MM2 to drain to near-zero lag, so the target holds every record;
4. flips the bootstrap endpoint to the target, opens the target's fence, and reverses
   the mirror direction (disables the old direction's connectors, then enables the new
   direction's on the other DC's `ConnectCluster`), never both at once;
5. moves the Lease to the target.

Because MM2 fully drained before the flip, no committed record is lost. This is a
hub-driven annotation, not a `KafkaOpsRequest` type: the engine-aware quiesce and MM2
drain run in the hub, not in the engine-agnostic `dr-controlplane`.

## Failback

Failback is not a rewind. When a failed DC returns, it becomes the MM2 target of the
new active. The records it accepted but never mirrored before the failover are a forked
tail Kafka cannot rewind, and because MM2 only adds and never deletes, a naive
re-mirror leaves those orphan records on top of the new active's data (and they could
resurface if that DC is ever made active again). For correctness:

- **re-seed the affected topics from the new active** (wipe the returned cluster's copy
  and re-mirror from scratch), or
- **accept and document the orphan tail** as bounded loss.

Once the returned DC is caught up, a drained planned switchover returns the active DC:

```bash
$ kubectl annotate kafka -n demo kf-dcdr dr.kubedb.com/switchover-to=dc-a
```

## Scaling and day-2 operations

The standard `KafkaOpsRequest` operations (`UpdateVersion`, `HorizontalScaling`,
`VerticalScaling`, `VolumeExpansion`, `Restart`, `Reconfigure`, `ReconfigureTLS`,
`RotateAuth`, `StorageMigration`) apply to a DC-DR cluster. They act on the per-DC
Kafka clusters. Horizontal scaling operates per DC (each Member DC's cluster scales its
own brokers or controllers and handles KRaft membership intra-DC via KIP-853), so a
scaling request targets the data centers rather than a single flat broker set.

There is no failover ops type: unplanned failover is driven by the Lease, and the
planned switchover is the `dr.kubedb.com/switchover-to` annotation, not an ops request.

> **Note:** the distributed Kafka substrate and the DC-DR layer are net-new for Kafka.
> Treat the field names and flows in this guide as the intended user experience;
> confirm availability in your release before relying on them in production.

## Deletion and cleanup

```bash
$ kubectl delete kafka -n demo kf-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC Kafka clusters, the per-DC
`ConnectCluster` objects, the MM2 `Connector` objects, and the cluster-scoped per-DC
`PlacementPolicies` it generated (these carry no owner reference, so the operator
deletes them explicitly). The user-provided base `PlacementPolicy` is left for you to
delete.

## Limitations

- **No zero-RPO on an unplanned failover.** MM2 is asynchronous, so an unplanned
  active-DC loss loses the un-mirrored tail (bounded by MM2 lag). Use a drained planned
  switchover for a zero-record-loss move.
- **No rewind on failback.** A returned old-active cluster's un-mirrored forked tail
  cannot be rewound. Re-seed the affected topics or accept the orphan tail as bounded
  loss.
- **Two data DCs only.** Active/passive MM2 is inherently two-cluster. Three or more
  data DCs (fan-out mirroring, three-way failover) is a separate, larger design.
- **Cross-DC reachability is required.** Kafka advertises in-cluster `.svc` listeners,
  so MM2 and the cross-DC bootstrap need flat pod networking (KubeSlice) or external
  listeners.
