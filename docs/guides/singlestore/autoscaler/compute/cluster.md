---
title: SingleStore Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: sdb-auto-scaling-cluster
    name: Cluster
    parent: sdb-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Autoscaling the Compute Resource of a SingleStore Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a singlestore cluster for aggregator and leaf nodes.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStoreAutoscaler](/docs/guides/singlestore/concepts/autoscaler.md)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/singlestore/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/singlestore](/docs/examples/singlestore) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of SingleStore Cluster

Here, we are going to deploy a `SingleStore` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `SingleStoreAutoscaler` to set up autoscaling.

#### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

#### Deploy SingleStore Cluster

In this section, we are going to deploy a SingleStore with version `8.7.10`.  Then, in the next section we will set up autoscaling for this database using `SingleStoreAutoscaler` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: 8.7.10
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "0.7"
                requests:
                  memory: "2Gi"
                  cpu: "0.7"
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 3
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "0.7"
                requests:
                  memory: "2Gi"
                  cpu: "0.7"
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut

```
Let's create the `SingleStore` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/autoscaling/compute/sdb-cluster.yaml
singlestore.kubedb.com/sdb-cluster created
```

Now, wait until `sdb-sample` has status `Ready`. i.e,

```bash
NAME                                TYPE                  VERSION   STATUS   AGE
singlestore.kubedb.com/sdb-sample   kubedb.com/v1alpha2   8.7.10    Ready    4m35s
```

Let's check the aggregator pod containers resources,

```bash
kubectl get pod -n demo sdb-sample-aggregator-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "700m",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "2Gi"
  }
}
```

Let's check the SingleStore aggregator node resources,
```bash
kubectl get singlestore -n demo sdb-sample -o json | jq '.spec.topology.aggregator.podTemplate.spec.containers[] | select(.name == "singlestore") | .resources'
{
  "limits": {
    "cpu": "700m",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "2Gi"
  }
}

```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the singlestore.

We are now ready to apply the `SingleStoreAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a SingleStoreAutoscaler Object.

#### Create SingleStoreAutoscaler Object

In order to set up compute resource autoscaling for this singlestore cluster, we have to create a `SingleStoreAutoscaler` CRO with our desired configuration. Below is the YAML of the `SingleStoreAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SinglestoreAutoscaler
metadata:
  name: sdb-cluster-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: sdb-sample
  compute:
    aggregator:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 900m
        memory: 3Gi
      maxAllowed:
        cpu: 2000m
        memory: 6Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
      resourceDiffPercentage: 10
```


Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `sdb-sample` cluster.
- `spec.compute.aggregator.trigger` or `spec.compute.leaf.trigger` specifies that compute autoscaling is enabled for this cluster.
- `spec.compute.aggregator.podLifeTimeThreshold` or `spec.compute.leaf.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.aggregator.resourceDiffPercentage` or `spec.compute.leaf.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.aggregator.minAllowed` or `spec.compute.leaf.minAllowed` specifies the minimum allowed resources for the cluster.
- `spec.compute.aggregator.maxAllowed` or `spec.compute.leaf.maxAllowed` specifies the maximum allowed resources for the cluster.
- `spec.compute.aggregator.controlledResources` or `spec.compute.leaf.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.aggregator.containerControlledValues` or `spec.compute.leaf.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields.
  - `timeout` specifies the timeout for the OpsRequest.
  - `apply` specifies when the OpsRequest should be applied. The default is "IfReady".

Let's create the `SinglestoreAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/autoscaler/compute/sdb-cluster-autoscaler.yaml
singlestoreautoscaler.autoscaling.kubedb.com/sdb-cluster-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `singlestoreautoscaler` resource is created successfully,

