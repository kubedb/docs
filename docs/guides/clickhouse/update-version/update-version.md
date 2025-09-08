---
title: Update Version of ClickHouse
menu:
  docs_{{ .version }}:
    identifier: ch-update-version-clickhouse
    name: ClickHouse Update Version
    parent: ch-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of ClickHouse

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `ClickHouse` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)
    - [Updating Overview](/docs/guides/clickhouse/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/clickhouse](/docs/examples/clickhouse) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare ClickHouse

Now, we are going to deploy a `ClickHouse` replicaset database with version `24.4.1`.

### Deploy ClickHouse

In this section, we are going to deploy a ClickHouse cluster. Then, in the next section we will update the version using `ClickHouseOpsRequest` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
        name: appscode-cluster
        shards: 2
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: clickhouse
                resources:
                  limits:
                    memory: 4Gi
                  requests:
                    cpu: 500m
                    memory: 2Gi
            initContainers:
              - name: clickhouse-init
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
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/update-version/clickhouse-cluster.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` created has status `Ready`. i.e,

```bash
➤ kubectl get clickhouse -n demo -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   16s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   51s
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m5s

```

We are now ready to apply the `ClickHouseOpsRequest` CR to update.

### update ClickHouse Version

Here, we are going to update `ClickHouse` from `24.4.1` to `25.7.1`.

#### Create ClickHouseOpsRequest:

In order to update the version, we have to create a `ClickHouseOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `ClickHouseOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: clickhouse-prod
  updateVersion:
    targetVersion: 25.7.1
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `clickhouse-prod` ClickHouse.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `25.7.1`.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/update-version/update-version.yaml
clickhouseopsrequest.ops.kubedb.com/ch-update-version created
```

#### Verify ClickHouse version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `ClickHouse` object and related `PetSets` and `Pods`.

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CR,

```bash
➤ kubectl get clickhouseopsrequest -n demo
NAME                TYPE            STATUS       AGE
ch-update-version   UpdateVersion   Successful   16m
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
➤ kubectl describe clickhouseopsrequest -n demo ch-update-version 
Name:         ch-update-version
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-26T08:24:28Z
  Generation:          1
  Resource Version:    936717
  UID:                 563aaded-ceca-490a-a9c0-44dd38853b87
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  5m
  Type:     UpdateVersion
  Update Version:
    Target Version:  25.7.1
Status:
  Conditions:
    Last Transition Time:  2025-08-26T08:24:28Z
    Message:               ClickHouse ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2025-08-26T08:24:36Z
    Message:               successfully reconciled the ClickHouse with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-26T08:24:36Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-26T08:26:26Z
    Message:               Successfully Restarted ClickHouse nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-26T08:24:41Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-26T08:24:41Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-26T08:24:46Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-26T08:25:11Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-26T08:25:11Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-26T08:25:46Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-26T08:25:46Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-26T08:26:06Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-26T08:26:06Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-26T08:26:26Z
    Message:               Successfully completed update clickhouse version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age   From                         Message
  ----     ------                                                                               ----  ----                         -------
  Normal   Starting                                                                             16m   KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/ch-update-version
  Normal   Starting                                                                             16m   KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           16m   KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: ch-update-version
  Warning  reconcile; ConditionStatus:True                                                      16m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      16m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      16m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        16m   KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with updated version
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    16m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   16m   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    16m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    15m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    15m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartPods                                                                          14m   KubeDB Ops-manager Operator  Successfully Restarted ClickHouse nodes
  Normal   Starting                                                                             14m   KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           14m   KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: ch-update-version
```

Now, we are going to verify whether the `ClickHouse` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
➤ kubectl get clickhouse -n demo clickhouse-prod -o=jsonpath='{.spec.version}{"\n"}'
25.7.1

➤ kubectl get petset -n demo clickhouse-prod-appscode-cluster-shard-0 -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
clickhouse/clickhouse-server:25.7.1@sha256:c2d7cdb388ab9a8c9e9ec31673fc9203e373f06ab33dfa4b3cdb0095b5d2a2d2
```

You can see from above, our `ClickHouse` has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete chops -n demo ch-update-version
kubectl delete ch -n demo clickhouse-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
