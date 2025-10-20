---
title: ClickHouse Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: ch-auto-scaling-storage-autoscale
    name: Cluster
    parent: ch-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a ClickHouse Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a ClickHouse cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseAutoscaler](/docs/guides/clickhouse/concepts/clickhouseautoscaler.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)
    - [Storage Autoscaling Overview](/docs/guides/clickhouse/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Storage Autoscaling of Cluster Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  6h2m
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   9m41s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   9m24s
```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `ClickHouse` cluster using a supported version by `KubeDB` operator. Then we are going to apply `ClickHouseAutoscaler` to set up autoscaling.

#### Deploy ClickHouse Cluster

In this section, we are going to deploy a ClickHouse cluster with version `3.13.2`.  Then, in the next section we will set up autoscaling for this database using `ClickHouseAutoscaler` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

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
                  memory: 2Gi
                requests:
                  memory: 1Gi
                  cpu: 900m
          initContainers:
            - name: clickhouse-init
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 1Gi
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/examples/clickhouse/autoscaling/storage/clickhouse-autoscale.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
➤ kubectl get clickhouse -n demo clickhouse-prod
NAME              TYPE                  VERSION   STATUS   AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready    3m17s
```

Let's check volume size from petset, and from the persistent volume,

```bash
➤ kubectl get petset -n demo clickhouse-prod-appscode-cluster-shard-0 -o json | jq '.spec.volumeClaimTemplates[0].spec.resources.requests.storage'
"1Gi"


➤ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                  STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-3a1cbeb5-ed56-4c20-9e29-960c202dd2f7   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-1                     longhorn       <unset>                          3m52s
pvc-53e703b6-b5d5-46c2-97b6-f724f684a448   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-0   longhorn       <unset>                          4m3s
pvc-817d3ee6-922c-471e-889a-4529f8822ce2   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-0                     longhorn       <unset>                          4m10s
pvc-892c1830-8521-4e84-80e8-fdd7995b0c03   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-1   longhorn       <unset>                          3m41s
pvc-a8027da7-92a7-4d70-9a12-6ea5e3280d16   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-1   longhorn       <unset>                          3m45s
pvc-e0722709-9674-4e41-b1be-756a6d77e330   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-2                     longhorn       <unset>                          3m33s
pvc-f2666ae8-1f53-4d0b-869e-ac62ffde5d5e   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-0   longhorn       <unset>                          4m             longhorn       <unset>                          21m
```

You can see the petset has 1Gi storage, and the capacity of all the persistent volume is also 600Mi.

We are now ready to apply the `ClickHouseAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a ClickHouseAutoscaler Object.

#### Create ClickHouseAutoscaler Object

In order to set up vertical autoscaling for this replicaset database, we have to create a `ClickHouseAutoscaler` CRO with our desired configuration. Below is the YAML of the `ClickHouseAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ClickHouseAutoscaler
metadata:
  name: ch-storage-autoscale
  namespace: demo
spec:
  databaseRef:
    name: clickhouse-prod
  storage:
    clickhouse:
      trigger: "On"
      usageThreshold: 40
      scalingThreshold: 50
      expansionMode: "Online"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `clickhouse-autoscale` database.
- `spec.storage.clickhouse.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.clickhouse.usageThreshold` specifies storage usage threshold, if storage usage exceeds `20%` then storage autoscaling will be triggered.
- `spec.storage.clickhouse.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `100%` of the current amount.
- `spec.storage.clickhouse.expansionMode` specifies the expansion mode of volume expansion `clickhouseOpsRequest` created by `clickhouseAutoscaler`. topolvm-provisioner supports online volume expansion so here `expansionMode` is set as "Online".

Let's create the `clickhouseAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/autoscaling/storage/clickhouse-autoscaler-ops.yaml
clickhouseautoscaler.autoscaling.kubedb.com/ch-storage-autoscale created
```

#### Storage Autoscaling is set up successfully

Let's check that the `clickhouseautoscaler` resource is created successfully,

```bash
➤ kubectl get clickhouseautoscaler -n demo 
NAME                   AGE
ch-storage-autoscale   8m34s


➤ kubectl describe clickhouseautoscaler -n demo ch-storage-autoscale 
Name:         ch-storage-autoscale
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         ClickHouseAutoscaler
Metadata:
  Creation Timestamp:  2025-10-07T06:02:08Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ClickHouse
    Name:                  clickhouse-prod
    UID:                   06f8478f-fb06-46cb-b1ec-29d4bed7959c
  Resource Version:        703430
  UID:                     8e3aabfc-b3fc-454e-80c5-f433ee88d89f
Spec:
  Database Ref:
    Name:  clickhouse-prod
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Clickhouse:
      Expansion Mode:  Online
      Scaling Rules:
        Applies Upto:     
        Threshold:        50pc
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    40
Events:                   <none>

```

So, the `clickhouseautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume(`var/lib/clickhouse`) using the following commands:

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ df -h /var/lib/clickhouse
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-53e703b6-b5d5-46c2-97b6-f724f684a448  974M   30M  929M   4% /var/lib/clickhouse
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ dd if=/dev/zero of=/var/lib/clickhouse/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 11.2442 s, 46.6 MB/s
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ df -h /var/lib/clickhouse
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-53e703b6-b5d5-46c2-97b6-f724f684a448  974M  531M  428M  56% /var/lib/clickhouse
```

So, from the above output we can see that the storage usage is 56%, which exceeded the `usageThreshold` 40%.

Let's watch the `clickhouseopsrequest` in the demo namespace to see if any `clickhouseopsrequest` object is created. After some time you'll see that a `clickhouseopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
➤ kubectl get clickhouseopsrequest -n demo
NAME                           TYPE              STATUS        AGE
chops-clickhouse-prod-u410po   VolumeExpansion   Progressing   24s
```

Let's wait for the ops request to become successful.

```bash
➤ kubectl get clickhouseopsrequest -n demo
NAME                           TYPE              STATUS       AGE
chops-clickhouse-prod-u410po   VolumeExpansion   Successful   16m
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-clickhouse-prod-u410po 
Name:         chops-clickhouse-prod-u410po
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=clickhouse-prod
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=clickhouses.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-10-07T06:15:06Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ClickHouseAutoscaler
    Name:                  ch-storage-autoscale
    UID:                   8e3aabfc-b3fc-454e-80c5-f433ee88d89f
  Resource Version:        707683
  UID:                     69ec7501-3dab-44d9-8e9e-a4c083c8ecca
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  clickhouse-prod
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Online
    Node:  1531054080
Status:
  Conditions:
    Last Transition Time:  2025-10-07T06:15:06Z
    Message:               ClickHouse ops-request has started to expand volume of ClickHouse nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-10-07T06:15:34Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2025-10-07T06:15:14Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2025-10-07T06:15:14Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2025-10-07T06:24:39Z
    Message:               successfully updated PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2025-10-07T06:15:39Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-10-07T06:15:39Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-10-07T06:24:34Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-10-07T06:25:02Z
    Message:               successfully reconciled the ClickHouse resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-10-07T06:24:44Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-10-07T06:25:07Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2025-10-07T06:25:07Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2025-10-07T06:25:07Z
    Message:               Successfully completed volumeExpansion for ClickHouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                  Age    From                         Message
  ----     ------                                  ----   ----                         -------
  Normal   Starting                                17m    KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-clickhouse-prod-u410po
  Normal   Starting                                17m    KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                              17m    KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-clickhouse-prod-u410po
  Warning  get petset; ConditionStatus:True        17m    KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Normal   OrphanPetSetPods                        17m    KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True           17m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False  17m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True           15m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   15m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           15m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True         15m    KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           15m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False  15m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True           14m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           12m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   12m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           12m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True         12m    KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           12m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False  12m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True           12m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   10m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           10m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True         10m    KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           10m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False  10m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True           10m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   8m15s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateNodePVCs                          8m10s  KubeDB Ops-manager Operator  successfully updated PVC sizes
  Warning  reconcile; ConditionStatus:True         8m5s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True         7m58s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True         7m47s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                           7m47s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse resources
  Warning  get pet set; ConditionStatus:True       7m42s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True       7m42s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                            7m42s  KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                7m42s  KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                              7m42s  KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-clickhouse-prod-u410po
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the replicaset database has expanded to meet the desired state, Let's check,

```bash
➤ kubectl get petset -n demo clickhouse-prod-appscode-cluster-shard-0 -o json | jq '.spec.volumeClaimTemplates[0].spec.resources.requests.storage'
"1531054080"
shuvo@shuvo-pc:~
➤ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                  STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-3a1cbeb5-ed56-4c20-9e29-960c202dd2f7   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-1                     longhorn       <unset>                          38m
pvc-53e703b6-b5d5-46c2-97b6-f724f684a448   1462Mi     RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-0   longhorn       <unset>                          38m
pvc-817d3ee6-922c-471e-889a-4529f8822ce2   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-0                     longhorn       <unset>                          38m
pvc-892c1830-8521-4e84-80e8-fdd7995b0c03   1462Mi     RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-1   longhorn       <unset>                          38m
pvc-a8027da7-92a7-4d70-9a12-6ea5e3280d16   1462Mi     RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-0-1   longhorn       <unset>                          38m
pvc-e0722709-9674-4e41-b1be-756a6d77e330   1Gi        RWO            Delete           Bound    demo/data-clickhouse-prod-keeper-2                     longhorn       <unset>                          38m
pvc-f2666ae8-1f53-4d0b-869e-ac62ffde5d5e   1462Mi     RWO            Delete           Bound    demo/data-clickhouse-prod-appscode-cluster-shard-1-0   longhorn       <unset>                          38m
```

The above output verifies that we have successfully autoscaled the volume of the Clickhouse replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete clickhouse -n demo clickhouse-prod
kubectl delete clickhouseautoscaler -n demo ch-storage-autoscale
kubectl delete ns demo
```
