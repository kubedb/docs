---
title: Vertical Scaling MySQL group replication
menu:
  docs_{{ .version }}:
    identifier: my-vertical-scaling-group
    name: Group Replication
    parent:  my-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="Vertical scaling is an Enterprise feature of KubeDB. You must have KubeDB Enterprise operator installed to test this feature." >}}

# Vertical Scale MySQL Group Replication

This guide will show you how to use `KubeDB` enterprise operator to update the resources of the members of a `MySQL` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/concepts/databases/mysql.md)
  - [MySQLOpsRequest](/docs/concepts/day-2-operations/mysqlopsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/mysql/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/day-2-operations](/docs/examples/day-2-operations) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Vertical Scaling on MySQL Group Replication

Here, we are going to deploy a  `MySQL` group replication using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

#### Prepare Group Replication

At first, we are going to deploy a group replication server using supported `MySQL` version. Then, we are going to update the resources of the members through vertical scaling.

**Find supported MySQL Version:**

When you have installed `KubeDB`, it has created `MySQLVersion` cr for all supported `MySQL` versions.  Let's check the supported MySQL versions,

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

The version above that does not show `DEPRECATED` `true` are supported by `KubeDB` for `MySQL`. You can use any non-deprecated version. Here, we are going to create a MySQL Group Replication using non-deprecated `MySQL` version `8.0.20`.

**Deploy MySQL Group Replication :**

In this section, we are going to deploy a MySQL group replication with 3 members. Then, in the next section we will update the resources of the members using vertical scaling. Below is the YAML of the `MySQL` cr that we are going to create,

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

Let's create the `MySQL` cr we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/verticalscaling/group_replication.yaml
mysql.kubedb.com/my-group created
```

**Wait for the cluster to be ready :**

`KubeDB` operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, `KubeDB` operator will create a new StatefulSet, Services and Secrets etc.
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

Let's check one of the StatefulSet's pod containers resources,

```console
$ kubectl get pod -n demo my-group-0 -o json | jq '.spec.containers[].resources'
{}
```

You can see that the Pod has empty resources that means the scheduler will choose a random node to place the container of the Pod on by default.
We are ready to apply the vertical scale on this group replication.

#### Vertical Scaling

Here, we are going to update the resources of the database cluster to meet up the desired resources after scaling.

**Create MySQLOpsRequest:**

In order to update the resources of your database cluster, you have to create a `MySQLOpsRequest` cr with your desired resources after scaling. Below is the YAML of the `MySQLOpsRequest` crd that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-vertical
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: my-group
  verticalScaling:
    mysql:
      requests:
        memory: "200Mi"
        cpu: "0.1"
      limits:
        memory: "300Mi"
        cpu: "0.2"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` `MySQL` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mysql` specifies the expected mysql container resources after scaling.

Let's create the `MySQLOpsRequest` cr we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/verticalscaling/vertical_scale.yaml
mysqlopsrequest.ops.kubedb.com/myops-vertical created
```

**Verify MySQL Group Replication resources updated successfully :**

If everything goes well, `KubeDB` enterprise operator will update the resources of the StatefulSet's `Pod` containers. After successful scaling process is done, the `KubeDB` enterprise operator update the resources of the `MySQL` cluster.

First, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` crd,

```console
Every 3.0s: kubectl get myops -n demo myops-vertical             suaas-appscode: Wed Jul  1 18:04:52 2020

NAME             TYPE              STATUS       AGE
myops-vertical   VerticalScaling   Successful   4m45s
```

You can see from the above output that the `MySQLOpsRequest` has succeeded. If you describe the `MySQLOpsRequest` you will see that the resources of the members of the `MySQL` group replication are updated.

```console
$ kubectl describe myops -n demo myops-vertical
Name:         myops-vertical
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2020-07-01T10:25:05Z
  Finalizers:
    mysql.ops.kubedb.com
  Generation:        2
  Resource Version:  16560
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mysqlopsrequests/myops-vertical
  UID:               bf8a3fc9-aab9-4fa6-a8f8-5c393e0bc166
