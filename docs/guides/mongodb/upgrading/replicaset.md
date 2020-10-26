---
title: Upgrading MongoDB Replicaset
menu:
  docs_{{ .version }}:
    identifier: mg-upgrading-replicaset
    name: ReplicaSet
    parent: mg-upgrading
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade version of MongoDB ReplicaSet

This guide will show you how to use `KubeDB` Enterprise operator to upgrade the version of `MongoDB` ReplicaSet.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Upgrading Overview](/docs/guides/mongodb/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare MongoDB ReplicaSet Database

Now, we are going to deploy a `MongoDB` replicaset database with version `3.6.8`.

### Deploy MongoDB replicaset

In this section, we are going to deploy a MongoDB replicaset database. Then, in the next section we will upgrade the version of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "3.6.8-v1"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/upgrading/mg-replicaset.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` created has status `Running`. i.e,

```bash
$ k get mongodb -n demo                                                                                                                                             
  NAME            VERSION    STATUS    AGE
  mg-replicaset   3.6.8-v1   Running   109s
```

We are now ready to apply the `MongoDBOpsRequest` CR to upgrade this database.

### Upgrade MongoDB Version

Here, we are going to upgrade `MongoDB` replicaset from `3.6.8` to `4.0.5`.

#### Create MongoDBOpsRequest:

In order to upgrade the replicaset database, we have to create a `MongoDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-replicaset-upgrade
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: mg-replicaset
  upgrade:
    targetVersion: 4.0.5-v3
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mg-replicaset` MongoDB database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `4.0.5`.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/upgrading/mops-upgrade-replicaset .yaml
mongodbopsrequest.ops.kubedb.com/mops-replicaset-upgrade created
```

#### Verify MongoDB version upgraded successfully 

If everything goes well, `KubeDB` Enterprise operator will update the image of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                      TYPE      STATUS       AGE
mops-replicaset-upgrade   Upgrade   Successful   84s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to upgrade the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-replicaset-upgrade
  Name:         mops-replicaset-upgrade
  Namespace:    demo
  Labels:       <none>
  Annotations:  API Version:  ops.kubedb.com/v1alpha1
  Kind:         MongoDBOpsRequest
  Metadata:
    Creation Timestamp:  2020-08-24T14:56:39Z
    Finalizers:
      kubedb.com
    Generation:  1
    Managed Fields:
      API Version:  ops.kubedb.com/v1alpha1
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
          f:type:
          f:upgrade:
            .:
            f:targetVersion:
      Manager:      kubectl
      Operation:    Update
      Time:         2020-08-24T14:56:39Z
      API Version:  ops.kubedb.com/v1alpha1
      Fields Type:  FieldsV1
      fieldsV1:
        f:metadata:
          f:finalizers:
        f:status:
          .:
          f:conditions:
          f:observedGeneration:
          f:phase:
      Manager:         kubedb-enterprise
      Operation:       Update
      Time:            2020-08-24T14:57:14Z
    Resource Version:  4812837
    Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-replicaset-upgrade
    UID:               99fc9dd2-11f7-44d6-9432-e93b4028ef7b
  Spec:
    Database Ref:
      Name:  mg-replicaset
    Type:    Upgrade
    Upgrade:
      Target Version:  4.0.5-v3
  Status:
    Conditions:
      Last Transition Time:  2020-08-24T14:56:39Z
      Message:               MongoDB ops request is starting to process
      Observed Generation:   1
      Reason:                UpgradingVersion
      Status:                True
      Type:                  UpgradingVersion
      Last Transition Time:  2020-08-24T14:56:39Z
      Message:               Successfully halted mongodb: mg-replicaset
      Observed Generation:   1
      Reason:                HaltDatabase
      Status:                True
      Type:                  HaltDatabase
      Last Transition Time:  2020-08-24T14:56:39Z
      Message:               Successfully updated statefulsets update strategy type
      Observed Generation:   1
      Reason:                UpdateStatefulSets
      Status:                True
      Type:                  UpdateStatefulSets
      Last Transition Time:  2020-08-24T14:57:14Z
      Message:               Successfully updated ReplicaSet images
      Observed Generation:   1
      Reason:                UpdateReplicaSetImage
      Status:                True
      Type:                  UpdateReplicaSetImage
      Last Transition Time:  2020-08-24T14:57:14Z
      Message:               Succefully Resumed mongodb: mg-replicaset
      Observed Generation:   1
      Reason:                ResumeDatabase
      Status:                True
      Type:                  ResumeDatabase
      Last Transition Time:  2020-08-24T14:57:14Z
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
    Normal  HaltDatabase          2m26s  KubeDB Enterprise Operator  Pausing Mongodb mg-replicaset in Namespace demo
    Normal  HaltDatabase          2m26s  KubeDB Enterprise Operator  Successfully Halted Mongodb mg-replicaset in Namespace demo
    Normal  Updating               2m26s  KubeDB Enterprise Operator  Updating StatefulSets
    Normal  Updating               2m26s  KubeDB Enterprise Operator  Successfully Updated StatefulSets
    Normal  UpdateReplicaSetImage  2m26s  KubeDB Enterprise Operator  Updating ReplicaSet Images
    Normal  UpdateReplicaSetImage  2m11s  KubeDB Enterprise Operator  Successfully Updated Images of Pod mg-replicaset-1
    Normal  UpdateReplicaSetImage  2m6s   KubeDB Enterprise Operator  Successfully Updated Images of Pod mg-replicaset-2
    Normal  UpdateReplicaSetImage  111s   KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-replicaset-0
    Normal  UpdateReplicaSetImage  111s   KubeDB Enterprise Operator  Successfully Updated ReplicaSet Images
    Normal  ResumeDatabase         111s   KubeDB Enterprise Operator  Resuming MongoDB
    Normal  ResumeDatabase         111s   KubeDB Enterprise Operator  Successfully Started Balancer
    Normal  Successful             111s   KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify whether the `MongoDB` and the related `StatefulSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get mg -n demo mg-replicaset -o=jsonpath='{.spec.version}{"\n"}'                                                                                           20:59:43
  4.0.5-v3

$ kubectl get sts -n demo mg-replicaset -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                                21:01:32
  kubedb/mongo:4.0.5-v3

$ kubectl get pods -n demo mg-replicaset-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           21:01:57
  kubedb/mongo:4.0.5-v3
```

You can see from above, our `MongoDB` replicaset database has been updated with the new version. So, the upgrade process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-replicaset-upgrade
```