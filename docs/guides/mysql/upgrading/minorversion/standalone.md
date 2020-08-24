---
title: Upgrading MySQL standalone minor version
menu:
  docs_{{ .version }}:
    identifier: my-upgrading-mysql-minor-standalone
    name: Standalone
    parent:  my-upgrading-mysql-minor
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="Upgrading is an Enterprise feature of KubeDB. You must have a KubeDB Enterprise operator installed to test this feature." >}}

# Upgrade minor version of MySQL Standalone

This guide will show you how to use `KubeDB` enterprise operator to upgrade the minor version of `MySQL` standalone.

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

> **Note:** YAML files used in this tutorial are stored in [docs/examples/day-2-operations/mysql](/docs/examples/day-2-operations/mysql) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Apply Version Upgrading on Standalone

Here, we are going to deploy a `MySQL` standalone using a supported version by `KubeDB` operator. Then we are going to apply upgrading on it.

#### Prepare Group Replication

At first, we are going to deploy a standalone using supported that `MySQL` version whether it is possible to upgrade from this version to another. In the next two sections, we are going to find out the supported version and version upgrade constraints.

**Find supported MySQLVersion:**

When you have installed `KubeDB`, it has created `MySQLVersion` cr for all supported `MySQL` versions. Let's check support versions,

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

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `MySQLVersion` for `MySQL` standalone that will be possible to upgrade from this version to another version. In the next section, we are going to verify version upgrade constraints.

**Check Upgrade Constraints:**

Database version upgrade constraints is a constraint that shows whether it is possible or not possible to upgrade from one version to another. Let's check the version upgrade constraints of `MySQL` `5.7.29`,

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

The above `spec.upgradeConstraints.denylist` is showing that upgrading below version of `5.7.29` is not possible for both standalone and group replication. That means, it is possible to upgrade any version above `5.7.29`. Here, we are going to create a `MySQL` standalone using MySQL  `5.7.29`. Then we are going to upgrade this version to `5.7.31`.

#### Prepare Standalone

Now, we are going to deploy a `MySQL` standalone using version `5.7.29`.

**Deploy MySQL standalone:**

In this section, we are going to deploy a MySQL standalone. Then, in the next section, we will upgrade the version of the database using upgrading. Below is the YAML of the `MySQL` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-standalone
  namespace: demo
spec:
  version: "5.7.29"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/upgrading/minorversin/standalone.yaml
mysql.kubedb.com/my-standalone created
```

**Wait for the database to be ready:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc. A secret called `my-standalone-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
Now, watch `MySQL` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```console
$ watch -n 3 kubectl get my -n demo my-standalone
Every 3.0s: kubectl get my -n demo my-standalone                 suaas-appscode: Thu Jun 18 01:28:43 2020

NAME            VERSION   STATUS    AGE
my-standalone   5.7.29    Running   3m

$ watch -n 3 kubectl get sts -n demo my-standalone
Every 3.0s: kubectl get sts -n demo my-standalone                suaas-appscode: Thu Jun 18 12:57:42 2020

NAME            READY   AGE
my-standalone   1/1     3m42s

$ watch -n 3 kubectl get pod -n demo my-standalone-0
Every 3.0s: kubectl get pod -n demo my-standalone-0              suaas-appscode: Thu Jun 18 12:56:23 2020

NAME              READY   STATUS    RESTARTS   AGE
my-standalone-0   1/1     Running   0          5m23s
```

Let's verify the `MySQL`, the `StatefulSet` and its `Pod` image version,

```console
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
5.7.29

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.29

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.29
```

We are ready to apply upgrading on this `MySQL` standalone.

#### Upgrade

Here, we are going to upgrade `MySQL` standalone from `5.7.29` to `5.7.31`.

**Create MySQLOpsRequest:**

To upgrade the standalone, you have to create a `MySQLOpsRequest` cr with your desired version that supported by `KubeDB`. Below is the YAML of the `MySQLOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-upgrade-minor-standalone
  namespace: demo
spec:
  databaseRef:
    name: my-standalone
  type: Upgrade
  upgrade:
    targetVersion: "5.7.31"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` MySQL database.
