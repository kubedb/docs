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

This guide will show you how to use the `KubeDB` operator to set up a `ProxySQL` server in front of an externally managed [Amazon Aurora](https://aws.amazon.com/rds/aurora/) (MySQL-compatible) cluster, with automatic read/write query splitting and failover-aware routing powered by ProxySQL's own native Aurora integration.

> Verified end-to-end against a live Amazon Aurora MySQL cluster and a real Kubernetes cluster, including five real `aws rds failover-db-cluster` events. The command output below reflects those verified runs (hostnames replaced with the placeholder values used throughout this guide).

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

Aurora is not a KubeDB-managed engine, so there's no `MySQL`/`MariaDB`/`PerconaXtraDB` CR for the operator to inspect. It also doesn't behave like ordinary MySQL replication for the purpose of picking a writer, and — this is the part that matters most — its **writer/reader cluster endpoints are floating DNS records**, not stable addresses.

For an Aurora backend, the operator configures ProxySQL's own native `mysql_aws_aurora_hostgroups` mechanism rather than the generic `mysql_replication_hostgroups`/`innodb_read_only`-polling approach used by simpler third-party guides. Given a single seed connection, ProxySQL:

- **Auto-discovers every instance** in the Aurora cluster by querying Aurora's own `information_schema.replica_host_status` table — you don't list readers by hand, and new/removed replicas are picked up automatically.
- **Tracks writer/reader role from that same Aurora-native metadata** (comparing each instance's `SESSION_ID` to `MASTER_SESSION_ID`), not by polling `innodb_read_only` through Aurora's floating cluster/reader DNS endpoints.
- **Connects to stable, individual instance hostnames** it discovers (e.g. `your-instance-1.xxxxx.<region>.rds.amazonaws.com`), which never change identity on a failover — only their reported role does. This is what makes it fail over cleanly and quickly: see [Failover behavior](#failover-behavior) below.

## Create an AppBinding for your Aurora cluster

Since Aurora isn't provisioned by KubeDB, you create the `AppBinding` (and its credentials `Secret`) by hand, pointing at your cluster's writer/cluster endpoint. This endpoint is used only as the **bootstrap seed connection** ProxySQL uses to start discovery — not for ongoing production traffic.

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
    url: tcp://aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com:3306
  secret:
    name: aurora-auth
  version: "8.0.mysql_aurora.3.05.2"
```

`spec.type` : Must be set to `kubedb.com/aws-aurora` — this is how the operator recognizes the AppBinding as an external Aurora backend rather than a KubeDB-managed one, and switches on Aurora-specific `mysql_servers`/`mysql_aws_aurora_hostgroups` generation.

`spec.clientConfig.url` : Your Aurora cluster's **writer (cluster) endpoint**, in `tcp://<host>:<port>` form. Use the `tcp://` scheme specifically — it's threaded through into the MySQL driver's own connection-string format internally, and any other scheme breaks that.

`spec.secret.name` : The secret holding the Aurora master username/password.

> **Note (TLS):** To connect to a TLS-enabled Aurora cluster (e.g. `require_secure_transport=ON`), set `spec.clientConfig.caBundle` to the base64-encoded RDS CA bundle (`https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem` or your region's bundle).

By default, the operator derives the **domain name** ProxySQL needs to turn discovered instance identifiers into hostnames from the writer endpoint, using AWS's stable naming convention — stripping the leading `<cluster-id>.cluster-` label (e.g. `aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com` becomes `.c9akciq32.us-east-1.rds.amazonaws.com`). If your cluster doesn't follow that convention — for example, an Aurora Global Database secondary region — set it explicitly via `spec.parameters`:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: aurora-appbinding
  namespace: demo
spec:
  type: kubedb.com/aws-aurora
  clientConfig:
    url: tcp://aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com:3306
  secret:
    name: aurora-auth
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: AuroraConfiguration
    domainName: .c9akciq32.us-east-1.rds.amazonaws.com
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

### Tuning discovery and routing weight

Set `spec.backend.aurora` to tune how ProxySQL's native Aurora monitor behaves for this specific ProxySQL instance — for example, checking for role changes more aggressively than the 1-second default, or weighting newly-discovered readers differently:

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
      newReaderWeight: 800
      maxLagMs: 30000
      checkIntervalMs: 1000
  deletionPolicy: WipeOut
```

`newReaderWeight` : the ProxySQL `mysql_servers.weight` assigned to each reader instance ProxySQL discovers (default `1000`).

`maxLagMs` : excludes a reader instance from the read pool once its measured replication lag exceeds this many milliseconds (default `600000` — 10 minutes).

`checkIntervalMs` : how often ProxySQL polls Aurora's `replica_host_status` for role/topology changes (default `1000`). Lower values detect a failover faster at the cost of more frequent checks.

```bash
$ kubectl apply -f sample-proxysql.yaml
proxysql.kubedb.com/aurora-proxy configured
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

(`admin`/`admin` is only the bootstrap default; once the pod finishes its first reconcile, the operator rotates the admin login to the generated credentials in the `<name>-auth` secret — read that secret's `username`/`password` keys if a later `kubectl exec` into the admin panel gets `Access denied`.)

The operator has configured `mysql_aws_aurora_hostgroups` with the domain name and tuning from above:

```bash
ProxySQLAdmin > select writer_hostgroup,reader_hostgroup,domain_name,new_reader_weight,max_lag_ms,check_interval_ms from mysql_aws_aurora_hostgroups;
+------------------+------------------+-----------------------------------------+-------------------+------------+--------------------+
| writer_hostgroup | reader_hostgroup | domain_name                              | new_reader_weight | max_lag_ms | check_interval_ms |
+------------------+------------------+-----------------------------------------+-------------------+------------+--------------------+
| 2                | 3                | .c9akciq32.us-east-1.rds.amazonaws.com  | 800               | 30000      | 1000               |
+------------------+------------------+-----------------------------------------+-------------------+------------+--------------------+
1 row in set (0.001 sec)
```

`mysql_servers` only holds the one bootstrap seed row you configured on the AppBinding — deliberately placed in the reader hostgroup (`3`), so that even a stale seed can only ever affect a read, never a write:

```bash
ProxySQLAdmin > select hostgroup_id,hostname,status from mysql_servers;
+--------------+-------------------------------------------------------------+---------+
| hostgroup_id | hostname                                                     | status  |
+--------------+-------------------------------------------------------------+---------+
| 3            | aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com    | ONLINE  |
+--------------+-------------------------------------------------------------+---------+
```

`runtime_mysql_servers` shows what ProxySQL actually discovered and is routing to — one stable, individual hostname per real Aurora instance, correctly classified into the writer (`2`) or reader (`3`) hostgroup:

```bash
ProxySQLAdmin > select hostgroup_id,hostname,status from runtime_mysql_servers;
+--------------+---------------------------------------------------------------------+---------+
| hostgroup_id | hostname                                                             | status  |
+--------------+---------------------------------------------------------------------+---------+
| 2            | aurora-demo-instance-1.c9akciq32.us-east-1.rds.amazonaws.com         | ONLINE  |
| 3            | aurora-demo-instance-1-reader.c9akciq32.us-east-1.rds.amazonaws.com  | ONLINE  |
| 3            | aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com            | ONLINE  |
+--------------+---------------------------------------------------------------------+---------+
```

### Check Traffic Proxy

Connect through the `aurora-proxy` service on port `6033` (data-plane, not the `6032` admin panel used above) as the Aurora master user and run a mix of writes and reads. Pass the password via the `MYSQL_PWD` environment variable rather than `-p` directly, so it doesn't end up in shell history or show up in `ps` output inside the container:

```bash
$ kubectl exec -it -n demo aurora-proxy-0 -c proxysql -- env MYSQL_PWD='<your-master-password>' mysql -uadmin -h127.0.0.1 -P6033 -e "
CREATE DATABASE IF NOT EXISTS proxytest;
CREATE TABLE IF NOT EXISTS proxytest.t1 (id INT PRIMARY KEY AUTO_INCREMENT, note VARCHAR(64));
INSERT INTO proxytest.t1 (note) VALUES ('via-proxysql-aurora-writer');
SELECT * FROM proxytest.t1;
"
+----+----------------------------+
| id | note                       |
+----+----------------------------+
|  1 | via-proxysql-aurora-writer |
+----+----------------------------+
```

The `SELECT` here is routed to the reader hostgroup right after a write to the writer hostgroup, and happened to see the row immediately in this run. Aurora replicas apply changes asynchronously, so under real load a read immediately following a write can occasionally miss it for a moment — if you don't see the row back, retry the `SELECT` rather than treating it as a failure.

Back in the admin panel, `stats_mysql_connection_pool` confirms the split: the three writes (`CREATE DATABASE`/`CREATE TABLE`/`INSERT`) went to hostgroup `2` (the discovered writer instance), and the `SELECT` went to hostgroup `3` (the discovered reader instance):

```bash
ProxySQLAdmin > select hostgroup,srv_host,Queries from stats_mysql_connection_pool where Queries > 0;
+-----------+----------------------------------------------------------------------+---------+
| hostgroup | srv_host                                                              | Queries |
+-----------+----------------------------------------------------------------------+---------+
| 2         | aurora-demo-instance-1.c9akciq32.us-east-1.rds.amazonaws.com         | 3       |
| 3         | aurora-demo-instance-1-reader.c9akciq32.us-east-1.rds.amazonaws.com  | 1       |
+-----------+----------------------------------------------------------------------+---------+
```

## Failover behavior

We tested this against **five real `aws rds failover-db-cluster` events** on a live cluster, watching a tight probe loop plus a direct, ProxySQL-bypassing connection to establish ground truth independently of ProxySQL's own view.

An earlier design (ProxySQL's generic `mysql_replication_hostgroups` + polling `innodb_read_only` through Aurora's floating writer/reader DNS endpoints — the approach most third-party guides describe) broke down badly under real failover conditions: ProxySQL's own internal DNS cache, which is consulted by every new backend connection it opens, could keep routing to a pre-failover IP for minutes after Aurora itself had already failed over cleanly. In our testing that produced a **sticky 1-2 minute write outage that did not self-correct** and needed manual operator intervention to clear.

The native `mysql_aws_aurora_hostgroups` mechanism this operator uses instead doesn't have that problem, because it doesn't route production traffic through Aurora's floating endpoints on an ongoing basis — only the one-time bootstrap seed connection touches them, and that seed is confined to the reader hostgroup specifically so it can never cause a write failure. In our final validation run: **a single transient connection error, zero write failures, and full convergence to the correct topology in about 25 seconds** — no manual intervention needed.

**Practical takeaway:** you should still expect a brief window (well under a minute, in our testing) of transient connection errors right as Aurora promotes a new writer — that's Aurora's own failover completing, not something any proxy can route around instantaneously. Application code should retry on connection errors for a few seconds after a known failover event. You should *not* need to intervene manually, and writes should not fail for an extended period the way they could with the floating-endpoint approach.

## Conclusion

In this tutorial we've seen how to point KubeDB ProxySQL at an externally managed AWS Aurora cluster using ProxySQL's native `mysql_aws_aurora_hostgroups` auto-discovery, how its writer/reader routing differs from the KubeDB-managed Group Replication and Galera backends, and what to expect from it during a real Aurora failover. Checkout the other backend guides and [Reconfigure](/docs/guides/proxysql/reconfigure/overview/index.md) docs to learn more.
