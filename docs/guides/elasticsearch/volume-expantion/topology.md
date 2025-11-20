---
title: Elasticsearch Topology Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: es-volume-expansion-topology
    name: Topology
    parent: es-voulume-expansion-elasticsearch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Topology Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Elasticsearch Topology Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Topology](/docs/guides/elasticsearch/clustering/topology-cluster/_index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Volume Expansion Overview](/docs/guides/elasticsearch/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Topology Elasticsearch Cluster

Here, we are going to deploy a `Elasticsearch` topology using a supported version by `KubeDB` operator. Then we are going to apply `ElasticsearchOpsRequest` to expand its volume.

### Prepare Elasticsearch Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Elasticsearch` combined cluster with version `xpack-8.11.1`.

### Deploy Elasticsearch

In this section, we are going to deploy a Elasticsearch topology cluster for broker and controller with 1GB volume. Then, in the next section we will expand its volume to 2GB using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  enableSSL: true 
  version: xpack-8.11.1
  storageType: Durable
  topology:
    master:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
              
    data:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}docs/examples/elasticsearch/clustering/topology-es.yaml
Elasticsearch.kubedb.com/es-cluster created
```

Now, wait until `es-cluster` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo
NAME         VERSION        STATUS   AGE
es-cluster   xpack-8.11.1   Ready    22h

```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo es-cluster-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get petset -n demo es-cluster-master -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get petset -n demo es-cluster-ingest -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-11b48c6e-d996-45a7-8ba2-f8d71a655912   1Gi        RWO            Delete           Bound    demo/data-es-cluster-ingest-2   local-path     <unset>                          22h
pvc-1904104c-bbf2-4754-838a-8a647b2bd23e   1Gi        RWO            Delete           Bound    demo/data-es-cluster-data-2     local-path     <unset>                          22h
pvc-19aa694a-29c0-43d9-a495-c84c77df2dd8   1Gi        RWO            Delete           Bound    demo/data-es-cluster-master-0   local-path     <unset>                          22h
pvc-33702b18-7e98-41b7-9b19-73762cb4f86a   1Gi        RWO            Delete           Bound    demo/data-es-cluster-master-1   local-path     <unset>                          22h
pvc-8604968f-f433-4931-82bc-8d240d6f52d8   1Gi        RWO            Delete           Bound    demo/data-es-cluster-data-0     local-path     <unset>                          22h
pvc-ae5ccc43-d078-4816-a553-8a3cd1f674be   1Gi        RWO            Delete           Bound    demo/data-es-cluster-ingest-0   local-path     <unset>                          22h
pvc-b4225042-c69f-41df-99b2-1b3191057a85   1Gi        RWO            Delete           Bound    demo/data-es-cluster-data-1     local-path     <unset>                          22h
pvc-bd4b7d5a-8494-4ee2-a25c-697a6f23cb79   1Gi        RWO            Delete           Bound    demo/data-es-cluster-ingest-1   local-path     <unset>                          22h
pvc-c9057b3b-4412-467f-8ae5-f6414e0059c3   1Gi        RWO            Delete           Bound    demo/data-es-cluster-master-2   local-path     <unset>                          22h
```

You can see the petsets have 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `ElasticsearchOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the Elasticsearch topology cluster.

#### Create ElasticsearchOpsRequest

In order to expand the volume of the database, we have to create a `ElasticsearchOpsRequest` CR with our desired volume size. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: volume-expansion-topology
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: es-cluster
  volumeExpansion:
    mode: "Online"
    master: 5Gi
    data: 5Gi
    ingest: 4Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `es-cluster`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.data` specifies the desired volume size for data node.
- `spec.volumeExpansion.master` specifies the desired volume size for master node.
- `spec.volumeExpansion.ingest` specifies the desired volume size for ingest node.

> If you want to expand the volume of only one node, you can specify the desired volume size for that node only.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/volume-expansion/elasticsearch-volume-expansion-topology.yaml
Elasticsearchopsrequest.ops.kubedb.com/volume-expansion-topology created
```

#### Verify Elasticsearch Topology volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Elasticsearch` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get Elasticsearchopsrequest -n demo
NAME                        TYPE              STATUS       AGE
volume-expansion-topology   VolumeExpansion   Successful   44m

