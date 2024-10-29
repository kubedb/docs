---
title: Solr Topology Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: sl-auto-scaling-topology
    name: Topology Cluster
    parent: sl-compute-autoscaling-solr
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of an Solr Topology Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. `cpu` and `memory` of an Solr topology cluster.

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

> **Note:** YAML files used in this tutorial are stored in this [directory](/docs/guides/Solr/autoscaler/compute/topology/yamls) of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Topology Cluster

Here, we are going to deploy an `Solr` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply `SolrAutoscaler` to set up autoscaling.

#### Deploy Solr Topology Cluster

In this section, we are going to deploy an Solr topology with SolrVersion `9.4.1`. Then, in the next section we will set up autoscaling for this database using `SolrAutoscaler` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.4.1
  zookeeperRef:
    name: zoo
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        storageClassName: longhorn
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        storageClassName: longhorn
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      storage:
        storageClassName: longhorn
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Let's create the `Solr` CRD we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscaler/topology.yaml
solr.kubedb.com/solr-cluster created
```

Now, wait until `solr-cluster` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.4.1     Ready    82s
```

Let's check an data node containers resources,

```bash
$ kubectl get pod -n demo solr-cluster-data-0 -o json | jq '.spec.containers[].resources'
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

Let's check the Solr CR for the data node resources,

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.data.podTemplate.spec.containers[0].resources'
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

Here, we are going to set up compute resource autoscaling using a SolrAutoscaler Object.

#### Create SolrAutoscaler Object

In order to set up compute resource autoscaling for the data nodes of the cluster, we have to create a `SolrAutoscaler` CRO with our desired configuration. Below is the YAML of the `SolrAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SolrAutoscaler
metadata:
  name: sl-data-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    data:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 5
      minAllowed:
        cpu: 1
        memory: 2.5Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `es-topology` cluster.
- `spec.compute.topology.data.trigger` specifies that compute autoscaling is enabled for the data nodes.
- `spec.compute.topology.data.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.topology.data.minAllowed` specifies the minimum allowed resources for the data nodes.
- `spec.compute.topology.data.maxAllowed` specifies the maximum allowed resources for the data nodes.
- `spec.compute.topology.data.controlledResources` specifies the resources that are controlled by the autoscaler.

> Note: In this demo, we are only setting up the autoscaling for the data nodes, that's why we only specified the data section of the autoscaler. You can enable autoscaling for the master and the data nodes in the same YAML, by specifying the `topology.master` and `topology.data` section, similar to the `topology.data` section we have configured in this demo.

Let's create the `SolrAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscaler/compute/topology-scaler.yaml
solrautoscaler.autoscaling.kubedb.com/sl-data-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `Solrautoscaler` resource is created successfully,

```bash
$ kubectl get solrautoscaler -n demo
NAME                 AGE
sl-data-autoscaler   94s

$ kubectl describe solrautoscaler -n demo sl-data-autoscaler
Name:         sl-data-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         SolrAutoscaler
Metadata:
  Creation Timestamp:  2024-10-29T13:39:52Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Solr
    Name:                  solr-cluster
    UID:                   b3fbf8b2-a05e-4502-be16-34b27b92a3ae
  Resource Version:        890538
  UID:                     6e95495b-2edf-4426-afa5-8de713bc3b2e
Spec:
  Compute:
    Data:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:                     1
        Memory:                  2.5Gi
      Pod Life Time Threshold:   5m
      Resource Diff Percentage:  5
      Trigger:                   On
    Overseer:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:                     1
        Memory:                  2.5Gi
      Pod Life Time Threshold:   5m
      Resource Diff Percentage:  5
      Trigger:                   On
  Database Ref:
    Name:  solr-cluster
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
      Reference Timestamp:  2024-10-29T13:40:00Z
      Total Weight:         0.09749048557222403
    First Sample Start:     2024-10-29T13:39:49Z
    Last Sample Start:      2024-10-29T13:39:49Z
    Last Update Time:       2024-10-29T13:40:07Z
    Memory Histogram:
      Reference Timestamp:  2024-10-29T13:45:00Z
    Ref:
      Container Name:     solr
      Vpa Object Name:    solr-cluster-data
    Total Samples Count:  1
    Version:              v3
  Conditions:
    Last Transition Time:  2024-10-29T13:40:35Z
    Message:               Successfully created solrOpsRequest demo/slops-solr-cluster-data-n3vjgi
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-10-29T13:40:07Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  solr
        Lower Bound:
          Cpu:     1
          Memory:  2560Mi
        Target:
          Cpu:     1
          Memory:  2560Mi
        Uncapped Target:
          Cpu:     100m
          Memory:  2162292018
        Upper Bound:
          Cpu:     2
          Memory:  3Gi
    Vpa Name:      solr-cluster-data
    Vpa Name:      solr-cluster-overseer
Events:            <none>

```

