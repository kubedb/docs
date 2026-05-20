---
title: Distributed Deployment
menu:
  docs_{{ .version }}:
    identifier: qdrant-distributed-deployment-overview
    name: Overview
    parent: qdrant-distributed-deployment
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Distributed Deployment

Since version v0.8.0, Qdrant supports a distributed deployment mode where multiple Qdrant services communicate with each other to distribute data across peers, extending storage capabilities and increasing stability. In this mode, Qdrant uses the [Raft](https://raft.github.io/) consensus protocol to maintain consistency of cluster topology and collection structure, while sharding enables horizontal scaling by splitting collections across multiple nodes. Replication further enhances reliability by keeping copies of shards across the cluster.

This tutorial will show you how to deploy a Qdrant database in distributed mode using KubeDB.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/quickstart](/docs/examples/qdrant/quickstart) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Find Available StorageClass

We will need to provide `StorageClass` in the Qdrant CR specification. Check available `StorageClass` in your cluster using the following command:

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  29d
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   26d
standard               rancher.io/local-path   Delete          WaitForFirstConsumer   false                  21h
```

We will use `standard` StorageClass in this tutorial.

## Find Available QdrantVersion

When you install KubeDB, it creates `QdrantVersion` CRDs for all supported Qdrant versions. Let's check available `QdrantVersion`s:

```bash
$ kubectl get qdrantversions
NAME     VERSION   DB_IMAGE                                       DEPRECATED   AGE
1.15.4   1.15.4    docker.io/qdrant/qdrant:v1.15.4-unprivileged                29d
1.16.2   1.16.2    docker.io/qdrant/qdrant:v1.16.2-unprivileged                29d
1.17.0   1.17.0    docker.io/qdrant/qdrant:v1.17.0-unprivileged                29d
```

In this tutorial, we will use `1.17.0` QdrantVersion CR to create a distributed Qdrant cluster.

## Deploy Distributed Qdrant

KubeDB implements a `Qdrant` CRD to define the specification of a Qdrant database. For distributed deployment, you need to set `spec.mode` to `Distributed` and specify the number of replicas.

Below is the `Qdrant` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  mode: Distributed
  replicas: 3
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Here,
- `spec.version` specifies the version of Qdrant to use
- `spec.mode` set to `Distributed` enables distributed mode
- `spec.replicas` specifies the number of Qdrant nodes (default is 1)
- `spec.storage` specifies the storage configuration for each node

Let's create the Qdrant object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
qdrant.kubedb.com/qdrant-sample created
```

## Verify the Deployment

Let's check the status of the Qdrant object:

```bash
$ kubectl get qdrant -n demo
NAME            VERSION   STATUS   AGE
qdrant-sample   1.17.0    Ready    50s
```

To see the distributed nodes, check the pods:

```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=qdrant-sample
NAME              READY   STATUS    RESTARTS   AGE
qdrant-sample-0   1/1     Running   0          48s
qdrant-sample-1   1/1     Running   0          43s
qdrant-sample-2   1/1     Running   0          39s
```

In distributed mode, Qdrant creates a Petset with the specified number of replicas.

## Interact with the Distributed Cluster

Now let's interact with the distributed Qdrant cluster. First, get the API key and forward a port:

```bash
$ kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 -d
F1UxwGOleYzmofu3

$ kubectl port-forward -n demo svc/qdrant-sample 6333:6333 &
```

Create a collection with sharding and replication:

```bash
$ curl -X PUT http://localhost:6333/collections/demo_vectors \
  -H "Content-Type: application/json" \
  -H "api-key: F1UxwGOleYzmofu3" \
  -d '{
    "shard_number": 6,
    "replication_factor": 2,
    "vectors": {
      "size": 8,
      "distance": "Cosine"
    }
  }'
{"result":true,"status":"ok","time":0.912871278}
```

Add some points to the collection:

```bash
$ curl -X PUT "http://localhost:6333/collections/demo_vectors/points?wait=true" \
  -H "Content-Type: application/json" \
  -H "api-key: F1UxwGOleYzmofu3" \
  -d '{
    "points": [
      {"id": 1, "vector": [0.15, 0.22, 0.31, 0.44, 0.51, 0.68, 0.73, 0.89], "payload": {"label": "apple"}},
      {"id": 2, "vector": [0.12, 0.28, 0.35, 0.42, 0.53, 0.64, 0.71, 0.85], "payload": {"label": "banana"}},
      {"id": 3, "vector": [0.18, 0.21, 0.33, 0.46, 0.50, 0.66, 0.77, 0.82], "payload": {"label": "cherry"}},
      {"id": 4, "vector": [0.14, 0.25, 0.32, 0.41, 0.54, 0.63, 0.75, 0.88], "payload": {"label": "date"}},
      {"id": 5, "vector": [0.16, 0.23, 0.38, 0.43, 0.55, 0.61, 0.72, 0.86], "payload": {"label": "elderberry"}}
    ]
  }'
{"result":{"operation_id":1,"status":"completed"},"status":"ok","time":0.002696282}
```

Verify that clustering is enabled by checking the root cluster endpoint:

```bash
$ curl http://localhost:6333/cluster -H "api-key: F1UxwGOleYzmofu3" | jq
{
  "result": {
    "status": "enabled",
    "peer_id": 5887768058245046,
    "peers": {
      "6780901721144166": {
        "uri": "http://qdrant-sample-2.qdrant-sample-pods.demo.svc.cluster.local:6335/"
      },
      "5887768058245046": {
        "uri": "http://qdrant-sample-0.qdrant-sample-pods.demo.svc.cluster.local:6335/"
      },
      "5462954126296684": {
        "uri": "http://qdrant-sample-1.qdrant-sample-pods.demo.svc.cluster.local:6335/"
      }
    },
    "raft_info": {
      "term": 1,
      "commit": 27,
      "pending_operations": 0,
      "leader": 5887768058245046,
      "role": "Leader",
      "is_voter": true
    },
    "consensus_thread_status": {
      "consensus_thread_status": "working"
    },
    "message_send_failures": {}
  },
  "status": "ok"
}
```

The output confirms distributed mode is **enabled** with 3 peers and Raft consensus active.

Now check the collection-level shard distribution:

```bash
$ curl http://localhost:6333/collections/demo_vectors/cluster \
  -H "api-key: F1UxwGOleYzmofu3" | jq
{
  "result": {
    "peer_id": 5887768058245046,
    "shard_count": 6,
    "local_shards": [
      {"shard_id": 0, "points_count": 1, "state": "Active"},
      {"shard_id": 2, "points_count": 1, "state": "Active"},
      {"shard_id": 3, "points_count": 2, "state": "Active"},
      {"shard_id": 5, "points_count": 0, "state": "Active"}
    ],
    "remote_shards": [
      {"shard_id": 0, "peer_id": 6780901721144166, "state": "Active"},
      {"shard_id": 1, "peer_id": 6780901721144166, "state": "Active"},
      {"shard_id": 1, "peer_id": 5462954126296684, "state": "Active"},
      {"shard_id": 2, "peer_id": 5462954126296684, "state": "Active"},
      {"shard_id": 3, "peer_id": 6780901721144166, "state": "Active"},
      {"shard_id": 4, "peer_id": 5462954126296684, "state": "Active"},
      {"shard_id": 4, "peer_id": 6780901721144166, "state": "Active"},
      {"shard_id": 5, "peer_id": 5462954126296684, "state": "Active"}
    ],
    "shard_transfers": []
  },
  "status": "ok"
}
```

The output shows that the collection `demo_vectors` is distributed across 3 peers with 6 shards and a replication factor of 2. Each shard is in `Active` state, and local/remote shards are balanced across the cluster nodes.

## Cleaning Up

To delete the Qdrant database and all associated resources:

```bash
$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted
```

> **Warning:** If you delete the Qdrant object with `deletionPolicy: WipeOut`, all data will be permanently deleted.
