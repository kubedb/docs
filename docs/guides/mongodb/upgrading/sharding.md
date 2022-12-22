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

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade version of MongoDB Sharded Database

This guide will show you how to use `KubeDB` Ops-manager operator to upgrade the version of `MongoDB` Sharded Database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

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

Now, wait until `mg-sharding` created has status `Ready`. i.e,

```bash
$ k get mongodb -n demo                                                                                                                                             
NAME          VERSION    STATUS    AGE
mg-sharding   3.6.8-v1   Ready     2m9s
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
  type: UpdateVersion
  databaseRef:
    name: mg-sharding
  upgrade:
    targetVersion: 4.0.5-v3
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mg-sharding` MongoDB database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `4.0.5`.
- Have a look [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/upgrading/mops-upgrade-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-shard-upgrade created
```

#### Verify MongoDB version upgraded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                 TYPE            STATUS       AGE
mops-shard-upgrade   UpdateVersion   Successful   2m31s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to upgrade the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-shard-upgrade

Name:         mops-shard-upgrade
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2022-10-26T10:27:24Z
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
        f:upgrade:
          .:
          f:targetVersion:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-10-26T10:27:24Z
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
    Time:            2022-10-26T10:36:12Z
  Resource Version:  610193
  UID:               6459a314-c759-4002-9dff-106b836c4db0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mg-sharding
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:                             UpdateVersion
  Upgrade:
    Target Version:  4.0.5-v3
Status:
  Conditions:
    Last Transition Time:  2022-10-26T10:36:12Z
    Message:               connection() error occurred during connection handshake: dial tcp 10.244.0.125:27017: i/o timeout
    Observed Generation:   1
    Reason:                Failed
    Status:                False
    Type:                  UpgradeVersion
    Last Transition Time:  2022-10-26T10:29:29Z
    Message:               Successfully stopped mongodb load balancer
    Observed Generation:   1
    Reason:                StoppingBalancer
    Status:                True
    Type:                  StoppingBalancer
    Last Transition Time:  2022-10-26T10:30:54Z
    Message:               Successfully updated statefulsets update strategy type
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2022-10-26T10:32:00Z
    Message:               Successfully Updated ConfigServer Image
    Observed Generation:   1
    Reason:                UpdateConfigServerImage
    Status:                True
    Type:                  UpdateConfigServerImage
    Last Transition Time:  2022-10-26T10:35:32Z
    Message:               Successfully Updated Shard Image
    Observed Generation:   1
    Reason:                UpdateShardImage
    Status:                True
    Type:                  UpdateShardImage
    Last Transition Time:  2022-10-26T10:36:07Z
    Message:               Successfully Updated Mongos Image
    Observed Generation:   1
    Reason:                UpdateMongosImage
    Status:                True
    Type:                  UpdateMongosImage
    Last Transition Time:  2022-10-26T10:36:07Z
    Message:               Successfully Started mongodb load balancer
    Observed Generation:   1
    Reason:                StartingBalancer
    Status:                True
    Type:                  StartingBalancer
    Last Transition Time:  2022-10-26T10:36:07Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Failed
Events:
  Type     Reason                   Age    From                         Message
  ----     ------                   ----   ----                         -------
  Normal   PauseDatabase            8m27s  KubeDB Ops-manager Operator  Pausing MongoDB demo/mg-sharding
  Normal   PauseDatabase            8m27s  KubeDB Ops-manager Operator  Successfully paused MongoDB demo/mg-sharding
  Normal   StoppingBalancer         8m27s  KubeDB Ops-manager Operator  Stopping Balancer
  Normal   StoppingBalancer         8m27s  KubeDB Ops-manager Operator  Successfully Stopped Balancer
  Normal   Updating                 8m27s  KubeDB Ops-manager Operator  Updating StatefulSets
  Normal   Updating                 7m2s   KubeDB Ops-manager Operator  Successfully Updated StatefulSets
  Normal   Updating                 7m2s   KubeDB Ops-manager Operator  Updating StatefulSets
  Normal   UpdateConfigServerImage  5m56s  KubeDB Ops-manager Operator  Successfully Updated ConfigServer Image
  Normal   Updating                 5m45s  KubeDB Ops-manager Operator  Successfully Updated StatefulSets
  Normal   UpdateShardImage         2m24s  KubeDB Ops-manager Operator  Successfully Updated Shard Image
  Normal   UpdateMongosImage        109s   KubeDB Ops-manager Operator  Successfully Updated Mongos Image
  Normal   Updating                 109s   KubeDB Ops-manager Operator  Starting Balancer
  Normal   StartingBalancer         109s   KubeDB Ops-manager Operator  Successfully Started Balancer
  Normal   ResumeDatabase           109s   KubeDB Ops-manager Operator  Resuming MongoDB demo/mg-sharding
  Normal   ResumeDatabase           109s   KubeDB Ops-manager Operator  Successfully resumed MongoDB demo/mg-sharding
  Normal   Successful               109s   KubeDB Ops-manager Operator  Successfully Updated Database
```

Now, we are going to verify whether the `MongoDB` and the related `StatefulSets` of `Mongos`, `Shard` and `ConfigeServer` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get mg -n demo mg-sharding -o=jsonpath='{.spec.version}{"\n"}'
4.0.5-v3

$ kubectl get sts -n demo mg-sharding-configsvr -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mongo:4.0.5

$ kubectl get sts -n demo mg-sharding-shard0 -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mongo:4.0.5

$ kubectl get sts -n demo mg-sharding-mongos -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mongo:4.0.5

$ kubectl get pods -n demo mg-sharding-configsvr-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mongo:4.0.5

$ kubectl get pods -n demo mg-sharding-shard0-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mongo:4.0.5

$ kubectl get pods -n demo mg-sharding-mongos-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mongo:4.0.5
```

You can see from above, our `MongoDB` sharded database has been updated with the new version. So, the upgrade process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-sharding
kubectl delete mongodbopsrequest -n demo mops-shard-upgrade
```