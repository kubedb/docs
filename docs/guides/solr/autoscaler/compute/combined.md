---
title: Solr Combined Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: sl-auto-scaling-combined
    name: Combined
    parent: sl-compute-autoscaling-solr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of an Solr Combined Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. `cpu` and `memory` of an Solr combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrAutoscaler](/docs/guides/solr/concepts/autoscaler.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Compute Resource Autoscaling Overview](/docs/guides/solr/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in this [directory](/docs/examples/solr/autoscaler) of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of a Combined Cluster

Here, we are going to deploy an `Solr` in combined cluster mode using a supported version by `KubeDB` operator. Then we are going to apply `SolrAutoscaler` to set up autoscaling.

### Deploy Solr Combined

In this section, we are going to deploy an Solr combined cluster with SolrVersion `9.6.1`.  Then, in the next section, we will set up autoscaling for this database using `SolrAutoscaler` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-combined
  namespace: demo
spec:
  version: 9.6.1
  replicas: 2
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `Solr` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscalers/combined.yaml
solr.kubedb.com/solr-combined created
```

Now, wait until `es-combined` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME            TYPE                  VERSION   STATUS   AGE
solr-combined   kubedb.com/v1alpha2   9.6.1     Ready    83s

```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo solr-combined-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "2Gi"
  }
}
```

Let's check the Solr resources,

```bash
$ kubectl get solr -n demo solr-combined -o json | jq '.spec.podTemplate.spec.containers[] | select(.name == "solr") | .resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "2Gi"
  }
}

```

You can see from the above outputs that the resources are the same as the ones we have assigned while deploying the Solr.

We are now ready to apply the `SolrAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (ie. `cpu` and `memory`) autoscaling using an SolrAutoscaler Object.

#### Create SolrAutoscaler Object

To set up compute resource autoscaling for this combined cluster, we have to create a `SolrAutoscaler` CRO with our desired configuration. Below is the YAML of the `SolrAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SolrAutoscaler
metadata:
  name: sl-node-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: solr-combined
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 5
      minAllowed:
        cpu: 1
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `solr-combined` database.
- `spec.compute.node.trigger` specifies that compute resource autoscaling is enabled for this cluster.
- `spec.compute.node.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.node.minAllowed` specifies the minimum allowed resources for the Solr node.
- `spec.compute.node.maxAllowed` specifies the maximum allowed resources for the Solr node.
- `spec.compute.node.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.node.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.node.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- - `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here : [timeout](/docs/guides/Solr/concepts/Solr-ops-request/index.md#spectimeout), [apply](/docs/guides/Solr/concepts/Solr-ops-request/index.md#specapply).

Let's create the `SolrAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscaler/compute/combined-scaler.yaml
solrautoscaler.autoscaling.kubedb.com/sl-node-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `Solrautoscaler` resource is created successfully,

```bash
$ kubectl get solrautoscaler -n demo
NAME                 AGE
sl-node-autoscaler   100s

$ kubectl describe solrautoscaler -n demo sl-node-autoscaler
Name:         sl-node-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         SolrAutoscaler
Metadata:
  Creation Timestamp:  2024-10-29T12:29:57Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Solr
    Name:                  solr-combined
    UID:                   422bb16e-2181-4ce3-9401-a4feef853b4e
  Resource Version:        883971
  UID:                     33bf5f3b-c6ad-4234-8119-fccebcb8d4b6
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
        Memory:                  2Gi
      Pod Life Time Threshold:   5m
      Resource Diff Percentage:  5
      Trigger:                   On
  Database Ref:
    Name:  solr-combined
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             7649
        Index:              1
        Weight:             7457
        Index:              3
        Weight:             10000
        Index:              4
        Weight:             9817
      Reference Timestamp:  2024-10-29T12:30:00Z
      Total Weight:         0.44924508934014207
    First Sample Start:     2024-10-29T12:29:42Z
    Last Sample Start:      2024-10-29T12:31:49Z
    Last Update Time:       2024-10-29T12:32:05Z
    Memory Histogram:
      Reference Timestamp:  2024-10-29T12:35:00Z
    Ref:
      Container Name:     solr
      Vpa Object Name:    solr-combined
    Total Samples Count:  4
    Version:              v3
  Conditions:
    Last Transition Time:  2024-10-29T12:30:35Z
    Message:               Successfully created solrOpsRequest demo/slops-solr-combined-04xbzd
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-10-29T12:30:05Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  solr
        Lower Bound:
          Cpu:     1
          Memory:  2Gi
        Target:
          Cpu:     1
          Memory:  2Gi
        Uncapped Target:
          Cpu:     100m
          Memory:  1555165137
        Upper Bound:
          Cpu:     2
          Memory:  3Gi
    Vpa Name:      solr-combined
Events:            <none>
```

