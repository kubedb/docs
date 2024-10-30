---
title: Solr Storage Autoscaling Topology
menu:
  docs_{{ .version }}:
    identifier: sl-storage-autoscaling-topology
    name: Solr Topology Autoscaling
    parent: sl-storage-autoscaling-solr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of Solr Topology Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a solr topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrAutoscaler](/docs/guides/solr/concepts/autoscaler.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Storage Autoscaling Overview](/docs/guides/solr/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in this [directory](/docs/examples/solr/autoscaler/storage) of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get sc
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  11d
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   7d22h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   7d22h

```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `Solr` topology cluster using a supported version by the `KubeDB` operator. Then we are going to apply `SolrAutoscaler` to set up autoscaling.

#### Deploy Solr Topology

In this section, we are going to deploy a Solr topology cluster with version `9.6.1`.  Then, in the next section we will set up autoscaling for this database using `SolrAutoscaler` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.6.1
  zookeeperRef:
    name: zoo
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        storageClassName: longhorn
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        storageClassName: longhorn
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      storage:
        storageClassName: longhorn
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi

```

Let's create the `Solr` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscaler/topology.yaml
Solr.kubedb.com/es-topology created
```

Now, wait until `solr-cluster` has status `Ready`. i.e,

```bash
 $ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.6.1     Ready    83s

```

Let's check volume size from the data petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo solr-cluster-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "1Gi"
  }
}

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-24431af2-8df5-4ad2-a6cd-795dcbdc6355   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-coordinator-0   longhorn       <unset>                          2m15s
pvc-5e3430da-545c-4234-a891-3385b100401d   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-overseer-0      longhorn       <unset>                          2m17s
pvc-aa75a15f-94cd-475a-a7ad-498023830020   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-data-0          longhorn       <unset>                          2m19s

```

You can see that the data PetSet has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `SolrAutoscaler` CRO to set up storage autoscaling for the data nodes.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using an SolrAutoscaler Object.

#### Create SolrAutoscaler Object

To set up vertical autoscaling for this topology cluster, we have to create a `SolrAutoscaler` CRO with our desired configuration. Below is the YAML of the `SolrAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SolrAutoscaler
metadata:
  name: sl-storage-autoscaler-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  storage:
    data:
      expansionMode: "Offline"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 100
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `es-topology` cluster.
- `spec.storage.topology.data.trigger` specifies that storage autoscaling is enabled for data nodes.
- `spec.storage.topology.data.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.topology.data.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.

> Note: In this demo we are only setting up the storage autoscaling for the data nodes, that's why we only specified the data section of the autoscaler. You can enable autoscaling for master nodes and ingest nodes in the same YAML, by specifying the `topology.master` and `topology.ingest` respectivly.

Let's create the `SolrAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscaler/storage/topology-scaler.yaml 
solrautoscaler.autoscaling.kubedb.com/sl-storage-autoscaler-topology created
```

#### Storage Autoscaling is set up successfully

Let's check that the `solrautoscaler` resource is created successfully,

```bash
$ kubectl get solrautoscaler -n demo
NAME                             AGE
sl-storage-autoscaler-topology   70s

$ kubectl describe solrautoscaler -n demo sl-storage-autoscaler-topology
Name:         sl-storage-autoscaler-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         SolrAutoscaler
Metadata:
  Creation Timestamp:  2024-10-30T06:55:55Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Solr
    Name:                  solr-cluster
    UID:                   0820762a-3b96-44db-8157-f1857bed410e
  Resource Version:        976749
  UID:                     8a5b2ca2-3fa1-4e22-9b4b-bde4a163aa08
Spec:
  Database Ref:
    Name:  solr-cluster
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Data:
      Expansion Mode:  Offline
      Scaling Rules:
        Applies Upto:     
        Threshold:        100pc
      Scaling Threshold:  100
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>


```

So, the `solrautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up one of the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the data nodes and fill the database volume using the following commands:

```bash
 $ kubectl exec -it -n demo solr-cluster-data-0 -- bash
Defaulted container "solr" out of: solr, init-solr (init)
solr@solr-combined-0:/opt/solr-9.6.1$ df -h /var/solr/data
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-d9c2f7c1-7c27-48bd-a87e-cb1935cc2e61  7.1G  601M  6.5G   9% /var/solr/data
solr@solr-cluster-data-0:/opt/solr-9.6.1$ dd if=/dev/zero of=/var/solr/data/file.img bs=300M count=2
2+0 records in
2+0 records out
629145600 bytes (629 MB, 600 MiB) copied, 1.95395 s, 322 MB/s
solr@solr-cluster-data-0:/opt/solr-9.6.1$ df -h /var/solr/data
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-aa75a15f-94cd-475a-a7ad-498023830020  974M  601M  358M  63% /var/solr/data

```

So, from the above output we can see that the storage usage is 69%, which exceeded the `usageThreshold` 60%.

Let's watch the `solropsrequest` in the demo namespace to see if any `solropsrequest` object is created. After some time you'll see that an `solropsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get slops -n demo
NAME                        TYPE              STATUS        AGE
slops-solr-cluster-0s6kgw   VolumeExpansion   Progressing   95s
```

Let's wait for the opsRequest to become successful.

```bash
$ kubectl get slops -n demo
NAME                        TYPE              STATUS       AGE
slops-solr-cluster-0s6kgw   VolumeExpansion   Successful   2m58s
```

We can see from the above output that the `solrOpsRequest` has succeeded. If we describe the `solrOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe slops -n demo slops-solr-cluster-0s6kgw 
Name:         slops-solr-cluster-0s6kgw
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=solr-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=solrs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-10-30T06:58:43Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  SolrAutoscaler
    Name:                  sl-storage-autoscaler-topology
    UID:                   8a5b2ca2-3fa1-4e22-9b4b-bde4a163aa08
  Resource Version:        977641
  UID:                     5411ed48-b2fe-40a2-b1e4-2d3d659668b1
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Data:  2041405440
    Mode:  Offline
Status:
  Conditions:
    Last Transition Time:  2024-10-30T06:58:43Z
    Message:               Solr ops-request has started to expand volume of solr nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-10-30T06:59:01Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-10-30T06:58:51Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2024-10-30T06:58:51Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2024-10-30T07:01:16Z
    Message:               successfully updated data node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionDataNode
    Status:                True
    Type:                  VolumeExpansionDataNode
    Last Transition Time:  2024-10-30T06:59:06Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-10-30T06:59:06Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2024-10-30T06:59:06Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-10-30T06:59:11Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-30T06:59:11Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2024-10-30T07:00:56Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-10-30T07:00:56Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-10-30T07:01:01Z
    Message:               running solr; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningSolr
    Last Transition Time:  2024-10-30T07:01:21Z
    Message:               successfully reconciled the Solr resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-30T07:01:26Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-10-30T07:01:26Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-10-30T07:01:26Z
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
  Normal   Starting                                 3m39s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-solr-cluster-0s6kgw
  Normal   Starting                                 3m39s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                               3m39s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: slops-solr-cluster-0s6kgw
  Warning  get petset; ConditionStatus:True         3m31s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True      3m31s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True         3m26s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True

```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the data nodes of the cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo solr-cluster-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "2041405440"
  }
}


