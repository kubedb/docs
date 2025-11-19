---
title: Vertical Scaling Elasticsearch Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: kf-vertical-scaling-combined
    name: Combined Cluster
    parent: kf-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Elasticsearch Combined Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Elasticsearch combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Combined](/docs/guides/elasticsearch/clustering/combined-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Vertical Scaling Overview](/docs/guides/elasticsearch/scaling/vertical/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Combined Cluster

Here, we are going to deploy a `Elasticsearch` combined cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Elasticsearch Combined Cluster

Now, we are going to deploy a `Elasticsearch` combined cluster database with version `xpack-8.11.1`.

### Deploy Elasticsearch Combined Cluster

In this section, we are going to deploy a Elasticsearch combined cluster. Then, in the next section we will update the resources of the database using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-combined
  namespace: demo
spec:
  version: xpack-8.11.1
  enableSSL: true
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut

```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/clustering/multi-node-es.yaml
Elasticsearch.kubedb.com/es-combined created
```

Now, wait until `es-combined` has status `Ready`. i.e,

```bash
$ kubectl get elasticsearch -n demo -w
NAME          VERSION        STATUS   AGE
es-combined   xpack-8.11.1   Ready    3h17m

```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo es-combined-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}

```
This is the default resources of the Elasticsearch combined cluster set by the `KubeDB` operator.

We are now ready to apply the `ElasticsearchOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the combined cluster to meet the desired resources after scaling.

#### Create ElasticsearchOpsRequest

In order to update the resources of the database, we have to create a `ElasticsearchOpsRequest` CR with our desired resources. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: vscale-combined
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: es-combined
  verticalScaling:
    node:
      resources:
        limits:
          cpu: 1500m
          memory: 2Gi
        requests:
          cpu: 600m
          memory: 2Gi

```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `es-combined` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on Elasticsearch.
- `spec.VerticalScaling.node` specifies the desired resources after scaling.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/clustering/topology-es.yaml
```

#### Verify Elasticsearch Combined cluster resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get elasticsearchopsrequest -n demo
NAME              TYPE              STATUS       AGE
vscale-combined   VerticalScaling   Successful   2m38s

```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo vscale-combined
Name:         vscale-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-19T08:55:15Z
  Generation:          1
  Resource Version:    66012
  UID:                 bb814c10-12af-438e-9553-5565120bbdb9
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-combined
  Type:    VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     1500m
          Memory:  2Gi
        Requests:
          Cpu:     600m
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2025-11-19T08:55:15Z
    Message:               Elasticsearch ops request is vertically scaling the nodes
    Observed Generation:   1
    Reason:                VerticalScale
    Status:                True
    Type:                  VerticalScale
    Last Transition Time:  2025-11-19T08:55:27Z
    Message:               successfully reconciled the Elasticsearch resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-11-19T08:55:32Z
    Message:               pod exists; ConditionStatus:True; PodName:es-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-combined-0
    Last Transition Time:  2025-11-19T08:55:32Z
    Message:               create es client; ConditionStatus:True; PodName:es-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-combined-0
    Last Transition Time:  2025-11-19T08:55:32Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-combined-0
    Last Transition Time:  2025-11-19T08:55:32Z
    Message:               evict pod; ConditionStatus:True; PodName:es-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-combined-0
    Last Transition Time:  2025-11-19T08:55:57Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-19T08:55:57Z
    Message:               re enable shard allocation; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReEnableShardAllocation
    Last Transition Time:  2025-11-19T08:56:02Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-11-19T08:56:07Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-19T08:56:07Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                 Age   From                         Message
  ----     ------                                                                 ----  ----                         -------
  Normal   PauseDatabase                                                          2m6s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-combined
  Normal   UpdatePetSets                                                          114s  KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Warning  pod exists; ConditionStatus:True; PodName:es-combined-0                109s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-combined-0
  Warning  create es client; ConditionStatus:True; PodName:es-combined-0          109s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-combined-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-combined-0  109s  KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-combined-0
  Warning  evict pod; ConditionStatus:True; PodName:es-combined-0                 109s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-combined-0
  Warning  create es client; ConditionStatus:False                                104s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                 84s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                       84s   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Normal   RestartNodes                                                           79s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   UpdateDatabase                                                         74s   KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                                                         74s   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-combined
  Normal   ResumeDatabase                                                         74s   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-combined
  Normal   Successful                                                             74s   KubeDB Ops-manager Operator  Successfully Updated Database

```

Now, we are going to verify from one of the Pod yaml whether the resources of the combined cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo es-combined-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1500m",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "2Gi"
  }
}

```

The above output verifies that we have successfully scaled up the resources of the Elasticsearch combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo es-combined
kubectl delete Elasticsearchopsrequest -n demo vscale-combined
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/_index.md).
- Monitor your Elasticsearch database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Elasticsearch database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
