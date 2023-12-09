---
title: Elasticsearch Combined Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: es-auto-scaling-combined
    name: Combined
    parent: es-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Autoscaling the Compute Resource of an Elasticsearch Combined Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. `cpu` and `memory` of an Elasticsearch combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
  - [ElasticsearchAutoscaler](/docs/guides/elasticsearch/concepts/autoscaler/index.md)
  - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/elasticsearch/autoscaler/compute/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in this [directory](/docs/guides/elasticsearch/autoscaler/compute/combined/yamls) of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of a Combined Cluster

Here, we are going to deploy an `Elasticsearch` in combined cluster mode using a supported version by `KubeDB` operator. Then we are going to apply `ElasticsearchAutoscaler` to set up autoscaling.

### Deploy Elasticsearch standalone

In this section, we are going to deploy an Elasticsearch combined cluster with ElasticsearchVersion `xpack-8.11.1`.  Then, in the next section, we will set up autoscaling for this database using `ElasticsearchAutoscaler` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-combined
  namespace: demo
spec:
  enableSSL: true
  version: xpack-8.2.3
  storageType: Durable
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "500m"
        limits:
          cpu: "500m"
          memory: "1.2Gi"
  terminationPolicy: WipeOut
```

Let's create the `Elasticsearch` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/compute/combined/yamls/es-combined.yaml 
elasticsearch.kubedb.com/es-combined created
```

Now, wait until `es-combined` has status `Ready`. i.e,

```bash
$ kubectl get elasticsearch -n demo -w
NAME          VERSION       STATUS         AGE
es-combined   xpack-8.2.3   Provisioning   4s
es-combined   xpack-8.2.3   Provisioning   7s
....
....
es-combined   xpack-8.2.3   Ready          60s

```

Let's check the Pod containers resources,

```json
$ kubectl get pod -n demo es-combined-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1288490188800m"
  }
}
```

Let's check the Elasticsearch resources,

```json
$ kubectl get elasticsearch -n demo es-combined -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1288490188800m"
  }
}
```

You can see from the above outputs that the resources are the same as the ones we have assigned while deploying the Elasticsearch.

We are now ready to apply the `ElasticsearchAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (ie. `cpu` and `memory`) autoscaling using an ElasticsearchAutoscaler Object.

#### Create ElasticsearchAutoscaler Object

To set up compute resource autoscaling for this combined cluster, we have to create a `ElasticsearchAutoscaler` CRO with our desired configuration. Below is the YAML of the `ElasticsearchAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ElasticsearchAutoscaler
metadata:
  name: es-combined-as
  namespace: demo
