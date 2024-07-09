---
title: Horizontal Scaling Postgres Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-scaling-horizontal-scale-horizontally
    name: Scale Horizontally
    parent: guides-postgres-scaling-horizontal
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Postgres Cluster

This guide will show you how to use `KubeDB` Ops Manager to increase/decrease the number of members of a `Postgres` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/postgres/scaling/horizontal-scaling/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/scaling/horizontal-scaling/scale-horizontally/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/scaling/horizontal-scaling/scale-horizontally/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

### Apply Horizontal Scaling on Postgres Cluster

Here, we are going to deploy a `Postgres` Cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

#### Prepare Cluster

At first, we are going to deploy a Cluster server with 3 members. Then, we are going to add two additional members through horizontal scaling. Finally, we will remove 1 member from the cluster again via horizontal scaling.

**Find supported Postgres Version:**

When you have installed `KubeDB`, it has created `PostgresVersion` CR for all supported `Postgres` versions. Let's check the supported Postgres versions,

```bash
$ kubectl get postgresversion
NAME                       VERSION   DISTRIBUTION   DB_IMAGE                               DEPRECATED   AGE
10.16                      10.16     Official       postgres:10.16-alpine                               63s
10.16-debian               10.16     Official       postgres:10.16                                      63s
10.19                      10.19     Official       postgres:10.19-bullseye                             63s
10.19-bullseye             10.19     Official       postgres:10.19-bullseye                             63s
11.11                      11.11     Official       postgres:11.11-alpine                               63s
11.11-debian               11.11     Official       postgres:11.11                                      63s
11.14                      11.14     Official       postgres:11.14-alpine                               63s
11.14-bullseye             11.14     Official       postgres:11.14-bullseye                             63s
11.14-bullseye-postgis     11.14     PostGIS        postgis/postgis:11-3.1                              63s
12.6                       12.6      Official       postgres:12.6-alpine                                63s
12.6-debian                12.6      Official       postgres:12.6                                       63s
12.9                       12.9      Official       postgres:12.9-alpine                                63s
12.9-bullseye              12.9      Official       postgres:12.9-bullseye                              63s
12.9-bullseye-postgis      12.9      PostGIS        postgis/postgis:12-3.1                              63s
13.2                       13.2      Official       postgres:13.2-alpine                                63s
13.2-debian                13.2      Official       postgres:13.2                                       63s
13.5                       13.5      Official       postgres:13.5-alpine                                63s
13.5-bullseye              13.5      Official       postgres:13.5-bullseye                              63s
13.5-bullseye-postgis      13.5      PostGIS        postgis/postgis:13-3.1                              63s
14.1                       14.1      Official       postgres:14.1-alpine                                63s
14.1-bullseye              14.1      Official       postgres:14.1-bullseye                              63s
14.1-bullseye-postgis      14.1      PostGIS        postgis/postgis:14-3.1                              63s
9.6.21                     9.6.21    Official       postgres:9.6.21-alpine                              63s
9.6.21-debian              9.6.21    Official       postgres:9.6.21                                     63s
9.6.24                     9.6.24    Official       postgres:9.6.24-alpine                              63s
9.6.24-bullseye            9.6.24    Official       postgres:9.6.24-bullseye                            63s
timescaledb-2.1.0-pg11     11.11     TimescaleDB    timescale/timescaledb:2.1.0-pg11-oss                63s
timescaledb-2.1.0-pg12     12.6      TimescaleDB    timescale/timescaledb:2.1.0-pg12-oss                63s
timescaledb-2.1.0-pg13     13.2      TimescaleDB    timescale/timescaledb:2.1.0-pg13-oss                63s
timescaledb-2.5.0-pg14.1   14.1      TimescaleDB    timescale/timescaledb:2.5.0-pg14-oss                63s
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `Postgres`. You can use any non-deprecated version. Here, we are going to create a Postgres Cluster using `Postgres` `13.2`.

**Deploy Postgres Cluster:**

In this section, we are going to deploy a Postgres Cluster with 3 members. Then, in the next section we will scale-up the cluster using horizontal scaling. Below is the YAML of the `Postgres` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: pg
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

Let's create the `Postgres` cr we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/horizontal-scaling/scale-horizontally/postgres.yaml
postgres.kubedb.com/pg created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `Postgres` objects using Kubernetes API. When a `Postgres` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc. A secret called `pg-auth` (format: <em>{postgres-object-name}-auth</em>) will be created storing the password for postgres superuser.
Now, watch `Postgres` is going to `Running` state and also watch `PetSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get postgres -n demo pg
Every 3.0s: kubectl get postgres -n demo pg                        emon-r7: Thu Dec  2 15:31:16 2021

