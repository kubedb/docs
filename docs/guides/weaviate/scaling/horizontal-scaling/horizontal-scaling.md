---
title: Scale Weaviate Horizontally
menu:
  docs_{{ .version }}:
    identifier: weaviate-horizontal-scaling-ops
    name: Scale Horizontally
    parent: weaviate-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Weaviate Cluster

This guide will show you how to use the `KubeDB` Ops Manager to scale the number of nodes of a Weaviate cluster up and down.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Horizontal Scaling Overview](/docs/guides/weaviate/scaling/horizontal-scaling/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/scaling/horizontal-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/scaling/horizontal-scaling) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate

In this section, we are going to deploy a Weaviate cluster with `3` nodes.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR and wait for it to become `Ready`:

```bash
$ kubectl get weaviate -n demo
NAME              TYPE                  VERSION   STATUS   AGE
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Ready    5m

$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          5m
weaviate-sample-1   1/1     Running   0          5m
weaviate-sample-2   1/1     Running   0          5m
```

## Scale Up

Here, we are going to scale up the cluster from `3` nodes to `5` nodes.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: weaviate-sample
  horizontalScaling:
    node: 5
```

- `spec.type` specifies that this is a `HorizontalScaling` operation.
- `spec.horizontalScaling.node` specifies the desired number of nodes after scaling.

Let's create the `WeaviateOpsRequest` CR:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/scaling/horizontal-scaling/scale-up.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-scale-up created
```

The Ops Manager adds the new nodes, waits for them to join and sync the schema, and rebalances the shard replicas.

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-scale-up
NAME                TYPE                STATUS       AGE
weaviate-scale-up   HorizontalScaling   Successful   3m

$ kubectl get weaviateopsrequest -n demo weaviate-scale-up -o yaml
...
status:
  conditions:
  - message: Weaviate ops-request has started horizontal scaling
    reason: HorizontalScaling
    status: "True"
    type: HorizontalScaling
  - message: patch petset; ConditionStatus:True
    status: "True"
    type: PatchPetset
  - message: is node in cluster; ConditionStatus:True
    status: "True"
    type: IsNodeInCluster
  - message: successfully restarted new Weaviate pods for schema sync
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - message: successfully rebalanced Weaviate shard replicas
    reason: RebalanceShards
    status: "True"
    type: RebalanceShards
  - message: Successfully scaled up nodes
    reason: HorizontalScaleUp
    status: "True"
    type: HorizontalScaleUp
  - message: successfully reconciled Weaviate with new replica count
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: Horizontal scaling completed
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

Verify the new node count:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          3m51s
weaviate-sample-1   1/1     Running   0          3m11s
weaviate-sample-2   1/1     Running   0          2m31s
weaviate-sample-3   1/1     Running   0          56s
weaviate-sample-4   1/1     Running   0          44s

$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.replicas}'
5
```

## Scale Down

Now, let's scale the cluster back down from `5` nodes to `2` nodes.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: weaviate-sample
  horizontalScaling:
    node: 2
```

Let's create the `WeaviateOpsRequest` CR:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/scaling/horizontal-scaling/scale-down.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-scale-down created
```

The Ops Manager moves the shards off the nodes that are going away before removing them.

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-scale-down
NAME                  TYPE                STATUS       AGE
weaviate-scale-down   HorizontalScaling   Successful   3m

$ kubectl get weaviateopsrequest -n demo weaviate-scale-down -o yaml
...
status:
  conditions:
  - message: Weaviate ops-request has started horizontal scaling
    reason: HorizontalScaling
    status: "True"
    type: HorizontalScaling
  - message: move shards; ConditionStatus:True; PodName:weaviate-sample-4
    status: "True"
    type: MoveShards--weaviate-sample-4
  - message: delete pvc; ConditionStatus:True; PodName:weaviate-sample-4
    status: "True"
    type: DeletePvc--weaviate-sample-4
  - message: move shards; ConditionStatus:True; PodName:weaviate-sample-3
    status: "True"
    type: MoveShards--weaviate-sample-3
  - message: delete pvc; ConditionStatus:True; PodName:weaviate-sample-3
    status: "True"
    type: DeletePvc--weaviate-sample-3
  - message: move shards; ConditionStatus:True; PodName:weaviate-sample-2
    status: "True"
    type: MoveShards--weaviate-sample-2
  - message: delete pvc; ConditionStatus:True; PodName:weaviate-sample-2
    status: "True"
    type: DeletePvc--weaviate-sample-2
  - message: Successfully scaled down nodes
    reason: HorizontalScaleDown
    status: "True"
    type: HorizontalScaleDown
  - message: successfully reconciled Weaviate with new replica count
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: Horizontal scaling completed
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

Verify the node count again:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          7m27s
weaviate-sample-1   1/1     Running   0          6m47s

$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.replicas}'
2
```

The cluster has been scaled horizontally — first up to `5` nodes, then back down to `2`.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- [Vertical Scaling](/docs/guides/weaviate/scaling/vertical-scaling/vertical-scaling.md) of a Weaviate cluster.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete weaviateopsrequest -n demo weaviate-scale-up weaviate-scale-down
$ kubectl delete weaviate -n demo weaviate-sample
$ kubectl delete ns demo
```
