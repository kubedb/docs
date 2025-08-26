---
title: Vertical Scaling ClickHouse Standalone
menu:
  docs_{{ .version }}:
    identifier: ch-vertical-scaling-standalone
    name: Standalone
    parent: ch-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale ClickHouse Standalone

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a ClickHouse standalone.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/clickhouse/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/clickhouse](/docs/examples/clickhouse) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on ClickHouse Standalone

Here, we are going to deploy a `ClickHouse` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare ClickHouse Standalone

Now, we are going to deploy a `ClickHouse` cluster database with version `24.4.1`.

### Deploy ClickHouse Standalone

In this section, we are going to deploy a ClickHouse standalone. Then, in the next section we will update the resources of the database using `ClickHouseOpsRequest` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/scaling/clickhouse-standalone.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
➤ kubectl get clickhouse -n demo -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   4s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   50s
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m5s
```

Let's check the Pod containers resources of the ClickHouse cluster. Run the following command to get the resources of the containers of the ClickHouse cluster

```bash
➤ kubectl get pod -n demo clickhouse-prod-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "4Gi"
  }
}
```

We are now ready to apply the `ClickHouseOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the clickhouse standalone to meet the desired resources after scaling.

#### Create ClickHouseOpsRequest

In order to update the resources of the database, we have to create a `ClickHouseOpsRequest` CR with our desired resources. Below is the YAML of the `ClickHouseOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-vertical-scale-standalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: clickhouse-prod
  verticalScaling:
    standalone:
      resources:
        requests:
          memory: "3Gi"
          cpu: "3"
        limits:
          memory: "3Gi"
          cpu: "3"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on clickhouse.
- `spec.verticalScaling.standalone` specifies the desired resources after scaling.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/scaling/vertical-scaling/ch-vertical-ops-standalone.yaml
clickhouseopsrequest.ops.kubedb.com/ch-vertical-scale-standalone created
```

#### Verify ClickHouse standalone resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `ClickHouse` object and related `PetSets` and `Pods`.

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CR,

```bash
➤ kubectl get clickhouseopsrequest -n demo ch-vertical-scale-standalone 
NAME                           TYPE              STATUS       AGE
ch-vertical-scale-standalone   VerticalScaling   Successful   68s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to scale the clickhouse standalone.

```bash
➤ kubectl describe clickhouseopsrequest -n demo ch-vertical-scale-standalone 
Name:         ch-vertical-scale-standalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-26T06:50:05Z
  Generation:          1
  Resource Version:    924257
  UID:                 1556f9bd-3728-4c5c-97cb-12396ab71cfb
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  clickhouse-prod
  Type:    VerticalScaling
  Vertical Scaling:
    Standalone:
      Resources:
        Limits:
          Cpu:     3
          Memory:  3Gi
        Requests:
          Cpu:     3
          Memory:  3Gi
Status:
  Conditions:
    Last Transition Time:  2025-08-26T06:50:05Z
    Message:               ClickHouse ops-request has started to vertically scaling the ClickHouse nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2025-08-26T06:50:08Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-26T06:50:33Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-26T06:50:13Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-0
    Last Transition Time:  2025-08-26T06:50:13Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-0
    Last Transition Time:  2025-08-26T06:50:18Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-26T06:50:33Z
    Message:               Successfully completed the vertical scaling for ClickHouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age   From                         Message
  ----     ------                                                      ----  ----                         -------
  Normal   Starting                                                    2m    KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/ch-vertical-scale-standalone
  Normal   Starting                                                    2m    KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                  2m    KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: ch-vertical-scale-standalone
  Normal   UpdatePetSets                                               117s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-0    112s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-0  112s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-0
  Warning  running pod; ConditionStatus:False                          107s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Normal   RestartPods                                                 92s   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                    92s   KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                  92s   KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: ch-vertical-scale-standalone
```
Now, we are going to verify from one of the Pod yaml whether the resources of the clickhouse standalone has updated to meet up the desired state, Let's check,

```bash
➤ kubectl get pod -n demo clickhouse-prod-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "3",
    "memory": "3Gi"
  },
  "requests": {
    "cpu": "3",
    "memory": "3Gi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the ClickHouse cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete clickhouse -n demo clickhouse-prod
kubectl delete clickhouseopsrequest -n demo ch-vertical-scale-standalone
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
