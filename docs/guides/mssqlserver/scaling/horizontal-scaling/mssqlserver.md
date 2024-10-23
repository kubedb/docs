---
title: Horizontal Scaling MSSQLServer Cluster
menu:
  docs_{{ .version }}:
    identifier: ms-scaling-horizontal-guide
    name: Scale Horizontally
    parent: ms-scaling-horizontal
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale MSSQLServer Cluster

This guide will show you how to use `KubeDB` Ops Manager to increase/decrease the number of members of a `MSSQLServer` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/mssqlserver/scaling/horizontal-scaling/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mssqlserver/scaling/horizontal-scaling/scale-horizontally/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mssqlserver/scaling/horizontal-scaling/scale-horizontally/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

### Apply Horizontal Scaling on MSSQLServer Cluster

Here, we are going to deploy a `MSSQLServer` Cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

#### Prepare Cluster

At first, we are going to deploy a Cluster server with 3 members. Then, we are going to add two additional members through horizontal scaling. Finally, we will remove 1 member from the cluster again via horizontal scaling.

**Find supported MSSQLServer Version:**

When you have installed `KubeDB`, it has created `MSSQLServerVersion` CR for all supported `MSSQLServer` versions. Let's check the supported MSSQLServer versions,

```bash
$ kubectl get mssqlserverversion
NAME                       VERSION   DISTRIBUTION   DB_IMAGE                               DEPRECATED   AGE
10.16                      10.16     Official       mssqlserver:10.16-alpine                               63s
10.16-debian               10.16     Official       mssqlserver:10.16                                      63s
10.19                      10.19     Official       mssqlserver:10.19-bullseye                             63s
10.19-bullseye             10.19     Official       mssqlserver:10.19-bullseye                             63s
11.11                      11.11     Official       mssqlserver:11.11-alpine                               63s
11.11-debian               11.11     Official       mssqlserver:11.11                                      63s
11.14                      11.14     Official       mssqlserver:11.14-alpine                               63s
11.14-bullseye             11.14     Official       mssqlserver:11.14-bullseye                             63s
11.14-bullseye-postgis     11.14     PostGIS        postgis/postgis:11-3.1                              63s
12.6                       12.6      Official       mssqlserver:12.6-alpine                                63s
12.6-debian                12.6      Official       mssqlserver:12.6                                       63s
12.9                       12.9      Official       mssqlserver:12.9-alpine                                63s
12.9-bullseye              12.9      Official       mssqlserver:12.9-bullseye                              63s
12.9-bullseye-postgis      12.9      PostGIS        postgis/postgis:12-3.1                              63s
13.2                       13.2      Official       mssqlserver:13.2-alpine                                63s
13.2-debian                13.2      Official       mssqlserver:13.2                                       63s
13.5                       13.5      Official       mssqlserver:13.5-alpine                                63s
13.5-bullseye              13.5      Official       mssqlserver:13.5-bullseye                              63s
13.5-bullseye-postgis      13.5      PostGIS        postgis/postgis:13-3.1                              63s
14.1                       14.1      Official       mssqlserver:14.1-alpine                                63s
14.1-bullseye              14.1      Official       mssqlserver:14.1-bullseye                              63s
14.1-bullseye-postgis      14.1      PostGIS        postgis/postgis:14-3.1                              63s
9.6.21                     9.6.21    Official       mssqlserver:9.6.21-alpine                              63s
9.6.21-debian              9.6.21    Official       mssqlserver:9.6.21                                     63s
9.6.24                     9.6.24    Official       mssqlserver:9.6.24-alpine                              63s
9.6.24-bullseye            9.6.24    Official       mssqlserver:9.6.24-bullseye                            63s
timescaledb-2.1.0-ms11     11.11     TimescaleDB    timescale/timescaledb:2.1.0-ms11-oss                63s
timescaledb-2.1.0-ms12     12.6      TimescaleDB    timescale/timescaledb:2.1.0-ms12-oss                63s
timescaledb-2.1.0-ms13     13.2      TimescaleDB    timescale/timescaledb:2.1.0-ms13-oss                63s
timescaledb-2.5.0-ms14.1   14.1      TimescaleDB    timescale/timescaledb:2.5.0-ms14-oss                63s
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MSSQLServer`. You can use any non-deprecated version. Here, we are going to create a MSSQLServer Cluster using `MSSQLServer` `13.2`.

**Deploy MSSQLServer Cluster:**

In this section, we are going to deploy a MSSQLServer Cluster with 3 members. Then, in the next section we will scale-up the cluster using horizontal scaling. Below is the YAML of the `MSSQLServer` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: ms
  namespace: demo
