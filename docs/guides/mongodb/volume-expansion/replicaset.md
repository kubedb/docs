---
title: MongoDB Replicaset Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: mg-volume-expansion-replicaset
    name: Replicaset
    parent: mg-volume-expansion
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Replicaset Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a MongoDB Replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/mongodb/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Replicaset

Here, we are going to deploy a  `MongoDB` replicaset using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBOpsRequest` to expand its volume.

### Prepare MongoDB Replicaset Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `MongoDB` replicaSet database with version `4.4.26`.

### Deploy MongoDB

In this section, we are going to deploy a MongoDB Replicaset database with 1GB volume. Then, in the next section we will expand its volume to 2GB using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet: 
    name: "replicaset"
  replicas: 3
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/volume-expansion/mg-replicaset.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME            VERSION    STATUS    AGE
mg-replicaset   4.4.26      Ready     10m
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo                                                                                          
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-2067c63d-f982-4b66-a008-5e9c3ff6218a   1Gi        RWO            Delete           Bound    demo/datadir-mg-replicaset-0   standard                10m
pvc-9db1aeb0-f1af-4555-93a3-0ca754327751   1Gi        RWO            Delete           Bound    demo/datadir-mg-replicaset-2   standard                9m45s
pvc-d38f42a8-50d4-4fa9-82ba-69fc7a464ff4   1Gi        RWO            Delete           Bound    demo/datadir-mg-replicaset-1   standard                10m
```

You can see the petset has 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `MongoDBOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the replicaset database.

#### Create MongoDBOpsRequest

In order to expand the volume of the database, we have to create a `MongoDBOpsRequest` CR with our desired volume size. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-volume-exp-replicaset
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: mg-replicaset
  volumeExpansion:
    replicaSet: 2Gi
    mode: Online
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `mops-volume-exp-replicaset` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.replicaSet` specifies the desired volume size.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/volume-expansion/mops-volume-exp-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-volume-exp-replicaset created
```

#### Verify MongoDB replicaset volume expanded successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `MongoDB` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
NAME                         TYPE              STATUS       AGE
mops-volume-exp-replicaset   VolumeExpansion   Successful   83s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-volume-exp-replicaset   
Name:         mops-volume-exp-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2020-08-25T18:21:18Z
  Finalizers:
    kubedb.com
  Generation:        1
  Resource Version:  84084
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-volume-exp-replicaset
  UID:               2cec0cd3-4abe-4114-813c-1326f28563cb
Spec:
  Database Ref:
    Name:  mg-replicaset
  Type:    VolumeExpansion
  Volume Expansion:
    ReplicaSet:  2Gi
Status:
  Conditions:
    Last Transition Time:  2020-08-25T18:21:18Z
    Message:               MongoDB ops request is being processed
    Observed Generation:   1
    Reason:                Scaling
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2020-08-25T18:22:38Z
    Message:               Successfully updated Storage
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2020-08-25T18:22:38Z
    Message:               Successfully Resumed mongodb: mg-replicaset
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-25T18:22:38Z
    Message:               Successfully completed the modification process
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason           Age    From                        Message
  ----    ------           ----   ----                        -------
  Normal  VolumeExpansion  3m11s  KubeDB Ops-manager operator  Successfully Updated Storage
  Normal  ResumeDatabase   3m11s  KubeDB Ops-manager operator  Resuming MongoDB
  Normal  ResumeDatabase   3m11s  KubeDB Ops-manager operator  Successfully Resumed mongodb
  Normal  Successful       3m11s  KubeDB Ops-manager operator  Successfully Scaled Database  
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo                                                                                          
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-2067c63d-f982-4b66-a008-5e9c3ff6218a   2Gi        RWO            Delete           Bound    demo/datadir-mg-replicaset-0   standard                19m
pvc-9db1aeb0-f1af-4555-93a3-0ca754327751   2Gi        RWO            Delete           Bound    demo/datadir-mg-replicaset-2   standard                18m
pvc-d38f42a8-50d4-4fa9-82ba-69fc7a464ff4   2Gi        RWO            Delete           Bound    demo/datadir-mg-replicaset-1   standard                19m
```

The above output verifies that we have successfully expanded the volume of the MongoDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-volume-exp-replicaset
```