NAME   VERSION   STATUS   AGE
pg     13.2      Ready    4h40m


$ watch -n 3 kubectl get sts -n demo pg
Every 3.0s: kubectl get sts -n demo pg                             emon-r7: Thu Dec  2 15:31:38 2021

NAME   READY   AGE
pg     3/3     4h41m



$ watch -n 3 kubectl get pods -n demo
Every 3.0s: kubectl get pod -n demo                                emon-r7: Thu Dec  2 15:33:24 2021

NAME   READY   STATUS    RESTARTS   AGE
pg-0   2/2     Running   0          4h25m
pg-1   2/2     Running   0          4h26m
pg-2   2/2     Running   0          4h26m

```

Let's verify that the PetSet's pods have joined into cluster,

```bash
$ kubectl get secrets -n demo pg-auth -o jsonpath='{.data.\username}' | base64 -d
postgres

$ kubectl get secrets -n demo pg-auth -o jsonpath='{.data.\password}' | base64 -d
b3b5838EhjwsiuFU

```

So, we can see that our cluster has 3 members. Now, we are ready to apply the horizontal scale to this Postgres cluster.

#### Scale Up

Here, we are going to add 2 replicas in our Cluster using horizontal scaling.

**Create PostgresOpsRequest:**

To scale up your cluster, you have to create a `PostgresOpsRequest` cr with your desired number of replicas after scaling. Below is the YAML of the `PostgresOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pg
  horizontalScaling:
    replicas: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `pg` `Postgres` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the expected number of replicas after the scaling.

Let's create the `PostgresOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/horizontal-scaling/scale-horizontally/yamls/pg-scale-up.yaml
postgresopsrequest.ops.kubedb.com/pg-scale-up created
```

**Verify Scale-Up Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale up the PetSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `Postgres` object.

First, we will wait for `PostgresOpsRequest` to be successful. Run the following command to watch `PostgresOpsRequest` cr,

```bash
$ watch kubectl get postgresopsrequest -n demo pg-scale-up
Every 2.0s: kubectl get postgresopsrequest -n demo pg-scale-up     emon-r7: Thu Dec  2 17:57:36 2021

NAME          TYPE                STATUS       AGE
pg-scale-up   HorizontalScaling   Successful   8m23s

```

You can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest`, we will see that the `Postgres` cluster is scaled up.

