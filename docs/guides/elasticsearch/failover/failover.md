---
title: Elasticsearch Failover and DR Scenarios
menu:
  docs_{{ .version }}:
    identifier: elasticsearch-failover
    name: Overview
    parent: guides-elasticsearch-fdr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Exploring Fault Tolerance in Elasticsearch with KubeDB

## Understanding Failover in Elasticsearch on KubeDB

`Failover` refers to the process of automatically recovering cluster coordination and data availability
when one or more nodes fail. In high-availability database systems, failover ensures that services remain
uninterrupted even when one or more nodes go down. This capability is critical in modern, cloud-native
infrastructure where downtime can lead to major disruptions.

A dedicated **topology cluster** has three node roles — `master`, `data`, and `ingest` — and KubeDB sets up
recovery for all three automatically, though only two of them involve an actual election/promotion:

- **Cluster-manager (master) election:**
  A subset of nodes are marked `master-eligible`. Exactly one of them is elected as the **active master**,
  which owns cluster-wide bookkeeping — index creation/deletion, shard allocation decisions, and cluster
  membership. Master-eligible nodes use a quorum-based (Raft-inspired) consensus protocol, so a **majority**
  of the configured master-eligible nodes must be reachable to elect (or keep) an active master. This is why
  production clusters should keep an **odd** master-eligible count (`3`, `5`, ...) — with `3` nodes, any `2`
  form a majority, so the cluster tolerates losing `1` without losing quorum. An **even** count like `2` gives
  you a second copy of cluster state, but genuinely **no fault tolerance** — losing either node already drops
  you below a majority of `2`. The demo cluster in this guide uses `2` master-eligible nodes to keep it
  lightweight, and Case 2 below demonstrates exactly this failure mode.

- **Data node (shard) failover:**
  Every index's data is split into shards, and each shard can have one or more replica copies stored on
  different `data` nodes. If a data node holding a primary shard goes down, the elected master promotes an
  in-sync replica of that shard to primary on a surviving node, and schedules a new replica to be built once
  capacity is available. This happens independently, per shard, and doesn't depend on master-node count.

- **Ingest node recovery:**
  Ingest nodes are stateless — they don't hold shard data and aren't master-eligible, so there's nothing to
  elect or promote. If an ingest pod goes down, in-flight requests through it fail, but the client
  (`es-topology`) service simply routes subsequent traffic to the remaining ingest pods. KubeDB restarts the
  pod, and since it carries no data, it rejoins the pool as soon as it's `Running` — no reconfiguration
  needed.

In the rest of this guide, we'll deploy a dedicated topology cluster (separate `master`, `data`, and
`ingest` nodes) and walk through master and data-node failover in detail.

## Before You Begin

Before proceeding:

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to
  communicate with your cluster. If you do not already have a cluster, you can create one by
  using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- Read the [Elasticsearch topology cluster concept](/docs/guides/elasticsearch/clustering/_index.md) to learn about dedicated node roles.

- To keep things isolated, this tutorial uses a separate namespace called `es-demo` throughout this tutorial.
  Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns es-demo
  namespace/es-demo created
  ```

## Deploy Elasticsearch Dedicated Topology Cluster

The following is an example `Elasticsearch` object which creates a dedicated topology cluster with 2
master-eligible nodes, 3 data nodes, and 2 ingest nodes.

>note: 2 master-eligible nodes is used here to keep the demo light. As explained above, this gives **zero**
> tolerance for a single master-node failure — production clusters should use an odd count of `3` or more.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-topology
  namespace: es-demo
spec:
  version: xpack-9.2.3
  topology:
    master:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: local-path
    data:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: local-path
    ingest:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: local-path
  storageType: Durable
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f es-topology.yaml
elasticsearch.kubedb.com/es-topology created
```

Here,

- `spec.topology.master.replicas` — number of master-eligible nodes.
- `spec.topology.data.replicas` — number of data nodes, which store primary and replica shards.
- `spec.topology.ingest.replicas` — number of ingest nodes, used for request routing and pre-processing pipelines.

KubeDB operator watches for `Elasticsearch` objects using the Kubernetes API. When one is created, the
operator creates a PetSet per node role, plus Services for client and inter-node communication.

You can monitor the status until all pods are ready:

```bash
$ watch kubectl get elasticsearch,petset,pods -n es-demo
```

