---
title: Scale Qdrant Horizontally
menu:
  docs_{{ .version }}:
    identifier: qdrant-scale-horizontally
    name: Scale Horizontally
    parent: qdrant-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Qdrant Cluster

This guide will show you how to use `KubeDB` Ops Manager to increase/decrease the number of nodes in a `Qdrant` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/qdrant/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/qdrant/scaling/horizontal-scaling/scale-horizontally/yamls](/docs/guides/qdrant/scaling/horizontal-scaling/scale-horizontally/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Horizontal Scaling on Qdrant Cluster

Here, we are going to deploy a `Qdrant` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

#### Prepare Cluster

At first, we are going to deploy a cluster with 3 nodes. Then, we are going to add two additional nodes through horizontal scaling. Finally, we will remove 1 node from the cluster again via horizontal scaling.

**Deploy Qdrant Cluster:**

In this section, we are going to deploy a Qdrant cluster with 3 nodes. Then, in the next section we will scale the cluster using horizontal scaling. Below is the YAML of the `Qdrant` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/scaling/horizontal-scaling/scale-horizontally/yamls/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `Qdrant` objects using Kubernetes API. When a `Qdrant` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc.

Now, watch `Qdrant` is going to `Running` state and also watch `PetSet` and its pods:

```bash
$ watch -n 3 kubectl get qdrant -n demo qdrant-sample
Every 3.0s: kubectl get qdrant -n demo qdrant-sample

NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    4m40m


$ watch -n 3 kubectl get petset -n demo qdrant-sample
Every 3.0s: kubectl get petset -n demo qdrant-sample

NAME              READY   AGE
qdrant-sample     3/3     4m41m


$ watch -n 3 kubectl get pods -n demo
Every 3.0s: kubectl get pod -n demo

NAME                READY   STATUS    RESTARTS   AGE
qdrant-sample-0     1/1     Running   0          4m25m
qdrant-sample-1     1/1     Running   0          4m26m
qdrant-sample-2     1/1     Running   0          4m26m
```

Let's check the current number of nodes:

```bash
$ kubectl get qdrant -n demo qdrant-sample -o=jsonpath='{.spec.replicas}{"\n"}'
3
```

We are ready to apply the `QdrantOpsRequest` CR to scale horizontally.

#### Scale Up

Here, we are going to scale up the cluster from 3 nodes to 5 nodes.

**Create QdrantOpsRequest:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: qdrant-sample
  horizontalScaling:
    node: 5
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling on `qdrant-sample` Qdrant database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.node` specifies the desired number of nodes after scaling.
- `spec.timeout` specifies the timeout for the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#spectimeout)).
- `spec.apply` specifies when to apply the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#specapply)).

Let's create the `QdrantOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/scaling/horizontal-scaling/scale-horizontally/yamls/hscale-up.yaml
qdrantopsrequest.ops.kubedb.com/qdops-hscale-up created
```

**Verify Qdrant scale-up completed successfully:**

```bash
$ watch -n 3 kubectl get QdrantOpsRequest -n demo qdops-hscale-up
Every 3.0s: kubectl get QdrantOpsRequest -n demo qdops-hscale-up

NAME               TYPE               STATUS       AGE
qdops-hscale-up    HorizontalScaling  Successful   3m57s
```

Now let's verify that the number of nodes has increased:

```bash
$ kubectl get qdrant -n demo qdrant-sample -o=jsonpath='{.spec.replicas}{"\n"}'
5

$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
qdrant-sample-0     1/1     Running   0          10m
qdrant-sample-1     1/1     Running   0          10m
qdrant-sample-2     1/1     Running   0          10m
qdrant-sample-3     1/1     Running   0          2m
qdrant-sample-4     1/1     Running   0          1m
```

#### Scale Down

Here, we are going to scale down the cluster from 5 nodes to 4 nodes.

**Create QdrantOpsRequest:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: qdrant-sample
  horizontalScaling:
    node: 4
```

Let's create the `QdrantOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/scaling/horizontal-scaling/scale-horizontally/yamls/hscale-down.yaml
qdrantopsrequest.ops.kubedb.com/qdops-hscale-down created
```

**Verify Qdrant scale-down completed successfully:**

```bash
$ watch -n 3 kubectl get QdrantOpsRequest -n demo qdops-hscale-down
Every 3.0s: kubectl get QdrantOpsRequest -n demo qdops-hscale-down

NAME                 TYPE               STATUS       AGE
qdops-hscale-down    HorizontalScaling  Successful   2m15s
```

Now let's verify that the number of nodes has decreased:

```bash
$ kubectl get qdrant -n demo qdrant-sample -o=jsonpath='{.spec.replicas}{"\n"}'
4

$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
qdrant-sample-0     1/1     Running   0          14m
qdrant-sample-1     1/1     Running   0          14m
qdrant-sample-2     1/1     Running   0          14m
qdrant-sample-3     1/1     Running   0          6m
```

We have successfully performed horizontal scaling on the Qdrant cluster.

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete QdrantOpsRequest -n demo qdops-hscale-up qdops-hscale-down
```