---
title: Upgrading MySQL group replication minor version
menu:
  docs_{{ .version }}:
    identifier: my-upgrading-mysql-minor-group
    name: Group Replication
    parent: my-upgrading-mysql-minor
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="Upgrading is an Enterprise feature of KubeDB. You must have a KubeDB Enterprise operator installed to test this feature." >}}

# Upgrade minor version of MySQL Group Replication

This guide will show you how to use `KubeDB` enterprise operator to upgrade the minor version of `MySQL` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/concepts/databases/mysql.md)
  - [MySQLOpsRequest](/docs/concepts/day-2-operations/mysqlopsrequest.md)
  - [Upgrading Overview](/docs/guides/mysql/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/day-2-operations/mysql](/docs/examples/day-2-operations/mysql) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Version Upgrading on Group Replication

Here, we are going to deploy a `MySQL` group replication using a supported version by `KubeDB` operator. Then we are going to apply upgrading on it.

#### Prepare Group Replication

At first, we are going to deploy a group replication using supported that `MySQL` version whether it is possible to upgrade from this version to another. In the next two sections, we are going to find out the supported version and version upgrade constraints.

**Find supported MySQL Version:**

When you have installed `KubeDB`, it has created `MySQLVersion` cr for all supported `MySQL` versions. Let’s check the supported `MySQL` versions,

```console
$ kubectl get mysqlversion
NAME        VERSION   DB_IMAGE                 DEPRECATED   AGE
5           5         kubedb/mysql:5           true         149m
5-v1        5         kubedb/mysql:5-v1        true         149m
5.7         5.7       kubedb/mysql:5.7         true         149m
5.7-v1      5.7       kubedb/mysql:5.7-v1      true         149m
5.7-v2      5.7.25    kubedb/mysql:5.7-v2      true         149m
5.7-v3      5.7.25    kubedb/mysql:5.7.25      true         149m
5.7-v4      5.7.29    kubedb/mysql:5.7.29      true         149m
5.7.25      5.7.25    kubedb/mysql:5.7.25      true         149m
5.7.25-v1   5.7.25    kubedb/mysql:5.7.25-v1                149m
5.7.29      5.7.29    kubedb/mysql:5.7.29                   149m
5.7.31      5.7.31    kubedb/mysql:5.7.31                   149m
8           8         kubedb/mysql:8           true         149m
8-v1        8         kubedb/mysql:8-v1        true         149m
8.0         8.0       kubedb/mysql:8.0         true         149m
8.0-v1      8.0.3     kubedb/mysql:8.0-v1      true         149m
8.0-v2      8.0.14    kubedb/mysql:8.0-v2      true         149m
8.0-v3      8.0.20    kubedb/mysql:8.0.20      true         149m
8.0.14      8.0.14    kubedb/mysql:8.0.14      true         149m
8.0.14-v1   8.0.14    kubedb/mysql:8.0.14-v1                149m
8.0.20      8.0.20    kubedb/mysql:8.0.20                   149m
8.0.21      8.0.21    kubedb/mysql:8.0.21                   149m
8.0.3       8.0.3     kubedb/mysql:8.0.3       true         149m
8.0.3-v1    8.0.3     kubedb/mysql:8.0.3-v1                 149m
```

The version above that does not show `DEPRECATED` true is supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `MySQLVersion` for `MySQL` group replication that will be possible to upgrade from this version to another version. In the next section, we are going to verify version upgrade constraints.

**Check Upgrade Constraints:**

Database version upgrade constraints is a constraint that shows whether it is possible or not possible to upgrade from one version to another. Let's check the version upgrade constraints of `MySQL` `5.7.29`,

```console
$ kubectl get mysqlversion 5.7.29 -o yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MySQLVersion
metadata:
  creationTimestamp: "2020-08-12T08:28:09Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb-catalog
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v0.14.0-beta.1
    helm.sh/chart: kubedb-catalog-v0.14.0-beta.1
  ...
    operation: Update
    time: "2020-08-12T08:28:09Z"
  name: 5.7.29
  resourceVersion: "42266"
  selfLink: /apis/catalog.kubedb.com/v1alpha1/mysqlversions/5.7.29
  uid: 6c595408-32a4-4815-9b18-fb26cc6f9045
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
    image: kubedb/mysql-replication-mode-detector:v0.1.0-beta.1
  tools:
    image: kubedb/mysql-tools:5.7.25
  upgradeConstraints:
    denylist:
      groupReplication:
      - < 5.7.29
      standalone:
      - < 5.7.29
  version: 5.7.29
```

