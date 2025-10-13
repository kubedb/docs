---
title: Failover & Disaster Recovery Overview for Solr
menu:
  docs_{{ .version }}:
    identifier: sl-failover-disaster-recovery-solr
    name: Overview
    parent: sl-failover-solr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ensuring Rock-Solid Solr Uptime

## High Availability with KubeDB: Auto-Failover and Disaster Recovery
In today's data-driven landscape, search service downtime is more than just an inconvenience, it can lead to 
critical business disruptions. For teams deploying Apache Solr on Kubernetes, ensuring high availability and 
resilience is crucial. That's where KubeDB comes in a cloud-native database management solution purpose-built for Kubernetes.

One of the standout features of KubeDB is its native support for High Availability (HA) and automated 
coordination for Solr clusters. The KubeDB operator works with ZooKeeper to monitor the health of your Solr 
cluster in real-time. In the event of a node failure, the system automatically redistributes leadership of 
affected shards, ensuring continuous service with minimal disruption.

This article explores how KubeDB handles automated recovery for Solr. You'll learn how to deploy a highly 
available Solr cluster on Kubernetes using KubeDB and then simulate various failure scenarios to observe its 
self-healing mechanisms in action.

By the end of this guide, you'll gain a deeper understanding of how KubeDB ensures that your Solr workloads 
remain highly available, even in the face of failures.

> Unlike traditional databases with primary-secondary architecture, Solr uses ZooKeeper for cluster coordination. Node failures are handled automatically through shard leader election, typically completing within seconds of detecting a failure.
>
> 📌 **Important:** All Solr nodes in a SolrCloud cluster are both **readable and writable**. There is no single “primary” node. Any node can receive indexing (write) or query (read) requests, and Solr automatically routes them to the correct shard leader or replica internally. Data created through any node is replicated to other replicas of the shard to maintain consistency and availability.


## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called 'demo' throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Find Available Solr Versions

When you have installed KubeDB, it has created `SolrVersion` CR for all supported Solr versions. Check available versions by:

```bash
$  kubectl get solrversions
NAME     VERSION   DB_IMAGE                              DEPRECATED   AGE
8.11.4   8.11.4    ghcr.io/appscode-images/solr:8.11.4                27d
9.4.1    9.4.1     ghcr.io/appscode-images/solr:9.4.1                 27d
9.6.1    9.6.1     ghcr.io/appscode-images/solr:9.6.1                 27d
9.7.0    9.7.0     ghcr.io/appscode-images/solr:9.7.0                 27d
9.8.0    9.8.0     ghcr.io/appscode-images/solr:9.8.0                 27d

```

## Deploy a Highly Available Solr Cluster

The KubeDB operator implements a Solr CRD to define the specification of a Solr database.

The KubeDB Solr runs in `solrcloud` mode. Hence, it needs a external zookeeper to distribute replicas among pods and save configurations.

We will use KubeDB ZooKeeper for this purpose.

The ZooKeeper instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zoo-com
  namespace: demo
spec:
  version: 3.8.3
  replicas: 3
  deletionPolicy: WipeOut
  adminServerPort: 8080
  storage:
    resources:
      requests:
        storage: "100Mi"
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
```

We have to apply zookeeper first and wait till atleast pods are running to make sure that a cluster has been formed.

Let's create the ZooKeeper CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/quickstart/overview/yamls/zookeeper/zookeeper.yaml
zooKeeper.kubedb.com/zoo-com created
```

The ZooKeeper's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the database.

```bash
$ kubectl get zookeeper -n demo -w
NAME      TYPE                  VERSION   STATUS   AGE
zoo-com   kubedb.com/v1alpha2   3.8.3     Ready    4d

```

Then we can deploy solr in our cluster.

The Solr instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-ha
  namespace: demo
spec:
  version: 9.4.1
  deletionPolicy: WipeOut
  replicas: 3
  zookeeperRef:
    name: zoo-com
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
    storageClassName: standard
```

Apply the manifest:

```bash
$ kubectl apply -f solr-ha.yaml
solr.kubedb.com/solr-ha created
```

Monitor the status:

```bash
$ kubectl get solr,pods -n demo
NAME                      TYPE                  VERSION   STATUS   AGE
solr.kubedb.com/solr-ha   kubedb.com/v1alpha2   9.4.1     Ready    4d

