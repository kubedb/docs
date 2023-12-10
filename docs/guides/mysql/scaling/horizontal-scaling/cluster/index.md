---
title: Horizontal Scaling MySQL Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-scaling-horizontal-cluster
    name: MySQL Cluster
    parent: guides-mysql-scaling-horizontal
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale MySQL Group Replication

This guide will show you how to use `KubeDB` Ops Manager to increase/decrease the number of members of a `MySQL` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/database/index.md)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest/index.md)
  - [Horizontal Scaling Overview](/docs/guides/mysql/scaling/horizontal-scaling/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mysql/scaling/horizontal-scaling/cluster/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/scaling/horizontal-scaling/cluster/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

### Apply Horizontal Scaling on MySQL Group Replication

Here, we are going to deploy a  `MySQL` group replication using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

#### Prepare Group Replication

At first, we are going to deploy a group replication server with 3 members. Then, we are going to add two additional members through horizontal scaling. Finally, we will remove 1 member from the cluster again via horizontal scaling.

**Find supported MySQL Version:**

When you have installed `KubeDB`, it has created `MySQLVersion` CR for all supported `MySQL` versions.  Let's check the supported MySQL versions,

```bash
$ kubectl get mysqlversion
NAME            VERSION   DISTRIBUTION   DB_IMAGE                    DEPRECATED   AGE
5.7.35-v1       5.7.35    Official       mysql:5.7.35                             4d2h
5.7.41          5.7.41    Official       mysql:5.7.41                             4d2h
8.0.17          8.0.17    Official       mysql:8.0.17                             4d2h
8.0.32          8.0.32    Official       mysql:8.0.32                             4d2h
8.0.32-innodb   8.0.32    MySQL          mysql/mysql-server:8.0.32                4d2h
8.0.32          8.0.32    Official       mysql:8.0.32                             4d2h
8.0.3-v4        8.0.3     Official       mysql:8.0.3                              4d2h
8.0.31          8.0.31    Official       mysql:8.0.31                             4d2h
8.0.31-innodb   8.0.31    MySQL          mysql/mysql-server:8.0.31                4d2h
```


The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Here, we are going to create a MySQL Group Replication using `MySQL`  `8.0.31`.

**Deploy MySQL Cluster:**


<ul class="nav nav-tabs" id="definationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="gr-tab" data-toggle="tab" href="#groupReplication" role="tab" aria-controls="groupReplication" aria-selected="true">Group Replication</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="ic-tab" data-toggle="tab" href="#innodbCluster" role="tab" aria-controls="innodbCluster" aria-selected="false">Innodb Cluster</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="sc-tab" data-toggle="tab" href="#semisync" role="tab" aria-controls="semisync" aria-selected="false">Semi sync </a>
  </li>
</ul>


<div class="tab-content" id="definationTabContent">
  <div class="tab-pane fade show active" id="groupReplication" role="tabpanel" aria-labelledby="gr-tab">

**Group Replication:**

In this section, we are going to deploy a MySQL group replication with 3 members. Then, in the next section we will scale-up the cluster using horizontal scaling. Below is the YAML of the `MySQL` cr that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "8.0.31"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/scaling/horizontal-scaling/cluster/yamls/group-replication.yaml
mysql.kubedb.com/my-group created
```

  </div>

  <div class="tab-pane fade" id="innodbCluster" role="tabpanel" aria-labelledby="sc-tab">

**MySQL Innodb Cluster:**

In this section, we are going to deploy a MySQL innodb with 3 members. Then, in the next section we will scale-up the cluster using horizontal scaling. Below is the YAML of the `MySQL` cr that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "8.0.31-innodb"
  replicas: 3
  topology:
    mode: InnoDBCluster
    innoDBCluster:
      router:
        replicas: 1
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/scaling/horizontal-scaling/cluster/yamls/innodb.yaml
mysql.kubedb.com/my-group created
```

  </div>

  <div class="tab-pane fade " id="semisync" role="tabpanel" aria-labelledby="sc-tab">

**MySQL Semi Sync Cluster:**

In this section, we are going to deploy a MySQL semi-sync with 3 members. Then, in the next section we will scale-up the cluster using horizontal scaling. Below is the YAML of the `MySQL` cr that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "8.0.31"
  replicas: 3
  topology:
    mode: SemiSync
    semiSync:
      sourceWaitForReplicaCount: 1
      sourceTimeout: 23h
      errantTransactionRecoveryPolicy: PseudoTransaction
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/scaling/horizontal-scaling/cluster/yamls/semi-sync.yaml
mysql.kubedb.com/my-group created
```


  </div>
</div>


**Wait for the cluster to be ready:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc. A secret called `my-group-auth` (format: <em>{mysql-object-name}-auth</em>) will be created storing the password for mysql superuser.
Now, watch `MySQL` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,


```bash
$ watch -n 3 kubectl get my -n demo my-group
Every 3.0s: kubectl get my -n demo my-group                     suaas-appscode: Tue Jun 30 22:43:57 2020

NAME       VERSION   STATUS    AGE
my-group   8.0.31    Running   16m

$ watch -n 3 kubectl get sts -n demo my-group
Every 3.0s: kubectl get sts -n demo my-group                     Every 3.0s: kubectl get sts -n demo my-group                    suaas-appscode: Tue Jun 30 22:44:35 2020

NAME       READY   AGE
my-group   3/3     16m

$ watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group
Every 3.0s: kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com  suaas-appscode: Tue Jun 30 22:45:33 2020

NAME         READY   STATUS    RESTARTS   AGE
my-group-0   2/2     Running   0          17m
my-group-1   2/2     Running   0          14m
my-group-2   2/2     Running   0          11m
```

Let's verify that the StatefulSet's pods have joined into a group replication cluster,

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
sWfUMoqRpOJyomgb

$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=sWfUMoqRpOJyomgb --host=my-group-0.my-group-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| group_replication_applier | 596be47b-baef-11ea-859a-02c946ef4fe7 | my-group-1.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | 815974c2-baef-11ea-bd7e-a695cbdbd6cc | my-group-2.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | ec61cef2-baee-11ea-adb0-9a02630bae5d | my-group-0.my-group-pods.demo |        3306 | ONLINE       | PRIMARY     | 8.0.23         |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
```

So, we can see that our group replication cluster has 3 members. Now, we are ready to apply the horizontal scale to this group replication.

#### Scale Up

Here, we are going to add 2 members in our group replication using horizontal scaling.

**Create MySQLOpsRequest:**

To scale up your cluster, you have to create a `MySQLOpsRequest` cr with your desired number of members after scaling. Below is the YAML of the `MySQLOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-scale-up
  namespace: demo
spec:
  type: HorizontalScaling  
  databaseRef:
    name: my-group
  horizontalScaling:
    member: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` `MySQL` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the expected number of members after the scaling.

Let's create the `MySQLOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/scaling/horizontal-scaling/cluster/yamls/scale_up.yaml
mysqlopsrequest.ops.kubedb.com/my-scale-up created
```

**Verify Scale-Up Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale up the StatefulSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `MySQL` object.

First, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```bash
Every 3.0s: kubectl get myops -n demo my-scale-up              suaas-appscode: Sat Jul 25 15:49:42 2020

NAME            TYPE                STATUS       AGE
my-scale-up     HorizontalScaling   Successful   2m55s
```

You can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest`, we shall see that the `MySQL` group replication is scaled up.

```bash
$ kubectl describe myops -n demo my-scale-up
$ Name:         my-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2021-03-10T11:18:39Z
  Generation:          2
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-10T11:20:45Z
  Resource Version:  1088850
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/my-scale-up
  UID:               f60f5bbc-3086-4b86-bf74-baf0b39ff358
Spec:
  Database Ref:
    Name:  my-group
  Horizontal Scaling:
    Member:              5
  Stateful Set Ordinal:  0
  Type:                  HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-03-10T11:18:39Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-scale-up
    Observed Generation:   2
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2021-03-10T11:18:40Z
    Message:               Horizontal scaling started in MySQL: demo/my-group for MySQLOpsRequest: my-scale-up
    Observed Generation:   2
    Reason:                HorizontalScalingStarted
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2021-03-10T11:20:45Z
    Message:               Horizontal scaling Up performed successfully in MySQL: demo/my-group for MySQLOpsRequest: my-scale-up
    Observed Generation:   2
    Reason:                SuccessfullyPerformedHorizontalScaling
    Status:                True
    Type:                  ScalingUp
    Last Transition Time:  2021-03-10T11:20:45Z
    Message:               Controller has successfully scaled the MySQL demo/my-scale-up
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    28m   KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-scale-up
  Normal  Starting    28m   KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-group
  Normal  Successful  28m   KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-group for MySQLOpsRequest: my-scale-up
  Normal  Starting    28m   KubeDB Enterprise Operator  Horizontal scaling started in MySQL: demo/my-group for MySQLOpsRequest: my-scale-up
  Normal  Successful  26m   KubeDB Enterprise Operator  Horizontal scaling Up performed successfully in MySQL: demo/my-group for MySQLOpsRequest: my-scale-up
  Normal  Starting    26m   KubeDB Enterprise Operator  Resuming MySQL database: demo/my-group
  Normal  Successful  26m   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-group
  Normal  Successful  26m   KubeDB Enterprise Operator  Controller has Successfully scaled the MySQL database: demo/my-group
```

Now, we are going to verify whether the number of members has increased to meet up the desired state, Let's check,

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
Y28qkWFQ8QHVzq2h

$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=Y28qkWFQ8QHVzq2h --host=my-group-0.my-group-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| group_replication_applier | 4b76f5c8-baff-11ea-9848-425294afbbbf | my-group-3.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | 73c1f150-baff-11ea-9394-4a8c424ea5c2 | my-group-4.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | 9f6c694c-bafd-11ea-8ad4-822669614bde | my-group-0.my-group-pods.demo |        3306 | ONLINE       | PRIMARY     | 8.0.23         |
| group_replication_applier | c9d82f09-bafd-11ea-ab3a-764d326534a6 | my-group-1.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | eff81073-bafd-11ea-9f3d-ca1e99c33106 | my-group-2.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
```

You can see above that our `MySQL` group replication now has a total of 5 members. It verifies that we have successfully scaled up.

#### Scale Down

Here, we are going to remove 1 member from our group replication using horizontal scaling.

**Create MysQLOpsRequest:**

To scale down your cluster, you have to create a `MySQLOpsRequest` cr with your desired number of members after scaling. Below is the YAML of the `MySQLOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-scale-down
  namespace: demo
spec:
  type: HorizontalScaling  
  databaseRef:
    name: my-group
  horizontalScaling:
    member: 4
```

Let's create the `MySQLOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/scaling/horizontal-scaling/cluster/yamls/scale_down.yaml
mysqlopsrequest.ops.kubedb.com/my-scale-down created
```

**Verify Scale-down Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale down the StatefulSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `MySQL` object.

Now, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```bash
Every 3.0s: kubectl get myops -n demo my-scale-down              suaas-appscode: Sat Jul 25 15:49:42 2020

NAME            TYPE                STATUS       AGE
my-scale-down   HorizontalScaling   Successful   2m55s
```

You can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest`, we shall see that the `MySQL` group replication is scaled down.

```bash
$ kubectl describe myops -n demo my-scale-down
Name:         my-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2021-03-10T11:48:42Z
  Generation:          2
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-10T11:49:18Z
  Resource Version:  1094487
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/my-scale-down
  UID:               7b9471ed-87b1-4c62-939b-dbb8a7641554
Spec:
  Database Ref:
    Name:  my-group
  Horizontal Scaling:
    Member:              4
  Stateful Set Ordinal:  0
  Type:                  HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-03-10T11:48:42Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-scale-down
    Observed Generation:   2
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2021-03-10T11:48:43Z
    Message:               Horizontal scaling started in MySQL: demo/my-group for MySQLOpsRequest: my-scale-down
    Observed Generation:   2
    Reason:                HorizontalScalingStarted
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2021-03-10T11:49:18Z
    Message:               Horizontal scaling down performed successfully in MySQL: demo/my-group for MySQLOpsRequest: my-scale-down
    Observed Generation:   2
    Reason:                SuccessfullyPerformedHorizontalScaling
    Status:                True
    Type:                  ScalingDown
    Last Transition Time:  2021-03-10T11:49:18Z
    Message:               Controller has successfully scaled the MySQL demo/my-scale-down
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     4
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    97s   KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-scale-down
  Normal  Starting    97s   KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-group
  Normal  Successful  97s   KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-group for MySQLOpsRequest: my-scale-down
  Normal  Starting    96s   KubeDB Enterprise Operator  Horizontal scaling started in MySQL: demo/my-group for MySQLOpsRequest: my-scale-down
  Normal  Successful  61s   KubeDB Enterprise Operator  Horizontal scaling down performed successfully in MySQL: demo/my-group for MySQLOpsRequest: my-scale-down
  Normal  Starting    61s   KubeDB Enterprise Operator  Resuming MySQL database: demo/my-group
  Normal  Successful  61s   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-group
  Normal  Successful  61s   KubeDB Enterprise Operator  Controller has Successfully scaled the MySQL database: demo/my-group
```

Now, we are going to verify whether the number of members has decreased to meet up the desired state, Let's check,

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
Y28qkWFQ8QHVzq2h

$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=5pwciRRUWHhSJ6qQ --host=my-group-0.my-group-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| group_replication_applier | 533602d0-ce5b-11ea-b866-5ad2598e5303 | my-group-1.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | 7d429240-ce5b-11ea-9fe2-0aaa5a845ec8 | my-group-2.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | c498302f-ce5b-11ea-96a3-72980d437abc | my-group-3.my-group-pods.demo |        3306 | ONLINE       | SECONDARY   | 8.0.23         |
| group_replication_applier | dfb1633a-ce5a-11ea-a9c8-6e4ef86119d0 | my-group-0.my-group-pods.demo |        3306 | ONLINE       | PRIMARY     | 8.0.23         |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
```

You can see above that our `MySQL` group replication now has a total of 4 members. It verifies that we have successfully scaled down.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo my-group
kubectl delete myops -n demo my-scale-up
kubectl delete myops -n demo my-scale-down
```