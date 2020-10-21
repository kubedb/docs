---
title: Upgrading MongoDB Sharded Database
menu:
  docs_{{ .version }}:
    identifier: mg-upgrading-sharding
    name: Sharding
    parent: mg-upgrading
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade version of MongoDB Sharded Database

This guide will show you how to use `KubeDB` Enterprise operator to upgrade the version of `MongoDB` Sharded Database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Sharding](/docs/guides/mongodb/clustering/sharding.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Upgrading Overview](/docs/guides/mongodb/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare MongoDB Sharded Database Database

Now, we are going to deploy a `MongoDB` sharded database with version `3.6.8`.

### Deploy MongoDB Sharded Database 

In this section, we are going to deploy a MongoDB sharded database. Then, in the next section we will upgrade the version of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-sharding
  namespace: demo
spec:
  version: 3.6.8-v1
  shardTopology:
    configServer:
      replicas: 2
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 2
      shards: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/upgrading/mg-shard.yaml
mongodb.kubedb.com/mg-sharding created
```

Now, wait until `mg-sharding` created has status `Running`. i.e,

```bash
$ k get mongodb -n demo                                                                                                                                             
NAME          VERSION    STATUS    AGE
mg-sharding   3.6.8-v1   Running   2m9s
```

We are now ready to apply the `MongoDBOpsRequest` CR to upgrade this database.

### Upgrade MongoDB Version

Here, we are going to upgrade `MongoDB` sharded database from `3.6.8` to `4.0.5`.

#### Create MongoDBOpsRequest

In order to upgrade the sharded database, we have to create a `MongoDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-shard-upgrade
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: mg-sharding
  upgrade:
    targetVersion: 4.0.5-v3
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mg-sharding` MongoDB database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `4.0.5`.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/upgrading/mops-upgrade-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-shard-upgrade created
```

#### Verify MongoDB version upgraded successfully

If everything goes well, `KubeDB` Enterprise operator will update the image of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                 TYPE      STATUS       AGE
mops-shard-upgrade   Upgrade   Successful   2m31s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to upgrade the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-shard-upgrade

Name:         mops-shard-upgrade
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2020-08-24T15:17:47Z
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
    Time:         2020-08-24T15:17:47Z
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
    Time:            2020-08-24T15:19:51Z
  Resource Version:  4830892
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-shard-upgrade
  UID:               9683d8ed-35fa-4782-990d-0fb64160cd4a
Spec:
  Database Ref:
    Name:  mg-sharding
  Type:    Upgrade
  Upgrade:
    Target Version:  4.0.5-v3
Status:
  Conditions:
    Last Transition Time:  2020-08-24T15:17:47Z
    Message:               MongoDB ops request is starting to process
    Observed Generation:   1
    Reason:                UpgradingVersion
    Status:                True
    Type:                  UpgradingVersion
    Last Transition Time:  2020-08-24T15:17:48Z
    Message:               Successfully halted mongodb: mg-sharding
    Observed Generation:   1
    Reason:                HaltDatabase
    Status:                True
    Type:                  HaltDatabase
    Last Transition Time:  2020-08-24T15:17:50Z
    Message:               Succesfully stopped mongodb load balancer
    Observed Generation:   1
    Reason:                StoppingBalancer
    Status:                True
    Type:                  StoppingBalancer
    Last Transition Time:  2020-08-24T15:17:50Z
    Message:               Successfully updated statefulsets update strategy type
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2020-08-24T15:18:15Z
    Message:               Successfully updated ConfigServer images
    Observed Generation:   1
    Reason:                UpdateConfigServerImage
    Status:                True
    Type:                  UpdateConfigServerImage
    Last Transition Time:  2020-08-24T15:19:25Z
    Message:               Successfully updated Shard images
    Observed Generation:   1
    Reason:                UpdateShardImage
    Status:                True
    Type:                  UpdateShardImage
    Last Transition Time:  2020-08-24T15:19:50Z
    Message:               Successfully updated Mongos images
    Observed Generation:   1
    Reason:                UpdateMongosImage
    Status:                True
    Type:                  UpdateMongosImage
    Last Transition Time:  2020-08-24T15:19:51Z
    Message:               Successfully Started mongodb load balancer
    Observed Generation:   1
    Reason:                StartingBalancer
    Status:                True
    Type:                  StartingBalancer
    Last Transition Time:  2020-08-24T15:19:51Z
    Message:               Succefully Resumed mongodb: mg-sharding
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-24T15:19:51Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                   Age    From                        Message
  ----    ------                   ----   ----                        -------
  Normal  HaltDatabase            3m23s  KubeDB Enterprise Operator  Pausing Mongodb mg-sharding in Namespace demo
  Normal  HaltDatabase            3m23s  KubeDB Enterprise Operator  Successfully Halted Mongodb mg-sharding in Namespace demo
  Normal  StoppingBalancer         3m23s  KubeDB Enterprise Operator  Stopping Balancer
  Normal  StoppingBalancer         3m21s  KubeDB Enterprise Operator  Successfully Stopped Balancer
  Normal  Updating                 3m21s  KubeDB Enterprise Operator  Updating StatefulSets
  Normal  Updating                 3m21s  KubeDB Enterprise Operator  Successfully Updated StatefulSets
  Normal  UpdateConfigServerImage  3m21s  KubeDB Enterprise Operator  Updating ConfigServer Images
  Normal  UpdateConfigServerImage  3m11s  KubeDB Enterprise Operator  Successfully Updated Images of Pod mg-sharding-configsvr-1
  Normal  UpdateConfigServerImage  2m56s  KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-sharding-configsvr-0
  Normal  UpdateConfigServerImage  2m56s  KubeDB Enterprise Operator  Successfully Updated ConfigServer Images
  Normal  UpdateShardImage         2m56s  KubeDB Enterprise Operator  Updating Shard Images
  Normal  UpdateShardImage         2m36s  KubeDB Enterprise Operator  Successfully Updated Images of Pod mg-sharding-shard0-1
  Normal  UpdateShardImage         2m21s  KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-sharding-shard0-0
  Normal  UpdateShardImage         2m16s  KubeDB Enterprise Operator  Successfully Updated Images of Pod mg-sharding-shard1-1
  Normal  UpdateShardImage         2m6s   KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-sharding-shard1-0
  Normal  UpdateShardImage         116s   KubeDB Enterprise Operator  Successfully Updated Images of Pod mg-sharding-shard2-1
  Normal  UpdateShardImage         106s   KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-sharding-shard2-0
  Normal  UpdateShardImage         106s   KubeDB Enterprise Operator  Successfully Updated Shard Images
  Normal  UpdateMongosImage        106s   KubeDB Enterprise Operator  Updating Mongos Images
  Normal  UpdateMongosImage        86s    KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-sharding-mongos-0
  Normal  UpdateMongosImage        81s    KubeDB Enterprise Operator  Successfully Updated Images of Pod (master): mg-sharding-mongos-1
  Normal  UpdateMongosImage        81s    KubeDB Enterprise Operator  Successfully Updated Mongos Images
  Normal  Updating                 81s    KubeDB Enterprise Operator  Starting Balancer
  Normal  StartingBalancer         80s    KubeDB Enterprise Operator  Successfully Started Balancer
  Normal  ResumeDatabase           80s    KubeDB Enterprise Operator  Resuming MongoDB
  Normal  ResumeDatabase           80s    KubeDB Enterprise Operator  Successfully Started Balancer
  Normal  Successful               80s    KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify whether the `MongoDB` and the related `StatefulSets` of `Mongos`, `Shard` and `ConfigeServer` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get mg -n demo mg-sharding -o=jsonpath='{.spec.version}{"\n"}'                                                                                           20:59:43
  4.0.5-v3

$ kubectl get sts -n demo mg-sharding-configsvr -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                                21:01:32
  kubedb/mongo:4.0.5-v3

$ kubectl get sts -n demo mg-sharding-shard0 -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                                21:01:32
  kubedb/mongo:4.0.5-v3

$ kubectl get sts -n demo mg-sharding-mongos -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                                21:01:32
  kubedb/mongo:4.0.5-v3

$ kubectl get pods -n demo mg-sharding-configsvr-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           21:01:57
  kubedb/mongo:4.0.5-v3

$ kubectl get pods -n demo mg-sharding-shard0-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           21:01:57
  kubedb/mongo:4.0.5-v3

$ kubectl get pods -n demo mg-sharding-mongos-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           21:01:57
  kubedb/mongo:4.0.5-v3
```

You can see from above, our `MongoDB` sharded database has been updated with the new version. So, the upgrade process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-sharding
kubectl delete mongodbopsrequest -n demo mops-shard-upgrade
```