The above `spec.upgradeConstraints.denylist` of `5.7.29` is showing that upgrading below version of `5.7.29` is not possible for both group replication and standalone. That means, it is possible to upgrade any version above `5.7.29`. Here, we are going to create a `MySQL` Group Replication using MySQL  `5.7.29`. Then we are going to upgrade this version to `5.7.31`.

**Deploy MySQL Group Replication:**

In this section, we are going to deploy a MySQL group replication with 3 members. Then, in the next section we will upgrade the version of the  members using upgrading. Below is the YAML of the `MySQL` cr that we are going to create,

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

Let's create the `MySQL` cr we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/upgrading/minorversion/group_replication.yaml
mysql.kubedb.com/my-group created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc. A secret called `my-group-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
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
my-group-0   2/2     Running   0          11m
my-group-1   2/2     Running   0          9m53s
my-group-2   2/2     Running   0          6m48s
```

Let's verify the `MySQL`, the `StatefulSet` and its `Pod` image version,

```console
$ kubectl get my -n demo my-group -o=jsonpath='{.spec.version}{"\n"}'
5.7.29

$ kubectl get sts -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.template.spec.containers[1].image'
"kubedb/my:5.7.29"

$ kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.containers[1].image'
"kubedb/my:5.7.29"
"kubedb/my:5.7.29"
"kubedb/my:5.7.29"
```

Let's also verify that the StatefulSet’s pods have joined into the group replication,

```console
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
sWfUMoqRpOJyomgb

$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=sWfUMoqRpOJyomgb --host=my-group-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| group_replication_applier | 356c03c0-b9cb-11ea-b856-7e6de479ee9d | my-group-1.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | 5be97bc0-b9cb-11ea-8b8b-a2eec7faa37d | my-group-2.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | c7089cb3-b9ca-11ea-b92a-228a3699132f | my-group-0.my-group-gvr.demo |        3306 | ONLINE       |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
```

We are ready to apply upgrading on this `MySQL` group replication.

#### Upgrade

Here, we are going to upgrade the `MySQL` group replication from `5.7.29` to `5.7.31`.

**Create MySQLOpsRequest:**

To upgrade your database cluster, you have to create a `MySQLOpsRequest` cr with your desired version that supported by `KubeDB`. Below is the YAML of the `MySQLOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-upgrade-minor-group
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: my-group
  upgrade:
    targetVersion: "5.7.31"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` MySQL database.
- `spec.type` specifies that we are going performing `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies expected version `5.7.31` after upgrading.

Let's create the `MySQLOpsRequest` cr we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/upgrading/minorversion/upgrade_minor_version_group.yaml
mysqlopsrequest.ops.kubedb.com/my-upgrade-minor-group created
```

**Verify MySQL version upgraded successfully:**

If everything goes well, `KubeDB` enterprise operator will update the image of `MySQL`, `StatefulSet`, and its `Pod`.

At first, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```console
$ watch -n 3 kubectl get myops -n demo my-upgrade-minor-group
Every 3.0s: kubectl get myops -n demo my-upgrade-minor-group     suaas-appscode: Sat Jul 25 21:41:39 2020

NAME                     TYPE      STATUS       AGE
my-upgrade-minor-group   Upgrade   Successful   5m26s
```

You can see from the above output that the `MySQLOpsRequest` has succeeded. If you describe the `MySQLOpsRequest` you will see that the `MySQL` group replication is updated with the new version and the `StatefulSet` is created with a new image.

