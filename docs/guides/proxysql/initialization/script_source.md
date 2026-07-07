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

ProxySQL needs a bootstrap configuration file, `proxysql.cnf`, to set up its `mysql_query_rules`, `mysql_variables` and `admin_variables` tables at the very first startup. KubeDB lets you provide this bootstrap configuration in two ways under `spec.configuration.init`:

- **`spec.configuration.init.secretName`** - point to a Secret holding the raw `proxysql.cnf` snippets. The values are patched into the config file verbatim, so you are responsible for the exact ProxySQL config syntax (and for supplying user passwords yourself).
- **`spec.configuration.init.inline`** - describe the same four sections in structured YAML. The operator renders this into `proxysql.cnf` for you, and for `mysqlUsers`, it automatically fetches the password from the backend server instead of asking you to write it in plaintext.

If both are set, `init.inline` always takes precedence over `init.secretName`. This tutorial will show you how to use both.

> Note: `spec.initConfig` and `spec.configSecret` are older, deprecated equivalents of `spec.configuration.init.inline` and `spec.configuration.init.secretName` respectively. Use the `spec.configuration.init` fields for any new ProxySQL object.

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

> Note: YAML files used in this tutorial are stored in [docs/guides/proxysql/initialization/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare MySQL Backend

ProxySQL acts as a proxy in front of MySQL servers. Before deploying ProxySQL, you need a running MySQL Group Replication backend. Apply the following YAML to create one:

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

Wait for the MySQL cluster to be `Ready`:

```bash
$ kubectl get mysql -n demo mysql-server
NAME           VERSION   STATUS   AGE
mysql-server   8.4.8     Ready    5m
```
## Option 1: Bootstrap using a raw configuration Secret

`spec.configuration.init.secretName` refers to a Secret containing up to four optional keys: `MySQLUsers.cnf`, `MySQLQueryRules.cnf`, `MySQLVariables.cnf`, and `AdminVariables.cnf`. Each key must contain valid **ProxySQL configuration (`proxysql.cnf`) syntax**. During initialization, KubeDB mounts the Secret into the ProxySQL pod and copies the provided configuration fragments into the generated `proxysql.cnf` without modifying or translating them. ProxySQL then loads this configuration during startup and applies it to its internal configuration database.

Unlike `init.inline`, which generates the ProxySQL configuration from structured Kubernetes YAML, `init.secretName` treats the Secret contents as raw ProxySQL configuration. Because KubeDB does not validate or merge these configuration fragments, you must ensure that the configuration is syntactically correct before applying it.

The supported Secret keys are:

* **`MySQLUsers.cnf`**: Defines frontend users (`mysql_users`).
* **`MySQLQueryRules.cnf`**: Defines query routing rules (`mysql_query_rules`).
* **`MySQLVariables.cnf`**: Configures global MySQL module variables.
* **`AdminVariables.cnf`**: Configures ProxySQL admin module variables.

The following Secret configures frontend users, query rules, MySQL module variables, and admin variables:

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

Deploy ProxySQL using the Secret:

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

* `configuration.init.secretName` references the `proxysql-init-raw` Secret containing the raw ProxySQL configuration fragments.
* Since `MySQLUsers.cnf` specifies each user's `password` explicitly, KubeDB does **not** automatically retrieve backend credentials, unlike `init.inline`.
* Each configuration fragment is copied directly into the generated `proxysql.cnf` and loaded by ProxySQL during startup. The resulting frontend users, query rules, and global variables are then available through the ProxySQL admin interface.
* Because the Secret is used as-is, ensure that every configuration fragment follows valid ProxySQL configuration syntax before applying it.

Apply the Secret and the ProxySQL object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples/proxysql-init-secret.yaml
secret/proxysql-init-raw created
proxysql.kubedb.com/proxy-init-secret created
```

Wait until ProxySQL goes into the `Ready` state:

```bash
$ kubectl get proxysql -n demo proxy-init-secret
NAME                VERSION        STATUS   AGE
proxy-init-secret   3.0.1-debian   Ready    2m
```

### Verify

Get the admin credentials and connect to the ProxySQL admin interface (port `6032`):

```bash
$ kubectl get secret -n demo proxy-init-inline-auth -o jsonpath='{.data.username}' | base64 -d
cluster⏎                                                                        $ kubectl get secret -n demo proxy-init-inline-auth -o jsonpath='{.data.password}' | base64 -d
0vm8A7bllYpPFTK7⏎             
```

```bash
$ kubectl exec -it -n demo proxy-init-secret-0 -- mysql -u cluster -pGQzOzNmCUBP7pSEv -h 127.0.0.1 -P 6032 
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 110
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> SELECT variable_name, variable_value
    -> FROM global_variables
    -> WHERE variable_name IN (
    ->   'mysql-max_connections',
    ->   'mysql-connect_timeout_server',
    ->   'mysql-threads',
    ->   'mysql-server_version',
    ->   'mysql-default_query_timeout'
    -> );
+------------------------------+----------------+
| variable_name                | variable_value |
+------------------------------+----------------+
| mysql-connect_timeout_server | 10000          |
| mysql-default_query_timeout  | 1234567        |
| mysql-max_connections        | 4096           |
| mysql-server_version         | 8.0.27         |
| mysql-threads                | 8              |
+------------------------------+----------------+
5 rows in set (0.002 sec)

MySQL [(none)]> SELECT username,
    ->        active,
    ->        default_hostgroup,
    ->        default_schema
    -> FROM mysql_users;
+-----------+--------+-------------------+----------------+
| username  | active | default_hostgroup | default_schema |
+-----------+--------+-------------------+----------------+
| wolverine | 1      | 2                 | secret_schema  |
| superman  | 1      | 3                 |                |
+-----------+--------+-------------------+----------------+
2 rows in set (0.000 sec)

```

The `mysql_users`, `mysql_query_rules` and the global variables all reflect exactly what was written in the `proxysql-init-raw` Secret.

## Option 2: Bootstrap using inline structured configuration

`spec.configuration.init.inline` lets you initialize ProxySQL using structured Kubernetes YAML instead of writing raw `proxysql.cnf` configuration. During initialization, KubeDB converts the inline YAML into the corresponding ProxySQL configuration, merges it with the operator-managed defaults (such as monitor user, cluster authentication, and other required internal settings), and generates the final `proxysql.cnf`. ProxySQL then loads the generated configuration into its internal configuration database during startup.

The inline configuration supports the following sections:

* **`mysqlQueryRules`** – Configures query routing rules (`mysql_query_rules`).
* **`mysqlVariables`** – Configures global MySQL module variables.
* **`adminVariables`** – Configures ProxySQL admin module variables.

The following example configures query rules, MySQL variables, and admin variables using inline YAML:

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

* `configuration.init.inline` provides a Kubernetes-native way to configure ProxySQL without writing raw `proxysql.cnf` syntax.
* `mysqlQueryRules`, `mysqlVariables`, and `adminVariables` are defined using structured YAML, making the configuration easier to read and maintain.
* During initialization, KubeDB converts these YAML sections into the corresponding ProxySQL configuration, merges them with the operator-managed defaults, and generates the final `proxysql.cnf`.
* ProxySQL loads the generated configuration during startup, making the configured query rules and global variables available through the ProxySQL admin interface.

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

```bash
$ kubectl get secret -n demo proxy-init-inline-auth -o jsonpath='{.data.username}' | base64 -d
cluster

$ kubectl get secret -n demo proxy-init-inline-auth -o jsonpath='{.data.password}' | base64 -d
yKdzMTY0RKaomsqn
```

```bash
$ kubectl exec -it -n demo proxy-init-inline-0 -- mysql -u cluster -pyKdzMTY0RKaomsqn -h 127.0.0.1 -P 6032 
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 5593
Server version: 8.4.8 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> SELECT variable_name, variable_value
    -> FROM global_variables
    -> WHERE variable_name IN (
    ->   'mysql-max_connections',
    ->   'mysql-connect_timeout_server',
    ->   'mysql-threads',
    ->   'mysql-server_version',
    ->   'mysql-default_query_timeout'
    -> );
+------------------------------+----------------+
| variable_name                | variable_value |
+------------------------------+----------------+
| mysql-connect_timeout_server | 10000          |
| mysql-default_query_timeout  | 36000000       |
| mysql-max_connections        | 2048           |
| mysql-server_version         | 8.4.8          |
| mysql-threads                | 4              |
+------------------------------+----------------+
5 rows in set (0.001 sec)
```
The query results confirm that the values defined under `mysqlQueryRules`, `mysqlVariables`, and `adminVariables` in `configuration.init.inline` were successfully rendered into the generated `ProxySQL` configuration and loaded during `ProxySQL` startup.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo proxysql/proxy-init-secret -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo proxysql/proxy-init-secret

$ kubectl patch -n demo proxysql/proxy-init-inline -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo proxysql/proxy-init-inline

$ kubectl patch -n demo mysql/mysql-server -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo mysql/mysql-server

$ kubectl delete -n demo secret/proxysql-init-raw
$ kubectl delete ns demo
```

## Next Steps

- Learn about [ProxySQL Declarative Configuration](/docs/guides/proxysql/concepts/declarative-configuration/index.md) in detail.
- Learn about [ProxySQL clustering](/docs/guides/proxysql/clustering/overview/index.md) with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
