---
title: Elasticsearch Topology Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: es-auto-scaling-topology
    name: Topology Cluster
    parent: es-compute-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of an Elasticsearch Topology Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. `cpu` and `memory` of an Elasticsearch topology cluster.

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

> **Note:** YAML files used in this tutorial are stored in this [directory](/docs/guides/elasticsearch/autoscaler/compute/topology/yamls) of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Topology Cluster

Here, we are going to deploy an `Elasticsearch` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply `ElasticsearchAutoscaler` to set up autoscaling.

#### Deploy Elasticsearch Topology Cluster

In this section, we are going to deploy an Elasticsearch topology with ElasticsearchVersion `opensearch-2.8.0`. Then, in the next section we will set up autoscaling for this database using `ElasticsearchAutoscaler` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-topology
  namespace: demo
spec:
  enableSSL: true 
  version: opensearch-2.8.0
  storageType: Durable
  topology:
    master:
      suffix: master
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      suffix: data
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      suffix: ingest
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Elasticsearch` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/computetopology/yamls/es-topology.yaml
elasticsearch.kubedb.com/es-topology created
```

Now, wait until `es-topology` has status `Ready`. i.e,

```bash
$ kubectl get elasticsearch -n demo -w
NAME          VERSION             STATUS         AGE
es-topology   opensearch-2.8.0   Provisioning   113s
es-topology   opensearch-2.8.0   Ready          115s
```

Let's check an ingest node containers resources,

```bash
$ kubectl get pod -n demo es-topology-ingest-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

Let's check the Elasticsearch CR for the ingest node resources,

```bash
$ kubectl get elasticsearch -n demo es-topology -o json | jq '.spec.topology.ingest.resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }

```

You can see from the above outputs that the resources are the same as the ones we have assigned while deploying the Elasticsearch.

We are now ready to apply the `ElasticsearchAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a ElasticsearchAutoscaler Object.

#### Create ElasticsearchAutoscaler Object

In order to set up compute resource autoscaling for the ingest nodes of the cluster, we have to create a `ElasticsearchAutoscaler` CRO with our desired configuration. Below is the YAML of the `ElasticsearchAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ElasticsearchAutoscaler
metadata:
  name: es-topology-as
  namespace: demo
spec:
  databaseRef:
    name: es-topology
  compute:
    ingest:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: ".4"
        memory: 500Mi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `es-topology` cluster.
- `spec.compute.topology.ingest.trigger` specifies that compute autoscaling is enabled for the ingest nodes.
- `spec.compute.topology.ingest.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.topology.ingest.minAllowed` specifies the minimum allowed resources for the ingest nodes.
- `spec.compute.topology.ingest.maxAllowed` specifies the maximum allowed resources for the ingest nodes.
- `spec.compute.topology.ingest.controlledResources` specifies the resources that are controlled by the autoscaler.

> Note: In this demo, we are only setting up the autoscaling for the ingest nodes, that's why we only specified the ingest section of the autoscaler. You can enable autoscaling for the master and the data nodes in the same YAML, by specifying the `topology.master` and `topology.data` section, similar to the `topology.ingest` section we have configured in this demo.

Let's create the `ElasticsearchAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/computetopology/yamls/es-topology-auto-scaler.yaml
elasticsearchautoscaler.autoscaling.kubedb.com/es-topology-as created
```

#### Verify Autoscaling is set up successfully

Let's check that the `elasticsearchautoscaler` resource is created successfully,

```bash
$ kubectl get elasticsearchautoscaler -n demo
NAME             AGE
es-topology-as   9s

$ kubectl describe elasticsearchautoscaler -n demo es-topology-as 
Name:         es-topology-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         ElasticsearchAutoscaler
Metadata:
  Creation Timestamp:  2021-03-22T13:03:55Z
  Generation:          1
  Resource Version:  18219
  UID:               c1855d8e-6430-48bb-87d7-9c7bc9ce6f42
Spec:
  Compute:
    Topology:
      Ingest:
        Controlled Resources:
          cpu
          memory
        Max Allowed:
          Cpu:     2
          Memory:  3Gi
        Min Allowed:
          Cpu:                    400m
          Memory:                 500Mi
        Pod Life Time Threshold:  5m0s
        Trigger:                  On
  Database Ref:
    Name:  es-topology
Events:    <none>
```

So, the `elasticsearchautoscaler` resource is created successfully.

Now, lets verify that the vertical pod autoscaler (vpa) resource is created successfully,

```bash
$ kubectl get vpa -n demo
NAME                     MODE   CPU    MEM          PROVIDED   AGE
vpa-es-topology-ingest   Off    400m   1102117711   True       30s

$ kubectl describe vpa -n demo vpa-es-topology-ingest 
Name:         vpa-es-topology-ingest
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.k8s.io/v1
Kind:         VerticalPodAutoscaler
Metadata:
  Creation Timestamp:  2021-03-22T13:03:55Z
  Generation:          2
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-topology-as
    UID:                   c1855d8e-6430-48bb-87d7-9c7bc9ce6f42
  Resource Version:        18253
  UID:                     1d32c133-7214-49bd-bf3b-aa4a99986058
