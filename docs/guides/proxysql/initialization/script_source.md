---
title: Initialize ProxySQL using Script Source
menu:
  docs_{{ .version }}:
    identifier: proxysql-script-source-initialization
    name: Using Script
    parent: proxysql-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize ProxySQL with Script

KubeDB supports ProxySQL initialization using SQL scripts stored in a ConfigMap. This tutorial will show you how to use KubeDB to initialize a ProxySQL instance from a script to pre-configure MySQL backend servers, users, and query routing rules.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/proxysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/proxysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare MySQL Backend

ProxySQL acts as a proxy in front of MySQL servers. Before deploying ProxySQL, you need a running MySQL Group Replication backend. Apply the following YAML to create a MySQL Group Replication:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.4.8"
  replicas: 3
  topology:
    mode: GroupReplication
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/initialization/mysql-server.yaml
mysql.kubedb.com/mysql-server created
```

Wait for the MySQL cluster to be `Ready`:

```bash
$ kubectl get mysql -n demo mysql-server
NAME           VERSION   STATUS   AGE
mysql-server   8.4.8     Ready    5m
```

## Prepare Initialization Scripts

ProxySQL supports initialization with `.sql` and `.sh` scripts. In this tutorial, we will use an `init.sql` script to pre-configure query routing rules that direct write traffic to the primary server and read traffic to replicas.

We will use a ConfigMap as the script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as a script source.

Let's create a ConfigMap with the initialization script:

```bash
$ kubectl create configmap -n demo proxysql-init-script \
--from-literal=init.sql="$(curl -fsSL https://raw.githubusercontent.com/kubedb/proxysql-init-scripts/master/init.sql)"
configmap/proxysql-init-script created
```

## Create ProxySQL with Script Source

Following YAML describes the ProxySQL object with `init.script`:

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: script-proxysql
  namespace: demo
spec:
  version: "3.0.1-debian"
  replicas: 1
  backend:
    name: mysql-server
  init:
    script:
      configMap:
        name: proxysql-init-script
  deletionPolicy: WipeOut
```

Here,

- `init.script` specifies the SQL scripts used to initialize ProxySQL when it is being created. Scripts are executed against the ProxySQL admin interface.
- `backend.name` references the MySQL Group Replication that ProxySQL will front.

VolumeSource provided in `init.script` will be mounted in the Pod and executed against ProxySQL's admin port (`6032`) during initialization.

Now, let's create the ProxySQL CRD using the YAML shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/initialization/script-proxysql.yaml
proxysql.kubedb.com/script-proxysql created
```

Now, wait until ProxySQL goes in `Ready` state. Verify that it is in `Ready` state using the following command:

```bash
$ kubectl get proxysql -n demo script-proxysql
NAME              VERSION        STATUS   AGE
script-proxysql   3.0.1-debian   Ready    2m
```

## Verify Initialization

Now let's connect to our ProxySQL instance to verify that the initialization scripts have been applied successfully.

**Connection Information:**

- Host name/address: you can use any of these
  - Service: `script-proxysql.demo`
  - Pod IP: (`$ kubectl get pods script-proxysql-0 -n demo -o yaml | grep podIP`)
- Port: `6033` (MySQL traffic proxy) or `6032` (ProxySQL admin)

- Username: Run the following command to get the *username*:

  ```bash
  $ kubectl get secret -n demo script-proxysql-auth -o jsonpath='{.data.username}' | base64 -d
  proxysql
  ```

- Password: Run the following command to get the *password*:

  ```bash
  $ kubectl get secret -n demo script-proxysql-auth -o jsonpath='{.data.password}' | base64 -d
  S3cur3P@ssw0rd
  ```

Connect to the ProxySQL admin interface and verify the initialized query rules:

```bash
$ kubectl exec -it -n demo script-proxysql-0 -- mysql -u proxysql -pS3cur3P@ssw0rd -h 127.0.0.1 -P 6032 \
  -e "SELECT rule_id, match_pattern, destination_hostgroup FROM mysql_query_rules;"
+---------+-------------------+----------------------+
| rule_id | match_pattern     | destination_hostgroup |
+---------+-------------------+----------------------+
|       1 | ^SELECT.*FOR UPDATE$ |                  2 |
|       2 | ^SELECT           |                  3 |
+---------+-------------------+----------------------+
```

Verify the configured MySQL servers in ProxySQL:

```bash
$ kubectl exec -it -n demo script-proxysql-0 -- mysql -u proxysql -pS3cur3P@ssw0rd -h 127.0.0.1 -P 6032 \
  -e "SELECT hostgroup_id, hostname, port, status FROM mysql_servers;"
+--------------+---------------------------------------------+------+--------+
| hostgroup_id | hostname                                    | port | status |
+--------------+---------------------------------------------+------+--------+
|            2 | mysql-server-0.mysql-server-pods.demo.svc   | 3306 | ONLINE |
|            3 | mysql-server-1.mysql-server-pods.demo.svc   | 3306 | ONLINE |
|            3 | mysql-server-2.mysql-server-pods.demo.svc   | 3306 | ONLINE |
+--------------+---------------------------------------------+------+--------+
```

We can see that the query routing rules and MySQL server configurations were applied through the initialization script.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo proxysql/script-proxysql -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo proxysql/script-proxysql

$ kubectl patch -n demo mysql/mysql-server -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo mysql/mysql-server

$ kubectl delete -n demo configmap/proxysql-init-script
$ kubectl delete ns demo
```

## Next Steps

- Learn about [ProxySQL clustering](/docs/guides/proxysql/clustering/overview/index.md) with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
