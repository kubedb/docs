---
title: FerretDB Autoscaling
menu:
  docs_{{ .version }}:
    identifier: fr-auto-scaling-ferretdb
    name: ferretdbCompute
    parent: fr-compute-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a FerretDB

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a FerretDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
    - [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md)
    - [FerretDBAutoscaler](/docs/guides/ferretdb/concepts/autoscaler.md)
    - [FerretDBOpsRequest](/docs/guides/ferretdb/concepts/opsrequest.md)
    - [Compute Resource Autoscaling Overview](/docs/guides/ferretdb/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/ferretdb](/docs/examples/ferretdb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of FerretDB

Here, we are going to deploy a `FerretDB` standalone using a supported version by `KubeDB` operator. Backend postgres of this FerretDB will be internally managed by KubeDB, or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/ferretdb/concepts/appbinding.md) yourself.
Then we are going to apply `FerretDBAutoscaler` to set up autoscaling.

#### Deploy FerretDB

In this section, we are going to deploy a FerretDB with version `1.23.0`  Then, in the next section we will set up autoscaling for this ferretdb using `FerretDBAutoscaler` CRD. Below is the YAML of the `FerretDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb-autoscale
  namespace: demo
spec:
  version: "1.23.0"
  replicas: 1
  backend:
    externallyManaged: false
  podTemplate:
    spec:
      containers:
        - name: ferretdb
          resources:
            requests:
              cpu: "200m"
              memory: "300Mi"
            limits:
              cpu: "200m"
              memory: "300Mi"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  deletionPolicy: WipeOut
```

Let's create the `FerretDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/autoscaling/compute/ferretdb-autoscale.yaml
ferretdb.kubedb.com/ferretdb-autoscale created
```

Now, wait until `ferretdb-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get fr -n demo
NAME                 NAMESPACE   VERSION   STATUS   AGE
ferretdb-autoscale   demo        1.23.0    Ready    6m1s
```

Let's check the FerretDB resources,
```bash
$ kubectl get ferretdb -n demo ferretdb-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the ferretdb.

