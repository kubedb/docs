---
title: Vertical Scaling MySQL group replication
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-scaling-vertical-group
    name: Group Replication
    parent: guides-mysql-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MySQL Group Replication

This guide will show you how to use `KubeDB` enterprise operator to update the resources of the members of a `MySQL` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/database/index.md)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest/index.md)
  - [Vertical Scaling Overview](/docs/guides/mysql/scaling/vertical-scaling/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mysql/scaling/vertical-scaling/group-replication/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/scaling/vertical-scaling/group-replication/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

### Apply Vertical Scaling on MySQL Group Replication

Here, we are going to deploy a  `MySQL` group replication using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

#### Prepare Group Replication

At first, we are going to deploy a group replication server using a supported `MySQL` version. Then, we are going to update the resources of the members through vertical scaling.

**Find supported MySQL Version:**

When you have installed `KubeDB`, it has created `MySQLVersion` CR for all supported `MySQL` versions.  Let's check the supported MySQL versions,

```bash
$ kubectl get mysqlversion
NAME        VERSION   DB_IMAGE                  DEPRECATED   AGE
5.7.25-v2   5.7.25    kubedb/mysql:5.7.25-v2                 3h55m
5.7.36   5.7.29    kubedb/mysql:5.7.36                 3h55m
5.7.36   5.7.31    kubedb/mysql:5.7.36                 3h55m
5.7.36   5.7.33    kubedb/mysql:5.7.36                 3h55m
8.0.14-v2   8.0.14    kubedb/mysql:8.0.14-v2                 3h55m
8.0.20-v1   8.0.20    kubedb/mysql:8.0.20-v1                 3h55m
8.0.27   8.0.21    kubedb/mysql:8.0.27                 3h55m
8.0.27      8.0.27    kubedb/mysql:8.0.27                    3h55m
8.0.3-v2    8.0.3     kubedb/mysql:8.0.3-v2                  3h55m
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Here, we are going to create a MySQL Group Replication using non-deprecated `MySQL` version `8.0.27`.

**Deploy MySQL Group Replication:**

In this section, we are going to deploy a MySQL group replication with 3 members. Then, in the next section we will update the resources of the members using vertical scaling. Below is the YAML of the `MySQL` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "8.0.27"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/scaling/vertical-scaling/group-replication/yamls/group-replication.yaml
mysql.kubedb.com/my-group created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc.
Now, watch `MySQL` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get my -n demo my-group
Every 3.0s: kubectl get my -n demo my-group                     suaas-appscode: Tue Jun 30 22:43:57 2020

NAME       VERSION   STATUS    AGE
my-group   8.0.27    Running   16m

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

Let's check one of the StatefulSet's pod containers resources,

```bash
$ kubectl get pod -n demo my-group-0 -o json | jq '.spec.containers[1].resources'
{}
```

You can see that the Pod has empty resources that mean the scheduler will choose a random node to place the container of the Pod on by default. Now, we are ready to apply the vertical scale on this group replication.

#### Vertical Scaling

Here, we are going to update the resources of the database cluster to meet up with the desired resources after scaling.

**Create MySQLOpsRequest:**

In order to update the resources of your database cluster, you have to create a `MySQLOpsRequest` cr with your desired resources after scaling. Below is the YAML of the `MySQLOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-scale-group
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: my-group
  verticalScaling:
    mysql:
      requests:
        memory: "1200Mi"
        cpu: "0.7"
      limits:
        memory: "1200Mi"
        cpu: "0.7"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` `MySQL` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mysql` specifies the expected mysql container resources after scaling.

Let's create the `MySQLOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/scaling/vertical-scaling/group-replication/yamls/my-scale-group.yaml
mysqlopsrequest.ops.kubedb.com/my-scale-group created
```

**Verify MySQL Group Replication resources updated successfully:**

If everything goes well, `KubeDB` enterprise operator will update the resources of the StatefulSet's `Pod` containers. After a successful scaling process is done, the `KubeDB` enterprise operator updates the resources of the `MySQL` cluster.

First, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```bash
$ watch -n 3 kubectl get myops -n demo my-scale-group
Every 3.0s: kubectl get myops -n demo my-sc...  suaas-appscode: Wed Aug 12 16:49:21 2020

NAME             TYPE              STATUS       AGE
my-scale-group   VerticalScaling   Successful   4m53s
```

You can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest`, we shall see that the resources of the members of the `MySQL` group replication are updated.

```bash
$ kubectl describe myops -n demo my-scale-group
Name:         my-scale-group
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2021-03-10T10:54:24Z
  Generation:          2
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-10T10:54:24Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-10T10:56:19Z
  Resource Version:  1083654
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/my-scale-group
  UID:               92ace405-666c-45da-aaa1-fd3d079c187d
Spec:
  Database Ref:
    Name:                my-group
  Stateful Set Ordinal:  0
  Type:                  VerticalScaling
  Vertical Scaling:
    Mysql:
      Limits:
        Cpu:     0.7
        Memory:  1200Mi
      Requests:
        Cpu:     0.7
        Memory:  1200Mi
Status:
  Conditions:
    Last Transition Time:  2021-03-10T10:54:24Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/my-scale-group
    Observed Generation:   2
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2021-03-10T10:54:24Z
    Message:               Vertical scaling started in MySQL: demo/my-group for MySQLOpsRequest: my-scale-group
    Observed Generation:   2
    Reason:                VerticalScalingStarted
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2021-03-10T10:56:19Z
    Message:               Vertical scaling performed successfully in MySQL: demo/my-group for MySQLOpsRequest: my-scale-group
    Observed Generation:   2
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2021-03-10T10:56:19Z
    Message:               Controller has successfully scaled the MySQL demo/my-scale-group
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    7m46s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/my-scale-group
  Normal  Starting    7m46s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/my-group
  Normal  Successful  7m46s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/my-group for MySQLOpsRequest: my-scale-group
  Normal  Starting    7m46s  KubeDB Enterprise Operator  Vertical scaling started in MySQL: demo/my-group for MySQLOpsRequest: my-scale-group
  Normal  Starting    7m41s  KubeDB Enterprise Operator  Restarting Pod: demo/my-group-1
  Normal  Starting    7m1s   KubeDB Enterprise Operator  Restarting Pod: demo/my-group-2
  Normal  Starting    6m31s  KubeDB Enterprise Operator  Restarting Pod (master): demo/my-group-0
  Normal  Successful  5m51s  KubeDB Enterprise Operator  Vertical scaling performed successfully in MySQL: demo/my-group for MySQLOpsRequest: my-scale-group
  Normal  Starting    5m51s  KubeDB Enterprise Operator  Resuming MySQL database: demo/my-group
  Normal  Successful  5m51s  KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/my-group
  Normal  Successful  5m51s  KubeDB Enterprise Operator  Controller has Successfully scaled the MySQL database: demo/my-group
```

Now, we are going to verify whether the resources of the members of the cluster have updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo my-group-0 -o json | jq '.spec.containers[1].resources'
{
  "limits": {
    "cpu": "700m",
    "memory": "1200Mi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "1200Mi"
  }
}
```

The above output verifies that we have successfully updated the resources of the `MySQL` group replication.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo my-group
kubectl delete myops -n demo my-scale-group
```