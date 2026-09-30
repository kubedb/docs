---
title: Proxy AWS Aurora With KubeDB ProxySQL
menu:
  docs_{{ .version }}:
    identifier: aurora-backend
    name: AWS Aurora
    parent: proxysql-backends
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB ProxySQL with AWS Aurora

This guide will show you how to use the `KubeDB` operator to set up a `ProxySQL` server in front of an externally managed [Amazon Aurora](https://aws.amazon.com/rds/aurora/) (MySQL-compatible) cluster, with automatic read/write query splitting between Aurora's writer and reader endpoints.

> This guide documents the shape of the feature as implemented in the operator. It has not yet been verified end-to-end against a live Aurora cluster — treat the example output below as illustrative of what the operator generates, not a captured session.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/index.md)
  - [AppBinding](/docs/guides/proxysql/concepts/appbinding/index.md)

- You need an existing Amazon Aurora (MySQL-compatible) DB cluster, reachable from your Kubernetes cluster, along with its master username/password. Unlike the [MySQL Group Replication](/docs/guides/proxysql/backends/mysqlgrp/index.md), [MariaDB Galera](/docs/guides/proxysql/backends/mariadb-galera/index.md), and [Percona XtraDB Galera](/docs/guides/proxysql/backends/xtradb-galera/external/index.md) backends, Aurora is never a KubeDB-managed database — KubeDB only knows how to point ProxySQL at it.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Why Aurora needs its own backend type

Aurora is not a KubeDB-managed engine, so there's no `MySQL`/`MariaDB`/`PerconaXtraDB` CR for the operator to inspect. It also doesn't behave like ordinary MySQL replication for the purpose of picking a writer:

- ProxySQL normally determines a server's role from the `read_only` global variable. **Aurora reports it via `innodb_read_only` instead.** The operator configures ProxySQL's `mysql_replication_hostgroups` table with `check_type = "innodb_read_only"` specifically for an Aurora backend, so the writer/reader split keeps tracking Aurora correctly (including across an Aurora failover).
- Aurora already exposes a stable **writer/cluster endpoint** and a separate, load-balanced **reader endpoint** — so, unlike the Galera/Group-Replication backends, the operator doesn't need to discover individual pod IPs. It registers exactly one `mysql_servers` row per endpoint: the writer endpoint in hostgroup `2`, the reader endpoint in hostgroup `3`.

## Create an AppBinding for your Aurora cluster

Since Aurora isn't provisioned by KubeDB, you create the `AppBinding` (and its credentials `Secret`) by hand, pointing at your cluster's writer/cluster endpoint.

First, create a secret with your Aurora cluster's master credentials:

```bash
$ kubectl create secret generic aurora-auth -n demo \
    --from-literal=username=admin \
    --from-literal=password='<your-master-password>' \
    --type=kubernetes.io/basic-auth
secret/aurora-auth created
```

Then create the AppBinding:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: aurora-appbinding
  namespace: demo
spec:
  type: kubedb.com/aws-aurora
  clientConfig:
    url: mysql://aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com:3306
  secret:
    name: aurora-auth
  version: "8.0.mysql_aurora.3.05.2"
```

`spec.type` : Must be set to `kubedb.com/aws-aurora` — this is how the operator recognizes the AppBinding as an external Aurora backend rather than a KubeDB-managed one, and switches on Aurora-specific `mysql_servers`/`mysql_replication_hostgroups` generation.

`spec.clientConfig.url` : Your Aurora cluster's **writer (cluster) endpoint**, in `mysql://<host>:<port>` form. Do not point this at an instance endpoint or the reader endpoint.

`spec.secret.name` : The secret holding the Aurora master username/password.

By default, the operator derives the **reader endpoint** from the writer endpoint using AWS's stable naming convention — replacing the `cluster-` label with `cluster-ro-` (e.g. `aurora-demo.cluster-c9akciq32...` becomes `aurora-demo.cluster-ro-c9akciq32...`). If your reader endpoint doesn't follow that convention — for example, an Aurora Global Database secondary-region endpoint, or a custom endpoint — set it explicitly via `spec.parameters`:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: aurora-appbinding
  namespace: demo
