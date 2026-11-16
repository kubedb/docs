---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: es-dr-guide-elasticsearch
    name: User Guide
    parent: es-dr-elasticsearch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running Elasticsearch in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed Elasticsearch in cross data
center disaster recovery (DC-DR) mode: the components, the naming contract, deployment,
what the operator creates, indexing into the active endpoint, monitoring CCR follow lag,
the follower-read-only fence, switchover, failback, scaling, and day-2 operations.

Read the [DC-DR Overview](/docs/guides/elasticsearch/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/elasticsearch/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the active write DC. This is the single cross-DC failover authority. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease for its DC and projects the Lease decision into the local spoke as the primary-dc marker the fence reads. |
| **KubeDB Elasticsearch operator (hub)** | the OCM hub | Expands the `Elasticsearch` CR into per-DC Elasticsearch clusters, registers each as a remote of the other, and sets auto-follow patterns on the standby. On a Lease change it promotes the followers, flips the search/index endpoint, and reverses the CCR direction, then writes `status.disasterRecovery`. |
| **Per-DC Elasticsearch clusters** | each Member DC | Each is a self-contained Elasticsearch with its own intra-DC master quorum, data nodes, and per-shard primaries and replicas. The master quorum never crosses the DC boundary. |
| **CCR remote-cluster registration + auto-follow patterns** | the standby Member DC | The standby registers the active as a remote cluster and follows its leader indices via follower indices. CCR runs only active to standby, and the operator owns the direction. |
| **The follower-read-only fence** | each Member DC | Follower indices are inherently read-only; a non-active cluster refuses client writes to would-be leader indices, fail closed, so only the Lease holder takes writes. |
| **KubeSlice (or external listeners)** | each spoke | Provides the flat cross-DC pod network so CCR can reach the remote cluster's transport port (9300) and the search/index endpoint (9200) resolves across DCs. |

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
  name: es-dcdr
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
  Elasticsearch cluster with its own master quorum. The **Arbiter** DC carries an empty
  `replicaIndices` and holds only the `dr-controlplane` etcd member, never Elasticsearch.
- `mode: TwoDC` expects exactly two Member DCs plus the Arbiter DC. Three or more data DCs
  is a separate design and out of scope.
- Roles are `Member` and `Arbiter` only.

### Elasticsearch

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-dcdr
  namespace: demo
spec:
  version: xpack-8.19.9
  distributed: true
  enableSSL: true
  replicas: 6
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: es-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

### What the operator creates

- **One self-contained Elasticsearch cluster per Member DC** (`es-dcdr` materialized into a
  cluster in `dc-a` and a cluster in `dc-b`), each with its own master-eligible voting
  quorum and its own per-shard primaries and replicas. The two clusters never share a master
  quorum. `spec.replicas: 6` is the total node count across both Member DCs; the
  `replicaIndices` split it into a three-node cluster per DC (ordinals 0, 1, 2 in `dc-a` and
  3, 4, 5 in `dc-b`).
- **A remote-cluster registration on each DC**, so the standby can reach the active over the
  transport port (9300) via `cluster.remote.<alias>.seeds` pointed at the active's transport
  endpoint. The reverse-direction registration stays provisioned so a failover can enable
  auto-follow the other way.
- **Auto-follow patterns on the standby DC's cluster** that create follower indices for the
  active cluster's leader indices, pulling operations asynchronously. CCR runs only
  active-to-standby; the operator owns the direction so the two directions never overlap.
- **The Lease-gated search/index endpoint** that resolves to the active cluster's nodes.

CCR itself is configured through the Elasticsearch CCR REST API, which the operator drives.
Conceptually, on the standby cluster (`dc-b` while `dc-a` is active) the operator registers
the remote and sets an auto-follow pattern:

```bash
# Register the active cluster as a remote (transport seeds, port 9300):
PUT /_cluster/settings
{
  "persistent": {
    "cluster.remote.dc-a.seeds": [ "es-dcdr-dc-a-master.demo.svc:9300" ]
  }
}

# Auto-follow every new leader index on the active into a follower index here:
PUT /_ccr/auto_follow/dc-a-autofollow
{
  "remote_cluster": "dc-a",
  "leader_index_patterns": [ "*" ],
  "follow_index_pattern": "{{leader_index}}"
}
```

The follower's index name matches the leader's, so a failover is transparent to clients. To
bound catch-up like a replication slot, the operator sets the leader indices'
`index.soft_deletes.retention_lease.period` so a follower that falls behind can still resume;
a follower past retention forces a full re-follow.

## Connecting and indexing

A DC-DR Elasticsearch exposes one user-facing **search/index endpoint** (client port 9200)
that resolves to the active cluster's nodes. Clients always connect to that single endpoint
and reach the write cluster without reconfiguration. Because CCR keeps index names identical
on both clusters, after the endpoint flips clients keep using the same indices.

```bash
# Index into the active cluster through the single search/index endpoint:
$ curl -k -u "admin:$PASSWORD" -X POST \
    "https://es-dcdr.demo.svc:9200/orders/_doc" \
    -H 'Content-Type: application/json' -d '{"id": 1, "item": "book"}'
```

Only the active cluster accepts client writes. If clients somehow reach a standby cluster,
its indices are follower (read-only) indices and reject the write (see the fence below),
which is the split-brain guard.

### Reads on the standby

The standby cluster's follower indices are queryable read-only, so you can serve local
reads from the standby DC. After a failover or switchover, the promoted indices on the new
active accept writes again under identical names, so clients reconnecting through the flipped
endpoint keep working.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC that holds the Lease and takes client writes. |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | `Member` or `Arbiter`. |
| `dataCenters[].writable` | True only for the active cluster. |
| `dataCenters[].nodesReady` | Ready node count in that DC's cluster. |
| `dataCenters[].followLagOps` | The standby's CCR follow lag in operations (`leader_global_checkpoint` minus `follower_global_checkpoint`, summed across follower indices). |
| `dataCenters[].healthy` | DC health: a Member DC is healthy when its nodes are ready; the Arbiter DC is healthy when its `dr-controlplane` etcd member is reachable (so an Arbiter reports `healthy: true` with `nodesReady: 0`). |

### CCR follow lag

Cross-DC lag comes from CCR's own follow-stats: `leader_global_checkpoint` minus
`follower_global_checkpoint` per follower index, a count of operations the follower has not
yet applied. The hub surfaces this into `followLagOps`; there is no lag field in the base
`ElasticsearchStatus`.

```bash
# Raw CCR follow-stats on the standby cluster:
$ curl -k -u "admin:$PASSWORD" "https://es-dcdr-dc-b.demo.svc:9200/_ccr/stats" | jq
```

### Useful checks

```bash
# Which DC the Lease intends as active (from the coordination plane):
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# Per-DC nodes and roles:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=es-dcdr \
    -L kubedb.com/role,open-cluster-management.io/cluster-name

# Follow lag per DC from status:
$ kubectl get elasticsearch -n demo es-dcdr \
    -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.followLagOps} healthy={.healthy}{"\n"}{end}'
```

## The follower-read-only fence

A non-active cluster must refuse client writes, fail closed: a cluster that cannot confirm it
holds the Lease keeps its indices as read-only followers. The fence is the CCR follower model
itself.

- **Follower indices are read-only.** While a cluster is standby, its indices are CCR
  followers, and Elasticsearch rejects writes to a follower index. There are no writable
  leader indices on the standby until the operator promotes them, and it only promotes when
  the Lease moves.
- **Fail closed on the Lease.** A cluster that cannot confirm it holds the primary-DC Lease
  never promotes its followers, so it stays read-only. A partitioned old-active cluster that
  loses its Lease renewal stops being writable on its own, before the hub reacts.

Two rules keep the fence from breaking replication:

- **Never promote followers without the Lease.** Promotion (`pause_follow`, `unfollow`,
  convert to writable) is the only thing that lifts the fence, and only the hub does it, only
  on a Lease change. A cluster never self-promotes.
- **Never run CCR both directions for the same index.** CCR is one-directional per index. The
  operator disables the old direction's auto-follow and promotes the old followers before (or
  atomically with) enabling the new direction, so an index is never a leader and a follower at
  once.

