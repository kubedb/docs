---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: rm-dr-guide-rabbitmq
    name: User Guide
    parent: rm-dr-rabbitmq
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running RabbitMQ in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed RabbitMQ in cross data center
disaster recovery (DC-DR) mode: the components, the naming contract, deployment, what
the operator creates, publishing to the active endpoint, consumers resuming after a
flip, monitoring federation lag, the publish fence, switchover, failback, scaling, and
day-2 operations.

Read the [DC-DR Overview](/docs/guides/rabbitmq/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/rabbitmq/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the active publish DC. This is the single cross-DC failover authority. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease for its DC and projects the Lease decision into the local spoke as the primary-dc marker the publish fence reads. |
| **KubeDB RabbitMQ operator (hub)** | the OCM hub | Expands the `RabbitMQ` CR into per-DC RabbitMQ clusters and the Federation upstreams and policies on the standby. On a Lease change it flips the AMQP endpoint, reverses the federation direction, and moves the publish fence, then writes `status.disasterRecovery`. |
| **Per-DC RabbitMQ clusters** | each Member DC | Each is a self-contained RabbitMQ with its own intra-DC nodes and quorum-queue Raft groups. Raft never crosses the DC boundary. |
| **Federation upstreams and policies** | on the standby DC | One upstream on the standby cluster pulling from the active cluster's endpoint, plus the policies that mark which exchanges and queues are federated, mirroring active to standby. |
| **The publish fence** | each Member DC | Denies client publishes on a non-active cluster, fail closed, so only the Lease holder takes publishes. |
| **KubeSlice (or external listeners)** | each spoke | Provides the flat cross-DC pod network so Federation can reach the remote cluster and the AMQP endpoint resolves across DCs. |

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
  name: rm-dcdr
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
  RabbitMQ cluster with its own quorum-queue Raft. The **Arbiter** DC carries an empty
  `replicaIndices` and holds only the `dr-controlplane` etcd member, never RabbitMQ.
- `mode: TwoDC` expects exactly two Member DCs plus the Arbiter DC. Three or more data
  DCs is a separate design and out of scope.
- Roles are `Member` and `Arbiter` only.

### RabbitMQ

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm-dcdr
  namespace: demo
spec:
  version: "3.13.2"
  distributed: true
  replicas: 6
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: rm-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

### What the operator creates

- **One self-contained RabbitMQ cluster per Member DC** (`rm-dcdr` materialized into a
  cluster in `dc-a` and a cluster in `dc-b`), each with its own nodes and quorum-queue
  Raft groups. The two clusters never share a Raft group. `spec.replicas: 6` is the
  total node count across both Member DCs; the `replicaIndices` split it into a
  three-node cluster per DC (ordinals 0, 1, 2 in `dc-a` and 3, 4, 5 in `dc-b`).
- **Federation upstreams and policies on the standby cluster**, because a federation
  upstream lives on the target (standby) side and pulls from the active source. The
  reverse-direction upstream stays defined but disabled until a failover needs it.
- **The Lease-gated AMQP Service** that resolves to the active cluster's nodes.

RabbitMQ Federation is configured through runtime parameters (the upstream) and a
policy (which resources federate). There is no KubeDB `Connector` CRD for RabbitMQ: the
operator manages the federation runtime parameters and policies directly on the standby
cluster (they ride on the broker configuration and the management/federation endpoint).
Conceptually the operator defines, on the standby cluster (`dc-b` while `dc-a` is
active), a federation upstream pointing at the active cluster:

```jsonc
// federation upstream parameter, set by the operator on the standby (dc-b) cluster
{
  "component": "federation-upstream",
  "name": "dcdr-upstream-from-dc-a",
  "value": {
    // the active cluster's Lease-routed AMQP endpoint
    "uri": "amqp://rm-dcdr-dc-a.demo.svc:5672",
    "ack-mode": "on-confirm",
    "trust-user-id": false
  }
}
```

and a policy selecting which queues federate from that upstream:

```jsonc
// federation policy, set by the operator on the standby (dc-b) cluster
{
  "name": "dcdr-federation",
  "pattern": "^(?!amq\\.).*",        // federate user queues and exchanges
  "apply-to": "queues",
  "definition": {
    "federation-upstream-set": "dcdr-upstream-from-dc-a"
  }
}
```

The operator owns these federation parameters and policies so the direction stays
operator-controlled and the two directions never overlap. Use quorum queues on both
clusters so intra-DC HA survives node loss; classic queues are non-replicated and
provide no intra-DC redundancy.

## Connecting and publishing

A DC-DR RabbitMQ exposes one user-facing **AMQP Service** that resolves to the active
cluster's nodes. Publishers and consumers always connect to that single endpoint and
reach the publish cluster without reconfiguration. Because Federation preserves queue
and exchange names on both clusters, after the endpoint flips clients keep using the
same queues.

```bash
# Publish to the active cluster through the single AMQP endpoint:
$ kubectl run perf-test -n demo --image=pivotalrabbitmq/perf-test -- \
    --uri "amqp://admin:password@rm-dcdr.demo.svc:5672/" --queue orders --quorum-queue
```

Only the active cluster accepts client publishes. If clients somehow reach a standby
cluster, its publish fence rejects the write (see below), which is the split-brain
guard. AMQP is on port 5672; inter-node traffic is on 25672; the management and
federation endpoint is on 15672.

### Consumers resume after a flip

Federation replicates messages (and, with the right upstream settings, acknowledgements)
into the standby cluster's queues, so after a failover or switchover a consumer
reconnecting through the flipped endpoint finds its queues and resumes from the
federated state on the new active cluster rather than re-reading from the beginning.
Because Federation is asynchronous and does not deduplicate across the flip, a
just-failed-over consumer can see a small window of redelivered messages: make consumers
idempotent, or apply a dedup window across the flip.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC that holds the Lease and takes client publishes. |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | `Member` or `Arbiter`. |
| `dataCenters[].writable` | True only for the active cluster. |
| `dataCenters[].nodesReady` | Ready RabbitMQ node count in that DC's cluster. |
| `dataCenters[].federationLagMessages` | The standby's federation backlog behind the active, in messages. |
| `dataCenters[].healthy` | DC health: a Member DC is healthy when its nodes are ready; the Arbiter DC is healthy when its `dr-controlplane` etcd member is reachable (so an Arbiter reports `healthy: true` with `nodesReady: 0`). |

### Federation lag

Cross-DC lag comes from Federation's own view: the gap between the active (upstream)
cluster's position and the standby (downstream) cluster's position for the federated
resources, exposed through the management/federation endpoint (federation link status
and per-queue message counts). The hub surfaces this into `federationLagMessages`; there
is no lag field in the base `RabbitMQStatus`.

### Useful checks

```bash
# Which DC the Lease intends as active (from the coordination plane):
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# Per-DC nodes and roles:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=rm-dcdr \
    -L kubedb.com/role,open-cluster-management.io/cluster-name

# Federation link status on the standby cluster (via rabbitmqctl):
$ kubectl exec -n demo rm-dcdr-dc-b-0 -- rabbitmqctl list_federation_links
```

## The publish fence

A non-active cluster must refuse client publishes, fail closed: a cluster that cannot
confirm it holds the Lease denies publishes. There are two fence mechanisms.

- **Permission fence:** revoke the `write` (configure/write/read) permission for the
  CLIENT users on the non-active cluster's virtual hosts, so client channels cannot
  publish. This requires the operator to manage per-user permissions per cluster.
- **Listener-gate fence:** gate the AMQP client listener (5672) on the non-active cluster
  so publishers cannot connect at all. This is the default-posture fence when
  per-user permission management is not in play.

DC-DR requires one of the two. Two rules keep the fence from breaking replication:

- **Never fence the federation user.** The fence must target only client users.
  Fencing the user the federation upstream authenticates as (or the operator's
  management user) would break replication along with client publish.
- **Do not federate the fence state.** The fence is per-cluster runtime state (revoked
  permissions or a gated listener). Do not let a synced-definitions or policy import
  path copy the active cluster's client permissions onto the standby and re-grant
  publish there, undoing the fence. The operator manages client permissions per cluster.

## Planned switchover (drained, zero message loss)

Move the active DC on purpose by annotating the RabbitMQ:

```bash
$ kubectl annotate rabbitmq -n demo rm-dcdr dr.kubedb.com/switchover-to=dc-b
```

The hub then:

1. checks the target is a known, healthy DC within the federation lag budget;
2. sets `phase: FailingOver` and quiesces publishers by closing the active cluster's
   publish fence;
3. waits for Federation to drain to near-zero lag, so the target holds every message;
4. flips the AMQP endpoint to the target, opens the target's fence, and reverses the
   federation direction (tears down the old direction's upstream, then sets up the new
   direction's on the other DC's cluster), never both at once;
5. moves the Lease to the target.

Because Federation fully drained before the flip, no confirmed message is lost. This is
a hub-driven annotation, not a `RabbitMQOpsRequest` type: the engine-aware quiesce and
federation drain run in the hub, not in the engine-agnostic `dr-controlplane`.

## Failback

Failback is not a rewind. When a failed DC returns, it becomes the Federation target of
the new active. The messages it accepted but never federated before the failover are a
forked tail RabbitMQ cannot rewind, and because Federation only adds and never deletes,
a naive re-federation leaves those orphan messages on top of the new active's data (and
they could resurface if that DC is ever made active again). For correctness:

