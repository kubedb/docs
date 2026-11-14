---
title: NATS Across Data Centers
menu:
  docs_{{ .version }}:
    identifier: nats-multi-dc
    name: Multi-DC
    parent: nats-guides
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# NATS Across Data Centers

**A NATS cluster never spans data centers.** Each DC runs its own NATS cluster with a full route mesh. Clusters in different DCs are joined by **gateways** into a *super-cluster*, and edge sites that can only dial out attach as **leaf nodes**. A gateway keeps one connection per remote cluster and forwards only the messages that have remote interest, which is why this scales across a WAN better than stretching one route mesh.

Every DC is active. There is no primary DC and no failover step: clients talk to their local DC, and the gateways carry messages to the DCs that have subscribers.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/multi-dc](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/multi-dc) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Two ways to build a super-cluster

| | Distributed mode | Hand-built gateways |
|---|---|---|
| You create | one `NATS` object with `spec.distributed: true` and a `PlacementPolicy` | one `NATS` object per DC, each with `spec.gateway` |
| KubeDB creates | a PetSet, governing Service and client Service per DC; gateway remotes; the Arbiter DC | only what a single `NATS` normally gets |
| Needs | OCM hub and spokes, KubeSlice | network reachability between the gateway ports |

## Gateways between separate NATS objects

Give each cluster a gateway whose name equals its cluster name (the object name). The server refuses to start otherwise. Gateways gossip, so a few remotes are enough to seed discovery.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: dc-a
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 3
  gateway:
    name: dc-a
    remotes:
      - name: dc-b
        urls:
          - nats://dc-b-pods.demo.svc.slice.local:7222
  deletionPolicy: WipeOut
```

A remote may instead use `natsRef` to point at another `NATS` object through its AppBinding, so KubeDB fills in the URLs.

- Gateway configuration is not reloadable. Changing a configured remote is a rolling restart of that cluster.
- **Adding a DC does not restart the existing DCs.** The new DC lists any one existing DC, and the existing clusters learn about it by gossip. They keep their original list until their next planned roll. KubeDB never sets `rejectUnknown`; set `gateway.rejectUnknown: true` only if you want a closed super-cluster and accept the restart.
- The gateway uses the `server` certificate when `spec.tls` is set (`gatewayTLS`, default on). Gateway TLS material reloads in place.
- Gateways use a generated shared credential, `<name>-gateway-auth`. In Operator mode, gateways between clusters of the same operator need no JWT work.

## Distributed mode: one cluster per DC from one object

Distributed mode needs a kubeops.dev `PlacementPolicy` with a `clusterSpreadConstraint`. It requires KubeSlice (`slice` is mandatory) and must **not** set `failoverPolicy`, because NATS has no primary DC.

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: nats-3dc
spec:
  clusterSpreadConstraint:
    slice:
      projectNamespace: kubeslice-demo
      sliceName: demo-slice
    distributionRules:
      - clusterName: dc-a
        replicaIndices: [0, 1, 2]
        role: Member
      - clusterName: dc-b
        replicaIndices: [3, 4, 5]
        role: Member
      - clusterName: dc-c
        replicaIndices: []
        role: Arbiter
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-dc
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 6
  distributed: true
  jetStream:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 10Gi
  arbiter:
    replicas: 1
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
  podTemplate:
    spec:
      podPlacementPolicy:
        name: nats-3dc
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/multi-dc/placement-policy.yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/multi-dc/nats-distributed.yaml
```

For each `Member` rule KubeDB creates:

- a PetSet `<name>-<dc>` with its own derived placement policy, so each DC is an ordinary cluster. `cluster.name` and `gateway.name` equal the DC name, and routes list only that DC's pods.
- a governing Service `<name>-pods-<dc>` and a client Service `<name>-<dc>`. Pods are exported over the KubeSlice network on ports 6222, 7222, 7422 and 8222.
- gateway remotes: every DC lists the other DCs. Cross-DC names use `*.svc.slice.local`.

Do not set `spec.gateway.name` when `distributed` is true; each DC's gateway is named after the DC.

### Clients stay local

Clients learn only the servers of their own cluster, so a client cannot discover another DC on its own. Give the application an explicit list:

```
nats://nats-dc-dc-a.demo.svc:4222,nats://nats-dc-dc-b.demo.svc:4222
```

The AppBinding of a distributed `NATS` lists every DC Service in `spec.clientConfig`, so KubeStash and other tools see them all.