## Planned switchover (drained, zero document loss)

Move the active DC on purpose by annotating the Elasticsearch:

```bash
$ kubectl annotate elasticsearch -n demo es-dcdr dr.kubedb.com/switchover-to=dc-b
```

The hub then:

1. checks the target is a known, healthy DC within the CCR follow-lag budget;
2. sets `phase: FailingOver` and quiesces indexing on the active cluster;
3. waits for CCR to drain to zero follow lag, so the target has every operation;
4. pauses and promotes the target's follower indices (`pause_follow`, `unfollow`, convert to
   writable), flips the search/index endpoint to the target, and starts CCR in the reverse
   direction (auto-follow on the old active once it returns), never both directions at once;
5. moves the Lease to the target.

Because CCR fully drained before the flip, no acknowledged document is lost. This is a
hub-driven annotation, not an `ElasticsearchOpsRequest` type: the engine-aware quiesce and CCR
drain run in the hub, not in the engine-agnostic `dr-controlplane`.

## Failback

Failback is not a rewind. When a failed DC returns, it becomes the CCR follower of the new
active via auto-follow, and catches up. The operations it accepted but never followed before
the failover are a forked tail Elasticsearch cannot rewind. For correctness:

- **reconcile the forked tail out of band or re-seed the affected indices** from the new
  active (delete the returned cluster's copy of an affected index and let auto-follow re-seed
  it from scratch), or
