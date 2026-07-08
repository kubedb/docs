---
title: Initialize ProxySQL with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: proxysql-script-source-initialization
    name: Using Custom Configuration
    parent: proxysql-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize ProxySQL with Custom Configuration

ProxySQL reads a bootstrap configuration file, `proxysql.cnf`, on first start to populate its in-memory configuration tables — `mysql_users`, `mysql_query_rules`, `mysql_variables`, and `admin_variables`. KubeDB lets you supply this bootstrap configuration through `spec.configuration.init`, in one of two ways:

- **`spec.configuration.init.secretName`** — reference a Secret that holds raw `proxysql.cnf` fragments. The values are written into the generated configuration verbatim, so you own the exact ProxySQL syntax and must supply user passwords yourself.
- **`spec.configuration.init.inline`** — describe the same sections as structured YAML. The operator renders them into `proxysql.cnf` for you and, for `mysqlUsers`, automatically retrieves each user's password from the backend server so you never store it in plaintext.

If both are set, `init.inline` takes precedence over `init.secretName`. This guide demonstrates both approaches and then shows how to add users after initialization with a `ProxySQLOpsRequest`.

> Note: `spec.initConfig` and `spec.configSecret` are the deprecated equivalents of `init.inline` and `init.secretName` respectively. Use `spec.configuration.init` for all new ProxySQL objects.

## Before You Begin

You will need a Kubernetes cluster and the `kubectl` command-line tool configured to communicate with it. If you do not already have a cluster, you can create one with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Install the KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the [setup instructions](/docs/setup/README.md).

This guide uses a dedicated namespace called `demo` to keep the resources isolated.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace demo
NAME    STATUS   AGE
demo    Active   5s
```

> Note: The YAML files used in this guide are stored under [docs/guides/proxysql/initialization/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples) in the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Prepare MySQL Backend

ProxySQL sits in front of one or more MySQL servers. Before deploying ProxySQL, create a MySQL backend running in Group Replication mode:

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Wait for the MySQL cluster to become `Ready`:

```bash
$ kubectl get mysql -n demo mysql-server
NAME           VERSION   STATUS   AGE
mysql-server   8.4.8     Ready    5m
```

## Option 1: Bootstrap from a raw configuration Secret

`spec.configuration.init.secretName` points to a Secret that may contain up to four optional keys: `MySQLUsers.cnf`, `MySQLQueryRules.cnf`, `MySQLVariables.cnf`, and `AdminVariables.cnf`. Each key must hold valid **ProxySQL (`proxysql.cnf`) syntax**. During initialization, KubeDB mounts the Secret into the ProxySQL pod and copies each fragment verbatim into the generated `proxysql.cnf` — it does not validate, translate, or merge them. ProxySQL then loads the file at startup and applies it to its in-memory configuration database.

Because the fragments are applied as-is, review the syntax carefully before applying it.

The supported Secret keys are:

| Key                   | ProxySQL table                     |
| --------------------- | ---------------------------------- |
| `MySQLUsers.cnf`      | Frontend users (`mysql_users`)     |
| `MySQLQueryRules.cnf` | Query routing rules (`mysql_query_rules`) |
| `MySQLVariables.cnf`  | MySQL module variables (`mysql_variables`) |
| `AdminVariables.cnf`  | Admin module variables (`admin_variables`) |

The following Secret defines frontend users, query rules, MySQL variables, and admin variables:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: proxysql-init-raw
  namespace: demo
type: Opaque
stringData:
  MySQLUsers.cnf: |
    mysql_users=
    (
      {
        username="wolverine"
        password="wolverine-pass"
        active=1
        default_hostgroup=2
        default_schema="secret_schema"
      },
      {
        username="superman"
        password="superman-pass"
        active=1
        default_hostgroup=3
      }
    )
  MySQLQueryRules.cnf: |
    mysql_query_rules=
    (
      {
        rule_id=100
        active=1
        apply=1
        match_pattern="^INSERT"
        destination_hostgroup=2
      },
      {
        rule_id=101
        active=1
        apply=1
        match_pattern="^SELECT"
        destination_hostgroup=3
      }
    )
  MySQLVariables.cnf: |
    mysql_variables=
    {
      max_connections="4096"
      threads="8"
      default_query_timeout="1234567"
    }
  AdminVariables.cnf: |
    admin_variables=
    {
      refresh_interval="3500"
      restapi_enabled="true"
      restapi_port="6090"
    }
```

