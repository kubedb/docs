---
title: Sentinel Autoscaling
menu:
  docs_{{ .version }}:
    identifier: rd-auto-scaling-sentinel
    name: Sentinel Autoscaling
    parent: rd-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Sentinel

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Redis standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [RedisSentinel](/docs/guides/redis/concepts/redissentinel.md)
  - [RedisAutoscaler](/docs/guides/redis/concepts/autoscaler.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/redis/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Sentinel

Here, we are going to deploy a `RedisSentinel` instance using a supported version by `KubeDB` operator. Then we are going to apply `RedisSentinelAutoscaler` to set up autoscaling.

#### Deploy Redis standalone

In this section, we are going to deploy a RedisSentinel instance with version `6.2.14`.  Then, in the next section we will set up autoscaling for this database using `RedisSentinelAutoscaler` CRD. Below is the YAML of the `RedisSentinel` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: RedisSentinel
metadata:
  name: sen-demo
  namespace: demo
spec:
  version: "6.2.14"
  storageType: Durable
  replicas: 3
  storage:
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      containers:
      - name: redissentinel
        resources:
          requests:
            cpu: "200m"
            memory: "300Mi"
          limits:
            cpu: "200m"
            memory: "300Mi"
  deletionPolicy: WipeOut
```

Let's create the `RedisSentinel` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/autoscaling/compute/sentinel.yaml
redissentinel.kubedb.com/sen-demo created
```

Now, wait until `sen-demo` has status `Ready`. i.e,

```bash
$ kubectl get redissentinel -n demo
NAME       VERSION   STATUS   AGE
sen-demo   6.2.14     Ready    86s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sen-demo-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

Let's check the RedisSentinel resources,
```bash
$ kubectl get redissentinel -n demo sen-demo -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the redissentinel.

We are now ready to apply the `RedisSentinelAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a RedisSentinelAutoscaler Object.

#### Create RedisSentinelAutoscaler Object

In order to set up compute resource autoscaling for this standalone database, we have to create a `RedisAutoscaler` CRO with our desired configuration. Below is the YAML of the `RedisAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RedisSentinelAutoscaler
metadata:
  name: sen-as
  namespace: demo
spec:
  databaseRef:
    name: sen-demo
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    sentinel:
      trigger: "On"
      podLifeTimeThreshold: 5m
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

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `sen-demo` database.
- `spec.compute.standalone.trigger` specifies that compute resource autoscaling is enabled for this database.
- `spec.compute.sentinel.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.sentinel.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.sentinel.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.sentinel.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.sentinel.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here :  [timeout](/docs/guides/redis/concepts/redisopsrequest.md#spectimeout), [apply](/docs/guides/redis/concepts/redisopsrequest.md#specapply).

If it was an `InMemory database`, we could also autoscaler the inMemory resources using Redis compute autoscaler, like below.


Let's create the `RedisAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/compute/autoscaling/sen-as.yaml
redissentinelautoscaler.autoscaling.kubedb.com/sen-as created
```

#### Verify Autoscaling is set up successfully

Let's check that the `redisautoscaler` resource is created successfully,

```bash
$ kubectl get redisautoscaler -n demo
NAME    AGE
sen-as   102s

$ kubectl describe redissentinelautoscaler sen-as -n demo
Name:         sen-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         RedisSentinelAutoscaler
Metadata:
  Creation Timestamp:  2023-02-09T11:14:18Z
  Generation:          1
  Managed Fields:
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:compute:
          .:
          f:sentinel:
            .:
            f:containerControlledValues:
            f:controlledResources:
            f:maxAllowed:
              .:
              f:cpu:
              f:memory:
            f:minAllowed:
              .:
              f:cpu:
              f:memory:
            f:podLifeTimeThreshold:
            f:resourceDiffPercentage:
            f:trigger:
        f:databaseRef:
        f:opsRequestOptions:
          .:
          f:apply:
          f:timeout:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2023-02-09T11:14:18Z
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:checkpoints:
        f:conditions:
        f:vpas:
    Manager:         kubedb-autoscaler
    Operation:       Update
    Subresource:     status
    Time:            2023-02-09T11:15:20Z
  Resource Version:  845618
  UID:               44da50a4-6e4f-49fa-b7e4-6c7f83c3e6c4
Spec:
  Compute:
    Sentinel:
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
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  sen-demo
  Ops Request Options:
    Apply:    IfReady
    Timeout:  3m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
      Reference Timestamp:  2023-02-09T00:00:00Z
      Total Weight:         0.4150619553793766
    First Sample Start:     2023-02-09T11:14:17Z
    Last Sample Start:      2023-02-09T11:14:32Z
    Last Update Time:       2023-02-09T11:14:35Z
    Memory Histogram:
      Reference Timestamp:  2023-02-10T00:00:00Z
    Ref:
      Container Name:     redissentinel
      Vpa Object Name:    sen-demo
    Total Samples Count:  3
    Version:              v3
  Conditions:
    Last Transition Time:  2023-02-09T11:15:20Z
    Message:               Successfully created RedisSentinelOpsRequest demo/rdsops-sen-demo-5emii6
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2023-02-09T11:14:35Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  redissentinel
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
    Vpa Name:      sen-demo
Events:            <none>
```
So, the `redisautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `redissentinelopsrequest` based on the recommendations, if the database pods are needed to scaled up or down.

Let's watch the `redissentinelopsrequest` in the demo namespace to see if any `redissentinelopsrequest` object is created. After some time you'll see that a `redissentinelopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get redissentinelopsrequest -n demo
Every 2.0s: kubectl get redissentinelopsrequest -n demo
NAME                         TYPE              STATUS       AGE
rdsops-sen-demo-5emii6       VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get redissentinelopsrequest -n demo
Every 2.0s: kubectl get redissentinelopsrequest -n demo
NAME                         TYPE              STATUS       AGE
rdsops-sen-demo-5emii6       VerticalScaling   Successfull  10s
```

We can see from the above output that the `RedisSentinelOpsRequest` has succeeded. 

Now, we are going to verify from the Pod, and the Redis yaml whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sen-demo-0 -o json | jq '.spec.containers[].resources'
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

$ kubectl get redis -n demo sen-demo -o json | jq '.spec.podTemplate.spec.resources'
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


The above output verifies that we have successfully auto-scaled the resources of the Redis standalone database.



## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo redissentinel/sen-demo -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-demo patched

$ kubectl delete redissentinel -n demo sen-demo
redissentinel.kubedb.com "sen-demo" deleted

$ kubectl delete redissentinelautoscaler -n demo sen-as
redissentinelautoscaler.autoscaling.kubedb.com "sen-as" deleted
```