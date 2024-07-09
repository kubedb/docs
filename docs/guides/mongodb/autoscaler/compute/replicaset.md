---
title: MongoDB Replicaset Autoscaling
menu:
  docs_{{ .version }}:
    identifier: mg-auto-scaling-replicaset
    name: Replicaset
    parent: mg-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a MongoDB Replicaset Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a MongoDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [MongoDBAutoscaler](/docs/guides/mongodb/concepts/autoscaler.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/mongodb/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Replicaset Database

Here, we are going to deploy a `MongoDB` Replicaset using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBAutoscaler` to set up autoscaling.

#### Deploy MongoDB Replicaset

In this section, we are going to deploy a MongoDB Replicaset database with version `4.4.26`.  Then, in the next section we will set up autoscaling for this database using `MongoDBAutoscaler` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "replicaset"
  replicas: 3
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
  deletionPolicy: WipeOut

```

Let's create the `MongoDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/compute/mg-rs.yaml
mongodb.kubedb.com/mg-rs created
```

Now, wait until `mg-rs` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME    VERSION    STATUS    AGE
mg-rs   4.4.26      Ready     2m53s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo mg-rs-0 -o json | jq '.spec.containers[].resources'
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

Let's check the MongoDB resources,
```bash
$ kubectl get mongodb -n demo mg-rs -o json | jq '.spec.podTemplate.spec.resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the mongodb.

We are now ready to apply the `MongoDBAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a MongoDBAutoscaler Object.

#### Create MongoDBAutoscaler Object

In order to set up compute resource autoscaling for this replicaset database, we have to create a `MongoDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MongoDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: mg-as-rs
  namespace: demo
spec:
  databaseRef:
    name: mg-rs
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    replicaSet:
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

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `mg-rs` database.
- `spec.compute.replicaSet.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.replicaSet.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.replicaset.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.replicaSet.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.replicaSet.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.replicaSet.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.replicaSet.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 3 fields. Know more about them here : [readinessCriteria](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria), [timeout](/docs/guides/mongodb/concepts/opsrequest.md#spectimeout), [apply](/docs/guides/mongodb/concepts/opsrequest.md#specapply).

If it was an `InMemory database`, we could also autoscaler the inMemory resources using MongoDB compute autoscaler, like below.

#### Autoscale inMemory database
To autoscale inMemory databases, you need to specify the `spec.compute.replicaSet.inMemoryStorage` section.

```yaml
  ...
  inMemoryStorage:
    usageThresholdPercentage: 80
    scalingFactorPercentage: 30
  ...
```
It has two fields inside it.
- `usageThresholdPercentage`. If db uses more than usageThresholdPercentage of the total memory, memoryStorage should be increased. Default usage threshold is 70%.
- `scalingFactorPercentage`. If db uses more than usageThresholdPercentage of the total memory, memoryStorage should be increased by this given scaling percentage. Default scaling percentage is 50%.

> Note: To inform you, We use `db.serverStatus().inMemory.cache["bytes currently in the cache"]` & `db.serverStatus().inMemory.cache["maximum bytes configured"]` to calculate the used & maximum inMemory storage respectively.

Let's create the `MongoDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/compute/mg-as-rs.yaml
mongodbautoscaler.autoscaling.kubedb.com/mg-as-rs created
```

#### Verify Autoscaling is set up successfully

Let's check that the `mongodbautoscaler` resource is created successfully,

```bash
$ kubectl get mongodbautoscaler -n demo
NAME        AGE
mg-as-rs    102s

$ kubectl describe mongodbautoscaler mg-as-rs -n demo
Name:         mg-as-rs
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MongoDBAutoscaler
Metadata:
  Creation Timestamp:  2022-10-27T06:56:34Z
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
          f:replicaSet:
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
    Time:         2022-10-27T06:56:34Z
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
    Time:            2022-10-27T07:01:05Z
  Resource Version:  640314
  UID:               ab03414a-67a2-4da4-8960-6e67ae56b503
Spec:
  Compute:
    Replica Set:
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
    Name:  mg-rs
  Ops Request Options:
    Apply:    IfReady
    Timeout:  3m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              2
        Weight:             10000
        Index:              3
        Weight:             5000
      Reference Timestamp:  2022-10-27T00:00:00Z
      Total Weight:         0.3673624107285783
    First Sample Start:     2022-10-27T07:00:42Z
    Last Sample Start:      2022-10-27T07:00:55Z
    Last Update Time:       2022-10-27T07:01:00Z
    Memory Histogram:
      Reference Timestamp:  2022-10-28T00:00:00Z
    Ref:
      Container Name:     mongodb
      Vpa Object Name:    mg-rs
    Total Samples Count:  3
    Version:              v3
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
      Reference Timestamp:  2022-10-27T00:00:00Z
      Total Weight:         0.3673624107285783
    First Sample Start:     2022-10-27T07:00:42Z
    Last Sample Start:      2022-10-27T07:00:55Z
    Last Update Time:       2022-10-27T07:01:00Z
    Memory Histogram:
      Reference Timestamp:  2022-10-28T00:00:00Z
    Ref:
      Container Name:     replication-mode-detector
      Vpa Object Name:    mg-rs
    Total Samples Count:  3
    Version:              v3
  Conditions:
    Last Transition Time:  2022-10-27T07:01:05Z
    Message:               Successfully created mongoDBOpsRequest demo/mops-mg-rs-cxhsy1
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2022-10-27T07:01:00Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  mongodb
        Lower Bound:
          Cpu:     400m
          Memory:  400Mi
        Target:
          Cpu:     400m
          Memory:  400Mi
        Uncapped Target:
          Cpu:     49m
          Memory:  262144k
        Upper Bound:
          Cpu:     1
          Memory:  1Gi
    Vpa Name:      mg-rs
Events:            <none>
```
So, the `mongodbautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `mongodbopsrequest` based on the recommendations, if the database pods are needed to scaled up or down.

Let's watch the `mongodbopsrequest` in the demo namespace to see if any `mongodbopsrequest` object is created. After some time you'll see that a `mongodbopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                    TYPE              STATUS       AGE
mops-mg-rs-cxhsy1       VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                    TYPE              STATUS       AGE
mops-mg-rs-cxhsy1       VerticalScaling   Successful   68s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-mg-rs-cxhsy1
Name:         mops-mg-rs-cxhsy1
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2022-10-27T07:01:05Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"ab03414a-67a2-4da4-8960-6e67ae56b503"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:timeout:
        f:type:
        f:verticalScaling:
          .:
          f:replicaSet:
            .:
            f:limits:
              .:
              f:cpu:
              f:memory:
            f:requests:
              .:
              f:cpu:
              f:memory:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2022-10-27T07:01:05Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:      kubedb-ops-manager
    Operation:    Update
    Subresource:  status
    Time:         2022-10-27T07:02:31Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MongoDBAutoscaler
    Name:                  mg-as-rs
    UID:                   ab03414a-67a2-4da4-8960-6e67ae56b503
  Resource Version:        640598
  UID:                     f7c6db00-dd0e-4850-8bad-5f0855ce3850
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mg-rs
  Timeout:  3m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Replica Set:
      Limits:
        Cpu:     400m
        Memory:  400Mi
      Requests:
        Cpu:     400m
        Memory:  400Mi
Status:
  Conditions:
    Last Transition Time:  2022-10-27T07:01:05Z
    Message:               MongoDB ops request is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-10-27T07:02:30Z
    Message:               Successfully Vertically Scaled Replicaset Resources
    Observed Generation:   1
    Reason:                UpdateReplicaSetResources
    Status:                True
    Type:                  UpdateReplicaSetResources
    Last Transition Time:  2022-10-27T07:02:31Z
    Message:               Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                     Age    From                         Message
  ----    ------                     ----   ----                         -------
  Normal  PauseDatabase              4m9s   KubeDB Ops-manager Operator  Pausing MongoDB demo/mg-rs
  Normal  PauseDatabase              4m9s   KubeDB Ops-manager Operator  Successfully paused MongoDB demo/mg-rs
  Normal  Starting                   4m9s   KubeDB Ops-manager Operator  Updating Resources of PetSet: mg-rs
  Normal  UpdateReplicaSetResources  4m9s   KubeDB Ops-manager Operator  Successfully updated replicaset Resources
  Normal  Starting                   4m9s   KubeDB Ops-manager Operator  Updating Resources of PetSet: mg-rs
  Normal  UpdateReplicaSetResources  4m9s   KubeDB Ops-manager Operator  Successfully updated replicaset Resources
  Normal  UpdateReplicaSetResources  2m44s  KubeDB Ops-manager Operator  Successfully Vertically Scaled Replicaset Resources
  Normal  ResumeDatabase             2m43s  KubeDB Ops-manager Operator  Resuming MongoDB demo/mg-rs
  Normal  ResumeDatabase             2m43s  KubeDB Ops-manager Operator  Successfully resumed MongoDB demo/mg-rs
  Normal  Successful                 2m43s  KubeDB Ops-manager Operator  Successfully Vertically Scaled Database
  Normal  UpdateReplicaSetResources  2m43s  KubeDB Ops-manager Operator  Successfully Vertically Scaled Replicaset Resources

```

Now, we are going to verify from the Pod, and the MongoDB yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo mg-rs-0 -o json | jq '.spec.containers[].resources'
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

$ kubectl get mongodb -n demo mg-rs -o json | jq '.spec.podTemplate.spec.resources'
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


The above output verifies that we have successfully auto scaled the resources of the MongoDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-rs
kubectl delete mongodbautoscaler -n demo mg-as-rs
```