Deploy a ProxySQL instance that consumes the Secret:

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: proxy-init-secret
  namespace: demo
spec:
  version: "3.0.1-debian"
  replicas: 1
  backend:
    name: mysql-server
  configuration:
    init:
      secretName: proxysql-init-raw
  deletionPolicy: WipeOut
```

Here:

- `configuration.init.secretName` references the `proxysql-init-raw` Secret that holds the raw configuration fragments.
- Because `MySQLUsers.cnf` supplies each user's `password` explicitly, KubeDB does **not** fetch credentials from the backend — unlike `init.inline`.
- Each fragment is copied directly into the generated `proxysql.cnf` and loaded by ProxySQL at startup. The resulting users, query rules, and variables are then visible through the ProxySQL admin interface.

Apply the Secret and the ProxySQL object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples/proxysql-init-secret.yaml
secret/proxysql-init-raw created
proxysql.kubedb.com/proxy-init-secret created
```

Wait until ProxySQL reaches the `Ready` state:

```bash
$ kubectl get proxysql -n demo proxy-init-secret
NAME                VERSION        STATUS   AGE
proxy-init-secret   3.0.1-debian   Ready    2m
```

### Verify

KubeDB stores the ProxySQL cluster-admin credentials in the `{ProxySQL-name}-auth` Secret (the username defaults to `cluster`, and the password is auto-generated). Read them and connect to the admin interface on port `6032`:

```bash
$ ADMIN_USER=$(kubectl get secret -n demo proxy-init-secret-auth -o jsonpath='{.data.username}' | base64 -d)
$ ADMIN_PASS=$(kubectl get secret -n demo proxy-init-secret-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -it -n demo proxy-init-secret-0 -- mysql -u"$ADMIN_USER" -p"$ADMIN_PASS" -h 127.0.0.1 -P 6032
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 110
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> SELECT variable_name, variable_value
    -> FROM global_variables
    -> WHERE variable_name IN
    ->   ('mysql-max_connections', 'mysql-threads', 'mysql-default_query_timeout');
+-----------------------------+----------------+
| variable_name               | variable_value |
+-----------------------------+----------------+
| mysql-default_query_timeout | 1234567        |
| mysql-max_connections       | 4096           |
| mysql-threads               | 8              |
+-----------------------------+----------------+
3 rows in set (0.00 sec)

MySQL [(none)]> SELECT username, active, default_hostgroup, default_schema
    -> FROM mysql_users;
+-----------+--------+-------------------+----------------+
| username  | active | default_hostgroup | default_schema |
+-----------+--------+-------------------+----------------+
| wolverine | 1      | 2                 | secret_schema  |
| superman  | 1      | 3                 | NULL           |
+-----------+--------+-------------------+----------------+
2 rows in set (0.00 sec)
```

The global variables from `MySQLVariables.cnf` and the frontend users from `MySQLUsers.cnf` are reflected exactly as written in the `proxysql-init-raw` Secret.

## Option 2: Bootstrap from inline structured configuration

`spec.configuration.init.inline` lets you initialize ProxySQL with structured Kubernetes YAML instead of raw `proxysql.cnf` syntax. At initialization, KubeDB translates the inline YAML into the corresponding ProxySQL configuration, merges it with the operator-managed defaults (monitor credentials, cluster authentication, and other required internal settings), and generates the final `proxysql.cnf`. ProxySQL loads this file at startup into its in-memory configuration database.

The inline configuration supports the same four sections:

- **`mysqlUsers`** — frontend users (`mysql_users`). Passwords are fetched automatically from the backend server.
- **`mysqlQueryRules`** — query routing rules (`mysql_query_rules`).
- **`mysqlVariables`** — MySQL module variables (`mysql_variables`).
- **`adminVariables`** — admin module variables (`admin_variables`).