NAME                          READY   STATUS    RESTARTS       AGE
pod/solr-ha-0                 1/1     Running   2 (103m ago)   3d
pod/solr-ha-1                 1/1     Running   2 (103m ago)   4d
pod/solr-ha-2                 1/1     Running   2 (103m ago)   4d
pod/zoo-com-0                 1/1     Running   2 (103m ago)   4d
pod/zoo-com-1                 1/1     Running   2 (103m ago)   4d
pod/zoo-com-2                 1/1     Running   2 (103m ago)   4d

```

Let's create a collection and add some data to test failover scenarios:

```bash
$ kubectl exec -it -n demo solr-ha-0 -- bash
Defaulted container "solr" out of: solr, init-solr (init)
solr@solr-ha-0:/opt/solr-9.4.1$ alias solr_curl='curl -u admin:c4d0IeGGDO**1h9y'
solr@solr-ha-0:/opt/solr-9.4.1$ solr_curl  "http://localhost:8983/solr/admin/collections?action=CREATE&name=sattriyam&numShards=1&replicationFactor=1&wt=json"
{
  "responseHeader":{
    "status":0,
    "QTime":918
  },
  "success":{
    "solr-ha-1.solr-ha-pods.demo:8983_solr":{
      "responseHeader":{
        "status":0,
        "QTime":226
      },
      "core":"sattriyam_shard1_replica_n1"
    }
  },
  "warning":"Using _default configset. Data driven schema functionality is enabled by default, which is NOT RECOMMENDED for production use. To turn it off: curl http://{host:port}/solr/sattriyam/config -d '{\"set-user-property\": {\"update.autoCreateFields\":\"false\"}}'"
}solr@solr-ha-0:/opt/solr-9.4.1$ solr_curl  "http://localhost:8983/solr/admin/collections?action=LIST&wt=json""
{
  "responseHeader":{
    "status":0,
    "QTime":0
  },
  "collections":["kubedb-system","sattriyam"]
}solr@solr-ha-0:/opt/solr-9.4.1$ 


```
If we check another pod, we can see the collection there as well:

```bash
$ kubectl exec -it -n demo solr-ha-1 -- bash
Defaulted container "solr" out of: solr, init-solr (init)