spec:
  type: kubedb.com/aws-aurora
  clientConfig:
    url: mysql://aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com:3306
  secret:
    name: aurora-auth
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: AuroraConfiguration
    readerEndpoint: aurora-demo.cluster-ro-c9akciq32.us-east-1.rds.amazonaws.com
  version: "8.0.mysql_aurora.3.05.2"
```

Apply the AppBinding:

```bash
$ kubectl apply -f aurora-appbinding.yaml
appbinding.appcatalog.appscode.com/aurora-appbinding created
```

## Deploy ProxySQL Server

With the following yaml we are going to create our ProxySQL server, referencing the AppBinding we just created:

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: aurora-proxy
  namespace: demo
spec:
  version: "3.0.1-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: aurora-appbinding
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f sample-proxysql.yaml
proxysql.kubedb.com/aurora-proxy created
```

### Tuning writer/reader routing weight

Two different ProxySQL deployments may want to route to the same Aurora cluster differently — for example, an OLTP-facing ProxySQL versus a reporting-facing one. Set `spec.backend.aurora` to override the default `mysql_servers` weight (default `1000` for both writer and reader) and reader replication-lag tolerance (default `0`, disabled) for this specific ProxySQL instance:

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: aurora-proxy
  namespace: demo
spec:
  version: "3.0.1-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: aurora-appbinding
    aurora:
      writerWeight: 1000
      readerWeight: 800
      maxReplicationLag: 30
  deletionPolicy: WipeOut
```

Let's wait for the ProxySQL to be Ready.

```bash
$ kubectl get prx -n demo
NAME           VERSION        STATUS   AGE
aurora-proxy   3.0.1-debian   Ready    2m
```

### Check Internal Configuration

Let's exec into the ProxySQL server pod and get into the admin panel.

```bash
$ kubectl exec -it -n demo aurora-proxy-0 -- bash
proxysql@aurora-proxy-0:/$  mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > "
ProxySQLAdmin >
```

The operator has configured `mysql_replication_hostgroups` with Aurora's `innodb_read_only` check type:

```bash
ProxySQLAdmin > select * from mysql_replication_hostgroups;
+------------------+------------------+------------------+------------+
| writer_hostgroup | reader_hostgroup | check_type       | comment    |
+------------------+------------------+------------------+------------+
| 2                | 3                | innodb_read_only | aws-aurora |
+------------------+------------------+------------------+------------+
1 row in set (0.001 sec)
```

And `mysql_servers` has one row for the writer endpoint (hostgroup `2`) and one for the reader endpoint (hostgroup `3`):

```bash
ProxySQLAdmin > select hostgroup_id,hostname,port,weight,max_replication_lag from mysql_servers;
+--------------+---------------------------------------------------------+------+--------+----------------------+
| hostgroup_id | hostname                                                 | port | weight | max_replication_lag |
+--------------+---------------------------------------------------------+------+--------+----------------------+
| 2            | aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com    | 3306 | 1000   | 0                    |
| 3            | aurora-demo.cluster-ro-c9akciq32.us-east-1.rds.amazonaws.com | 3306 | 1000   | 0                    |
+--------------+---------------------------------------------------------+------+--------+----------------------+
2 rows in set (0.001 sec)
```

As Aurora fails over, ProxySQL's monitor keeps polling `innodb_read_only` on both endpoints (the monitor user is granted `REPLICATION CLIENT` for this) and moves servers between the writer/reader hostgroups accordingly — you don't need to update the AppBinding or the ProxySQL CR when that happens, since the endpoint hostnames themselves don't change, only which underlying instance answers them.

From here on, connecting through the `aurora-proxy` service on port `6033` and verifying read/write query splitting works the same way as in the [MySQL Group Replication guide](/docs/guides/proxysql/backends/mysqlgrp/index.md#check-traffic-proxy).

## Conclusion

In this tutorial we've seen how to point KubeDB ProxySQL at an externally managed AWS Aurora cluster, and how its writer/reader routing differs from the KubeDB-managed Group Replication and Galera backends. Checkout the other backend guides and [Reconfigure](/docs/guides/proxysql/reconfigure/overview/index.md) docs to learn more.
