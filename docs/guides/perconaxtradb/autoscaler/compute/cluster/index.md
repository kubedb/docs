---
title: PerconaXtraDB Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-autoscaling-compute-cluster
    name: Cluster
    parent: guides-perconaxtradb-autoscaling-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Autoscaling the Compute Resource of a PerconaXtraDB Cluster Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a PerconaXtraDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/perconaxtradb/concepts/perconaxtradb)
  - [PerconaXtraDBAutoscaler](/docs/guides/perconaxtradb/concepts/autoscaler)
  - [PerconaXtraDBOpsRequest](/docs/guides/perconaxtradb/concepts/opsrequest)
  - [Compute Resource Autoscaling Overview](/docs/guides/perconaxtradb/autoscaler/compute/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
## Autoscaling of Cluster Database

Here, we are going to deploy a `PerconaXtraDB` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `PerconaXtraDBAutoscaler` to set up autoscaling.

#### Deploy PerconaXtraDB Cluster

In this section, we are going to deploy a PerconaXtraDB Cluster with version `8.0.26`. Then, in the next section we will set up autoscaling for this database using `PerconaXtraDBAutoscaler` CRD. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,
> If you want to autoscale PerconaXtraDB `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
  storageType: Durable
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
          cpu: "200m"
          memory: "300Mi"
        limits:
          cpu: "200m"
          memory: "300Mi"
  terminationPolicy: WipeOut
```

Let's create the `PerconaXtraDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/autoscaler/compute/cluster/examples/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo
NAME             VERSION   STATUS   AGE
sample-pxc       8.0.26    Ready    14m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sample-pxc-0 -o json | jq '.spec.containers[].resources'
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

Let's check the PerconaXtraDB resources,
```bash
$ kubectl get perconaxtradb -n demo sample-pxc -o json | jq '.spec.podTemplate.spec.resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the perconaxtradb.

We are now ready to apply the `PerconaXtraDBAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a PerconaXtraDBAutoscaler Object.

#### Create PerconaXtraDBAutoscaler Object

In order to set up compute resource autoscaling for this database cluster, we have to create a `PerconaXtraDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `PerconaXtraDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PerconaXtraDBAutoscaler
metadata:
  name: px-as-compute
  namespace: demo
spec:
  databaseRef:
    name: sample-pxc
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    perconaxtradb:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 250m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      containerControlledValues: "RequestsAndLimits"
      controlledResources: ["cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `sample-pxc` database.
- `spec.compute.perconaxtradb.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.perconaxtradb.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.perconaxtradb.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.perconaxtradb.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.perconaxtradb.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.perconaxtradb.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.perconaxtradb.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions.apply` has two supported value : `IfReady` & `Always`.
Use `IfReady` if you want to process the opsReq only when the database is Ready. And use `Always` if you want to process the execution of opsReq irrespective of the Database state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest(in seconds).
If a step doesn't finish within the specified timeout, the ops request will result in failure.


Let's create the `PerconaXtraDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/autoscaler/compute/cluster/examples/pxas-compute.yaml
perconaxtradbautoscaler.autoscaling.kubedb.com/pxas-compute created
```

#### Verify Autoscaling is set up successfully

Let's check that the `perconaxtradbautoscaler` resource is created successfully,

```bash
$ kubectl get perconaxtradbautoscaler -n demo
NAME            AGE
px-as-compute   5m56s

$ kubectl describe perconaxtradbautoscaler px-as-compute -n demo
Name:         px-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         PerconaXtraDBAutoscaler
Metadata:
  Creation Timestamp:  2022-09-16T11:26:58Z
  Generation:          1
  Managed Fields:
  ...
  Resource Version:  846645
  UID:               44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58
Spec:
  Compute:
    Mariadb:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:                     250m
        Memory:                  400Mi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  sample-pxc
  Ops Request Options:
    Apply:    IfReady
    Timeout:  3m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
        Index:              46
        Weight:             555
      Reference Timestamp:  2022-09-16T00:00:00Z
      Total Weight:         2.648440345821337
    First Sample Start:     2022-09-16T11:26:48Z
    Last Sample Start:      2022-09-16T11:32:52Z
    Last Update Time:       2022-09-16T11:33:02Z
    Memory Histogram:
      Bucket Weights:
        Index:              1
        Weight:             10000
      Reference Timestamp:  2022-09-17T00:00:00Z
      Total Weight:         1.391848625060675
    Ref:
      Container Name:     px-coordinator
      Vpa Object Name:    sample-pxc
    Total Samples Count:  19
    Version:              v3
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
        Index:              3
        Weight:             556
      Reference Timestamp:  2022-09-16T00:00:00Z
      Total Weight:         2.648440345821337
    First Sample Start:     2022-09-16T11:26:48Z
    Last Sample Start:      2022-09-16T11:32:52Z
    Last Update Time:       2022-09-16T11:33:02Z
    Memory Histogram:
      Reference Timestamp:  2022-09-17T00:00:00Z
    Ref:
      Container Name:     perconaxtradb
      Vpa Object Name:    sample-pxc
    Total Samples Count:  19
    Version:              v3
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Successfully created mariaDBOpsRequest demo/pxops-sample-pxc-6xc1kc
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2022-09-16T11:27:02Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  perconaxtradb
        Lower Bound:
          Cpu:     250m
          Memory:  400Mi
        Target:
          Cpu:     250m
          Memory:  400Mi
        Uncapped Target:
          Cpu:     25m
          Memory:  262144k
        Upper Bound:
          Cpu:     1
          Memory:  1Gi
    Vpa Name:      sample-pxc
Events:            <none>

```
So, the `perconaxtradbautoscaler` resource is created successfully.

We can verify from the above output that `status.vpas` contains the `RecommendationProvided` condition to true. And in the same time, `status.vpas.recommendation.containerRecommendations` contain the actual generated recommendation.

Our autoscaler operator continuously watches the recommendation generated and creates an `perconaxtradbopsrequest` based on the recommendations, if the database pod resources are needed to scaled up or down. 

Let's watch the `perconaxtradbopsrequest` in the demo namespace to see if any `perconaxtradbopsrequest` object is created. After some time you'll see that a `perconaxtradbopsrequest` will be created based on the recommendation.

```bash
$ kubectl get perconaxtradbopsrequest -n demo
NAME                          TYPE              STATUS       AGE
pxops-sample-pxc-6xc1kc   VerticalScaling   Progressing  7s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get perconaxtradbopsrequest -n demo
NAME                              TYPE              STATUS       AGE
pxops-vpa-sample-pxc-z43wc8   VerticalScaling   Successful   3m32s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe perconaxtradbopsrequest -n demo pxops-vpa-sample-pxc-z43wc8
Name:         pxops-sample-pxc-6xc1kc
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PerconaXtraDBOpsRequest
Metadata:
  Creation Timestamp:  2022-09-16T11:27:07Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
          .:
          f:name:
        f:timeout:
        f:type:
        f:verticalScaling:
          .:
          f:perconaxtradb:
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
    Time:         2022-09-16T11:27:07Z
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
    Time:         2022-09-16T11:27:07Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  PerconaXtraDBAutoscaler
    Name:                  px-as-compute
    UID:                   44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58
  Resource Version:        846324
  UID:                     c2b30107-c6d3-44bb-adf3-135edc5d615b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   sample-pxc
  Timeout:  2m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Mariadb:
      Limits:
        Cpu:     250m
        Memory:  400Mi
      Requests:
        Cpu:     250m
        Memory:  400Mi
Status:
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Controller has started to Progress the PerconaXtraDBOpsRequest: demo/pxops-sample-pxc-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-09-16T11:30:42Z
    Message:               Successfully restarted PerconaXtraDB pods for PerconaXtraDBOpsRequest: demo/pxops-sample-pxc-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyRestatedStatefulSet
    Status:                True
    Type:                  RestartStatefulSet
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Vertical scale successful for PerconaXtraDBOpsRequest: demo/pxops-sample-pxc-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Controller has successfully scaled the PerconaXtraDB demo/pxops-sample-pxc-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Start processing for PerconaXtraDBOpsRequest: demo/pxops-sample-pxc-6xc1kc
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Pausing PerconaXtraDB databse: demo/sample-pxc
  Normal  Successful  8m48s  KubeDB Enterprise Operator  Successfully paused PerconaXtraDB database: demo/sample-pxc for PerconaXtraDBOpsRequest: pxops-sample-pxc-6xc1kc
  Normal  Starting    8m43s  KubeDB Enterprise Operator  Restarting Pod: demo/sample-pxc-0
  Normal  Starting    7m33s  KubeDB Enterprise Operator  Restarting Pod: demo/sample-pxc-1
  Normal  Starting    6m23s  KubeDB Enterprise Operator  Restarting Pod: demo/sample-pxc-2
  Normal  Successful  5m13s  KubeDB Enterprise Operator  Successfully restarted PerconaXtraDB pods for PerconaXtraDBOpsRequest: demo/pxops-sample-pxc-6xc1kc
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Vertical scale successful for PerconaXtraDBOpsRequest: demo/pxops-sample-pxc-6xc1kc
  Normal  Starting    5m8s   KubeDB Enterprise Operator  Resuming PerconaXtraDB database: demo/sample-pxc
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Successfully resumed PerconaXtraDB database: demo/sample-pxc
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Controller has Successfully scaled the PerconaXtraDB database: demo/sample-pxc
```

Now, we are going to verify from the Pod, and the PerconaXtraDB yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sample-pxc-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "400Mi"
  }
}

$ kubectl get perconaxtradb -n demo sample-pxc -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "400Mi"
  }
}
```


The above output verifies that we have successfully autoscaled the resources of the PerconaXtraDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete perconaxtradb -n demo sample-pxc
kubectl delete perconaxtradbautoscaler -n demo px-as-compute
kubectl delete ns demo
```