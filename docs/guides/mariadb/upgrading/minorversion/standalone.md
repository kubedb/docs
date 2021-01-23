---
title: Upgrading MariaDB standalone minor version
menu:
  docs_{{ .version }}:
    identifier: my-upgrading-mariadb-minor-standalone
    name: Standalone
    parent:  my-upgrading-mariadb-minor
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade minor version of MariaDB Standalone

This guide will show you how to use `KubeDB` enterprise operator to upgrade the minor version of `MariaDB` standalone.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb.md)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest.md)
  - [Upgrading Overview](/docs/guides/mariadb/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mariadb](/docs/examples/mariadb) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Apply Version Upgrading on Standalone

Here, we are going to deploy a `MariaDB` standalone using a supported version by `KubeDB` operator. Then we are going to apply upgrading on it.

#### Prepare Group Replication

At first, we are going to deploy a standalone using supported that `MariaDB` version whether it is possible to upgrade from this version to another. In the next two sections, we are going to find out the supported version and version upgrade constraints.

**Find supported MariaDBVersion:**

When you have installed `KubeDB`, it has created `MariaDBVersion` CR for all supported `MariaDB` versions. Let's check support versions,

```bash
$ kubectl get mariadbversion
NAME        VERSION   DB_IMAGE                 DEPRECATED   AGE
5           5         kubedb/mariadb:5           true         149m
5-v1        5         kubedb/mariadb:5-v1        true         149m
5.7         5.7       kubedb/mariadb:5.7         true         149m
5.7-v1      5.7       kubedb/mariadb:5.7-v1      true         149m
5.7-v2      5.7.25    kubedb/mariadb:5.7-v2      true         149m
5.7-v3      5.7.25    kubedb/mariadb:5.7.25      true         149m
5.7-v4      5.7.29    kubedb/mariadb:5.7.29      true         149m
5.7.25      5.7.25    kubedb/mariadb:5.7.25      true         149m
5.7.25-v1   5.7.25    kubedb/mariadb:5.7.25-v1                149m
5.7.29      5.7.29    kubedb/mariadb:5.7.29                   149m
5.7.31      5.7.31    kubedb/mariadb:5.7.31                   149m
8           8         kubedb/mariadb:8           true         149m
8-v1        8         kubedb/mariadb:8-v1        true         149m
8.0         8.0       kubedb/mariadb:8.0         true         149m
8.0-v1      8.0.3     kubedb/mariadb:8.0-v1      true         149m
8.0-v2      8.0.14    kubedb/mariadb:8.0-v2      true         149m
8.0-v3      8.0.20    kubedb/mariadb:8.0.20      true         149m
8.0.14      8.0.14    kubedb/mariadb:8.0.14      true         149m
8.0.14-v1   8.0.14    kubedb/mariadb:8.0.14-v1                149m
8.0.20      8.0.20    kubedb/mariadb:8.0.20                   149m
8.0.21      8.0.21    kubedb/mariadb:8.0.21                   149m
8.0.3       8.0.3     kubedb/mariadb:8.0.3       true         149m
8.0.3-v1    8.0.3     kubedb/mariadb:8.0.3-v1                 149m
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MariaDB`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `MariaDBVersion` for `MariaDB` standalone that will be possible to upgrade from this version to another version. In the next section, we are going to verify version upgrade constraints.

**Check Upgrade Constraints:**

Database version upgrade constraints is a constraint that shows whether it is possible or not possible to upgrade from one version to another. Let's check the version upgrade constraints of `MariaDB` `5.7.29`,

```bash
$ kubectl get mariadbversion 5.7.29 -o yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MariaDBVersion
metadata:
  name: 5.7.29
  ...
spec:
  db:
    image: kubedb/mariadb:5.7.29
  exporter:
    image: kubedb/mariadbd-exporter:v0.11.0
  initContainer:
    image: kubedb/busybox
  podSecurityPolicies:
    databasePolicyName: mariadb-db
  replicationModeDetector:
    image: kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1
  tools:
    image: kubedb/mariadb-tools:5.7.25
  upgradeConstraints:
    denylist:
      groupReplication:
      - < 5.7.29
      standalone:
      - < 5.7.29
  version: 5.7.29
```

The above `spec.upgradeConstraints.denylist` is showing that upgrading below version of `5.7.29` is not possible for both standalone and group replication. That means, it is possible to upgrade any version above `5.7.29`. Here, we are going to create a `MariaDB` standalone using MariaDB  `5.7.29`. Then we are going to upgrade this version to `5.7.31`.

#### Prepare Standalone

Now, we are going to deploy a `MariaDB` standalone using version `5.7.29`.

**Deploy MariaDB standalone:**

In this section, we are going to deploy a MariaDB standalone. Then, in the next section, we will upgrade the version of the database using upgrading. Below is the YAML of the `MariaDB` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
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