```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to expand the volume of Elasticsearch.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo volume-expansion-topology
Name:         volume-expansion-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-20T10:07:17Z
  Generation:          1
  Resource Version:    115931
  UID:                 38107c4f-4249-4597-b8b4-06a445891872
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Data:    5Gi
    Ingest:  4Gi
    Master:  5Gi
    Mode:    Offline
Status:
  Conditions:
    Last Transition Time:  2025-11-20T10:07:17Z
    Message:               Elasticsearch ops request is expanding volume of the Elasticsearch nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-11-20T10:07:25Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2025-11-20T10:07:25Z
    Message:               delete pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  deletePetSet
    Last Transition Time:  2025-11-20T10:07:55Z
    Message:               successfully deleted the PetSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2025-11-20T10:08:00Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2025-11-20T10:08:00Z
    Message:               patch opsrequest; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsrequest
    Last Transition Time:  2025-11-20T10:20:20Z
    Message:               create db client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateDbClient
    Last Transition Time:  2025-11-20T10:08:00Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2025-11-20T10:08:05Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-11-20T10:19:55Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-11-20T10:11:05Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2025-11-20T10:11:40Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-11-20T10:13:55Z
    Message:               successfully updated ingest node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionIngestNode
    Status:                True
    Type:                  VolumeExpansionIngestNode
    Last Transition Time:  2025-11-20T10:14:00Z
    Message:               db operation; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DbOperation
    Last Transition Time:  2025-11-20T10:17:15Z
    Message:               successfully updated data node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionDataNode
    Status:                True
    Type:                  VolumeExpansionDataNode
    Last Transition Time:  2025-11-20T10:20:25Z
    Message:               successfully updated master node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionMasterNode
    Status:                True
    Type:                  VolumeExpansionMasterNode
    Last Transition Time:  2025-11-20T10:21:02Z
    Message:               successfully reconciled the Elasticsearch resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-11-20T10:21:07Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2025-11-20T10:21:12Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-20T10:21:12Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                   Age   From                         Message
  ----     ------                                   ----  ----                         -------
  Normal   PauseDatabase                            45m   KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-cluster
  Warning  get pet set; ConditionStatus:True        45m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  delete pet set; ConditionStatus:True     45m   KubeDB Ops-manager Operator  delete pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        44m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        44m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  delete pet set; ConditionStatus:True     44m   KubeDB Ops-manager Operator  delete pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        44m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        44m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  delete pet set; ConditionStatus:True     44m   KubeDB Ops-manager Operator  delete pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        44m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                         44m   KubeDB Ops-manager Operator  successfully deleted the PetSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   44m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   44m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         44m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   44m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            44m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            42m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    41m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         41m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   41m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  41m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   40m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   40m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   40m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         40m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          40m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   40m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    40m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         40m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   40m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  40m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            40m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   39m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   39m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   39m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         39m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          39m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   39m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            39m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    39m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         39m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   39m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  38m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   38m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Normal   VolumeExpansionIngestNode                38m   KubeDB Ops-manager Operator  successfully updated ingest node PVC sizes
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   38m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   38m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       38m   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         38m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          38m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   38m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            38m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    38m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         38m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   38m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  37m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   37m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       37m   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   37m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   37m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       37m   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         37m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          37m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   37m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            37m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    36m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         36m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   36m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  36m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   36m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       36m   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   36m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   36m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       36m   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         36m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          36m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   36m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            36m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    35m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         35m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   35m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  35m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   35m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       35m   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Normal   VolumeExpansionDataNode                  35m   KubeDB Ops-manager Operator  successfully updated data node PVC sizes
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   35m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   35m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         35m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            35m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          35m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   35m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    34m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         34m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   34m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  34m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   34m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   34m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   34m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         34m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            34m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          34m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   34m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    33m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         33m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   33m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  33m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   33m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   33m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   33m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         33m   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          32m   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   32m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    32m   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         32m   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   32m   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  32m   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   32m   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Normal   VolumeExpansionMasterNode                32m   KubeDB Ops-manager Operator  successfully updated master node PVC sizes
  Normal   UpdatePetSets                            31m   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Warning  get pet set; ConditionStatus:True        31m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        31m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        31m   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             31m   KubeDB Ops-manager Operator  PetSet is recreated
  Normal   UpdateDatabase                           31m   KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                           31m   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-cluster
  Normal   ResumeDatabase                           31m   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-cluster
  Normal   Successful                               31m   KubeDB Ops-manager Operator  Successfully Updated Database
  Normal   UpdatePetSets                            31m   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources

```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo es-cluster-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"5Gi"
$ kubectl get petset -n demo es-cluster-master -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"5Gi"
$ kubectl get petset -n demo es-cluster-ingest -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"4Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-37f7398d-0251-4d3c-a439-d289b8cec6d2   5Gi        RWO            Delete           Bound    demo/data-es-cluster-master-2   standard       <unset>                          111m
pvc-3a5d2b3e-dd39-4468-a8da-5274992a6502   5Gi        RWO            Delete           Bound    demo/data-es-cluster-master-0   standard       <unset>                          111m
pvc-3cf21868-4b51-427b-b7ef-d0d26c753c8b   5Gi        RWO            Delete           Bound    demo/data-es-cluster-master-1   standard       <unset>                          111m
pvc-56e6ed8f-a729-4532-bdec-92b8101f7813   5Gi        RWO            Delete           Bound    demo/data-es-cluster-data-2     standard       <unset>                          111m
pvc-783d51f7-3bf2-4121-8f18-357d14d003ad   4Gi        RWO            Delete           Bound    demo/data-es-cluster-ingest-0   standard       <unset>                          111m
pvc-81d6c1d3-0aa6-4190-9ee0-dd4a8d62b6b3   4Gi        RWO            Delete           Bound    demo/data-es-cluster-ingest-2   standard       <unset>                          111m
pvc-942c6dce-4701-4e1a-b6f9-bf7d4ab56a11   5Gi        RWO            Delete           Bound    demo/data-es-cluster-data-1     standard       <unset>                          111m
pvc-b706647d-c9ba-4296-94aa-2f6ef2230b6e   4Gi        RWO            Delete           Bound    demo/data-es-cluster-ingest-1   standard       <unset>                          111m
pvc-c274f913-5452-47e1-ab42-ba584bdae297   5Gi        RWO            Delete           Bound    demo/data-es-cluster-data-0     standard       <unset>                          111m
```

The above output verifies that we have successfully expanded the volume of the Elasticsearch.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete Elasticsearchopsrequest -n demo volume-expansion-topology
kubectl delete es -n demo es-cluster
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