We are now ready to apply the `FerretDBAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a FerretDBAutoscaler Object.

#### Create FerretDBAutoscaler Object

In order to set up compute resource autoscaling for this ferretdb, we have to create a `FerretDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `FerretDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: FerretDBAutoscaler
metadata:
  name: ferretdb-autoscale-ops
  namespace: demo
spec:
  databaseRef:
    name: ferretdb-autoscale
  compute:
    ferretdb:
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

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `ferretdb-autoscale`.
- `spec.compute.ferretdb.trigger` specifies that compute resource autoscaling is enabled for this ferretdb.
- `spec.compute.ferretdb.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.replicaset.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.ferretdb.minAllowed` specifies the minimum allowed resources for this ferretdb.
- `spec.compute.ferretdb.maxAllowed` specifies the maximum allowed resources for this ferretdb.
- `spec.compute.ferretdb.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.ferretdb.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here :  [timeout](/docs/guides/ferretdb/concepts/opsrequest.md#spectimeout), [apply](/docs/guides/ferretdb/concepts/opsrequest.md#specapply).

Let's create the `FerretDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/autoscaling/compute/autoscaler.yaml
ferretdbautoscaler.autoscaling.kubedb.com/ferretdb-autoscaler-ops created
```

#### Verify Autoscaling is set up successfully

Let's check that the `ferretdbautoscaler` resource is created successfully,

```bash
$ kubectl get ferretdbautoscaler -n demo
NAME                   AGE
ferretdb-autoscale-ops   6m55s

$ kubectl describe ferretdbautoscaler ferretdb-autoscale-ops -n demo
Name:         ferretdb-autoscale-ops
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         FerretDBAutoscaler
Metadata:
  Creation Timestamp:  2024-10-14T08:30:37Z
  Generation:          1
  Resource Version:    11066
  UID:                 62387d58-1cd2-4cb6-9d97-91515531fcea
Spec:
  Compute:
    Ferretdb:
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
      Pod Life Time Threshold:   5m
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  ferretdb-autoscale
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
      Reference Timestamp:  2024-10-14T08:30:00Z
      Total Weight:         0.2536082343117003
    First Sample Start:     2024-10-14T08:31:16Z
    Last Sample Start:      2024-10-14T08:32:08Z
    Last Update Time:       2024-10-14T08:32:34Z
    Memory Histogram:
      Reference Timestamp:  2024-10-14T08:35:00Z
    Ref:
      Container Name:     ferretdb
      Vpa Object Name:    ferretdb-autoscale
    Total Samples Count:  2
    Version:              v3
  Conditions:
    Last Transition Time:  2024-10-14T08:32:29Z
    Message:               Successfully created FerretDBOpsRequest demo/frops-ferretdb-autoscale-5eo9wo
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-10-14T08:31:34Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  ferretdb
        Lower Bound:
          Cpu:     400m
          Memory:  400Mi
        Target:
          Cpu:     400m
          Memory:  400Mi
        Uncapped Target:
          Cpu:     100m
          Memory:  262144k
        Upper Bound:
          Cpu:     1
          Memory:  1Gi
    Vpa Name:      ferretdb-autoscale
Events:            <none>
```
So, the `ferretdbautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our ferretdb. Our autoscaler operator continuously watches the recommendation generated and creates an `ferretdbopsrequest` based on the recommendations, if the ferretdb pods are needed to scaled up or down.

Let's watch the `ferretdbopsrequest` in the demo namespace to see if any `ferretdbopsrequest` object is created. After some time you'll see that a `ferretdbopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME                               TYPE              STATUS        AGE
frops-ferretdb-autoscale-5eo9wo    VerticalScaling   Progressing   10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME                              TYPE              STATUS       AGE
frops-ferretdb-autoscale-5eo9wo   VerticalScaling   Successful   31s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed to scale the ferretdb.

```bash
$ kubectl describe ferretdbopsrequest -n demo frops-ferretdb-autoscale-5eo9wo
Name:         frops-ferretdb-autoscale-5eo9wo
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=ferretdb-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=ferretdbs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2024-10-14T08:32:29Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  FerretDBAutoscaler
    Name:                  ferretdb-autoscale-ops
    UID:                   62387d58-1cd2-4cb6-9d97-91515531fcea
  Resource Version:        11153
  UID:                     f14acbf1-bd46-4b93-9ee7-d944d9f1f8fd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ferretdb-autoscale
  Type:    VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     400m
          Memory:  400Mi
        Requests:
          Cpu:     400m
          Memory:  400Mi
Status:
  Conditions:
    Last Transition Time:  2024-10-14T08:32:29Z
    Message:               FerretDB ops-request has started to vertically scaling the FerretDB nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-14T08:32:32Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-14T08:32:32Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-14T08:32:37Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-autoscale-0
    Last Transition Time:  2024-10-14T08:32:37Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-autoscale-0
    Last Transition Time:  2024-10-14T08:32:42Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-autoscale-0
    Last Transition Time:  2024-10-14T08:32:47Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-14T08:32:48Z
    Message:               Successfully completed the VerticalScaling for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                 Age    From                         Message
  ----     ------                                                                 ----   ----                         -------
  Normal   Starting                                                               3m7s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/frops-ferretdb-autoscale-5eo9wo
  Normal   Starting                                                               3m7s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/ferretdb-autoscale
  Normal   Successful                                                             3m7s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/ferretdb-autoscale for FerretDBOpsRequest: frops-ferretdb-autoscale-5eo9wo
  Normal   UpdatePetSets                                                          3m4s   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-autoscale-0            2m59s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-autoscale-0
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-autoscale-0          2m59s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-autoscale-0
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-autoscale-0  2m54s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-autoscale-0
  Normal   RestartPods                                                            2m49s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                               2m49s  KubeDB Ops-manager Operator  Resuming FerretDB database: demo/ferretdb-autoscale
  Normal   Successful                                                             2m48s  KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/ferretdb-autoscale for FerretDBOpsRequest: frops-ferretdb-autoscale-5eo9wo
```

Now, we are going to verify from the Pod, and the FerretDB yaml whether the resources of the ferretdb has updated to meet up the desired state, Let's check,

```bash
$ kubectl get ferretdb -n demo ferretdb-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
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


The above output verifies that we have successfully auto-scaled the resources of the FerretDB.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete fr -n demo ferretdb-autoscale
kubectl delete ferretdbautoscaler -n demo ferretdb-autoscale-ops
```