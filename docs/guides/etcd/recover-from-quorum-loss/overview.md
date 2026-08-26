---
title: Etcd Recover from Quorum Loss Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-recover-from-quorum-loss-overview
    name: Overview
    parent: etcd-recover-from-quorum-loss
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Recovery from Quorum Loss

This guide gives an overview of how the KubeDB Ops-manager operator rebuilds an `Etcd` cluster that
has **permanently** lost its Raft quorum, from a single surviving member.

> **This is a break-glass, disaster-recovery procedure, not routine maintenance.** It destroys the
> data of every member except the one you confirm, and it can lose writes. Read this whole page
> before you create a `RecoverFromQuorumLoss` `EtcdOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/etcd/scaling/horizontal-scaling/overview.md) — the
    learner-add/promote mechanism this recovery relies on to regrow the cluster.

## What Quorum Loss Actually Is

An etcd cluster is a Raft consensus group. Every write has to be acknowledged by a quorum of
`floor(N/2) + 1` **voting** members, so a 3-member cluster needs 2 and a 5-member cluster needs 3.

When a *majority* of the voting members is gone **for good**, Raft can never make progress again:

- the surviving minority cannot elect a leader, so no write — and no read that has to be
  linearizable — can be served;
- the dead members cannot be removed from the membership either, because a membership change is
  itself a Raft proposal and therefore needs the very quorum that no longer exists.

That deadlock is not something the cluster can heal from on its own. The only way out is etcd's own
documented `--force-new-cluster` procedure, which rewrites one surviving member's data directory so
that the cluster it describes consists of that member alone. `RecoverFromQuorumLoss` is the
`EtcdOpsRequest` type that performs that procedure for you, safely and resumably.

## When to Use It

Use it when a **majority of members is permanently gone**, and there is no prospect of them coming
back. In practice that means the member's node *and* its data volume are both gone:

- several nodes were deleted, or their disks failed, and the PVs went with them;
- a cloud availability zone holding two of three members was lost;
- someone deleted the PVCs of a majority of the members.

**Do not use it for any of the following** — in every one of these cases the cluster either still
has quorum, or will heal on its own once the missing members come back, and running this recovery
would throw away good data for nothing:

| Situation | What to do instead |
|---|---|
| A single member of a 3-member cluster is down | Nothing. The cluster still has quorum. Fix the node or let the `PetSet` recreate the pod. |
| A member is down but its PVC still exists | Wait. When the pod is rescheduled on the volume, it rejoins with its own data. |
| Members are up but slow / flapping | Investigate disk latency and networking. See [monitoring](/docs/guides/etcd/monitoring/overview.md). |
| Every member is down but every volume is intact | Wait, or perform a [Restart](/docs/guides/etcd/restart/restart.md). The members re-elect a leader as they come back. |
| The data itself is bad, and you want a known-good copy back | Use an [in-place Restore](/docs/guides/etcd/restore/overview.md) from a KubeStash snapshot. |

## The `QuorumLost` Signal

The Provisioner operator's health checker dials every member and asks etcd whether a quorum is
answering. When it is not, the operator sets a `QuorumLost` condition on the `Etcd` object:

```yaml
status:
  conditions:
    - type: QuorumLost
      status: "True"
      reason: RaftQuorumLost
      message: >-
        The Etcd: demo/etcd-cluster has lost its Raft quorum: 2 of 3 members are not answering.
        ...