Spec:
  Database Ref:
    Name:                my-group
  Stateful Set Ordinal:  0
  Type:                  VerticalScaling
  Vertical Scaling:
    Mysql:
      Limits:
        Cpu:     0.2
        Memory:  300Mi
      Requests:
        Cpu:     0.1
        Memory:  200Mi
Status:
  Conditions:
    Last Transition Time:  2020-07-01T10:25:06Z
    Message:               The controller has started to Progress the OpsRequest
    Observed Generation:   1
    Reason:                OpsRequestOpsRequestProgressing
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-07-01T10:25:06Z
    Message:               MySQLOpsRequestDefinition: myops-vertical for Pausing MySQL: demo/my-group
    Observed Generation:   1
    Reason:                PausingDatabase
    Status:                True
    Type:                  PausingDatabase
    Last Transition Time:  2020-07-01T10:25:06Z
    Message:               MySQLOpsRequestDefinition: myops-vertical for Paused MySQL: demo/my-group
    Observed Generation:   1
    Reason:                PausedDatabase
    Status:                True
    Type:                  PausedDatabase
    Last Transition Time:  2020-07-01T10:25:06Z
    Message:               MySQLOpsRequestDefinition: myops-vertical for Scaling MySQL: demo/my-group
    Observed Generation:   1
    Reason:                OpsRequestScalingDatabase
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2020-07-01T10:32:35Z
    Message:               MySQLOpsRequestDefinition: myops-vertical for Vertical Scaling MySQL: demo/my-group
    Observed Generation:   1
    Reason:                OpsRequestVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2020-07-01T10:32:36Z
    Message:               MySQLOpsRequestDefinition: myops-vertical for Resuming MySQL: demo/my-group
    Observed Generation:   2
    Reason:                ResumingDatabase
    Status:                True
    Type:                  ResumingDatabase
    Last Transition Time:  2020-07-01T10:32:36Z
    Message:               MySQLOpsRequestDefinition: myops-vertical for Reasumed MySQL: demo/my-group
    Observed Generation:   2
    Reason:                ResumedDatabase
    Status:                True
    Type:                  ResumedDatabase
    Last Transition Time:  2020-07-01T10:32:36Z
    Message:               The controller has scaled/upgraded the MySQL successfully
    Observed Generation:   2
    Reason:                OpsRequestSuccessful
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type    Reason           Age   From                        Message
  ----    ------           ----  ----                        -------
  Normal  SuccessfulPause  29m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, successfully paused: demo/my-group
  Normal  Starting         29m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical for Scaling MySQL: demo/my-group
  Normal  Pausing          29m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, Pausing MySQL: demo/my-group
  Normal  Successful       26m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-1
  Normal  Successful       26m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-1
  Normal  Successful       26m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-1
  Normal  Successful       24m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-2
  Normal  Successful       24m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-1
  Normal  Successful       24m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-2
  Normal  Starting         22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, Vertical Scaling MySQL: demo/my-group
  Normal  Successful       22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-2
  Normal  Successful       22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-1
  Normal  Successful       22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-0
  Normal  Successful       22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-0
  Normal  Successful       22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-2
  Normal  Successful       22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, resources successfully updated for Pod: demo/my-group-1
  Normal  Resuming         22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, Resuming MySQL: demo/my-group
  Normal  Successful       22m   KubeDB Enterprise Operator  MySQLOpsRequestDefinition: myops-vertical, Resumed for MySQL: demo/my-group
```

Now, we are going to verify whether the resources of the members of the cluster have updated to meet up the desire state, Let's check,

```console
$ kubectl get pod -n demo my-group-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "200Mi"
  }
}
```

The above output verify that we have successfully updated the resources of the `MySQL` group replication.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo my-group
kubectl delete myops -n demo myops-vertical
```