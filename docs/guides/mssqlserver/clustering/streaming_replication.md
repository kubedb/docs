---
title: Using MSSQLServer Streaming Replication
menu:
  docs_{{ .version }}:
    identifier: pg-streaming-replication-clustering
    name: Streaming Replication
    parent: pg-clustering-mssqlserver
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Streaming Replication

Streaming Replication provides *asynchronous* replication to one or more *standby* servers.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster.
If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create PostgreSQL with Streaming replication

The example below demonstrates KubeDB PostgreSQL for Streaming Replication

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: ha-mssqlserver
  namespace: demo
spec:
  version: "13.13"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

In this examples:

- This `MSSQLServer` object creates three PostgreSQL servers, indicated by the **`replicas`** field.
- One server will be *primary* and two others will be *warm standby* servers, default of **`spec.standbyMode`**

### What is Streaming Replication

Streaming Replication allows a *standby* server to stay more up-to-date by shipping and applying the [WAL XLOG](http://www.mssqlserverql.org/docs/9.6/static/wal.html)
records continuously. The *standby* connects to the *primary*, which streams WAL records to the *standby* as they're generated, without waiting for the WAL file to be filled.

Streaming Replication is **asynchronous** by default. As a result, there is a small delay between committing a transaction in the *primary* and the changes becoming visible in the *standby*.

### Streaming Replication setup

Following parameters are set in `mssqlserverql.conf` for both *primary* and *standby* server

```bash
wal_level = replica
max_wal_senders = 99
wal_keep_segments = 32
```

Here,

- _wal_keep_segments_ specifies the minimum number of past log file segments kept in the pg_xlog directory.

And followings are in `recovery.conf` for *standby* server

```bash
standby_mode = on
trigger_file = '/tmp/pg-failover-trigger'
recovery_target_timeline = 'latest'
primary_conninfo = 'application_name=$HOSTNAME host=$PRIMARY_HOST'
```

Here,

- _trigger_file_ is created to trigger a *standby* to take over as *primary* server.
- *$PRIMARY_HOST* holds the Kubernetes Service name that targets *primary* server

Now create this MSSQLServer object with Streaming Replication support

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/clustering/ha-mssqlserver.yaml
mssqlserver.kubedb.com/ha-mssqlserver created
```

KubeDB operator creates three Pod as PostgreSQL server.

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=ha-mssqlserver" --show-labels
NAME            READY   STATUS    RESTARTS   AGE   LABELS
ha-mssqlserver-0   1/1     Running   0          20s   controller-revision-hash=ha-mssqlserver-6b7998ccfd,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver,kubedb.com/role=primary,petset.kubernetes.io/pod-name=ha-mssqlserver-0
ha-mssqlserver-1   1/1     Running   0          16s   controller-revision-hash=ha-mssqlserver-6b7998ccfd,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver,kubedb.com/role=replica,petset.kubernetes.io/pod-name=ha-mssqlserver-1
ha-mssqlserver-2   1/1     Running   0          10s   controller-revision-hash=ha-mssqlserver-6b7998ccfd,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver,kubedb.com/role=replica,petset.kubernetes.io/pod-name=ha-mssqlserver-2
```

Here,

- Pod `ha-mssqlserver-0` is serving as *primary* server, indicated by label `kubedb.com/role=primary`
- Pod `ha-mssqlserver-1` & `ha-mssqlserver-2` both are serving as *standby* server, indicated by label `kubedb.com/role=replica`

And two services for MSSQLServer `ha-mssqlserver` are created.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=ha-mssqlserver"
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
ha-mssqlserver            ClusterIP   10.102.19.49   <none>        5432/TCP   4m
ha-mssqlserver-replicas   ClusterIP   10.97.36.117   <none>        5432/TCP   4m
```

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=ha-mssqlserver" -o=custom-columns=NAME:.metadata.name,SELECTOR:.spec.selector
NAME                   SELECTOR
ha-mssqlserver            map[app.kubernetes.io/name:mssqlserveres.kubedb.com app.kubernetes.io/instance:ha-mssqlserver kubedb.com/role:primary]
ha-mssqlserver-replicas   map[app.kubernetes.io/name:mssqlserveres.kubedb.com app.kubernetes.io/instance:ha-mssqlserver]
```

Here,

- Service `ha-mssqlserver` targets Pod `ha-mssqlserver-0`, which is *primary* server, by selector `app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver,kubedb.com/role=primary`.
- Service `ha-mssqlserver-replicas` targets all Pods (*`ha-mssqlserver-0`*, *`ha-mssqlserver-1`* and *`ha-mssqlserver-2`*) with label `app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver`.

>These *standby* servers are asynchronous *warm standby* server. That means, you can only connect to *primary* sever.

Now connect to this *primary* server Pod `ha-mssqlserver-0` using pgAdmin installed in [quickstart](/docs/guides/mssqlserver/quickstart/quickstart.md#before-you-begin) tutorial.

**Connection information:**

- Host name/address: you can use any of these
  - Service: `ha-mssqlserver.demo`
  - Pod IP: (`$kubectl get pods ha-mssqlserver-0 -n demo -o yaml | grep podIP`)
- Port: `5432`
- Maintenance database: `mssqlserver`
- Username: Run following command to get *username*,

  ```bash
  $ kubectl get secrets -n demo ha-mssqlserver-auth -o jsonpath='{.data.\POSTGRES_USER}' | base64 -d
  mssqlserver
  ```

- Password: Run the following command to get *password*,

  ```bash
  $ kubectl get secrets -n demo ha-mssqlserver-auth -o jsonpath='{.data.\POSTGRES_PASSWORD}' | base64 -d
  MHRrOcuyddfh3YpU
  ```

You can check `pg_stat_replication` information to know who is currently streaming from *primary*.

```bash
mssqlserver=# select * from pg_stat_replication;
```

 pid | usesysid | usename  | application_name | client_addr | client_port |         backend_start         |   state   | sent_location | write_location | flush_location | replay_location | sync_priority | sync_state
-----|----------|----------|------------------|-------------|-------------|-------------------------------|-----------|---------------|----------------|----------------|-----------------|---------------|------------
  89 |       10 | mssqlserver | ha-mssqlserver-2    | 172.17.0.8  |       35306 | 2018-02-09 04:27:11.674828+00 | streaming | 0/5000060     | 0/5000060      | 0/5000060      | 0/5000060       |             0 | async
  90 |       10 | mssqlserver | ha-mssqlserver-1    | 172.17.0.7  |       42400 | 2018-02-09 04:27:13.716104+00 | streaming | 0/5000060     | 0/5000060      | 0/5000060      | 0/5000060       |             0 | async

Here, both `ha-mssqlserver-1` and `ha-mssqlserver-2` are streaming asynchronously from *primary* server.

### Lease Duration

Get the mssqlserver CRD at this point.

```yaml
$ kubectl get pg -n demo   ha-mssqlserver -o yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  creationTimestamp: "2019-02-07T12:14:05Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: ha-mssqlserver
  namespace: demo
  resourceVersion: "44966"
  selfLink: /apis/kubedb.com/v1/namespaces/demo/mssqlserveres/ha-mssqlserver
  uid: dcf6d96a-2ad1-11e9-9d44-080027154f61
spec:
  authSecret:
    name: ha-mssqlserver-auth
  leaderElection:
    leaseDurationSeconds: 15
    renewDeadlineSeconds: 10
    retryPeriodSeconds: 2
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    dataSource: null
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: Halt
  version: "10.2"-v5
status:
  observedGeneration: 2$4213139756412538772
  phase: Running
```

There are three fields under MSSQLServer CRD's `spec.leaderElection`. These values defines how fast the leader election can happen.

- leaseDurationSeconds: This is the duration in seconds that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack. Default 15 secs.
- renewDeadlineSeconds: This is the duration in seconds that the acting master will retry refreshing leadership before giving up. Normally, LeaseDuration * 2 / 3. Default 10 secs.
- retryPeriodSeconds: This is the duration in seconds the LeaderElector clients should wait between tries of actions. Normally, LeaseDuration / 3. Default 2 secs.

If the Cluster machine is powerful, user can reduce the times. But, Do not make it so little, in that case MSSQLServer will restarts very often.

### Automatic failover

If *primary* server fails, another *standby* server will take over and serve as *primary*.

Delete Pod `ha-mssqlserver-0` to see the failover behavior.

```bash
kubectl delete pod -n demo ha-mssqlserver-0
```

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=ha-mssqlserver" --show-labels
NAME            READY     STATUS    RESTARTS   AGE       LABELS
ha-mssqlserver-0   1/1       Running   0          10s       controller-revision-hash=ha-mssqlserver-b8b4b5fc4,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver,kubedb.com/role=replica,petset.kubernetes.io/pod-name=ha-mssqlserver-0
ha-mssqlserver-1   1/1       Running   0          52m       controller-revision-hash=ha-mssqlserver-b8b4b5fc4,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver,kubedb.com/role=primary,petset.kubernetes.io/pod-name=ha-mssqlserver-1
ha-mssqlserver-2   1/1       Running   0          51m       controller-revision-hash=ha-mssqlserver-b8b4b5fc4,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=ha-mssqlserver,kubedb.com/role=replica,petset.kubernetes.io/pod-name=ha-mssqlserver-2
```

Here,

- Pod `ha-mssqlserver-1` is now serving as *primary* server
- Pod `ha-mssqlserver-0` and `ha-mssqlserver-2` both are serving as *standby* server

And result from `pg_stat_replication`

```bash
mssqlserver=# select * from pg_stat_replication;
```

 pid | usesysid | usename  | application_name | client_addr | client_port |         backend_start         |   state   | sent_location | write_location | flush_location | replay_location | sync_priority | sync_state
-----|----------|----------|------------------|-------------|-------------|-------------------------------|-----------|---------------|----------------|----------------|-----------------|---------------|------------
  57 |       10 | mssqlserver | ha-mssqlserver-0    | 172.17.0.6  |       52730 | 2018-02-09 04:33:06.051716|00 | streaming | 0/7000060     | 0/7000060      | 0/7000060      | 0/7000060       |             0 | async
  58 |       10 | mssqlserver | ha-mssqlserver-2    | 172.17.0.8  |       42824 | 2018-02-09 04:33:09.762168|00 | streaming | 0/7000060     | 0/7000060      | 0/7000060      | 0/7000060       |             0 | async

You can see here, now `ha-mssqlserver-0` and `ha-mssqlserver-2` are streaming asynchronously from `ha-mssqlserver-1`, our *primary* server.

<p align="center">
  <kbd>
    <img alt="recovered-mssqlserver"  src="/docs/images/mssqlserver/ha-mssqlserver.gif">
  </kbd>
</p>

[//]: # (If you want to know how this failover process works, [read here])

## Streaming Replication with `hot standby`

Streaming Replication also works with one or more *hot standby* servers.

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: hot-mssqlserver
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
```

In this examples:

- This `MSSQLServer` object creates three PostgreSQL servers, indicated by the **`replicas`** field.
- One server will be *primary* and two others will be *hot standby* servers, as instructed by **`spec.standbyMode`**

### `hot standby` setup

Following parameters are set in `mssqlserverql.conf` for *standby* server

```bash
hot_standby = on
```

Here,

- _hot_standby_ specifies that *standby* server will act as *hot standby*.

Now create this MSSQLServer object

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/clustering/hot-mssqlserver.yaml
mssqlserver "hot-mssqlserver" created
```

KubeDB operator creates three Pod as PostgreSQL server.

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=hot-mssqlserver" --show-labels
NAME             READY     STATUS    RESTARTS   AGE       LABELS
hot-mssqlserver-0   1/1       Running   0          1m        controller-revision-hash=hot-mssqlserver-6c48cfb5bb,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=hot-mssqlserver,kubedb.com/role=primary,petset.kubernetes.io/pod-name=hot-mssqlserver-0
hot-mssqlserver-1   1/1       Running   0          1m        controller-revision-hash=hot-mssqlserver-6c48cfb5bb,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=hot-mssqlserver,kubedb.com/role=replica,petset.kubernetes.io/pod-name=hot-mssqlserver-1
hot-mssqlserver-2   1/1       Running   0          48s       controller-revision-hash=hot-mssqlserver-6c48cfb5bb,app.kubernetes.io/name=mssqlserveres.kubedb.com,app.kubernetes.io/instance=hot-mssqlserver,kubedb.com/role=replica,petset.kubernetes.io/pod-name=hot-mssqlserver-2
```

Here,

- Pod `hot-mssqlserver-0` is serving as *primary* server, indicated by label `kubedb.com/role=primary`
- Pod `hot-mssqlserver-1` & `hot-mssqlserver-2` both are serving as *standby* server, indicated by label `kubedb.com/role=replica`

> These *standby* servers are asynchronous *hot standby* servers.

That means, you can connect to both *primary* and *standby* sever. But these *hot standby* servers only accept read-only queries.

Now connect to one of our *hot standby* servers Pod `hot-mssqlserver-2` using pgAdmin installed in [quickstart](/docs/guides/mssqlserver/quickstart/quickstart.md#before-you-begin) tutorial.

**Connection information:**

- Host name/address: you can use any of these
  - Service: `hot-mssqlserver-replicas.demo`
  - Pod IP: (`$kubectl get pods hot-mssqlserver-2 -n demo -o yaml | grep podIP`)
- Port: `5432`
- Maintenance database: `mssqlserver`
- Username: Run following command to get *username*,

  ```bash
  $ kubectl get secrets -n demo hot-mssqlserver-auth -o jsonpath='{.data.\POSTGRES_USER}' | base64 -d
  mssqlserver
  ```

- Password: Run the following command to get *password*,

  ```bash
  $ kubectl get secrets -n demo hot-mssqlserver-auth -o jsonpath='{.data.\POSTGRES_PASSWORD}' | base64 -d
  ZZgjjQMUdKJYy1W9
  ```

Try to create a database (write operation)

```bash
mssqlserver=# CREATE DATABASE standby;
ERROR:  cannot execute CREATE DATABASE in a read-only transaction
```

Failed to execute write operation. But it can execute following read query

```bash
mssqlserver=# select pg_last_xlog_receive_location();
 pg_last_xlog_receive_location
-------------------------------
 0/7000220
```

So, you can see here that you can connect to *hot standby* and it only accepts read-only queries.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo pg/ha-mssqlserver pg/hot-mssqlserver -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo pg/ha-mssqlserver pg/hot-mssqlserver

$ kubectl delete ns demo
```

## Next Steps

- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/mssqlserver/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
