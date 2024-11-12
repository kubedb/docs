---
title: Solr Combined Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: sl-volume-expansion-combined
    name: Combined
    parent: sl-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Combined Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Solr Combined Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [Combined](/docs/guides/solr/clustering/combined_cluster.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Volume Expansion Overview](/docs/guides/solr/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/Solr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Solr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Combined Solr Cluster

Here, we are going to deploy a `Solr` combined using a supported version by `KubeDB` operator. Then we are going to apply `SolrOpsRequest` to expand its volume.

### Prepare Solr Combined CLuster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get sc
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  24d
```

We can see from the output the `local-path` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Solr` combined cluster with version `9.4.1`.

### Deploy Solr

In this section, we are going to deploy a Solr combined cluster with 1GB volume. Then, in the next section we will expand its volume to 2GB using `SolrOpsRequest` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-combined
  namespace: demo
spec:
  version: 9.4.1
  replicas: 2
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/volume-expansion/combined.yaml
Solr.kubedb.com/Solr-dev created
```

Now, wait until `Solr-dev` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME            TYPE                  VERSION   STATUS   AGE
solr-combined   kubedb.com/v1alpha2   9.4.1     Ready    23m
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo solr-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-02cddba5-1d6a-4f1b-91b2-a7e55857b6b7   1Gi        RWO            Delete           Bound    demo/solr-combined-data-solr-combined-1   longhorn       <unset>                          23m
pvc-61b8f97a-a588-4125-99f3-604f6a70d560   1Gi        RWO            Delete           Bound    demo/solr-combined-data-solr-combined-0   longhorn       <unset>                          24m
```

You can see the petset has 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `SolrOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the Solr combined cluster.

#### Create SolrOpsRequest

In order to expand the volume of the database, we have to create a `SolrOpsRequest` CR with our desired volume size. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: sl-volume-exp-combined
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    node: 11Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `Solr-dev`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.node` specifies the desired volume size.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/volume-expansion/solr-volume-expansion-combined.yaml
solropsrequest.ops.kubedb.com/sl-volume-exp-combined created
```

#### Verify Solr Combined volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Solr` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CR,

```bash
$ kubectl get slops -n demo
NAME                     TYPE              STATUS       AGE
sl-volume-exp-topology   VolumeExpansion   Successful   3m
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe slops -n demo sl-volume-exp-topology 
Name:         sl-volume-exp-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-12T07:59:08Z
  Generation:          1
  Resource Version:    2453072
  UID:                 efa404a1-0cdf-46a9-9995-3f3fca88fa4a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-combined
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Offline
    Node:  11Gi
Status:
  Conditions:
    Last Transition Time:  2024-11-12T07:59:08Z
    Message:               Solr ops-request has started to expand volume of solr nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-11-12T07:59:26Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-11-12T07:59:16Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2024-11-12T07:59:16Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2024-11-12T08:01:31Z
    Message:               successfully updated combined node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionCombinedNode
    Status:                True
    Type:                  VolumeExpansionCombinedNode
    Last Transition Time:  2024-11-12T08:01:06Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-11-12T07:59:31Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2024-11-12T07:59:31Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-11-12T08:00:06Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-11-12T08:00:06Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2024-11-12T08:01:21Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-11-12T08:00:21Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-11-12T08:01:36Z
    Message:               successfully reconciled the Solr resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-12T08:01:41Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-11-12T08:01:41Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-11-12T08:01:41Z
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
  Normal   Starting                                 3m29s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/sl-volume-exp-topology
  Normal   Starting                                 3m29s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-combined
  Normal   Successful                               3m29s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-combined for SolrOpsRequest: sl-volume-exp-topology
  Warning  get petset; ConditionStatus:True         3m21s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True      3m21s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True         3m16s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Normal   OrphanPetSetPods                         3m11s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True            3m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  3m6s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         3m6s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False           3m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            2m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          2m31s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   2m31s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            2m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    2m16s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         2m16s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  2m16s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  2m6s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         2m6s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False           2m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            91s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            91s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          91s    KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   91s    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            86s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            86s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            81s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            81s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            76s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            76s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    76s    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         76s    KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  76s    KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            71s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   VolumeExpansionCombinedNode              66s    KubeDB Ops-manager Operator  successfully updated combined node PVC sizes
  Normal   UpdatePetSets                            61s    KubeDB Ops-manager Operator  successfully reconciled the Solr resources
  Warning  get pet set; ConditionStatus:True        56s    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             56s    KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 56s    KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-combined
  Normal   Successful                               56s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-combined for SolrOpsRequest: sl-volume-exp-topology
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo solr-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"11Gi"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-02cddba5-1d6a-4f1b-91b2-a7e55857b6b7   11Gi       RWO            Delete           Bound    demo/solr-combined-data-solr-combined-1   longhorn       <unset>                          33m
pvc-61b8f97a-a588-4125-99f3-604f6a70d560   11Gi       RWO            Delete           Bound    demo/solr-combined-data-solr-combined-0   longhorn       <unset>                          33m
```

The above output verifies that we have successfully expanded the volume of the Solr.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete solropsrequest -n demo sl-volume-exp-combined
kubectl delete sl -n demo solr-combined
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
