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
NAME            VERSION   DISTRIBUTION   DB_IMAGE                    DEPRECATED   AGE
5.7.35-v1       5.7.35    Official       mysql:5.7.35                             13d
5.7.36          5.7.36    Official       mysql:5.7.36                             13d
8.0.17          8.0.17    Official       mysql:8.0.17                             13d
8.0.27          8.0.27    Official       mysql:8.0.27                             13d
8.0.27-innodb   8.0.27    MySQL          mysql/mysql-server:8.0.27                13d
8.0.29          8.0.29    Official       mysql:8.0.29                             13d
8.0.3-v4        8.0.3     Official       mysql:8.0.3                              13d
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `MySQLVersion` for `MySQL` standalone that will be possible to upgrade from this version to another version. In the next section, we are going to verify version upgrade constraints.

**Check Upgrade Constraints:**

Database version upgrade constraints is a constraint that shows whether it is possible or not possible to upgrade from one version to another. Let's check the version upgrade constraints of `MySQL` `5.7.36`,

```bash
$ kubectl get mysqlversion 5.7.36 -o yaml | kubectl neat
apiVersion: catalog.kubedb.com/v1alpha1
kind: MySQLVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb-catalog
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2022-06-16T13:52:58Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb-catalog
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2022.03.28
    helm.sh/chart: kubedb-catalog-v2022.03.28
  name: 5.7.36
  resourceVersion: "1092465"
  uid: 4cc87fc8-efd7-4e69-bb12-4454a2b1bf06
spec:
  coordinator:
    image: kubedb/mysql-coordinator:v0.5.0
  db:
    image: mysql:5.7.36
  distribution: Official
  exporter:
    image: kubedb/mysqld-exporter:v0.13.1
  initContainer:
    image: kubedb/mysql-init:5.7-v2
  podSecurityPolicies:
    databasePolicyName: mysql-db
  replicationModeDetector:
    image: kubedb/replication-mode-detector:v0.13.0
  stash:
    addon:
      backupTask:
        name: mysql-backup-5.7.25
      restoreTask:
        name: mysql-restore-5.7.25
  upgradeConstraints:
    denylist:
      groupReplication:
      - < 5.7.36
      standalone:
      - < 5.7.36
  version: 5.7.36

```

The above `spec.upgradeConstraints.denylist` is showing that upgrading below version of `5.7.36` is not possible for both standalone and group replication. That means, it is possible to upgrade any version above `5.7.36`. Here, we are going to create a `MySQL` standalone using MySQL  `5.7.36`. Then we are going to upgrade this version to `8.0.29`.

**Deploy MySQL standalone:**

In this section, we are going to deploy a MySQL standalone. Then, in the next section, we will upgrade the version of the database using upgrading. Below is the YAML of the `MySQL` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-standalone
  namespace: demo
spec:
  version: "5.7.36"
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

NAME            VERSION      STATUS    AGE
my-standalone   5.7.36    Running   3m

$ watch -n 3 kubectl get sts -n demo my-standalone

NAME            READY   AGE
my-standalone   1/1     3m42s

$ watch -n 3 kubectl get pod -n demo my-standalone-0

NAME              READY   STATUS    RESTARTS   AGE
my-standalone-0   1/1     Running   0          5m23s
```

Let's verify the `MySQL`, the `StatefulSet` and its `Pod` image version,

```bash
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
5.7.36

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.36

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.36
```

We are ready to apply upgrading on this `MySQL` standalone.

#### Upgrade

Here, we are going to upgrade `MySQL` standalone from `5.7.36` to `8.0.29`.

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
  type: UpdateVersion
  upgrade:
    targetVersion: "8.0.29"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` MySQL database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies expected version `8.0.29` after upgrading.

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
  Creation Timestamp:  2022-06-30T07:55:16Z
    Manager:         kubedb-ops-manager
    Operation:       Update
    Time:            2022-06-30T07:55:16Z
  Resource Version:  1708721
  UID:               460319fc-8dc4-45d7-8958-fa84e24d5f51
Spec:
  Database Ref:
    Name:  my-standalone
  Type:    Upgrade
  Upgrade:
    Target Version:  8.0.29
Status:
  Conditions:
    Last Transition Time:  2022-06-30T07:55:16Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-upgrade-major-standalone
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-06-30T07:55:16Z
    Message:               MySQL version UpgradeFunc stated for MySQLOpsRequest: demo/my-upgrade-major-standalone
    Observed Generation:   1
    Reason:                DatabaseVersionUpgradingStarted
    Status:                True
    Type:                  Upgrading
    Last Transition Time:  2022-06-30T07:59:16Z
    Message:               Image successfully updated in MySQL: demo/my-standalone for MySQLOpsRequest: my-upgrade-major-standalone 
    Observed Generation:   1
    Reason:                SuccessfullyUpgradedDatabaseVersion
    Status:                True
    Type:                  UpgradeVersion
    Last Transition Time:  2022-06-30T07:59:16Z
    Message:               Controller has successfully upgraded the MySQL demo/my-upgrade-major-standalone
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    8m47s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-upgrade-major-standalone
  Normal  Starting    8m47s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-standalone
  Normal  Successful  8m47s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-standalone for MySQLOpsRequest: my-upgrade-major-standalone
  Normal  Starting    6m2s   KubeDB Enterprise Operator  Restarting Pod: my-standalone-0/demo
  Normal  Successful  4m47s  KubeDB Enterprise Operator  Image successfully updated in MySQL: demo/my-standalone for MySQLOpsRequest: my-upgrade-major-standalone
  Normal  Starting    4m47s  KubeDB Enterprise Operator  Resuming MySQL database: demo/my-standalone
  Normal  Successful  4m47s  KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-standalone
  Normal  Successful  4m47s  KubeDB Enterprise Operator  Controller has Successfully upgraded the version of MySQL : demo/my-standalone
```

Now, we are going to verify whether the `MySQL`, `StatefulSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
8.0.29

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mysql:8.0.29

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mysql:8.0.29
```

You can see above that our `MySQL`standalone has been updated with the new version. It verifies that we have successfully upgraded our standalone.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo my-standalone
kubectl delete myops -n demo my-upgrade-major-standalone
```