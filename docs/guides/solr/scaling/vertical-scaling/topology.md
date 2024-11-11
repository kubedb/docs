---
title: Solr Vertical Scaling Topology
menu:
  docs_{{ .version }}:
    identifier: sl-scaling-vertical-topology
    name: Solr Vertical Scaling Topology
    parent: sl-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Solr Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Solr topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [Topology](/docs/guides/solr/clustering/topology_cluster.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Vertical Scaling Overview](/docs/guides/solr/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/solr](/docs/examples/solr) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Topology Cluster

Here, we are going to deploy a `Solr` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Solr Topology Cluster

Now, we are going to deploy a `Solr` topology cluster database with version `3.6.1`.

### Deploy Solr Topology Cluster

In this section, we are going to deploy a Solr topology cluster. Then, in the next section we will update the resources of the database using `SolrOpsRequest` CRD. Below is the YAML of the `Solr` CR that we are going to create,

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
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi

```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/scaling/vertical/topology/solr.yaml
solr.kubedb.com/solr-cluster created
```

Now, wait until `solr-cluster` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.4.1     Ready    63m
```

Let's check the Pod containers resources for `data`, `overseer` and `coordinator` of the solr topology cluster. Run the following command to get the resources of the `broker` and `controller` containers of the Solr topology cluster

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

```bash
$ kubectl get pod -n demo solr-cluster-overseer-0 -o json | jq '.spec.containers[].resources'
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

```bash
$ kubectl get pod -n demo solr-cluster-coordinator-0 -o json | jq '.spec.containers[].resources'
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
This is the default resources of the Solr topology cluster set by the `KubeDB` operator.

We are now ready to apply the `SolrOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the topology cluster to meet the desired resources after scaling.

#### Create SolrOpsRequest

In order to update the resources of the database, we have to create a `SolrOpsRequest` CR with our desired resources. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-vscale-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: VerticalScaling
  verticalScaling:
    data:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
    overseer:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
    coordinator:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `solr-cluster` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on Solr.
- `spec.VerticalScaling.data`, `spec.VerticalScaling.overseer` and `spec.VerticalScaling.coordinator` specifies the desired resources for topologies after scaling.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/scaling/vertical/topology/scaling.yaml
solropsrequest.ops.kubedb.com/slops-slops-vscale-topology-topology created
```

#### Verify Solr Topology cluster resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Solr` object and related `PetSets` and `Pods`.

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CR,

```bash
$ kubectl get slops -n demo
NAME     TYPE              STATUS       AGE
slops-vscale-topology   VerticalScaling   Successful   3m9s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe slops -n demo slops-vscale-topology 
Name:         slops-vscale-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-11T11:24:23Z
  Generation:          1
  Resource Version:    2353035
  UID:                 f30a3bd6-9903-4747-96af-e4f50948afbc
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Type:    VerticalScaling
  Vertical Scaling:
    Coordinator:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2.5Gi
        Requests:
          Cpu:     1
          Memory:  2.5Gi
    Data:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2.5Gi
        Requests:
          Cpu:     1
          Memory:  2.5Gi
    Overseer:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2.5Gi
        Requests:
          Cpu:     1
          Memory:  2.5Gi
Status:
  Conditions:
    Last Transition Time:  2024-11-11T11:24:23Z
    Message:               Solr ops-request has started to vertically scaling the Solr nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-11-11T11:24:23Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-11T11:26:58Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-11T11:24:28Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-11T11:24:28Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-11T11:24:33Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-11T11:25:18Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-data-0
    Last Transition Time:  2024-11-11T11:25:18Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-data-0
    Last Transition Time:  2024-11-11T11:26:08Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-11T11:26:08Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-11T11:26:58Z
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
  Normal   Starting                                                             3m55s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-vscale-topology
  Normal   UpdatePetSets                                                        3m55s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0       3m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0     3m50s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  running pod; ConditionStatus:False                                   3m45s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-data-0           3m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0         3m     KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0    2m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0  2m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Normal   RestartPods                                                          80s    KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                             80s    KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                                                           80s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: slops-vscale-topology
```
Now, we are going to verify from one of the Pod yaml whether the resources of the topology cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo solr-cluster-coordinator-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2560Mi"
  }
}
$ kubectl get pod -n demo solr-cluster-data-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2560Mi"
  }
}
$ kubectl get pod -n demo solr-cluster-overseer-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2560Mi"
  }
}

```

The above output verifies that we have successfully scaled up the resources of the Solr topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete sl -n demo solr-cluster
kubectl delete solropsrequest -n demo slops-vscale-topology
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