solr@solr-ha-1:/opt/solr-9.4.1$ alias solr_curl='curl -u admin:c4d0IeGGDO**1h9y'
solr@solr-ha-1:/opt/solr-9.4.1$ solr_curl  "http://localhost:8983/solr/admin/collections?action=LIST&wt=json"
{
  "responseHeader":{
    "status":0,
    "QTime":1
  },
  "collections":["kubedb-system","sattriyam"]
}solr@solr-ha-1:/opt/solr-9.4.1$solr_curl -X POST -H 'Content-type:application/json' \\
"http://localhost:8983/solr/sattriyam/update?commit=true" \
-d '[
  {"id": "1"},
  {"id": "2"},
  {"id": "3"}
]'
{
  "responseHeader":{
    "rf":1,
    "status":0,
    "QTime":98
  }
}solr@solr-ha-1:/opt/solr-9.4.1$solr_curl "http://localhost:8983/solr/sattriyam/select?q=*:*&wt=json&rows=10""
{
  "responseHeader":{
    "zkConnected":true,
    "status":0,
    "QTime":3,
    "params":{
      "q":"*:*",
      "rows":"10",
      "wt":"json"
    }
  },
  "response":{
    "numFound":3,
    "start":0,
    "numFoundExact":true,
    "docs":[{
      "id":"1",
      "_version_":1845866210875408384
    },{
      "id":"2",
      "_version_":1845866210893234176
    },{
      "id":"3",
      "_version_":1845866210893234177
    }]
  }

```
📌 Note: Because every Solr node is both readable and writable, data created in one pod is automatically
replicated to other pods that host replicas of the same shard. If any pod is deleted or fails, 
there will be no data loss as long as other replicas are available, since the data is stored in persistent
volumes and replicated across the cluster.

## Understanding Failover Scenarios

Unlike traditional primary-secondary database setups, Solr's failover behavior works differently because:

1. All nodes are peers - there is no primary/secondary relationship

2. ZooKeeper manages cluster state and shard leadership

3. Each shard has a leader, but leaders are elected per shard

4. Any node can serve queries for any shard

5. Writes can be sent to any node, and Solr automatically forwards the updates to the correct shard leader and then replicates to other replicas.


Let's explore different failure scenarios:

### Scenario 1: Single Node Failure

When a Solr node fails:
1. ZooKeeper detects the node failure
2. For any shards where the failed node was leader:
   - New leader is elected from remaining replicas
   - Queries are redirected automatically
3. Cluster continues serving requests through remaining nodes

Let's simulate by deleting a pod:

```bash
$ kubectl delete pod -n demo solr-ha-0
pod "solr-ha-0" deleted
```

Watch the recovery:

```bash
$ watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
```shell
solr-ha-0
solr-ha-1
solr-ha-2
zoo-com-0
zoo-com-1
zoo-com-2
```

During this process:
- The cluster remains available
- Queries continue being served
- Data remains accessible through replicas
- The failed node automatically recovers
Let's verify the collection is still accessible:

```bash
$ $ kubectl exec -it -n demo solr-ha-1 -- bash
Defaulted container "solr" out of: solr, init-solr (init)

solr@solr-ha-1:/opt/solr-9.4.1$ alias solr_curl='curl -u admin:c4d0IeGGDO**1h9y'
solr@solr-ha-1:/opt/solr-9.4.1$ solr_curl  "http://localhost:8983/solr/admin/collections?action=LIST&wt=json"
{
  "responseHeader":{
    "status":0,
    "QTime":1
  },
  "collections":["kubedb-system","sattriyam"]
}solr@solr-ha-1:/opt/solr-9.4.1$solr_curl -X POST -H 'Content-type:application/json' \\
"http://localhost:8983/solr/sattriyam/update?commit=true" \
-d '[
  {"id": "1"},
  {"id": "2"},
  {"id": "3"}
]'
{
  "responseHeader":{
    "rf":1,
    "status":0,
    "QTime":98
  }
}solr@solr-ha-1:/opt/solr-9.4.1$solr_curl "http://localhost:8983/solr/sattriyam/select?q=*:*&wt=json&rows=10""
{
  "responseHeader":{
    "zkConnected":true,
    "status":0,
    "QTime":3,
    "params":{
      "q":"*:*",
      "rows":"10",
      "wt":"json"
    }
  },
  "response":{
    "numFound":3,
    "start":0,
    "numFoundExact":true,
    "docs":[{
      "id":"1",
      "_version_":1845866210875408384
    },{
      "id":"2",
      "_version_":1845866210893234176
    },{
      "id":"3",
      "_version_":1845866210893234177
    }]
  }
```

### Scenario 2: Multiple Node Failure

Even with multiple node failures, Solr remains available as long as:
- At least one replica of each shard is available
- ZooKeeper ensemble remains functional

Let's simulate multiple failures:

```bash
$ kubectl delete pod -n demo solr-ha-0 solr-ha-1
pod "solr-ha-0" deleted
pod "solr-ha-1" deleted
```
Watch the recovery:

```bash
$ watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
```shell
solr-ha-0
solr-ha-1
solr-ha-2
zoo-com-0
zoo-com-1
zoo-com-2
```

### Scenario 3: Full Cluster Recovery

In case all nodes fail:

```bash
$ kubectl delete pod -n demo solr-ha-0 solr-ha-1 solr-ha-2
```

The cluster will recover automatically, but full availability requires:
- All nodes to restart
- Data recovery from persistent storage
- ZooKeeper ensemble to be functional
- Leader election for all shards

## Disaster Recovery

For disaster recovery, KubeDB supports:

1. **Backup & Restore**: Regular backups using Stash
2. **Volume Snapshots**: For point-in-time recovery
3. **Cross-cluster Replication**: For geographic redundancy

### Handling Storage Issues

If a Solr node's storage becomes full:

1. Use Volume Expansion to recover:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: solr-ops-volume-expand
  namespace: demo
spec:
  type: VolumeExpansion
  volumeExpansion:
    mode: Online
    solr: 20Gi
  databaseRef:
    name: solr-ha
```

## Cleanup

```bash
$ kubectl delete solr -n demo solr-ha
$ kubectl delete ns demo
```

