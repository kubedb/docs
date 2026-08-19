---
title: Etcd Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-horizontal-scaling-overview
    name: Overview
    parent: etcd-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Horizontal Scaling

This guide gives an overview of how KubeDB changes the number of members of an `Etcd` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## Etcd Membership Is Not a Replica Count

Read this section before anything else — horizontal scaling for etcd does not work the way it does
for a stateless Deployment, or even for most other KubeDB databases.

An etcd cluster is a Raft consensus group. Its membership lives **inside etcd itself**, not in the
`PetSet`: every member holds a copy of the member list, and every write has to be acknowledged by a
quorum of `N/2+1` **voting** members. Adding or removing a member is therefore a Raft configuration
change that has to be committed by the cluster, not a number you increment on a controller object.

Two consequences shape everything below:

1. **A membership change and a pod change are two different writes** that must be ordered correctly.
   Adding a pod without adding the member gives you a process that cannot join; adding the member
   without the pod gives you a cluster whose quorum requirement has already grown while the new
   member does not exist yet.
2. **Scaling is never atomic.** A `3 → 5` scale up is two separate member additions, each with its own
   catch-up period. KubeDB deliberately performs **one membership mutation at a time**, so a scale up
   or down is a sequence of steps, not a single instantaneous change.

### Learners

etcd solves the "a brand new member drags the cluster down while it catches up" problem with
**learners**. A learner:

- receives the full Raft log and applies it, exactly like a voting member;
- **does not count towards quorum** and cannot vote;
- cannot serve linearizable reads.

KubeDB adds every new member as a learner first (`MemberAddAsLearner`), waits for it to catch up, and
only then **promotes** it to a voting member (`MemberPromote`). Because a learner is not part of
quorum, the cluster's fault tolerance is completely unchanged while it is catching up — a 3-member
cluster growing to 4 still needs 2 of its 3 voters, not 3 of 4.

> **Why readiness is not the gate.** etcd's `/health` endpoint performs a linearizable read, which a
> learner *cannot* serve — so a learner pod never becomes `Ready`. KubeDB therefore does **not** gate
> promotion on pod readiness (that would deadlock the scale up). Instead it compares the learner's
> applied revision against the leader's and promotes once the ratio reaches `0.9`, mirroring etcd's
> own `IsLearnerReady` heuristic. etcd rejects a premature promotion anyway; the ratio check just
> avoids hammering it with doomed calls.

## Who Actually Does the Work

The membership reconciliation lives in the **Provisioner** operator, not in the ops request. It runs
continuously, comparing etcd's real member list against `spec.replicas` on the `Etcd` object, and
performs **at most one mutation per reconcile pass**:

- If a learner exists, finish that first — check whether it has caught up, and promote it. Nothing
  else happens in that pass. A learner is never left in place while another member is added.
- Otherwise, if the member count is below `spec.replicas`, add exactly one member as a learner.
- Otherwise, if the member count is above `spec.replicas`, remove exactly one member.
- If the `PetSet` replica count and the etcd member count disagree (for example because the operator
  restarted between the two writes), realign the `PetSet` with the member list first. **etcd's member
  list is the authority** on who is in the cluster.

This is why membership converges even without an ops request: if you edit `spec.replicas` on the
`Etcd` object directly, the same sequence runs. The `HorizontalScaling` `EtcdOpsRequest` is a thin,
auditable wrapper around that edit — it patches `spec.replicas` and then waits for the provisioner to
converge, reporting progress through its own status.

> **The one ops type that does not pause the database.** Every other `EtcdOpsRequest` pauses the
> `Etcd` object so the Provisioner operator keeps its hands off. `HorizontalScaling` deliberately
> does **not** — pausing would stop the very controller that performs the membership changes, and the
> convergence the ops request is waiting for would never happen. Mutual exclusion is still
> guaranteed: only one `EtcdOpsRequest` per database may be in `Progressing` at a time.

## How Scale Up Works

