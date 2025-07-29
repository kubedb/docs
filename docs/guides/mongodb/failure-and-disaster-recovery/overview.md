---
title: MongoDB Failover and DR Scenarios
menu:
  docs_{{ .version }}:
    identifier: mg-failover-disaster-recovery
    name: Overview
    parent: mg-failover-dr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ensuring Rock-Solid MongoDB Uptime

## Introduction to MongoDB Failover and Replica Sets
MongoDB provides high availability and data redundancy through a feature called replica sets. A replica set consists of a group of mongod processes working together to ensure that the database remains operational even if individual members fail.

In a typical replica set:

- The Primary node handles all write operations and replicates changes to the secondaries.

- Secondaries replicate operations from the primary to maintain an identical dataset. These nodes can also be configured with special roles such as non-voting members or priority settings.

- An optional Arbiter may participate in elections but does not store any data.

Replica sets enable automatic failover, where the remaining members detect if the primary becomes unavailable and initiate an election to promote a secondary to primary. This allows the cluster to maintain availability with minimal manual intervention.

When a primary fails to communicate with other members for longer than the configured electionTimeoutMillis (10 seconds by default), an eligible secondary calls for an election. If it receives a majority of votes, it is promoted to primary, and the cluster resumes normal operation.

To ensure data consistency and fault-tolerance, MongoDB includes:

- **Replica Set Elections:** Replica sets use elections to determine which member becomes primary, ensuring continuous availability without manual failover.

- **Rollback:** If a primary steps down before its writes are replicated to a majority, those writes are rolled back when it rejoins as a secondary.

- **Retryable Writes:** Certain write operations can be safely retried by clients, preventing duplication and ensuring success even in the presence of transient errors or failovers.

While these features require multiple MongoDB nodes, understanding how replica sets and automatic failover work is essential for designing resilient, production-grade systems — even if the current deployment starts with a single-node setup.


### Before You Start

To follow along with this tutorial, you will need:

1. A running Kubernetes cluster.
2. KubeDB [installed](https://kubedb.com/docs/v2025.5.30/setup/install/kubedb/) in your cluster.
3. kubectl command-line tool configured to communicate with your cluster.


### Step 1: Create a High-Availability MongoDBQL Cluster

First, we need to deploy a `MongoDB` cluster configured for high availability.
Unlike a Standalone instance, a HA cluster consists of a primary pod
and one or more standby pods that are ready to take over if the leader
fails.

Save the following YAML as `mg-ha-demo.yaml`. This manifest
defines a 3-node MongoDBQL cluster with streaming replication enabled.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-ha-demo
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "rs1"
  replicas: 3
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

Now, create the namespace and apply the manifest:

```shell
# Create the namespace if it doesn't exist
kubectl create ns demo

# Apply the manifest to deploy the cluster
kubectl apply -f mg-ha-demo.yaml
```

You can monitor the status until all pods are ready:
```shell
watch kubectl get mg,petset,pods -n demo
```
See the database is ready.

```shell
➤ kubectl get mg,petset,pods -n demo
NAME                            VERSION   STATUS   AGE
mongodb.kubedb.com/mg-ha-demo   4.4.26    Ready    3m58s

NAME                                      AGE
petset.apps.k8s.appscode.com/mg-ha-demo   3m52s

NAME               READY   STATUS    RESTARTS   AGE
pod/mg-ha-demo-0   2/2     Running   0          3m52s
pod/mg-ha-demo-1   2/2     Running   0          3m27s
pod/mg-ha-demo-2   2/2     Running   0          3m3s
```

Inspect who is primary and who is standby.

```shell
# you can inspect who is primary
# and who is secondary like below

➤ kubectl get pods -n demo --show-labels | grep role
mg-ha-demo-0   2/2     Running   0          5m6s    app.kubernetes.io/component=database,app.kubernetes.io/instance=mg-ha-demo,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mongodbs.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=mg-ha-demo-6b559c9645,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=mg-ha-demo-0
mg-ha-demo-1   2/2     Running   0          4m41s   app.kubernetes.io/component=database,app.kubernetes.io/instance=mg-ha-demo,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mongodbs.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=mg-ha-demo-6b559c9645,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=mg-ha-demo-1
mg-ha-demo-2   2/2     Running   0          4m17s   app.kubernetes.io/component=database,app.kubernetes.io/instance=mg-ha-demo,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mongodbs.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=mg-ha-demo-6b559c9645,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=mg-ha-demo-2
```
The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=standby` are the standby's.


Lets create a table in the primary.

```shell
➤ kubectl get secrets -n demo mg-ha-demo-auth -o jsonpath='{.data.\username}' | base64 -d
root⏎          
➤ kubectl get secrets -n demo mg-ha-demo-auth -o jsonpath='{.data.\password}' | base64 -d
JUIevJ)ISh!Srg4y⏎              
# find the primary pod
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
# exec into the primary pod
➤ mongodb@mg-ha-demo-0:/$ mongo admin
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("57604543-ec8b-478a-bca3-bdbcf4dda0b6") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://docs.mongodb.com/
Questions? Try the MongoDB Developer Community Forums
	https://community.mongodb.com
