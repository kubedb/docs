---
title: About ProxySQL
menu:
  docs_{{ .version }}:
    identifier: about-proxysql
    name: About ProxySQL
    parent: proxysql-overview
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Introduction

High availability and better performance are crucial for any database. To keep pace with the growing traffic/connection/data size, often we add multiple database servers and use replication among themselves. For smooth service, we often want to route the read and write intelligently.

Say we may have some secondary (slave) servers along with primary (master) servers and we want to split the read traffic to the slaves and the write traffic to the masters. On top of that, what will happen if the replication delays or some server crashes?

[ProxySQL](http://www.proxysql.com/) comes in hand in such a situation.

## What is ProxySQL

ProxySQL is a high performance, high availability, protocol aware proxy for MySQL and forks (like Percona Server and MariaDB).

[ProxySQL](http://www.proxysql.com/) is an open-source MySQL proxy server. It can improve performance by distributing traffic among multiple database servers. It also improves the availability by automatically failing over if servers fail.

ProxySQL is an intelligent router for Galera and Group Replication. It can defer reads, writes and route the write traffic to the primary and read traffic to the secondary. ProxySQL routes the traffic to backend MySQL, Percona Server, MariaDB (in fact MySQL and MySQL forks).

<p align="center">
    <img alt="proxysql-query-filtering"  src="/docs/images/proxysql/proxysql.svg">
</p>

In the event of any failure, ProxySQL can identify and forward traffic to an available one. It continuously monitors the failed node and includes into the group when the failed node comes back.

## Architecture

Here is the architecture of ProxySQL. It treats as a gateway for the traffic coming from applications/clients. Clients connect to the ProxySQL instead of the backend database and send requests. Then the requests are evaluated by ProxySQL and corresponding actions are performed. To evaluate the client requests and perform an action, there are rules defined in ProxySQL.

<p align="center">
    <img alt="proxysql-query-filtering"  src="/docs/images/proxysql/proxysql-architecture.png">
</p>

## Features

ProxySQL has several features. Below are some of the notable of them.

- Complex query routing and read/write split
- Load balancing
- Real-time statistics
- Monitoring
- High availability and scalability
- Seamless failover
- Runtime reconfiguration
- Scheduler
- Support for Galera and Group Replication
- Support for millions of users
- Support for tens of thousands of database servers
- Native ProxySQL Clustering solution

## Inside of ProxySQL

The details about ProxySQL is huge and can be [here](https://github.com/sysown/proxysql/wiki/). Here we will discuss some of them.

### Admin Interface

After starting ProxySQL, it uses a package-provided configuration file (default is `/etc/proxysql.cnf`) to initialize default values for all of its configuration variables. After this initialization, ProxySQL stores its configuration in a database that you can manage and modify via the admin interface from the command line.

First, access the administration interface. You’ll be prompted for a password which, on a default installation, is `admin`.

```console
$ mysql --user=admin --password=admin --host=127.0.0.1 --port=6032 --prompt='ProxySQLAdmin> '
```

- `--user` specifies the user we are connecting to, which here is **`admin`** as the default user for administrative tasks.
- `--password` specifies the password for the user we are connecting to.
- `--host` tells **`mysql`** to connect to the local ProxySQL instance. We need to define this explicitly because ProxySQL doesn’t listen on the socket file that **`mysql`** assumes by default.
- `--port` specifies the admin port to connect to and it is `6032`.
- `--prompt` is just an optional flag. Normally the default prompt for `mysql` is `mysql> `, but here we are changing it to `ProxySQLAdmin> `. So, when you connect to the admin interface, you will see like `ProxySQLAdmin> ` in the prompt.

From the admin interface, we can see that there are a few databases available.

```console
ProxySQLAdmin> show databases;
+-----+---------------+-------------------------------------+
| seq | name          | file                                |
+-----+---------------+-------------------------------------+
| 0   | main          |                                     |
| 2   | disk          | /var/lib/proxysql/proxysql.db       |
| 3   | stats         |                                     |
| 4   | monitor       |                                     |
| 5   | stats_history | /var/lib/proxysql/proxysql_stats.db |
+-----+---------------+-------------------------------------+
5 rows in set (0.00 sec)
```

- `main`: the in-memory configuration database.
- `disk`: the disk-based mirror of "main".
- `stats`: contains runtime metrics collected from the internal functioning.
- `monitor`: contains monitoring metrics related to the backend servers to which ProxySQL connects.

#### Main - Runtime

The tables from the `main` database are as follows:

```sql
ProxySQLAdmin> SHOW TABLES FROM main;
+--------------------------------------------+
| tables                                     |
+--------------------------------------------+
| global_variables                           |
| mysql_collations                           |
| mysql_galera_hostgroups                    |
| mysql_group_replication_hostgroups         |
| mysql_query_rules                          |
| mysql_query_rules_fast_routing             |
| mysql_replication_hostgroups               |
| mysql_servers                              |
| mysql_users                                |
| proxysql_servers                           |
| runtime_checksums_values                   |
| runtime_global_variables                   |
| runtime_mysql_galera_hostgroups            |
| runtime_mysql_group_replication_hostgroups |
| runtime_mysql_query_rules                  |
| runtime_mysql_query_rules_fast_routing     |
| runtime_mysql_replication_hostgroups       |
| runtime_mysql_servers                      |
| runtime_mysql_users                        |
| runtime_proxysql_servers                   |
| runtime_scheduler                          |
| scheduler                                  |
+--------------------------------------------+
22 rows in set (0.00 sec)
```

For Group Replication, the following tables are the concern here: `mysql_group_replication_hostgroups`, `mysql_galera_hostgroups`, `mysql_servers`, `mysql_users`, `mysql_query_rules` and `global_variables`.

##### mysql_group_replication_hostgroups

Hostgroups for Group Replication are defined in table `mysql_group_replication_hostgroups`.

```sql
ProxySQLAdmin> show create table mysql_group_replication_hostgroups\G
*************************** 1. row ***************************
       table: mysql_group_replication_hostgroups
Create Table: CREATE TABLE mysql_group_replication_hostgroups (
    writer_hostgroup INT CHECK (writer_hostgroup>=0) NOT NULL PRIMARY KEY,
    backup_writer_hostgroup INT CHECK (backup_writer_hostgroup>=0 AND backup_writer_hostgroup<>writer_hostgroup) NOT NULL,
    reader_hostgroup INT NOT NULL CHECK (reader_hostgroup<>writer_hostgroup AND backup_writer_hostgroup<>reader_hostgroup AND reader_hostgroup>0),
    offline_hostgroup INT NOT NULL CHECK (offline_hostgroup<>writer_hostgroup AND offline_hostgroup<>reader_hostgroup AND backup_writer_hostgroup<>offline_hostgroup AND offline_hostgroup>=0),
    active INT CHECK (active IN (0,1)) NOT NULL DEFAULT 1,
    max_writers INT NOT NULL CHECK (max_writers >= 0) DEFAULT 1,
    writer_is_also_reader INT CHECK (writer_is_also_reader IN (0,1)) NOT NULL DEFAULT 0,
    max_transactions_behind INT CHECK (max_transactions_behind>=0) NOT NULL DEFAULT 0,
    comment VARCHAR,
    UNIQUE (reader_hostgroup),
    UNIQUE (offline_hostgroup),
    UNIQUE (backup_writer_hostgroup))
1 row in set (0.00 sec)
```

- **`writer_hostgroup`** - by default all the traffic are sent to this group, nodes having `read_only=0` are in this host group
- **`backup_writer_hostgroup`** - if the number of cluster nodes with `read_only=0` is greater than the value of `max_writers`, ProxySQL will put the additional nodes in this group
- **`reader_hostgroup`** - it is the host group to which read traffic be sent, nodes having `read_only=1` are assigned to this host group
- **`offline_hostgroup`** - when ProxySQL's monitoring determines a node is `OFFLINE`, it will be put into the offline_hostgroup
- **`active`** - when enabled, ProxySQL monitors the host groups and moves nodes to the appropriate host groups
- **`max_writers`** -the maximum number of nodes that should be in the `writer_hostgroup`, extra nodes will be put into the `backup_writer_hostgroup`
- **`writer_is_also_reader`** - tells ProxySQL that if a node should be added to the `reader_hostgroup` as well as the `writer_hostgroup` after being promoted
- **`max_transactions_behind`** - determines the maximum number of transactions behind the writers that ProxySQL should allow before shunning the node to prevent stale reads (this is determined by querying the `transactions_behind` field of the `sys.gr_member_routing_candidate_status` table in MySQL)
- **`comment`** - text field that can be used for any purposed defined by the user. Could be a description of what the cluster stores, a reminder of when the host group was added or disabled, or a JSON processed by some checker script

##### mysql_galera_hostgroups

This table is available from ProxySQL 2.x. It defines the host groups for MySQL servers using Galera Cluster such as Percona XtraDB Cluster, MariaDB Galera Cluster.

```sql
ProxySQLAdmin> show create table mysql_galera_hostgroups\G
*************************** 1. row ***************************
       table: mysql_galera_hostgroups
Create Table: CREATE TABLE mysql_galera_hostgroups (
    writer_hostgroup INT CHECK (writer_hostgroup>=0) NOT NULL PRIMARY KEY,
    backup_writer_hostgroup INT CHECK (backup_writer_hostgroup>=0 AND backup_writer_hostgroup<>writer_hostgroup) NOT NULL,
    reader_hostgroup INT NOT NULL CHECK (reader_hostgroup<>writer_hostgroup AND backup_writer_hostgroup<>reader_hostgroup AND reader_hostgroup>0),
    offline_hostgroup INT NOT NULL CHECK (offline_hostgroup<>writer_hostgroup AND offline_hostgroup<>reader_hostgroup AND backup_writer_hostgroup<>offline_hostgroup AND offline_hostgroup>=0),
    active INT CHECK (active IN (0,1)) NOT NULL DEFAULT 1,
    max_writers INT NOT NULL CHECK (max_writers >= 0) DEFAULT 1,
    writer_is_also_reader INT CHECK (writer_is_also_reader IN (0,1)) NOT NULL DEFAULT 0,
    max_transactions_behind INT CHECK (max_transactions_behind>=0) NOT NULL DEFAULT 0,
    comment VARCHAR,
    UNIQUE (reader_hostgroup),
    UNIQUE (offline_hostgroup),
    UNIQUE (backup_writer_hostgroup))
1 row in set (0.00 sec)
```

- **`writer_hostgroup`** - by default all the traffic are sent to this group, nodes having `read_only=0` are in this host group
- **`backup_writer_hostgroup`** - if the number of cluster nodes with `read_only=0` is greater than the value of `max_writers`, ProxySQL will put the additional nodes in this group
- **`reader_hostgroup`** - it is the host group to which read traffic be sent, nodes having `read_only=1` are assigned to this host group
- **`offline_hostgroup`** - when ProxySQL's monitoring determines a host is `OFFLINE`, it will be put into the offline_hostgroup
- **`active`** - when enabled, ProxySQL monitors the host groups and moves nodes to the appropriate host groups
- **`max_writers`** - the maximum number of nodes that should be in the `writer_hostgroup`, extra nodes will be put into the `backup_writer_hostgroup`
- **`writer_is_also_reader`** - tells ProxySQL that if a node should be added to the `reader_hostgroup` as well as the `writer_hostgroup` after being promoted
- **`max_transactions_behind`** - determines the maximum number of write sets behind the cluster that ProxySQL should allow before shunning the node to prevent stale reads (this is determined by querying the `wsrep_local_recv_queue` Galera variable)
- **`comment`** - text field that can be used for any purposed defined by the user. Could be a description of what the cluster stores, a reminder of when the host group was added or disabled, or a JSON processed by some checker script

##### mysql_servers

Table **`mysql_servers`** contains all the backend MySQL servers' information. Its schema definition is as follows:

```sql
ProxySQLAdmin> SHOW CREATE TABLE mysql_servers\G
*************************** 1. row ***************************
       table: mysql_servers
Create Table: CREATE TABLE mysql_servers (
    hostgroup_id INT CHECK (hostgroup_id>=0) NOT NULL DEFAULT 0,
    hostname VARCHAR NOT NULL,
    port INT NOT NULL DEFAULT 3306,
    gtid_port INT CHECK (gtid_port <> port) NOT NULL DEFAULT 0,
    status VARCHAR CHECK (UPPER(status) IN ('ONLINE','SHUNNED','OFFLINE_SOFT', 'OFFLINE_HARD')) NOT NULL DEFAULT 'ONLINE',
    weight INT CHECK (weight >= 0 AND weight <=10000000) NOT NULL DEFAULT 1,
    compression INT CHECK (compression >=0 AND compression <= 102400) NOT NULL DEFAULT 0,
    max_connections INT CHECK (max_connections >=0) NOT NULL DEFAULT 1000,
    max_replication_lag INT CHECK (max_replication_lag >= 0 AND max_replication_lag <= 126144000) NOT NULL DEFAULT 0,
    use_ssl INT CHECK (use_ssl IN(0,1)) NOT NULL DEFAULT 0,
    max_latency_ms INT UNSIGNED CHECK (max_latency_ms>=0) NOT NULL DEFAULT 0,
    comment VARCHAR NOT NULL DEFAULT '',
    PRIMARY KEY (hostgroup_id, hostname, port) )
1 row in set (0.00 sec)
```

- **`hostgroup_id`**: the host group in which this server resides. The same server can be part of more than one host group
- **`hostname`**, **`port`**: the TCP endpoint of the server
- **`gtid_port`**: the port at which ProxySQL Binlog Reader listens on for GTID tracking
- **`status`**:
 - `ONLINE` - server is fully operational
 - `SHUNNED` - either too many connection errors in time or replication lag exceeded the allowed threshold and therefore the server is temporarily out
 - `OFFLINE_SOFT` - in this mode, new connections aren't accepted anymore, until the existing connections became inactive. That means connections are kept in use until the current transaction is completed. This allows to gracefully detach a backend server.
 - `OFFLINE_HARD` - in this mode, the existing connections are dropped, while new connections aren't accepted either. It means deleting the server from a host group, or taking it out of the host group temporarily
- **`weight`** - the probability of choosing a server from a host group with a larger weight value is high.
- **`compression`** - value greater than 0, new connections to that server will use compression
- **`max_connections`** - the maximum number of connections ProxySQL will open to this server
- **`max_replication_lag`** - value must be in range, 0 <= `max_replication_lag` <= 126144000. In that range, ProxySQL will regularly monitor replication lag and if it goes beyond such threshold it will temporary shun the host until replication catch-ups
- **`use_ssl`** - if set to 1, connections use SSL.
- **`max_latency_ms`** - if a host has a greater ping time than `max_latency_ms`, it is excluded from the connection pool (although the server stays `ONLINE`)
- **`comment`** - user-defined text field

##### mysql_users

This table describes the users of MySQL that are used to connect to the backend server.

```sql
ProxySQLAdmin> SHOW CREATE TABLE mysql_users\G
*************************** 1. row ***************************
       table: mysql_users
Create Table: CREATE TABLE mysql_users (
    username VARCHAR NOT NULL,
    password VARCHAR,
    active INT CHECK (active IN (0,1)) NOT NULL DEFAULT 1,
    use_ssl INT CHECK (use_ssl IN (0,1)) NOT NULL DEFAULT 0,
    default_hostgroup INT NOT NULL DEFAULT 0,
    default_schema VARCHAR,
    schema_locked INT CHECK (schema_locked IN (0,1)) NOT NULL DEFAULT 0,
    transaction_persistent INT CHECK (transaction_persistent IN (0,1)) NOT NULL DEFAULT 0,
    fast_forward INT CHECK (fast_forward IN (0,1)) NOT NULL DEFAULT 0,
    backend INT CHECK (backend IN (0,1)) NOT NULL DEFAULT 1,
    frontend INT CHECK (frontend IN (0,1)) NOT NULL DEFAULT 1,
    max_connections INT CHECK (max_connections >=0) NOT NULL DEFAULT 10000,
    comment VARCHAR NOT NULL DEFAULT '',
    PRIMARY KEY (username, backend),
    UNIQUE (username, frontend))
1 row in set (0.00 sec)
```

- **`username`**, **`password`** - credentials for connecting to the MySQL or ProxySQL instance
- **`active`** - if active = 0, the user will be tracked in the database, but will be never loaded in the in-memory data structures
- **`default_hostgroup`** - if no matching rule is found for the queries sent by this user, the traffic is sent to the default host group
- **`default_schema`** - the schema to which the connection should change by default
- **`schema_locked`** - not supported yet (TODO: check)
- **`transaction_persistent`** - if set for this user, transactions started within a host group will remain within that host group regardless of any other rules
- **`fast_forward`** - if set, it bypasses the query processing layer (rewriting, caching) and passes through the query directly as is to the backend server
- **`frontend`** - if set, this (username, password) pair is used to authenticate to ProxySQL
- **`backend`** - if set, this (username, password) pair is used to authenticate to the MySQL servers against any host group
- **`max_connections`** - defines the maximum number of allowable connections for a user.
- **`comment`** - user-defined text field to describe something

##### mysql_query_rules

This table `mysql_query_rules` defines routing rules for the incoming traffic.

```sql
Admin> SHOW CREATE TABLE mysql_query_rules\G
*************************** 1. row ***************************
       table: mysql_query_rules
Create Table: CREATE TABLE mysql_query_rules (
    rule_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    active INT CHECK (active IN (0,1)) NOT NULL DEFAULT 0,
    username VARCHAR,
    schemaname VARCHAR,
    flagIN INT NOT NULL DEFAULT 0,
    client_addr VARCHAR,
    proxy_addr VARCHAR,
    proxy_port INT,
    digest VARCHAR,
    match_digest VARCHAR,
    match_pattern VARCHAR,
    negate_match_pattern INT CHECK (negate_match_pattern IN (0,1)) NOT NULL DEFAULT 0,
    re_modifiers VARCHAR DEFAULT 'CASELESS',
    flagOUT INT,
    replace_pattern VARCHAR,
    destination_hostgroup INT DEFAULT NULL,
    cache_ttl INT CHECK(cache_ttl > 0),
    cache_empty_result INT CHECK (cache_empty_result IN (0,1)) DEFAULT NULL,
    reconnect INT CHECK (reconnect IN (0,1)) DEFAULT NULL,
    timeout INT UNSIGNED,
    retries INT CHECK (retries>=0 AND retries <=1000),
    delay INT UNSIGNED,
    next_query_flagIN INT UNSIGNED,
    mirror_flagOUT INT UNSIGNED,
    mirror_hostgroup INT UNSIGNED,
    error_msg VARCHAR,
    OK_msg VARCHAR,
    sticky_conn INT CHECK (sticky_conn IN (0,1)),
    multiplex INT CHECK (multiplex IN (0,1)),
    gtid_from_hostgroup INT UNSIGNED,
    log INT CHECK (log IN (0,1)),
    apply INT CHECK(apply IN (0,1)) NOT NULL DEFAULT 0,
    comment VARCHAR)
1 row in set (0.00 sec)
```

Some these field, important to set a query rule, are shortly described below:

- **`rule_id`** - the unique id of the rule
- **`active`** - only active rules are applied
- **`username`** - If it is not empty, a query will match only if the connection is from the correct username
- **`client_addr`** - match traffic from a specific source
- **`match_digest`** - regular expression that matches the query digest
- **`match_pattern`** - regular expression that matches the query text
- **`destination_hostgroup`** - route matched queries to this host group
- **`timeout`** - the maximum timeout in milliseconds within which the matched or rewritten query should be executed. If `timeout` is not specified, global variable `mysql-default_query_timeout` applies
- **`retries`** - the maximum number of times a query needs to be re-executed in case of detected failure during the execution of the query. If not specified, global variable `mysql-query_retries_on_failure` applies
- **`log`** - If set to 1, the query will be logged
- **`apply`** - if set to 1, no further queries will be evaluated after this rule is matched and processed
- **`comment`** - a descriptive comment of the query rule

##### global_variables

The table `global_variables` is two columns table like a key-value store. It defines global variables used by ProxySQL.

There are 2 classes of global variables currently:

- **`admin`** - these are prefixed with `admin-` and relevant for admin module allow tweaking the admin interface
- **`mysql`** - these are prefixed with `mysql-` and relevant for MySQL modules allow tweaking of MySQL-related features. These include configuring variables to handle MySQL traffic, monitor operations (further prefixed with `mysql-monitor_`), query caching.

For more information, please see section [global variables](https://github.com/sysown/proxysql/wiki/global_variables.md).

```sql
ProxySQLAdmin> SHOW CREATE TABLE global_variables\G
*************************** 1. row ***************************
       table: global_variables
Create Table: CREATE TABLE global_variables (
    variable_name VARCHAR NOT NULL PRIMARY KEY,
    variable_value VARCHAR NOT NULL)
1 row in set (0.00 sec)
```

You can see all the variables, by running the following query,

```sql
ProxySQLAdmin> SELECT * FROM global_variables ORDER BY variable_name;
```

## Next Steps

- Configure ProxySQL for Group Replication [here](/docs/guides/proxysql/overview/configure-proxysql.md)
- Detail concepts of ProxySQL CRD [here](/docs/concepts/database-proxy/proxysql.md).
- Detail concepts of ProxySQLVersion CRD [here](/docs/concepts/catalog/proxysql.md).
- Quickstart ProxySQL to Load Balance MySQL Group Replication with KubeDB Operator [here](/docs/guides/proxysql/quickstart/load-balance-mysql-group-replication.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