spec:
  databaseRef:
    name: es-combined
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 5
      minAllowed:
        cpu: 1
        memory: "2.1Gi"
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `es-combined` database.
- `spec.compute.node.trigger` specifies that compute resource autoscaling is enabled for this cluster.
- `spec.compute.node.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.node.minAllowed` specifies the minimum allowed resources for the Elasticsearch node.
- `spec.compute.node.maxAllowed` specifies the maximum allowed resources for the Elasticsearch node.
- `spec.compute.node.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.node.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.node.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- - `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here : [timeout](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md#spectimeout), [apply](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md#specapply).

Let's create the `ElasticsearchAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/compute/combined/yamls/es-auto-scaler.yaml
elasticsearchautoscaler.autoscaling.kubedb.com/es-combined-as created
```

#### Verify Autoscaling is set up successfully

Let's check that the `elasticsearchautoscaler` resource is created successfully,

```bash
$kubectl get elasticsearchautoscaler -n demo
NAME             AGE
es-combined-as   14s

$ kubectl describe elasticsearchautoscaler -n demo  es-combined-as
Name:         es-combined-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         ElasticsearchAutoscaler
Metadata:
  Creation Timestamp:  2022-12-29T10:54:00Z
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
          f:node:
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
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-12-29T10:54:00Z
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
    Time:            2022-12-29T10:54:27Z
  Resource Version:  12469
  UID:               35640903-7aaf-46c6-9bc4-bd1771313e30
Spec:
  Compute:
    Node:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:                     1
        Memory:                  2254857830400m
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  5
      Trigger:                   On
  Database Ref:
    Name:  es-combined
  Ops Request Options:
    Apply:  IfReady
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             2849
        Index:              1
        Weight:             10000
        Index:              2
        Weight:             2856
        Index:              3
        Weight:             714
        Index:              5
        Weight:             714
        Index:              6
        Weight:             713
        Index:              7
        Weight:             714
        Index:              12
        Weight:             713
        Index:              21
        Weight:             713
        Index:              25
        Weight:             2138
      Reference Timestamp:  2022-12-29T00:00:00Z
      Total Weight:         4.257959878725071
    First Sample Start:     2022-12-29T10:54:03Z
    Last Sample Start:      2022-12-29T11:04:18Z
    Last Update Time:       2022-12-29T11:04:26Z
    Memory Histogram:
      Reference Timestamp:  2022-12-30T00:00:00Z
    Ref:
      Container Name:     elasticsearch
      Vpa Object Name:    es-combined
    Total Samples Count:  31
    Version:              v3
  Conditions:
    Last Transition Time:  2022-12-29T10:54:27Z
    Message:               Successfully created elasticsearchOpsRequest demo/esops-es-combined-ujb5hy
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2022-12-29T10:54:26Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  elasticsearch
        Lower Bound:
          Cpu:     1
          Memory:  2254857830400m
        Target:
          Cpu:     1
          Memory:  2254857830400m
        Uncapped Target:
          Cpu:     442m
          Memory:  1555165137
        Upper Bound:
          Cpu:     2
          Memory:  3Gi
    Vpa Name:      es-combined
Events:            <none>
```

So, the `elasticsearchautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation section`, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `elasticsearchopsrequest` based on the recommendations, if the database pods are needed to scaled up or down.

Let's watch the `elasticsearchopsrequest` in the demo namespace to see if any `elasticsearchopsrequest` object is created. After some time you'll see that an `elasticsearchopsrequest` will be created based on the recommendation.

```bash
$  kubectl get elasticsearchopsrequest -n demo
NAME                       TYPE              STATUS       AGE
esops-es-combined-ujb5hy   VerticalScaling   Progessing   1m
```

Let's wait for the opsRequest to become successful.

```bash
$  kubectl get elasticsearchopsrequest -n demo
NAME                       TYPE              STATUS       AGE
esops-es-combined-ujb5hy   VerticalScaling   Successful   1m
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe elasticsearchopsrequest -n demo esops-es-combined-ujb5hy
Name:         esops-es-combined-ujb5hy
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2022-12-29T10:54:27Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"35640903-7aaf-46c6-9bc4-bd1771313e30"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:type:
        f:verticalScaling:
          .:
          f:node:
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
    Time:         2022-12-29T10:54:27Z
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
    Time:         2022-12-29T10:54:27Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-combined-as
    UID:                   35640903-7aaf-46c6-9bc4-bd1771313e30
  Resource Version:        11992
  UID:                     4aa5295f-0702-45ac-9ae8-3cb496b0e740
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-combined
  Type:    VerticalScaling
  Vertical Scaling:
    Node:
      Limits:
        Cpu:     1
        Memory:  2254857830400m
      Requests:
        Cpu:     1
        Memory:  2254857830400m
Status:
  Conditions:
    Last Transition Time:  2022-12-29T10:54:27Z
    Message:               Elasticsearch ops request is vertically scaling the nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-12-29T10:54:39Z
    Message:               successfully reconciled the Elasticsearch resources
    Observed Generation:   1
    Reason:                Reconciled
    Status:                True
    Type:                  Reconciled
    Last Transition Time:  2022-12-29T10:58:39Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2022-12-29T10:58:44Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateElasticsearchCR
    Status:                True
    Type:                  UpdateElasticsearchCR
    Last Transition Time:  2022-12-29T10:58:45Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                         Message
  ----    ------                 ----   ----                         -------
  Normal  PauseDatabase          8m25s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-combined
  Normal  Reconciled             8m13s  KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Normal  RestartNodes           4m13s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal  UpdateElasticsearchCR  4m7s   KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal  ResumeDatabase         4m7s   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-combined
  Normal  Successful             4m7s   KubeDB Ops-manager Operator  Successfully Updated Database
```

Now, we are going to verify from the Pod, and the Elasticsearch YAML whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```json
$ kubectl get pod -n demo es-combined-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1288490188800m"
  }
}

$ kubectl get elasticsearch -n demo es-combined -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2254857830400m"
  },
  "requests": {
    "cpu": "1",
    "memory": "2254857830400m"
  }
}
```

The above output verifies that we have successfully auto-scaled the resources of the Elasticsearch standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete es -n demo es-combined 
$ kubectl delete elasticsearchautoscaler -n demo es-combined-as
$ kubectl delete ns demo
```