The example below configures query rules, MySQL variables, and admin variables. Users are intentionally omitted here so that the `mysql_users` table starts empty; the [Reconfigure mysql_users](#reconfigure-mysql_users) section that follows adds users to this same instance with a `ProxySQLOpsRequest`, demonstrating how KubeDB retrieves passwords from the backend.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: proxy-init-inline
  namespace: demo
spec:
  version: "3.0.1-debian"
  replicas: 1
  backend:
    name: mysql-server
  configuration:
    init:
      inline:
        mysqlQueryRules:
          - rule_id: 1
            active: 1
            match_pattern: "^SELECT .* FOR UPDATE$"
            destination_hostgroup: 2
            apply: 1
          - rule_id: 2
            active: 1
            match_pattern: "^SELECT"
            destination_hostgroup: 3
            apply: 1
        mysqlVariables:
          max_connections: 2048
          connect_timeout_server: 10000
          threads: 4
          server_version: "8.4.8"
          default_query_timeout: "36000000"
        adminVariables:
          restapi_enabled: "true"
          restapi_port: "6070"
          refresh_interval: "2000"
          cluster_check_interval_ms: "200"
  deletionPolicy: WipeOut
```

Here:

- `configuration.init.inline` provides a Kubernetes-native way to configure ProxySQL without writing raw `proxysql.cnf` syntax.
- `mysqlQueryRules`, `mysqlVariables`, and `adminVariables` are expressed as structured YAML, which is easier to read, diff, and maintain.
- At initialization, KubeDB renders these sections into `proxysql.cnf`, merges them with its own defaults, and lets ProxySQL load the result at startup.

Apply the configuration:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples/proxysql-init-inline.yaml
proxysql.kubedb.com/proxy-init-inline created
```

Wait until ProxySQL reaches the `Ready` state:

```bash
$ kubectl get proxysql -n demo proxy-init-inline
NAME                VERSION        STATUS   AGE
proxy-init-inline   3.0.1-debian   Ready    2m
```

### Verify

Connect to the admin interface and confirm that the query rules and global variables were applied:

```bash
$ ADMIN_USER=$(kubectl get secret -n demo proxy-init-inline-auth -o jsonpath='{.data.username}' | base64 -d)
$ ADMIN_PASS=$(kubectl get secret -n demo proxy-init-inline-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -it -n demo proxy-init-inline-0 -- mysql -u"$ADMIN_USER" -p"$ADMIN_PASS" -h 127.0.0.1 -P 6032
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 61
Server version: 8.4.8 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> SELECT rule_id, active, match_pattern, destination_hostgroup, apply
    -> FROM mysql_query_rules
    -> ORDER BY rule_id;
+---------+--------+------------------------+-----------------------+-------+
| rule_id | active | match_pattern          | destination_hostgroup | apply |
+---------+--------+------------------------+-----------------------+-------+
| 1       | 1      | ^SELECT .* FOR UPDATE$ | 2                     | 1     |
| 2       | 1      | ^SELECT                | 3                     | 1     |
+---------+--------+------------------------+-----------------------+-------+
2 rows in set (0.00 sec)

MySQL [(none)]> SELECT variable_name, variable_value
    -> FROM global_variables
    -> WHERE variable_name IN
    ->   ('mysql-max_connections', 'mysql-connect_timeout_server',
    ->    'mysql-threads', 'mysql-server_version', 'mysql-default_query_timeout');
+------------------------------+----------------+
| variable_name                | variable_value |
+------------------------------+----------------+
| mysql-connect_timeout_server | 10000          |
| mysql-default_query_timeout  | 36000000       |
| mysql-max_connections        | 2048           |
| mysql-server_version         | 8.4.8          |
| mysql-threads                | 4              |
+------------------------------+----------------+
5 rows in set (0.00 sec)
```

The query rules and global variables match the values defined under `mysqlQueryRules`, `mysqlVariables`, and `adminVariables`, confirming that KubeDB rendered them into the generated configuration and ProxySQL loaded them at startup. The `mysql_users` table is empty, since no users were configured inline.

## Reconfigure mysql_users

After initialization, you can manage the `mysql_users` table with a `ProxySQLOpsRequest` — adding, updating, or removing users without re-deploying ProxySQL. As with `init.inline`, you do **not** supply passwords: KubeDB fetches each user's password from the backend MySQL server, so every user you reference must already exist there.

To reconfigure `mysql_users`, set `.spec.type` to `Reconfigure`, point `.spec.proxyRef.name` at your ProxySQL instance, list the users under `.spec.configuration.mysqlUsers.users`, and set `.spec.configuration.mysqlUsers.reqType` to `add`, `update`, or `delete`.

### Create the users in the MySQL backend

First, create the users in the backend MySQL server. The example below creates `wolverine` and `superman`:

```bash
$ kubectl exec -it -n demo mysql-server-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
root@mysql-server-0:/# mysql -uroot -p$MYSQL_ROOT_PASSWORD
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 50594
Server version: 8.4.3 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE USER 'wolverine'@'%' IDENTIFIED BY 'wolverine-pass';
Query OK, 0 rows affected (0.01 sec)

mysql> CREATE USER 'superman'@'%' IDENTIFIED BY 'superman-pass';
Query OK, 0 rows affected (0.01 sec)
```

### Inspect the current mysql_users table

Confirm that the `mysql_users` table is empty on `proxy-init-inline`. This requires that `spec.syncUsers` was **not** enabled on the ProxySQL object — otherwise ProxySQL would automatically import every backend user and the table would not be empty.

```bash
$ kubectl exec -it -n demo proxy-init-inline-0 -- mysql -u"$ADMIN_USER" -p"$ADMIN_PASS" -h 127.0.0.1 -P 6032
...
MySQL [(none)]> SELECT username FROM mysql_users;
Empty set (0.00 sec)
```

### Add users

Submit an ops request to add `wolverine` and `superman`. Because no password is supplied, KubeDB retrieves it from the backend — if a user does not exist there, the request fails.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: add-user
  namespace: demo
spec:
  type: Reconfigure
  proxyRef:
    name: proxy-init-inline
  configuration:
    mysqlUsers:
      users:
        - username: wolverine
          active: 1
          default_hostgroup: 2
        - username: superman
          active: 1
          default_hostgroup: 2
      reqType: add
```

Apply it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples/proxyops-add-users.yaml
proxysqlopsrequest.ops.kubedb.com/add-user created
```

Wait for the ops request to succeed:

```bash
$ kubectl get proxysqlopsrequest -n demo add-user
NAME       TYPE          STATUS       AGE
add-user   Reconfigure   Successful   20s
```

### Verify

Back in the admin interface, the two users now appear with the passwords KubeDB fetched from the backend (stored as hashes), the `active` flag, and the `default_hostgroup` from the ops request:

```bash
MySQL [(none)]> SELECT username, password, active, default_hostgroup
    -> FROM mysql_users;
+-----------+-------------------------------------------+--------+-------------------+
| username  | password                                  | active | default_hostgroup |
+-----------+-------------------------------------------+--------+-------------------+
| superman  | *AE9C3C2838160D2591B6B15FA281CE712ABE94F0 | 1      | 2                 |
| wolverine | *1BB8830D52D091A226FB7990D996CBC20F913475 | 1      | 2                 |
+-----------+-------------------------------------------+--------+-------------------+
2 rows in set (0.00 sec)

```

> To change a user's attributes later, submit another ops request with `reqType: update` (`username` is the key). To remove a user, use `reqType: delete` and list only the username. The [Reconfigure ProxySQL](/docs/guides/proxysql/reconfigure/cluster/index.md) guide covers the full set of operations, including query rules and global variables.

## Cleaning up

To remove the resources created in this guide, run:

```bash
$ kubectl delete -n demo proxysqlopsrequest --all
$ kubectl delete -n demo proxysql proxy-init-secret proxy-init-inline
$ kubectl delete -n demo mysql mysql-server
$ kubectl delete -n demo secret proxysql-init-raw
$ kubectl delete namespace demo
```

## Next Steps

- Learn about [ProxySQL Declarative Configuration](/docs/guides/proxysql/concepts/declarative-configuration/index.md) in detail.
- Learn about [ProxySQL clustering](/docs/guides/proxysql/clustering/overview/index.md) with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
