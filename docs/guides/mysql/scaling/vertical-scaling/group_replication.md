---
title: Vertical Scaling MySQL group replication
menu:
  docs_{{ .version }}:
    identifier: my-vertical-scale
    name: my-vertical-scale
    parent: my-upgrading-mysql
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# Vertical Scale MySQL Group Replication

This guide will show you how to use `KubeDB` enterprise operator to update the resources of server nodes of a `MySQL` Group Replication.

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

### Apply Vertical Scaling on MySQL Group Replication

Here, we are going to deploy a  `MySQL` group replication using a supported version by `KubeDB` operator. Below section will check the supported `MySQL` versions.

**Find supported MySQL Version:**

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

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MySQL`. Now we will select a version from `MySQLVersion` for `MySQL` group replication. For `MySQL` group replication deployment, we will select version `8.0.20`.

#### Prepare Group Replication

Now, we are going to deploy a `MySQL` group replication using version `8.0.20`.

**Create MySQL Object:**

Below is the YAML of the `MySQL` crd that we are going to create,

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

Let's create the `MySQL` crd we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/group_replication2.yaml
mysql.kubedb.com/my-group created
```

**Check MySQL group Ready to Scale:**

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

You can see the Pod has empty resources that means the scheduler will choose a random node to place the container of the Pod on by default

We are ready to apply horizontal scale on this group replication.

#### Vertical Scaling

Here, we are going to update the resources of server nodes of the `MySQL` group replication.

**Create MySQLOpsRequest:**

Below is the YAML of the `MySQLOpsRequest` crd that we are going to create,

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

- `spec.databaseRef.name` refers to the `my-group` MySQL object for operation.
- `spec.type` specifies that this is an `VerticalScaling` type operation
- `spec.VerticalScaling.mysql` specifies the mysql container resources to apply

Let's create the `MySQLOpsRequest` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2operations/vertical_scale.yaml
mysqlopsrequest.ops.kubedb.com/myops-vertical created
```

**Check MySQL Group Replication resources updated:**

If everything goes well, `KubeDB` enterprise operator will update the resources of the StatefulSet's `Pod` containers. After successful scaling process is done, the `KubeDB` enterprise operator update the resources of the `MySQL` object.

First, we will wait for `MySQLOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` crd,

```console
Every 3.0s: kubectl get myops -n demo myops-vertical             suaas-appscode: Wed Jul  1 18:04:52 2020

NAME             TYPE              STATUS       AGE
myops-vertical   VerticalScaling   Successful   4m45s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If you describe the `MySQLOpsRequest` you will see that the resources of the server nodes of the `MySQL` group replication are updated.

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

Now, we are going to verify whether the resources of the server nodes have updated to meet up the desire state, Let's check,

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

The above output verify that we have successfully scaled up the resources of the server nodes.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo my-group
kubectl delete myops -n demo myops-vertical
```