---
title: Weaviate Compute Autoscaler Cluster
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-compute-cluster
    name: Cluster
    parent: weaviate-autoscaler-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Weaviate Cluster

This guide will show you how to use `KubeDB` to auto-scale compute resources i.e. CPU and memory of a Weaviate cluster database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/weaviate/autoscaler/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Autoscaling of Cluster Database

Here, we are going to deploy a `Weaviate` cluster using a supported version by `KubeDB` operator. Then we are going to apply `WeaviateAutoscaler` to set up autoscaling.

### Deploy Weaviate Cluster

In this section, we are going to deploy a Weaviate cluster with version `1.33.1`. Then, in the next section we will set up autoscaling for this database using `WeaviateAutoscaler` CRD. Below is the YAML of the `Weaviate` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: "1.33.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
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
            cpu: "200m"
            memory: "512Mi"
          limits:
            cpu: "200m"
            memory: "512Mi"
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/autoscaler/compute/weaviate-cluster.yaml
weaviate.kubedb.com/weaviate-sample created
```

Now, wait until `weaviate-sample` has status `Ready`:

```bash
$ kubectl get weaviate -n demo
NAME              VERSION   STATUS   AGE
weaviate-sample   1.33.1    Ready    4m
```

Let's check the Pod container resources:

```bash
$ kubectl get pod -n demo weaviate-sample-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "512Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "512Mi"
  }
}
```

We are now ready to apply the `WeaviateAutoscaler` CRD to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a `WeaviateAutoscaler` Object.

#### Create WeaviateAutoscaler Object

In order to set up compute resource autoscaling for this database cluster, we have to create a `WeaviateAutoscaler` CR with our desired configuration. Below is the YAML of the `WeaviateAutoscaler` object that we are going to create:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: WeaviateAutoscaler
metadata:
  name: weaviate-as-compute
  namespace: demo
spec:
  databaseRef:
    name: weaviate-sample
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 10m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 400m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute autoscaling on `weaviate-sample` database.
- `spec.compute.node.trigger` specifies that compute resource autoscaling is enabled for the Weaviate nodes.
- `spec.compute.node.podLifeTimeThreshold` specifies the minimum age of a Pod before the `VerticalPodAutoscaler` can recommend a resource update.
- `spec.compute.node.resourceDiffPercentage` specifies the minimum percentage change needed before applying a new resource recommendation.
- `spec.compute.node.minAllowed` specifies the minimum allowed resources for the Weaviate nodes.
- `spec.compute.node.maxAllowed` specifies the maximum allowed resources for the Weaviate nodes.
- `spec.compute.node.controlledResources` specifies the resource types that will be auto-scaled.
- `spec.compute.node.containerControlledValues` specifies which resource values should be controlled, here both requests and limits.

Let's create the `WeaviateAutoscaler` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/autoscaler/compute/weaviate-as-compute.yaml
weaviateautoscaler.autoscaling.kubedb.com/weaviate-as-compute created
```

#### Verify Autoscaler is set up successfully

Let's check that the `WeaviateAutoscaler` resource is created successfully:

```bash
$ kubectl get weaviateautoscaler -n demo
NAME                  AGE
weaviate-as-compute   5s

$ kubectl describe weaviateautoscaler weaviate-as-compute -n demo
Name:         weaviate-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         WeaviateAutoscaler
Spec:
  Compute:
    Node:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  2Gi
      Min Allowed:
        Cpu:     400m
        Memory:  400Mi
      Pod Life Time Threshold:      10m0s
      Resource Diff Percentage:     20
      Trigger:                      On
  Database Ref:
    Name:  weaviate-sample
  Ops Request Options:
    Apply:    IfReady
    Timeout:  3m0s
Events:       <none>
```

So, the `WeaviateAutoscaler` resource is created successfully. The operator will now watch the resource usage of the Weaviate pods and create `WeaviateOpsRequest` resources to scale the cluster when needed.

After some time, you can observe that the autoscaler has created a `WeaviateOpsRequest` with type `VerticalScaling`:

```bash
$ kubectl get weaviateopsrequest -n demo
NAME                               TYPE              STATUS       AGE
wvops-weaviate-sample-xxxxxxxx     VerticalScaling   Successful   5m
```

You can then verify the updated resources on the pods:

```bash
$ kubectl get pod -n demo weaviate-sample-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "512Mi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "512Mi"
  }
}
```

The above output verifies that we have successfully autoscaled the resources of the Weaviate cluster database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-sample
kubectl delete weaviateautoscaler -n demo weaviate-as-compute
kubectl delete ns demo
```