spec:
  version: "13.13"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MSSQLServer` cr we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/scaling/horizontal-scaling/scale-horizontally/mssqlserver.yaml
mssqlserver.kubedb.com/ms created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `MSSQLServer` objects using Kubernetes API. When a `MSSQLServer` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc. A secret called `ms-auth` (format: <em>{mssqlserver-object-name}-auth</em>) will be created storing the password for mssqlserver superuser.
Now, watch `MSSQLServer` is going to `Running` state and also watch `PetSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get mssqlserver -n demo ms
Every 3.0s: kubectl get mssqlserver -n demo ms                        emon-r7: Thu Dec  2 15:31:16 2021

NAME   VERSION   STATUS   AGE
ms     13.2      Ready    4h40m


$ watch -n 3 kubectl get sts -n demo ms
Every 3.0s: kubectl get sts -n demo ms                             emon-r7: Thu Dec  2 15:31:38 2021

NAME   READY   AGE
ms     3/3     4h41m



$ watch -n 3 kubectl get pods -n demo
Every 3.0s: kubectl get pod -n demo                                emon-r7: Thu Dec  2 15:33:24 2021

NAME   READY   STATUS    RESTARTS   AGE
ms-0   2/2     Running   0          4h25m
ms-1   2/2     Running   0          4h26m
ms-2   2/2     Running   0          4h26m

```

Let's verify that the PetSet's pods have joined into cluster,

```bash
$ kubectl get secrets -n demo ms-auth -o jsonpath='{.data.\username}' | base64 -d
mssqlserver

$ kubectl get secrets -n demo ms-auth -o jsonpath='{.data.\password}' | base64 -d
b3b5838EhjwsiuFU

```

So, we can see that our cluster has 3 members. Now, we are ready to apply the horizontal scale to this MSSQLServer cluster.

#### Scale Up

Here, we are going to add 2 replicas in our Cluster using horizontal scaling.

**Create MSSQLServerOpsRequest:**

To scale up your cluster, you have to create a `MSSQLServerOpsRequest` cr with your desired number of replicas after scaling. Below is the YAML of the `MSSQLServerOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: ms-scale-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ms
  horizontalScaling:
    replicas: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `ms` `MSSQLServer` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the expected number of replicas after the scaling.

Let's create the `MSSQLServerOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/scaling/horizontal-scaling/scale-horizontally/yamls/ms-scale-up.yaml
mssqlserveropsrequest.ops.kubedb.com/ms-scale-up created
```

**Verify Scale-Up Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale up the PetSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `MSSQLServer` object.

First, we will wait for `MSSQLServerOpsRequest` to be successful. Run the following command to watch `MSSQLServerOpsRequest` cr,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo ms-scale-up
Every 2.0s: kubectl get mssqlserveropsrequest -n demo ms-scale-up     emon-r7: Thu Dec  2 17:57:36 2021

NAME          TYPE                STATUS       AGE
ms-scale-up   HorizontalScaling   Successful   8m23s

```

You can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest`, we will see that the `MSSQLServer` cluster is scaled up.

```bash
kubectl describe mssqlserveropsrequest -n demo ms-scale-up
Name:         ms-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2021-12-02T11:49:13Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:horizontalScaling:
          .:
          f:replicas:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-12-02T11:49:13Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-12-02T11:49:13Z
  Resource Version:  49610
  UID:               cc62fe84-5c13-4c77-b130-f748c0beff27
Spec:
  Database Ref:
    Name:  ms
  Horizontal Scaling:
    Replicas:  5
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-12-02T11:49:13Z
    Message:               MSSQLServer ops request is horizontally scaling database
    Observed Generation:   1
    Reason:                Progressing
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2021-12-02T11:50:38Z
    Message:               Successfully Horizontally Scaled Up
    Observed Generation:   1
    Reason:                ScalingUp
    Status:                True
    Type:                  ScalingUp
    Last Transition Time:  2021-12-02T11:50:38Z
    Message:               Successfully Horizontally Scaled MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason          Age    From                        Message
  ----    ------          ----   ----                        -------
  Normal  PauseDatabase   10m    KubeDB Enterprise Operator  Pausing MSSQLServer demo/ms
  Normal  PauseDatabase   10m    KubeDB Enterprise Operator  Successfully paused MSSQLServer demo/ms
  Normal  ScalingUp       9m17s  KubeDB Enterprise Operator  Successfully Horizontally Scaled Up
  Normal  ResumeDatabase  9m17s  KubeDB Enterprise Operator  Resuming MSSQLServer demo/ms
  Normal  ResumeDatabase  9m17s  KubeDB Enterprise Operator  Successfully resumed MSSQLServer demo/ms
  Normal  Successful      9m17s  KubeDB Enterprise Operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify whether the number of members has increased to meet up the desired state. So let's check the new pods logs to see if they have joined in the cluster as new replica.

