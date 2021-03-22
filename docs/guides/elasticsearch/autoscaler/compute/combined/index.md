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
  
- Install `Vertical Pod Autoscaler` from [here](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#installation)

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

In this section, we are going to deploy an Elasticsearch combined cluster with ElasticsearchVersion `searchguard-7.9.3`.  Then, in the next section, we will set up autoscaling for this database using `ElasticsearchAutoscaler` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-combined
  namespace: demo
spec:
  enableSSL: true 
  version: searchguard-7.9.3
  storageType: Durable
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
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
NAME          VERSION             STATUS         AGE
es-combined   searchguard-7.9.3   Provisioning   1m2s
es-combined   searchguard-7.9.3   Provisioning   2m8s
es-combined   searchguard-7.9.3   Ready          2m8s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo es-combined-0 -o json | jq '.spec.containers[].resources'
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

Let's check the Elasticsearch resources,

```bash
$ kubectl get elasticsearch -n demo es-combined -o json | jq '.spec.podTemplate.spec.resources'
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
      minAllowed:
        cpu: ".4"
        memory: "1Gi"
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `es-combined` database.
- `spec.compute.node.trigger` specifies that compute resource autoscaling is enabled for this cluster.
- `spec.compute.node.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.node.minAllowed` specifies the minimum allowed resources for the Elasticsearch node.
- `spec.compute.node.maxAllowed` specifies the maximum allowed resources for the Elasticsearch node.
- `spec.compute.node.controlledResources` specifies the resources that are controlled by the autoscaler.

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
  Creation Timestamp:  2021-03-22T10:01:05Z
  Generation:          1
  Resource Version:  33465
  UID:               a5e671a5-22df-48bc-8949-e902270c85f4
Spec:
  Compute:
    Node:
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:                    400m
        Memory:                 1Gi
      Pod Life Time Threshold:  5m0s
      Trigger:                  On
  Database Ref:
    Name:  es-combined
Events:    <none>
```

So, the `elasticsearchautoscaler` resource is created successfully.

Now, let's verify that the vertical pod autoscaler (vpa) resource is created successfully,

```bash
$ kubectl get vpa -n demo
NAME              AGE
vpa-es-combined   1m32s

$ kubectl describe vpa -n demo vpa-es-combined 
Name:         vpa-es-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.k8s.io/v1
Kind:         VerticalPodAutoscaler
Metadata:
  Creation Timestamp:  2021-03-22T10:01:05Z
  Generation:          2
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-combined-as
    UID:                   a5e671a5-22df-48bc-8949-e902270c85f4
  Resource Version:        33488
  UID:                     5f49956c-ba9d-4896-a083-adc2a3138083
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
        Memory:  1Gi
  Target Ref:
    API Version:  apps/v1
    Kind:         StatefulSet
    Name:         es-combined
  Update Policy:
    Update Mode:  Off
Status:
  Conditions:
    Last Transition Time:  2021-03-22T10:00:18Z
    Status:                False
    Type:                  RecommendationProvided
Events:          <none>
```

So, we can verify from the above output that the `vpa` resource is created successfully. But you can see that the `RecommendationProvided` is false and also the `Recommendation` section of the `vpa` is empty. Let's wait some time and describe the vpa again.

```shell
$ kubectl describe vpa -n demo vpa-es-combined 
Name:         vpa-es-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.k8s.io/v1
Kind:         VerticalPodAutoscaler
Metadata:
  Creation Timestamp:  2021-03-22T10:01:05Z
  Generation:          2
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-combined-as
    UID:                   a5e671a5-22df-48bc-8949-e902270c85f4
  Resource Version:        33488
  UID:                     5f49956c-ba9d-4896-a083-adc2a3138083
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
        Memory:  1Gi
  Target Ref:
    API Version:  apps/v1
    Kind:         StatefulSet
    Name:         es-combined
  Update Policy:
    Update Mode:  Off
Status:
  Conditions:
    Last Transition Time:  2021-03-22T10:01:18Z
    Status:                True
    Type:                  RecommendationProvided
  Recommendation:
    Container Recommendations:
      Container Name:  elasticsearch
      Lower Bound:
        Cpu:     400m
        Memory:  1Gi
      Target:
        Cpu:     400m
        Memory:  1Gi
      Uncapped Target:
        Cpu:     126m
        Memory:  920733364
      Upper Bound:
        Cpu:     2
        Memory:  3Gi
Events:          <none>
```

As you can see from the output the vpa has generated a recommendation for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `elasticsearchopsrequest` based on the recommendations, if the database pods are needed to scale up or down.

Let's watch the `elasticsearchopsrequest` in the demo namespace to see if any `elasticsearchopsrequest` object is created. After some time you'll see that an `elasticsearchopsrequest` will be created based on the recommendation.

```bash
$  kubectl get elasticsearchopsrequest -n demo
NAME                           TYPE              STATUS       AGE
esops-vpa-es-combined-a6be0c   VerticalScaling   Progessing   1m
```

Let's wait for the opsRequest to become successful.

```bash
$  kubectl get elasticsearchopsrequest -n demo
NAME                           TYPE              STATUS       AGE
esops-vpa-es-combined-a6be0c   VerticalScaling   Successful   7m
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe elasticsearchopsrequest -n demo esops-vpa-es-combined-a6be0c
Name:         esops-vpa-es-combined-a6be0c
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=es-combined
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=elasticsearches.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2021-03-22T10:04:21Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-combined-as
    UID:                   a5e671a5-22df-48bc-8949-e902270c85f4
  Resource Version:        34809
  UID:                     d9f04043-7e42-42a5-827e-fe8b1b873425
Spec:
  Database Ref:
    Name:  es-combined
  Type:    VerticalScaling
  Vertical Scaling:
    Node:
      Limits:
        Cpu:     400m
        Memory:  1Gi
      Requests:
        Cpu:     400m
        Memory:  1Gi
Status:
  Conditions:
    Last Transition Time:  2021-03-22T10:04:21Z
    Message:               Elasticsearch ops request is vertically scaling the nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2021-03-22T10:04:21Z
    Message:               Successfully updated statefulSet resources.
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2021-03-22T10:11:21Z
    Message:               Successfully updated all node resources
    Observed Generation:   1
    Reason:                UpdateNodeResources
    Status:                True
    Type:                  UpdateNodeResources
    Last Transition Time:  2021-03-22T10:11:21Z
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
  Normal  PauseDatabase        7m    KubeDB Enterprise Operator  Pausing Elasticsearch demo/es-combined
  Normal  Updating             7m    KubeDB Enterprise Operator  Updating StatefulSets
  Normal  Updating             7m    KubeDB Enterprise Operator  Successfully Updated StatefulSets
  Normal  UpdateNodeResources  2m    KubeDB Enterprise Operator  Successfully updated all node resources
  Normal  Updating             2m    KubeDB Enterprise Operator  Updating Elasticsearch
  Normal  Updating             2m    KubeDB Enterprise Operator  Successfully Updated Elasticsearch
  Normal  ResumeDatabase       2m    KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-combined
  Normal  Successful           2m    KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify from the Pod, and the Elasticsearch YAML whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo es-combined-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "1Gi"
  }
}

$ kubectl get elasticsearch -n demo es-combined -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "1Gi"
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