rs1:PRIMARY> db.auth("root","JUIevJ)ISh!Srg4y")
1
rs1:PRIMARY> show dbs;
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
rs1:PRIMARY> use Ballet
switched to db Ballet
rs1:PRIMARY> db.performances.insertMany([
...   {
...     title: "Swan Lake",
...     composer: "Tchaikovsky",
...     venue: "Royal Opera House",
...     date: ISODate("2025-08-01T19:00:00Z"),
...     leadDancer: "Anna Pavlova"
...   },
...   {
...     title: "The Firebird",
...     composer: "Stravinsky",
...     venue: "Lincoln Center",
...     date: ISODate("2025-09-15T20:00:00Z"),
...     leadDancer: "Misty Copeland"
...   }
... ])
{
	"acknowledged" : true,
	"insertedIds" : [
		ObjectId("688763830573023467abe95b"),
		ObjectId("688763830573023467abe95c")
	]
}
rs1:PRIMARY> show dbs
Ballet         0.000GB
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

```
### Step 2: Simulating a Failover
Before simulating failover, let's understand how KubeDB handles failover scenarios in a `MongoDB` 
replica set.
Each `MongoDB` pod is part of a replica set where one pod acts as the primary, handling all write 
operations, and the others act as secondaries, replicating data from the primary.

KubeDB continuously monitors the health of the MongoDB replica set. If the current primary 
(e.g., `mg-ha-demo-0`) becomes unavailable, the remaining voting members automatically initiate an 
election to choose a new primary. This election process is typically completed within seconds,
minimizing downtime and ensuring continued availability.

Now, the current primary is `mg-ha-demo-0`. Let’s open another terminal and run the following command to simulate a failover.


```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
It will show current mg cluster roles.

```shell
mg-ha-demo-0 standby
mg-ha-demo-1 primary
mg-ha-demo-2 standby
```

#### Case 1: Delete the current primary

Lets delete the current primary and see how the role change happens almost immediately.

```shell
➤ kubectl delete pods -n demo mg-ha-demo-0 
pod "mg-ha-demo-0" deleted
```

```shell
mg-ha-demo-0 standby
mg-ha-demo-1 primary
mg-ha-demo-2 standby
```

You see almost immediately the failover happened. Here's what happened internally:

- Distributed raft algorithm implementation is running 24 * 7 in your each db sidecar. You can configure this behavior as shown below.
- As soon as `mg-ha-demo-0` was being deleted and raft inside `mg-ha-demo-0` senses the termination, it immediately switches the leadership to any other viable leader before termination.
- In our case, raft inside `mg-ha-demo-1` got the leadership.
- Now this leader switch only means raft leader switch, not the **database leader switch(aka failover)** yet. So `mg-ha-demo-1` still running as replica. It will be primary after the next step.
- Once raft sidecar inside `mg-ha-demo-1` see it has become leader of the cluster, it initiates the database failover process and start running as primary.
- So, now `mg-ha-demo-1` is running as primary.

```yaml
# You can find this part in your db yaml by running
# kubectl get mg -n demo mg-ha-demo -oyaml
# under db.spec section
# vist below link for more information
# https://github.com/kubedb/apimachinery/blob/97c18a62d4e33a112e5f887dc3ad910edf3f3c82/apis/kubedb/v1/MongoDB_types.go#L204

leaderElection:
  electionTick: 10
  heartbeatTick: 1
  maximumLagBeforeFailover: 67108864
  period: 300ms
  transferLeadershipInterval: 1s
  transferLeadershipTimeout: 1m0s
  
```

Now we know how failover is done, let's check if the new primary is working.

```shell
➤ kubectl exec -it -n demo mg-ha-demo-1  -- bash
Defaulted container "MongoDB" out of: MongoDB, mg-coordinator, MongoDB-init-container (init)
mg-ha-demo-1:/$ psql
psql (17.2)
Type "help" for help.

MongoDB=# create table hi(id int);
CREATE TABLE # See we were able to create the database. so failover was successful.
MongoDB=# 

```