So, the `Solrautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation section`, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `solropsrequest` based on the recommendations, if the database pods are needed to scaled up or down.

Let's watch the `solropsrequest` in the demo namespace to see if any `solropsrequest` object is created. After some time you'll see that an `Solropsrequest` will be created based on the recommendation.

```bash
$ kubectl get slops -n demo
NAME                         TYPE              STATUS        AGE
slops-solr-combined-04xbzd   VerticalScaling   Progressing   2m24s
```

Let's wait for the opsRequest to become successful.

```bash
$ kubectl get slops -n demo
NAME                         TYPE              STATUS       AGE
slops-solr-combined-04xbzd   VerticalScaling   Successful   2m24s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe slops -n demo slops-solr-combined-04xbzd
Name:         slops-solr-combined-04xbzd
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=solr-combined
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=solrs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-10-29T12:30:35Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  SolrAutoscaler
    Name:                  sl-node-autoscaler
    UID:                   33bf5f3b-c6ad-4234-8119-fccebcb8d4b6
  Resource Version:        883905
  UID:                     709d9b24-cd19-4605-bb41-92d099758ec0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   solr-combined
  Timeout:  5m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Memory:  2Gi
        Requests:
          Cpu:     1
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-29T12:30:35Z
    Message:               Solr ops-request has started to vertically scaling the Solr nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-29T12:30:38Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-29T12:31:23Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-29T12:30:43Z
    Message:               get pod; ConditionStatus:True; PodName:solr-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-combined-0
    Last Transition Time:  2024-10-29T12:30:43Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-combined-0
    Last Transition Time:  2024-10-29T12:30:48Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-10-29T12:31:03Z
    Message:               get pod; ConditionStatus:True; PodName:solr-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-combined-1
    Last Transition Time:  2024-10-29T12:31:03Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-combined-1
    Last Transition Time:  2024-10-29T12:31:23Z
    Message:               Successfully completed the vertical scaling for RabbitMQ
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  3m35s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-solr-combined-04xbzd
  Normal   Starting                                                  3m35s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-combined
  Normal   Successful                                                3m35s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-combined for SolrOpsRequest: slops-solr-combined-04xbzd
  Normal   UpdatePetSets                                             3m32s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:solr-combined-0    3m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-combined-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-combined-0  3m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-combined-0
  Warning  running pod; ConditionStatus:False                        3m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-combined-1    3m7s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-combined-1
  Warning  evict pod; ConditionStatus:True; PodName:solr-combined-1  3m7s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-combined-1
  Normal   RestartPods                                               2m47s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                  2m47s  KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-combined
  Normal   Successful                                                2m47s  KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-combined for SolrOpsRequest: slops-solr-combined-04xbzd
```

Now, we are going to verify from the Pod, and the Solr YAML whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo solr-combined-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}

$ kubectl get solr -n demo solr-combined -o json | jq '.spec.podTemplate.spec.containers[] | select(.name == "solr") | .resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}

```

The above output verifies that we have successfully auto-scaled the resources of the Solr standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete sl -n demo solr-combined 
$ kubectl delete solrautoscaler -n demo sl-node-autoscaler
$ kubectl delete ns demo
```
