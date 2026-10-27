---
title: YugabyteDB Read Replicas
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-clustering-read-replicas
    name: Read Replicas
    parent: yugabytedb-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Read Replicas

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

A read replica is an optional second group of tservers, `spec.topology.readReplica`. Its tservers hold **non-voting** copies of every tablet in their own placement (`placementUUID`), so they never take part in writes or in leader elections. They add read capacity, or put read copies close to remote clients, without changing the write quorum.

## Topology

- The read replica pods run in their own PetSet, `<name>-readreplica`, next to `<name>-master` and `<name>-tserver`.
- `readReplica` takes the same node fields as the tservers (`replicas`, `podTemplate`, `storageType`, `storage`, `dataVolumes`), plus `placementUUID`, `region`, `zone` and `replicationFactor`.
- `readReplica.replicationFactor` is the **total number of read copies of every tablet** (default 1). It must not exceed the number of read replica pods.
- `placementUUID` defaults to `rr` and cannot be changed once the read replica exists. YugabyteDB needs the read replica zones to differ from the zones of the primary tservers.
- `region` defaults to `spec.placement.region`. `zone` overrides the zone; by default the zone of every pod is read from its node.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: YugabyteDB
metadata:
  name: yb-rr
  namespace: demo
spec:
  version: 2025.2.8
  replicationFactor: 3
  topology:
    tserver:
      replicas: 3
    readReplica:
      replicas: 2
      replicationFactor: 2
      placementUUID: rr
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/clustering/read-replica/yugabytedb-read-replica.yaml
```

The operator registers the read replica placement and its replication factor on the masters (the `SetReadReplicaPlacement` step).

## Follower reads

Reads on a read replica are only allowed when the session accepts stale data. In YSQL, enable follower reads in the session (or per transaction):

```sql
SET yb_read_from_followers = true;
SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY;
```

With `yb_read_from_followers` on, a read-only transaction can be served by a follower or a read replica close to the client. The data can lag behind the leader by a bounded staleness (YugabyteDB parameter `yb_follower_read_staleness_ms`). Connect to a read replica tserver (for example the pod `yb-rr-readreplica-0`), not through the primary tservers' Service, to read from it. Whether a given connection really reads from the read replica was not observed.

## Scaling

Use `horizontalScaling.readReplica`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-scale-read-replica
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: yb-rr
  horizontalScaling:
    readReplica: 4
```

The webhook enforces:

- the database must have `spec.topology.readReplica`;
- the target must be `0` or at least the read replica `replicationFactor`;
- the target must differ from the current number of replicas;
- the masters cannot be scaled at all, and `horizontalScaling.tserver` must be at least the replication factor.

Scaling in runs the `DrainReadReplicas` step first, which moves the tablet copies off the read replicas that are being removed; scaling out runs `ScaleReadReplicas`.

Other operations that take a read replica role: `verticalScaling.readReplica`, `volumeExpansion.readReplica`, and a rolling restart runs `RollReadReplicas` after the tservers. The `spec.topology.readReplica.placementUUID` cannot be changed, and the `readReplica` block cannot be added to or removed from an existing database through the GitOps operator.

## Cleaning up

```bash
$ kubectl delete yb -n demo yb-rr
```
