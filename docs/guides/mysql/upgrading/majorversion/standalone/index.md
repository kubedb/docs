---
title: Upgrading MySQL standalone major version
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-upgrading-major-standalone
    name: Standalone
    parent:  guides-mysql-upgrading-major
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade major version of MySQL Standalone

This guide will show you how to use `KubeDB` enterprise operator to upgrade the major version of `MySQL` standalone.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/database/index.md)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest/index.md)
  - [Upgrading Overview](/docs/guides/mysql/upgrading/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mysql/upgrading/majorversion/standalone/yamls](/docs/guides/mysql/upgrading/majorversion/standalone/yamls) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Apply Version Upgrading on Standalone

Here, we are going to deploy a `MySQL` standalone using a supported version by `KubeDB` operator. Then we are going to apply upgrading on it.

#### Prepare Standalone

At first, we are going to deploy a standalone using supported that `MySQL` version whether it is possible to upgrade from this version to another. In the next two sections, we are going to find out the supported version and version upgrade constraints.

**Find supported MySQLVersion:**

When you have installed `KubeDB`, it has created `MySQLVersion` CR for all supported `MySQL` versions. Let's check support versions,

```bash
$ kubectl get mysqlversion
NAME        VERSION   DB_IMAGE                  DEPRECATED   AGE
5.7.25-v2   5.7.25    kubedb/mysql:5.7.25-v2                 128m
5.7.29-v1   5.7.29    suaas21/mysql:5.7.29-v1                128m
5.7.31-v1   5.7.31    suaas21/mysql:5.7.31-v1                128m
5.7.33      5.7.33    suaas21/mysql:5.7.33                   128m
8.0.14-v2   8.0.14    kubedb/mysql:8.0.14-v2                 128m
8.0.20-v1   8.0.20    kubedb/mysql:8.0.20-v1                 128m
8.0.21-v1   8.0.21    suaas21/mysql:8.0.21-v1                128m
8.0.23      8.0.23    kubedb/mysql:8.0.23                    128m
8.0.3-v2    8.0.3     kubedb/mysql:8.0.3-v2                  128m
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `MySQLVersion` for `MySQL` standalone that will be possible to upgrade from this version to another version. In the next section, we are going to verify version upgrade constraints.

**Check Upgrade Constraints:**

Database version upgrade constraints is a constraint that shows whether it is possible or not possible to upgrade from one version to another. Let's check the version upgrade constraints of `MySQL` `5.7.31-v1`,

```bash
$ kubectl get mysqlversion 5.7.31-v1 -o yaml | kubectl neat
apiVersion: catalog.kubedb.com/v1alpha1
kind: MySQLVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb-catalog
    meta.helm.sh/release-namespace: kube-system
  labels:
    app.kubernetes.io/instance: kubedb-catalog
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v0.16.2
    helm.sh/chart: kubedb-catalog-v0.16.2
  name: 5.7.31-v1
spec:
  db:
    image: suaas21/mysql:5.7.31-v1
  distribution: Oracle
  exporter:
    image: kubedb/mysqld-exporter:v0.11.0
  initContainer:
    image: kubedb/toybox:0.8.4
  podSecurityPolicies:
    databasePolicyName: mysql-db
  replicationModeDetector:
    image: kubedb/replication-mode-detector:v0.3.2
  stash:
    addon:
      backupTask:
        name: mysql-backup-5.7.25-v7
      restoreTask:
        name: mysql-restore-5.7.25-v7
  upgradeConstraints:
    denylist:
      groupReplication:
      - < 5.7.31
      standalone:
      - < 5.7.31
  version: 5.7.31
```

The above `spec.upgradeConstraints.denylist` is showing that upgrading below version of `5.7.31-v1` is not possible for both standalone and group replication. That means, it is possible to upgrade any version above `5.7.31-v1`. Here, we are going to create a `MySQL` standalone using MySQL  `5.7.31-v1`. Then we are going to upgrade this version to `8.0.21-v1`.

**Deploy MySQL standalone:**

In this section, we are going to deploy a MySQL standalone. Then, in the next section, we will upgrade the version of the database using upgrading. Below is the YAML of the `MySQL` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-standalone
  namespace: demo
spec:
  version: "5.7.33"
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/upgrading/majorversion/standalone/yamls/standalone.yaml
mysql.kubedb.com/my-standalone created
```