$ kubectl get pvc -n demo 
NAME                                           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
solr-cluster-data-solr-cluster-coordinator-0   Bound    pvc-24431af2-8df5-4ad2-a6cd-795dcbdc6355   1Gi        RWO            longhorn       <unset>                 18m
solr-cluster-data-solr-cluster-data-0          Bound    pvc-aa75a15f-94cd-475a-a7ad-498023830020   1948Mi     RWO            longhorn       <unset>                 18m
solr-cluster-data-solr-cluster-overseer-0      Bound    pvc-5e3430da-545c-4234-a891-3385b100401d   1Gi        RWO            longhorn       <unset>                 18m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-24431af2-8df5-4ad2-a6cd-795dcbdc6355   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-coordinator-0   longhorn       <unset>                          18m
pvc-5e3430da-545c-4234-a891-3385b100401d   1Gi        RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-overseer-0      longhorn       <unset>                          18m
pvc-aa75a15f-94cd-475a-a7ad-498023830020   1948Mi     RWO            Delete           Bound    demo/solr-cluster-data-solr-cluster-data-0          longhorn       <unset>                          18m
```

The above output verifies that we have successfully autoscaler the volume of the data nodes of this Solr topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete Solr -n demo solr-cluster
$ kubectl delete solrautoscaler -n demo sl-storage-autoscaler-topology
```
