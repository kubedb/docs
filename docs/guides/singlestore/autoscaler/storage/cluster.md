---
title: SingleStore Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: sdb-storage-auto-scaling-cluster
    name: Cluster
    parent: sdb-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a SingleStore Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a SingleStore cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
    - [SingleStoreAutoscaler](/docs/guides/singlestore/concepts/autoscaler.md)
    - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
    - [Storage Autoscaling Overview](/docs/guides/singlestore/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/singlestore](/docs/examples/singlestore) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of SingleStore Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

#### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

#### Deploy SingleStore Cluster

In this section, we are going to deploy a SingleStore with version `8.7.10`.  Then, in the next section we will set up autoscaling for this database using `SingleStoreAutoscaler` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: 8.7.10
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "0.7"
                requests:
                  memory: "2Gi"
                  cpu: "0.7"
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 3
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "0.7"
                requests:
                  memory: "2Gi"
                  cpu: "0.7"
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut

```
Let's create the `SingleStore` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/autoscaling/storage/sdb-cluster.yaml
singlestore.kubedb.com/sdb-cluster created
```

Now, wait until `sdb-sample` has status `Ready`. i.e,

```bash
NAME                                TYPE                  VERSION   STATUS   AGE
singlestore.kubedb.com/sdb-sample   kubedb.com/v1alpha2   8.7.10    Ready    4m35s
```

> **Note:** You can manage storage autoscale for aggregator and leaf nodes separately. Here, we will focus on leaf nodes.

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo sdb-sample-leaf -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"10Gi"

$ kubectl get pv -n demo | grep 'leaf'
pvc-5cf8638e365544dd   10Gi       RWO            Retain           Bound      demo/data-sdb-sample-leaf-0         linode-block-storage-retain   <unset>                          50s
pvc-a99e7adb282a4f9c   10Gi       RWO            Retain           Bound      demo/data-sdb-sample-leaf-2         linode-block-storage-retain   <unset>                          60s
pvc-da8e9e5162a748df   10Gi       RWO            Retain           Bound      demo/data-sdb-sample-leaf-1         linode-block-storage-retain   <unset>                          70s

```

You can see the petset of leaf has 10GB storage, and the capacity of all the persistent volume is also 10GB.

We are now ready to apply the `SingleStoreAutoscaler` CRO to set up storage autoscaling for this cluster.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a SingleStoreAutoscaler Object.

#### Create SingleStoreAutoscaler Object

In order to set up vertical autoscaling for this singlestore cluster, we have to create a `SinglestoreAutoscaler` CRO with our desired configuration. Below is the YAML of the `SinglestoreAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SinglestoreAutoscaler
metadata:
  name: sdb-cluster-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: sdb-sample
  storage:
    leaf:
      trigger: "On"
      usageThreshold: 30
      scalingThreshold: 50
      expansionMode: "Online"
      upperBound: "100Gi"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `sdb-sample` cluster.
- `spec.storage.leaf.trigger` specifies that storage autoscaling is enabled for leaf nodes on this cluster.
- `spec.storage.leaf.usageThreshold` specifies storage usage threshold, if storage usage exceeds `30%` then storage autoscaling will be triggered.
- `spec.storage.leaf.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- It has another field `spec.storage.leaf.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `SinglestoreAutoscaler` CR we have shown above,


```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/autoscaling/storage/sdb-storage-autoscaler.yaml
singlestoreautoscaler.autoscaling.kubedb.com/sdb-storage-autoscaler created
```

#### Storage Autoscaling is set up successfully

Let's check that the `singlestoreautoscaler` resource is created successfully,

```bash
$ kubectl get singlestoreautoscaler -n demo
NAME                     AGE
sdb-cluster-autoscaler   2m5s


$ kubectl describe singlestoreautoscaler -n demo sdb-cluster-autoscaler
Name:         sdb-cluster-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         SinglestoreAutoscaler
Metadata:
  Creation Timestamp:  2024-09-11T07:05:11Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Singlestore
    Name:                  sdb-sample
    UID:                   e08e1f37-d869-437d-9b15-14c6aef3f406
  Resource Version:        4904325
  UID:                     471afa65-6d12-4e7d-a2a6-6d28ce440c4d
Spec:
  Database Ref:
    Name:  sdb-sample
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Leaf:
      Expansion Mode:  Online
      Scaling Rules:
        Applies Upto:     
        Threshold:        50pc
      Scaling Threshold:  50
      Trigger:            On
      Upper Bound:        100Gi
      Usage Threshold:    30
Events:                   <none>


```

So, the `singlestoreautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` creating new database with partitions 6 to see if storage autoscaling is working or not.

Let's exec into the cluster pod and fill the cluster volume using the following commands:

```bash
$ kubectl exec -it -n demo sdb-sample-leaf-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sdb-sample-leaf-0 /]$ df -h var/lib/memsql
Filesystem                                               Size  Used Avail Use% Mounted on
/dev/disk/by-id/scsi-0Linode_Volume_pvcc50e0d73d07349f9  9.8G  1.4G  8.4G  15% /var/lib/memsql

$ kubectl exec -it -n demo sdb-sample-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sdb-sample-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 113
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)
Copyright (c) 2000, 2022, Oracle and/or its affiliates.
Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.
Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
singlestore> create database demo partitions 6;
Query OK, 1 row affected (3.78 sec)

$ kubectl exec -it -n demo sdb-sample-leaf-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sdb-sample-leaf-0 /]$ df -h var/lib/memsql
Filesystem                                               Size  Used Avail Use% Mounted on
/dev/disk/by-id/scsi-0Linode_Volume_pvcc50e0d73d07349f9  9.8G  3.2G  6.7G  33% /var/lib/memsql

```

