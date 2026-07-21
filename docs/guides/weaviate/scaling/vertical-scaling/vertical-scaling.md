---
title: Scale Weaviate Vertically
menu:
  docs_{{ .version }}:
    identifier: weaviate-vertical-scaling-ops
    name: Scale Vertically
    parent: weaviate-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Weaviate Cluster

This guide will show you how to use the `KubeDB` Ops Manager to update the resources (CPU and Memory) of a Weaviate cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Vertical Scaling Overview](/docs/guides/weaviate/scaling/vertical-scaling/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/scaling/vertical-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/scaling/vertical-scaling) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate

In this section, we are going to deploy a Weaviate cluster using KubeDB. Then, in the next section we will update the resources using a `WeaviateOpsRequest`.

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
  podTemplate:
    spec:
      containers:
        - name: weaviate
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 500m
              memory: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR and wait for it to become `Ready`:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/scaling/vertical-scaling/weaviate.yaml
weaviate.kubedb.com/weaviate-sample created

$ kubectl get weaviate -n demo
NAME              TYPE                  VERSION   STATUS   AGE
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Ready    5m
```

Let's check the current resources of one of the pods:

```bash
$ kubectl get pod -n demo weaviate-sample-0 -o jsonpath='{.spec.containers[0].resources}'
{"limits":{"cpu":"500m","memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}
```

## Apply Vertical Scaling on the Weaviate Cluster

Here, we are going to update the resources of the cluster to meet the resources after vertical scaling.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: weaviate-sample
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

- `spec.type` specifies that this is a `VerticalScaling` operation.
- `spec.databaseRef.name` specifies that we are performing the operation on `weaviate-sample`.
- `spec.verticalScaling.node` specifies the desired resources for the Weaviate nodes after scaling.
- `spec.verticalScaling.mode` specifies how the scaling is actuated — `Restart` (default, restarts the Pods) or `InPlace` (resizes the running Pods without a restart, falling back to restart if a Node can't fit the new resources). See [Vertical Scaling Modes](/docs/guides/weaviate/scaling/vertical-scaling/overview.md#vertical-scaling-modes).

Let's create the `WeaviateOpsRequest` CR:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/scaling/vertical-scaling/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/wvops-vertical-scale created
```

The Ops Manager will update the PetSet resources and restart the pods one by one to apply the new resources.

```bash
$ kubectl get weaviateopsrequest -n demo wvops-vertical-scale
NAME                   TYPE              STATUS       AGE
wvops-vertical-scale   VerticalScaling   Successful   2m
```

Let's look at the `status.conditions` of the `WeaviateOpsRequest`:

```bash
$ kubectl get weaviateopsrequest -n demo wvops-vertical-scale -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-vertical-scale
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: weaviate-sample
  maxRetries: 1
  timeout: 5m
  type: VerticalScaling
  verticalScaling:
    node:
      resources:
        limits:
          cpu: "1"
          memory: 2Gi
        requests:
          cpu: "1"
          memory: 2Gi
status:
  conditions:
  - message: Weaviate ops-request has started to vertically scaling the Weaviate nodes
    reason: VerticalScaling
    status: "True"
    type: VerticalScaling
  - message: Successfully updated PetSets Resources
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: get pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: GetPod--weaviate-sample-0
  - message: evict pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: EvictPod--weaviate-sample-0
  - message: running pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: RunningPod--weaviate-sample-0
  - message: get pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: GetPod--weaviate-sample-1
  - message: evict pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: EvictPod--weaviate-sample-1
  - message: running pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: RunningPod--weaviate-sample-1
  - message: get pod; ConditionStatus:True; PodName:weaviate-sample-2
    status: "True"
    type: GetPod--weaviate-sample-2
  - message: evict pod; ConditionStatus:True; PodName:weaviate-sample-2
    status: "True"
    type: EvictPod--weaviate-sample-2
  - message: running pod; ConditionStatus:True; PodName:weaviate-sample-2
    status: "True"
    type: RunningPod--weaviate-sample-2
  - message: Successfully Restarted Pods With Resources
    reason: RestartPods
    status: "True"
    type: RestartPods
  - message: Successfully completed the vertical scaling for Weaviate
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

Now, let's verify the resources of the cluster have been updated:

```bash
$ kubectl get pod -n demo weaviate-sample-0 -o jsonpath='{.spec.containers[0].resources}'
{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}}

$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.podTemplate.spec.containers[0].resources}'
{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}}
```

The resources have been updated successfully.

### In-Place Vertical Scaling

To resize the Pods **without a restart**, set `spec.verticalScaling.mode` to `InPlace` in the
`WeaviateOpsRequest`. The operator resizes the running containers via the Kubernetes `pods/resize`
subresource and only restarts a Pod if its Node cannot accommodate the new resources.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-vertical-scale-inplace
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: weaviate-sample
  verticalScaling:
    mode: InPlace
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

Apply it the same way as above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/scaling/vertical-scaling/ops-request-inplace.yaml
weaviateopsrequest.ops.kubedb.com/wvops-vertical-scale-inplace created
```

The resources update in place with no Pod restart.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- [Horizontal Scaling](/docs/guides/weaviate/scaling/horizontal-scaling/horizontal-scaling.md) of a Weaviate cluster.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete weaviateopsrequest -n demo wvops-vertical-scale
$ kubectl delete weaviate -n demo weaviate-sample
$ kubectl delete ns demo
```
