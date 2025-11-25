---
title: Elasticsearch Combined Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: es-volume-expansion-combined
    name: Combined
    parent: es-voulume-expansion-elasticsearch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Combined Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Elasticsearch Combined Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Combined](/docs/guides/elasticsearch/clustering/combined-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Volume Expansion Overview](/docs/guides/elasticsearch/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Combined Elasticsearch Cluster

Here, we are going to deploy a `Elasticsearch` combined using a supported version by `KubeDB` operator. Then we are going to apply `ElasticsearchOpsRequest` to expand its volume.

### Prepare Elasticsearch Combined CLuster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Elasticsearch` combined cluster with version `xpack-8.11.1`.

### Deploy Elasticsearch

In this section, we are going to deploy a Elasticsearch combined cluster with 1GB volume. Then, in the next section we will expand its volume to 2GB using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-combined
  namespace: demo
spec:
  version: xpack-8.11.1
  enableSSL: true
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut

```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/clustering/multi-nodes-es.yaml
Elasticsearch.kubedb.com/es-combined created
```

Now, wait until `es-combined` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo -w
NAME          VERSION        STATUS   AGE
es-combined   xpack-8.11.1   Ready    75s

```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo es-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-edeeff75-9823-4aeb-9189-37adad567ec7   1Gi        RWO            Delete           Bound    demo/data-es-combined-0   longhorn       <unset>                          2m21s

```

You can see the petset has 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `ElasticsearchOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the Elasticsearch combined cluster.

#### Create ElasticsearchOpsRequest

In order to expand the volume of the database, we have to create a `ElasticsearchOpsRequest` CR with our desired volume size. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: es-volume-expansion-combined
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: es-combined
  volumeExpansion:
    mode: "Online"
    node: 4Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `es-combined`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.node` specifies the desired volume size.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/volume-expansion/elasticsearch-volume-expansion-combined.yaml
Elasticsearchopsrequest.ops.kubedb.com/es-volume-exp-combined created
```

#### Verify Elasticsearch Combined volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Elasticsearch` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get Elasticsearchopsrequest -n demo
NAME                     TYPE              STATUS       AGE
es-volume-exp-combined   VolumeExpansion   Successful   2m4s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo es-volume-expansion-combined
Name:         es-volume-expansion-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-20T12:19:05Z
  Generation:          1
  Resource Version:    127891
  UID:                 4199c88c-d3c4-44d0-8084-efdaa49b9c03
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-combined
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Offline
    Node:  4Gi
Status:
  Conditions:
    Last Transition Time:  2025-11-20T12:19:05Z
    Message:               Elasticsearch ops request is expanding volume of the Elasticsearch nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-11-20T12:19:13Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2025-11-20T12:19:13Z
    Message:               delete pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  deletePetSet
    Last Transition Time:  2025-11-20T12:19:23Z
    Message:               successfully deleted the PetSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2025-11-20T12:19:28Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2025-11-20T12:19:28Z
    Message:               patch opsrequest; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsrequest
    Last Transition Time:  2025-11-20T12:20:23Z
    Message:               create db client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateDbClient
    Last Transition Time:  2025-11-20T12:19:28Z
    Message:               db operation; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DbOperation
    Last Transition Time:  2025-11-20T12:19:28Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2025-11-20T12:19:33Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-11-20T12:19:33Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-11-20T12:19:58Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-11-20T12:19:58Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2025-11-20T12:20:28Z
    Message:               successfully updated combined node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionCombinedNode
    Status:                True
    Type:                  VolumeExpansionCombinedNode
    Last Transition Time:  2025-11-20T12:20:37Z
    Message:               successfully reconciled the Elasticsearch resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-11-20T12:20:42Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2025-11-20T12:20:48Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-20T12:20:48Z
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
  Normal   PauseDatabase                            114s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-combined
  Warning  get pet set; ConditionStatus:True        106s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  delete pet set; ConditionStatus:True     106s  KubeDB Ops-manager Operator  delete pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        101s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                         96s   KubeDB Ops-manager Operator  successfully deleted the PetSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True            91s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   91s   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   91s   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       91s   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         91s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            86s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            86s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          86s   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   86s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            81s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            81s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            76s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            76s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            71s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            71s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            66s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            66s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            61s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            61s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    61s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         61s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch opsrequest; ConditionStatus:True   61s   KubeDB Ops-manager Operator  patch opsrequest; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            56s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:False  56s   KubeDB Ops-manager Operator  create db client; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            51s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            46s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            41s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            36s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  create db client; ConditionStatus:True   36s   KubeDB Ops-manager Operator  create db client; ConditionStatus:True
  Warning  db operation; ConditionStatus:True       36s   KubeDB Ops-manager Operator  db operation; ConditionStatus:True
  Normal   VolumeExpansionCombinedNode              31s   KubeDB Ops-manager Operator  successfully updated combined node PVC sizes
  Normal   UpdatePetSets                            22s   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Warning  get pet set; ConditionStatus:True        17s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             17s   KubeDB Ops-manager Operator  PetSet is recreated
  Normal   UpdateDatabase                           11s   KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                           11s   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-combined
  Normal   ResumeDatabase                           11s   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-combined
  Normal   Successful                               11s   KubeDB Ops-manager Operator  Successfully Updated Database

```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$  kubectl get petset -n demo es-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"4Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-edeeff75-9823-4aeb-9189-37adad567ec7   4Gi        RWO            Delete           Bound    demo/data-es-combined-0   longhorn       <unset>                          13m
```

The above output verifies that we have successfully expanded the volume of the Elasticsearch.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete Elasticsearchopsrequest -n demo es-volume-expansion-combined
kubectl delete es -n demo es-combined
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/topology-cluster/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