## The Arbiter DC and why you need three DCs

JetStream has **one meta group over every server in the super-cluster**. Gateways always merge the groups. A whole-DC loss is survivable only if a majority of meta peers remains, which needs at least three DCs.

- A gateway-joined `NATS` with JetStream needs an **odd number of DCs counting any Arbiter DC, and at least 3**. The webhook enforces this.
- With **two data DCs, add an Arbiter DC**: an `Arbiter` rule with *empty* `replicaIndices`. KubeDB runs `spec.arbiter.replicas` (default 1) servers there in their own PetSet `<name>-<dc>-arbiter`, with a small PVC (default `1Gi`). They are tagged `!jetstream`, so they vote in the meta group but hold no stream data, serve no clients and open no leaf port. Two DCs of three servers plus one arbiter make 7 meta peers; losing a data DC leaves 4 of 7, so stream and consumer management continues.
- An arbiter can still be elected meta-leader. KubeDB steps the leader down to a data DC after every ops request and when its health checker sees the leader on an arbiter. This is a preference, not a guarantee.
- **No third site?** Connect the two DCs with leaf nodes and a distinct JetStream domain per DC (each DC keeps its own meta group, data crosses by mirror or source) rather than gateways. The webhook requires this for two data DCs with JetStream and no Arbiter.

50 ms between DCs is fine: Raft heartbeats are 1 s and elections 4 to 9 s. KubeDB reads the gateway round-trip time per remote and raises a warning condition above 100 ms.

### Placing streams

Create each stream inside one DC so its commits stay at LAN latency:

```bash
$ nats stream add ORDERS --subjects 'orders.*' --replicas 3 --cluster dc-a
```

Copy a stream to another DC with a **mirror** or **source**. These are NATS-native asynchronous replication, declared as stream configuration rather than KubeDB API:

```bash
$ nats stream add ORDERS_DR --mirror ORDERS --replicas 3 --cluster dc-b
```

To promote the copy after losing `dc-a`, edit the mirror stream and remove its `mirror` setting so it accepts writes. Mirrors are asynchronous, so the replication lag is the data you can lose.

## Leaf nodes

A leaf node attaches an edge site to a hub without being part of its mesh. The leaf dials out, so the hub needs no inbound route to the edge.

Hub side, in any DC's cluster:

```yaml
spec:
  leafNodes:
    accept: true
  serviceTemplates:
    - alias: leafnodes
      spec:
        type: LoadBalancer
```

Port 7422 opens on every pod and on the client Service. Its type comes from the `leafnodes` alias (default `ClusterIP`). Choose `LoadBalancer` so an edge site can reach it.

Leaf side, another `NATS` object, usually one replica:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-edge
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 1
  leafNodes:
    jetStreamDomain: edge
    remotes:
      - natsRef:
          name: nats-hub
          namespace: demo
        account: APP
  deletionPolicy: WipeOut
```

- A remote has `urls` or `natsRef`, an optional `account`, an optional `credentialsSecret` (a Secret with a `creds` key, Operator mode) and an optional JetStream `domain`.
- In Config mode the leaf binds the default account unless `account` is set, and then the hub must define the same account. In Operator mode the leaf presents its credentials and the hub attaches it to the matching account. A leaf on the default account has no subject boundary, so the webhook warns about it.
- **A leaf that runs its own JetStream needs `jetStreamDomain`.** The webhook requires it whenever JetStream is on and `leafNodes.remotes` is set, otherwise the leaf would extend the hub's JetStream.
- Changing a remote's URLs, account or credentials path reloads live as remove-plus-add. Rotating the contents of the same `.creds` file needs nothing, because the next reconnect reads it. The hub's own `leafnodes` block (listen, authorization users) does not reload, so changing hub leaf credentials is a rolling restart.
- A leaf's local JetStream is backed up like any `NATS`; the hub's copy is independent.

## Day-2 operations in multi-DC

Ops requests run inside one DC at a time and one pod at a time, never two DCs concurrently. The Arbiter DC goes first and the DC holding the meta leader goes last. See [ops requests](/docs/guides/nats/ops-requests.md#multi-dc-clusters).

## Cleanup

```bash
$ kubectl delete -n demo nats/nats-dc nats/nats-hub nats/nats-edge nats/dc-a
$ kubectl delete placementpolicy nats-3dc
```
