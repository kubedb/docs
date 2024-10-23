---
title: Updating ZooKeeper Cluster
menu:
  docs_{{ .version }}:
    identifier: zk-cluster-update-version
    name: Update Version
    parent: zk-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of ZooKeeper Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `ZooKeeper` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
    - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)
    - [Updating Overview](/docs/guides/zookeeper/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/zookeeper](/docs/examples/zookeeper) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare ZooKeeper cluster

Now, we are going to deploy a `ZooKeeper` cluster with version `3.8.3`.

### Deploy ZooKeeper

In this section, we are going to deploy a ZooKeeper cluster. Then, in the next section we will update the version of the database using `ZooKeeperOpsRequest` CRD. Below is the YAML of the `ZooKeeper` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/update-version/zookeeper.yaml
zookeeper.kubedb.com/zk-quickstart created
```

Now, wait until `zk-quickstart` created has status `Ready`. i.e,

```bash
$ kubectl get zk -n demo                                                                                                                                             
NAME            VERSION    STATUS    AGE
zk-quickstart      3.12.12   Ready     109s
```

We are now ready to apply the `ZooKeeperOpsRequest` CR to update this database.

### update ZooKeeper Version

Here, we are going to update `ZooKeeper` cluster from `3.8.3` to `3.9.1`.

#### Create ZooKeeperOpsRequest:

In order to update the version of the cluster, we have to create a `ZooKeeperOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `ZooKeeperOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: upgrade-topology
  namespace: demo
spec:
  databaseRef:
    name: zk-quickstart
  type: UpdateVersion
  updateVersion:
    targetVersion: 3.9.1
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `zk-quickstart` ZooKeeper database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `3.9.1`.
- Have a look [here](/docs/guides/zookeeper/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/update-version/zk-version-upgrade-ops.yaml
zookeeperopsrequest.ops.kubedb.com/upgrade-topology created
```

#### Verify ZooKeeper version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `ZooKeeper` object and related `PetSets` and `Pods`.

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CR,

```bash
$ kubectl get zookeeperopsrequest -n demo
Every 2.0s: kubectl get zookeeperopsrequest -n demo
NAME                      TYPE            STATUS       AGE
upgrade-topology    UpdateVersion   Successful   84s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe zookeeperopsrequest -n demo upgrade-topology
Name:         upgrade-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-10-23T10:46:27Z
  Generation:          1
  Resource Version:    1112190
  UID:                 6a1baef3-74cb-4a44-9b8f-f4fa49a4cfca
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   zk-quickstart
  Timeout:  5m
  Type:     UpdateVersion
  Update Version:
    Target Version:  3.9.1
Status:
  Conditions:
    Last Transition Time:  2024-10-23T10:46:27Z
    Message:               Zookeeper ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2024-10-23T10:46:35Z
    Message:               successfully reconciled the ZooKeeper with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-23T10:49:25Z
    Message:               Successfully Restarted ZooKeeper nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-23T10:46:40Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-10-23T10:46:40Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-10-23T10:46:45Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-10-23T10:47:25Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-10-23T10:47:25Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-10-23T10:48:05Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-10-23T10:48:05Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
    Last Transition Time:  2024-10-23T10:48:45Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-3
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-3
    Last Transition Time:  2024-10-23T10:48:45Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-3
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-3
    Last Transition Time:  2024-10-23T10:49:25Z
    Message:               Successfully updated ZooKeeper version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  10m    KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/upgrade-topology
  Normal   Starting                                                  10m    KubeDB Ops-manager Operator  Pausing ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                10m    KubeDB Ops-manager Operator  Successfully paused ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: upgrade-topology
  Normal   UpdatePetSets                                             10m    KubeDB Ops-manager Operator  successfully reconciled the ZooKeeper with updated version
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-0    10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-0  10m    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  running pod; ConditionStatus:False                        10m    KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-1    9m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-1  9m25s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-2    8m45s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-2  8m45s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-2
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-3    8m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-3
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-3  8m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-3
  Normal   RestartPods                                               7m25s  KubeDB Ops-manager Operator  Successfully Restarted ZooKeeper nodes
  Normal   Starting                                                  7m25s  KubeDB Ops-manager Operator  Resuming ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                7m25s  KubeDB Ops-manager Operator 
```

Now, we are going to verify whether the `ZooKeeper` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get zk -n demo zk-quickstart -o=jsonpath='{.spec.version}{"\n"}'
3.9.1

$ kubectl get petset -n demo zk-quickstart -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/zookeeper:3.9.1@sha256:21365fd1bd55cacd6bf556394d6dcb76ad559ad3767adc304e62db205e4b10b7

$ kubectl get pods -n demo zk-quickstart-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/zookeeper:3.9.1
```

You can see from above, our `ZooKeeper` cluster has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete zk -n demo zk-quickstart
kubectl delete zookeeperopsrequest -n demo upgrade-topology
```