You will see the deleted pod (mg-ha-demo-0) is brought back by the kubedb operator and it is now assigned to standby role.


![img_2.png](/docs/guides/MongoDB/failure-and-disaster-recovery/img_2.png)

Lets check if the standby(`mg-ha-demo-0`) got the updated data from new primary `mg-ha-demo-1`.

```shell
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "MongoDB" out of: MongoDB, mg-coordinator, MongoDB-init-container (init)
mg-ha-demo-0:/$ psql
psql (17.2)
Type "help" for help.

MongoDB=# \dt
               List of relations
 Schema |        Name        | Type  |  Owner   
--------+--------------------+-------+----------
 public | hello              | table | MongoDB
 public | hi                 | table | MongoDB # this was created in the new primary
 public | kubedb_write_check | table | MongoDB
(3 rows)

```

#### Case 2: Delete the current primary and One replica

```shell
➤ kubectl delete pods -n demo mg-ha-demo-1 mg-ha-demo-2
pod "mg-ha-demo-1" deleted
pod "mg-ha-demo-2" deleted
```
Again we can see the failover happened pretty quickly.

![img_3.png](/docs/guides/MongoDB/failure-and-disaster-recovery/img_3.png)

After 10-30 second, the deleted pods will be back and will have its role.

![img_4.png](/docs/guides/MongoDB/failure-and-disaster-recovery/img_4.png)

Lets validate the cluster state from new primary(`mg-ha-demo-0`).

```shell
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "MongoDB" out of: MongoDB, mg-coordinator, MongoDB-init-container (init)
mg-ha-demo-0:/$ psql
psql (17.2)
Type "help" for help.

MongoDB=# select * from mg_stat_replication;
 pid  | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |    flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
------+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+-----------------+-----------------+---------------+------------+-------------------------------
 1098 |       10 | MongoDB | mg-ha-demo-1     | 10.42.0.191 |                 |       49410 | 2025-06-20 09:56:36.989448+00 |              | streaming | 0/70016A8 | 0/70016A8 | 0/70016A8 | 0/70016A8  | 00:00:00.000142 | 00:00:00.00066  | 00:00:00.000703 |             0 | async      | 2025-06-20 09:59:40.217223+00
 1129 |       10 | MongoDB | mg-ha-demo-2     | 10.42.0.192 |                 |       35216 | 2025-06-20 09:56:39.042789+00 |              | streaming | 0/70016A8 | 0/70016A8 | 0/70016A8 | 0/70016A8  | 00:00:00.000219 | 00:00:00.000745 | 00:00:00.00079  |             0 | async      | 2025-06-20 09:59:40.217308+00
(2 rows)

```

#### Case3: Delete any of the replica's

Let's delete both of the standby's.

```shell
kubectl delete pods -n demo mg-ha-demo-1 mg-ha-demo-2
pod "mg-ha-demo-1" deleted
pod "mg-ha-demo-2" deleted

```

![img_5.png](/docs/guides/MongoDB/failure-and-disaster-recovery/img_5.png)

Shortly both of the pods will be back with its role.

![img_6.png](/docs/guides/MongoDB/failure-and-disaster-recovery/img_6.png)

Lets verify cluster state.
```shell
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "MongoDB" out of: MongoDB, mg-coordinator, MongoDB-init-container (init)
mg-ha-demo-0:/$ psql
psql (17.2)
Type "help" for help.

MongoDB=# select * from mg_stat_replication;
 pid  | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |    flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
------+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+-----------------+-----------------+---------------+------------+-------------------------------
 5564 |       10 | MongoDB | mg-ha-demo-2     | 10.42.0.194 |                 |       51560 | 2025-06-20 10:06:26.988807+00 |              | streaming | 0/7014A58 | 0/7014A58 | 0/7014A58 | 0/7014A58  | 00:00:00.000178 | 00:00:00.000811 | 00:00:00.000848 |             0 | async      | 2025-06-20 10:07:50.218299+00
 5572 |       10 | MongoDB | mg-ha-demo-1     | 10.42.0.193 |                 |       36158 | 2025-06-20 10:06:27.980841+00 |              | streaming | 0/7014A58 | 0/7014A58 | 0/7014A58 | 0/7014A58  | 00:00:00.000194 | 00:00:00.000818 | 00:00:00.000895 |             0 | async      | 2025-06-20 10:07:50.218337+00
(2 rows)


```