- **accept and document the forked tail** as bounded loss.

Once the returned DC is caught up (low follow lag), a drained planned switchover returns the
active DC:

```bash
$ kubectl annotate elasticsearch -n demo es-dcdr dr.kubedb.com/switchover-to=dc-a
```

## Scaling and day-2 operations

The standard `ElasticsearchOpsRequest` operations (`UpdateVersion`, `HorizontalScaling`,
`VerticalScaling`, `VolumeExpansion`, `Restart`, `Reconfigure`, `ReconfigureTLS`,
`RotateAuth`) apply to a DC-DR cluster. They act on the per-DC Elasticsearch clusters.
Horizontal scaling operates per DC (each Member DC's cluster scales its own nodes and handles
master-quorum membership intra-DC), so a scaling request targets the data centers rather than
a single flat node set.

There is no failover ops type: unplanned failover is driven by the Lease, and the planned
switchover is the `dr.kubedb.com/switchover-to` annotation, not an ops request.

> **Note:** the distributed Elasticsearch substrate and the DC-DR layer are net-new for
> Elasticsearch. Treat the field names and flows in this guide as the intended user
> experience; confirm availability in your release before relying on them in production.

## Deletion and cleanup

```bash
$ kubectl delete elasticsearch -n demo es-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC Elasticsearch clusters, the
remote-cluster registrations and auto-follow patterns, and the cluster-scoped per-DC
`PlacementPolicies` it generated (these carry no owner reference, so the operator deletes them
explicitly). The user-provided base `PlacementPolicy` is left for you to delete.

## Limitations

- **CCR licensing.** CCR is an Elastic Platinum/Enterprise feature, so the licensed
  Elasticsearch image must support it. OpenSearch uses its own cross-cluster replication
  plugin (OSS) for the equivalent leader/follower model. Confirm the feature is available in
  your image before relying on DC-DR.
- **No zero-RPO on an unplanned failover.** CCR is asynchronous, so an unplanned active-DC
  loss loses the un-followed tail (bounded by CCR follow lag). Use a drained planned
  switchover for a zero-document-loss move.
- **No rewind on failback.** A returned old-active cluster's un-followed forked tail cannot be
  rewound. Reconcile it out of band, re-seed the affected indices, or accept the forked tail
  as bounded loss.
- **Two data DCs only.** Active/passive CCR is inherently two-cluster. Three or more data DCs
  (fan-out following, three-way failover) is a separate, larger design.
- **Cross-DC reachability is required.** Elasticsearch nodes need routable cross-DC access to
  the transport port (9300, for the remote-cluster seeds) and the client endpoint (9200), so
  flat pod networking (KubeSlice) or external listeners are required.