Spec:
  Resource Policy:
    Container Policies:
      Container Name:  elasticsearch
      Controlled Resources:
        cpu
        memory
      Controlled Values:  RequestsAndLimits
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:     400m
        Memory:  500Mi
  Target Ref:
    API Version:  apps/v1
    Kind:         PetSet
    Name:         es-topology-ingest
  Update Policy:
    Update Mode:  Off
Status:
  Conditions:
    Last Transition Time:  2021-03-22T13:04:12Z
    Status:                True
    Type:                  RecommendationProvided
  Recommendation:
    Container Recommendations:
      Container Name:  elasticsearch
      Lower Bound:
        Cpu:     400m
        Memory:  1054147415
      Target:
        Cpu:     400m
        Memory:  1102117711
      Uncapped Target:
        Cpu:     224m
        Memory:  1102117711
      Upper Bound:
        Cpu:     2
        Memory:  3Gi
Events:          <none>
```

As you can see from the output the vpa has generated a recommendation for the ingest node of the Elasticsearch cluster. Our autoscaler operator continuously watches the recommendation generated and creates an `elasticsearchopsrequest` based on the recommendations, if the Elasticsearch nodes are needed to be scaled up or down.

Let's watch the `elasticsearchopsrequest` in the demo namespace to see if any `elasticsearchopsrequest` object is created. After some time you'll see that an `elasticsearchopsrequest` will be created based on the recommendation.

```bash
$ kubectl get elasticsearchopsrequest -n demo
NAME                                  TYPE              STATUS        AGE
esops-vpa-es-topology-ingest-37m2wi   VerticalScaling   Progressing   44s
```

Let's wait for the opsRequest to become successful.

```bash
$  kubectl get elasticsearchopsrequest -n demo -w
NAME                                  TYPE              STATUS        AGE
esops-vpa-es-topology-ingest-37m2wi   VerticalScaling   Progressing   8m2s
esops-vpa-es-topology-ingest-37m2wi   VerticalScaling   Successful    9m20s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ Name:         esops-vpa-es-topology-ingest-37m2wi
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=es-topology
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=elasticsearches.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2021-03-22T13:04:21Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-topology-as
    UID:                   c1855d8e-6430-48bb-87d7-9c7bc9ce6f42
  Resource Version:        19553
  UID:                     aed024b7-3779-416c-86c4-43120bba7bd3
Spec:
  Database Ref:
    Name:  es-topology
  Type:    VerticalScaling
  Vertical Scaling:
    Topology:
      Ingest:
        Limits:
          Cpu:     400m
          Memory:  1102117711
        Requests:
          Cpu:     400m
          Memory:  1102117711
Status:
  Conditions:
    Last Transition Time:  2021-03-22T13:04:21Z
    Message:               Elasticsearch ops request is vertically scaling the nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2021-03-22T13:04:21Z
    Message:               Successfully updated petSet resources.
    Observed Generation:   1
    Reason:                UpdatePetSetResources
    Status:                True
    Type:                  UpdatePetSetResources
    Last Transition Time:  2021-03-22T13:13:41Z
    Message:               Successfully updated all node resources
    Observed Generation:   1
    Reason:                UpdateNodeResources
    Status:                True
    Type:                  UpdateNodeResources
    Last Transition Time:  2021-03-22T13:13:41Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason               Age   From                        Message
  ----    ------               ----  ----                        -------
  Normal  PauseDatabase        10m   KubeDB Enterprise Operator  Pausing Elasticsearch demo/es-topology
  Normal  Updating             10m   KubeDB Enterprise Operator  Updating PetSets
  Normal  Updating             10m   KubeDB Enterprise Operator  Successfully Updated PetSets
  Normal  UpdateNodeResources  56s   KubeDB Enterprise Operator  Successfully updated all node resources
  Normal  Updating             56s   KubeDB Enterprise Operator  Updating Elasticsearch
  Normal  Updating             56s   KubeDB Enterprise Operator  Successfully Updated Elasticsearch
  Normal  ResumeDatabase       56s   KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-topology
  Normal  Successful           56s   KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify from the Pod, and the Elasticsearch YAML whether the resources of the ingest node of the cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo es-topology-ingest-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "1102117711"
  },
  "requests": {
    "cpu": "400m",
    "memory": "1102117711"
  }
}

$ kubectl get elasticsearch -n demo es-topology -o json | jq '.spec.topology.ingest.resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "1102117711"
  },
  "requests": {
    "cpu": "400m",
    "memory": "1102117711"
  }
}
```

The above output verifies that we have successfully auto-scaled the resources of the Elasticsearch topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete elasticsearch -n demo es-topology
$ kubectl delete elasticsearchautoscaler -n demo es-topology-as 
$ kubectl delete ns demo
```