So, from the above output we can see that the storage usage is 33%, which exceeded the `usageThreshold` 30%.

Let's watch the `singlestoreopsrequest` in the demo namespace to see if any `singlestoreopsrequest` object is created. After some time you'll see that a `singlestoreopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get singlestoreopsrequest -n demo
Every 2.0s: kubectl get singlestoreopsrequest -n demo       ashraful: Wed Sep 11 13:39:25 2024

NAME                       TYPE              STATUS       AGE
sdbops-sdb-sample-th2r62   VolumeExpansion   Progressing   10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get singlestoreopsrequest -n demo
Every 2.0s: kubectl get singlestoreopsrequest -n demo       ashraful: Wed Sep 11 13:41:12 2024

NAME                       TYPE              STATUS       AGE
sdbops-sdb-sample-th2r62   VolumeExpansion   Successful   2m31s

```

We can see from the above output that the `SinglestoreOpsRequest` has succeeded. If we describe the `SinglestoreOpsRequest` we will get an overview of the steps that were followed to expand the volume of the cluster.

```bash
$ kubectl describe singlestoreopsrequest -n demo sdbops-sdb-sample-th2r62 
Name:         sdbops-sdb-sample-th2r62
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sdb-sample
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=singlestores.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SinglestoreOpsRequest
Metadata:
  Creation Timestamp:  2024-09-11T07:36:42Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  SinglestoreAutoscaler
    Name:                  sdb-cluster-autoscaler
    UID:                   471afa65-6d12-4e7d-a2a6-6d28ce440c4d
  Resource Version:        4909632
  UID:                     3dce68d0-b5ee-4ad6-bd1f-f712bae39630
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  sdb-sample
  Type:    VolumeExpansion
  Volume Expansion:
    Leaf:  15696033792
    Mode:  Online
Status:
  Conditions:
    Last Transition Time:  2024-09-11T07:36:42Z
    Message:               Singlestore ops-request has started to expand volume of singlestore nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-09-11T07:36:45Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-09-11T07:37:00Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-09-11T07:36:50Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-09-11T07:36:50Z
    Message:               delete pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetSet
    Last Transition Time:  2024-09-11T07:37:40Z
    Message:               successfully updated Leaf node PVC sizes
    Observed Generation:   1
    Reason:                UpdateLeafNodePVCs
    Status:                True
    Type:                  UpdateLeafNodePVCs
    Last Transition Time:  2024-09-11T07:37:05Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-09-11T07:37:06Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-09-11T07:37:15Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-09-11T07:37:46Z
    Message:               successfully reconciled the Singlestore resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-09-11T07:37:51Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-09-11T07:38:19Z
    Message:               Successfully completed volumeExpansion for Singlestore
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                  Age    From                         Message
  ----     ------                                  ----   ----                         -------
  Normal   Starting                                6m4s   KubeDB Ops-manager Operator  Start processing for SinglestoreOpsRequest: demo/sdbops-sdb-sample-th2r62
  Normal   Starting                                6m4s   KubeDB Ops-manager Operator  Pausing Singlestore database: demo/sdb-sample
  Normal   Successful                              6m4s   KubeDB Ops-manager Operator  Successfully paused Singlestore database: demo/sdb-sample for SinglestoreOpsRequest: sdbops-sdb-sample-th2r62
  Warning  get pet set; ConditionStatus:True       5m56s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  delete pet set; ConditionStatus:True    5m56s  KubeDB Ops-manager Operator  delete pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True       5m51s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                        5m46s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True           5m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True    5m40s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False  5m36s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True           5m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   5m31s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True    5m26s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   5m21s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True    5m16s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True           5m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True   5m11s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateLeafNodePVCs                      5m6s   KubeDB Ops-manager Operator  successfully updated Leaf node PVC sizes
  Normal   UpdatePetSets                           5m     KubeDB Ops-manager Operator  successfully reconciled the Singlestore resources
  Warning  get pet set; ConditionStatus:True       4m55s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                            4m55s  KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                4m27s  KubeDB Ops-manager Operator  Resuming Singlestore database: demo/sdb-sample
  Normal   Successful                              4m27s  KubeDB Ops-manager Operator  Successfully resumed Singlestore database: demo/sdb-sample for SinglestoreOpsRequest: sdbops-sdb-sample-th2r62
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the combined cluster has expanded to meet the desired state, Let's check,

```bash
kubectl get petset -n demo sdb-sample-leaf -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"15696033792"

~ $ kubectl get pv -n demo | grep 'leaf'
pvc-8df67f3178964106   15328158Ki   RWO            Retain           Bound      demo/data-sdb-sample-leaf-2         linode-block-storage-retain   <unset>                          42m
pvc-c50e0d73d07349f9   15328158Ki   RWO            Retain           Bound      demo/data-sdb-sample-leaf-0         linode-block-storage-retain   <unset>                          43m
pvc-f8b95ff9a9bd4fa2   15328158Ki   RWO            Retain           Bound      demo/data-sdb-sample-leaf-1         linode-block-storage-retain   <unset>                          42m

```

The above output verifies that we have successfully autoscaled the volume of the SingleStore cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete singlestoreopsrequests -n demo sdbops-sdb-sample-th2r62
kubectl delete singlestoreautoscaler -n demo sdb-storage-autoscaler
kubectl delete sdb -n demo sdb-sample
```

## Next Steps

- Detail concepts of [SingleStore object](/docs/guides/singlestore/concepts/singlestore.md).
- Different SingleStore clustering modes [here](/docs/guides/singlestore/clustering/_index.md).
- Monitor your SingleStore database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/singlestore/monitoring/prometheus-operator/index.md).
- Monitor your SingleStore database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/singlestore/monitoring/builtin-prometheus/index.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).