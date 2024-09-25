---
title: Memcached Autoscaling
menu:
  docs_{{ .version }}:
    identifier: mc-auto-scaling
    name: Compute Autoscaler
    parent: compute-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Memcached Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Memcached database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [Memcached](/docs/guides/memcached/concepts/memcached.md)
  - [MemcachedAutoscaler](/docs/guides/memcached/concepts/memcached-autoscaler.md)
  - [MemcachedOpsRequest](/docs/guides/memcached/concepts/memcached-opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/memcached/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/memcached](/docs/examples/memcached) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Memcached Database

Here, we are going to deploy a `Memcached` database using a supported version by `KubeDB` operator. Then we are going to apply `MemcachedAutoscaler` to set up autoscaling.

#### Deploy Memcached Database

In this section, we are going to deploy a Memcached database with version `1.6.22`.  Then, in the next section we will set up autoscaling for this database using `MemcachedAutoscaler` CRD. Below is the YAML of the `Memcached` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: mc-autoscaler-compute
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 100m
              memory: 100Mi
  deletionPolicy: WipeOut
```

Let's create the `Memcached` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/autoscaler/compute/mc-compute-autoscaler.yaml
Memcached.kubedb.com/mc-compute-autoscaler created
```

Now, wait until `mc-compute-autoscaler` has status `Ready`. i.e,

```bash
$ kubectl get mc -n demo
NAME                    VERSION     STATUS    AGE
mc-autoscaler-compute   1.6.22      Ready     2m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo mc-autoscaler-compute-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "100m",
    "memory": "100Mi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "100Mi"
  }
}
```

Let's check the Memcached resources,
```bash
$ kubectl get Memcached -n demo mc-autoscaler-compute -o json | jq '.spec.podTemplate.spec.containers[] | select(.name == "memcached") | .resources'
{
  "limits": {
    "cpu": "100m",
    "memory": "100Mi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "100Mi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the Memcached.

We are now ready to apply the `MemcachedAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a MemcachedAutoscaler Object.

#### Create MemcachedAutoscaler Object

In order to set up compute resource autoscaling for this database, we have to create a `MemcachedAutoscaler` CRO with our desired configuration. Below is the YAML of the `MemcachedAutoscaler` object that we are going to create:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MemcachedAutoscaler
metadata:
  name: mc-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: mc-autoscaler-compute
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    memcached:
      trigger: "On"
      podLifeTimeThreshold: 1m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 400m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `mc-compute-autoscaler` database.
- `spec.compute.memcached.trigger` specifies that compute resource autoscaling is enabled for this database.
- `spec.compute.memcached.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.memcached.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.memcached.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.memcached.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.memcached.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.memcahced.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here : [timeout](/docs/guides/memcached/concepts/memcached-opsrequest.md#spectimeout), [apply](/docs/guides/memcached/concepts/memcached-opsrequest.md#specapply).

Let's create the `MemcachedAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/autoscaling/compute/mc-compute-autoscaler.yaml
Memcachedautoscaler.autoscaling.kubedb.com/rd-as created
```

#### Verify Autoscaling is set up successfully

Let's check that the `Memcachedautoscaler` resource is created successfully,

```bash
$ kubectl get memcachedautoscaler -n demo
NAME            AGE
mc-autoscaler   16m

$ kubectl describe memcachedautoscaler mc-autoscaler -n demo
Name:         mc-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MemcachedAutoscaler
Metadata:
  Creation Timestamp:  2024-09-10T12:55:35Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Memcached
    Name:                  mc-autoscaler-compute
    UID:                   56a15163-0f8b-4f35-8cd9-ae9bd0976ea7
  Resource Version:        105259
  UID:                     2ef29276-dc47-4b2d-8995-ad5114b419f3
Spec:
  Compute:
    Memcached:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:                     400m
        Memory:                  400Mi
      Pod Life Time Threshold:   1m
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  mc-autoscaler-compute
  Ops Request Options:
    Apply:    IfReady
    Timeout:  3m
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
      Reference Timestamp:  2024-09-10T13:10:00Z
      Total Weight:         0.42972012872296605
    First Sample Start:     2024-09-10T13:08:51Z
    Last Sample Start:      2024-09-10T13:12:00Z
    Last Update Time:       2024-09-10T13:12:04Z
    Memory Histogram:
      Reference Timestamp:  2024-09-10T13:15:00Z
    Ref:
      Container Name:     memcached
      Vpa Object Name:    mc-autoscaler-compute
    Total Samples Count:  4
    Version:              v3
  Conditions:
    Last Transition Time:  2024-09-10T13:10:04Z
    Message:               Successfully created MemcachedOpsRequest demo/mcops-mc-autoscaler-compute-p1usdl
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-09-10T13:09:04Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  memcached
        Lower Bound:
          Cpu:     400m
          Memory:  400Mi
        Target:
          Cpu:     400m
          Memory:  400Mi
        Uncapped Target:
          Cpu:     100m
          Memory:  262144k
        Upper Bound:
          Cpu:     1
          Memory:  1Gi
    Vpa Name:      mc-autoscaler-compute
Events:            <none>
```
So, the `Memcachedautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `Memcachedopsrequest` based on the recommendations, if the database pods are needed to scaled up or down.

Let's watch the `Memcachedopsrequest` in the demo namespace to see if any `Memcachedopsrequest` object is created. After some time you'll see that a `Memcachedopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get memcachedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME                                 TYPE              STATUS       AGE
mcops-mc-autoscaler-compute-p1usdl   VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get memcachedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME                                 TYPE              STATUS       AGE
mcops-mc-autoscaler-compute-p1usdl   VerticalScaling   Successful   1m
```

We can see from the above output that the `memcachedOpsRequest` has succeeded. 


```bash
$ kubectl get pod -n demo mc-autoscaler-compute-0 -o json | jq '.spec.containers[].resources'
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

$ kubectl get Memcached -n demo mc-autoscaler-compute -o json | jq '.spec.podTemplate.spec.containers[] | select(.name == "memcached") | .resources'
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

The above output verifies that we have successfully auto-scaled the resources of the Memcached database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo mc/mc-autoscaler-compute -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
memcached.kubedb.com/mc-autoscaler-compute patched

$ kubectl delete mc -n demo mc-autoscaler-compute
memcached.kubedb.com "mc-autoscaler-compute" deleted

$ kubectl delete memcachedautoscaler -n demo mc-autoscaler
memcachedautoscaler.autoscaling.kubedb.com "mc-autoscaler" deleted
```