---
title: Elasticsearch Topology Cluster Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: es-storage-auto-scaling-topology
    name: Topology Cluster
    parent: es-storage-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of Elasticsearch Topology Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of an Elasticsearch topology cluster.

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

> **Note:** YAML files used in this tutorial are stored in this [directory](/docs/guides/elasticsearch/autoscaler/storage/topology/yamls) of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  9h
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   9h
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `Elasticsearch` topology cluster using a supported version by the `KubeDB` operator. Then we are going to apply `ElasticsearchAutoscaler` to set up autoscaling.

#### Deploy Elasticsearch Topology

In this section, we are going to deploy a Elasticsearch topology cluster with version `xpack-8.11.1`.  Then, in the next section we will set up autoscaling for this database using `ElasticsearchAutoscaler` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-topology
  namespace: demo
spec:
  enableSSL: true 
  version: xpack-8.11.1
  storageType: Durable
  topology:
    master:
      suffix: master
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      suffix: data
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      suffix: ingest
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `Elasticsearch` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/storage/topology/yamls/es-topology.yaml
elasticsearch.kubedb.com/es-topology created
```

Now, wait until `es-topology` has status `Ready`. i.e,

```bash
$ kubectl get elasticsearch -n demo -w
NAME          VERSION             STATUS         AGE
es-topology   xpack-8.11.1   Provisioning   12s
es-topology   xpack-8.11.1   Ready          1m50s
```

Let's check volume size from the data statefulset, and from the persistent volume,

```bash
$ kubectl get sts -n demo es-topology-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "1Gi"
  }
}

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS        CLAIM                            STORAGECLASS          REASON   AGE
pvc-1a22f743-2b03-487b-92db-e75ce14a3994   1Gi        RWO            Delete           Bound         demo/data-es-topology-ingest-0   topolvm-provisioner            2m8s
pvc-82c60733-22a3-4dbb-bac0-2fcd386650dd   1Gi        RWO            Delete           Bound         demo/data-es-topology-data-0     topolvm-provisioner            2m7s
pvc-a610cbb8-dece-4d2e-8870-b66a2f1fe458   1Gi        RWO            Delete           Bound         demo/data-es-topology-master-0   topolvm-provisioner            2m8s
pvc-edb7f4f7-f8ba-4af9-a507-b707462ddc3c   1Gi        RWO            Delete           Bound         demo/data-es-topology-data-1     topolvm-provisioner            119s
```

You can see that the data StatefulSet has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `ElasticsearchAutoscaler` CRO to set up storage autoscaling for the data nodes.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using an ElasticsearchAutoscaler Object.

#### Create ElasticsearchAutoscaler Object

To set up vertical autoscaling for this topology cluster, we have to create a `ElasticsearchAutoscaler` CRO with our desired configuration. Below is the YAML of the `ElasticsearchAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ElasticsearchAutoscaler
metadata:
  name: es-topology-storage-as
  namespace: demo
spec:
  databaseRef:
    name: es-topology
  storage:
    topology:
      data:
        trigger: "On"
        usageThreshold: 60
        scalingThreshold: 50
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `es-topology` cluster.
- `spec.storage.topology.data.trigger` specifies that storage autoscaling is enabled for data nodes.
- `spec.storage.topology.data.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.topology.data.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.

> Note: In this demo we are only setting up the storage autoscaling for the data nodes, that's why we only specified the data section of the autoscaler. You can enable autoscaling for master nodes and ingest nodes in the same YAML, by specifying the `topology.master` and `topology.ingest` respectivly.

Let's create the `ElasticsearchAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/autoscaler/storage/topology/yamls/es-topology-storage-as.yaml 
elasticsearchautoscaler.autoscaling.kubedb.com/es-topology-storage-as created
```

#### Storage Autoscaling is set up successfully

Let's check that the `elasticsearchautoscaler` resource is created successfully,

```bash
$ kubectl get elasticsearchautoscaler -n demo
NAME                     AGE
es-topology-storage-as   4m16s

$ kubectl describe elasticsearchautoscaler -n demo es-topology-storage-as 
Name:         es-topology-storage-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         ElasticsearchAutoscaler
Metadata:
  Creation Timestamp:  2021-03-22T15:47:18Z
  Generation:          1
  Resource Version:  19096
  UID:               3ea0516f-e272-463e-be7f-903c86a8e084
Spec:
  Database Ref:
    Name:  es-topology
  Storage:
    Topology:
      Data:
        Scaling Threshold:  50
        Trigger:            On
        Usage Threshold:    60
Events:                     <none>

```

So, the `elasticsearchautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up one of the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the data nodes and fill the database volume using the following commands:

```bash
$ kubectl exec -it -n demo es-topology-data-0 -- bash
[root@es-topology-data-0 elasticsearch]# df -h /usr/share/elasticsearch/data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/fb6d30c8-8bf7-4c19-884e-937f150f4763 1014M   40M  975M   4% /usr/share/elasticsearch/data
[root@es-topology-data-0 elasticsearch]# dd if=/dev/zero of=/usr/share/elasticsearch/data/file.img bs=650M count=1
1+0 records in
1+0 records out
681574400 bytes (682 MB) copied, 2.25556 s, 302 MB/s
[root@es-topology-data-0 elasticsearch]# df -h /usr/share/elasticsearch/data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/fb6d30c8-8bf7-4c19-884e-937f150f4763 1014M  690M  325M  69% /usr/share/elasticsearch/data

```

