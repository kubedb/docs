---
title: Etcd Maintenance Operations Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-maintenance-overview
    name: Overview
    parent: etcd-maintenance
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Maintenance Operations

etcd is unlike the other databases KubeDB manages in one important way: it is its own consensus
layer, and it exposes its housekeeping as first-class RPCs on the cluster API rather than as
shell commands or SQL. KubeDB surfaces three of those RPCs as `EtcdOpsRequest` types:

| OpsRequest type | etcd RPC      | What it does                                                        |
| --------------- | ------------- | ------------------------------------------------------------------- |
| `MoveLeader`    | `MoveLeader`  | Hands Raft leadership to another member, deliberately.              |
| `Compact`       | `Compact`     | Discards superseded key revisions from the keyspace history.        |
| `Defragment`    | `Defragment`  | Shrinks the on-disk backend file so the freed pages return to the OS. |

None of these three has an analogue in any other KubeDB database — there is no Postgres-style
"force failover" or "reconnect standby" step for etcd, because Raft already handles failover on
its own; what you occasionally want is to *schedule* the handover rather than wait for a failure.
Likewise there is nothing like `VACUUM FULL` — but etcd's MVCC store has the same
"deleting data does not free disk" property, which is what compaction and defragmentation
address together.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## Raft leadership transfer

Every etcd cluster elects one member as the Raft **leader**. The leader is the only member that
appends new entries to the replicated log; the other members (**followers**) forward writes to it
and replicate what it commits. Reads are served by any member, but a write always goes through
the leader.

If the leader disappears — the node is drained, the pod is evicted, the kubelet restarts — the
remaining members notice the missing heartbeats, hold an election, and pick a new leader. That
works, but it is not free: for the duration of the election timeout plus the campaign, there is
no leader, and writes fail or block.

A **leadership transfer** avoids that gap. The current leader tells a chosen follower "you take
over now", the follower campaigns immediately with an up-to-date log, and the handover completes
in roughly a round trip instead of an election timeout.

You would use `MoveLeader` when you are about to do something disruptive to the node that
currently hosts the leader, and you want to choose the moment yourself:

- Before cordoning or draining a Kubernetes node for maintenance.
- Before deliberately deleting or restarting the leader's pod.
- To move leadership off a member that sits on a saturated or badly-placed node.

Note that KubeDB already does this for you inside the operations it drives itself: a
[Restart](/docs/guides/etcd/restart/restart.md), a `VerticalScaling` in `Restart` mode and the
other rolling operations all move leadership off the leader before recreating its pod, and they
visit the leader last. `MoveLeader` exists for the times when *you* are the one about to disturb
the node.

Read the guide: [Move the Raft leadership of an Etcd cluster](/docs/guides/etcd/maintenance/move-leader.md).

## Compaction

etcd is a multi-version store. Every write creates a new **revision** of the keyspace, and the
older revisions are kept so that watches can be resumed from a past point and so that historical
reads work. That history is not a rounding error — a workload that repeatedly overwrites the same
few keys still accumulates one revision per write, forever, until something removes them.

**Compaction** is that something. Compacting to revision *N* discards every superseded version of
every key at revisions below *N*, keeping only the newest value of each key at or before *N*.
Live data is never lost; only history is. After compaction, reads and watches at revisions below
*N* fail with `mvcc: required revision has been compacted` — which is why you should not compact
to "now" while a consumer is deliberately reading a lagging revision.

etcd can compact on its own if you configure automatic compaction. In KubeDB, that is
`spec.configuration.tuning.autoCompactionMode` (`periodic` or `revision`) together with
`spec.configuration.tuning.autoCompactionRetention` on the `Etcd` object, which map onto etcd's
own `--auto-compaction-mode` and `--auto-compaction-retention` flags. The `Compact`
`EtcdOpsRequest` is the on-demand counterpart, for when you want to reclaim history right now, or
at a precise revision.

Read the guide: [Compact the keyspace history of an Etcd cluster](/docs/guides/etcd/maintenance/compact.md).

## Defragmentation

Here is the part that surprises people: **compaction alone does not shrink anything on disk.**

Each etcd member stores its data in a bbolt file. When compaction drops old revisions, the pages
those revisions occupied are marked free *inside* that file and become available for etcd to
reuse — but the file itself keeps its size, and the operating system never sees the space come
back. Over a long-lived cluster with a churny keyspace, the gap between "logical data size" and
"backend file size" can become large.

**Defragmentation** closes that gap. It rewrites a member's backend into a compact form and
releases the freed space back to the filesystem. Two properties matter operationally:

- It is a **per-member, local** operation. Unlike compaction, it does not go through Raft. Each
  member has to be defragmented individually, and each ends up with its own file size.
- It **blocks the member it runs against** for the whole duration. The member stops answering
  while its backend is being rewritten, which on a large backend can take a while.

That second property is why KubeDB never fires defragmentation at the whole cluster at once. The
`Defragment` `EtcdOpsRequest` walks the members strictly one at a time, re-checking that the
cluster still has quorum before moving on to the next one, and it defragments the leader last.

Defragmentation also matters for a specific failure mode. When a member's backend exceeds
`--quota-backend-bytes` (`spec.configuration.tuning.quotaBackendBytes`), etcd raises a cluster-wide
`NOSPACE` alarm and refuses further writes until the alarm is disarmed. Disarming the alarm
before actually reclaiming the space just re-raises it, so the correct order is compact, then
defragment, then disarm — which is exactly what the `Defragment` ops request does as its final
step.

Read the guide: [Defragment the backends of an Etcd cluster](/docs/guides/etcd/maintenance/defragment.md).

## Putting them together

The three operations are commonly used as a sequence rather than in isolation:

1. **`Compact`** — make the old revisions eligible for reuse.
2. **`Defragment`** — actually return the freed space to the filesystem, and clear any alarm the
   full backend had raised.
3. **`MoveLeader`** — optional, and generally used the other way around: as preparation for
   node-level maintenance rather than as a follow-up.

A useful mental model: compaction decides *what data is allowed to go away*, defragmentation
decides *when the disk space comes back*, and leadership transfer decides *which member is in
charge while you work on the cluster*.

Some practical guidance:

- These three ops requests never pause the `Etcd` object, never patch the PetSet and never
  restart a pod. They are pure control-plane calls against the running cluster.
- Only one `EtcdOpsRequest` per database can be in the `Progressing` phase at a time, so run them
  one after the other rather than applying all three at once.
- Compaction and defragmentation both add load. Schedule them when the cluster is not at its
  busiest, and prefer automatic compaction (`spec.configuration.tuning`) for the routine case,
  reserving the ops requests for on-demand work.

## Next Steps

- [Move the Raft leadership](/docs/guides/etcd/maintenance/move-leader.md).
- [Compact the keyspace history](/docs/guides/etcd/maintenance/compact.md).
- [Defragment the member backends](/docs/guides/etcd/maintenance/defragment.md).
- Detail concepts of the [Etcd object](/docs/guides/etcd/concepts/etcd.md).
