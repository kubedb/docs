---
title: Configure ProxySQL
menu:
  docs_{{ .version }}:
    identifier: prx-configure-proxysql-overview
    name: Configure ProxySQL
    parent: prx-overview-proxysql
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Configure ProxySQL

ProxySQL has native support for Galera Cluster and Group Replication. ProxySQL can,

- monitor of the backend servers in real-time
- redirect the traffic transparently
- reconfigure automatically

## ProxySQL for Group Replication

Here we will discuss how ProxySQL works with Group Replication.

### Group Replication

Group Replication is a plugin for the standard MySQL 5.7 and higher Server developed by Oracle. It is synchronous replication, with built-in conflict detection/handling and consistency guarantees. It allows you to move from a stand-alone instance of MySQL, which is a single point of failure, to a natively distributed highly available MySQL group made up of N MySQL instances (the group members). The servers keep strong coordination through message passing to build a fault-tolerant system.

Groups can operate in a single-primary mode, where only one server accepts updates at a time. Groups can be deployed in multi-primary mode, where all servers can accept updates. Group Replication has the following features:

- Multi-primary / active-active clustered MySQL solution
- Single-primary is the default one
- Synchronous replication
- InnoDB compliant
- State transfer based on GTID matching across all servers
- Fault-tolerant
- Built-in membership service keeps the view of the group consistent

<p align="center">
    <img alt="mysql-group-replication" src="/docs/images/proxysql/mysql-group-replication.svg">
</p>

### How to Configure

The key tables for Group Replication in ProxySQL Admin are:

- mysql_group_replication_hostgroups
- runtime_mysql_group_replication_hostgroups
- mysql_server_group_replication_log

Say we have a replication group of 3 MySQL servers.

| Host | Port |
| :--: | :--: |
| mysql-0 | 3306 |
| mysql-1 | 3306 |
| mysql-2 | 3306 |