So, from the above output we can see that the storage usage is 69%, which exceeded the `usageThreshold` 60%.

Let's watch the `elasticsearchopsrequest` in the demo namespace to see if any `elasticsearchopsrequest` object is created. After some time you'll see that an `elasticsearchopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get esops -n demo -w
NAME                       TYPE              STATUS   AGE
esops-es-topology-79zpaf   VolumeExpansion            0s
esops-es-topology-79zpaf   VolumeExpansion   Progressing   0s
```

Let's wait for the opsRequest to become successful.

```bash
$ kubectl get esops -n demo
NAME                       TYPE              STATUS   AGE
esops-es-topology-79zpaf   VolumeExpansion   Successful    110s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe elasticsearchopsrequest -n demo esops-es-topology-79zpaf 
Name:         esops-es-topology-79zpaf
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=es-topology
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=elasticsearches.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2021-03-22T16:03:54Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ElasticsearchAutoscaler
    Name:                  es-topology-storage-as
    UID:                   aae135af-b203-47db-baeb-f51ffeb66e57
  Resource Version:        23727
  UID:                     378b28e8-9a7f-49c2-9e4d-49ee6ecad4d0
Spec:
  Database Ref:
    Name:  es-topology
  Type:    VolumeExpansion
  Volume Expansion:
    Topology:
      Data:  1594884096
Status:
  Conditions:
    Last Transition Time:  2021-03-22T16:03:54Z
    Message:               Elasticsearch ops request is expanding volume of the Elasticsearch nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2021-03-22T16:05:24Z
    Message:               successfully expanded data nodes
    Observed Generation:   1
    Reason:                UpdateDataNodePVCs
    Status:                True
    Type:                  UpdateDataNodePVCs
    Last Transition Time:  2021-03-22T16:05:39Z
    Message:               successfully deleted the statefulSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanStatefulSetPods
    Status:                True
    Type:                  OrphanStatefulSetPods
    Last Transition Time:  2021-03-22T16:05:44Z
    Message:               StatefulSet is recreated
    Observed Generation:   1
    Reason:                ReadyStatefulSets
    Status:                True
    Type:                  ReadyStatefulSets
    Last Transition Time:  2021-03-22T16:05:44Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                        Message
  ----    ------                 ----   ----                        -------
  Normal  PauseDatabase          3m18s  KubeDB Enterprise Operator  Pausing Elasticsearch demo/es-topology
  Normal  UpdateDataNodePVCs     108s   KubeDB Enterprise Operator  successfully expanded data nodes
  Normal  OrphanStatefulSetPods  93s    KubeDB Enterprise Operator  successfully deleted the statefulSets with orphan propagation policy
  Normal  ResumeDatabase         93s    KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-topology
  Normal  ResumeDatabase         93s    KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-topology
  Normal  ReadyStatefulSets      88s    KubeDB Enterprise Operator  StatefulSet is recreated
  Normal  ResumeDatabase         88s    KubeDB Enterprise Operator  Resuming Elasticsearch demo/es-topology
  Normal  Successful             88s    KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volume` whether the volume of the data nodes of the cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo es-topology-data -o json | jq '.spec.volumeClaimTemplates[].spec.resources'
{
  "requests": {
    "storage": "1594884096"
  }
}

$ kubectl get pvc -n demo
NAME                        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS          AGE
data-es-topology-data-0     Bound    pvc-82c60733-22a3-4dbb-bac0-2fcd386650dd   2Gi        RWO            topolvm-provisioner   11m
data-es-topology-data-1     Bound    pvc-edb7f4f7-f8ba-4af9-a507-b707462ddc3c   2Gi        RWO            topolvm-provisioner   11m
data-es-topology-ingest-0   Bound    pvc-1a22f743-2b03-487b-92db-e75ce14a3994   1Gi        RWO            topolvm-provisioner   11m
data-es-topology-master-0   Bound    pvc-a610cbb8-dece-4d2e-8870-b66a2f1fe458   1Gi        RWO            topolvm-provisioner   11m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS        CLAIM                            STORAGECLASS          REASON   AGE
pvc-1a22f743-2b03-487b-92db-e75ce14a3994   1Gi        RWO            Delete           Bound         demo/data-es-topology-ingest-0   topolvm-provisioner            10m
pvc-82c60733-22a3-4dbb-bac0-2fcd386650dd   2Gi        RWO            Delete           Bound         demo/data-es-topology-data-0     topolvm-provisioner            10m
pvc-a610cbb8-dece-4d2e-8870-b66a2f1fe458   1Gi        RWO            Delete           Bound         demo/data-es-topology-master-0   topolvm-provisioner            10m
pvc-edb7f4f7-f8ba-4af9-a507-b707462ddc3c   2Gi        RWO            Delete           Bound         demo/data-es-topology-data-1     topolvm-provisioner            10m
```

The above output verifies that we have successfully autoscaled the volume of the data nodes of this Elasticsearch topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete elasticsearch -n demo es-topology
$ kubectl delete elasticsearchautoscaler -n demo es-topology-storage-as
```
