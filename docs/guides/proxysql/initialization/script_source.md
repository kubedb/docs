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

ProxySQL requires a bootstrap configuration file, `proxysql.cnf`, to populate its `mysql_users`, `mysql_query_rules`, `mysql_variables`, and `admin_variables` tables during its very first startup. KubeDB lets you provide this bootstrap configuration in two ways under `spec.configuration.init`:

- **`spec.configuration.init.secretName`** - Points to a Secret holding the raw `proxysql.cnf` snippets. The values are patched into the config file verbatim, so you are responsible for the exact ProxySQL config syntax (and for supplying user passwords yourself).
- **`spec.configuration.init.inline`** - Describes the same four sections in structured YAML. The operator renders this into `proxysql.cnf` for you, and for `mysqlUsers`, it automatically fetches the password from the backend server instead of requiring you to write it in plaintext.

If both are set, `init.inline` always takes precedence over `init.secretName`. This tutorial demonstrates both approaches.

> Note: `spec.initConfig` and `spec.configSecret` are older, deprecated equivalents of `spec.configuration.init.inline` and `spec.configuration.init.secretName` respectively. Use the `spec.configuration.init` fields for any new ProxySQL object.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout.

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

`spec.configuration.init.secretName` references a Secret with up to four keys: `MySQLUsers.cnf`, `MySQLQueryRules.cnf`, `MySQLVariables.cnf`, and `AdminVariables.cnf`. Each key's value must already be in valid `proxysql.cnf` syntax, since the operator copies it into the config file exactly as-is.

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
---
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

Here,

