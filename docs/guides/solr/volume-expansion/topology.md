---
title: Solr Topology Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: sl-volume-expansion-topology
    name: Topology
    parent: sl-volume-expansion
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Topology Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Solr Topology Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [Topology](/docs/guides/solr/clustering/topology_cluster.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Volume Expansion Overview](/docs/guides/solr/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/Solr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Solr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Topology Solr Cluster

Here, we are going to deploy a `Solr` topology using a supported version by `KubeDB` operator. Then we are going to apply `SolrOpsRequest` to expand its volume.

### Prepare Solr Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get sc
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  24d
```

We can see from the output the `local-path` storage class has `ALLOWVOLUMEEXPANSION` field as false. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Solr` combined cluster with version `9.4.1`.

### Deploy Solr

In this section, we are going to deploy a Solr topology cluster for broker and controller with 1GB volume. Then, in the next section we will expand its volume to 2GB using `SolrOpsRequest` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.4.1
  zookeeperRef:
    name: zoo
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/volume-expansion/topology.yaml
solr.kubedb.com/solr-cluster created
```

Now, wait until `solr-cluster` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.4.1     Ready    41m

```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo solr-cluster-overseer -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get petset -n demo solr-cluster-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get petset -n demo solr-cluster-coordinator -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-31538e3e-2d02-4ca0-9b76-5da7c63cea70   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-data-0          longhorn       <unset>                          44m
pvc-8c5b14ab-3da4-4492-abf4-edd7faa265ef   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-overseer-0      longhorn       <unset>                          44m
pvc-95522f35-52bd-4978-b66f-1979cec34982   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-coordinator-0   longhorn       <unset>                          44m
```

You can see the petsets have 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `SolrOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the Solr topology cluster.

#### Create SolrOpsRequest

In order to expand the volume of the database, we have to create a `SolrOpsRequest` CR with our desired volume size. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: sl-volume-exp-topology
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    data: 11Gi
    overseer : 11Gi

```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `Solr-prod`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.data` specifies the desired volume size for data node.
- `spec.volumeExpansion.overseer` specifies the desired volume size for overseer node.
- `spec.volumeExpansion.coordinator` specifies the desired volume size for coordinator node.

> If you want to expand the volume of only one node, you can specify the desired volume size for that node only.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/volume-expansion/solr-volume-expansion-topology.yaml
solropsrequest.ops.kubedb.com/sl-volume-exp-topology created
```

#### Verify Solr Topology volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Solr` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CR,

```bash
$ kubectl get solropsrequest -n demo
NAME                     TYPE              STATUS       AGE
sl-volume-exp-topology   VolumeExpansion   Successful   3m1s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to expand the volume of Solr.

```bash
$ kubectl describe slops -n demo sl-volume-exp-topology 
Name:         sl-volume-exp-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-12T06:38:29Z
  Generation:          1
  Resource Version:    2444852
  UID:                 2ea88297-45d1-4f48-b21a-8ede43d3ee69
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Data:      11Gi
    Mode:      Offline
    Overseer:  11Gi
Status:
  Conditions:
    Last Transition Time:  2024-11-12T06:38:29Z
    Message:               Solr ops-request has started to expand volume of solr nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-11-12T06:39:03Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-11-12T06:38:43Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2024-11-12T06:38:43Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2024-11-12T06:40:13Z
    Message:               successfully updated data node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionDataNode
    Status:                True
    Type:                  VolumeExpansionDataNode
    Last Transition Time:  2024-11-12T06:40:53Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-11-12T06:39:08Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2024-11-12T06:39:08Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-11-12T06:39:43Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-11-12T06:39:43Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2024-11-12T06:41:13Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-11-12T06:40:03Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-11-12T06:41:38Z
    Message:               successfully updated overseer node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionOverseerNode
    Status:                True
    Type:                  VolumeExpansionOverseerNode
    Last Transition Time:  2024-11-12T06:41:18Z
    Message:               running solr; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningSolr
    Last Transition Time:  2024-11-12T06:41:44Z
    Message:               successfully reconciled the Solr resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-12T06:41:49Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-11-12T06:41:49Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-11-12T06:41:49Z
    Message:               Successfully completed volumeExpansion for Solr
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                   Age    From                         Message
  ----     ------                                   ----   ----                         -------
  Normal   Starting                                 3m39s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/sl-volume-exp-topology
  Normal   Starting                                 3m39s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                               3m39s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: sl-volume-exp-topology
  Warning  get petset; ConditionStatus:True         3m25s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True      3m25s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True         3m20s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True         3m15s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True      3m15s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True         3m10s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Normal   OrphanPetSetPods                         3m5s   KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True            3m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  3m     KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         3m     KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False           2m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            2m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m25s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          2m25s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   2m25s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            2m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m20s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m15s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m10s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m5s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    2m5s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         2m5s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  2m5s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   VolumeExpansionDataNode                  115s   KubeDB Ops-manager Operator  successfully updated data node PVC sizes
  Warning  get pod; ConditionStatus:True            110s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  110s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         110s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False           105s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            75s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            75s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          75s    KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   75s    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            70s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            70s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            65s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            65s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            60s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            60s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            55s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            55s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    55s    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         55s    KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  55s    KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            50s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  running solr; ConditionStatus:False      50s    KubeDB Ops-manager Operator  running solr; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            45s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   VolumeExpansionOverseerNode              30s    KubeDB Ops-manager Operator  successfully updated overseer node PVC sizes
  Normal   UpdatePetSets                            24s    KubeDB Ops-manager Operator  successfully reconciled the Solr resources
  Warning  get pet set; ConditionStatus:True        19s    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        19s    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             19s    KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 19s    KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                               19s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: sl-volume-exp-topology
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo solr-cluster-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"11Gi"
$ kubectl get petset -n demo solr-cluster-overseer -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"11Gi"
$ kubectl get petset -n demo solr-cluster-coordinator -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-31538e3e-2d02-4ca0-9b76-5da7c63cea70   11Gi       RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-data-0          longhorn       <unset>                          52m
pvc-8c5b14ab-3da4-4492-abf4-edd7faa265ef   11Gi       RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-overseer-0      longhorn       <unset>                          52m
pvc-95522f35-52bd-4978-b66f-1979cec34982   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-coordinator-0   longhorn       <unset>                          52m
```

The above output verifies that we have successfully expanded the volume of the Solr.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete solropsrequest -n demo ksl-volume-exp-topology
kubectl delete sl -n demo solr-cluster
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