```

Two things about this condition matter:

1. **Nothing acts on it automatically.** It is only ever a signal for a human (or for a
   `RecoverFromQuorumLoss` ops request authored by one). Rebuilding a cluster from one survivor
   throws data away, so KubeDB will never do it behind your back.
2. **It is a live gate on the recovery.** The ops request reads the condition straight from the API
   server, and refuses to run at all unless the `Etcd` currently reports `QuorumLost=True`. The
   condition is flipped back to `False` (with reason `RaftQuorumHealthy`) rather than removed once
   quorum returns, so a cluster that healed on its own cannot be torn down by a request that was
   written while it was broken.

## What Gets Lost

Two distinct things:

- **Every other member's data directory is discarded, permanently.** After the recovery the cluster
  is a brand new one-member cluster with a new cluster ID, so a member still carrying the old
  cluster ID could never rejoin it anyway. Those members come back as empty learners.
- **Any write the lost majority had committed but which the survivor had not yet applied is gone.**
  There is no way to get it back. This is inherent to `--force-new-cluster`, not a KubeDB
  limitation — the writes only ever existed on machines you no longer have.

Choosing the survivor with the highest Raft applied index is what minimises the second loss, which
is exactly what the operator does when you do not name a member yourself.

## The Confirmation Gate

The recovery has a **mandatory two-step handshake**, and it is the most important thing to
understand about this ops request:

1. You create the `EtcdOpsRequest`. `spec.recoverFromQuorumLoss.member` is optional — leave it
   empty and the operator picks the reachable member with the **highest Raft applied index**; set it
   (a pod name, or a bare ordinal like `"0"`) to choose one yourself.

2. The operator resolves the survivor **exactly once**, records it in an
   `EtcdQuorumLossSurvivorResolved--<pod>` condition, and then **stops**. It reports the choice in
   the `EtcdQuorumLossAwaitingConfirmation` condition and waits — indefinitely, with no timeout and
   no failure — until `spec.recoverFromQuorumLoss.confirmMember` names exactly that pod.

3. You read the resolved name, agree with it, and patch `confirmMember` to match. Only then does
   anything destructive happen.

This is deliberate. The gate exists so that a mistyped `member` — or a survivor you did not expect
the operator to pick — cannot silently destroy the wrong member's data. And because the resolution
happens only once and is never re-picked on a later pass, the member you confirmed is guaranteed to
be the member the recovery keeps. If the resolved survivor is not the one you want, delete the ops
request and create a new one naming the member you do want; there is no way to change the answer
in place.

## How the Recovery Process Works

Every step is gated by its own condition on the ops request, so an operator restart resumes exactly
where it left off rather than starting over.

1. The user creates an `EtcdOpsRequest` of type `RecoverFromQuorumLoss` with a `spec.timeout`.

2. **Admission.** The operator reads the `Etcd` object live and refuses the request outright unless
   `QuorumLost=True`. A refusal is terminal — it does not burn `spec.maxRetries` and retrying it
   cannot change the answer.

3. The Ops-manager operator pauses the `Etcd` object, so the Provisioner operator does not fight
   the recovery.

4. **The survivor is resolved** and the request waits for `confirmMember`, as described above.

5. **Preflight.** The last look at the cluster while everything is still intact: the quorum really
   is still lost (a cluster that recovered on its own while the request waited for confirmation is
   refused here), and the pod of ordinal 0 exists, since its manifest is what the rebuilt member is
   recreated from.

6. The `PetSet` is deleted **with an orphan propagation policy** — it would otherwise recreate
   every pod the recovery deletes and re-provision every claim it discards.

7. **The survivor is relocated onto ordinal 0.** If the survivor already *is* ordinal 0, its pod is
   simply taken down. Otherwise its volume is rebound under ordinal 0's claim name: no data is
   copied, only the name the volume answers to changes, and the rebuilt member is pinned to the node
   the survivor was running on.

8. **Every other member is discarded**, one per pass: the pod is taken down (tolerating members
   whose pod object is already gone) and its claim is dropped.

9. The cluster-state ConfigMap is rewritten for a **single member** cluster.

10. **Ordinal 0 is booted with `--force-new-cluster`**, which is the step that actually rebuilds the
    cluster. The operator then waits until it is `Ready` and reports exactly one member.

11. **The flag is taken back off** by recreating the member once more without it.
    `--force-new-cluster` is not self-clearing: left on the command line it would rewrite the data
    directory into a fresh single-member cluster on *every* later restart, silently discarding every
    member added back in the meantime.

12. The reclaim policies parked while the volumes were being swapped are restored — **this is what
    finally destroys the discarded members' data**, and it is deliberately left until the rebuilt
    member is up and healthy, so a crash mid-recovery has thrown nothing away.

13. The `PetSet` is recreated around the rebuilt member, the `Etcd` object is resumed, and the
    request is marked `Successful`.

Nothing in the recovery regrows the cluster. Once the database is resumed, the ordinary membership
reconciliation sees one member against `spec.replicas` and adds the rest back through **exactly the
same learner-add/promote path an ordinary scale up uses** — each new member joins as a learner,
streams the recovered keyspace from the survivor, and is promoted to a voting member once it has
caught up.

## Preconditions

- The `Etcd` object must currently report `QuorumLost=True`. Anything else is refused.
- **`spec.timeout` is required.** Several steps — waiting for the rebuilt member to come up, waiting
  for a volume to rebind — have no other deadline at all.
- `spec.storageType` must be `Durable`. An ephemeral member's data died with its pod, so there is
  nothing to rescue.
- At least one member must still answer, and if you name one in `spec.recoverFromQuorumLoss.member`
  it must be that one — a cluster cannot be rebuilt from a member that cannot be reached.

In the [next](/docs/guides/etcd/recover-from-quorum-loss/recover-from-quorum-loss.md) doc, we are
going to show a step-by-step guide on recovering an Etcd cluster from a permanent quorum loss using
the `EtcdOpsRequest` CRD.
