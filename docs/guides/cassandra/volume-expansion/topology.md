---
title: Cassandra Topology Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: cas-volume-expansion-topology
    name: Topology
    parent: cas-volume-expansion
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra Topology Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Cassandra Topology Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)
    - [Volume Expansion Overview](/docs/guides/cassandra/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Topology Cassandra Cluster

Here, we are going to deploy a `Cassandra` topology using a supported version by `KubeDB` operator. Then we are going to apply `CassandraOpsRequest` to expand its volume.

### Prepare Cassandra Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  5d22h
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   6s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   3s

```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Cassandra` combined cluster with version `5.0.3`.

### Deploy Cassandra

In this section, we are going to deploy a Cassandra topology cluster for broker and controller with 1GB volume. Then, in the next section we will expand its volume to 2GB using `CassandraOpsRequest` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          storageClassName:
            longhorn
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 3Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/volume-expansion/cassandra.yaml
cassandra.kubedb.com/cassandra-prod created
```

Now, wait until `cassandra-prod` has status `Ready`. i.e,

```bash
$ kubectl get cas -n demo -w
NAME             TYPE                  VERSION   STATUS         AGE
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   4s
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   37s
..
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Ready          2m3s

```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo cassandra-prod-rack-r0 -o json | jq '.spec.volumeClaimTemplates[0].spec.resources.requests.storage'
"1Gi"


$  kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                              STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-623e4d80-f508-4bb1-a4cb-4ebdbcbd8495   1Gi        RWO            Delete           Bound    demo/data-cassandra-prod-rack-r0-0                 longhorn       <unset>                          82s
pvc-76b5a0a7-d234-426c-a4cc-ec740d6456ba   1Gi        RWO            Delete           Bound    demo/data-cassandra-prod-rack-r0-1                 longhorn       <unset>                          64s
pvc-84588238-9fea-4ac3-9cfa-f7b04640697c   1Gi        RWO            Delete           Bound    demo/nodetool-cassandra-prod-rack-r0-0             longhorn       <unset>                          82s
pvc-849d404c-d078-4802-a28f-6834d8c81998   1Gi        RWO            Delete           Bound    demo/nodetool-cassandra-prod-rack-r0-1             longhorn       <unset>                          64s
pvc-85a6902d-a596-45db-94f9-1ce355600323   1Gi        RWO            Delete           Bound    demo/main-config-volume-cassandra-prod-rack-r0-1   longhorn       <unset>                          64s
pvc-88d6586e-b502-481d-91fc-dd6381d9b1c0   1Gi        RWO            Delete           Bound    demo/main-config-volume-cassandra-prod-rack-r0-0   longhorn       <unset>                          82s
```

You can see the petsets have 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `CassandraOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the cassandra topology cluster.

#### Create CassandraOpsRequest

In order to expand the volume of the database, we have to create a `CassandraOpsRequest` CR with our desired volume size. Below is the YAML of the `CassandraOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cas-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: cassandra-prod
  volumeExpansion:
    node: 2Gi
    mode: Online
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `cassandra-prod`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.broker` specifies the desired volume size for broker node.
- `spec.volumeExpansion.controller` specifies the desired volume size for controller node.

> If you want to expand the volume of only one node, you can specify the desired volume size for that node only.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/volume-expansion/cassandra-volume-expansion-opsreq.yaml
cassandraopsrequest.ops.kubedb.com/cas-volume-expansion created
```

#### Verify Cassandra Topology volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Cassandra` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CR,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                   TYPE              STATUS       AGE
cas-volume-expansion   VolumeExpansion   Successful   5m49s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to expand the volume of cassandra.

```bash
$ kubectl describe cassandraopsrequest -n demo cas-volume-expansion
Name:         cas-volume-expansion
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-29T10:07:19Z
  Generation:          1
  Resource Version:    306444
  UID:                 ccc1253b-b69f-4242-a270-fd36c2a15daa
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-prod
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Online
    Node:  2Gi
Status:
  Conditions:
    Last Transition Time:  2025-07-29T10:08:29Z
    Message:               Cassandra ops-request has started to expand volume of cassandra nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-07-29T10:08:56Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2025-07-29T10:08:46Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2025-07-29T10:08:46Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2025-07-29T10:10:56Z
    Message:               successfully updated PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2025-07-29T10:09:01Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-07-29T10:09:01Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-07-29T10:10:51Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-07-29T10:11:09Z
    Message:               successfully reconciled the Cassandra resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-29T10:11:19Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2025-07-29T10:11:14Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2025-07-29T10:11:14Z
    Message:               Successfully completed volumeExpansion for Cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                  Age    From                         Message
  ----     ------                                  ----   ----                         -------
  Normal   Starting                                7m56s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/cas-volume-expansion
  Normal   Starting                                7m56s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                              7m56s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cas-volume-expansion
  Warning  get petset; ConditionStatus:True        7m39s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True     7m39s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True        7m34s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Normal   OrphanPetSetPods                        7m29s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True           7m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True         7m24s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           7m19s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False  7m19s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True           7m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           7m9s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           7m4s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m29s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   6m24s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m19s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True         6m19s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False  6m14s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True           6m9s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           6m4s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   5m34s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateNodePVCs                          5m29s  KubeDB Ops-manager Operator  successfully updated PVC sizes
  Normal   UpdatePetSets                           5m16s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra resources
  Normal   UpdatePetSets                           5m13s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra resources
  Warning  get pet set; ConditionStatus:True       5m11s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                            5m11s  KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                5m11s  KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                              5m11s  KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cas-volume-expansion
  Normal   Starting                                5m11s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                              5m11s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cas-volume-expansion
  Warning  get pet set; ConditionStatus:True       5m6s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                            5m6s   KubeDB Ops-manager Operator  PetSet is recreated
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo cassandra-prod-rack-r0 -o json | jq '.spec.volumeClaimTemplates[0].spec.resources.requests.storage'
"2Gi"


$  kubectl get pv -n demo 
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                              STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-7efa007f-5fb2-4e64-aea0-233ec456703f   1Gi        RWO            Delete           Bound    demo/main-config-volume-cassandra-prod-rack-r0-1   longhorn       <unset>                          11m
pvc-88f8ec55-ae5b-48dd-9b3a-0dc113cdaa43   2Gi        RWO            Delete           Bound    demo/data-cassandra-prod-rack-r0-1                 longhorn       <unset>                          11m
pvc-c455344e-0f17-42d2-8fd9-062fa2f1b0a1   1Gi        RWO            Delete           Bound    demo/main-config-volume-cassandra-prod-rack-r0-0   longhorn       <unset>                          11m
pvc-d19d62ac-b37b-406e-a7d6-a10f4f74d929   2Gi        RWO            Delete           Bound    demo/data-cassandra-prod-rack-r0-0                 longhorn       <unset>                          11m
pvc-df152a7a-12ea-4690-b64e-ccb2decce8cd   1Gi        RWO            Delete           Bound    demo/nodetool-cassandra-prod-rack-r0-1             longhorn       <unset>                          11m
pvc-f8420d54-05f8-4ea2-b70c-11a3737e04e4   1Gi        RWO            Delete           Bound    demo/nodetool-cassandra-prod-rack-r0-0             longhorn       <unset>                          11m
```

The above output verifies that we have successfully expanded the data related volume of the Cassandra.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cassandraopsrequest -n demo cas-volume-expansion
kubectl delete cas -n demo cassandra-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).
- 
[//]: # (- Monitor your Cassandra database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/cassandra/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
