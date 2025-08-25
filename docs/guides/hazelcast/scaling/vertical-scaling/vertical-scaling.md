---
title: Vertical Scaling Hazelcast
menu:
  docs_{{ .version }}:
    identifier: hazelcast-vertical-scaling
    name: Vertical Scaling
    parent: hz-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Hazelcast

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a Hazelcast database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/hazelcast/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Apply Vertical Scaling on Hazelcast

Here, we are going to deploy a `Hazelcast` database using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Hazelcast Database

Now, we are going to deploy a `Hazelcast` database with version `5.5.2`.

### Deploy Hazelcast

Before deploying hazelcast we need to create license secret since we are running enterprise version of hazelcast.

```bash
kubectl create secret generic hz-license-key -n demo --from-literal=licenseKey=TrialLicense#10Nodes#eyJhbGxvd2VkTmF0aXZlTWVtb3J5U2l6ZSI6MTAwLCJhbGxvd2VkTnVtYmVyT2ZOb2RlcyI6MTAsImFsbG93ZWRUaWVyZWRTdG9yZVNpemUiOjAsImFsbG93ZWRUcGNDb3JlcyI6MCwiY3JlYXRpb25EYXRlIjoxNzQ4ODQwNDc3LjYzOTQ0NzgxNiwiZXhwaXJ5RGF0ZSI6MTc1MTQxNDM5OS45OTk5OTk5OTksImZlYXR1cmVzIjpbMCwyLDMsNCw1LDYsNyw4LDEwLDExLDEzLDE0LDE1LDE3LDIxLDIyXSwiZ3JhY2VQZXJpb2QiOjAsImhhemVsY2FzdFZlcnNpb24iOjk5LCJvZW0iOmZhbHNlLCJ0cmlhbCI6dHJ1ZSwidmVyc2lvbiI6IlY3In0=.6PYD6i-hejrJ5Czgc3nYsmnwF7mAI-78E8LFEuYp-lnzXh_QLvvsYx4ECD0EimqcdeG2J5sqUI06okLD502mCA==
secret/hz-license-key created
```

In this section, we are going to deploy a Hazelcast database. Then, in the next section we will update the resources using `HazelcastOpsRequest` CRD. Below is the YAML of the `Hazelcast` CR that we are going to create,

```yaml

```

Let's create the `Hazelcast` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/scaling/vertical-scaling/hazelcast.yaml
hazelcast.kubedb.com/hz-prod created
```

Now, wait until `hz-prod` has status `Ready`. i.e,

```bash
$ kubectl get hz -n demo
NAME      TYPE                  VERSION   STATUS   AGE
hz-prod   kubedb.com/v1alpha2   5.5.2     Ready    4m
```

Let's check the container resources,

```bash
$ kubectl get pod -n demo hz-prod-0 -o json | jq '.spec.containers[].resources'
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

You can see the container has `500m` CPU and `1Gi` memory as resource limits.

We are now ready to apply the `HazelcastOpsRequest` CR to update the resources of this database.

## Vertical Scaling

Here, we are going to update the resources of the database to meet the desired resources after scaling.

### Create HazelcastOpsRequest

In order to update the resources of the database, we have to create a `HazelcastOpsRequest` CR with our desired resources. Below is the YAML of the `HazelcastOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hz-vscale-up
  namespace: demo
spec:
  databaseRef:
    name: hz-prod
  type: VerticalScaling
  verticalScaling:
    hazelcast:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `hz-prod` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.hazelcast` specifies the desired resources after scaling.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/scaling/vertical-scaling/hz-vscale-up.yaml
hazelcastopsrequest.ops.kubedb.com/hz-vscale-up created
```

### Verify Hazelcast resources updated successfully

If everything goes well, `KubeDB` Enterprise operator will update the resources of `Hazelcast` object and related `PetSets` and `Pods`.

Let's wait for `HazelcastOpsRequest` to be `Successful`. Run the following command to watch `HazelcastOpsRequest` CR,

```bash
$ kubectl get hazelcastopsrequest -n demo
NAME           TYPE              STATUS       AGE
hz-vscale-up   VerticalScaling   Successful   3m2s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed to update the database.

```bash
$ kubectl describe hazelcastopsrequest -n demo hz-vscale-up
Name:         hz-vscale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T11:10:56Z
  Generation:          1
  Resource Version:    5478364
  UID:                 e0c7e3a5-b04f-4756-a70f-aec54235b9ad
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hz-prod
  Type:    VerticalScaling
  Vertical Scaling:
    Hazelcast:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2.5Gi
        Requests:
          Cpu:     1
          Memory:  2.5Gi
Status:
  Conditions:
    Last Transition Time:  2025-08-19T11:10:56Z
    Message:               Hazelcast ops-request has started to vertically scaling the Hazelcast nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2025-08-19T11:11:00Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-19T11:12:50Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-19T11:11:10Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-0
    Last Transition Time:  2025-08-19T11:11:10Z
    Message:               evict pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hz-prod-0
    Last Transition Time:  2025-08-19T11:11:20Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-19T11:12:00Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-1
    Last Transition Time:  2025-08-19T11:12:00Z
    Message:               evict pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hz-prod-1
    Last Transition Time:  2025-08-19T11:12:50Z
    Message:               Successfully completed the vertical scaling for RabbitMQ
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                              Age    From                         Message
  ----     ------                                              ----   ----                         -------
  Normal   Starting                                            3m1s   KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hz-vscale-up
  Normal   Starting                                            3m1s   KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hz-prod
  Normal   Successful                                          3m     KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hz-vscale-up
  Normal   UpdateStatefulSets                                  2m57s  KubeDB Ops-manager Operator  Successfully updated StatefulSets Resources
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-0    2m47s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  evict pod; ConditionStatus:True; PodName:hz-prod-0  2m47s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                  2m37s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-1    117s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  evict pod; ConditionStatus:True; PodName:hz-prod-1  117s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hz-prod-1
  Normal   RestartPods                                         67s    KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                            67s    KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hz-prod
  Normal   Successful                                          67s    KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hz-vscale-up

```

Now, we are going to verify from the Pod, and the PetSet that the resources of the database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo hz-prod-0 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the Hazelcast database.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcastopsrequest -n demo hz-vscale-up
kubectl delete hazelcast -n demo hz-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).
- Monitor your Hazelcast database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
