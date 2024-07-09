---
title: Updating MongoDB Standalone
menu:
  docs_{{ .version }}:
    identifier: mg-updating-standalone
    name: Standalone
    parent: mg-updating
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of MongoDB Standalone

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `MongoDB` standalone.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/mongodb/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare MongoDB Standalone Database

Now, we are going to deploy a `MongoDB` standalone database with version `3.6.8`.

### Deploy MongoDB standalone :

In this section, we are going to deploy a MongoDB standalone database. Then, in the next section we will update the version of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-standalone
  namespace: demo
spec:
  version: "4.4.26"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/update-version/mg-standalone.yaml
mongodb.kubedb.com/mg-standalone created
```

Now, wait until `mg-standalone` created has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
  NAME            VERSION    STATUS    AGE
  mg-standalone   4.4.26   Ready     8m58s
```

We are now ready to apply the `MongoDBOpsRequest` CR to update this database.

### update MongoDB Version

Here, we are going to update `MongoDB` standalone from `3.6.8` to `4.0.5`.

#### Create MongoDBOpsRequest:

In order to update the standalone database, we have to create a `MongoDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mg-standalone
  updateVersion:
    targetVersion: 4.4.26
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mg-standalone` MongoDB database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `4.0.5`.
- Have a look [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.


Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/update-version/mops-update-standalone.yaml
mongodbopsrequest.ops.kubedb.com/mops-update created
```

#### Verify MongoDB version updated successfully :

If everything goes well, `KubeDB` Ops-manager operator will update the image of `MongoDB` object and related `PetSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME           TYPE            STATUS       AGE
mops-update   UpdateVersion   Successful   3m45s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to update the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-update
Name:         mops-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2022-10-26T10:06:50Z
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
    Time:         2022-10-26T10:06:50Z
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
    Time:            2022-10-26T10:08:25Z
  Resource Version:  605817
  UID:               79faadf6-7af9-4b74-9907-febe7d543386
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mg-standalone
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:                             UpdateVersion
  UpdateVersion:
    Target Version:  4.4.26
Status:
  Conditions:
    Last Transition Time:  2022-10-26T10:07:10Z
    Message:               MongoDB ops request is update-version database version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2022-10-26T10:07:30Z
    Message:               Successfully updated petsets update strategy type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2022-10-26T10:08:25Z
    Message:               Successfully Updated Standalone Image
    Observed Generation:   1
    Reason:                UpdateStandaloneImage
    Status:                True
    Type:                  UpdateStandaloneImage
    Last Transition Time:  2022-10-26T10:08:25Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age   From                         Message
  ----    ------                 ----  ----                         -------
  Normal  PauseDatabase          2m5s  KubeDB Ops-manager Operator  Pausing MongoDB demo/mg-standalone
  Normal  PauseDatabase          2m5s  KubeDB Ops-manager Operator  Successfully paused MongoDB demo/mg-standalone
  Normal  Updating               2m5s  KubeDB Ops-manager Operator  Updating PetSets
  Normal  Updating               105s  KubeDB Ops-manager Operator  Successfully Updated PetSets
  Normal  UpdateStandaloneImage  50s   KubeDB Ops-manager Operator  Successfully Updated Standalone Image
  Normal  ResumeDatabase         50s   KubeDB Ops-manager Operator  Resuming MongoDB demo/mg-standalone
  Normal  ResumeDatabase         50s   KubeDB Ops-manager Operator  Successfully resumed MongoDB demo/mg-standalone
  Normal  Successful             50s   KubeDB Ops-manager Operator  Successfully Updated Database

```

Now, we are going to verify whether the `MongoDB` and the related `PetSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get mg -n demo mg-standalone -o=jsonpath='{.spec.version}{"\n"}'                                                                                          
4.4.26

$ kubectl get sts -n demo mg-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                               
mongo:4.0.5

$ kubectl get pods -n demo mg-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           
mongo:4.0.5
```

You can see from above, our `MongoDB` standalone database has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-standalone
kubectl delete mongodbopsrequest -n demo mops-update
```