# Configure ProxySQL

Now ProxySQL has native support for Galera Cluster and Group Replication. ProxySQL can,

- monitor of the backend servers in real time
- redirect the traffic transparently
- reconfigure automatically

## ProxySQL for Group Replication

Here we will discuss how ProxySQL works with Group Replication.

### Group Replication

Group Replication is a plugin for the standard MySQL 5.7 and higher Server developed by Oracle. It is a synchronous replication, with built-in conflict detection/handling and consistency guarantees. It allows you to move from a stand-alone instance of MySQL, which is a single point of failure, to a natively distributed highly available MySQL group made up of N MySQL instances (the group members). The servers keep strong coordination through message passing to build fault-tolerant system.

Groups can operate in a single-primary mode, where only one server accepts updates at a time. Groups can be deployed in multi-primary mode, where all servers can accept updates.

If we say specifically about the Group Replication,

- Multi-primary / active-active clustered MySQL solution
  - Single-primary is the default one
- Synchronous replication
- InnoDB compilant
- State transfer based on GTID matching across all servers
- Fault-tolerant
- Built-in membership service keeps the view of the group consistent
