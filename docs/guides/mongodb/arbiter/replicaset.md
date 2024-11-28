---
title: MongoDB ReplicaSet with Arbiter
menu:
  docs_{{ .version }}:
    identifier: mg-arbiter-replicaset
    name: ReplicaSet with Arbiter
    parent: mg-arbiter
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MongoDB ReplicaSet with Arbiter

This tutorial will show you how to use KubeDB to run a MongoDB ReplicaSet with arbiter.

## Before You Begin

Before proceeding:

- Read [mongodb arbiter concept](/docs/guides/mongodb/arbiter/concept.md) to get the concept about MongoDB Replica Set Arbiter.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MongoDB ReplicaSet with arbiter

To deploy a MongoDB ReplicaSet, user have to specify `spec.replicaSet` option in `Mongodb` CRD.

The following is an example of a `Mongodb` object which creates MongoDB ReplicaSet of three members.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongo-arb
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "rs0"
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  arbiter:
    podTemplate: {}
  deletionPolicy: WipeOut

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/arbiter/replicaset.yaml
mongodb.kubedb.com/mongo-arb created
```

Here,

- `spec.replicaSet` represents the configuration for replicaset.
  - `name` denotes the name of mongodb replicaset.
- `spec.keyFileSecret` (optional) is a secret name that contains keyfile (a random string)against `key.txt` key. Each mongod instances in the replica set and `shardTopology` uses the contents of the keyfile as the shared password for authenticating other members in the replicaset. Only mongod instances with the correct keyfile can join the replica set. _User can provide the `keyFileSecret` by creating a secret with key `key.txt`. See [here](https://docs.mongodb.com/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/#create-a-keyfile) to create the string for `keyFileSecret`._ If `keyFileSecret` is not given, KubeDB operator will generate a `keyFileSecret` itself.
- `spec.replicas` denotes the number of data-bearing members in `rs0` mongodb replicaset.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.arbiter` denotes arbiter spec of the deployed MongoDB CRD. There are two fields under it : configSecret & podTemplate. `spec.arbiter.configSecret` is an optional field to provide custom configuration file for database (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise default configuration file will be used. `spec.arbiter.podTemplate` holds the arbiter-podSpec. `null` value of it, instructs kubedb operator to use  the default arbiter podTemplate.

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create two new PetSets (one for replicas & one for arbiter) and a Service with the matching MongoDB object name. This service will always point to the primary of the replicaset. KubeDB operator will also create a governing service for the pods of those two PetSets with the name `<mongodb-name>-pods`.

```bash
$ kubectl dba describe mg -n demo mongo-arb
Name:               mongo-arb
Namespace:          demo
CreationTimestamp:  Thu, 21 Apr 2022 14:39:32 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-arb","namespace":"demo"},"spec":{"arbiter":{"podTemplat...
Replicas:           2  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          500Mi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  WipeOut

PetSet:          
  Name:               mongo-arb
  CreationTimestamp:  Thu, 21 Apr 2022 14:39:32 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mongo-arb
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
                        mongodb.kubedb.com/node.type=replica
  Annotations:        <none>
  Replicas:           824639168104 desired | 2 total
  Pods Status:        2 Running / 0 Waiting / 0 Succeeded / 0 Failed

PetSet:          
  Name:               mongo-arb-arbiter
  CreationTimestamp:  Thu, 21 Apr 2022 14:40:21 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mongo-arb
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
                        mongodb.kubedb.com/node.type=arbiter
  Annotations:        <none>
  Replicas:           824645537528 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mongo-arb
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mongo-arb
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.148.184
  Port:         primary  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.3.23:27017

Service:        
  Name:         mongo-arb-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mongo-arb
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.1.9:27017,10.244.2.18:27017,10.244.3.23:27017

Auth Secret:
  Name:         mongo-arb-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mongo-arb
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         Opaque
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-arb","namespace":"demo"},"spec":{"arbiter":{"podTemplate":null},"replicaSet":{"name":"rs0"},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"500Mi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","version":"4.4.26"}}

    Creation Timestamp:  2022-04-21T08:40:21Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mongo-arb
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mongodbs.kubedb.com
    Name:                            mongo-arb
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mongo-arb
        Port:    27017
        Scheme:  mongodb
    Parameters:
      API Version:  config.kubedb.com/v1alpha1
      Kind:         MongoConfiguration
      Replica Sets:
        host-0:  rs0/mongo-arb-0.mongo-arb-pods.demo.svc:27017,mongo-arb-1.mongo-arb-pods.demo.svc:27017,mongo-arb-arbiter-0.mongo-arb-pods.demo.svc:27017
      Stash:
        Addon:
          Backup Task:
            Name:  mongodb-backup-4.4.6
          Restore Task:
            Name:  mongodb-restore-4.4.6
    Secret:
      Name:   mongo-arb-auth
    Type:     kubedb.com/mongodb
    Version:  4.4.26

Events:
  Type    Reason      Age   From               Message
  ----    ------      ----  ----               -------
  Normal  Successful  1m    Postgres operator  Successfully created governing service
  Normal  Successful  1m    Postgres operator  Successfully created Primary Service
  Normal  Successful  1m    Postgres operator  Successfully created appbinding



$ kubectl get petset -n demo
NAME                READY   AGE
mongo-arb           2/2     2m37s
mongo-arb-arbiter   1/1     108s


$ kubectl get pvc -n demo
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
datadir-mongo-arb-0           Bound    pvc-93a2681f-096d-4af1-b1fb-93cd7b7b6020   500Mi      RWO            standard       2m57s
datadir-mongo-arb-1           Bound    pvc-fb06ea3b-a9dd-4479-87b2-de73ca272718   500Mi      RWO            standard       2m35s
datadir-mongo-arb-arbiter-0   Bound    pvc-169fd172-0e41-48e3-81a5-3abae4a85056   500Mi      RWO            standard       2m8s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                              STORAGECLASS   REASON   AGE
pvc-169fd172-0e41-48e3-81a5-3abae4a85056   500Mi      RWO            Delete           Bound    demo/datadir-mongo-arb-arbiter-0   standard                2m23s
pvc-93a2681f-096d-4af1-b1fb-93cd7b7b6020   500Mi      RWO            Delete           Bound    demo/datadir-mongo-arb-0           standard                3m11s
pvc-fb06ea3b-a9dd-4479-87b2-de73ca272718   500Mi      RWO            Delete           Bound    demo/datadir-mongo-arb-1           standard                2m50s


$ kubectl get service -n demo
NAME             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
mongo-arb        ClusterIP   10.96.148.184   <none>        27017/TCP   3m32s
mongo-arb-pods   ClusterIP   None            <none>        27017/TCP   3m32s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mongo-arb -o yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-arb","namespace":"demo"},"spec":{"arbiter":{"podTemplate":null},"replicaSet":{"name":"rs0"},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"500Mi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","version":"4.4.26"}}
  creationTimestamp: "2022-04-21T08:39:32Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: mongo-arb
  namespace: demo
  resourceVersion: "22168"
  uid: c4a3dc69-5556-42b6-a2b8-11d3547015d3
spec:
  allowedSchemas:
    namespaces:
      from: Same
  arbiter:
    podTemplate:
      controller: {}
      metadata: {}
      spec:
        resources:
          limits:
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 1Gi
  authSecret:
    name: mongo-arb-auth
  clusterAuthMode: keyFile
  keyFileSecret:
    name: mongo-arb-key
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: mongo-arb
  replicaSet:
    name: rs0
  replicas: 2
  sslMode: disabled
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
    storageClassName: standard
  storageEngine: wiredTiger
  storageType: Durable
  deletionPolicy: WipeOut
  version: 4.4.26
status:
  conditions:
  - lastTransitionTime: "2022-04-21T08:39:32Z"
    message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mongo-arb'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2022-04-21T08:40:42Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2022-04-21T08:39:56Z"
    message: 'The MongoDB: demo/mongo-arb is accepting client requests.'
    observedGeneration: 3
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2022-04-21T08:39:56Z"
    message: 'The MongoDB: demo/mongo-arb is ready.'
    observedGeneration: 3
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2022-04-21T08:40:21Z"
    message: 'The MongoDB: demo/mongo-arb is successfully provisioned.'
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 3
  phase: Ready

```

Please note that KubeDB operator has created a new Secret called `mongo-arb-auth` *(format: {mongodb-object-name}-auth)* for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the *username* for MongoDB superuser and a `password` key which contains the *password* for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

## Redundancy and Data Availability

Now, you can connect to this database through [mongo-arb](https://docs.mongodb.com/v3.4/mongo/). In this tutorial, we will insert document on the primary member, and we will see if the data becomes available on secondary members.

At first, insert data inside primary member `rs0:PRIMARY`.

```bash
$ kubectl get secrets -n demo mongo-arb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mongo-arb-auth -o jsonpath='{.data.\password}' | base64 -d
OX4yb!IFm;~yAHkD

$ kubectl exec -it mongo-arb-0 -n demo bash

mongodb@mongo-arb-0:/$ mongo admin -u root -p 'OX4yb!IFm;~yAHkD'
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.

rs0:PRIMARY> rs.status()
{
	"set" : "rs0",
	"date" : ISODate("2022-04-21T08:46:28.786Z"),
	"myState" : 1,
	"term" : NumberLong(1),
	"syncSourceHost" : "",
	"syncSourceId" : -1,
	"heartbeatIntervalMillis" : NumberLong(2000),
	"majorityVoteCount" : 2,
	"writeMajorityCount" : 2,
	"votingMembersCount" : 3,
	"writableVotingMembersCount" : 2,
	"optimes" : {
		"lastCommittedOpTime" : {
			"ts" : Timestamp(1650530787, 1),
			"t" : NumberLong(1)
		},
		"lastCommittedWallTime" : ISODate("2022-04-21T08:46:27.247Z"),
		"readConcernMajorityOpTime" : {
			"ts" : Timestamp(1650530787, 1),
			"t" : NumberLong(1)
		},
		"readConcernMajorityWallTime" : ISODate("2022-04-21T08:46:27.247Z"),
		"appliedOpTime" : {
			"ts" : Timestamp(1650530787, 1),
			"t" : NumberLong(1)
		},
		"durableOpTime" : {
			"ts" : Timestamp(1650530787, 1),
			"t" : NumberLong(1)
		},
		"lastAppliedWallTime" : ISODate("2022-04-21T08:46:27.247Z"),
		"lastDurableWallTime" : ISODate("2022-04-21T08:46:27.247Z")
	},
	"lastStableRecoveryTimestamp" : Timestamp(1650530747, 1),
	"electionCandidateMetrics" : {
		"lastElectionReason" : "electionTimeout",
		"lastElectionDate" : ISODate("2022-04-21T08:39:47.205Z"),
		"electionTerm" : NumberLong(1),
		"lastCommittedOpTimeAtElection" : {
			"ts" : Timestamp(0, 0),
			"t" : NumberLong(-1)
		},
		"lastSeenOpTimeAtElection" : {
			"ts" : Timestamp(1650530387, 1),
			"t" : NumberLong(-1)
		},
		"numVotesNeeded" : 1,
		"priorityAtElection" : 1,
		"electionTimeoutMillis" : NumberLong(10000),
		"newTermStartDate" : ISODate("2022-04-21T08:39:47.221Z"),
		"wMajorityWriteAvailabilityDate" : ISODate("2022-04-21T08:39:47.234Z")
	},
	"members" : [
		{
			"_id" : 0,
			"name" : "mongo-arb-0.mongo-arb-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 1,
			"stateStr" : "PRIMARY",
			"uptime" : 412,
			"optime" : {
				"ts" : Timestamp(1650530787, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-04-21T08:46:27Z"),
			"syncSourceHost" : "",
			"syncSourceId" : -1,
			"infoMessage" : "",
			"electionTime" : Timestamp(1650530387, 2),
			"electionDate" : ISODate("2022-04-21T08:39:47Z"),
			"configVersion" : 3,
			"configTerm" : 1,
			"self" : true,
			"lastHeartbeatMessage" : ""
		},
		{
			"_id" : 1,
			"name" : "mongo-arb-1.mongo-arb-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 375,
			"optime" : {
				"ts" : Timestamp(1650530787, 1),
				"t" : NumberLong(1)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1650530787, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-04-21T08:46:27Z"),
			"optimeDurableDate" : ISODate("2022-04-21T08:46:27Z"),
			"lastHeartbeat" : ISODate("2022-04-21T08:46:27.456Z"),
			"lastHeartbeatRecv" : ISODate("2022-04-21T08:46:27.591Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mongo-arb-0.mongo-arb-pods.demo.svc.cluster.local:27017",
			"syncSourceId" : 0,
			"infoMessage" : "",
			"configVersion" : 3,
			"configTerm" : 1
		},
		{
			"_id" : 2,
			"name" : "mongo-arb-arbiter-0.mongo-arb-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 7,
			"stateStr" : "ARBITER",
			"uptime" : 353,
			"lastHeartbeat" : ISODate("2022-04-21T08:46:27.450Z"),
			"lastHeartbeatRecv" : ISODate("2022-04-21T08:46:27.607Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "",
			"syncSourceId" : -1,
			"infoMessage" : "",
			"configVersion" : 3,
			"configTerm" : 1
		}
	],
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1650530787, 1),
		"signature" : {
			"hash" : BinData(0,"N6pWJaxVqaZch7cKLKWX8bdfkBM="),
			"keyId" : NumberLong("7088974033219223556")
		}
	},
	"operationTime" : Timestamp(1650530787, 1)
}
```

Here you can see the arbiter pod in the members list of `rs.status()` output.

```bash
rs0:PRIMARY> > rs.isMaster().primary
mongo-arb-0.mongo-arb-pods.demo.svc.cluster.local:27017

rs0:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

rs0:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("af3c1344-d052-496a-bdbb-5bd41486d878"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	],
	"mechanisms" : [
		"SCRAM-SHA-1",
		"SCRAM-SHA-256"
	]
}


rs0:PRIMARY> use mydb
switched to db mydb
rs0:PRIMARY> db.songs.insert({"pink floyd": "shine on you crazy diamond"})
WriteResult({ "nInserted" : 1 })
rs0:PRIMARY> db.songs.find().pretty()
{
	"_id" : ObjectId("62611ae33583279dfca0a5e4"),
	"pink floyd" : "shine on you crazy diamond"
}

rs0:PRIMARY> exit
bye
```

Now, check the redundancy and data availability in secondary members.
We will exec in `mongo-arb-1`(which is secondary member right now) to check the data availability.

```bash
$ kubectl exec -it mongo-arb-1 -n demo bash
mongodb@mongo-arb-1:/$ mongo admin -u root -p 'OX4yb!IFm;~yAHkD'
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.

rs0:SECONDARY> rs.slaveOk()
rs0:SECONDARY> > show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
mydb           0.000GB

rs0:SECONDARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("af3c1344-d052-496a-bdbb-5bd41486d878"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	],
	"mechanisms" : [
		"SCRAM-SHA-1",
		"SCRAM-SHA-256"
	]
}

rs0:SECONDARY> use mydb
switched to db mydb

rs0:SECONDARY> db.songs.find().pretty()
{
	"_id" : ObjectId("62611ae33583279dfca0a5e4"),
	"pink floyd" : "shine on you crazy diamond"
}

rs0:SECONDARY> exit
bye

```

## Automatic Failover

To test automatic failover, we will force the primary member to restart. As the primary member (`pod`) becomes unavailable, the rest of the members will elect a primary member by election.

```bash
$ kubectl get pods -n demo
NAME                  READY   STATUS    RESTARTS   AGE
mongo-arb-0           2/2     Running   0          15m
mongo-arb-1           2/2     Running   0          14m
mongo-arb-arbiter-0   1/1     Running   0          14m

$ kubectl delete pod -n demo mongo-arb-0
pod "mongo-arb-0" deleted

$ kubectl get pods -n demo
NAME                  READY   STATUS        RESTARTS   AGE
mongo-arb-0           2/2     Terminating   0          16m
mongo-arb-1           2/2     Running       0          15m
mongo-arb-arbiter-0   1/1     Running       0          15m
```

Now verify the automatic failover, Let's exec in `mongo-arb-0` pod,

```bash
$ kubectl exec -it mongo-arb-0  -n demo bash
mongodb@mongo-arb-1:/$ mongo admin -u root -p 'OX4yb!IFm;~yAHkD'
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.

rs0:SECONDARY> rs.isMaster().primary
mongo-arb-1.mongo-arb-pods.demo.svc.cluster.local:27017

# Also verify, data persistency
rs0:SECONDARY> rs.slaveOk()
rs0:SECONDARY> > show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
mydb           0.000GB

rs0:SECONDARY> use mydb
switched to db mydb

rs0:SECONDARY> db.songs.find().pretty()
{
	"_id" : ObjectId("62611ae33583279dfca0a5e4"),
	"pink floyd" : "shine on you crazy diamond"
}

```

## Halt Database

When [DeletionPolicy](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy) is set to halt, and you delete the mongodb object, the KubeDB operator will delete the PetSet and its pods but leaves the PVCs, secrets and database backup (snapshots) intact. Learn details of all `DeletionPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy).

You can also keep the mongodb object and halt the database to resume it again later. If you halt the database, the kubedb will delete the petsets and services but will keep the mongodb object, pvcs, secrets and backup (snapshots).

To halt the database, first you have to set the deletionPolicy to `Halt` in existing database. You can use the below command to set the deletionPolicy to `Halt`, if it is not already set.

```bash
$ kubectl patch -n demo mg/mongo-arb -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"
mongodb.kubedb.com/mongo-arb patched
```

Then, you have to set the `spec.halted` as true to set the database in a `Halted` state. You can use the below command.

```bash
$ kubectl patch -n demo mg/mongo-arb -p '{"spec":{"halted":true}}' --type="merge"
mongodb.kubedb.com/mongo-arb patched
```

After that, kubedb will delete the petsets and services and you can see the database Phase as `Halted`.

Now, you can run the following command to get all mongodb resources in demo namespaces,

```bash
$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mongo-arb   4.4.26     Halted   21m

NAME                         TYPE                                  DATA   AGE
secret/default-token-nzk64   kubernetes.io/service-account-token   3      146m
secret/mongo-arb-auth        Opaque                                2      21m
secret/mongo-arb-key         Opaque                                1      21m

NAME                                                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mongo-arb-0           Bound    pvc-93a2681f-096d-4af1-b1fb-93cd7b7b6020   500Mi      RWO            standard       21m
persistentvolumeclaim/datadir-mongo-arb-1           Bound    pvc-fb06ea3b-a9dd-4479-87b2-de73ca272718   500Mi      RWO            standard       21m
persistentvolumeclaim/datadir-mongo-arb-arbiter-0   Bound    pvc-169fd172-0e41-48e3-81a5-3abae4a85056   500Mi      RWO            standard       21m
```


## Resume Halted Database

Now, to resume the database, i.e. to get the same database setup back again, you have to set the the `spec.halted` as false. You can use the below command.

```bash
$ kubectl patch -n demo mg/mongo-arb  -p '{"spec":{"halted":false}}' --type="merge"
mongodb.kubedb.com/mongo-arb patched
```

When the database is resumed successfully, you can see the database Status is set to `Ready`.

```bash
$ kubectl get mg -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mongo-arb   4.4.26     Ready    23m
```

Now, If you again exec into the primary `pod` and look for previous data, you will see that, all the data persists.

```bash
$ kubectl exec -it mongo-arb-1 -n demo bash

mongodb@mongo-arb-1:/$ mongo admin -u root -p 'OX4yb!IFm;~yAHkD'

rs0:PRIMARY> use mydb
switched to db mydb
rs0:PRIMARY> db.songs.find().pretty()
{
	"_id" : ObjectId("62611ae33583279dfca0a5e4"),
	"pink floyd" : "shine on you crazy diamond"
}
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mongo-arb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mongo-arb

kubectl delete ns demo
```

## Next Steps

- Deploy MongoDB shard [with Arbiter](/docs/guides/mongodb/arbiter/sharding.md).
- [Backup and Restore](/docs/guides/mongodb/backup/stash/overview/index.md) process of MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
