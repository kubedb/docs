---
title: Horizontal Scaling MySQL group replication
menu:
  docs_{{ .version }}:
    identifier: my-horizontal-scale
    name: my-horizontal-scale
    parent: my-upgrading-mysql
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# Horizontal Scale MySQL Group Replication

This guide will show you how to use `KubeDB` enterprise operator to scale the number of server nodes of a `MySQL` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/concepts/databases/mysql.md)
  - [MySQLOpsRequest](/docs/concepts/day-2-operations/mysqlopsrequest.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/day-2-operations](/docs/examples/day-2-operations) directory of [stashed/docs](https://github.com/stashed/docs) repository.

### Apply Horizontal Scaling on MySQL Group Replication

Here, we are going to deploy a  `MySQL` group replication using a supported version by `KubeDB` operator. Below section will check the supported `MySQL` versions.

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

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. Now we will select a version from `MySQLVersion` for `MySQL` group replication. For `MySQL` group replication deployment, we will select version `8.0.20`.

#### Prepare Group Replication

Now, we are going to deploy a `MySQL` group replication using version `8.0.20`.

**Create MySQL Object:**

Below is the YAML of the `MySQL` crd that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "8.0.20"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/group_replication2.yaml
mysql.kubedb.com/my-group created
```

**Check MySQL group Ready to Scale:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services and Secrets etc. A secret called `my-group-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
Now, watch `MySQL` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```console
$ watch -n 3 kubectl get my -n demo my-group
Every 3.0s: kubectl get my -n demo my-group                     suaas-appscode: Tue Jun 30 22:43:57 2020

NAME       VERSION   STATUS    AGE
my-group   8.0.20    Running   16m

$ watch -n 3 kubectl get sts -n demo my-group
Every 3.0s: kubectl get sts -n demo my-group                     Every 3.0s: kubectl get sts -n demo my-group                    suaas-appscode: Tue Jun 30 22:44:35 2020

NAME       READY   AGE
my-group   3/3     16m

$ watch -n 3 kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group
Every 3.0s: kubectl get pod -n demo -l kubedb.com/kind=MySQ...  suaas-appscode: Tue Jun 30 22:45:33 2020

NAME         READY   STATUS    RESTARTS   AGE
my-group-0   1/1     Running   0          17m
my-group-1   1/1     Running   0          14m
my-group-2   1/1     Running   0          11m
```

Let's check the StatefulSet pods have formed a MySQL group replication,

```console
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
sWfUMoqRpOJyomgb

$ kubectl exec -it -n demo my-group-0 -- mysql -u root --password=sWfUMoqRpOJyomgb --host=my-group-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| group_replication_applier | 596be47b-baef-11ea-859a-02c946ef4fe7 | my-group-1.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
| group_replication_applier | 815974c2-baef-11ea-bd7e-a695cbdbd6cc | my-group-2.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
| group_replication_applier | ec61cef2-baee-11ea-adb0-9a02630bae5d | my-group-0.my-group-gvr.demo |        3306 | ONLINE       | PRIMARY     | 8.0.20         |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
```

We are ready to apply horizontal scale on this group replication.

#### Horizontal Scaling

Here, we are going to scale up the number of server nodes of the `MySQL` group replication.

**Create MySQLOpsRequest:**

Below is the YAML of the `MySQLOpsRequest` crd that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  type: HorizontalScaling  
  databaseRef:
    name: my-group
  horizontalScaling:
    member: 5
```

Here,

- `spec.databaseRef.name` refers to the `my-group` MySQL object for operation.
- `spec.type` specifies that this is an `HorizontalScaling` type operation
- `spec.horizontalScaling.member` specifies final expected number of nodes for group replication.

Let's create the `MySQLOpsRequest` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/horizontal_scale.yaml
mysqlopsrequest.ops.kubedb.com/myopsreq-group created
```

**Check MySQL server nodes scaled up:**

If everything goes well, `KubeDB` enterprise operator will scale up the StatefulSet's `Pod`. After successful scaling process is done, the `KubeDB` enterprise operator update the replicas of the `MySQL` object.

First, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` crd,

```console
$ watch -n 3 kubectl get myops -n demo myopsreq-group
Every 3.0s: kubectl get myops -n demo myopsreq-group             suaas-appscode: Thu Jun 18 14:48:05 2020

NAME             TYPE      STATUS       AGE
myopsreq-group   Upgrade   Successful   8m30s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If you describe the `MySQLOpsRequest` you will see that the `MySQL` group replication is scaled up.

```console
$ kubectl describe myops -n demo myops
Name:         myops-horizontal
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2020-06-30T17:35:48Z
  Finalizers:
    mysql.ops.kubedb.com
  Generation:        3
  Resource Version:  13681
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/myops
  UID:               cf58a544-bc8f-4d40-a565-cece7b85e44e
Spec:
  Database Ref:
    Name:  my-group
  Horizontal Scaling:
    Member:              5
    Member Weight:       50
  Stateful Set Ordinal:  0
  Type:                  HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2020-06-30T17:35:48Z
    Message:               The controller has started to Progress the OpsRequest
    Observed Generation:   1
    Reason:                OpsRequestOpsRequestProgressing
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-06-30T17:35:48Z
    Message:               MySQLOpsRequestDefinition: myops for Pausing MySQL: demo/my-group
    Observed Generation:   1
    Reason:                PausingDatabase
    Status:                True
    Type:                  PausingDatabase
    Last Transition Time:  2020-06-30T17:35:48Z
    Message:               MySQLOpsRequestDefinition: myops for Paused MySQL: demo/my-group
    Observed Generation:   1
    Reason:                PausedDatabase
    Status:                True
    Type:                  PausedDatabase
    Last Transition Time:  2020-06-30T17:35:49Z
    Message:               MySQLOpsRequestDefinition: myops for Scaling MySQL: demo/my-group
    Observed Generation:   1
    Reason:                OpsRequestScalingDatabase
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2020-06-30T17:38:29Z
    Message:               MySQLOpsRequestDefinition: myops for Horizontal Scaling MySQL: demo/my-group
    Observed Generation:   1
    Reason:                OpsRequestHorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2020-06-30T17:38:29Z
    Message:               MySQLOpsRequestDefinition: myops for Resuming MySQL: demo/my-group
    Observed Generation:   3
    Reason:                ResumingDatabase
    Status:                True
    Type:                  ResumingDatabase
    Last Transition Time:  2020-06-30T17:38:30Z
    Message:               MySQLOpsRequestDefinition: myops for Reasumed MySQL: demo/my-group
    Observed Generation:   3
    Reason:                ResumedDatabase
    Status:                True
    Type:                  ResumedDatabase
    Last Transition Time:  2020-06-30T17:38:30Z
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
  Normal  Pausing          7m35s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops, Pausing MySQL: demo/my-group
  Normal  SuccessfulPause  7m35s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops, successfully paused: demo/my-group
  Normal  Starting         7m34s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops for Scaling MySQL: demo/my-group
  Normal  Starting         4m54s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops, Horizontal Scaling MySQL: demo/my-group
  Normal  Resuming         4m54s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops, Resuming MySQL: demo/my-group
  Normal  Successful       4m54s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops, Resumed for MySQL: demo/my-group
```

Now, we are going to verify whether the number of server nodes have increased to meet up the desire state, Let's check,

```console
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
Y28qkWFQ8QHVzq2h

$ kubectl exec -it -n demo my-group-0 -- mysql -u root --password=Y28qkWFQ8QHVzq2h --host=my-group-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| group_replication_applier | 4b76f5c8-baff-11ea-9848-425294afbbbf | my-group-3.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
| group_replication_applier | 73c1f150-baff-11ea-9394-4a8c424ea5c2 | my-group-4.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
| group_replication_applier | 9f6c694c-bafd-11ea-8ad4-822669614bde | my-group-0.my-group-gvr.demo |        3306 | ONLINE       | PRIMARY     | 8.0.20         |
| group_replication_applier | c9d82f09-bafd-11ea-ab3a-764d326534a6 | my-group-1.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
| group_replication_applier | eff81073-bafd-11ea-9f3d-ca1e99c33106 | my-group-2.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.20         |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
```

You can see above that our `MySQL` group replication now have total 5 server nodes. It verify that we have successfully scaled up.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo my-group
kubectl delete myops -n demo myops
```