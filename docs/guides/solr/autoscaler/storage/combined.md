---
title: Solr Storage Autoscaling Combined
menu:
  docs_{{ .version }}:
    identifier: sl-storage-autoscaling-combined
    name: Solr Combined Autoscaling
    parent: sl-storage-autoscaling-solr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of Solr Combined Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of an Solr combined cluster.

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

## Storage Autoscaling of Combined cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get sc
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  11d
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   7d21h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   7d21h
```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install longhorn from [here](https://longhorn.io/docs/1.7.2/deploy/install/install-with-kubectl/)

Now, we are going to deploy a `Solr` combined cluster using a supported version by the `KubeDB` operator. Then we are going to apply `SolrAutoscaler` to set up autoscaling.

#### Deploy Solr Combined Cluster

In this section, we are going to deploy a solr combined cluster with version `9.6.1`.  Then, in the next section we will set up autoscaling for this database using `SolrAutoscaler` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-combined
  namespace: demo
spec:
  version: 9.6.1
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
    storageClassName: longhorn
```

Let's create the `Solr` CRD we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscaler/storage/combined-scaler.yaml
solr.kubedb.com/solr-combined created
```

Now, wait until `solr-combined` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME            TYPE                  VERSION   STATUS   AGE
solr-combined   kubedb.com/v1alpha2   9.6.1     Ready    17m

```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo solr-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "1Gi"
  }
}


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-ceee299c-5c50-4f5c-83d5-97e2423bf286   7332Mi     RWO            Delete           Bound    demo/solr-combined-data-solr-combined-1   longhorn       <unset>                          19m
pvc-d9c2f7c1-7c27-48bd-a87e-cb1935cc2e61   7332Mi     RWO            Delete           Bound    demo/solr-combined-data-solr-combined-0   longhorn       <unset>                          19m
```

You can see the PetSet has 1GB storage, and the capacity of the persistent volume is also 1GB.

We are now ready to apply the `SolrAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a SolrAutoscaler Object.

#### Create SolrAutoscaler Object

To set up vertical autoscaling for the combined cluster nodes, we have to create a `SolrAutoscaler` CRO with our desired configuration. Below is the YAML of the `SolrAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SolrAutoscaler
metadata:
  name: sl-storage-autoscaler-combined
  namespace: demo
spec:
  databaseRef:
    name: solr-combined
  storage:
    node:
      expansionMode: "Offline"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 100
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `es-combined` cluster.
- `spec.storage.node.trigger` specifies that storage autoscaling is enabled for the Solr nodes.
- `spec.storage.node.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.node.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.

Let's create the `SolrAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/autoscaler/storage/combined-scaler.yaml 
solrautoscaler.autoscaling.kubedb.com/sl-storage-autoscaler-combined created
```

#### Storage Autoscaling is set up successfully

Let's check that the `Solrautoscaler` resource is created successfully,

```bash
$ kubectl get solrautoscaler -n demo
NAME                             AGE
sl-storage-autoscaler-combined   20m


$ kubectl describe solrautoscaler -n demo sl-storage-autoscaler-combined 
Name:         sl-storage-autoscaler-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         SolrAutoscaler
Metadata:
  Creation Timestamp:  2024-10-30T05:57:51Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Solr
    Name:                  solr-combined
    UID:                   2f180d2f-27ef-4f94-8563-83fc0ae2bf66
  Resource Version:        971668
  UID:                     f2d12f05-790a-40ad-97b1-28f890a45dd7
Spec:
  Database Ref:
    Name:  solr-combined
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Node:
      Expansion Mode:  Offline
      Scaling Rules:
        Applies Upto:     
        Threshold:        100pc
      Scaling Threshold:  100
      Trigger:            On
      Usage Threshold:    10
Status:
  Conditions:
    Last Transition Time:  2024-10-30T06:11:43Z
    Message:               Successfully created solrOpsRequest demo/slops-solr-combined-gzqvx7
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
Events:                    <none>
```

So, the `solrautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume using the following commands:

```bash
$ kubectl exec -it -n demo solr-combined-0 -- bash
Defaulted container "solr" out of: solr, init-solr (init)
solr@solr-combined-0:/opt/solr-9.6.1$ df -h /var/solr/data
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-d9c2f7c1-7c27-48bd-a87e-cb1935cc2e61  7.1G  601M  6.5G   9% /var/solr/data

[root@es-combined-0 Solr]# dd if=/dev/zero of=/var/solrdata/file.img bs=300M count=2
1+0 records in
1+0 records out
629145600 bytes (629 MB) copied, 1.95767 s, 321 MB/s

[root@es-combined-0 Solr]# df -h /usr/share/Solr/data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-d9c2f7c1-7c27-48bd-a87e-cb1935cc2e61  7.1G  601M  6.5G 63% /var/solr/data
```

So, from the above output, we can see that the storage usage is 64%, which exceeded the `usageThreshold` 60%.

Let's watch the `solropsrequest` in the demo namespace to see if any `solropsrequest` object is created. After some time you'll see that a `Solropsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get slops -n demo
NAME                         TYPE              STATUS       AGE
slops-solr-combined-gzqvx7   VolumeExpansion   Progressing   9m42s
```

Let's wait for the opsRequest to become successful.

```bash
$ kubectl get esops -n demo
NAME                         TYPE              STATUS        AGE
slops-solr-combined-gzqvx7   VolumeExpansion   Successful    19m
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe slops -n demo slops-solr-combined-gzqvx7 
Name:         slops-solr-combined-gzqvx7
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=solr-combined
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=solrs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-10-30T06:11:43Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  SolrAutoscaler
    Name:                  sl-storage-autoscaler-combined
    UID:                   f2d12f05-790a-40ad-97b1-28f890a45dd7
  Resource Version:        972599
  UID:                     662a363f-9ce9-4c93-b2eb-8200c23978f6
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-combined
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Offline
    Node:  7687602176
Status:
  Conditions:
    Last Transition Time:  2024-10-30T06:11:43Z
    Message:               Solr ops-request has started to expand volume of solr nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-10-30T06:12:01Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-10-30T06:11:51Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2024-10-30T06:11:51Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2024-10-30T06:16:26Z
    Message:               successfully updated combined node PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionCombinedNode
    Status:                True
    Type:                  VolumeExpansionCombinedNode
    Last Transition Time:  2024-10-30T06:12:06Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-10-30T06:12:06Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2024-10-30T06:12:06Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-10-30T06:12:11Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-30T06:12:11Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2024-10-30T06:16:06Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-10-30T06:14:01Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-10-30T06:16:11Z
    Message:               running solr; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningSolr
    Last Transition Time:  2024-10-30T06:16:31Z
    Message:               successfully reconciled the Solr resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-30T06:16:36Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-10-30T06:16:36Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-10-30T06:16:36Z
    Message:               Successfully completed volumeExpansion for Solr
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the combined cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo solr-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "7687602176"
  }
}


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-ceee299c-5c50-4f5c-83d5-97e2423bf286   7332Mi     RWO            Delete           Bound    demo/solr-combined-data-solr-combined-1   longhorn       <unset>                          26m
pvc-d9c2f7c1-7c27-48bd-a87e-cb1935cc2e61   7332Mi     RWO            Delete           Bound    demo/solr-combined-data-solr-combined-0   longhorn       <unset>
```

The above output verifies that we have successfully autoscaler the volume of the Solr combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete solr -n demo solr-combined
$ kubectl delete solrautoscaler -n demo sl-storage-autoscaler-combined
```
