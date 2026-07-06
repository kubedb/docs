---
title: Scale Qdrant Vertically
menu:
  docs_{{ .version }}:
    identifier: qdrant-vertical-scaling-ops
    name: Scale Vertically
    parent: qdrant-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Qdrant Cluster

This guide will show you how to use `KubeDB` Ops Manager to update the resources of a `Qdrant` instance.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/qdrant/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/scaling/vertical-scaling](/docs/examples/qdrant/scaling/vertical-scaling) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Qdrant Cluster

Here, we are going to deploy a `Qdrant` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

**Deploy Qdrant:**

In this section, we are going to deploy a Qdrant cluster. Then, in the next section, we will update the resources of the database nodes using vertical scaling. Below is the YAML of the `Qdrant` CR that we are going to create:

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/scaling/vertical-scaling/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

**Wait for the cluster to be ready:**

```bash
$ watch -n 3 kubectl get qdrant -n demo qdrant-sample
Every 3.0s: kubectl get qdrant -n demo qdrant-sample

NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    3m16s

$ watch -n 3 kubectl get petset -n demo qdrant-sample
Every 3.0s: kubectl get petset -n demo qdrant-sample

NAME              READY   AGE
qdrant-sample     3/3     3m54s

$ watch -n 3 kubectl get pod -n demo
Every 3.0s: kubectl get pod -n demo

NAME                READY   STATUS    RESTARTS   AGE
qdrant-sample-0     1/1     Running   0          4m51s
qdrant-sample-1     1/1     Running   0          3m50s
qdrant-sample-2     1/1     Running   0          3m46s
```

Let's check the resources of the `qdrant-sample-0` pod:

```bash
$ kubectl get pod -n demo qdrant-sample-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

We are ready to apply the `QdrantOpsRequest` CR to vertically scale the cluster.

#### Create QdrantOpsRequest for Vertical Scaling

In order to update the resources of the database, we have to create a `QdrantOpsRequest` CR with our desired resources. Below is the YAML of the `QdrantOpsRequest` CR that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: qdrant-sample
  verticalScaling:
    node:
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "1"
          memory: "2Gi"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling on `qdrant-sample` Qdrant database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.node.resources` specifies the desired CPU and memory resources for the Qdrant nodes.
- `spec.timeout` specifies the timeout for the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#spectimeout)).
- `spec.apply` specifies when to apply the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#specapply)).

Let's create the `QdrantOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/scaling/vertical-scaling/vscale.yaml
qdrantopsrequest.ops.kubedb.com/qdops-vscale created
```

#### Verify Qdrant vertical scaling completed successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of the `Qdrant` object and related `PetSet`.

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ watch -n 3 kubectl get QdrantOpsRequest -n demo qdops-vscale
Every 3.0s: kubectl get QdrantOpsRequest -n demo qdops-vscale

NAME            TYPE              STATUS       AGE
qdops-vscale    VerticalScaling   Successful   3m12s
```

Now, let's verify that the resources of the pods have been updated:

```bash
$ kubectl get pod -n demo qdrant-sample-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see from the above output that the resources of the `qdrant-sample-0` pod have been updated successfully. All pods in the cluster will have the same updated resource configuration.

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete QdrantOpsRequest -n demo qdops-vscale
```