Add these 3 members into the [mysql_servers](/docs/guides/proxysql/overview/overview.md#mysql_servers) table:

```sql
ProxySQLAdmin> INSERT INTO mysql_servers (hostgroup_id,hostname,port) VALUES (2,'mysql-0',3306);
Query OK, 1 row affected (0.00 sec)

ProxySQLAdmin> INSERT INTO mysql_servers (hostgroup_id,hostname,port) VALUES (2,'mysql-1',3306);
Query OK, 1 row affected (0.00 sec)

ProxySQLAdmin> INSERT INTO mysql_servers (hostgroup_id,hostname,port) VALUES (2,'mysql-2',3306);
Query OK, 1 row affected (0.00 sec)


ProxySQLAdmin> SELECT * FROM mysql_servers;
+--------------+----------+------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname | port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+----------+------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | mysql-0  | 3306 | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mysql-1  | 3306 | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mysql-2  | 3306 | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+----------+------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
```

Do not forget to save the configuration to disk and to load them on runtime:

```sql
ProxySQLAdmin> SAVE mysql servers to disk;
Query OK, 0 rows affected (0.01 sec)

ProxySQLAdmin> LOAD mysql servers to runtime;
Query OK, 0 rows affected (0.00 sec)
```

Now, define the hostgroups for group replication in [mysql_group_replication_hostgroup](/docs/guides/proxysql/overview/overview.md#mysql_group_replication_hostgroups) like following:

| writer_hostgroup | backup_writer_hostgroup | reader_hostgroup | offline_hostgroup | active | max_writers | writer_is_also_reader | max_transactions_behind |
| :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
| 2 | 4 | 3 | 1 | 1 | 1 | 1 | 0 |

To insert these values, we have to execute like the following query,

```sql
ProxySQLAdmin> INSERT INTO mysql_group_replication_hostgroups
(writer_hostgroup,backup_writer_hostgroup,reader_hostgroup,offline_hostgroup,active,max_writers,writer_is_also_reader,max_transactions_behind)
VALUES (2,4,3,1,1,1,1,0);
```

To apply this change we have to do additional tasks because of how ProxySQL’s configuration system works.

- **memory** which is altered when making modifications from the command-line interface.
- **runtime** which is used by ProxySQL as an effective configuration.
- **disk** which is used to make a configuration persist across restarts.

The change we made is in memory. To apply the change, we have to load the change from memory to runtime realm, then save them to disk to make them persist.

```sql
ProxySQLAdmin> LOAD admin variables to disk;
ProxySQLAdmin> LOAD admin variables to runtime;
```

The following view and helper functions also need to be added to the replication group so that the replication group can be compatible with ProxySQL’s monitoring:

- CREATE VIEW gr_member_routing_candidate_status
- CREATE FUNCTION gr_member_in_primary_partition
- CREATE FUNCTION gr_applier_queue_length
- CREATE FUNCTION GTID_COUNT
- CREATE FUNCTION GTID_NORMALIZE
- CREATE FUNCTION LOCATE2
- CREATE FUNCTION IFZERO

**Ref**: it can be found as [addition_to_sys.sql](https://gist.github.com/lefred/77ddbde301c72535381ae7af9f968322)

Execute it from the primary member,

```bash
$ mysql --user=root --password={MYSQL_ROOT_PASSWORD} --host={PRIMARY_GROUP_MEMBER} < addition_to_sys.sql
```

To verify that the above FUNCTIONs are added successfully, we can run the following statement from every member of the group,

```sql
--- Status of the primary node (mysql_node1)
mysql> SELECT * FROM sys.gr_member_routing_candidate_status;
+------------------+-----------+---------------------+----------------------+
| viable_candidate | read_only | transactions_behind | transactions_to_cert |
+------------------+-----------+---------------------+----------------------+
| YES              | NO        |                   0 |                    0 |
+------------------+-----------+---------------------+----------------------+

--- Status of a secondary node (mysql_node2)
mysql> SELECT * FROM sys.gr_member_routing_candidate_status;
+------------------+-----------+---------------------+----------------------+
| viable_candidate | read_only | transactions_behind | transactions_to_cert |
+------------------+-----------+---------------------+----------------------+
| YES              | YES       |                   0 |                    0 |
+------------------+-----------+---------------------+----------------------+
```

Once deployed ProxySQL will query the view to retrieve the status of each of the group members (the view can also be used for troubleshooting purposes).

ProxySQL needs a dedicated user to communicate with the MySQL nodes to be able to assess their condition. So create a new user for this purpose. We called it **proxysql** here. Also use a strong password for this user. Then grant the read permission on the `sys` database to this user so that it can monitor the group. If you want applications connect to this user **proxysql** to access the other databases, then grant appropriate permissions too.

```sql
(primary_member) mysql> CREATE USER '$MYSQL_PROXY_USER'@'%' IDENTIFIED BY '$MYSQL_PROXY_PASSWORD';
(primary_member) mysql> GRANT SELECT on sys.* to '$MYSQL_PROXY_USER'@'%' IDENTIFIED BY '$MYSQL_PROXY_PASSWORD';
--- If you want applications connect to this user **proxysql** to access the other databases, then grant appropriate permissions too.
--- (primary_member) mysql> GRANT ALL on *.* to '$MYSQL_PROXY_USER'@'%' IDENTIFIED BY '$MYSQL_PROXY_PASSWORD';
(primary_member) mysql> FLUSH PRIVILEGES;
```

In the above query replace the variables `$MYSQL_PROXY_USER` and `$MYSQL_PROXY_PASSWORD` with proper values.

Update ProxySQL about this new user so that it can access the group members. To tell ProxySQL we will update the right variables `mysql-monitor-username` and `mysql-monitor-password`.

```sql
ProxySQLAdmin> UPDATE global_variables
ProxySQLAdmin> SET variable_value='$MYSQL_PROXY_USER'
ProxySQLAdmin> WHERE variable_name='mysql-monitor_username';
ProxySQLAdmin> UPDATE global_variables
ProxySQLAdmin> SET variable_value='$MYSQL_PROXY_PASSWORD'
ProxySQLAdmin> WHERE variable_name='mysql-monitor_password';

ProxySQLAdmin> LOAD MYSQL VARIABLES TO RUNTIME;
ProxySQLAdmin> SAVE MYSQL VARIABLES TO DISK;
```

In the above query replace the variables `$MYSQL_PROXY_USER` and `$MYSQL_PROXY_PASSWORD` with proper values.

See ProxySQL has distributed the servers in the hostgroups:

```sql
ProxySQLAdmin>  SELECT hostgroup_id, hostname, status  FROM runtime_mysql_servers;
+--------------+----------+--------+
| hostgroup_id | hostname | status |
+--------------+----------+--------+
| 2            | mysql-0  | ONLINE |
| 3            | mysql-1  | ONLINE |
| 3            | mysql-2  | ONLINE |
+--------------+----------+--------+
```

 Here, the primary `mysql-0` is in the writer hostgroup and the secondaries `mysql-1` and `mysql-2` are in reader hostgroup.

Allow applications to connect to ProxySQL with the **proxysql** user, and send traffic to the backend servers.

To do so, we need to set configuration variables in the [mysql_users](/docs/guides/proxysql/overview/overview.md#mysql_users) table, which holds user credentials along with the default hostgroup information (which is 2 here for `writer_hostgroup`).

```sql
ProxySQLAdmin> INSERT INTO mysql_users(username, password, active, default_hostgroup, max_connections) VALUES ('$MYSQL_PROXY_USER', '$MYSQL_PROXY_PASSWORD', 1, 2, 200);
ProxySQLAdmin> LOAD MYSQL USERS TO RUNTIME;
ProxySQLAdmin> SAVE MYSQL USERS TO DISK;
```

If you take a little monitoring effort on server load, you might see that all traffic comes to the writer group (2 this case). Even when you setup single-primary for MySQL Group Replication, ProxySQL does not “automatically” route read traffics to reader nodes. Routing and balancing load are the DBA job, so it needs to be done with [mysql_query_rules](/docs/guides/proxysql/overview/overview.md#mysql_query_rules) table.

Learn more about splitting the traffic from [ProxySQL Wiki entry](https://github.com/sysown/proxysql/wiki/ProxySQL-Read-Write-Split-(HOWTO)). For example we can use the following simple rules:

```sql
ProxySQLAdmin> INSERT INTO mysql_query_rules(rule_id,active,match_digest,destination_hostgroup,apply) VALUES (1,1,'^SELECT.*FOR UPDATE$',2,1), (2,1,'^SELECT',3,1), (3,1,'.*',2,1);

ProxySQLAdmin> LOAD MYSQL QUERY RULES TO RUNTIME;
ProxySQLAdmin> SAVE MYSQL QUERY RULES TO DISK;
```

Now routing will work as follow:

- All `SELECT FOR UPDATE` will be sent to hostgroup 2
- All other `SELECT` will be sent to hostgroup 3
- Others will be sent to hostgroup 2 (the default)

We should take more effort on monitoring and on routing traffic to achieve better load distribution among DB nodes.

So here we can say that there is no more need to create a scheduler calling an external script with complex rules to move the servers in the right hostgroup.

You can view the group replication server log statistics for `mysql_server_group_replication_log` from `monitor` table:

```sql
ProxySQLAdmin> SELECT * FROM mysql_server_group_replication_log
 order by time_start_us DESC LIMIT 3\G
*************************** 1. row ***************************
 hostname: mysql-0
 port: 3306
 time_start_us: 1515079109821971
 success_time_us: 1582
 viable_candidate: YES
 read_only: NO
transactions_behind: 0
 error: NULL
*************************** 2. row ***************************
 hostname: mysql-0
 port: 3306
 time_start_us: 1515079109822292
 success_time_us: 1845
 viable_candidate: YES
 read_only: YES
transactions_behind: 0
 error: NULL
...
```

## Next Steps

- Overview of ProxySQL [here](/docs/guides/proxysql/overview/overview.md).
- Detail concepts of ProxySQL CRD [here](/docs/guides/proxysql/concepts/proxysql.md).
- Detail concepts of ProxySQLVersion CRD [here](/docs/guides/proxysql/concepts/catalog.md).
- Quickstart ProxySQL to Load Balance MySQL Group Replication with KubeDB Operator [here](/docs/guides/proxysql/quickstart/load-balance-mysql-group-replication.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
