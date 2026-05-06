---
title: ClickHouse Cluster Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: ch-volume-expansion-cluster
    name: Cluster
    parent: ch-volume-expansion
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouse Cluster Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a ClickHouse Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)
    - [Volume Expansion Overview](/docs/guides/clickhouse/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/clickhouse](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/clickhouse) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of ClickHouse Cluster

Here, we are going to deploy a `ClickHouse` using a supported version by `KubeDB` operator. Then we are going to apply `ClickHouseOpsRequest` to expand its volume.

### Prepare ClickHouse Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
➤ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  7d22h
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   6d23h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   6d23h
```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `ClickHouse` combined cluster with version `24.4.1`.

### Deploy ClickHouse

In this section, we are going to deploy a ClickHouse cluster with 1GB volume. Then, in the next section we will expand its volume to 2GB using `ClickHouseOpsRequest` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

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
                    memory: 1Gi
            initContainers:
              - name: clickhouse-init
                resources:
                  limits:
                    memory: 1Gi
                  requests:
                    cpu: 500m
                    memory: 1Gi
        storage:
          storageClassName: longhorn
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/volume-expansion/clickhouse-cluster.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
➤ kubectl get clickhouse -n demo clickhouse-prod -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   64s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   88s
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m9s
```

Let's check volume size from petset, and from the persistent volume,

```bash
➤ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                  STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-03fe7758-fc16-4c77-8072-4948331386a4   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-1                     longhorn       <unset>                          4m14s
pvc-348df1d0-037c-4284-9556-1bd2e2089a37   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-0   longhorn       <unset>                          4m22s
pvc-39d65c1d-86dc-4028-ad68-91a45a2d4be5   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-0   longhorn       <unset>                          4m19s
pvc-6bd28008-6fc4-4b24-b377-257268ae60b2   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-2                     longhorn       <unset>                          4m
pvc-73abd357-2d97-4469-a34a-17ce41364fe1   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-0                     longhorn       <unset>                          4m28s
pvc-7ab3ae88-bdd1-45c2-8aa4-dc58e50e0551   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-1   longhorn       <unset>                          4m8s
pvc-e4004397-b6f7-4a6b-bef6-a87e0f8c2014   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-1   longhorn       <unset>                          4m3s
```

You can see the petsets have 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `ClickHouseOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the clickhouse cluster.

#### Create ClickHouseOpsRequest

In order to expand the volume of the database, we have to create a `ClickHouseOpsRequest` CR with our desired volume size. Below is the YAML of the `ClickHouseOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-offline-volume-expansion
  namespace: demo
spec:
  apply: "IfReady"
  type: VolumeExpansion
  databaseRef:
    name: clickhouse-prod
  volumeExpansion:
    mode: "Offline"
    node: 2Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `clickhouse-prod`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.node` specifies the desired volume size for cluster node.

> If you want to expand the volume of only one node, you can specify the desired volume size for that node only.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/volume-expansion/chops-volume-expansion.yaml
clickhouseopsrequest.ops.kubedb.com/ch-offline-volume-expansion created
```

#### Verify ClickHouse Cluster volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `ClickHouse` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CR,

```bash
➤ kubectl get clickhouseopsrequest -n demo
NAME                          TYPE              STATUS       AGE
ch-offline-volume-expansion   VolumeExpansion   Successful   6m17s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to expand the volume of clickhouse.

```bash
➤ kubectl describe clickhouseopsrequest -n demo ch-offline-volume-expansion 
Name:         ch-offline-volume-expansion
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-26T09:49:40Z
  Generation:          1
  Resource Version:    949957
  UID:                 29f060ad-1d94-4019-81ad-eaddad4e1359
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  clickhouse-prod
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Offline
    Node:  2Gi