Let's create the `MariaDB` cr we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/upgrading/minorversin/standalone.yaml
mariadb.kubedb.com/my-standalone created
```

**Wait for the database to be ready:**

`KubeDB` operator watches for `MariaDB` objects using Kubernetes API. When a `MariaDB` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc. A secret called `my-standalone-auth` (format: <em>{mariadb-object-name}-auth</em>) will be created storing the password for mariadb superuser.
Now, watch `MariaDB` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
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

Let's verify the `MariaDB`, the `StatefulSet` and its `Pod` image version,

```bash
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
5.7.29

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.29

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.29
```

We are ready to apply upgrading on this `MariaDB` standalone.

#### Upgrade

Here, we are going to upgrade `MariaDB` standalone from `5.7.29` to `5.7.31`.

**Create MariaDBOpsRequest:**

To upgrade the standalone, you have to create a `MariaDBOpsRequest` cr with your desired version that supported by `KubeDB`. Below is the YAML of the `MariaDBOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
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

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` MariaDB database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies expected version `5.7.31` after upgrading.

Let's create the `MariaDBOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/upgrading/minorversion/upgrade_minor_version_standalone.yaml
mariadbopsrequest.ops.kubedb.com/my-upgrade-minor-standalone created
```

**Verify MariaDB version upgraded successfully:**

If everything goes well, `KubeDB` enterprise operator will update the image of `MariaDB`, `StatefulSet`, and its `Pod`.

At first, we will wait for `MariaDBOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```bash
$ watch -n 3 kubectl get myops -n demo my-upgrade-minor-standalone
Every 3.0s: kubectl get myops -n demo my-up...  suaas-appscode: Wed Aug 12 16:12:03 2020

NAME                          TYPE      STATUS       AGE
my-upgrade-minor-standalone   Upgrade   Successful   3m57s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest`, we shall see that the `MariaDB`, `StatefulSet`, and its `Pod` have updated with a new image.

```bash
$ kubectl describe myops -n demo my-upgrade-minor-standalone
Name:         my-upgrade-minor-standalone
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2020-08-12T10:08:06Z
  Finalizers:
    mariadb.ops.kubedb.com
  Generation:  2
  ...
  Resource Version:  61103
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mariadbopsrequests/my-upgrade-minor-standalone
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
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/my-upgrade-minor-standalone
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-08-12T10:08:06Z
    Message:               Controller has successfully Halted the MariaDB database: demo/my-standalone 
    Observed Generation:   1
    Reason:                SuccessfullyHaltedDatabase
    Status:                True
    Type:                  HaltDatabase
    Last Transition Time:  2020-08-12T10:08:06Z
    Message:               MariaDB version upgrading stated for MariaDBOpsRequest: demo/my-upgrade-minor-standalone
    Observed Generation:   1
    Reason:                DatabaseVersionUpgradingStarted
    Status:                True
    Type:                  Upgrading
    Last Transition Time:  2020-08-12T10:10:06Z
    Message:               Image successfully updated in MariaDB: demo/my-standalone for MariaDBOpsRequest: my-upgrade-minor-standalone 
    Observed Generation:   1
    Reason:                SuccessfullyUpgradedDatabaseVersion
    Status:                True
    Type:                  UpgradeVersion
    Last Transition Time:  2020-08-12T10:10:06Z
    Message:               Controller has successfully Resumed the MariaDB database: demo/my-standalone
    Observed Generation:   2
    Reason:                SuccessfullyResumedDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-12T10:10:06Z
    Message:               Controller has successfully scaled/upgraded the MariaDB demo/my-upgrade-minor-standalone
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    4m48s  KubeDB Enterprise Operator  Start processing for MariaDBOpsRequest: demo/my-upgrade-minor-standalone
  Normal  Starting    4m48s  KubeDB Enterprise Operator  Pausing MariaDB databse: demo/my-standalone
  Normal  Successful  4m48s  KubeDB Enterprise Operator  Successfully halted MariaDB database: demo/my-standalone for MariaDBOpsRequest: my-upgrade-minor-standalone
  Normal  Starting    4m48s  KubeDB Enterprise Operator  Upgrading MariaDB images: demo/my-standalone for MariaDBOpsRequest: my-upgrade-minor-standalone
  Normal  Successful  3m8s   KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-standalone-0
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Image successfully updated in MariaDB: demo/my-standalone for MariaDBOpsRequest: my-upgrade-minor-standalone
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-standalone-0
  Normal  Starting    2m48s  KubeDB Enterprise Operator  Resuming MariaDB database: demo/my-standalone
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Successfully resumed MariaDB database: demo/my-standalone
  Normal  Successful  2m48s  KubeDB Enterprise Operator  Controller has Successfully upgraded the version of MariaDB : demo/my-standalone
```

Now, we are going to verify whether the `MariaDB`, `StatefulSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
5.7.31

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.31

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.31
```

You can see above that our `MariaDB`standalone has been updated with the new version. It verifies that we have successfully upgraded our standalone.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo my-standalone
kubectl delete myops -n demo my-upgrade-minor-standalone
```