- `spec.type` specifies that we are going performing `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies expected version `5.7.31` after upgrading.

Let's create the `MySQLOpsRequest` cr we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/upgrading/minorversion/upgrade_minor_version_standalone.yaml
mysqlopsrequest.ops.kubedb.com/my-upgrade-minor-standalone created
```

**Verify MySQL version upgraded successfully:**

If everything goes well, `KubeDB` enterprise operator will update the image of `MySQL`, `StatefulSet`, and its `Pod`.

At first, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```console
$ watch -n 3 kubectl get myops -n demo my-upgrade-minor-standalone
Every 3.0s: kubectl get myops -n demo my-up...  suaas-appscode: Wed Aug 12 16:12:03 2020

NAME                          TYPE      STATUS       AGE
my-upgrade-minor-standalone   Upgrade   Successful   3m57s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If you describe the `MySQLOpsRequest` you will see that the `MySQL`, `StatefulSet`, and its `Pod` have updated with a new image.

```console
$ kubectl describe myops -n demo my-upgrade-minor-standalone
Name:         my-upgrade-minor-standalone
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2020-08-12T10:08:06Z
  Finalizers:
    mysql.ops.kubedb.com
  Generation:  2
  ...
  Resource Version:  61103
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/my-upgrade-minor-standalone
  UID:               03ea52ac-5b74-47b9-bfba-426f386e8a34
Spec:
  Database Ref:
    Name:                my-standalone
  Stateful Set Ordinal:  0
  Type:                  Upgrade
  Upgrade:
    Target Version:  5.7.31
Status:
  Conditions:
    Last Transition Time:  2020-08-12T10:08:06Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-upgrade-minor-standalone
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-08-12T10:08:06Z
    Message:               Controller has successfully Paused the MySQL database: demo/my-standalone 
    Observed Generation:   1
    Reason:                SuccessfullyPausedDatabase
    Status:                True
    Type:                  PauseDatabase
    Last Transition Time:  2020-08-12T10:08:06Z
    Message:               MySQL version upgrading stated for MySQLOpsRequest: demo/my-upgrade-minor-standalone
    Observed Generation:   1
    Reason:                DatabaseVersionUpgradingStarted
    Status:                True
    Type:                  Upgrading
    Last Transition Time:  2020-08-12T10:10:06Z
    Message:               Image successfully updated in MySQL: demo/my-standalone for MySQLOpsRequest: my-upgrade-minor-standalone 
    Observed Generation:   1
    Reason:                SuccessfullyUpgradedDatabaseVersion
    Status:                True
    Type:                  UpgradeVersion
    Last Transition Time:  2020-08-12T10:10:06Z
    Message:               Controller has successfully Resumed the MySQL database: demo/my-standalone
    Observed Generation:   2
    Reason:                SuccessfullyResumedDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-12T10:10:06Z
    Message:               Controller has successfully scaled/upgraded the MySQL demo/my-upgrade-minor-standalone
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    4m48s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-upgrade-minor-standalone
  Normal  Starting    4m48s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-standalone
  Normal  Successful  4m48s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-standalone for MySQLOpsRequest: my-upgrade-minor-standalone
  Normal  Starting    4m48s  KubeDB Enterprise Operator  Upgrading MySQL images: demo/my-standalone for MySQLOpsRequest: my-upgrade-minor-standalone
  Normal  Successful  3m8s   KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-standalone-0
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Image successfully updated in MySQL: demo/my-standalone for MySQLOpsRequest: my-upgrade-minor-standalone
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-standalone-0
  Normal  Starting    2m48s  KubeDB Enterprise Operator  Resuming MySQL database: demo/my-standalone
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-standalone
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Controller has Successfully upgraded the version of MySQL : demo/my-standalone
```

Now, we are going to verify whether the `MySQL`, `StatefulSet` and it's `Pod` have updated with new image. Let's check,

```console
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
5.7.31

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.31

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.31
```

You can see above that our `MySQL`standalone has been updated with the new version. It verifies that we have successfully upgraded our standalone.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo my-standalone
kubectl delete myops -n demo my-upgrade-minor-standalone
```