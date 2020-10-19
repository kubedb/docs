---
title: Upgrading MongoDB Standalone
menu:
  docs_{{ .version }}:
    identifier: mg-upgrading-standalone
    name: Standalone
    parent: mg-upgrading
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade version of MongoDB Standalone

This guide will show you how to use `KubeDB` Enterprise operator to upgrade the version of `MongoDB` standalone.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Upgrading Overview](/docs/guides/mongodb/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare MongoDB Standalone Database

Now, we are going to deploy a `MongoDB` standalone database with version `3.6.8`.

### Deploy MongoDB standalone :

In this section, we are going to deploy a MongoDB standalone database. Then, in the next section we will upgrade the version of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/upgrading/mg-standalone.yaml
mongodb.kubedb.com/mg-standalone created
```

Now, wait until `mg-standalone` created has status `Running`. i.e,

```console
$ kubectl get mg -n demo                                                                                                                                             20:05:47
  NAME            VERSION    STATUS    AGE
  mg-standalone   3.6.8-v1   Running   8m58s
```

We are now ready to apply the `MongoDBOpsRequest` CR to upgrade this database.

### Upgrade MongoDB Version

Here, we are going to upgrade `MongoDB` standalone from `3.6.8` to `4.0.5`.

#### Create MongoDBOpsRequest:

In order to upgrade the standalone database, we have to create a `MongoDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-upgrade
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: mg-standalone
  upgrade:
    targetVersion: 4.0.5-v3
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mg-standalone` MongoDB database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `4.0.5`.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/upgrading/mops-upgrade-standalone.yaml
mongodbopsrequest.ops.kubedb.com/mops-upgrade created
```

#### Verify MongoDB version upgraded successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME           TYPE      STATUS       AGE
mops-upgrade   Upgrade   Successful   3m45s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to upgrade the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-upgrade
  Name:         mops-upgrade
  Namespace:    demo
  Labels:       <none>
  Annotations:  API Version:  ops.kubedb.com/v1alpha1
  Kind:         MongoDBOpsRequest
  Metadata:
    Creation Timestamp:  2020-08-24T14:22:10Z
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
      Time:         2020-08-24T14:22:10Z
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
      Time:            2020-08-24T14:22:26Z
    Resource Version:  4786082
    Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-upgrade
    UID:               c5f35015-75b7-4843-8f05-9181d8bf14a5
  Spec:
    Database Ref:
      Name:  mg-standalone
    Type:    Upgrade
    Upgrade:
      Target Version:  4.0.5-v3
  Status:
    Conditions:
      Last Transition Time:  2020-08-24T14:22:10Z
      Message:               MongoDB ops request is starting to process
      Observed Generation:   1
      Reason:                UpgradingVersion
      Status:                True
      Type:                  UpgradingVersion
      Last Transition Time:  2020-08-24T14:22:11Z
      Message:               Successfully paused mongodb: mg-standalone
      Observed Generation:   1
      Reason:                PauseDatabase
      Status:                True
      Type:                  PauseDatabase
      Last Transition Time:  2020-08-24T14:22:11Z
      Message:               Successfully updated statefulsets update strategy type
      Observed Generation:   1
      Reason:                UpdateStatefulSets
      Status:                True
      Type:                  UpdateStatefulSets
      Last Transition Time:  2020-08-24T14:22:26Z
      Message:               Successfully updated ReplicaSet images
      Observed Generation:   1
      Reason:                UpdateReplicaSetImage
      Status:                True
      Type:                  UpdateReplicaSetImage
      Last Transition Time:  2020-08-24T14:22:26Z
      Message:               Succefully Resumed mongodb: mg-standalone
      Observed Generation:   1
      Reason:                ResumeDatabase
      Status:                True
      Type:                  ResumeDatabase
      Last Transition Time:  2020-08-24T14:22:26Z
      Message:               Successfully completed the modification process.
      Observed Generation:   1
      Reason:                Successful
      Status:                True
      Type:                  Successful
    Observed Generation:     1
    Phase:                   Successful
  Events:
    Type    Reason                 Age   From                        Message
    ----    ------                 ----  ----                        -------
    Normal  PauseDatabase          10m   KubeDB Enterprise Operator  Pausing Mongodb mg-standalone in Namespace demo
    Normal  PauseDatabase          10m   KubeDB Enterprise Operator  Successfully Paused Mongodb mg-standalone in Namespace demo
    Normal  Updating               10m   KubeDB Enterprise Operator  Updating StatefulSets
    Normal  Updating               10m   KubeDB Enterprise Operator  Successfully Updated StatefulSets
    Normal  UpdateReplicaSetImage  10m   KubeDB Enterprise Operator  Updating ReplicaSet Images
    Normal  UpdateReplicaSetImage  10m   KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-standalone-0
    Normal  UpdateReplicaSetImage  10m   KubeDB Enterprise Operator  Successfully Updated ReplicaSet Images
    Normal  ResumeDatabase         10m   KubeDB Enterprise Operator  Resuming MongoDB
    Normal  ResumeDatabase         10m   KubeDB Enterprise Operator  Successfully Started Balancer
    Normal  Successful             10m   KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify whether the `MongoDB` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```console
$ kubectl get mg -n demo mg-standalone -o=jsonpath='{.spec.version}{"\n"}'                                                                                          
  4.0.5-v3

$ kubectl get sts -n demo mg-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                               
  kubedb/mongo:4.0.5-v3

$ kubectl get pods -n demo mg-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           
  kubedb/mongo:4.0.5-v3
```

You can see from above, our `MongoDB` standalone database has been updated with the new version. So, the upgrade process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-standalone
kubectl delete mongodbopsrequest -n demo mops-upgrade
```