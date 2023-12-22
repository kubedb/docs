---
title: Redis Autoscaling
menu:
  docs_{{ .version }}:
    identifier: rd-storage-auto-scaling-standalone
    name: Redis Autoscaling
    parent: rd-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Redis Standalone Database

This guide will show you how to use `KubeDB` to autoscale the storage of a Redis standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.
  
- You should be familiar with the following `KubeDB` concepts:
    - [Redis](/docs/guides/redis/concepts/redis.md)
    - [RedisAutoscaler](/docs/guides/redis/concepts/autoscaler.md)
    - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
    - [Storage Autoscaling Overview](/docs/guides/redis/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of Standalone Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  9h
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   9h
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `Redis` standalone using a supported version by `KubeDB` operator. Then we are going to apply `RedisAutoscaler` to set up autoscaling.

#### Deploy Redis standalone

> If you want to autoscale Redis in `Cluster` or `Sentinel` mode, just deploy a Redis database in respective Mode and rest of the steps are same.


In this section, we are going to deploy a Redis standalone database with version `6.2.14`.  Then, in the next section we will set up autoscaling for this database using `RedisAutoscaler` CRD. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-standalone
  namespace: demo
spec:
  version: "6.2.14"
  storageType: Durable
  storage:
    storageClassName: topolvm-provisioner
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `Redis` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/autoscaling/storage/rd-standalone.yaml
redis.kubedb.com/rd-standalone created
```

Now, wait until `rd-standalone` has status `Ready`. i.e,

```bash
$ kubectl get rd -n demo
NAME            VERSION    STATUS    AGE
rd-standalone   6.2.14      Ready     2m53s
```

Let's check volume size from statefulset, and from the persistent volume,

```bash
$ kubectl get sts -n demo rd-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS          REASON   AGE
pvc-cf469ed8-a89a-49ca-bf7c-8c76b7889428   1Gi        RWO            Delete           Bound    demo/datadir-rd-standalone-0   topolvm-provisioner            7m41s
```

You can see the statefulset has 1GB storage, and the capacity of the persistent volume is also 1GB.

We are now ready to apply the `RedisAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a RedisAutoscaler Object.

#### Create RedisAutoscaler Object

In order to set up vertical autoscaling for this standalone database, we have to create a `RedisAutoscaler` CRO with our desired configuration. Below is the YAML of the `RedisAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RedisAutoscaler
metadata:
  name: rd-as
  namespace: demo
spec:
  databaseRef:
    name: rd-standalone
  storage:
    standalone:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

> If you want to autoscale Redis in Cluster mode, the field in `spec.storage` should be `cluster` and for sentinel it should be `sentinel`. The subfields are same inside `spec.storage.standalone`, `spec.storage.cluster` and `spec.storage.sentinel`


Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `rd-standalone` database.
- `spec.storage.standalone.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.standalone.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.standalone.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- It has another field `spec.storage.replicaSet.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `RedisAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/autoscaling/storage/rd-as.yaml
redisautoscaler.autoscaling.kubedb.com/rd-as created
```

#### Storage Autoscaling is set up successfully

Let's check that the `redisautoscaler` resource is created successfully,

```bash
$ kubectl get redisautoscaler -n demo
NAME    AGE
rd-as   102s

$ kubectl describe redisautoscaler rd-as -n demo
Name:         rd-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         RedisAutoscaler
Metadata:
  Creation Timestamp:  2023-02-09T11:02:26Z
  Generation:          1
  Managed Fields:
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:storage:
          .:
          f:standalone:
            .:
            f:scalingThreshold:
            f:trigger:
            f:usageThreshold:
    Manager:         kubectl-client-side-apply
    Operation:       Update
    Time:            2023-02-09T11:02:26Z
  Resource Version:  134423
  Self Link:         /apis/autoscaling.kubedb.com/v1alpha1/namespaces/demo/redisautoscalers/rd-as
  UID:               999a2dc9-7eb7-4ed2-9e90-d3f8b21c091a
Spec:
  Database Ref:
    Name:  rd-standalone
  Storage:
    Standalone:
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```
So, the `redisautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Lets exec into the database pod and fill the database volume using the following commands:

```bash
$ kubectl exec -it -n demo rd-standalone-0 -- bash
root@rd-standalone-0:/# df -h /data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/1df4ee9e-b900-4c0f-9d2c-8493fb30bdc0 1014M  334M  681M  33% /data/db
root@rd-standalone-0:/# dd if=/dev/zero of=/data/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 0.359202 s, 1.5 GB/s
root@rd-standalone-0:/# df -h /data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/1df4ee9e-b900-4c0f-9d2c-8493fb30bdc0 1014M  835M  180M  83% /data/db
```

So, from the above output we can see that the storage usage is 84%, which exceeded the `usageThreshold` 60%.

Let's watch the `redisopsrequest` in the demo namespace to see if any `redisopsrequest` object is created. After some time you'll see that a `redisopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                         TYPE              STATUS        AGE
rdops-rd-standalone-p27c11   VolumeExpansion   Progressing   26s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                         TYPE              STATUS        AGE
rdops-rd-standalone-p27c11   VolumeExpansion   Successful    73s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 

Now, we are going to verify from the `Statefulset`, and the `Persistent Volume` whether the volume of the standalone database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo rd-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS          REASON   AGE
pvc-cf469ed8-a89a-49ca-bf7c-8c76b7889428   2Gi        RWO            Delete           Bound    demo/datadir-rd-standalone-0   topolvm-provisioner            26m
```

The above output verifies that we have successfully autoscaled the volume of the Redis standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo rd/rd-standalone -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-standalone patched

$ kubectl delete rd -n demo rd-standalone
redis.kubedb.com "rd-standalone" deleted

$ kubectl delete redisautoscaler -n demo rd-as
redisautoscaler.autoscaling.kubedb.com "rd-as" deleted
```
