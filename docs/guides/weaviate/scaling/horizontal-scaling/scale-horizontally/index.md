---
title: Scale Weaviate Horizontally
menu:
  docs_{{ .version }}:
    identifier: weaviate-scale-horizontally
    name: Scale Horizontally
    parent: weaviate-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Weaviate Cluster

This guide will show you how to use `KubeDB` Ops Manager to increase/decrease the number of nodes in a `Weaviate` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/weaviate/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/weaviate/scaling/horizontal-scaling/scale-horizontally/yamls](/docs/guides/weaviate/scaling/horizontal-scaling/scale-horizontally/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Horizontal Scaling on Weaviate Cluster

Here, we are going to deploy a `Weaviate` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

#### Prepare Cluster

At first, we are going to deploy a cluster with 3 nodes. Then, we are going to add two additional nodes through horizontal scaling. Finally, we will remove 1 node from the cluster again via horizontal scaling.

**Deploy Weaviate Cluster:**

In this section, we are going to deploy a Weaviate cluster with 3 nodes. Then, in the next section we will scale the cluster using horizontal scaling. Below is the YAML of the `Weaviate` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: "1.26.4"
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

Let's create the `Weaviate` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/weaviate/scaling/horizontal-scaling/scale-horizontally/yamls/weaviate.yaml
weaviate.kubedb.com/weaviate-sample created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `Weaviate` objects using Kubernetes API. When a `Weaviate` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc.

Now, watch `Weaviate` is going to `Running` state and also watch `PetSet` and its pods:

```bash
$ watch -n 3 kubectl get weaviate -n demo weaviate-sample
Every 3.0s: kubectl get weaviate -n demo weaviate-sample

NAME              VERSION   STATUS   AGE
weaviate-sample   1.26.4    Ready    4m40m


$ watch -n 3 kubectl get petset -n demo weaviate-sample
Every 3.0s: kubectl get petset -n demo weaviate-sample

NAME              READY   AGE
weaviate-sample   3/3     4m41m


$ watch -n 3 kubectl get pods -n demo
Every 3.0s: kubectl get pod -n demo

NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          4m25m
weaviate-sample-1   1/1     Running   0          4m26m
weaviate-sample-2   1/1     Running   0          4m26m
```

Let's check the current number of nodes:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o=jsonpath='{.spec.replicas}{"\n"}'
3
```

We are ready to apply the `WeaviateOpsRequest` CR to scale horizontally.

#### Scale Up

Here, we are going to scale up the cluster from 3 nodes to 5 nodes.

**Create WeaviateOpsRequest:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: weaviate-sample
  horizontalScaling:
    node: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling on `weaviate-sample` Weaviate database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.node` specifies the desired number of nodes after scaling.

Let's create the `WeaviateOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/weaviate/scaling/horizontal-scaling/scale-horizontally/yamls/hscale-up.yaml
weaviateopsrequest.ops.kubedb.com/wvops-hscale-up created
```

**Verify Weaviate scale-up completed successfully:**

```bash
$ watch -n 3 kubectl get WeaviateOpsRequest -n demo wvops-hscale-up
Every 3.0s: kubectl get WeaviateOpsRequest -n demo wvops-hscale-up

NAME               TYPE               STATUS       AGE
wvops-hscale-up    HorizontalScaling  Successful   3m57s
```

Now let's verify that the number of nodes has increased:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o=jsonpath='{.spec.replicas}{"\n"}'
5

$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          10m
weaviate-sample-1   1/1     Running   0          10m
weaviate-sample-2   1/1     Running   0          10m
weaviate-sample-3   1/1     Running   0          2m
weaviate-sample-4   1/1     Running   0          1m
```

#### Scale Down

Here, we are going to scale down the cluster from 5 nodes to 4 nodes.

**Create WeaviateOpsRequest:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: weaviate-sample
  horizontalScaling:
    node: 4
```

Let's create the `WeaviateOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/weaviate/scaling/horizontal-scaling/scale-horizontally/yamls/hscale-down.yaml
weaviateopsrequest.ops.kubedb.com/wvops-hscale-down created
```

**Verify Weaviate scale-down completed successfully:**

```bash
$ watch -n 3 kubectl get WeaviateOpsRequest -n demo wvops-hscale-down
Every 3.0s: kubectl get WeaviateOpsRequest -n demo wvops-hscale-down

NAME                 TYPE               STATUS       AGE
wvops-hscale-down    HorizontalScaling  Successful   2m15s
```

Now let's verify that the number of nodes has decreased:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o=jsonpath='{.spec.replicas}{"\n"}'
4

$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          14m
weaviate-sample-1   1/1     Running   0          14m
weaviate-sample-2   1/1     Running   0          14m
weaviate-sample-3   1/1     Running   0          6m
```

We have successfully performed horizontal scaling on the Weaviate cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-sample
kubectl delete WeaviateOpsRequest -n demo wvops-hscale-up wvops-hscale-down
```
