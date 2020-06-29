---
title: Upgrading MySQL standalone
menu:
  docs_{{ .version }}:
    identifier: my-upgrade-standalone
    name: my-upgrade-standalone
    parent: my-upgrading-mysql
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# Upgrade MySQL Group Replication using KubeDB enterprise operator

This guide will show you how to use `KubeDB` enterprise operator to upgrade the `MySQL` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` general and enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/concepts/databases/mysql.md)
  - [MySQLOpsRequest](/docs/concepts/day-2-operations/mysqlopsrequest.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/day-2-operations](/docs/examples/day-2-operations) directory of [stashed/docs](https://github.com/stashed/docs) repository.

### Upgrade MySQL Group Replication

Here, we are going to deploy a  `MySQL` group replication using a supported version by `KubeDB` operator. Below two sections will check the supported `MySQL` versions and then check whether it is possible to upgrade from this version to another.

**Find supported MySQL Version:**

When you have installed `KubeDB`, it has created `MySQLVersion` crd for all supported `MySQL` versions. Let's check support versions,

```console
$ kubectl get mysqlversion
NAME        VERSION   DB_IMAGE                 DEPRECATED   AGE
5           5         kubedb/mysql:5           true         49s
5-v1        5         kubedb/mysql:5-v1        true         49s
5.7         5.7       kubedb/mysql:5.7         true         49s
5.7-v1      5.7       kubedb/mysql:5.7-v1      true         49s
5.7-v2      5.7.25    kubedb/mysql:5.7-v2      true         49s
5.7-v3      5.7.25    kubedb/mysql:5.7.25      true         49s
5.7-v4      5.7.29    kubedb/mysql:5.7.29                   49s
5.7.25      5.7.25    kubedb/mysql:5.7.25      true         49s
5.7.25-v1   5.7.25    kubedb/mysql:5.7.25-v1                49s
5.7.29      5.7.29    kubedb/mysql:5.7.29                   49s
8           8         kubedb/mysql:8           true         49s
8-v1        8         kubedb/mysql:8-v1        true         49s
8.0         8.0       kubedb/mysql:8.0         true         49s
8.0-v1      8.0.3     kubedb/mysql:8.0-v1      true         49s
8.0-v2      8.0.14    kubedb/mysql:8.0-v2      true         49s
8.0-v3      8.0.20    kubedb/mysql:8.0.20                   49s
8.0.14      8.0.14    kubedb/mysql:8.0.14      true         49s
8.0.14-v1   8.0.14    kubedb/mysql:8.0.14-v1                49s
8.0.18      8.0.18    kubedb/mysql:8.0.18                   49s
8.0.19      8.0.19    kubedb/mysql:8.0.19                   49s
8.0.20      8.0.20    kubedb/mysql:8.0.20                   49s
8.0.3       8.0.3     kubedb/mysql:8.0.3       true         49s
8.0.3-v1    8.0.3     kubedb/mysql:8.0.3-v1                 49s
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. Now we will select a version from `MySQLVersion` for `MySQL` group replication that will be possible to upgrade from this version to another version. For `MySQL` group replication deployment, we will select version `5.7.29`. The below section will check this version's upgrade constraints.

**Check Upgrade Constraints:**

"Version upgrade constraints" is a way to show whether it is possible or not possible to upgrade from one version to another. Let's check the version upgrade constraints of `5.7.29`,

```console
$ kubectl get mysqlversion 5.7.29 -o yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MySQLVersion
metadata:
  name: 5.7.29
  ...
spec:
  db:
    image: kubedb/mysql:5.7.29
  exporter:
    image: kubedb/mysqld-exporter:v0.11.0
  initContainer:
    image: kubedb/busybox
  podSecurityPolicies:
    databasePolicyName: mysql-db
  replicationModeDetector:
    image: kubedb/mysql-replication-mode-detector:v0.0.1
  tools:
    image: kubedb/mysql-tools:5.7.25
  upgradeConstraints:
    blacklist:
      groupReplication:
      - < 5.7.29
      standalone:
      - < 5.7.29
  version: 5.7.29
```

The above `spec.upgradeConstraints.blacklist` is showing that upgrading below version `5.7.29` is not possible for both group replication and standalone. That means, it is possible to upgrade any version above `5.7.29`. The below section will describe deploying `MySQL` group replication using version `5.7.29`.

#### Prepare Group Replication

Now, we are going to deploy a `MySQL` group replication using version `5.7.29`.

**Create MySQL Object:**

