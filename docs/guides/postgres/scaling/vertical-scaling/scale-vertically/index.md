---
title: Vertical Scaling Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-scaling-vertical-scale-vertically
    name: scale vertically
    parent: guides-postgres-scaling-vertical
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale Postgres Instance

This guide will show you how to use `kubeDP-Ops-Manager` to update the resources of a Postgres instance.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB-Provisioner` and `kubeDP-Ops-Manager` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/postgres/scaling/vertical-scaling/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/scaling/vertical-scaling/scale-vertically/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/scaling/vertical-scaling/scale-vertically/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

### Apply Vertical Scaling on Postgres Instance

Here, we are going to deploy a `Postgres` instance using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

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

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `Postgres`. You can use any non-deprecated version. Here, we are going to create a postgres using non-deprecated `Postgres` version `13.2`.

**Deploy Postgres:**

In this section, we are going to deploy a Postgres instance. Then, in the next section, we will update the resources of the database server using vertical scaling. Below is the YAML of the `Postgres` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: pg
  namespace: demo
spec:
  version: "13.2"
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
  terminationPolicy: WipeOut
```

Let's create the `Postgres` cr we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/vertical-scaling/scale-vertically/yamls/postgres.yaml
postgres.kubedb.com/pg created
```

**Check postgres Ready to Scale:**

`KubeDB-Provisioner` watches for `Postgres` objects using Kubernetes API. When a `Postgres` object is created, `KubeDB-Provisioner` will create a new StatefulSet, Services, and Secrets, etc.
Now, watch `Postgres` is going to be in `Running` state and also watch `StatefulSet` and its pod is created and going to be in `Running` state,

```bash
$ watch -n 3 kubectl get postgres -n demo pg
Every 3.0s: kubectl get postgres -n demo pg                         emon-r7: Thu Dec  2 10:53:54 2021

NAME   VERSION   STATUS   AGE
pg     13.2      Ready    3m16s

$ watch -n 3 kubectl get sts -n demo pg
Every 3.0s: kubectl get sts -n demo pg                              emon-r7: Thu Dec  2 10:54:31 2021

NAME   READY   AGE
pg     3/3     3m54s

$ watch -n 3 kubectl get pod -n demo
Every 3.0s: kubectl get pod -n demo                                 emon-r7: Thu Dec  2 10:55:29 2021

NAME   READY   STATUS    RESTARTS   AGE
pg-0   2/2     Running   0          4m51s
pg-1   2/2     Running   0          3m50s
pg-2   2/2     Running   0          3m46s

```

Let's check the `pg-0` Pod's postgres container's resources, As there are two containers, And Postgres container is the first container So it's index will be 0.

```bash
$ kubectl get pod -n demo pg-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}

```

Now, We are ready to apply a vertical scale on this postgres database.

#### Vertical Scaling

Here, we are going to update the resources of the postgres to meet up with the desired resources after scaling.

**Create PostgresOpsRequest:**

In order to update the resources of your database, you have to create a `PostgresOpsRequest` cr with your desired resources after scaling. Below is the YAML of the `PostgresOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: pg
  verticalScaling:
    postgres:
      requests:
        memory: "1200Mi"
        cpu: "0.7"
      limits:
        memory: "1200Mi"
        cpu: "0.7"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `pg` `Postgres` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.postgres` specifies the expected postgres container resources after scaling.

Let's create the `PostgresOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/vertical-scaling/scale-vertically/yamls/pg-vertical-scaling.yaml
postgresopsrequest.ops.kubedb.com/pg-scale-vertical created
```

**Verify Postgres resources updated successfully:**

If everything goes well, `KubeDB-Ops-Manager` will update the resources of the StatefulSet's `Pod` containers. After a successful scaling process is done, the `KubeDB-Ops-Manager` updates the resources of the `Postgres` object.

First, we will wait for `PostgresOpsRequest` to be successful. Run the following command to watch `PostgresOpsRequest` cr,

```bash
$ watch kubectl get postgresopsrequest -n demo pg-scale-vertical

Every 2.0s: kubectl get postgresopsrequest -n demo pg-scale-ve...  emon-r7: Thu Dec  2 11:09:49 2021

NAME                TYPE              STATUS       AGE
pg-scale-vertical   VerticalScaling   Successful   3m42s

```

We can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest`, we will see that the postgres resources are updated.

```bash
$ kubectl describe postgresopsrequest -n demo pg-scale-vertical
Name:         pg-scale-vertical
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
Metadata:
  Creation Timestamp:  2021-12-02T05:06:07Z
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
        f:type:
        f:verticalScaling:
          .:
          f:postgres:
            .:
            f:limits:
              .:
              f:cpu:
              f:memory:
            f:requests:
              .:
              f:cpu:
              f:memory:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-12-02T05:06:07Z
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
    Time:            2021-12-02T05:06:07Z
  Resource Version:  8452
  UID:               92d1e69f-c99a-4d0b-b8bf-e904e1336083
Spec:
  Database Ref:
    Name:  pg
  Type:    VerticalScaling
  Vertical Scaling:
    Postgres:
      Limits:
        Cpu:     0.7
        Memory:  1200Mi
      Requests:
        Cpu:     0.7
        Memory:  1200Mi
Status:
  Conditions:
    Last Transition Time:  2021-12-02T05:06:07Z
    Message:               Postgres ops request is vertically scaling database
    Observed Generation:   1
    Reason:                Progressing
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2021-12-02T05:06:07Z
    Message:               Successfully updated statefulsets resources
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2021-12-02T05:08:02Z
    Message:               SuccessfullyPerformedVerticalScaling
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2021-12-02T05:08:02Z
    Message:               Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason           Age    From                        Message
  ----    ------           ----   ----                        -------
  Normal  PauseDatabase    4m17s  KubeDB Enterprise Operator  Pausing Postgres demo/pg
  Normal  PauseDatabase    4m17s  KubeDB Enterprise Operator  Successfully paused Postgres demo/pg
  Normal  VerticalScaling  2m22s  KubeDB Enterprise Operator  SuccessfullyPerformedVerticalScaling
  Normal  ResumeDatabase   2m22s  KubeDB Enterprise Operator  Resuming PostgreSQL demo/pg
  Normal  ResumeDatabase   2m22s  KubeDB Enterprise Operator  Successfully resumed PostgreSQL demo/pg
  Normal  Successful       2m22s  KubeDB Enterprise Operator  Successfully Vertically Scaled Database

```

Now, we are going to verify whether the resources of the postgres instance has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo pg-0 -o json | jq '.spec.containers[0].resources'
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

The above output verifies that we have successfully scaled up the resources of the Postgres.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete postgres -n demo pg
kubectl delete postgresopsrequest -n demo pg-scale-vertical
```