**Wait for the database to be ready:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc. A secret called `my-standalone-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
Now, watch `MySQL` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get my -n demo my-standalone
Every 3.0s: kubectl get my -n demo my-standalone                 suaas-appscode: Thu Jun 18 01:28:43 2020

NAME            VERSION      STATUS    AGE
my-standalone   5.7.31-v1    Running   3m

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

```bash
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
5.7.31-v1

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.31-v1

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.31-v1
```

We are ready to apply upgrading on this `MySQL` standalone.

#### Upgrade

Here, we are going to upgrade `MySQL` standalone from `5.7.31-v1` to `8.0.21-v1`.

**Create MySQLOpsRequest:**

To upgrade the standalone, you have to create a `MySQLOpsRequest` cr with your desired version that supported by `KubeDB`. Below is the YAML of the `MySQLOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-upgrade-major-standalone
  namespace: demo
spec:
  databaseRef:
    name: my-standalone
  type: Upgrade
  upgrade:
    targetVersion: "8.0.21-v1"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` MySQL database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies expected version `8.0.21-v1` after upgrading.

Let's create the `MySQLOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/upgrading/majorversion/standalone/yamls/upgrade_major_version_standalone.yaml
mysqlopsrequest.ops.kubedb.com/my-upgrade-major-standalone created
```

**Verify MySQL version upgraded successfully:**

If everything goes well, `KubeDB` enterprise operator will update the image of `MySQL`, `StatefulSet`, and its `Pod`.

At first, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```bash
$ watch -n 3 kubectl get myops -n demo my-upgrade-major-standalone
Every 3.0s: kubectl get myops -n demo my-up...  suaas-appscode: Wed Aug 12 16:12:03 2020

NAME                          TYPE      STATUS       AGE
my-upgrade-major-standalone   Upgrade   Successful   3m57s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest`, we shall see that the `MySQL`, `StatefulSet`, and its `Pod` have updated with a new image.

```bash
$ kubectl describe myops -n demo my-upgrade-major-standalone
Name:         my-upgrade-major-standalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2021-03-10T08:38:48Z
  Generation:          2
    Operation:    Update
    Time:         2021-03-10T08:38:48Z
    API Version:  ops.kubedb.com/v1alpha1
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-10T08:40:23Z
  Resource Version:  1055436
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/my-upgrade-major-standalone
  UID:               d8a43a2e-3e75-4b9f-be67-ffd95cff54b3
Spec:
  Database Ref:
    Name:                my-standalone
  Stateful Set Ordinal:  0
  Type:                  Upgrade
  Upgrade:
    Target Version:  8.0.21-v1
Status:
  Conditions:
    Last Transition Time:  2021-03-10T08:38:48Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-upgrade-major-standalone
    Observed Generation:   2
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2021-03-10T08:38:48Z
    Message:               MySQL version UpgradeFunc stated for MySQLOpsRequest: demo/my-upgrade-major-standalone
    Observed Generation:   2
    Reason:                DatabaseVersionUpgradingStarted
    Status:                True
    Type:                  Upgrading
    Last Transition Time:  2021-03-10T08:40:23Z
    Message:               Image successfully updated in MySQL: demo/my-standalone for MySQLOpsRequest: my-upgrade-major-standalone 
    Observed Generation:   2
    Reason:                SuccessfullyUpgradedDatabaseVersion
    Status:                True
    Type:                  UpgradeVersion
    Last Transition Time:  2021-03-10T08:40:23Z
    Message:               Controller has successfully upgraded the MySQL demo/my-upgrade-major-standalone
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    2m19s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-upgrade-major-standalone
  Normal  Starting    2m19s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-standalone
  Normal  Successful  2m19s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-standalone for MySQLOpsRequest: my-upgrade-major-standalone
  Normal  Starting    2m19s  KubeDB Enterprise Operator  Upgrading MySQL images: demo/my-standalone for MySQLOpsRequest: my-upgrade-major-standalone
  Normal  Starting    2m14s  KubeDB Enterprise Operator  Restarting Pod (master): demo/my-standalone-0
  Normal  Successful  44s    KubeDB Enterprise Operator  Image successfully updated in MySQL: demo/my-standalone for MySQLOpsRequest: my-upgrade-major-standalone
  Normal  Starting    44s    KubeDB Enterprise Operator  Resuming MySQL database: demo/my-standalone
  Normal  Successful  44s    KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-standalone
  Normal  Successful  44s    KubeDB Enterprise Operator  Controller has Successfully upgraded the version of MySQL : demo/my-standalone
```

Now, we are going to verify whether the `MySQL`, `StatefulSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
8.0.21-v1

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:8.0.21-v1

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:8.0.21-v1
```

You can see above that our `MySQL`standalone has been updated with the new version. It verifies that we have successfully upgraded our standalone.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo my-standalone
kubectl delete myops -n demo my-upgrade-major-standalone
```