- **re-seed the affected queues from the new active** (purge the returned cluster's copy
  and re-federate from scratch), or
- **accept and document the orphan tail** as bounded loss, and make consumers idempotent
  (or apply a dedup window) across the flip.

Once the returned DC is caught up, a drained planned switchover returns the active DC:

```bash
$ kubectl annotate rabbitmq -n demo rm-dcdr dr.kubedb.com/switchover-to=dc-a
```

## Scaling and day-2 operations

The standard `RabbitMQOpsRequest` operations (`UpdateVersion`, `HorizontalScaling`,
`VerticalScaling`, `VolumeExpansion`, `Restart`, `Reconfigure`, `ReconfigureTLS`,
`RotateAuth`) apply to a DC-DR cluster. They act on the per-DC RabbitMQ clusters.
Horizontal scaling operates per DC (each Member DC's cluster scales its own nodes and
handles quorum-queue membership intra-DC), so a scaling request targets the data centers
rather than a single flat node set.

There is no failover ops type: unplanned failover is driven by the Lease, and the
planned switchover is the `dr.kubedb.com/switchover-to` annotation, not an ops request.

> **Note:** the distributed RabbitMQ substrate and the DC-DR layer are net-new for
> RabbitMQ. Treat the field names and flows in this guide as the intended user
> experience; confirm availability in your release before relying on them in production.

