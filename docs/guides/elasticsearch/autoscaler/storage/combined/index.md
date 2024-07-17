---
title: Elasticsearch Combined Cluster Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: es-storage-auto-scaling-combined
    name: Combined Cluster
    parent: es-storage-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of Elasticsearch Combined Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of an Elasticsearch combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.
  
- You should be familiar with the following `KubeDB` concepts:
  - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
  - [ElasticsearchAutoscaler](/docs/guides/elasticsearch/concepts/autoscaler/index.md)
  - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
  - [Storage Autoscaling Overview](/docs/guides/elasticsearch/autoscaler/storage/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in this [directory](/docs/guides/elasticsearch/autoscaler/storage/combined/yamls) of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of Combined cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  9h
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   9h
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `Elasticsearch` combined cluster using a supported version by the `KubeDB` operator. Then we are going to apply `ElasticsearchAutoscaler` to set up autoscaling.

#### Deploy Elasticsearch Combined Cluster

In this section, we are going to deploy an Elasticsearch combined cluster with version `xpack-8.11.1`.  Then, in the next section we will set up autoscaling for this database using `ElasticsearchAutoscaler` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-combined
  namespace: demo
spec:
  enableSSL: true 
  version: xpack-8.11.1
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "topolvm-provisioner"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Elasticsearch` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/storage/combined/yamls/es-combined.yaml
elasticsearch.kubedb.com/es-combined created
```

Now, wait until `es-combined` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo -w
NAME          VERSION             STATUS         AGE
es-combined   xpack-8.11.1   Provisioning   5s
es-combined   xpack-8.11.1   Ready          50s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo es-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "1Gi"
  }
}

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                     STORAGECLASS          REASON   AGE
pvc-efe67aee-21bf-4320-9873-5d58d68182ae   1Gi        RWO            Delete           Bound    demo/data-es-combined-0   topolvm-provisioner            8m3s
```

You can see the PetSet has 1GB storage, and the capacity of the persistent volume is also 1GB.

We are now ready to apply the `ElasticsearchAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a ElasticsearchAutoscaler Object.

#### Create ElasticsearchAutoscaler Object

To set up vertical autoscaling for the combined cluster nodes, we have to create a `ElasticsearchAutoscaler` CRO with our desired configuration. Below is the YAML of the `ElasticsearchAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ElasticsearchAutoscaler
metadata:
  name: es-combined-storage-as
  namespace: demo
spec:
  databaseRef:
    name: es-combined
  storage:
    node:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `es-combined` cluster.
- `spec.storage.node.trigger` specifies that storage autoscaling is enabled for the Elasticsearch nodes.
- `spec.storage.node.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.node.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.

Let's create the `ElasticsearchAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/storage/combined/yamls/es-combined-storage-as.yaml 
elasticsearchautoscaler.autoscaling.kubedb.com/es-combined-storage-as created
```

#### Storage Autoscaling is set up successfully

Let's check that the `elasticsearchautoscaler` resource is created successfully,

```bash
$ kubectl get elasticsearchautoscaler -n demo
NAME                     AGE
es-combined-storage-as   9s

$ kubectl describe elasticsearchautoscaler -n demo es-combined-storage-as 
Name:         es-combined-storage-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         ElasticsearchAutoscaler
Metadata:
  Creation Timestamp:  2021-03-22T14:57:58Z
  Generation:          1
  Resource Version:  7906
  UID:               f4e6b550-b566-458b-af05-84e0581b93f0
Spec:
  Database Ref:
    Name:  es-combined
  Storage:
    Node:
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```

So, the `elasticsearchautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume using the following commands:

```bash
$ kubectl exec -it -n demo es-combined-0 -- bash
[root@es-combined-0 elasticsearch]# df -h /usr/share/elasticsearch/data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/026b4152-c7d8-47c1-afe2-0a7c7b708857 1014M   40M  975M   4% /usr/share/elasticsearch/data

[root@es-combined-0 elasticsearch]# dd if=/dev/zero of=/usr/share/elasticsearch/data/file.img bs=600M count=1
1+0 records in
1+0 records out
629145600 bytes (629 MB) copied, 1.95767 s, 321 MB/s

[root@es-combined-0 elasticsearch]# df -h /usr/share/elasticsearch/data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/026b4152-c7d8-47c1-afe2-0a7c7b708857 1014M  640M  375M  64% /usr/share/elasticsearch/data
```

So, from the above output, we can see that the storage usage is 64%, which exceeded the `usageThreshold` 60%.

Let's watch the `elasticsearchopsrequest` in the demo namespace to see if any `elasticsearchopsrequest` object is created. After some time you'll see that a `elasticsearchopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$  kubectl get esops -n demo -w
NAME                       TYPE              STATUS   AGE
esops-es-combined-8ub9ca   VolumeExpansion   Progressing   30s
```

Let's wait for the opsRequest to become successful.

```bash
$ kubectl get esops -n demo
NAME                       TYPE              STATUS        AGE
esops-es-combined-8ub9ca   VolumeExpansion   Successful    50s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe esops -n demo esops-es-combined-8ub9ca 
Name:         esops-es-combined-8ub9ca
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=es-combined
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=elasticsearches.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2021-03-22T15:08:54Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-combined-storage-as
    UID:                   f4e6b550-b566-458b-af05-84e0581b93f0
  Resource Version:        11064
  UID:                     65ca8078-ae75-4b90-8e11-c09dc287c993
Spec:
  Database Ref:
    Name:  es-combined
  Type:    VolumeExpansion
  Volume Expansion:
    Node:  1594884096
Status:
  Conditions:
    Last Transition Time:  2021-03-22T15:08:54Z
    Message:               Elasticsearch ops request is expanding volume of the Elasticsearch nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2021-03-22T15:09:24Z
    Message:               successfully expanded combined nodes
    Observed Generation:   1
    Reason:                UpdateCombinedNodePVCs
    Status:                True
    Type:                  UpdateCombinedNodePVCs
    Last Transition Time:  2021-03-22T15:09:39Z
    Message:               successfully deleted the petSet with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2021-03-22T15:09:44Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2021-03-22T15:09:44Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                  Age   From                        Message
  ----    ------                  ----  ----                        -------
  Normal  PauseDatabase           17m   KubeDB Enterprise Operator  Pausing Elasticsearch demo/es-combined
  Normal  UpdateCombinedNodePVCs  17m   KubeDB Enterprise Operator  successfully expanded combined nodes
  Normal  OrphanPetSetPods   16m   KubeDB Enterprise Operator  successfully deleted the petSet with orphan propagation policy
  Normal  ResumeDatabase          16m   KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-combined
  Normal  ResumeDatabase          16m   KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-combined
  Normal  ReadyPetSets       16m   KubeDB Enterprise Operator  PetSet is recreated
  Normal  ResumeDatabase          16m   KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-combined
  Normal  Successful              16m   KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the combined cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo es-combined -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "1594884096"
  }
}

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS        CLAIM                     STORAGECLASS          REASON   AGE
pvc-efe67aee-21bf-4320-9873-5d58d68182ae   2Gi        RWO            Delete           Bound         demo/data-es-combined-0   topolvm-provisioner            43m
```

The above output verifies that we have successfully autoscaled the volume of the Elasticsearch combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete elasticsearch -n demo es-combined
$ kubectl delete elasticsearchautoscaler -n demo es-combined-storage-as
```