For a `3 → 5` scale up, with pods `etcd-cluster-0..2` already running:

1. The user creates an `EtcdOpsRequest` of type `HorizontalScaling` with
   `spec.horizontalScaling.replicas: 5`.
2. The Ops-manager operator moves the request to `Progressing` and patches `spec.replicas: 5` on the
   `Etcd` object (`UpdateDatabase` condition).
3. The Provisioner operator sees `3 members < 5 desired` and, in one pass:
   - rewrites the cluster-state ConfigMap for 4 members with `--initial-cluster-state=existing` —
     this is written **first**, because the new pod reads its bootstrap membership from it at start-up
     and etcd validates that a joining member's `--initial-cluster` matches the cluster's member list
     exactly;
   - calls `MemberAddAsLearner` with the peer URL of the next ordinal
     (`http://etcd-cluster-3.etcd-cluster-pods.demo.svc.cluster.local:2380`);
   - bumps the `PetSet` to 4 replicas, so the pod is created and the learner starts syncing.
4. On subsequent passes the provisioner sees a learner and does nothing but poll its revision against
   the leader's. Once the ratio reaches `0.9` it calls `MemberPromote`. The cluster now has 4 voting
   members.
5. Only now does the provisioner add `etcd-cluster-4` as a learner, and the same catch-up/promote
   cycle repeats.
6. The ops request polls until **all** of the following hold, and only then reports `Successful`:
   the `PetSet` carries 5 replicas, etcd's member list has exactly 5 members, **none of them is still
   a learner**, and the cluster reports a healthy quorum.

At no point is more than one membership change in flight, and at no point does the cluster contain
more than one non-voting member — so quorum is never put at risk by the scale up itself.

## How Scale Down Works

Scale down is the mirror image, and the **ordering is the important part**:

1. The provisioner identifies the **highest-ordinal** member. It maps each etcd member back to a
   `PetSet` ordinal by peer URL (a member that has not started yet reports an empty `--name`, so the
   peer URL is the reliable key), and picks the largest.
2. It calls `MemberRemove` on that member **first**, while its pod is still running.
3. Only after the removal is committed does it rewrite the cluster-state ConfigMap and shrink the
   `PetSet` by one, which deletes the pod.

Doing it the other way round — deleting the pod and then removing the member — would leave a member
in the Raft configuration whose process is gone. A 5-member cluster would still require 3 votes while
only 4 members exist, so the cluster would be one failure away from losing quorum for no reason. By
removing first, the quorum requirement drops as soon as the configuration change commits, and etcd's
membership never disagrees with the pod count for longer than the gap between two API writes (and
even that gap is repaired on the next pass).

One member is removed per pass, so a `5 → 3` scale down is two sequential removals. The provisioner
refuses to scale below a single member.

> **PVCs are retained.** Shrinking the `PetSet` does not delete the data volumes of the removed
> members. `data-etcd-cluster-3` and `data-etcd-cluster-4` stay bound after a `5 → 3` scale down; if
> you scale back up later they are adopted by the new pods (which is harmless — the member is added
> fresh as a learner and re-syncs from the leader). Delete them by hand if you want the space back.

## Choosing a Member Count

- Only **voting** members count towards quorum, and quorum is `N/2+1`. So 3 members tolerate 1
  failure, 4 members *also* tolerate only 1, and 5 members tolerate 2. **Odd counts are what you
  want**; an even count buys you extra write latency without extra fault tolerance.
- The webhook only rejects `spec.horizontalScaling.replicas < 1`. Even counts are allowed, just
  discouraged.
- Larger clusters are not faster. Every write has to be replicated to a quorum, so 7 members are
  slower to write than 3. Scale out for fault tolerance and read capacity, not for write throughput.

In the [next](/docs/guides/etcd/scaling/horizontal-scaling/horizontal-scaling.md) doc, we are going
to show a step-by-step guide on scaling an Etcd cluster up and down using the `EtcdOpsRequest` CRD.