## Deletion and cleanup

```bash
$ kubectl delete rabbitmq -n demo rm-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC RabbitMQ clusters, the
operator-managed Federation upstreams and policies, and the cluster-scoped per-DC
`PlacementPolicies` it generated (these carry no owner reference, so the operator
deletes them explicitly). The user-provided base `PlacementPolicy` is left for you to
delete.

## Limitations

- **No zero-RPO on an unplanned failover.** Federation is asynchronous, so an unplanned
  active-DC loss loses the un-federated tail (bounded by federation lag). Use a drained
  planned switchover for a zero-message-loss move.
- **No rewind on failback.** A returned old-active cluster's un-federated forked tail
  cannot be rewound. Re-seed the affected queues or accept the orphan tail as bounded
  loss, and make consumers idempotent across the flip.
- **Two data DCs only.** Active/passive Federation is inherently two-cluster. Three or
  more data DCs (fan-out federation, three-way failover) is a separate, larger design.
- **Cross-DC reachability is required.** RabbitMQ advertises in-cluster `.svc`
  endpoints, so Federation and the cross-DC AMQP endpoint need flat pod networking
  (KubeSlice) or external listeners.
- **Use quorum queues.** Only quorum queues replicate intra-DC and survive node loss;
  classic queues are non-replicated and are not covered by the intra-DC HA guarantee.
