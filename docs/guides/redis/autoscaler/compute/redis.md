---
title: Redis Autoscaling
menu:
  docs_{{ .version }}:
    identifier: rd-auto-scaling-standalone
    name: Redis Autoscaling
    parent: rd-compute-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Redis Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Redis standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisAutoscaler](/docs/guides/redis/concepts/autoscaler.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/redis/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Standalone Database

Here, we are going to deploy a `Redis` standalone using a supported version by `KubeDB` operator. Then we are going to apply `RedisAutoscaler` to set up autoscaling.

#### Deploy Redis standalone

In this section, we are going to deploy a Redis standalone database with version `6.2.5`.  Then, in the next section we will set up autoscaling for this database using `RedisAutoscaler` CRD. Below is the YAML of the `Redis` CR that we are going to create,

> If you want to autoscale Redis in `Cluster` or `Sentinel` mode, just deploy a Redis database in respective Mode and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-standalone
  namespace: demo
spec:
  version: "6.2.5"
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "200m"
          memory: "300Mi"
        limits:
          cpu: "200m"
          memory: "300Mi"
  terminationPolicy: WipeOut
```

Let's create the `Redis` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/autoscaling/compute/rd-standalone.yaml
redis.kubedb.com/rd-standalone created
```

Now, wait until `rd-standalone` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME            VERSION    STATUS    AGE
rd-standalone   6.2.5      Ready     2m53s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo rd-standalone-0 -o json | jq '.spec.containers[].resources'
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

Let's check the Redis resources,
```bash
$ kubectl get redis -n demo rd-standalone -o json | jq '.spec.podTemplate.spec.resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the redis.

We are now ready to apply the `RedisAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a RedisAutoscaler Object.

#### Create RedisAutoscaler Object

In order to set up compute resource autoscaling for this standalone database, we have to create a `RedisAutoscaler` CRO with our desired configuration. Below is the YAML of the `RedisAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RedisAutoscaler
metadata:
  name: rd-as
  namespace: demo
spec:
  databaseRef:
    name: rd-standalone
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    standalone:
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

> If you want to autoscale Redis in Cluster mode, the field in `spec.compute` should be `cluster` and for sentinel it should be `sentinel`. The subfields are same inside `spec.computer.standalone`, `spec.compute.cluster` and `spec.compute.sentinel`

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `rd-standalone` database.
- `spec.compute.standalone.trigger` specifies that compute resource autoscaling is enabled for this database.
- `spec.compute.standalone.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.replicaset.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.standalone.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.standalone.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.standalone.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.standalone.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here : [timeout](/docs/guides/redis/concepts/redisopsrequest.md#spectimeout), [apply](/docs/guides/redis/concepts/redisopsrequest.md#specapply).

Let's create the `RedisAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/autoscaling/compute/rd-as-standalone.yaml
redisautoscaler.autoscaling.kubedb.com/rd-as created
```

#### Verify Autoscaling is set up successfully

Let's check that the `redisautoscaler` resource is created successfully,

```bash
$ kubectl get redisautoscaler -n demo
NAME    AGE
rd-as   102s

$ kubectl describe redisautoscaler rd-as -n demo
Name:         rd-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         RedisAutoscaler
Metadata:
  Creation Timestamp:  2023-02-09T10:02:26Z
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
          f:standalone:
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
    Time:         2023-02-09T10:02:26Z
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:checkpoints:
        f:vpas:
    Manager:         kubedb-autoscaler
    Operation:       Update
    Subresource:     status
    Time:            2023-02-09T10:02:29Z
  Resource Version:  839366
  UID:               5a5dedc1-fbef-4afa-93f3-0ca8dfb8a30b
Spec:
  Compute:
    Standalone:
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
    Name:  rd-standalone
  Ops Request Options:
    Apply:    IfReady
    Timeout:  3m0s
Status:
  Checkpoints:
    Cpu Histogram:
    Last Update Time:  2023-02-09T10:03:29Z
    Memory Histogram:
    Ref:
      Container Name:   redis
      Vpa Object Name:  rd-standalone
    Version:            v3
  Vpas:
    Conditions:
      Last Transition Time:  2023-02-09T10:02:29Z
      Status:                False
      Type:                  RecommendationProvided
    Recommendation:
    Vpa Name:  rd-standalone
Events:        <none>

```
So, the `redisautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `redisopsrequest` based on the recommendations, if the database pods are needed to scaled up or down.

Let's watch the `redisopsrequest` in the demo namespace to see if any `redisopsrequest` object is created. After some time you'll see that a `redisopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                         TYPE              STATUS       AGE
rdops-rd-standalone-q2zozm   VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                         TYPE              STATUS       AGE
rdops-rd-standalone-q2zozm   VerticalScaling   Successful   68s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 

Now, we are going to verify from the Pod, and the Redis yaml whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo rd-standalone-0 -o json | jq '.spec.containers[].resources'
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

$ kubectl get redis -n demo rd-standalone -o json | jq '.spec.podTemplate.spec.resources'
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
$ kubectl patch -n demo rd/rd-standalone -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-standalone patched

$ kubectl delete rd -n demo rd-standalone
redis.kubedb.com "rd-standalone" deleted

$ kubectl delete redisautoscaler -n demo rd-as
redisautoscaler.autoscaling.kubedb.com "rd-as" deleted
```