```bash
$ kubectl get elasticsearch,petset,pods -n es-demo
NAME                                   VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-topology   xpack-9.2.3   Ready    16m

NAME                                              AGE
petset.apps.k8s.appscode.com/es-topology-data     16m
petset.apps.k8s.appscode.com/es-topology-ingest   16m
petset.apps.k8s.appscode.com/es-topology-master   16m

NAME                       READY   STATUS    RESTARTS   AGE
pod/es-topology-data-0     1/1     Running   0          16m
pod/es-topology-data-1     1/1     Running   0          12m
pod/es-topology-data-2     1/1     Running   0          11m
pod/es-topology-ingest-0   1/1     Running   0          16m
pod/es-topology-ingest-1   1/1     Running   0          12m
pod/es-topology-master-0   1/1     Running   0          16m
pod/es-topology-master-1   1/1     Running   0          12m
```

```bash
$ kubectl get svc -n es-demo
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
es-topology           ClusterIP   10.96.67.225   <none>        9200/TCP   4m2s
es-topology-master    ClusterIP   None           <none>        9300/TCP   4m2s
es-topology-pods      ClusterIP   None           <none>        9200/TCP   4m2s
```

## Verify Cluster Reachability and Node Roles

Port-forward the client service and export credentials as shown in the
[topology cluster guide](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md#connect-with-elasticsearch-database).

```bash
$ kubectl port-forward -n es-demo svc/es-topology 9200   # in one terminal

# in another terminal
$ export ES_USER=$(kubectl get secret -n es-demo es-topology-auth -o jsonpath='{.data.username}' | base64 -d)
$ export ES_PASS=$(kubectl get secret -n es-demo es-topology-auth -o jsonpath='{.data.password}' | base64 -d)
```

List every node along with its roles, and see which one is currently elected master (marked with `*`):

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cat/nodes?v&h=name,node.role,master"
name                 node.role master
es-topology-ingest-0 i         -
es-topology-data-0   d         -
es-topology-data-2   d         -
es-topology-master-1 m         *
es-topology-ingest-1 i         -
es-topology-master-0 m         -
es-topology-data-1   d         -
```

You can also ask directly who the active master is:

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cat/master?v"
id                     host        ip          node
i0hv_X6NQy6raWiWt2CE6A 10.42.0.210 10.42.0.210 es-topology-master-1
```

`es-topology-master-1` is currently the active master among the two master-eligible nodes.

## Insert Data and Check Shard Distribution

Create an index with 1 primary shard and 2 replicas, so every data node holds a copy.

```bash
$ curl -s -u "$ES_USER:$ES_PASS" -XPUT "http://localhost:9200/info?pretty" -H 'Content-Type: application/json' -d'
{
  "settings": { "number_of_shards": 1, "number_of_replicas": 2 }
}'
{
  "acknowledged" : true,
  "shards_acknowledged" : true,
  "index" : "info"
}
```

Index one document with an explicit `_id` of `1` (using `PUT`, not `POST`, which would auto-generate a
random ID instead) — this way, we can reliably look the document back up at `/info/_doc/1` later.

```bash
$ curl -s -u "$ES_USER:$ES_PASS" -XPUT "http://localhost:9200/info/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "Company": "AppsCode Inc",
  "Product": "KubeDB"
}'
{
  "_index" : "info",
  "_id" : "1",
  "_version" : 1,
  "result" : "created",
  "_shards" : {
    "total" : 3,
    "successful" : 3,
    "failed" : 0
  },
  "_seq_no" : 0,
  "_primary_term" : 1
}
```

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cat/shards/info?v"
index shard prirep state   docs store ip          node
info  0     p      STARTED    1  4kb 10.42.0.211 es-topology-data-1
info  0     r      STARTED    1  4kb 10.42.0.205 es-topology-data-0
info  0     r      STARTED    1  4kb 10.42.0.214 es-topology-data-2
```

The primary (`p`) shard lives on `es-topology-data-1`, with replicas (`r`) on the other two data nodes.

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cluster/health?pretty" | grep '"status"'
  "status" : "green",
```

## How Failover Works

Elasticsearch runs two independent election-based failover flows, plus stateless recovery for ingest nodes:

**Master election.** Master-eligible nodes exchange heartbeats. If the active master stops responding
(crash, network partition, or graceful shutdown), the remaining master-eligible nodes detect the loss and
try to elect a new one — but a candidate only wins with votes from a **majority** of the configured
master-eligible nodes. With `2` master-eligible nodes (as in this demo), that majority is `2` — so losing
*either one* already means no master can be elected until it returns. With `3` (the production-recommended
minimum), any `2` form a majority, so a single node loss is tolerated. Either way, when quorum is lost,
cluster-management operations (index creation, shard allocation, settings changes) stall — this is a
deliberate safety measure to avoid split-brain, not a bug.

**Data node failover.** The active master continuously tracks which data nodes hold in-sync copies of each
shard. If a data node goes down, any primary shards it held are promoted from one of their in-sync replicas
on a surviving data node — this is near-instantaneous since the replica already has the data. Cluster health
transitions to `yellow` (some replicas unavailable) while reallocation is worked out, and back to `green`
once enough replicas are rebuilt. If a node holding the **only** copy of a shard goes down, health goes to
`red` for that shard until the node returns.

**Ingest node recovery.** No election happens here — ingest nodes hold no data and aren't master-eligible.
Losing one just means fewer nodes to spread ingest-pipeline/coordinating traffic across until KubeDB
restarts the pod; existing indices and cluster health are unaffected.

### Hands-on Failover Testing

#### Case 1: Delete the data node holding the primary shard

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cat/shards/info?v"
index shard prirep state   docs store dataset ip          node
info  0     r      STARTED    1 5.6kb   5.6kb 10.42.0.205 es-topology-data-0
info  0     r      STARTED    1 5.6kb   5.6kb 10.42.0.214 es-topology-data-2
info  0     p      STARTED    1 5.6kb   5.6kb 10.42.0.211 es-topology-data-1

$ kubectl delete pod -n es-demo es-topology-data-1
pod "es-topology-data-1" deleted
```

The master promotes one of the in-sync replicas to primary almost immediately:

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cluster/health?pretty" | grep '"status"'
  "status" : "yellow",

$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cat/shards/info?v"
index shard prirep state      docs store dataset ip          node
info  0     r      STARTED       2  11kb    11kb 10.42.0.214 es-topology-data-2
info  0     p      STARTED       2  11kb    11kb 10.42.0.211 es-topology-data-1
info  0     r      UNASSIGNED
```

Once `es-topology-data-1` returns, it's reused to host a fresh replica copy, and the cluster returns to
`green`. The document survives the promotion:

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/info/_doc/1?pretty" | grep -A2 _source
  "_source" : {
    "Company" : "AppsCode Inc",
    "Product" : "KubeDB"
```

#### Case 2: Delete a master-eligible node

This demo cluster has only 2 master-eligible nodes, so the majority needed to elect or keep an active
master is `2` — both must be reachable. There's no spare to lose here; deleting either one already breaks
quorum.

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cat/master?v"
id                     host        ip          node
i0hv_X6NQy6raWiWt2CE6A 10.42.0.210 10.42.0.210 es-topology-master-1

$ kubectl delete pod -n es-demo es-topology-master-1 es-topology-master-0
pod "es-topology-master-1" deleted
pod "es-topology-master-0" deleted
```

With `0` of `2` master-eligible nodes left, the cluster can't elect a master:

```bash
$ curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cat/master?v"
{"error":{"root_cause":[{"type":"null_pointer_exception","reason":"Cannot invoke \"Object.hashCode()\" because \"pk\" is null"}],"type":"null_pointer_exception","reason":"Cannot invoke \"Object.hashCode()\" because \"pk\" is null"},"status":500}
```

Master-dependent operations (index creation, shard allocation, cluster settings changes) stall until quorum
returns. Once both `es-topology-master-0` and `es-topology-master-1` are back up, the 2-node majority is
restored and a master is elected again — bringing back only one of the two is **not** enough, since a
majority of `2` still requires both.

This is exactly why production topology clusters should run an **odd** number of master-eligible nodes —
`3` or more — so a single node failure still leaves a majority standing.

#### Case 3: Delete every node in the cluster

```bash
$ kubectl delete pod -n es-demo -l app.kubernetes.io/instance=es-topology
pod "es-topology-master-0" deleted
pod "es-topology-master-1" deleted
pod "es-topology-data-0" deleted
pod "es-topology-data-1" deleted
pod "es-topology-data-2" deleted
pod "es-topology-ingest-0" deleted
pod "es-topology-ingest-1" deleted
```

As the PetSets bring pods back up (each reattaching its own PVC), the master-eligible nodes rediscover each
other, elect a master once both are reachable again, and data nodes reattach their shards from disk. Cluster
health transitions `red` → `yellow` → `green` as shards are recovered from local data first, and any
missing replicas are rebuilt.

```bash
$ watch curl -s -u "$ES_USER:$ES_PASS" "http://localhost:9200/_cluster/health?pretty"
```

## CleanUp

For cleaning up what we created in this tutorial follow the following commands:

```bash
$ kubectl patch -n es-demo elasticsearch es-topology -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
elasticsearch.kubedb.com/es-topology patched

$ kubectl delete elasticsearch -n es-demo es-topology
elasticsearch.kubedb.com "es-topology" deleted

$ kubectl delete ns es-demo
namespace "es-demo" deleted
```

## Next Steps

- Learn about [backup and restore](/docs/guides/elasticsearch/backup/stash/overview/index.md) Elasticsearch database using Stash.
- Want to explore other topologies? Check the [Elasticsearch clustering guides](/docs/guides/elasticsearch/clustering/_index.md).
- Monitor your Elasticsearch database with KubeDB using [built-in Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