```bash
$ kubectl describe singlestoreautoscaler -n demo sdb-cluster-autoscaler 
Name:         sdb-cluster-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         SinglestoreAutoscaler
Metadata:
  Creation Timestamp:  2024-09-10T08:55:26Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Singlestore
    Name:                  sdb-sample
    UID:                   f81d0592-9dda-428a-b0b4-e72ab3643e22
  Resource Version:        424275
  UID:                     6b7b3d72-b92f-4e6f-88eb-4e891c24c550
Spec:
  Compute:
    Aggregator:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  6Gi
      Min Allowed:
        Cpu:                     900m
        Memory:                  3Gi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  10
      Trigger:                   On
  Database Ref:
    Name:  sdb-sample
  Ops Request Options:
    Apply:  IfReady
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             2455
        Index:              1
        Weight:             2089
        Index:              2
        Weight:             10000
        Index:              3
        Weight:             361
      Reference Timestamp:  2024-09-10T09:05:00Z
      Total Weight:         5.5790751974302655
    First Sample Start:     2024-09-10T08:59:26Z
    Last Sample Start:      2024-09-10T09:15:18Z
    Last Update Time:       2024-09-10T09:15:27Z
    Memory Histogram:
      Bucket Weights:
        Index:              1
        Weight:             1821
        Index:              2
        Weight:             10000
      Reference Timestamp:  2024-09-10T09:05:00Z
      Total Weight:         14.365194626381038
    Ref:
      Container Name:     singlestore-coordinator
      Vpa Object Name:    sdb-sample-aggregator
    Total Samples Count:  32
    Version:              v3
    Cpu Histogram:
      Bucket Weights:
        Index:              5
        Weight:             3770
        Index:              6
        Weight:             10000
        Index:              7
        Weight:             132
        Index:              20
        Weight:             118
      Reference Timestamp:  2024-09-10T09:05:00Z
      Total Weight:         6.533759718059768
    First Sample Start:     2024-09-10T08:59:26Z
    Last Sample Start:      2024-09-10T09:16:19Z
    Last Update Time:       2024-09-10T09:16:28Z
    Memory Histogram:
      Bucket Weights:
        Index:              17
        Weight:             8376
        Index:              18
        Weight:             10000
      Reference Timestamp:  2024-09-10T09:05:00Z
      Total Weight:         17.827743425726553
    Ref:
      Container Name:     singlestore
      Vpa Object Name:    sdb-sample-aggregator
    Total Samples Count:  34
    Version:              v3
  Conditions:
    Last Transition Time:  2024-09-10T08:59:43Z
    Message:               Successfully created SinglestoreOpsRequest demo/sdbops-sdb-sample-aggregator-c0u141
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-09-10T08:59:42Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  singlestore
        Lower Bound:
          Cpu:     900m
          Memory:  3Gi
        Target:
          Cpu:     900m
          Memory:  3Gi
        Uncapped Target:
          Cpu:     100m
          Memory:  351198544
        Upper Bound:
          Cpu:     2
          Memory:  6Gi
    Vpa Name:      sdb-sample-aggregator
Events:            <none>
```
So, the `singlestoreautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `singlestoreopsrequest` based on the recommendations, if the database pods resources are needed to scaled up or down.

Let's watch the `singlestoreopsrequest` in the demo namespace to see if any `singlestoreopsrequest` object is created. After some time you'll see that a `singlestoreopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get singlestoreopsrequest -n demo
Every 2.0s: kubectl get singlestoreopsrequest -n demo
NAME                                       TYPE              STATUS       AGE
sdbops-sdb-sample-aggregator-c0u141     VerticalScaling    Progressing    10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                                       TYPE              STATUS       AGE
sdbops-sdb-sample-aggregator-c0u141      VerticalScaling    Successful    3m2s
```

We can see from the above output that the `SinglestoreOpsRequest` has succeeded. If we describe the `SinglestoreOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe singlestoreopsrequest -n demo sdbops-sdb-sample-aggregator-c0u141 
Name:         sdbops-sdb-sample-aggregator-c0u141
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sdb-sample
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=singlestores.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SinglestoreOpsRequest
Metadata:
  Creation Timestamp:  2024-09-10T08:59:43Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  SinglestoreAutoscaler
    Name:                  sdb-cluster-autoscaler
    UID:                   6b7b3d72-b92f-4e6f-88eb-4e891c24c550
  Resource Version:        406111
  UID:                     978a1a00-f217-4326-b103-f66bbccf2943
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  sdb-sample
  Type:    VerticalScaling
  Vertical Scaling:
    Aggregator:
      Resources:
        Limits:
          Cpu:     900m
          Memory:  3Gi
        Requests:
          Cpu:     900m
          Memory:  3Gi
