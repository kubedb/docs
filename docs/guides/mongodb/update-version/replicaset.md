---
title: Updating MongoDB Replicaset
menu:
  docs_{{ .version }}:
    identifier: mg-updating-replicaset
    name: ReplicaSet
    parent: mg-updating
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of MongoDB ReplicaSet

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `MongoDB` ReplicaSet.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/mongodb/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare MongoDB ReplicaSet Database

Now, we are going to deploy a `MongoDB` replicaset database with version `3.6.8`.

### Deploy MongoDB replicaset

In this section, we are going to deploy a MongoDB replicaset database. Then, in the next section we will update the version of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/update-version/mg-replicaset.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` created has status `Ready`. i.e,

```bash
$ k get mongodb -n demo                                                                                                                                             
NAME            VERSION    STATUS    AGE
mg-replicaset   4.4.26   Ready     109s
```

We are now ready to apply the `MongoDBOpsRequest` CR to update this database.

### update MongoDB Version

Here, we are going to update `MongoDB` replicaset from `3.6.8` to `4.0.5`.

#### Create MongoDBOpsRequest:

In order to update the version of the replicaset database, we have to create a `MongoDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-replicaset-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mg-replicaset
  updateVersion:
    targetVersion: 4.4.26
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mg-replicaset` MongoDB database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `4.0.5`.
- Have a look [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/update-version/mops-update-replicaset .yaml
mongodbopsrequest.ops.kubedb.com/mops-replicaset-update created
```

#### Verify MongoDB version updated successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the image of `MongoDB` object and related `PetSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                      TYPE            STATUS       AGE
mops-replicaset-update   UpdateVersion   Successful   84s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-replicaset-update
Name:         mops-replicaset-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2022-10-26T10:19:55Z
  Generation:          1
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
        f:apply:
        f:databaseRef:
        f:readinessCriteria:
          .:
          f:objectsCountDiffPercentage:
          f:oplogMaxLagSeconds:
        f:timeout:
        f:type:
        f:updateVersion:
          .:
          f:targetVersion:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-10-26T10:19:55Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2022-10-26T10:23:09Z
  Resource Version:  607814
  UID:               38053605-47bd-4d94-9f53-ce9474ad0a98
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mg-replicaset
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:                             UpdateVersion
  UpdateVersion:
    Target Version:  4.4.26
Status:
  Conditions:
    Last Transition Time:  2022-10-26T10:21:20Z
    Message:               MongoDB ops request is update-version database version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2022-10-26T10:21:39Z
    Message:               Successfully updated petsets update strategy type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2022-10-26T10:23:09Z
    Message:               Successfully Updated Standalone Image
    Observed Generation:   1
    Reason:                UpdateStandaloneImage
    Status:                True
    Type:                  UpdateStandaloneImage
    Last Transition Time:  2022-10-26T10:23:09Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                         Message
  ----    ------                 ----   ----                         -------
  Normal  PauseDatabase          2m27s  KubeDB Ops-manager Operator  Pausing MongoDB demo/mg-replicaset
  Normal  PauseDatabase          2m27s  KubeDB Ops-manager Operator  Successfully paused MongoDB demo/mg-replicaset
  Normal  Updating               2m27s  KubeDB Ops-manager Operator  Updating PetSets
  Normal  Updating               2m8s   KubeDB Ops-manager Operator  Successfully Updated PetSets
  Normal  UpdateStandaloneImage  38s    KubeDB Ops-manager Operator  Successfully Updated Standalone Image
  Normal  ResumeDatabase         38s    KubeDB Ops-manager Operator  Resuming MongoDB demo/mg-replicaset
  Normal  ResumeDatabase         38s    KubeDB Ops-manager Operator  Successfully resumed MongoDB demo/mg-replicaset
  Normal  Successful             38s    KubeDB Ops-manager Operator  Successfully Updated Database
```

Now, we are going to verify whether the `MongoDB` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get mg -n demo mg-replicaset -o=jsonpath='{.spec.version}{"\n"}'
4.4.26

$ kubectl get sts -n demo mg-replicaset -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mongo:4.0.5

$ kubectl get pods -n demo mg-replicaset-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mongo:4.0.5
```

You can see from above, our `MongoDB` replicaset database has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-replicaset-update
```