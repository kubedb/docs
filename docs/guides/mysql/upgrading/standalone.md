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

# Upgrade MySQL Standalone

This guide will show you how to use `KubeDB` enterprise operator to upgrade the `MySQL` Standalone.

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

### Upgrade MySQL Standalone

Here, we are going to deploy a  `MySQL` standalone using a supported version by `KubeDB` operator. Below two sections will check the supported `MySQL` versions and then check whether it is possible to upgrade from this version to another.

**Find supported MySQLVersion:**

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

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. Now we will select a version from `MySQLVersion` for `MySQL` standalone that will be possible to upgrade from this version to another. For `MySQL` standalone deployment, we will select version `5.7.29`. The below section will check this version's upgrade constraints.

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

The above `spec.upgradeConstraints.blacklist` is showing that upgrading below version `5.7.29` is not possible for both standalone and group replication. That means, it is possible to upgrade any version above `5.7.29`. The below section will describe deploying `MySQL` standalone using version `5.7.29`.

#### Prepare Standalone

Now, we are going to deploy a `MySQL` standalone using version `5.7.29`.

**Create MySQL Object:**

Below is the YAML of the `MySQL` crd that we are going to create,

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

Let's create the `MySQL` crd we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/standalone.yaml
mysql.kubedb.com/my-standalone created
```

**Check MySQL Ready to Upgrade:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services and Secrets etc. A secret called `my-standalone-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
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

Let's check the `MySQL`, `StatefulSet` and its `Pod` image version,

```console
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
5.7.29

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.29

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:5.7.29
```

We are ready to upgrade the above `MySQL` standalone.

#### Upgrade

Here, we are going to upgrade `MySQL` standalone from `5.7.29` to `8.0.20`.

**Create MySQLOpsRequest:**

Below is the YAML of the `MySQLOpsRequest` crd that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myopsreq-standalone
  namespace: demo
spec:
  databaseRef:
    name: my-standalone
  type: Upgrade
  upgrade:
    targetVersion: "8.0.20"
```

Here,

- `spec.databaseRef.name` refers to the `my-standalone` MySQL object for operation.
- `spec.type` specifies that this is an `Upgrade` type operation
- `spec.upgrade.targetVersion` specifies version `8.0.20` that will be upgraded from version `5.7.29`.

Let's create the `MySQLOpsRequest` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/upgrade_standalone.yaml
mysqlopsrequest.ops.kubedb.com/myopsreq-standalone created
```

**Check MySQL version upgraded:**

If everything goes well, `KubeDB` enterprise operator will update the images of `MySQL`, `StatefulSet` and its `Pod`.

First, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` crd,

```console
$ kubectl get myops -n demo myopsreq-standalone
Every 3.0s: kubectl get myops -n demo myopsreq-standalone                         suaas-appscode: Thu Jun 18 10:34:24 2020

NAME                  TYPE      STATUS       AGE
myopsreq-standalone   Upgrade   Successful   4m26s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If you describe the `MySQLOpsRequest` you will see that the `MySQL`, `StatefulSet`, and its `Pod` have updated with new images.

```console
$ kubectl describe myops -n demo myopsreq-standalone
Name:         myopsreq-standalone
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2020-06-18T04:29:58Z
  Finalizers:
    mysql.ops.kubedb.com
  Generation:        2
  Resource Version:  3745
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/myopsreq-standalone
  UID:               fc425394-d545-45bb-90e3-9ba799623105
Spec:
  Database Ref:
    Name:                my-standalone
  Stateful Set Ordinal:  0
  Type:                  Upgrade
  Upgrade:
    Target Version:  8.0.20
Status:
  Conditions:
    Last Transition Time:  2020-06-18T04:29:58Z
    Message:               The controller has started to Progress the OpsRequest
    Observed Generation:   1
    Reason:                OpsRequestProgressing
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-06-18T04:29:58Z
    Message:               MySQLOpsRequestDefinition: myopsreq-standalone for Pausing MySQL: demo/my-standalone
    Observed Generation:   1
    Reason:                PausingDatabase
    Status:                True
    Type:                  PausingDatabase
    Last Transition Time:  2020-06-18T04:29:58Z
    Message:               MySQLOpsRequestDefinition: myopsreq-standalone for Paused MySQL: demo/my-standalone
    Observed Generation:   1
    Reason:                PausedDatabase
    Status:                True
    Type:                  PausedDatabase
    Last Transition Time:  2020-06-18T04:29:58Z
    Message:               MySQLOpsRequestDefinition: myopsreq-standalone for Upgrading MySQL version: demo/my-standalone
    Observed Generation:   1
    Reason:                OpsRequestUpgradingVersion
    Status:                True
    Type:                  UpgradingVersion
    Last Transition Time:  2020-06-18T04:33:38Z
    Message:               MySQLOpsRequestDefinition: myopsreq-standalone for image successfully upgraded for MySQ: demo/my-standalone
    Observed Generation:   1
    Reason:                OpsRequestUpgradedVersion
    Status:                True
    Type:                  UpgradedVersion
    Last Transition Time:  2020-06-18T04:33:38Z
    Message:               MySQLOpsRequestDefinition: myopsreq-standalone for Resuming MySQL: demo/my-standalone
    Observed Generation:   2
    Reason:                ResumingDatabase
    Status:                True
    Type:                  ResumingDatabase
    Last Transition Time:  2020-06-18T04:33:38Z
    Message:               MySQLOpsRequestDefinition: myopsreq-standalone for Reasumed MySQL: demo/my-standalone
    Observed Generation:   2
    Reason:                ResumedDatabase
    Status:                True
    Type:                  ResumedDatabase
    Last Transition Time:  2020-06-18T04:33:38Z
    Message:               The controller has scaled/upgraded the MySQL successfully
    Observed Generation:   2
    Reason:                OpsRequestSuccessful
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type    Reason           Age    From                        Message
  ----    ------           ----   ----                        -------
  Normal  Pausing          8m16s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, Pausing MySQL: demo/my-standalone
  Normal  SuccessfulPause  8m16s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, successfully paused: demo/my-standalone
  Normal  Starting         8m16s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, Upgrading MySQL images: demo/my-standalone
  Normal  Successful       4m56s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, image successfully upgraded for Pod: demo/my-standalone-0
  Normal  Successful       4m36s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, image successfully upgraded for MySQL: demo/my-standalone
  Normal  Successful       4m36s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, image successfully upgraded for Pod: demo/my-standalone-0
  Normal  Resuming         4m36s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, Resuming MySQL: demo/my-standalone
  Normal  Successful       4m36s  KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myopsreq-standalone, Resumed for MySQL: demo/my-standalone
```

Now, we are going to verify whether the `MySQL`, `StatefulSet` and it's `Pod` images have updated. Let's check,

```console
$ kubectl get my -n demo my-standalone -o=jsonpath='{.spec.version}{"\n"}'
8.0.20

$ kubectl get sts -n demo my-standalone -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/my:8.0.20

$ kubectl get pod -n demo my-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/my:8.0.20
```

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo my-standalone
kubectl delete myops -n demo myopsreq-standalone
```