#### Case 4: Delete both primary and all replicas

Let's delete all the pods.

```shell
➤ kubectl delete pods -n demo mg-ha-demo-0 mg-ha-demo-1 mg-ha-demo-2
pod "mg-ha-demo-0" deleted
pod "mg-ha-demo-1" deleted
pod "mg-ha-demo-2" deleted

```
![img_7.png](/docs/guides/MongoDB/failure-and-disaster-recovery/img_7.png)

Within 20-30 second, all of the pod should be back.

![img_8.png](/docs/guides/MongoDB/failure-and-disaster-recovery/img_8.png)

Lets verify the cluster state now.

```shell
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "MongoDB" out of: MongoDB, mg-coordinator, MongoDB-init-container (init)
mg-ha-demo-0:/$ psql
psql (17.2)
Type "help" for help.

MongoDB=# select * from mg_stat_replication;
 pid | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |    flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
-----+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+-----------------+-----------------+---------------+------------+-------------------------------
 132 |       10 | MongoDB | mg-ha-demo-2     | 10.42.0.197 |                 |       34244 | 2025-06-20 10:09:20.27726+00  |              | streaming | 0/9001848 | 0/9001848 | 0/9001848 | 0/9001848  | 00:00:00.00021  | 00:00:00.000841 | 00:00:00.000894 |             0 | async      | 2025-06-20 10:11:02.527633+00
 133 |       10 | MongoDB | mg-ha-demo-1     | 10.42.0.196 |                 |       40102 | 2025-06-20 10:09:20.279987+00 |              | streaming | 0/9001848 | 0/9001848 | 0/9001848 | 0/9001848  | 00:00:00.000225 | 00:00:00.000848 | 00:00:00.000905 |             0 | async      | 2025-06-20 10:11:02.527653+00
(2 rows)


```

> **We make sure the pod with highest lsn (you can think lsn as the highest data point available in your cluster) always run as primary, so if a case occur where the pod with highest lsn is being terminated, we will not perform the failover until the highest lsn pod is back online. So in a case, where that highest lsn primary is not recoverable, read [this](https://appscode.com/blog/post/kubedb-v2025.2.19/#forcefailover) to do a force failover.**


## A Guide to MongoDB Backup And Restore

You can configure Backup and Restore following the below documentation.

[Backup and Restore](/docs/guides/MongoDB/backup)

Youtube video Links: [link](https://www.youtube.com/watch?v=j9y5MsB-guQ)


## A Guide to MongoDB PITR
Documentaion Link: [PITR](/docs/guides/MongoDB/pitr)

Concepts and Demo: [link](https://www.youtube.com/watch?v=gR5UdN6Y99c)

Basic Demo: [link](https://www.youtube.com/watch?v=BdMSVjNQtCA)

Full Demo: [link](https://www.youtube.com/watch?v=KAl3rdd8i6k)

## A Guide to Handling MongoDB Storage

It is often possible that your database storage become full and your database has stopped working. We have got you covered. You just apply a VolumeExpansion `MongoDBOpsRequest` and your database storage will be increased, and the database will be ready to use again.

### Disaster Scenario and Recovery

#### Scenario

You deploy a `MongoDBQL` database. The database was running fine. Someday, your database storage becomes full. As your MongoDB process can't write to the filesystem,
clients won't be able to connect to the database. Your database status will be `Not Ready`.

#### Recovery

In order to recover from this, you can create a `VolumeExpansion` `MongoDBOpsRequest` with expanded resource requests.
As soon as you create this, KubeDB will trigger the necessary steps to expand your volume based on your specifications on the `MongoDBOpsRequest` manifest. A sample `MongoDBOpsRequest` manifest for `VolumeExpansion` is given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mgops-vol-exp-ha-demo
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: mg-ha-demo
  type: VolumeExpansion
  volumeExpansion:
    mode: Online # see the notes, your storageclass must support this mode
    MongoDB: 20Gi # expanded resource
```


For more details, please check the full section [here](/docs/guides/MongoDB/volume-expansion/Overview/overview.md).

> **Note**: There are two ways to update your volume: 1.Online 2.Offline. Which Mode to choose?
> It depends on your `StorageClass`. If your storageclass supports online volume expansion, you can go with it. Otherwise, you can go with `Ofline` Volume Expansion.

## CleanUp

```shell
kubectl delete mg -n demo mg-ha-demo
```
