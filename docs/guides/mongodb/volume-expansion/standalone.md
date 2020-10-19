---
title: MongoDB Standalone Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: mg-volume-expansion-standalone
    name: Standalone
    parent: mg-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# MongoDB Standalone Volume Expansion

This guide will show you how to use `KubeDB` Enterprise operator to expand the volume of a MongoDB standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Volume Expansion Overview](/docs/guides/mongodb/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Expand Volume of Standalone Database

Here, we are going to deploy a `MongoDB` standalone using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBOpsRequest` to expand its volume.

### Prepare MongoDB Standalone Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```console
$ kubectl get storageclass                                                                                                                                           20:22:33
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `MongoDB` standalone database with version `3.6.8`.

#### Deploy MongoDB standalone

In this section, we are going to deploy a MongoDB standalone database with 1GB volume. Then, in the next section we will expand its volume to 2GB using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-standalone
  namespace: demo
spec:
  version: "3.6.8-v1"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `MongoDB` CR we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/volume-expansion/mg-standalone.yaml
mongodb.kubedb.com/mg-standalone created
```

Now, wait until `mg-standalone` has status `Running`. i.e,

```console
$ kubectl get mg -n demo                                                                                                                                             20:05:47
  NAME            VERSION    STATUS    AGE
  mg-standalone   3.6.8-v1   Running   2m53s
```

Let's check volume size from statefulset, and from the persistent volume,

```console
$ kubectl get sts -n demo mg-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-d0b07657-a012-4384-862a-b4e437774287   1Gi        RWO            Delete           Bound    demo/datadir-mg-standalone-0   standard                49s
```

You can see the statefulset has 1GB storage, and the capacity of the persistent volume is also 1GB.

We are now ready to apply the `MongoDBOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the standalone database.

#### Create MongoDBOpsRequest

In order to expand the volume of the database, we have to create a `MongoDBOpsRequest` CR with our desired volume size. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-volume-exp-standalone
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: mg-standalone
  volumeExpansion:
    standalone: 2Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `mops-volume-exp-standalone` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.standalone` specifies the desired volume size.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/volume-expansion/mops-volume-exp-standalone.yaml
mongodbopsrequest.ops.kubedb.com/mops-volume-exp-standalone created
```

#### Verify MongoDB Standalone volume expanded successfully

If everything goes well, `KubeDB` Enterprise operator will update the volume size of `MongoDB` object and related `StatefulSets` and `Persistent Volume`.

Let's wait for `MongoDBOpsRequest` to be `Successful`. Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-volume-exp-standalone                                                                                              23:50:28
  Name:         mops-volume-exp-standalone
  Namespace:    demo
  Labels:       <none>
  Annotations:  API Version:  ops.kubedb.com/v1alpha1
  Kind:         MongoDBOpsRequest
  Metadata:
    Creation Timestamp:  2020-08-25T17:48:33Z
    Finalizers:
      kubedb.com
    Generation:        1
    Resource Version:  72899
    Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-volume-exp-standalone
    UID:               007fe35a-25f6-45e7-9e85-9add488b2622
  Spec:
    Database Ref:
      Name:  mg-standalone
    Type:    VolumeExpansion
    Volume Expansion:
      Standalone:  2Gi
  Status:
    Conditions:
      Last Transition Time:  2020-08-25T17:48:33Z
      Message:               MongoDB ops request is being processed
      Observed Generation:   1
      Reason:                Scaling
      Status:                True
      Type:                  Scaling
      Last Transition Time:  2020-08-25T17:50:03Z
      Message:               Successfully updated Storage
      Observed Generation:   1
      Reason:                VolumeExpansion
      Status:                True
      Type:                  VolumeExpansion
      Last Transition Time:  2020-08-25T17:50:03Z
      Message:               Successfully Resumed mongodb: mg-standalone
      Observed Generation:   1
      Reason:                ResumeDatabase
      Status:                True
      Type:                  ResumeDatabase
      Last Transition Time:  2020-08-25T17:50:03Z
      Message:               Successfully completed the modification process
      Observed Generation:   1
      Reason:                Successful
      Status:                True
      Type:                  Successful
    Observed Generation:     1
    Phase:                   Successful
  Events:
    Type    Reason           Age   From                        Message
    ----    ------           ----  ----                        -------
    Normal  VolumeExpansion  29s   KubeDB Enterprise Operator  Successfully Updated Storage
    Normal  ResumeDatabase   29s   KubeDB Enterprise Operator  Resuming MongoDB
    Normal  ResumeDatabase   29s   KubeDB Enterprise Operator  Successfully Resumed mongodb
    Normal  Successful       29s   KubeDB Enterprise Operator  Successfully Scaled Database
```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volume` whether the volume of the standalone database has expanded to meet the desired state, Let's check,

```console
$ kubectl get sts -n demo mg-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-d0b07657-a012-4384-862a-b4e437774287   2Gi        RWO            Delete           Bound    demo/datadir-mg-standalone-0   standard                4m29s
```

The above output verifies that we have successfully expanded the volume of the MongoDB standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-standalone
kubectl delete mongodbopsrequest -n demo mops-volume-exp-standalone
```
