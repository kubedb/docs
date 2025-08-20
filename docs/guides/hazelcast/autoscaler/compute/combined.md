---
title: Hazelcast Combined Autoscaling
menu:
  docs_{{ .version }}:
    identifier: hz-auto-scaling-combined
    name: Combined Cluster
    parent: hz-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Hazelcast Combined Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Hazelcast combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastAutoscaler](/docs/guides/hazelcast/concepts/hazelcastautoscaler.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/hazelcast/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/hazelcast](/docs/examples/hazelcast) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Combined Cluster

Before deploying hazelcast we need to create license secret since we are running enterprise version of hazelcast.

```bash
kubectl create secret generic hz-license-key -n demo --from-literal=licenseKey=TrialLicense#10Nodes#eyJhbGxvd2VkTmF0aXZlTWVtb3J5U2l6ZSI6MTAwLCJhbGxvd2VkTnVtYmVyT2ZOb2RlcyI6MTAsImFsbG93ZWRUaWVyZWRTdG9yZVNpemUiOjAsImFsbG93ZWRUcGNDb3JlcyI6MCwiY3JlYXRpb25EYXRlIjoxNzQ4ODQwNDc3LjYzOTQ0NzgxNiwiZXhwaXJ5RGF0ZSI6MTc1MTQxNDM5OS45OTk5OTk5OTksImZlYXR1cmVzIjpbMCwyLDMsNCw1LDYsNyw4LDEwLDExLDEzLDE0LDE1LDE3LDIxLDIyXSwiZ3JhY2VQZXJpb2QiOjAsImhhemVsY2FzdFZlcnNpb24iOjk5LCJvZW0iOmZhbHNlLCJ0cmlhbCI6dHJ1ZSwidmVyc2lvbiI6IlY3In0=.6PYD6i-hejrJ5Czgc3nYsmnwF7mAI-78E8LFEuYp-lnzXh_QLvvsYx4ECD0EimqcdeG2J5sqUI06okLD502mCA==
secret/hz-license-key created
```

Here, we are going to deploy a `Hazelcast` Combined Cluster using a supported version by `KubeDB` operator. Then we are going to apply `HazelcastAutoscaler` to set up autoscaling.

#### Deploy Hazelcast Combined Cluster

In this section, we are going to deploy a Hazelcast Topology database with version `5.5.2`.  Then, in the next section we will set up autoscaling for this database using `HazelcastAutoscaler` CRD. Below is the YAML of the `Hazelcast` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hazelcast-dev
  namespace: demo
spec:
  replicas: 2
  version: 5.5.2
  licenseSecret:
    name: hz-license-key
  podTemplate:
    spec:
      containers:
        - name: hazelcast
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: longhorn
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Hazelcast` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/autoscaler/hazelcast-combined.yaml
hazelcast.kubedb.com/hazelcast-dev created
```

Now, wait until `hazelcast-dev` has status `Ready`. i.e,

```bash
$ kubectl get hz -n demo -w
NAME             TYPE                    VERSION   STATUS         AGE
hazelcast-dev    kubedb.com/v1alpha2     5.5.2     Provisioning   0s
hazelcast-dev    kubedb.com/v1alpha2     5.5.2     Provisioning   24s
.
.
hazelcast-dev    kubedb.com/v1alpha2     5.5.2     Ready          92s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo hazelcast-dev-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

Let's check the Hazelcast resources,
```bash
$ kubectl get hazelcast -n demo hazelcast-dev -o json | jq '.spec.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the hazelcast.

We are now ready to apply the `HazelcastAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a HazelcastAutoscaler Object.

#### Create HazelcastAutoscaler Object

In order to set up compute resource autoscaling for this combined cluster, we have to create a `HazelcastAutoscaler` CRO with our desired configuration. Below is the YAML of the `HazelcastAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: HazelcastAutoscaler
metadata:
  name: hz-combined-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: hazelcast-dev
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    hazelcast:
      trigger: "On"
      podLifeTimeThreshold: 2m
      resourceDiffPercentage: 1
      minAllowed:
        cpu: 600m
        memory: 1.6Gi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `hazelcast-dev` cluster.
- `spec.compute.node.trigger` specifies that compute autoscaling is enabled for this cluster.
- `spec.compute.node.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.node.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.node.minAllowed` specifies the minimum allowed resources for the cluster.
- `spec.compute.node.maxAllowed` specifies the maximum allowed resources for the cluster.
- `spec.compute.node.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.node.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields.
  - `timeout` specifies the timeout for the OpsRequest.
  - `apply` specifies when the OpsRequest should be applied. The default is "IfReady".

Let's create the `HazelcastAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/autoscaler/compute/hazelcast-combined-autoscaler.yaml
hazelcastautoscaler.autoscaling.kubedb.com/hz-combined-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `hazelcastautoscaler` resource is created successfully,

