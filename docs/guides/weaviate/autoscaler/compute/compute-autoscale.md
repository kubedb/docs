---
title: Weaviate Compute Autoscaler
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-compute-description
    name: Autoscale Compute Resources
    parent: weaviate-autoscaler-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Weaviate Database

This guide will show you how to use `KubeDB` to auto-scale the compute resources (CPU and Memory) of a Weaviate database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-Manager, and Autoscaler operators in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation). The compute autoscaler relies on metrics to make recommendations.

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Compute Autoscaling Overview](/docs/guides/weaviate/autoscaler/compute/overview.md)
  - [Vertical Scaling](/docs/guides/weaviate/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Autoscaling of Database

Here, we are going to deploy a `Weaviate` database and then set up autoscaling with a `WeaviateAutoscaler`.

### Deploy Weaviate Database

In this section, we are going to deploy a Weaviate database with `500m` CPU and `1Gi` memory. Below is the YAML of the `Weaviate` CR that we are going to create:

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

Let's create the `Weaviate` CR and wait for it to become `Ready`. Then check the current container resources:

```bash
$ kubectl get pod -n demo weaviate-sample-0 -o jsonpath='{.spec.containers[0].resources}'
{"limits":{"cpu":"500m","memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}
```

### Create WeaviateAutoscaler

Now, we are going to set up compute resource autoscaling using a `WeaviateAutoscaler` object. Note the resource knob is under `spec.compute.weaviate`:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: WeaviateAutoscaler
metadata:
  name: weaviate-sample-autoscale
  namespace: demo
spec:
  databaseRef:
    name: weaviate-sample
  compute:
    weaviate:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 1.2Gi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute autoscaling on the `weaviate-sample` database.
- `spec.compute.weaviate.trigger` enables compute resource autoscaling for the Weaviate nodes.
- `spec.compute.weaviate.podLifeTimeThreshold` specifies the minimum age of a Pod before a resource update can be recommended.
- `spec.compute.weaviate.resourceDiffPercentage` specifies the minimum percentage difference required before applying a new recommendation.
- `spec.compute.weaviate.minAllowed` / `maxAllowed` specify the lower and upper bounds of the autoscaled resources.
- `spec.compute.weaviate.controlledResources` specifies the resources that will be auto-scaled.
- `spec.compute.weaviate.containerControlledValues` specifies whether both requests and limits are controlled.

Let's create the `WeaviateAutoscaler`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/autoscaler/compute/weaviate-compute-autoscaler.yaml
weaviateautoscaler.autoscaling.kubedb.com/weaviate-sample-autoscale created
```

### Verify Autoscaling

Let's describe the `WeaviateAutoscaler`. Because the initial resources (`500m`/`1Gi`) are below the `minAllowed` floor (`600m`/`1.2Gi`), the autoscaler quickly produces a recommendation:

```bash
$ kubectl describe weaviateautoscaler -n demo weaviate-sample-autoscale
...
Status:
  Vpas:
    Conditions:
      Status:  True
      Type:    RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  weaviate
        Lower Bound:
          Cpu:     600m
          Memory:  1288490188800m
        Target:
          Cpu:     600m
          Memory:  1288490188800m
        Upper Bound:
          Cpu:     1
          Memory:  2Gi
```

After the `podLifeTimeThreshold` passes, the autoscaler operator creates a `WeaviateOpsRequest` of type `VerticalScaling`:

```bash
$ kubectl get weaviateopsrequest -n demo
NAME                           TYPE              STATUS       AGE
wvops-weaviate-sample-0oyvzl   VerticalScaling   Successful   119s
```

```bash
$ kubectl get weaviateopsrequest -n demo wvops-weaviate-sample-0oyvzl -o jsonpath='{.spec.verticalScaling}'
{"node":{"resources":{"limits":{"cpu":"600m","memory":"1288490188"},"requests":{"cpu":"600m","memory":"1288490188"}}}}
```

Once the ops request completes, verify the updated resources on the pods:

```bash
$ kubectl get pod -n demo weaviate-sample-0 -o jsonpath='{.spec.containers[0].resources}'
{"limits":{"cpu":"600m","memory":"1288490188"},"requests":{"cpu":"600m","memory":"1288490188"}}
```

The compute resources of the Weaviate database have been autoscaled up to the `minAllowed` floor (`600m` CPU / `1.2Gi` memory). When the actual usage grows, the autoscaler will continue to recommend higher resources (up to `maxAllowed`).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete weaviateautoscaler -n demo weaviate-sample-autoscale
$ kubectl delete weaviate -n demo weaviate-sample
$ kubectl delete ns demo
```