```bash
$ kubectl logs   -n demo                 ms-4  -c mssqlserver -f
waiting for the role to be decided ...
running the initial script ...
Running as Replica
Attempting ms_isready on primary
Attempting query on primary
take base basebackup...
2021-12-02 11:50:11.062 GMT [17] LOG:  skipping missing configuration file "/etc/config/user.conf"
2021-12-02 11:50:11.062 GMT [17] LOG:  skipping missing configuration file "/etc/config/user.conf"
2021-12-02 11:50:11.075 UTC [17] LOG:  starting MSSQLServer 13.2 on x86_64-pc-linux-musl, compiled by gcc (Alpine 10.2.1_pre1) 10.2.1 20201203, 64-bit
2021-12-02 11:50:11.075 UTC [17] LOG:  listening on IPv4 address "0.0.0.0", port 5432
2021-12-02 11:50:11.075 UTC [17] LOG:  listening on IPv6 address "::", port 5432
2021-12-02 11:50:11.081 UTC [17] LOG:  listening on Unix socket "/var/run/mssqlserverql/.s.PGSQL.5432"
2021-12-02 11:50:11.088 UTC [30] LOG:  database system was interrupted; last known up at 2021-12-02 11:50:10 UTC
2021-12-02 11:50:11.148 UTC [30] LOG:  entering standby mode
2021-12-02 11:50:11.154 UTC [30] LOG:  redo starts at 0/8000028
2021-12-02 11:50:11.157 UTC [30] LOG:  consistent recovery state reached at 0/8000100
2021-12-02 11:50:11.157 UTC [17] LOG:  database system is ready to accept read only connections
2021-12-02 11:50:11.162 UTC [35] LOG:  started streaming WAL from primary at 0/9000000 on timeline 2

```

You can see above that this pod is streaming wal from primary as replica. It verifies that we have successfully scaled up.

#### Scale Down

Here, we are going to remove 1 replica from our cluster using horizontal scaling.

**Create MSSQLServerOpsRequest:**

To scale down your cluster, you have to create a `MSSQLServerOpsRequest` cr with your desired number of members after scaling. Below is the YAML of the `MSSQLServerOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: ms-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ms
  horizontalScaling:
    replicas: 4
```

Let's create the `MSSQLServerOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/scaling/horizontal-scaling/scale-horizontally/yamls/ms-scale-down.yaml
mssqlserveropsrequest.ops.kubedb.com/ms-scale-down created
```

**Verify Scale-down Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale down the PetSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `MSSQLServer` object.

Now, we will wait for `MSSQLServerOpsRequest` to be successful. Run the following command to watch `MSSQLServerOpsRequest` cr,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo ms-scale-down
Every 2.0s: kubectl get mssqlserveropsrequest -n demo ms-scale-down    emon-r7: Thu Dec  2 18:15:37 2021

NAME            TYPE                STATUS       AGE
ms-scale-down   HorizontalScaling   Successful   115s


```

You can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest`, we shall see that the `MSSQLServer` cluster is scaled down.

```bash
$ kubectl describe  mssqlserveropsrequest -n demo ms-scale-down
Name:         ms-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2021-12-02T12:13:42Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:horizontalScaling:
          .:
          f:replicas:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-12-02T12:13:42Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-12-02T12:13:42Z
  Resource Version:  52120
  UID:               c69ea56e-e21c-4b1e-8a80-76f1b74ef2ba
Spec:
  Database Ref:
    Name:  ms
  Horizontal Scaling:
    Replicas:  4
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-12-02T12:13:42Z
    Message:               MSSQLServer ops request is horizontally scaling database
    Observed Generation:   1
    Reason:                Progressing
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2021-12-02T12:14:42Z
    Message:               Successfully Horizontally Scaled Down
    Observed Generation:   1
    Reason:                ScalingDown
    Status:                True
    Type:                  ScalingDown
    Last Transition Time:  2021-12-02T12:14:42Z
    Message:               Successfully Horizontally Scaled MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason          Age    From                        Message
  ----    ------          ----   ----                        -------
  Normal  PauseDatabase   2m31s  KubeDB Enterprise Operator  Pausing MSSQLServer demo/ms
  Normal  PauseDatabase   2m31s  KubeDB Enterprise Operator  Successfully paused MSSQLServer demo/ms
  Normal  ScalingDown     91s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Down
  Normal  ResumeDatabase  91s    KubeDB Enterprise Operator  Resuming MSSQLServer demo/ms
  Normal  ResumeDatabase  91s    KubeDB Enterprise Operator  Successfully resumed MSSQLServer demo/ms
  Normal  Successful      91s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify whether the number of members has decreased to meet up the desired state, Let's check, the mssqlserver status if it's ready then the scale-down is successful.

```bash
$ kubectl get mssqlserver -n demo ms
Every 3.0s: kubectl get mssqlserver -n demo ms                         emon-r7: Thu Dec  2 18:16:39 2021

NAME   VERSION   STATUS   AGE
ms     13.2      Ready    7h26m

```

You can see above that our `MSSQLServer` cluster now has a total of 4 members. It verifies that we have successfully scaled down.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ms -n demo ms
kubectl delete mssqlserveropsrequest -n demo ms-scale-up
kubectl delete mssqlserveropsrequest -n demo ms-scale-down
```
