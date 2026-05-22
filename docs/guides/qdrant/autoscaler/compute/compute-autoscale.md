---
title: Qdrant Compute Autoscaler
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-compute-description
    name: Autoscale Compute Resources
    parent: qdrant-autoscaler-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Qdrant Database

This guide will show you how to use `KubeDB` to auto-scale compute resources i.e. CPU and memory of a Qdrant database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/)
  - [QdrantAutoscaler](/docs/guides/qdrant/concepts/autoscaler.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/qdrant/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Autoscaling of Database

Here, we are going to deploy a `Qdrant` database using a supported version by `KubeDB` operator. Then we are going to apply `QdrantAutoscaler` to set up autoscaling.

### Deploy Qdrant Database

In this section, we are going to deploy a Qdrant database with version `1.17.0`. Then, in the next section we will set up autoscaling for this database using `QdrantAutoscaler` CRD. Below is the YAML of the `Qdrant` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
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
      - name: qdrant
        resources:
          requests:
            cpu: "200m"
            memory: "512Mi"
          limits:
            cpu: "200m"
            memory: "512Mi"
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/autoscaler/compute/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

Now, wait until `qdrant-sample` has status `Ready`:

```bash
$ kubectl get qdrant -n demo
NAME            VERSION   STATUS   AGE
qdrant-sample   1.17.0    Ready    51s
```

Let's check the Pod container resources:

```bash
$ kubectl get pod -n demo qdrant-sample-0 -o json | jq '.spec.containers[].resources'
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

We are now ready to apply the `QdrantAutoscaler` CRD to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a `QdrantAutoscaler` Object.

#### Create QdrantAutoscaler Object

In order to set up compute resource autoscaling for this database, we have to create a `QdrantAutoscaler` CR with our desired configuration. Below is the YAML of the `QdrantAutoscaler` object that we are going to create:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: QdrantAutoscaler
metadata:
  name: qdrant-as-compute
  namespace: demo
spec:
  databaseRef:
    name: qdrant-sample
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

- `spec.databaseRef.name` specifies that we are performing compute autoscaling on `qdrant-sample` database.
- `spec.compute.node.trigger` specifies that compute resource autoscaling is enabled for the Qdrant nodes.
- `spec.compute.node.podLifeTimeThreshold` specifies the minimum age of a Pod before the `VerticalPodAutoscaler` can recommend a resource update.
- `spec.compute.node.resourceDiffPercentage` specifies the minimum percentage change needed before applying a new resource recommendation.
- `spec.compute.node.minAllowed` specifies the minimum allowed resources for the Qdrant nodes.
- `spec.compute.node.maxAllowed` specifies the maximum allowed resources for the Qdrant nodes.
- `spec.compute.node.controlledResources` specifies the resource types that will be auto-scaled.
- `spec.compute.node.containerControlledValues` specifies which resource values should be controlled, here both requests and limits.
- `spec.opsRequestOptions.apply` has two supported values: `IfReady` and `Always`. Use `IfReady` to process the ops request only when the database is Ready. Use `Always` to process the execution irrespective of the database state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the ops request. If a step doesn't finish within the specified timeout, the ops request will result in failure.

Let's create the `QdrantAutoscaler` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/autoscaler/compute/qdrant-as-compute.yaml
qdrantautoscaler.autoscaling.kubedb.com/qdrant-as-compute created
```

#### Verify Autoscaler is set up successfully

Let's check that the `QdrantAutoscaler` resource is created successfully:

```bash
$ kubectl get qdrantautoscaler -n demo
NAME                AGE
qdrant-as-compute   0s

$ kubectl describe qdrantautoscaler qdrant-as-compute -n demo
Name:         qdrant-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         QdrantAutoscaler
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
        Cpu:                     400m
        Memory:                  400Mi
      Pod Life Time Threshold:   10m
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  qdrant-sample
  Ops Request Options:
    Apply:        IfReady
    Max Retries:  1
    Timeout:      3m
Status:
  Vpas:
    Vpa Name:  qdrant-sample
Events:        <none>
```

So, the `QdrantAutoscaler` resource is created successfully. The operator will now watch the resource usage of the Qdrant pods and create `QdrantOpsRequest` resources to scale when needed.

After some time, you can observe that the autoscaler has created a `QdrantOpsRequest` with type `VerticalScaling`:

```bash
$ kubectl get qdrantopsrequest -n demo
NAME                           TYPE              STATUS       AGE
qdops-qdrant-sample-829lnp     VerticalScaling   Successful   45s
```

You can then verify the updated resources on the pods:

```bash
$ kubectl get pod -n demo qdrant-sample-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "400Mi"
  }
}
```

The above output verifies that we have successfully autoscaled the resources of the Qdrant database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete qdrantautoscaler -n demo qdrant-as-compute
kubectl delete ns demo
```