```bash
kubectl describe postgresopsrequest -n demo pg-scale-up
Name:         pg-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
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
    Name:  pg
  Horizontal Scaling:
    Replicas:  5
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-12-02T11:49:13Z
    Message:               Postgres ops request is horizontally scaling database
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
    Message:               Successfully Horizontally Scaled Postgres
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason          Age    From                        Message
  ----    ------          ----   ----                        -------
  Normal  PauseDatabase   10m    KubeDB Enterprise Operator  Pausing Postgres demo/pg
  Normal  PauseDatabase   10m    KubeDB Enterprise Operator  Successfully paused Postgres demo/pg
  Normal  ScalingUp       9m17s  KubeDB Enterprise Operator  Successfully Horizontally Scaled Up
  Normal  ResumeDatabase  9m17s  KubeDB Enterprise Operator  Resuming PostgreSQL demo/pg
  Normal  ResumeDatabase  9m17s  KubeDB Enterprise Operator  Successfully resumed PostgreSQL demo/pg
  Normal  Successful      9m17s  KubeDB Enterprise Operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify whether the number of members has increased to meet up the desired state. So let's check the new pods logs to see if they have joined in the cluster as new replica.

```bash
$ kubectl logs   -n demo                 pg-4  -c postgres -f
waiting for the role to be decided ...
running the initial script ...
Running as Replica
Attempting pg_isready on primary
Attempting query on primary
take base basebackup...
2021-12-02 11:50:11.062 GMT [17] LOG:  skipping missing configuration file "/etc/config/user.conf"
2021-12-02 11:50:11.062 GMT [17] LOG:  skipping missing configuration file "/etc/config/user.conf"
2021-12-02 11:50:11.075 UTC [17] LOG:  starting PostgreSQL 13.2 on x86_64-pc-linux-musl, compiled by gcc (Alpine 10.2.1_pre1) 10.2.1 20201203, 64-bit
2021-12-02 11:50:11.075 UTC [17] LOG:  listening on IPv4 address "0.0.0.0", port 5432
2021-12-02 11:50:11.075 UTC [17] LOG:  listening on IPv6 address "::", port 5432
2021-12-02 11:50:11.081 UTC [17] LOG:  listening on Unix socket "/var/run/postgresql/.s.PGSQL.5432"
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

**Create PostgresOpsRequest:**

To scale down your cluster, you have to create a `PostgresOpsRequest` cr with your desired number of members after scaling. Below is the YAML of the `PostgresOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pg
  horizontalScaling:
    replicas: 4
```

Let's create the `PostgresOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/horizontal-scaling/scale-horizontally/yamls/pg-scale-down.yaml
postgresopsrequest.ops.kubedb.com/pg-scale-down created
```

**Verify Scale-down Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale down the PetSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `Postgres` object.

Now, we will wait for `PostgresOpsRequest` to be successful. Run the following command to watch `PostgresOpsRequest` cr,

```bash
$ watch kubectl get postgresopsrequest -n demo pg-scale-down
Every 2.0s: kubectl get postgresopsrequest -n demo pg-scale-down    emon-r7: Thu Dec  2 18:15:37 2021

NAME            TYPE                STATUS       AGE
pg-scale-down   HorizontalScaling   Successful   115s


```

You can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest`, we shall see that the `Postgres` cluster is scaled down.

```bash
$ kubectl describe  postgresopsrequest -n demo pg-scale-down
Name:         pg-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
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
    Name:  pg
  Horizontal Scaling:
    Replicas:  4
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-12-02T12:13:42Z
    Message:               Postgres ops request is horizontally scaling database
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
    Message:               Successfully Horizontally Scaled Postgres
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason          Age    From                        Message
  ----    ------          ----   ----                        -------
  Normal  PauseDatabase   2m31s  KubeDB Enterprise Operator  Pausing Postgres demo/pg
  Normal  PauseDatabase   2m31s  KubeDB Enterprise Operator  Successfully paused Postgres demo/pg
  Normal  ScalingDown     91s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Down
  Normal  ResumeDatabase  91s    KubeDB Enterprise Operator  Resuming PostgreSQL demo/pg
  Normal  ResumeDatabase  91s    KubeDB Enterprise Operator  Successfully resumed PostgreSQL demo/pg
  Normal  Successful      91s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify whether the number of members has decreased to meet up the desired state, Let's check, the postgres status if it's ready then the scale-down is successful.

```bash
$ kubectl get postgres -n demo pg
Every 3.0s: kubectl get postgres -n demo pg                         emon-r7: Thu Dec  2 18:16:39 2021

NAME   VERSION   STATUS   AGE
pg     13.2      Ready    7h26m

```

You can see above that our `Postgres` cluster now has a total of 4 members. It verifies that we have successfully scaled down.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pg -n demo pg
kubectl delete postgresopsrequest -n demo pg-scale-up
kubectl delete postgresopsrequest -n demo pg-scale-down
```
