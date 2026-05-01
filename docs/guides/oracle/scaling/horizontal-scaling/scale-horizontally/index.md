---
title: Scale Oracle Horizontally
menu:
  docs_{{ .version }}:
    identifier: oracle-scale-horizontally
    name: Scale Horizontally
    parent: oracle-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Oracle Cluster

This guide will show you how to use `KubeDB` Ops Manager to increase/decrease the number of nodes in a `Oracle` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/oracle/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/oracle/scaling/horizontal-scaling/scale-horizontally/yamls](/docs/guides/oracle/scaling/horizontal-scaling/scale-horizontally/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Horizontal Scaling on Oracle Cluster

Here, we are going to deploy a `Oracle` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

#### Prepare Cluster

At first, we are going to deploy a cluster with 3 nodes. Then, we are going to add two additional nodes through horizontal scaling. Finally, we will remove 1 node from the cluster again via horizontal scaling.

**Deploy Oracle Cluster:**

In this section, we are going to deploy a Oracle cluster with 3 nodes. Then, in the next section we will scale the cluster using horizontal scaling. Below is the YAML of the `Oracle` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "21.3.0"
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

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/scaling/horizontal-scaling/scale-horizontally/yamls/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `Oracle` objects using Kubernetes API. When a `Oracle` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc.

Now, watch `Oracle` is going to `Running` state and also watch `PetSet` and its pods:

```bash
$ watch -n 3 kubectl get oracle -n demo oracle-sample
Every 3.0s: kubectl get oracle -n demo oracle-sample

NAME             VERSION   STATUS   AGE
oracle-sample    1.17.0    Ready    4m40m


$ watch -n 3 kubectl get petset -n demo oracle-sample
Every 3.0s: kubectl get petset -n demo oracle-sample

NAME              READY   AGE
oracle-sample     3/3     4m41m


$ watch -n 3 kubectl get pods -n demo
Every 3.0s: kubectl get pod -n demo

NAME                READY   STATUS    RESTARTS   AGE
oracle-sample-0     1/1     Running   0          4m25m
oracle-sample-1     1/1     Running   0          4m26m
oracle-sample-2     1/1     Running   0          4m26m
```

Let's check the current number of nodes:

```bash
$ kubectl get oracle -n demo oracle-sample -o=jsonpath='{.spec.replicas}{"\n"}'
3
```

We are ready to apply the `OracleOpsRequest` CR to scale horizontally.

#### Scale Up

Here, we are going to scale up the cluster from 3 nodes to 5 nodes.

**Create OracleOpsRequest:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: oracle-sample
  horizontalScaling:
    node: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling on `oracle-sample` Oracle database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.node` specifies the desired number of nodes after scaling.

Let's create the `OracleOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/scaling/horizontal-scaling/scale-horizontally/yamls/hscale-up.yaml
oracleopsrequest.ops.kubedb.com/qdops-hscale-up created
```

**Verify Oracle scale-up completed successfully:**

```bash
$ watch -n 3 kubectl get OracleOpsRequest -n demo qdops-hscale-up
Every 3.0s: kubectl get OracleOpsRequest -n demo qdops-hscale-up

NAME               TYPE               STATUS       AGE
qdops-hscale-up    HorizontalScaling  Successful   3m57s
```

Now let's verify that the number of nodes has increased:

```bash
$ kubectl get oracle -n demo oracle-sample -o=jsonpath='{.spec.replicas}{"\n"}'
5

$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
oracle-sample-0     1/1     Running   0          10m
oracle-sample-1     1/1     Running   0          10m
oracle-sample-2     1/1     Running   0          10m
oracle-sample-3     1/1     Running   0          2m
oracle-sample-4     1/1     Running   0          1m
```

#### Scale Down

Here, we are going to scale down the cluster from 5 nodes to 4 nodes.

**Create OracleOpsRequest:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: oracle-sample
  horizontalScaling:
    node: 4
```

Let's create the `OracleOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/scaling/horizontal-scaling/scale-horizontally/yamls/hscale-down.yaml
oracleopsrequest.ops.kubedb.com/qdops-hscale-down created
```

**Verify Oracle scale-down completed successfully:**

```bash
$ watch -n 3 kubectl get OracleOpsRequest -n demo qdops-hscale-down
Every 3.0s: kubectl get OracleOpsRequest -n demo qdops-hscale-down

NAME                 TYPE               STATUS       AGE
qdops-hscale-down    HorizontalScaling  Successful   2m15s
```

Now let's verify that the number of nodes has decreased:

```bash
$ kubectl get oracle -n demo oracle-sample -o=jsonpath='{.spec.replicas}{"\n"}'
4

$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
oracle-sample-0     1/1     Running   0          14m
oracle-sample-1     1/1     Running   0          14m
oracle-sample-2     1/1     Running   0          14m
oracle-sample-3     1/1     Running   0          6m
```

We have successfully performed horizontal scaling on the Oracle cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo oracle-sample
kubectl delete OracleOpsRequest -n demo qdops-hscale-up qdops-hscale-down
```