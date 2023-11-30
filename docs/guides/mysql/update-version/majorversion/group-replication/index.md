---
title: Updating MySQL group replication major version
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-updating-major-group
    name: Group Replication
    parent: guides-mysql-updating-major
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# update major version of MySQL Group Replication

This guide will show you how to use `KubeDB` enterprise operator to update the major version of `MySQL` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/database/index.md)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest/index.md)
  - [Updating Overview](/docs/guides/mysql/update-version/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mysql/update-version/majorversion/group-replication/yamls](/docs/guides/mysql/update-version/majorversion/group-replication/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Version updating on Group Replication

Here, we are going to deploy a `MySQL` group replication using a supported version by `KubeDB` operator. Then we are going to apply updating on it.

#### Prepare Group Replication

At first, we are going to deploy a group replication using supported that `MySQL` version whether it is possible to update from this version to another. In the next two sections, we are going to find out the supported version and version update constraints.

**Find supported MySQL Version:**

When you have installed `KubeDB`, it has created `MySQLVersion` CR for all supported `MySQL` versions. Let’s check the supported `MySQL` versions,

```bash
$ kubectl get mysqlversion 
NAME            VERSION   DISTRIBUTION   DB_IMAGE                    DEPRECATED   AGE
5.7.35-v1       5.7.35    Official       mysql:5.7.35                             13d
5.7.41          5.7.41    Official       mysql:5.7.41                             13d
8.0.17          8.0.17    Official       mysql:8.0.17                             13d
8.0.32          8.0.32    Official       mysql:8.0.32                             13d
8.0.32-innodb   8.0.32    MySQL          mysql/mysql-server:8.0.32                13d
8.0.32          8.0.32    Official       mysql:8.0.32                             13d
8.0.3-v4        8.0.3     Official       mysql:8.0.3                              13d

```

The version above that does not show `DEPRECATED` true is supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `MySQLVersion` for `MySQL` group replication that will be possible to update from this version to another version. In the next section, we are going to verify version update constraints.

**Check update Constraints:**

Database version update constraints is a constraint that shows whether it is possible or not possible to update from one version to another. Let's check the version update constraints of `MySQL` `5.7.41`,

```bash
$ kubectl get mysqlversion 5.7.41 -o yaml | kubectl neat
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
  name: 5.7.41
  resourceVersion: "1092465"
  uid: 4cc87fc8-efd7-4e69-bb12-4454a2b1bf06
spec:
  coordinator:
    image: kubedb/mysql-coordinator:v0.5.0
  db:
    image: mysql:5.7.41
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
  updateConstraints:
    denylist:
      groupReplication:
      - < 5.7.41
      standalone:
      - < 5.7.41
  version: 5.7.41

```

The above `spec.updateConstraints.denylist` of `5.7.41` is showing that updating below version of `5.7.41` is not possible for both group replication and standalone. That means, it is possible to update any version above `5.7.41`. Here, we are going to create a `MySQL` Group Replication using MySQL  `5.7.41`. Then we are going to update this version to `8.0.32`.

**Deploy MySQL Group Replication:**

In this section, we are going to deploy a MySQL group replication with 3 members. Then, in the next section we will update the version of the  members using updating. Below is the YAML of the `MySQL` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "5.7.41"
  replicas: 3
  topology:
    mode: GroupReplication
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/update-version/majorversion/group-replication/yamls/group_replication.yaml
mysql.kubedb.com/my-group created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc. A secret called `my-group-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
Now, watch `MySQL` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get my -n demo my-group

NAME       VERSION      STATUS    AGE
my-group   5.7.41    Running   5m52s

$ watch -n 3 kubectl get sts -n demo my-group

NAME       READY   AGE
my-group   3/3     7m12s

$ watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group

NAME         READY   STATUS    RESTARTS   AGE
my-group-0   2/2     Running   0          11m
my-group-1   2/2     Running   0          9m53s
my-group-2   2/2     Running   0          6m48s
```

Let's verify the `MySQL`, the `StatefulSet` and its `Pod` image version,

```bash
$ kubectl get my -n demo my-group -o=jsonpath='{.spec.version}{"\n"}'
5.7.41

$ kubectl get sts -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group -o json | jq '.items[].spec.template.spec.containers[1].image'
"kubedb/mysql:5.7.41"

$ kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group -o json | jq '.items[].spec.containers[1].image'
"kubedb/mysql:5.7.41"
"kubedb/mysql:5.7.41"
"kubedb/mysql:5.7.41"
```

Let's also verify that the StatefulSet’s pods have joined into a group replication,

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
7gUARa&Jkg.ypJE8

$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='7gUARa&Jkg.ypJE8' --host=my-group-0.my-group-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                       | MEMBER_PORT | MEMBER_STATE |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+
| group_replication_applier | b0e71e0c-f849-11ec-a315-46392c50e39c | my-group-1.my-group-pods.demo.svc |        3306 | ONLINE       |
| group_replication_applier | b34b16d7-f849-11ec-9362-a2f432876ee4 | my-group-2.my-group-pods.demo.svc |        3306 | ONLINE       |
| group_replication_applier | b5542a4a-f849-11ec-9a75-3e8abd17fee6 | my-group-0.my-group-pods.demo.svc |        3306 | ONLINE       |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+

```

We are ready to apply updating on this `MySQL` group replication.

#### UpdateVesion

Here, we are going to update the `MySQL` group replication from `5.7.41` to `8.0.32`.

**Create MySQLOpsRequest:**

To update your database cluster, you have to create a `MySQLOpsRequest` cr with your desired version that supported by `KubeDB`. Below is the YAML of the `MySQLOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-update-major-group
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: my-group
  updateVersion:
    targetVersion: "8.0.32"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` MySQL database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies expected version `8.0.32` after updating.

Let's create the `MySQLOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/update-version/majorversion/group-replication/yamls/update_major_version_group.yaml
mysqlopsrequest.ops.kubedb.com/my-update-major-group created
```

> Note: During the upgradation of the major version of MySQL group replication, a new StatefulSet is created by the KubeDB enterprise operator and the old one is deleted. The name of the newly created StatefulSet is formed as follows: `<mysql-name>-<suffix>`.
Here, `<suffix>` is a positive integer number and starts with 1. It's determined as follows:
For one-time major version updating of group replication, the suffix will be 1.
For the 2nd time major version updating of group replication, the suffix will be 2.
It will be continued...

**Verify MySQL version updated successfully:**

If everything goes well, `KubeDB` enterprise operator will create a new `StatefulSet` named `my-group-1` with the desire updated version and delete the old one.

At first, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```bash
$ watch -n 3 kubectl get myops -n demo my-update-major-group

NAME                    TYPE            STATUS       AGE
my-update-major-group   UpdateVersion   Successful   5m26s
```

You can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest`, we shall see that the `MySQL` group replication is updated with new images and the `StatefulSet` is created with a new image.

```bash
$ kubectl describe myops -n demo my-update-major-group
Name:         my-update-major-group
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
    Name:  my-group
  Type:    UpdateVersion
  UpdateVersion:
    Target Version:  8.0.32
Status:
  Conditions:
    Last Transition Time:  2022-06-30T07:55:16Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-update-major-group
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-06-30T07:55:16Z
    Message:               MySQL version updateFunc stated for MySQLOpsRequest: demo/my-update-major-group
    Observed Generation:   1
    Reason:                DatabaseVersionupdatingStarted
    Status:                True
    Type:                  updating
    Last Transition Time:  2022-06-30T07:59:16Z
    Message:               Image successfully updated in MySQL: demo/my-group for MySQLOpsRequest: my-update-major-group 
    Observed Generation:   1
    Reason:                SuccessfullyUpdatedDatabaseVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2022-06-30T07:59:16Z
    Message:               Controller has successfully updated the MySQL demo/my-update-major-group
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    5m24s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-update-major-group
  Normal  Starting    5m24s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-group
  Normal  Successful  5m24s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-group for MySQLOpsRequest: my-update-major-group
  Normal  Starting    5m24s  KubeDB Enterprise Operator  Updating MySQL images: demo/my-group for MySQLOpsRequest: my-update-major-group
  Normal  Starting    5m19s  KubeDB Enterprise Operator  Restarting Pod: my-group-1/demo
  Normal  Starting    3m59s  KubeDB Enterprise Operator  Restarting Pod: my-group-2/demo
  Normal  Starting    2m39s  KubeDB Enterprise Operator  Restarting Pod: my-group-0/demo
  Normal  Successful  84s    KubeDB Enterprise Operator  Image successfully updated in MySQL: demo/my-group for MySQLOpsRequest: my-update-major-group
  Normal  Starting    84s    KubeDB Enterprise Operator  Resuming MySQL database: demo/my-group
  Normal  Successful  84s    KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-group
  Normal  Successful  84s    KubeDB Enterprise Operator  Controller has Successfully updated the version of MySQL : demo/my-group
```

Now, we are going to verify whether the `MySQL` and `StatefulSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get my -n demo my-group -o=jsonpath='{.spec.version}{"\n"}'
8.0.32

$ kubectl get sts -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group -o json | jq '.items[].spec.template.spec.containers[1].image'
"kubedb/mysql:8.0.32"

$ kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group -o json | jq '.items[].spec.containers[1].image'
"kubedb/mysql:8.0.32"
"kubedb/mysql:8.0.32"
"kubedb/mysql:8.0.32"
```

Let's also check the StatefulSet pods have joined the `MySQL` group replication,

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
7gUARa&Jkg.ypJE8

$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='7gUARa&Jkg.ypJE8' --host=my-group-0.my-group-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                       | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION | MEMBER_COMMUNICATION_STACK |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| group_replication_applier | b0e71e0c-f849-11ec-a315-46392c50e39c | my-group-1.my-group-pods.demo.svc |        3306 | ONLINE       | PRIMARY     | 8.0.32         | XCom                       |
| group_replication_applier | b34b16d7-f849-11ec-9362-a2f432876ee4 | my-group-2.my-group-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.32         | XCom                       |
| group_replication_applier | b5542a4a-f849-11ec-9a75-3e8abd17fee6 | my-group-0.my-group-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.32         | XCom                       |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+

```

You can see above that our `MySQL` group replication now has updated members. It verifies that we have successfully updated our cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo my-group
kubectl delete myops -n demo my-update-major-group
```