So, the `solrautoscaler` resource is created successfully.


As you can see from the output the vpa has generated a recommendation for the data node of the Solr cluster. Our autoscaler operator continuously watches the recommendation generated and creates an `Solropsrequest` based on the recommendations, if the Solr nodes are needed to be scaled up or down.

Let's watch the `solropsrequest` in the demo namespace to see if any `solropsrequest` object is created. After some time you'll see that an `Solropsrequest` will be created based on the recommendation.

```bash
$ kubectl get slops -n demo
NAME                             TYPE              STATUS        AGE
slops-solr-cluster-data-n3vjgi   VerticalScaling   Progressing   2m7s
```

Let's wait for the opsRequest to become successful.

```bash
$ kubectl get slops -n demo
NAME                             TYPE              STATUS       AGE
slops-solr-cluster-data-n3vjgi   VerticalScaling   Successful   2m38s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe slops -n demo slops-solr-cluster-data-n3vjgi 
Name:         slops-solr-cluster-data-n3vjgi
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=solr-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=solrs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-10-29T13:40:35Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  SolrAutoscaler
    Name:                  sl-data-autoscaler
    UID:                   6e95495b-2edf-4426-afa5-8de713bc3b2e
  Resource Version:        891009
  UID:                     e641e2cb-a3cd-4f34-8664-2d4a9079a93c
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   solr-cluster
  Timeout:  5m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Data:
      Resources:
        Limits:
          Memory:  2560Mi
        Requests:
          Cpu:     1
          Memory:  2560Mi
Status:
  Conditions:
    Last Transition Time:  2024-10-29T13:40:35Z
    Message:               Solr ops-request has started to vertically scaling the Solr nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-29T13:40:38Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-29T13:43:13Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-29T13:40:43Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-overseer-0
    Last Transition Time:  2024-10-29T13:40:43Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-overseer-0
    Last Transition Time:  2024-10-29T13:40:48Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-10-29T13:41:28Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-data-0
    Last Transition Time:  2024-10-29T13:41:28Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-data-0
    Last Transition Time:  2024-10-29T13:42:23Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-10-29T13:42:23Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-10-29T13:43:13Z
    Message:               Successfully completed the vertical scaling for RabbitMQ
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                               Age    From                         Message
  ----     ------                                                               ----   ----                         -------
  Normal   Starting                                                             3m10s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-solr-cluster-data-n3vjgi
  Normal   Starting                                                             3m10s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                                                           3m10s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: slops-solr-cluster-data-n3vjgi
  Normal   UpdatePetSets                                                        3m7s   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0       3m2s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0     3m2s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  running pod; ConditionStatus:False                                   2m57s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-data-0           2m17s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0         2m17s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0    82s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0  82s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Normal   RestartPods                                                          32s    KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                             32s    KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                                                           32s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: slops-solr-cluster-data-n3vjgi
```

Now, we are going to verify from the Pod, and the Solr YAML whether the resources of the data node of the cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo solr-cluster-data-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2560Mi"
  }
}

$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.data.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2560Mi"
  }
}

```

The above output verifies that we have successfully auto-scaled the resources of the Solr topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete solr -n demo solr-cluster
$ kubectl delete solrautoscaler -n demo sl-data-autoscaler
$ kubectl delete ns demo
```