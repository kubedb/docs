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

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/index.md)
  - [AppBinding](/docs/guides/proxysql/concepts/appbinding/index.md)

- You need an existing Amazon Aurora (MySQL-compatible) DB cluster, reachable from your Kubernetes cluster, along with its master username/password.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

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
    url: tcp://aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com:3306
    caBundle: <base64-encoded-ca-bundle>
  secret:
    name: aurora-auth
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: AuroraConfiguration
    domainName: .c9akciq32.us-east-1.rds.amazonaws.com
  version: "8.0.mysql_aurora.3.05.2"
```

`spec.type` : Must be set to `kubedb.com/aws-aurora`.

`spec.clientConfig.url` : Your Aurora cluster's **writer (cluster) endpoint**, in `tcp://<host>:<port>` form.

`spec.clientConfig.caBundle` : The base64-encoded RDS CA bundle, used for TLS between ProxySQL and Aurora. You can get it with the following command. If you don't want backend TLS, remove this field.

```bash
$ curl -s https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem | base64 -w0
```

`spec.secret.name` : The secret holding the Aurora master username/password.

`spec.parameters.domainName` : The domain suffix of your Aurora instance endpoints, starting with a dot. An instance endpoint is `<instance-id><domainName>`, for example `aurora-demo-instance-1.c9akciq32.us-east-1.rds.amazonaws.com`. This field is optional. If it is not set, it is taken from the writer endpoint by removing the leading `<cluster-id>.cluster-` part.

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

Set `spec.configuration.init.inline.mysqlAWSAuroraHostgroups` to tune how ProxySQL's native Aurora monitor behaves for this specific ProxySQL instance — for example, checking for role changes more aggressively than the 1-second default, or weighting newly-discovered readers differently. The keys are the column names of ProxySQL's `mysql_aws_aurora_hostgroups` table:

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
  configuration:
    init:
      inline:
        mysqlAWSAuroraHostgroups:
          new_reader_weight: 800
          max_lag_ms: 30000
          check_interval_ms: 1000
  deletionPolicy: WipeOut
```

`new_reader_weight` : the ProxySQL `mysql_servers.weight` assigned to each reader instance ProxySQL discovers (default `1000`).

`max_lag_ms` : excludes a reader instance from the read pool once its measured replication lag exceeds this many milliseconds (default `600000` — 10 minutes).

`check_interval_ms` : how often ProxySQL polls Aurora's `replica_host_status` for role/topology changes (default `1000`). Lower values detect a failover faster at the cost of more frequent checks.

```bash
$ kubectl apply -f sample-proxysql.yaml
proxysql.kubedb.com/aurora-proxy created
```

Let's wait for the ProxySQL to be Ready.

```bash
$ kubectl get prx -n demo
NAME           VERSION        STATUS   AGE
aurora-proxy   3.0.1-debian   Ready    2m
```

### Check Internal Configuration

Let's get the admin credentials from the `aurora-proxy-auth` secret and get into the admin panel.

```bash
$ ADMIN_USER=$(kubectl get secret -n demo aurora-proxy-auth -o jsonpath='{.data.username}' | base64 -d)
$ ADMIN_PASS=$(kubectl get secret -n demo aurora-proxy-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -it -n demo aurora-proxy-0 -- mysql -u"$ADMIN_USER" -p"$ADMIN_PASS" -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > "
ProxySQLAdmin >
```

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

`runtime_mysql_servers` shows the Aurora instances ProxySQL discovered. The writer is in hostgroup `2` and the readers are in hostgroup `3`.

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

To test the traffic routing through the ProxySQL server let's first create a pod with ubuntu base image in it. We will use the following yaml.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ubuntu
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ubuntu
  template:
    metadata:
      labels:
        app: ubuntu
    spec:
      containers:
        - image: ubuntu
          imagePullPolicy: IfNotPresent
          name: ubuntu
          command: ["/bin/sleep", "3650d"]
```

Let's apply the yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/aurora/examples/ubuntu.yaml
deployment.apps/ubuntu created
```

Let's exec into the pod and install mysql-client.

```bash
$ kubectl exec -it -n demo ubuntu-bb47d8d6c-7wndq -- bash
root@ubuntu-bb47d8d6c-7wndq:/# apt update
... ... ..
root@ubuntu-bb47d8d6c-7wndq:/# apt install mysql-client -y
Reading package lists... Done
... .. ...
```

Now let's connect with the ProxySQL server through the `aurora-proxy` service as the Aurora master user.

```bash
root@ubuntu-bb47d8d6c-7wndq:/# mysql -uadmin -p'<your-master-password>' -haurora-proxy.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.

mysql> create database proxytest;
Query OK, 1 row affected (0.03 sec)

mysql> create table proxytest.testtb(name varchar(103), primary key(name));
Query OK, 0 rows affected (0.05 sec)

mysql> insert into proxytest.testtb(name) values("Kim Torres");
Query OK, 1 row affected (0.02 sec)

mysql> insert into proxytest.testtb(name) values("Tony SoFua");
Query OK, 1 row affected (0.02 sec)

mysql> select * from proxytest.testtb;
+------------+
| name       |
+------------+
| Kim Torres |
| Tony SoFua |
+------------+
2 rows in set (0.01 sec)
```

We can see the queries are successfully executed through the ProxySQL server.

Let's check the query splits inside the ProxySQL server by going back to the ProxySQLAdmin panel.

```bash
ProxySQLAdmin > select hostgroup,srv_host,Queries from stats_mysql_connection_pool;
+-----------+----------------------------------------------------------------------+---------+
| hostgroup | srv_host                                                             | Queries |
+-----------+----------------------------------------------------------------------+---------+
| 2         | aurora-demo-instance-1.c9akciq32.us-east-1.rds.amazonaws.com         | 4       |
| 3         | aurora-demo-instance-1-reader.c9akciq32.us-east-1.rds.amazonaws.com  | 1       |
| 3         | aurora-demo.cluster-c9akciq32.us-east-1.rds.amazonaws.com            | 0       |
+-----------+----------------------------------------------------------------------+---------+
```

We can see that the write queries went to the writer instance and the read query went to the reader instance. So the ProxySQL server is ready to use.

## Conclusion

In this tutorial, we have seen how to set up KubeDB ProxySQL for an AWS Aurora cluster and how it splits the read and write queries. Checkout the other docs to learn more.