Status:
  Conditions:
    Last Transition Time:  2024-09-10T09:01:55Z
    Message:               Timeout: request did not complete within requested timeout - context deadline exceeded
    Observed Generation:   1
    Reason:                Failed
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-09-10T08:59:46Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-09-10T08:59:46Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-09-10T09:01:21Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-09-10T08:59:52Z
    Message:               get pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sdb-sample-aggregator-0
    Last Transition Time:  2024-09-10T08:59:52Z
    Message:               evict pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sdb-sample-aggregator-0
    Last Transition Time:  2024-09-10T09:00:31Z
    Message:               check pod ready; ConditionStatus:True; PodName:sdb-sample-aggregator-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--sdb-sample-aggregator-0
    Last Transition Time:  2024-09-10T09:00:36Z
    Message:               get pod; ConditionStatus:True; PodName:sdb-sample-aggregator-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sdb-sample-aggregator-1
    Last Transition Time:  2024-09-10T09:00:36Z
    Message:               evict pod; ConditionStatus:True; PodName:sdb-sample-aggregator-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sdb-sample-aggregator-1
    Last Transition Time:  2024-09-10T09:01:16Z
    Message:               check pod ready; ConditionStatus:True; PodName:sdb-sample-aggregator-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--sdb-sample-aggregator-1
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                   Age   From                         Message
  ----     ------                                                                   ----  ----                         -------
  Normal   Starting                                                                 25m   KubeDB Ops-manager Operator  Start processing for SinglestoreOpsRequest: demo/sdbops-sdb-sample-aggregator-c0u141
  Normal   Starting                                                                 25m   KubeDB Ops-manager Operator  Pausing Singlestore database: demo/sdb-sample
  Normal   Successful                                                               25m   KubeDB Ops-manager Operator  Successfully paused Singlestore database: demo/sdb-sample for SinglestoreOpsRequest: sdbops-sdb-sample-aggregator-c0u141
  Normal   UpdatePetSets                                                            25m   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0           25m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0
  Warning  evict pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0         25m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0
  Warning  check pod ready; ConditionStatus:False; PodName:sdb-sample-aggregator-0  25m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:sdb-sample-aggregator-0
  Warning  check pod ready; ConditionStatus:True; PodName:sdb-sample-aggregator-0   24m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:sdb-sample-aggregator-0
  Warning  get pod; ConditionStatus:True; PodName:sdb-sample-aggregator-1           24m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sdb-sample-aggregator-1
  Warning  evict pod; ConditionStatus:True; PodName:sdb-sample-aggregator-1         24m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sdb-sample-aggregator-1
  Warning  check pod ready; ConditionStatus:False; PodName:sdb-sample-aggregator-1  24m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:sdb-sample-aggregator-1
  Warning  check pod ready; ConditionStatus:True; PodName:sdb-sample-aggregator-1   24m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:sdb-sample-aggregator-1
  Normal   RestartPods                                                              24m   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal Starting
  Normal Successful
```

Now, we are going to verify from the Pod, and the singlestore yaml whether the resources of the topology database has updated to meet up the desired state, Let's check,

```bash
kubectl get pod -n demo sdb-sample-aggregator-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "900m",
    "memory": "3Gi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "3Gi"
  }
}


kubectl get singlestore -n demo sdb-sample -o json | jq '.spec.topology.aggregator.podTemplate.spec.containers[] | select(.name == "singlestore") | .resources'
{
  "limits": {
    "cpu": "900m",
    "memory": "3Gi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "3Gi"
  }
}

```


The above output verifies that we have successfully auto scaled the resources of the SingleStore cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete singlestoreopsrequest -n demo sdbops-sdb-sample-aggregator-c0u141
kubectl delete singlestoreautoscaler -n demo sdb-cluster-autoscaler
kubectl delete kf -n demo sdb-sample
kubectl delete ns demo
```
## Next Steps

- Detail concepts of [SingleStore object](/docs/guides/singlestore/concepts/singlestore.md).
- Different SingleStore clustering modes [here](/docs/guides/singlestore/clustering/_index.md).
- Monitor your singlestore database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/singlestore/monitoring/using-prometheus-operator.md).
- Monitor your singlestore database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/singlestore/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