Below is the YAML of the `MySQL` crd that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "5.7.29"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
      baseServerID: 100
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `MySQL` crd we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/group_replication.yaml
mysql.kubedb.com/my-group created
```

**Check MySQL group Ready to Upgrade:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services and Secrets etc. A secret called `my-group-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
Now, watch `MySQL` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```console
$ watch -n 3 kubectl get my -n demo my-group
Every 3.0s: kubectl get my -n demo my-group                      suaas-appscode: Thu Jun 18 14:30:24 2020

NAME       VERSION   STATUS    AGE
my-group   5.7.29    Running   5m52s

$ watch -n 3 kubectl get sts -n demo my-group
Every 3.0s: kubectl get sts -n demo my-group                     suaas-appscode: Thu Jun 18 14:31:44 2020

NAME       READY   AGE
my-group   3/3     7m12s

$ watch -n 3 kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group
Every 3.0s: kubectl get pod -n demo -l kubedb.com/kind=MySQL...  suaas-appscode: Thu Jun 18 14:35:35 2020

NAME         READY   STATUS    RESTARTS   AGE
my-group-0   1/1     Running   0          11m
my-group-1   1/1     Running   0          9m53s
my-group-2   1/1     Running   0          6m48s
```

Let's check the `MySQL`, `StatefulSet` and its `Pod` image version,

```console
$ kubectl get my -n demo my-group -o=jsonpath='{.spec.version}{"\n"}'
5.7.29

$ kubectl get sts -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.template.spec.containers[0].image'
"kubedb/my:5.7.29"

$ kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.containers[0].image'
"kubedb/my:5.7.29"
"kubedb/my:5.7.29"
"kubedb/my:5.7.29"
```

Let's also check the StatefulSet pods have formed a MySQL group replication,

```console
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
sWfUMoqRpOJyomgb

kubectl exec -it -n demo my-group-0 -- mysql -u root --password=sWfUMoqRpOJyomgb --host=my-group-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| group_replication_applier | 356c03c0-b9cb-11ea-b856-7e6de479ee9d | my-group-1.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | 5be97bc0-b9cb-11ea-8b8b-a2eec7faa37d | my-group-2.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | c7089cb3-b9ca-11ea-b92a-228a3699132f | my-group-0.my-group-gvr.demo |        3306 | ONLINE       |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
```

We are ready to upgrade the above `MySQL` group replication.

#### Upgrade

Here, we are going to upgrade `MySQL` group replication from version `5.7.29` to `8.0.20`.

**Create MySQLOpsRequest:**

Below is the YAML of the `MySQLOpsRequest` crd that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myopsreq-group
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: Upgrade
  upgrade:
    targetVersion: "8.0.20"
```

Here,

- `spec.databaseRef.name` refers to the `my-group` MySQL object for operation.
- `spec.type` specifies that this is an `Upgrade` type operation
- `spec.upgrade.targetVersion` specifies version `8.0.20` that will be upgraded from version `5.7.29`.

Let's create the `MySQLOpsRequest` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/upgrade_group.yaml
mysqlopsrequest.ops.kubedb.com/myopsreq-group created
```

> Note: In the group replication of mysql, a new statefulset is created by KubeDB enterprise operator in the field of major version upgrading and the old one is deleted. The name of the StatefulSet is formed as follows: `<mysql-name>-<suffix>`.
Here, `<suffix>` is a positive integer number and starts with 1. It's determined as follows:
For one-time major version upgrading of group replication, suffix will be 1.
For the 2nd time major version upgrading of group replication, suffix will be 2.
It will be continued...

**Check MySQL version upgraded:**

If everything goes well, `KubeDB` enterprise operator will update the images of `MySQL`and will create a new `StatefulSet` named `my-group-1`.

First, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` crd,

```console
$ watch -n 3 kubectl get myops -n demo myopsreq-group
Every 3.0s: kubectl get myops -n demo myopsreq-group             suaas-appscode: Thu Jun 18 14:48:05 2020

NAME             TYPE      STATUS       AGE
myopsreq-group   Upgrade   Successful   8m30s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If you describe the `MySQLOpsRequest` you will see that the `MySQL` is updated with new images and the `StatefulSet` and its `Pod` is created with new images.

