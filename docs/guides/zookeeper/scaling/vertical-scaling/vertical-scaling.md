---
title: Vertical Scaling ZooKeeper
menu:
  docs_{{ .version }}:
    identifier: zk-vertical-scaling-ops
    name: Scale Vertically
    parent: zk-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale ZooKeeper Standalone

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a ZooKeeper standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
    - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/zookeeper/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/zookeeper](/docs/examples/zookeeper) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Standalone

Here, we are going to deploy a  `ZooKeeper` standalone using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare ZooKeeper Standalone Database

Now, we are going to deploy a `ZooKeeper` standalone database with version `3.8.3`.

### Deploy ZooKeeper standalone

In this section, we are going to deploy a ZooKeeper standalone database. Then, in the next section we will update the resources of the database using `ZooKeeperOpsRequest` CRD. Below is the YAML of the `ZooKeeper` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-quickstart
  namespace: demo
spec:
  version: "3.8.3"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"

```

Let's create the `ZooKeeper` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/scaling/zookeeper.yaml
zookeeper.kubedb.com/zk-quickstart created
```

Now, wait until `zk-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get zk -n demo
NAME            VERSION    STATUS    AGE
zk-quickstart   3.8.3      Ready     5m56s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo zk-quickstart-0 -o json | jq '.spec.containers[].resources'
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

You can see the Pod has default resources which is assigned by the Kubedb operator.

We are now ready to apply the `ZooKeeperOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the standalone database to meet the desired resources after scaling.

#### Create ZooKeeperOpsRequest

In order to update the resources of the database, we have to create a `ZooKeeperOpsRequest` CR with our desired resources. Below is the YAML of the `ZooKeeperOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: vscale
  namespace: demo
spec:
  databaseRef:
    name: zk-quickstart
  type: VerticalScaling
  verticalScaling:
    node:
      resources:
        limits:
          cpu: 1
          memory: 2Gi
        requests:
          cpu: 1
          memory: 2Gi
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `vscale` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.node` specifies the desired resources after scaling.
- Have a look [here](/docs/guides/zookeeper/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/scaling/vertical-scaling/zk-vscale.yaml
zookeeperopsrequest.ops.kubedb.com/vscale created
```

#### Verify ZooKeeper Standalone resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `ZooKeeper` object and related `Petsets` and `Pods`.

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CR,

```bash
$ kubectl get zookeeperopsrequest -n demo
Every 2.0s: kubectl get zookeeperopsrequest -n demo
NAME        TYPE              STATUS       AGE
vscale      VerticalScaling   Successful   108s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe zookeeperopsrequest -n demo vscale
Name:         vscale
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-10-24T11:21:28Z
  Generation:          1
  Resource Version:    1151711
  UID:                 53ba9aef-cfa6-40f1-a5a8-6055bafb0c7b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   zk-quickstart
  Timeout:  5m
  Type:     VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2Gi
        Requests:
          Cpu:     1
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-24T11:21:28Z
    Message:               ZooKeeper ops-request has started to vertically scaling the ZooKeeper nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-24T11:21:31Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-24T11:21:31Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                False
    Type:                  RestartPods
    Last Transition Time:  2024-10-24T11:21:36Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-10-24T11:21:36Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-10-24T11:21:41Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-10-24T11:22:16Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-10-24T11:22:16Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-10-24T11:22:56Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-10-24T11:22:56Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
  Observed Generation:     1
  Phase:                   Progressing
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  3m24s  KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/vscale
  Normal   Starting                                                  3m24s  KubeDB Ops-manager Operator  Pausing ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                3m24s  KubeDB Ops-manager Operator  Successfully paused ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: vscale
  Normal   UpdatePetSets                                             3m21s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-0    3m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-0  3m16s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  running pod; ConditionStatus:False                        3m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-1    2m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-1  2m36s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-2    116s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-2  116s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-2

```

Now, we are going to verify from the Pod yaml whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo zk-quickstart-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the ZooKeeper standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete zk -n demo zk-quickstart
kubectl delete zookeeperopsrequest -n demo vscale
```