```bash
$ kubectl describe hazelcastautoscaler hz-combined-autoscaler -n demo
Name:         hz-combined-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         HazelcastAutoscaler
Metadata:
  Creation Timestamp:  2025-08-20T05:04:48Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Hazelcast
    Name:                  hazelcast-dev
    UID:                   ad17f549-4b10-4064-99fe-578894872a92
  Resource Version:        5631182
  UID:                     860b6bb9-55a0-48d1-b02f-35b7a4bb696d
Spec:
  Compute:
    Hazelcast:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  2Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  1717986918400m
      Pod Life Time Threshold:   2m0s
      Resource Diff Percentage:  1
      Trigger:                   On
  Database Ref:
    Name:  hazelcast-dev
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              9
        Weight:             10000
        Index:              10
        Weight:             2515
        Index:              11
        Weight:             6221
      Reference Timestamp:  2025-08-20T05:05:00Z
      Total Weight:         0.7947440525150773
    First Sample Start:     2025-08-20T05:05:19Z
    Last Sample Start:      2025-08-20T05:08:24Z
    Last Update Time:       2025-08-20T05:08:38Z
    Memory Histogram:
      Reference Timestamp:  2025-08-20T05:10:00Z
    Ref:
      Container Name:     hazelcast
      Vpa Object Name:    hazelcast-dev
    Total Samples Count:  6
    Version:              v3
  Conditions:
    Last Transition Time:  2025-08-20T05:06:11Z
    Message:               Successfully created HazelcastOpsRequest demo/hzops-hazelcast-dev-68lrza
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2025-08-20T05:05:38Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  hazelcast
        Lower Bound:
          Cpu:     600m
          Memory:  1717986918400m
        Target:
          Cpu:     600m
          Memory:  1717986918400m
        Uncapped Target:
          Cpu:     182m
          Memory:  380258472
        Upper Bound:
          Cpu:     1
          Memory:  2Gi
    Vpa Name:      hazelcast-dev
Events:            <none>

```
So, the `hazelcastautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `hazelcastopsrequest` based on the recommendations, if the database pods resources are needed to scaled up or down.

Let's watch the `hazelcastopsrequest` in the demo namespace to see if any `hazelcastopsrequest` object is created. After some time you'll see that a `hazelcastopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get hazelcastopsrequest -n demo
Every 2.0s: kubectl get hazelcastopsrequest -n demo
NAME                         TYPE              STATUS       AGE
hzops-hazelcast-dev-68lrza   VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get hazelcastopsrequest -n demo
NAME                         TYPE              STATUS       AGE
hzops-hazelcast-dev-68lrza VerticalScaling   Successful   3m2s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$kubectl describe hzops -n demo hzops-hazelcast-dev-68lrza 
Name:         hzops-hazelcast-dev-68lrza
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=hazelcast-dev
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=hazelcasts.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-20T05:06:11Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  HazelcastAutoscaler
    Name:                  hz-combined-autoscaler
    UID:                   860b6bb9-55a0-48d1-b02f-35b7a4bb696d
  Resource Version:        5631147
  UID:                     586fc38a-16d7-4a26-8c89-4c04395298dc
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   hazelcast-dev
  Timeout:  5m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Hazelcast:
      Resources:
        Limits:
          Memory:  1717986918
        Requests:
          Cpu:     600m
          Memory:  1717986918
Status:
  Conditions:
    Last Transition Time:  2025-08-20T05:06:11Z
    Message:               Hazelcast ops-request has started to vertically scaling the Hazelcast nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2025-08-20T05:06:14Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-20T05:08:24Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-20T05:06:24Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-dev-0
    Last Transition Time:  2025-08-20T05:06:24Z
    Message:               evict pod; ConditionStatus:True; PodName:hazelcast-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hazelcast-dev-0
    Last Transition Time:  2025-08-20T05:06:34Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-20T05:07:24Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-dev-1
    Last Transition Time:  2025-08-20T05:07:24Z
    Message:               evict pod; ConditionStatus:True; PodName:hazelcast-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hazelcast-dev-1
    Last Transition Time:  2025-08-20T05:08:24Z
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
  Normal   Starting                                                  4m40s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-hazelcast-dev-68lrza
  Normal   Starting                                                  4m40s  KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hazelcast-dev
  Normal   Successful                                                4m40s  KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hazelcast-dev for HazelcastOpsRequest: hzops-hazelcast-dev-68lrza
  Normal   UpdateStatefulSets                                        4m37s  KubeDB Ops-manager Operator  Successfully updated StatefulSets Resources
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-dev-0    4m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-dev-0
  Warning  evict pod; ConditionStatus:True; PodName:hazelcast-dev-0  4m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hazelcast-dev-0
  Warning  running pod; ConditionStatus:False                        4m17s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-dev-1    3m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-dev-1
  Warning  evict pod; ConditionStatus:True; PodName:hazelcast-dev-1  3m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hazelcast-dev-1
  Normal   RestartPods                                               2m27s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                  2m27s  KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hazelcast-dev
  Normal   Successful                                                2m27s  KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hazelcast-dev for HazelcastOpsRequest: hzops-hazelcast-dev-68lrza
```

Now, we are going to verify from the Pod, and the Hazelcast yaml whether the resources of the database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo hazelcast-dev-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1717986918"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1717986918"
  }
}



$ kubectl get hazelcast -n demo hazelcast-dev -o json | jq '.spec.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1717986918"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1717986918"
  }
}

```


The above output verifies that we have successfully auto scaled the resources of the Hazelcast combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcastopsrequest -n demo hzops-hazelcast-dev-68lrza 
kubectl delete hazelcastautoscaler -n demo hz-combined-autoscaler
kubectl delete hz -n demo hazelcast-dev
kubectl delete ns demo
```
## Next Steps

- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).
- Monitor your Hazelcast database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).

[//]: # (- Monitor your Hazelcast database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/hazelcast/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