```console
$ kubectl describe myops -n demo myopsreq-group
Name:         myopsreq-group
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2020-06-18T08:39:35Z
  Finalizers:
    mysql.ops.kubedb.com
  Generation:        3
  Resource Version:  12961
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/myopsreq-group
  UID:               e091e283-373e-4633-9b9f-f43983a122be
Spec:
  Database Ref:
    Name:                my-group
  Stateful Set Ordinal:  1
  Type:                  Upgrade
  Upgrade:
    Target Version:  8.0.20
Status:
  Conditions:
    Last Transition Time:  2020-06-18T08:39:35Z
    Message:               The controller has started to Progress the OpsRequest
    Observed Generation:   1
    Reason:                OpsRequestOpsRequestProgressing
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-06-18T08:39:35Z
    Message:               MySQLOpsRequestDefinition: myopsreq-group for Pausing MySQL: demo/my-group
    Observed Generation:   1
    Reason:                PausingDatabase
    Status:                True
    Type:                  PausingDatabase
    Last Transition Time:  2020-06-18T08:39:35Z
    Message:               MySQLOpsRequestDefinition: myopsreq-group for Paused MySQL: demo/my-group
    Observed Generation:   1
    Reason:                PausedDatabase
    Status:                True
    Type:                  PausedDatabase
    Last Transition Time:  2020-06-18T08:39:35Z
    Message:               MySQLOpsRequestDefinition: myopsreq-group for Upgrading MySQL version: demo/my-group
    Observed Generation:   1
    Reason:                OpsRequestUpgradingVersion
    Status:                True
    Type:                  UpgradingVersion
    Last Transition Time:  2020-06-18T08:47:55Z
    Message:               MySQLOpsRequestDefinition: myopsreq-group for image successfully upgraded for MySQ: demo/my-group
    Observed Generation:   1
    Reason:                OpsRequestUpgradedVersion
    Status:                True
    Type:                  UpgradedVersion
    Last Transition Time:  2020-06-18T08:47:56Z
    Message:               MySQLOpsRequestDefinition: myopsreq-group for Resuming MySQL: demo/my-group
    Observed Generation:   3
    Reason:                ResumingDatabase
    Status:                True
    Type:                  ResumingDatabase
    Last Transition Time:  2020-06-18T08:47:56Z
    Message:               MySQLOpsRequestDefinition: myopsreq-group for Reasumed MySQL: demo/my-group
    Observed Generation:   3
    Reason:                ResumedDatabase
    Status:                True
    Type:                  ResumedDatabase
    Last Transition Time:  2020-06-18T08:47:56Z
    Message:               The controller has scaled/upgraded the MySQL successfully
    Observed Generation:   3
    Reason:                OpsRequestSuccessful
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason           Age    From                        Message
  ----    ------           ----   ----                        -------
  Normal  Pausing          10m    KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, Pausing MySQL: demo/my-group
  Normal  SuccessfulPause  10m    KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, successfully paused: demo/my-group
  Normal  Starting         10m    KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, Upgrading MySQL images: demo/my-group
  Normal  Successful       8m1s   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, image successfully upgraded for Pod: demo/my-group-1-0
  Normal  Successful       4m41s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, image successfully upgraded for Pod: demo/my-group-1-1
  Normal  Successful       2m21s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, image successfully upgraded for Pod: demo/my-group-1-2
  Normal  Successful       2m1s   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, image successfully upgraded for MySQL: demo/my-group
  Normal  Resuming         2m     KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, Resuming MySQL: demo/my-group
  Normal  Successful       2m     KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-group, Resumed for MySQL: demo/my-group
```

Now, we are going to verify whether the `MySQL`, `StatefulSet` and it's `Pod` images have updated. Let's check,

```console
$ kubectl get my -n demo my-group -o=jsonpath='{.spec.version}{"\n"}'
8.0.20

$ kubectl get sts -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.template.spec.containers[0].image'
"kubedb/my:8.0.20"

$ kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.containers[0].image'
"kubedb/my:8.0.20"
"kubedb/my:8.0.20"
"kubedb/my:8.0.20"
```

Let's also check the StatefulSet pods have joined the `MySQL` group replication,

```console
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
sWfUMoqRpOJyomgb

$ kubectl exec -it -n demo my-group-1-0 -- mysql -u root --password=sWfUMoqRpOJyomgb --host=my-group-1-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+--------------------------------+-------------+--------------+-------------+----------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                    | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION |
+---------------------------+--------------------------------------+--------------------------------+-------------+--------------+-------------+----------------+
| group_replication_applier | 4d594b52-b9e8-11ea-b389-22889501aae8 | my-group-1-1.my-group-gvr.demo |        3306 | ONLINE       | PRIMARY     | 8.0.20         |
| group_replication_applier | 866bb020-b9e8-11ea-bc06-52c624549b83 | my-group-1-2.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
| group_replication_applier | d4b2dc04-b9e7-11ea-a833-72f402a520fd | my-group-1-0.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
+---------------------------+--------------------------------------+--------------------------------+-------------+--------------+-------------+----------------+
```

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo my-standalone
kubectl delete myops -n demo myopsreq-standalone
```