Status:
  Conditions:
    Last Transition Time:  2025-08-26T09:49:40Z
    Message:               ClickHouse ops-request has started to expand volume of ClickHouse nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-08-26T09:50:17Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2025-08-26T09:49:57Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2025-08-26T09:49:57Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2025-08-26T09:54:52Z
    Message:               successfully updated node PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2025-08-26T09:54:02Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2025-08-26T09:50:22Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2025-08-26T09:50:22Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2025-08-26T09:50:32Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-08-26T09:50:32Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-08-26T09:54:22Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-08-26T09:51:12Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2025-08-26T09:51:17Z
    Message:               running click house; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningClickHouse
    Last Transition Time:  2025-08-26T09:54:57Z
    Message:               successfully reconciled the ClickHouse resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-26T09:54:57Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-26T09:55:02Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2025-08-26T09:55:02Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2025-08-26T09:55:02Z
    Message:               Successfully completed volumeExpansion for ClickHouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                      Age    From                         Message
  ----     ------                                      ----   ----                         -------
  Normal   Starting                                    7m16s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/ch-offline-volume-expansion
  Normal   Starting                                    7m16s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                  7m16s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: ch-offline-volume-expansion
  Warning  get petset; ConditionStatus:True            6m59s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True         6m59s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True            6m54s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True            6m49s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True         6m49s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True            6m44s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Normal   OrphanPetSetPods                            6m39s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True               6m34s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     6m34s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True            6m34s  KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False              6m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True               6m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               6m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True             6m24s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False      6m24s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True               6m19s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               6m19s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               6m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               6m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               6m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               6m9s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               6m4s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               6m4s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m59s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               5m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m54s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               5m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m49s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               5m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               5m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True       5m44s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True            5m44s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     5m44s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  running click house; ConditionStatus:False  5m39s  KubeDB Ops-manager Operator  running click house; ConditionStatus:False
  Warning  get pod; ConditionStatus:True               5m34s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m19s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               5m4s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     5m4s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True            5m4s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m59s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               4m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True             4m59s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False      4m59s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True               4m54s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               4m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m49s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               4m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               4m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               4m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True       4m39s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True            4m39s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     4m39s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m34s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m19s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               4m4s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     4m4s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True            4m4s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m59s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               3m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True             3m59s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False      3m59s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True               3m54s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               3m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m49s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               3m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               3m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               3m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True       3m39s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True            3m39s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     3m39s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m34s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m19s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               3m4s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     3m4s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True            3m4s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False              2m59s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True               2m54s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               2m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True             2m54s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False      2m54s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True               2m49s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               2m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               2m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               2m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m34s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True               2m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True       2m34s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True            2m34s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True     2m34s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m19s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True               2m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateNodePVCs                              2m4s   KubeDB Ops-manager Operator  successfully updated node PVC sizes
  Warning  reconcile; ConditionStatus:True             119s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True             119s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True             119s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                               119s   KubeDB Ops-manager Operator  successfully reconciled the ClickHouse resources
  Warning  get pet set; ConditionStatus:True           114s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True           114s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                                114s   KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                    114s   KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                  114s   KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: ch-offline-volume-expansion
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
➤ kubectl get petset -n demo clickhouse-prod-appscode-cluster-shard-0 -o json | jq '.spec.volumeClaimTemplates[0].spec.resources.requests.storage'
"2Gi"


➤ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                  STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-03fe7758-fc16-4c77-8072-4948331386a4   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-1                     longhorn       <unset>                          15m
pvc-348df1d0-037c-4284-9556-1bd2e2089a37   2Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-0   longhorn       <unset>                          16m
pvc-39d65c1d-86dc-4028-ad68-91a45a2d4be5   2Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-0   longhorn       <unset>                          16m
pvc-6bd28008-6fc4-4b24-b377-257268ae60b2   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-2                     longhorn       <unset>                          15m
pvc-73abd357-2d97-4469-a34a-17ce41364fe1   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-0                     longhorn       <unset>                          16m
pvc-7ab3ae88-bdd1-45c2-8aa4-dc58e50e0551   2Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-1   longhorn       <unset>                          15m
pvc-e4004397-b6f7-4a6b-bef6-a87e0f8c2014   2Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-1   longhorn       <unset>                          15m
```

The above output verifies that we have successfully expanded the data related volume of the ClickHouse.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete clickhouseopsrequests -n demo ch-offline-volume-expansion
kubectl delete ch -n demo clickhouse-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
