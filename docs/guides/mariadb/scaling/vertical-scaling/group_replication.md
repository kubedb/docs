---
title: Vertical Scaling MariaDB group replication
menu:
  docs_{{ .version }}:
    identifier: my-vertical-scaling-group
    name: Group Replication
    parent:  my-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MariaDB Group Replication

This guide will show you how to use `KubeDB` enterprise operator to update the resources of the members of a `MariaDB` Group Replication.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb.md)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/mariadb/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mariadb](/docs/examples/mariadb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Vertical Scaling on MariaDB Group Replication

Here, we are going to deploy a  `MariaDB` group replication using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

#### Prepare Group Replication

At first, we are going to deploy a group replication server using a supported `MariaDB` version. Then, we are going to update the resources of the members through vertical scaling.

**Find supported MariaDB Version:**

When you have installed `KubeDB`, it has created `MariaDBVersion` CR for all supported `MariaDB` versions.  Let's check the supported MariaDB versions,

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

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MariaDB`. You can use any non-deprecated version. Here, we are going to create a MariaDB Group Replication using non-deprecated `MariaDB` version `8.0.20`.

**Deploy MariaDB Group Replication:**

In this section, we are going to deploy a MariaDB group replication with 3 members. Then, in the next section we will update the resources of the members using vertical scaling. Below is the YAML of the `MariaDB` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
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

Let's create the `MariaDB` cr we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/verticalscaling/group_replication.yaml
mariadb.kubedb.com/my-group created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `MariaDB` objects using Kubernetes API. When a `MariaDB` object is created, `KubeDB` operator will create a new StatefulSet, Services, and Secrets, etc.
Now, watch `MariaDB` is going to  `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get my -n demo my-group
Every 3.0s: kubectl get my -n demo my-group                     suaas-appscode: Tue Jun 30 22:43:57 2020

NAME       VERSION   STATUS    AGE
my-group   8.0.20    Running   16m

$ watch -n 3 kubectl get sts -n demo my-group
Every 3.0s: kubectl get sts -n demo my-group                     Every 3.0s: kubectl get sts -n demo my-group                    suaas-appscode: Tue Jun 30 22:44:35 2020

NAME       READY   AGE
my-group   3/3     16m

$ watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/name=mariadbs.kubedb.com,app.kubernetes.io/instance=my-group
Every 3.0s: kubectl get pod -n demo -l app.kubernetes.io/name=mariadbs.kubedb.com  suaas-appscode: Tue Jun 30 22:45:33 2020

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

**Create MariaDBOpsRequest:**

In order to update the resources of your database cluster, you have to create a `MariaDBOpsRequest` cr with your desired resources after scaling. Below is the YAML of the `MariaDBOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: my-scale-group
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: my-group
  verticalScaling:
    mariadb:
      requests:
        memory: "200Mi"
        cpu: "0.1"
      limits:
        memory: "300Mi"
        cpu: "0.2"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `my-group` `MariaDB` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mariadb` specifies the expected mariadb container resources after scaling.

Let's create the `MariaDBOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/verticalscaling/vertical_scale_group.yaml
mariadbopsrequest.ops.kubedb.com/my-scale-group created
```

**Verify MariaDB Group Replication resources updated successfully:**

If everything goes well, `KubeDB` enterprise operator will update the resources of the StatefulSet's `Pod` containers. After a successful scaling process is done, the `KubeDB` enterprise operator updates the resources of the `MariaDB` cluster.

First, we will wait for `MariaDBOpsRequest` to be successful.  Run the following command to watch `MySQlOpsRequest` cr,

```bash
$ watch -n 3 kubectl get myops -n demo my-scale-group
Every 3.0s: kubectl get myops -n demo my-sc...  suaas-appscode: Wed Aug 12 16:49:21 2020

NAME             TYPE              STATUS       AGE
my-scale-group   VerticalScaling   Successful   4m53s
```

You can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest`, we shall see that the resources of the members of the `MariaDB` group replication are updated.

```bash
$ kubectl describe myops -n demo my-scale-group
Name:         my-scale-group
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2020-08-12T10:44:28Z
  Finalizers:
    mariadb.ops.kubedb.com
  Generation:  2
  ...
  Resource Version:  68442
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mariadbopsrequests/my-scale-group
  UID:               ae8f1ab5-ab89-4ba9-bf41-3ce11b1d0cf0
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
    Last Transition Time:  2020-08-12T10:44:28Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/my-scale-group
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-08-12T10:44:28Z
    Message:               Controller has successfully Halted the MariaDB database: demo/my-group 
    Observed Generation:   1
    Reason:                SuccessfullyHaltedDatabase
    Status:                True
    Type:                  HaltDatabase
    Last Transition Time:  2020-08-12T10:44:28Z
    Message:               Vertical scaling started in MariaDB: demo/my-group for MariaDBOpsRequest: my-scale-group
    Observed Generation:   1
    Reason:                VerticalScalingStarted
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2020-08-12T10:49:08Z
    Message:               Vertical scaling performed successfully in MariaDB: demo/my-group for MariaDBOpsRequest: my-scale-group
    Observed Generation:   1
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2020-08-12T10:49:08Z
    Message:               Controller has successfully Resumed the MariaDB database: demo/my-group
    Observed Generation:   2
    Reason:                SuccessfullyResumedDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-12T10:49:08Z
    Message:               Controller has successfully scaled/upgraded the MariaDB demo/my-scale-group
    Observed Generation:   2
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    5m50s  KubeDB Enterprise Operator  Start processing for MariaDBOpsRequest: demo/my-scale-group
  Normal  Starting    5m50s  KubeDB Enterprise Operator  Pausing MariaDB databse: demo/my-group
  Normal  Successful  5m50s  KubeDB Enterprise Operator  Successfully halted MariaDB database: demo/my-group for MariaDBOpsRequest: my-scale-group
  Normal  Starting    5m50s  KubeDB Enterprise Operator  Vertical scaling started in MariaDB: demo/my-group for MariaDBOpsRequest: my-scale-group
  Normal  Successful  4m10s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  3m50s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  3m30s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  3m10s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  2m50s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-1
  Normal  Successful  2m50s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  2m30s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  2m30s  KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-group-1
  Normal  Successful  2m10s  KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  2m10s  KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-group-1
  Normal  Successful  110s   KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  110s   KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-group-1
  Normal  Successful  90s    KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-0
  Normal  Successful  90s    KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  90s    KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-group-1
  Normal  Successful  70s    KubeDB Enterprise Operator  Vertical scaling performed successfully in MariaDB: demo/my-group for MariaDBOpsRequest: my-scale-group
  Normal  Successful  70s    KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-0
  Normal  Successful  70s    KubeDB Enterprise Operator  Image successfully upgraded for Pod: demo/my-group-2
  Normal  Successful  70s    KubeDB Enterprise Operator  Image successfully upgraded for standalone/master: demo/my-group-1
  Normal  Starting    70s    KubeDB Enterprise Operator  Resuming MariaDB database: demo/my-group
  Normal  Successful  70s    KubeDB Enterprise Operator  Successfully resumed MariaDB database: demo/my-group
  Normal  Successful  70s    KubeDB Enterprise Operator  Controller has Successfully scaled the MariaDB database: demo/my-group
```

Now, we are going to verify whether the resources of the members of the cluster have updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo my-group-0 -o json | jq '.spec.containers[1].resources'
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

The above output verifies that we have successfully updated the resources of the `MariaDB` group replication.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo my-group
kubectl delete myops -n demo my-scale-group
```