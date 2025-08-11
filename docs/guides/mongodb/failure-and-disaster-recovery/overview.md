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
2. KubeDB [installed](https://kubedb.com/docs/{{< param "info.version" >}}/setup/install/kubedb/) in your cluster.
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

Now, connect to a secondary node to inspect how the data reflects changes from the primary, and
observe any visible differences between their states.

```shell
kubectl exec -it -n demo mg-ha-demo-1  -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mg-ha-demo-1:/$ mongo admin
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("9ec7b2cc-8972-4f07-996f-f6c9573acc36") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://docs.mongodb.com/
Questions? Try the MongoDB Developer Community Forums
	https://community.mongodb.com
rs1:PRIMARY> db.auth("root","JUIevJ)ISh!Srg4y")
1

rs1:SECONDARY> rs.slaveOk()
WARNING: slaveOk() is deprecated and may be removed in the next major release. Please use secondaryOk() instead.
rs1:SECONDARY> show dbs
Ballet         0.000GB
Kathak         0.000GB
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
rs1:SECONDARY> use new_db
switched to db new_db
rs1:SECONDARY> db.performances.insertMany([
... ... ...   {
... ... ...     title: "Swan Lake",
... ... ...     composer: "Tchaikovsky",
... ... ...     venue: "Royal Opera House",
... ... ...     date: ISODate("2025-08-01T19:00:00Z"),
... ... ...     leadDancer: "Anna Pavlova"
... ... ...   }
... ... ])
uncaught exception: WriteCommandError({
	"topologyVersion" : {
		"processId" : ObjectId("68884ca04453725f9ae4b166"),
		"counter" : NumberLong(3)
	},
	"operationTime" : Timestamp(1753777717, 1),
	"ok" : 0,
	"errmsg" : "not master",
	"code" : 10107,
	"codeName" : "NotWritablePrimary",
	"$clusterTime" : {
		"clusterTime" : Timestamp(1753777717, 1),
		"signature" : {
			"hash" : BinData(0,"feVYdK9HiGlqlKF5dLbcpWrbOTs="),
			"keyId" : NumberLong("7532089958085951493")
		}
	}
}) :
WriteCommandError({
	"topologyVersion" : {
		"processId" : ObjectId("68884ca04453725f9ae4b166"),
		"counter" : NumberLong(3)
	},
	"operationTime" : Timestamp(1753777717, 1),
	"ok" : 0,
	"errmsg" : "not master",
	"code" : 10107,
	"codeName" : "NotWritablePrimary",
	"$clusterTime" : {
		"clusterTime" : Timestamp(1753777717, 1),
		"signature" : {
			"hash" : BinData(0,"feVYdK9HiGlqlKF5dLbcpWrbOTs="),
			"keyId" : NumberLong("7532089958085951493")
		}
	}
})
WriteCommandError@src/mongo/shell/bulk_api.js:417:48
executeBatch@src/mongo/shell/bulk_api.js:915:23
Bulk/this.execute@src/mongo/shell/bulk_api.js:1163:21
DBCollection.prototype.insertMany@src/mongo/shell/crud_api.js:326:5
@(shell):1:1

```

### Step 2: Simulating a Failover
#### Replica set elections
You will see almost immediately the failover happened. Before simulating failover, let's understand how 
KubeDB handles failover scenarios in a `MongoDB` replica set. Here's what happened internally:

- MongoDB replica sets support automatic failover to ensure high availability.
- All replica set members continuously exchange heartbeats every 2 seconds to monitor each other’s status.
- If the primary becomes unreachable (e.g., crash, network issue), secondaries detect the failure based on missed heartbeats.
- Once a majority of voting members agree the primary is down, an election is triggered.
- Any eligible secondary can declare itself a candidate and request votes from other members.
- A new primary is elected if a candidate receives majority votes (e.g., 2 out of 3 in a 3-node setup).
- The election considers oplog freshness, priority settings, and election term history to prevent stale nodes from taking over.
- Once elected, the new primary begins accepting write operations.
- MongoDB clients using replica set URIs automatically reroute traffic to the new primary.
- The old primary, when it rejoins, steps down to secondary and syncs missed operations if necessary.
- During the election, writes are paused, but reads can still occur if the read preference is set to allow secondaries.
- A majority quorum must be present to elect a new primary; otherwise, the replica set becomes read-only until quorum is restored.

Now, the current primary is `mg-ha-demo-0`. Let’s open another terminal and run the following command to simulate a failover.


```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
It will show current mg cluster roles.

```shell
mg-ha-demo-0 primary
mg-ha-demo-1 standby
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

Now we know how failover is done, let's check if the new primary is working.

```shell
➤ `kubectl exec -it -n demo mg-ha-demo-1  -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mg-ha-demo-1:/$ mongo admin
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("9ec7b2cc-8972-4f07-996f-f6c9573acc36") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://docs.mongodb.com/
Questions? Try the MongoDB Developer Community Forums
	https://community.mongodb.com
rs1:PRIMARY> db.auth("root","JUIevJ)ISh!Srg4y")
1
rs1:PRIMARY> show dbs
Ballet         0.000GB
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB`
rs1:PRIMARY> use Kathak
switched to db Kathak

rs1:PRIMARY> db.performances.insertMany([
... ...   {
... ...     title: "Swan Lake",
... ...     composer: "Tchaikovsky",
... ...     venue: "Royal Opera House",
... ...     date: ISODate("2025-08-01T19:00:00Z"),
... ...     leadDancer: "Anna Pavlova"
... ...   }
... ])
{
	"acknowledged" : true,
	"insertedIds" : [
		ObjectId("68887a44f9b3c13aeffdcff1")
	]
}
rs1:PRIMARY> show dbs
Ballet         0.000GB
Kathak         0.000GB
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

```

You will see the deleted pod `mg-ha-demo-0` is brought back by the kubedb operator and it is now assigned to standby role.
Lets check if the standby `mg-ha-demo-0` got the updated data from new primary `mg-ha-demo-1`.

```shell
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mg-ha-demo-0:/$ mongo admin
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("8c1a76f9-61cf-4105-b3b1-d82190640c87") }
MongoDB server version: 4.4.26
rs1:SECONDARY> db.auth("root","JUIevJ)ISh!Srg4y")
1
rs1:SECONDARY> rs.slaveOk()
WARNING: slaveOk() is deprecated and may be removed in the next major release. Please use secondaryOk() instead.
rs1:SECONDARY> show dbs
Ballet         0.000GB
Kathak         0.000GB
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
rs1:SECONDARY> 

```

#### Case 2: Delete the current primary and One replica

```shell
➤ kubectl delete pods -n demo mg-ha-demo-1 mg-ha-demo-2
pod "mg-ha-demo-1" deleted
pod "mg-ha-demo-2" deleted
```
Again we can see the failover happened pretty quickly.
```shell
mg-ha-demo-0 primary
mg-ha-demo-1 
mg-ha-demo-2 
```
After 10-30 second, the deleted pods will be back and will have its role.
```shell
mg-ha-demo-0 primary
mg-ha-demo-1 standby
mg-ha-demo-2 standby
```
You can validate the replica set status from the new primary `mg-ha-demo-0` by checking the role, state, and health of each member.
```shell
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mg-ha-demo-0:/$ mongo admin
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("8c1a76f9-61cf-4105-b3b1-d82190640c87") }
MongoDB server version: 4.4.26
rs1:SECONDARY> db.auth("root","JUIevJ)ISh!Srg4y")
1
rs1:PRIMARY> rs.status()
{
	"set" : "rs1",
	"date" : ISODate("2025-07-29T09:03:41.332Z"),
	"myState" : 1,
	"term" : NumberLong(4),
	"syncSourceHost" : "",
	"syncSourceId" : -1,
	"heartbeatIntervalMillis" : NumberLong(2000),
	"majorityVoteCount" : 2,
	"writeMajorityCount" : 2,
	"votingMembersCount" : 3,
	"writableVotingMembersCount" : 3,
	"optimes" : {
		"lastCommittedOpTime" : {
			"ts" : Timestamp(1753779817, 1),
			"t" : NumberLong(4)
		},
		"lastCommittedWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
		"readConcernMajorityOpTime" : {
			"ts" : Timestamp(1753779817, 1),
			"t" : NumberLong(4)
		},
		"readConcernMajorityWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
		"appliedOpTime" : {
			"ts" : Timestamp(1753779817, 1),
			"t" : NumberLong(4)
		},
		"durableOpTime" : {
			"ts" : Timestamp(1753779817, 1),
			"t" : NumberLong(4)
		},
		"lastAppliedWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
		"lastDurableWallTime" : ISODate("2025-07-29T09:03:37.338Z")
	},
	"lastStableRecoveryTimestamp" : Timestamp(1753779787, 1),
	"electionCandidateMetrics" : {
		"lastElectionReason" : "stepUpRequestSkipDryRun",
		"lastElectionDate" : ISODate("2025-07-29T08:38:57.291Z"),
		"electionTerm" : NumberLong(4),
		"lastCommittedOpTimeAtElection" : {
			"ts" : Timestamp(1753778337, 1),
			"t" : NumberLong(3)
		},
		"lastSeenOpTimeAtElection" : {
			"ts" : Timestamp(1753778337, 1),
			"t" : NumberLong(3)
		},
		"numVotesNeeded" : 2,
		"priorityAtElection" : 1,
		"electionTimeoutMillis" : NumberLong(10000),
		"priorPrimaryMemberId" : 1,
		"numCatchUpOps" : NumberLong(0),
		"newTermStartDate" : ISODate("2025-07-29T08:38:57.312Z"),
		"wMajorityWriteAvailabilityDate" : ISODate("2025-07-29T08:39:03.560Z")
	},
	"members" : [
		{
			"_id" : 0,
			"name" : "mg-ha-demo-0.mg-ha-demo-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 1,
			"stateStr" : "PRIMARY",
			"uptime" : 16845,
			"optime" : {
				"ts" : Timestamp(1753779817, 1),
				"t" : NumberLong(4)
			},
			"optimeDate" : ISODate("2025-07-29T09:03:37Z"),
			"lastAppliedWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
			"lastDurableWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
			"syncSourceHost" : "",
			"syncSourceId" : -1,
			"infoMessage" : "",
			"electionTime" : Timestamp(1753778337, 2),
			"electionDate" : ISODate("2025-07-29T08:38:57Z"),
			"configVersion" : 3,
			"configTerm" : 4,
			"self" : true,
			"lastHeartbeatMessage" : ""
		},
		{
			"_id" : 1,
			"name" : "mg-ha-demo-1.mg-ha-demo-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 1442,
			"optime" : {
				"ts" : Timestamp(1753779817, 1),
				"t" : NumberLong(4)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1753779817, 1),
				"t" : NumberLong(4)
			},
			"optimeDate" : ISODate("2025-07-29T09:03:37Z"),
			"optimeDurableDate" : ISODate("2025-07-29T09:03:37Z"),
			"lastAppliedWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
			"lastDurableWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
			"lastHeartbeat" : ISODate("2025-07-29T09:03:39.341Z"),
			"lastHeartbeatRecv" : ISODate("2025-07-29T09:03:39.569Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mg-ha-demo-0.mg-ha-demo-pods.demo.svc.cluster.local:27017",
			"syncSourceId" : 0,
			"infoMessage" : "",
			"configVersion" : 3,
			"configTerm" : 4
		},
		{
			"_id" : 2,
			"name" : "mg-ha-demo-2.mg-ha-demo-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 1442,
			"optime" : {
				"ts" : Timestamp(1753779817, 1),
				"t" : NumberLong(4)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1753779817, 1),
				"t" : NumberLong(4)
			},
			"optimeDate" : ISODate("2025-07-29T09:03:37Z"),
			"optimeDurableDate" : ISODate("2025-07-29T09:03:37Z"),
			"lastAppliedWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
			"lastDurableWallTime" : ISODate("2025-07-29T09:03:37.338Z"),
			"lastHeartbeat" : ISODate("2025-07-29T09:03:39.341Z"),
			"lastHeartbeatRecv" : ISODate("2025-07-29T09:03:40.606Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mg-ha-demo-0.mg-ha-demo-pods.demo.svc.cluster.local:27017",
			"syncSourceId" : 0,
			"infoMessage" : "",
			"configVersion" : 3,
			"configTerm" : 4
		}
	],
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1753779817, 1),
		"signature" : {
			"hash" : BinData(0,"8AgSQtGZAjbr4NWOcFl7R7ia1ZU="),
			"keyId" : NumberLong("7532089958085951493")
		}
	},
	"operationTime" : Timestamp(1753779817, 1)
}
# Here "stateStr" field shows if the node is primary or secondary.
rs1:PRIMARY> rs.status().members.forEach(function(member) {
...   const hostParts = member.name.split(":");
...   const host = hostParts[0];
...   const port = hostParts[1];
...   const state = member.stateStr;
...   const role = state === "PRIMARY" ? "PRIMARY" : "SECONDARY";
...   print(
...     host.padEnd(45) + "\t" +
...     port.padEnd(8) + "\t" +
...     "ONLINE".padEnd(10) + "\t" +
...     role
...   );
... });
mg-ha-demo-0.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	PRIMARY
mg-ha-demo-1.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	SECONDARY
mg-ha-demo-2.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	SECONDARY

```

#### Case3: Delete any of the replica's

Let's delete both of the standby's.

```shell
kubectl delete pods -n demo mg-ha-demo-1 mg-ha-demo-2
pod "mg-ha-demo-1" deleted
pod "mg-ha-demo-2" deleted

```
```shell
mg-ha-demo-0 primary
mg-ha-demo-1 
```

Shortly both of the pods will be back with its role.
```shell
mg-ha-demo-0 primary
mg-ha-demo-1 standby
mg-ha-demo-2 standby

```
Lets verify cluster state.
```shell
➤ kubectl exec -it -n demo mg-ha-demo-0  -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mg-ha-demo-0:/$ mongo admin
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("8c1a76f9-61cf-4105-b3b1-d82190640c87") }
MongoDB server version: 4.4.26
rs1:SECONDARY> db.auth("root","JUIevJ)ISh!Srg4y")
1

rs1:PRIMARY> rs.status().members.forEach(function(member) {
...   const hostParts = member.name.split(":");
...   const host = hostParts[0];
...   const port = hostParts[1];
...   const state = member.stateStr;
...   const role = state === "PRIMARY" ? "PRIMARY" : "SECONDARY";
...   print(
...     host.padEnd(45) + "\t" +
...     port.padEnd(8) + "\t" +
...     "ONLINE".padEnd(10) + "\t" +
...     role
...   );
... });
mg-ha-demo-0.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	PRIMARY
mg-ha-demo-1.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	SECONDARY
mg-ha-demo-2.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	SECONDARY

```

#### Case 4: Delete both primary and all replicas

Let's delete all the pods.

```shell
➤ kubectl delete pods -n demo mg-ha-demo-0 mg-ha-demo-1 mg-ha-demo-2
pod "mg-ha-demo-0" deleted
pod "mg-ha-demo-1" deleted
pod "mg-ha-demo-2" deleted

```

```shell
mg-ha-demo-0 standby
mg-ha-demo-1
mg-ha-demo-2
```

Within 20-30 second, all of the pod should be back.

```shell
mg-ha-demo-0 standby
mg-ha-demo-1 primary
mg-ha-demo-2 standby

```

Lets verify the cluster state now.

```shell
➤  kubectl exec -it -n demo mg-ha-demo-1  -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mg-ha-demo-1:/$ mongo admin
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("f99795f2-9fb1-4221-bf2c-916f0e92d152") }
MongoDB server version: 4.4.26
rs1:PRIMARY> db.auth("root","JUIevJ)ISh!Srg4y")
1
rs1:PRIMARY> rs.status().members.forEach(function(member) {
... ...   const hostParts = member.name.split(":");
... ...   const host = hostParts[0];
... ...   const port = hostParts[1];
... ...   const state = member.stateStr;
... ...   const role = state === "PRIMARY" ? "PRIMARY" : "SECONDARY";
... ...   print(
... ...     host.padEnd(45) + "\t" +
... ...     port.padEnd(8) + "\t" +
... ...     "ONLINE".padEnd(10) + "\t" +
... ...     role
... ...   );
... ... });
mg-ha-demo-0.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	SECONDARY
mg-ha-demo-1.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	PRIMARY
mg-ha-demo-2.mg-ha-demo-pods.demo.svc.cluster.local	27017   	ONLINE    	SECONDARY
```
#### Retryable Writes
Retryable writes allow MongoDB drivers to safely retry certain write operations 
(like insert, update, delete) once automatically if a network error or primary failover occurs.

Operational Details: 
- Each retryable write includes a unique transaction identifier.
- If a write fails due to network issues, failover, or not acknowledging, the driver retries the operation once.
- `MongoDB` ensures that even if retried, the operation is only executed once (idempotent behavior).
- The server uses the transaction ID to recognize duplicate attempts and suppress duplicates.

#### Rollback in MongoDB
A `rollback` in `MongoDB` occurs when a node that was previously acting as the primary performs write 
operations but then loses its primary status before those writes are replicated to a majority of the
replica set members.

This typically happens during a `network partition`, where the isolated primary continues accepting writes,
unaware that it no longer holds the majority. Meanwhile, the remaining members elect a new primary. 
When the original primary rejoins the replica set, `MongoDB` detects a divergence between its data and the new
primary’s state. To resolve this, MongoDB rolls back the unreplicated writes from the old primary,
restoring consistency with the current primary.

## CleanUp
Run the following command for cleanup:
```shell
# delete the cluster of MongoDB
kubectl delete mg -n demo mg-ha-demo

# or delete the namespace
kubectl delete ns demo
```