- `configuration.init.secretName` points to the `proxysql-init-raw` Secret above. Since `MySQLUsers.cnf` sets each user's `password` explicitly, this path does **not** auto-fetch credentials from the backend the way `init.inline` does.
- Unlike `init.inline`, values under `init.secretName` are not merged with anything else - only what you put in the Secret (plus KubeDB's internal defaults for things like cluster auth and TLS) ends up in `proxysql.cnf`. Double-check the syntax carefully before applying.

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
$ kubectl get secret -n demo proxy-init-secret-auth -o jsonpath='{.data.username}' | base64 -d
cluster

$ kubectl get secret -n demo proxy-init-secret-auth -o jsonpath='{.data.password}' | base64 -d
S3cur3P@ssw0rd
```

```bash
$ kubectl exec -it -n demo proxy-init-secret-0 -- mysql -u cluster -pS3cur3P@ssw0rd -h 127.0.0.1 -P 6032 \
  -e "SELECT username, active, default_hostgroup, default_schema FROM mysql_users;"
+-----------+--------+-------------------+------------------+
| username  | active | default_hostgroup | default_schema   |
+-----------+--------+-------------------+------------------+
| wolverine |      1 |                 2 | secret_schema    |
| superman  |      1 |                 3 |                  |
+-----------+--------+-------------------+------------------+

$ kubectl exec -it -n demo proxy-init-secret-0 -- mysql -u cluster -pS3cur3P@ssw0rd -h 127.0.0.1 -P 6032 \
  -e "SELECT rule_id, match_pattern, destination_hostgroup FROM mysql_query_rules;"
+---------+---------------+------------------------+
| rule_id | match_pattern | destination_hostgroup |
+---------+---------------+------------------------+
|     100 | ^INSERT       |                      2 |
|     101 | ^SELECT       |                      3 |
+---------+---------------+------------------------+

$ kubectl exec -it -n demo proxy-init-secret-0 -- mysql -u cluster -pS3cur3P@ssw0rd -h 127.0.0.1 -P 6032 \
  -e "SELECT variable_name, variable_value FROM global_variables WHERE variable_name IN ('mysql-max_connections','mysql-threads','mysql-default_query_timeout','admin-restapi_enabled','admin-restapi_port','admin-refresh_interval');"
+------------------------------+----------------+
| variable_name                | variable_value |
+------------------------------+----------------+
| mysql-max_connections        | 4096           |
| mysql-threads                | 8              |
| mysql-default_query_timeout  | 1234567        |
| admin-restapi_enabled        | true           |
| admin-restapi_port           | 6090           |
| admin-refresh_interval       | 3500           |
+------------------------------+----------------+
```

The `mysql_users`, `mysql_query_rules`, and the global variables all reflect exactly what was written in the `proxysql-init-raw` Secret.

## Option 2: Bootstrap using inline structured configuration

`spec.configuration.init.inline` describes the same four sections (`mysqlUsers`, `mysqlQueryRules`, `mysqlVariables`, `adminVariables`) in structured YAML instead of raw config syntax. The operator renders these into `proxysql.cnf`, merging them with KubeDB's own defaults (such as monitor and cluster-auth variables).

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

Here,

- `configuration.init.inline.mysqlUsers` does not take a `password` field - KubeDB fetches each user's password from the backend MySQL server automatically, so no credential is ever written in plaintext YAML.
- `configuration.init.inline.mysqlQueryRules`, `.mysqlVariables`, and `.adminVariables` accept the same keys as their raw `proxysql.cnf` counterparts, expressed in YAML key-value form.

See the [Declarative Configuration](/docs/guides/proxysql/concepts/declarative-configuration/index.md) concept page for the full field-by-field reference.

Apply the YAML:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/initialization/examples/proxysql-init-inline.yaml
proxysql.kubedb.com/proxy-init-inline created
```

Wait until ProxySQL goes into the `Ready` state:

```bash
$ kubectl get proxysql -n demo proxy-init-inline
NAME                VERSION        STATUS   AGE
proxy-init-inline   3.0.1-debian   Ready    2m
```

### Verify

#### Create Users in the MySQL Database

Before ProxySQL can fetch credentials from the backend, the corresponding users must exist there. Create two users on the backend MySQL server:

```bash
$ kubectl exec -it -n demo mysql-server-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-5.1$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 512
Server version: 8.4.3 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE USER `wolverine` IDENTIFIED BY 'wolverine-pass';
Query OK, 0 rows affected (0.02 sec)

mysql> CREATE USER `superman` IDENTIFIED BY 'superman-pass';
Query OK, 0 rows affected (0.02 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.00 sec)

mysql> exit
Bye
```

#### Add Users via Reconfigure Ops-request

You can also add users to a running ProxySQL server dynamically, using a `ProxySQLOpsRequest`. The following example adds users `testA` and `testB` to the `proxy-server` ProxySQL instance. Since no `password` field is provided in the YAML, the KubeDB operator fetches each user's password from the backend MySQL server automatically - so the corresponding users must already exist there before applying this OpsRequest. If a user is not present on the backend, the operator will be unable to fetch its password and the OpsRequest will fail.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
    name: add-user
    namespace: demo
spec:
    type: Reconfigure  
    proxyRef:
      name: proxy-server
    configuration:
      mysqlUsers:
        users: 
        - username: testA
          active: 1
          default_hostgroup: 2  
        - username: testB
          active: 1
          default_hostgroup: 2
        reqType: add
```

Apply the OpsRequest YAML:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-add-users.yaml
proxysqlopsrequest.ops.kubedb.com/add-user created
```

Wait for the OpsRequest to reach the `Successful` state:

```bash
$ kubectl get proxysqlopsrequest -n demo     
NAME       TYPE          STATUS       AGE
add-user   Reconfigure   Successful   20s
```

Now verify the `mysql_users` table on the ProxySQL server:


```bash
$ kubectl exec -it -n demo proxy-init-inline-0 -- bash
proxysql@proxy-init-inline-0:/$ mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin > "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 90
Server version: 8.4.8 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select * from mysql_users;
+-----------+------------------------------------------------------------------------------------------------+--------+---------+-------------------+----------------+---------------+------------------------+--------------+---------+----------+-----------------+------------+---------+
| username  | password                                                                                       | active | use_ssl | default_hostgroup | default_schema | schema_locked | transaction_persistent | fast_forward | backend | frontend | max_connections | attributes | comment |
+-----------+------------------------------------------------------------------------------------------------+--------+---------+-------------------+----------------+---------------+------------------------+--------------+---------+----------+-----------------+------------+---------+
| wolverine | $A$005$7GF&7o.%\x10;.%jP5Ui??\\KmKM45B50Xt3wzJ2fmWo9FHAi2Yt7mhhCa8wlZpHZ/1                     | 1      | 0       | 2                 | NULL           | 0             | 1                      | 0            | 1       | 1        | 10000           |            |         |
| superman  | $A$005$\t!MR\x18p{\x12\x06TE\x1e\t\\,\x1bI8\x15\x1bcxpfFLuOBmbfiLF9rs8MTeYQ3yJVQghmaM4/evor/XC | 1      | 0       | 2                 | NULL           | 0             | 1                      | 0            | 1       | 1        | 10000           |            |         |
+-----------+------------------------------------------------------------------------------------------------+--------+---------+-------------------+----------------+---------------+------------------------+--------------+---------+----------+-----------------+------------+---------+
2 rows in set (0.001 sec)

ProxySQLAdmin > SELECT rule_id, match_pattern, destination_hostgroup FROM mysql_query_rules;
+---------+------------------------+-----------------------+
| rule_id | match_pattern          | destination_hostgroup |
+---------+------------------------+-----------------------+
| 1       | ^SELECT .* FOR UPDATE$ | 2                     |
| 2       | ^SELECT                | 3                     |
+---------+------------------------+-----------------------+
2 rows in set (0.002 sec)

ProxySQLAdmin > SELECT variable_name, variable_value FROM global_variables WHERE variable_name IN ('mysql-max_connections','mysql-threads','admin-restapi_enabled','admin-restapi_port');
+-----------------------+----------------+
| variable_name         | variable_value |
+-----------------------+----------------+
| admin-restapi_enabled | true           |
| admin-restapi_port    | 6070           |
| mysql-max_connections | 2048           |
| mysql-threads         | 4              |
+-----------------------+----------------+
4 rows in set (0.002 sec)

```

Since `wolverine` and `superman` also exist on the MySQL backend, ProxySQL was able to log in and fetch their passwords automatically - confirming that a client can connect through ProxySQL using those credentials without a password ever being written into the YAML.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run the following commands:

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