```console
$ kubectl describe myops -n demo my-upgrade-minor-group
Name:         my-upgrade-minor-group
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2020-08-12T08:34:33Z
  Finalizers:
    mysql.ops.kubedb.com
  Generation:  2
  ...
  Resource Version:  45275
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/my-upgrade-minor-group
  UID:               915f0964-cb63-41ed-ad82-b6b799cfcc45
Spec:
  Database Ref:
    Name:                my-group
  Stateful Set Ordinal:  0
  Type:                  Upgrade
  Upgrade:
    Target Version:  5.7.31
Status:
  Conditions:
    Last Transition Time:  2020-08-12T08:34:33Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-upgrade-minor-group
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-08-12T08:34:33Z
    Message:               Controller has successfully Paused the MySQL database: demo/my-group 
    Observed Generation:   1
    Reason:                SuccessfullyPausedDatabase
    Status:                True
    Type:                  PauseDatabase
    Last Transition Time:  2020-08-12T08:34:33Z
    Message:               MySQL version upgrading stated for MySQLOpsRequest: demo/my-upgrade-minor-group
    Observed Generation:   1
    Reason:                DatabaseVersionUpgradingStarted
    Status:                True
    Type:                  Upgrading
    Last Transition Time:  2020-08-12T08:42:23Z
    Message:               Image successfully updated in MySQL: demo/my-group for MySQLOpsRequest: my-upgrade-minor-group 
    Observed Generation:   1
    Reason:                SuccessfullyUpgradedDatabaseVersion
    Status:                True
    Type:                  UpgradeVersion
    Last Transition Time:  2020-08-12T08:42:23Z
    Message:               Controller has successfully Resumed the MySQL database: demo/my-group
    Observed Generation:   2
    Reason:                SuccessfullyResumedDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-12T08:42:23Z
    Message:               Controller has successfully scaled/upgraded the MySQL demo/my-upgrade-minor-group
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    7m59s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-upgrade-minor-group
  Normal  Starting    7m59s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-group
  Normal  Successful  7m59s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-group for MySQLOpsRequest: my-upgrade-minor-group
  Normal  Starting    7m59s  KubeDB Enterprise Operator  Upgrading MySQL images: demo/my-group for MySQLOpsRequest: my-upgrade-minor-group
  Normal  Successful  4m59s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  4m39s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  2m28s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  2m28s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  2m28s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  9s     KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-0
  Normal  Successful  9s     KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  9s     KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-group-1
  Normal  Successful  9s     KubeDB Enterprise Operator  Image successfully updated in MySQL: demo/my-group for MySQLOpsRequest: my-upgrade-minor-group
  Normal  Successful  9s     KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-0
  Normal  Successful  9s     KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  9s     KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-group-1
  Normal  Starting    9s     KubeDB Enterprise Operator  Resuming MySQL database: demo/my-group
  Normal  Successful  9s     KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-group
  Normal  Successful  9s     KubeDB Enterprise Operator  Controller has Successfully upgraded the version of MySQL : demo/my-group
```

Now, we are going to verify whether the `MySQL` and `StatefulSet` and it's `Pod` have updated with new image. Let's check,

```console
$ kubectl get my -n demo my-group -o=jsonpath='{.spec.version}{"\n"}'
5.7.31

$ kubectl get sts -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.template.spec.containers[1].image'
"kubedb/mysql:5.7.31"

$ kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group -o json | jq '.items[].spec.containers[1].image'
"kubedb/mysql:5.7.31"
"kubedb/mysql:5.7.31"
"kubedb/mysql:5.7.31"
```

Let's also check the StatefulSet pods have joined the `MySQL` group replication,

```console
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
sWfUMoqRpOJyomgb

$ kkubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=OUta0G6QnyBuKmi0 --host=my-group-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| group_replication_applier | 0f29bc2e-dc76-11ea-b66f-02f24f164817 | my-group-1.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | 3741c056-dc76-11ea-ac3e-120989b9c9d5 | my-group-2.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | e2ba6b64-dc75-11ea-93e3-f60ec5bc86b8 | my-group-0.my-group-gvr.demo |        3306 | ONLINE       |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
```

You can see above that our `MySQL` group replication now has updated members. It verifies that we have successfully upgraded our cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo my-group
kubectl delete myops -n demo my-upgrade-minor-group
```