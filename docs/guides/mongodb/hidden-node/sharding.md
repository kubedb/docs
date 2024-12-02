---
title: MongoDB Sharding Guide with Hidden node
menu:
  docs_{{ .version }}:
    identifier: mg-hidden-sharding
    name: Sharding with Hidden node
    parent: mg-hidden
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Sharding with Hidden-node

This tutorial will show you how to use KubeDB to run a sharded MongoDB cluster with hidden node.

## Before You Begin

Before proceeding:

- Read [mongodb hidden-node concept](/docs/guides/mongodb/hidden-node/concept.md) to get the concept about MongoDB Replica Set Hidden-node.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Sharded MongoDB Cluster

To deploy a MongoDB Sharding, user have to specify `spec.shardTopology` option in `Mongodb` CRD.

The following is an example of a `Mongodb` object which creates MongoDB Sharding of three type of members.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongo-sh-hid
  namespace: demo
spec:
  version: "percona-7.0.4"
  shardTopology:
    configServer:
      replicas: 3
      ephemeralStorage: {}
    mongos:
      replicas: 2
    shard:
      replicas: 3
      shards: 2
      ephemeralStorage: {}
  storageEngine: inMemory
  storageType: Ephemeral
  hidden:
    podTemplate:
      spec:
        resources:
          requests:
            cpu: "400m"
            memory: "400Mi"
    replicas: 2
    storage:
      storageClassName: "standard"
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/hidden-node/sharding.yaml
mongodb.kubedb.com/mongo-sh-hid created
```

Here,

- `spec.shardTopology` represents the topology configuration for sharding.
  - `shard` represents configuration for Shard component of mongodb.
    - `shards` represents number of shards for a mongodb deployment. Each shard is deployed as a [replicaset](/docs/guides/mongodb/clustering/replication_concept.md).
    - `replicas` represents number of replicas of each shard replicaset.
    - `prefix` represents the prefix of each shard node.
    - `configSecret` is an optional field to provide custom configuration file for shards (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used.
    - `podTemplate` is an optional configuration for pods.
    - `storage` to specify pvc spec for each node of sharding. You can specify any StorageClass available in your cluster with appropriate resource requests.
  - `configServer` represents configuration for ConfigServer component of mongodb.
    - `replicas` represents number of replicas for configServer replicaset. Here, configServer is deployed as a replicaset of mongodb.
    - `prefix` represents the prefix of configServer nodes.
    - `configSecret` is an optional field to provide custom configuration file for configSource (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used.
    - `podTemplate` is an optional configuration for pods.
    - `storage` to specify pvc spec for each node of configServer. You can specify any StorageClass available in your cluster with appropriate resource requests.
  - `mongos` represents configuration for Mongos component of mongodb. `Mongos` instances run as stateless components (deployment).
    - `replicas` represents number of replicas of `Mongos` instance. Here, Mongos is not deployed as replicaset.
    - `prefix` represents the prefix of mongos nodes.
    - `configSecret` is an optional field to provide custom configuration file for mongos (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used.
    - `podTemplate` is an optional configuration for pods.
- `spec.keyFileSecret` (optional) is a secret name that contains keyfile (a random string)against `key.txt` key. Each mongod instances in the replica set and `shardTopology` uses the contents of the keyfile as the shared password for authenticating other members in the replicaset. Only mongod instances with the correct keyfile can join the replica set. _User can provide the `keyFileSecret` by creating a secret with key `key.txt`. See [here](https://docs.mongodb.com/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/#create-a-keyfile) to create the string for `keyFileSecret`._ If `keyFileSecret` is not given, KubeDB operator will generate a `keyFileSecret` itself.
- `spec.storageEngine` is set to inMemory, & `spec.storageType` to ephemeral.
- `spec.shardTopology.(configSerer/shard).ephemeralStorage` holds the emptyDir volume specifications. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this ephemeral storage configuration.
- `spec.hidden` denotes hidden-node spec of the deployed MongoDB CRD. There are four fields under it :
  - `spec.hidden.podTemplate` holds the hidden-node podSpec. `null` value of it, instructs kubedb operator to use  the default hidden-node podTemplate.
  - `spec.hidden.configSecret` is an optional field to provide custom configuration file for database (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise default configuration file will be used.
  - `spec.hidden.replicas` holds the number of hidden-node in the replica set.
  - `spec.hidden.storage` specifies the StorageClass of PVC dynamically allocated to store data for these hidden-nodes. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create some new PetSets : 1 for mongos, 1 for configServer, and 1 for each of the shard & hidden node. It creates a primary Service with the matching MongoDB object name. KubeDB operator will also create governing services for PetSets with the name `<mongodb-name>-<node-type>-pods`.

MongoDB `mongo-sh-hid` state,
All the types of nodes `Shard`, `ConfigServer` & `Mongos` are deployed as petset.

```bash
$ kubectl get mg,sts,svc,pvc,pv -n demo
NAME                              VERSION          STATUS   AGE
mongodb.kubedb.com/mongo-sh-hid   percona-7.0.4   Ready    4m46s

NAME                                          READY   AGE
petset.apps/mongo-sh-hid-configsvr       3/3     4m46s
petset.apps/mongo-sh-hid-mongos          2/2     2m52s
petset.apps/mongo-sh-hid-shard0          3/3     4m46s
petset.apps/mongo-sh-hid-shard0-hidden   2/2     3m45s
petset.apps/mongo-sh-hid-shard1          3/3     4m46s
petset.apps/mongo-sh-hid-shard1-hidden   2/2     3m36s

NAME                                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
service/mongo-sh-hid                  ClusterIP   10.96.57.155   <none>        27017/TCP   4m46s
service/mongo-sh-hid-configsvr-pods   ClusterIP   None           <none>        27017/TCP   4m46s
service/mongo-sh-hid-mongos-pods      ClusterIP   None           <none>        27017/TCP   4m46s
service/mongo-sh-hid-shard0-pods      ClusterIP   None           <none>        27017/TCP   4m46s
service/mongo-sh-hid-shard1-pods      ClusterIP   None           <none>        27017/TCP   4m46s

NAME                                                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mongo-sh-hid-shard0-hidden-0   Bound    pvc-9a4fd907-8225-4ed2-90e3-8ca43c0521d2   2Gi        RWO            standard       3m45s
persistentvolumeclaim/datadir-mongo-sh-hid-shard0-hidden-1   Bound    pvc-b77cd5d1-d5c1-433b-90dd-3784c5207cd6   2Gi        RWO            standard       3m23s
persistentvolumeclaim/datadir-mongo-sh-hid-shard1-hidden-0   Bound    pvc-61712454-2038-4692-a6ea-88685d7f34e1   2Gi        RWO            standard       3m36s
persistentvolumeclaim/datadir-mongo-sh-hid-shard1-hidden-1   Bound    pvc-489fb5c9-edee-4cf9-985f-48e04f14f695   2Gi        RWO            standard       3m14s

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                       STORAGECLASS   REASON   AGE
persistentvolume/pvc-489fb5c9-edee-4cf9-985f-48e04f14f695   2Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-hid-shard1-hidden-1   standard                3m11s
persistentvolume/pvc-61712454-2038-4692-a6ea-88685d7f34e1   2Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-hid-shard1-hidden-0   standard                3m33s
persistentvolume/pvc-9a4fd907-8225-4ed2-90e3-8ca43c0521d2   2Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-hid-shard0-hidden-0   standard                3m42s
persistentvolume/pvc-b77cd5d1-d5c1-433b-90dd-3784c5207cd6   2Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-hid-shard0-hidden-1   standard                3m20s

```


KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. It has also defaulted some field of crd object. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mongo-sh-hid -o yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-sh-hid","namespace":"demo"},"spec":{"hidden":{"podTemplate":{"spec":{"resources":{"requests":{"cpu":"400m","memory":"400Mi"}}}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"2Gi"}},"storageClassName":"standard"}},"shardTopology":{"configServer":{"ephemeralStorage":{},"replicas":3},"mongos":{"replicas":2},"shard":{"ephemeralStorage":{},"replicas":3,"shards":2}},"storageEngine":"inMemory","storageType":"Ephemeral","deletionPolicy":"WipeOut","version":"percona-7.0.4"}}
  creationTimestamp: "2022-10-31T05:59:43Z"
  finalizers:
    - kubedb.com
  generation: 3
  name: mongo-sh-hid
  namespace: demo
  resourceVersion: "721561"
  uid: 20f66240-669d-4556-b729-f6d0956a9241
spec:
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: mongo-sh-hid-auth
  autoOps: {}
  clusterAuthMode: keyFile
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  hidden:
    podTemplate:
      controller: {}
      metadata: {}
      spec:
        resources:
          limits:
            memory: 400Mi
          requests:
            cpu: 400m
            memory: 400Mi
    replicas: 2
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 2Gi
      storageClassName: standard
  keyFileSecret:
    name: mongo-sh-hid-key
  shardTopology:
    configServer:
      ephemeralStorage: {}
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
          serviceAccountName: mongo-sh-hid
      replicas: 3
    mongos:
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
          serviceAccountName: mongo-sh-hid
      replicas: 2
    shard:
      ephemeralStorage: {}
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
          serviceAccountName: mongo-sh-hid
      replicas: 3
      shards: 2
  sslMode: disabled
  storageEngine: inMemory
  storageType: Ephemeral
  deletionPolicy: WipeOut
  version: percona-7.0.4
status:
  conditions:
    - lastTransitionTime: "2022-10-31T05:59:43Z"
      message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mongo-sh-hid'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2022-10-31T06:02:05Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2022-10-31T06:01:47Z"
      message: 'The MongoDB: demo/mongo-sh-hid is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2022-10-31T06:01:47Z"
      message: 'The MongoDB: demo/mongo-sh-hid is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2022-10-31T06:02:05Z"
      message: 'The MongoDB: demo/mongo-sh-hid is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 3
  phase: Ready

```

Please note that KubeDB operator has created a new Secret called `mongo-sh-hid-auth` _(format: {mongodb-object-name}-auth)_ for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the _username_ for MongoDB superuser and a `password` key which contains the _password_ for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

## Connection Information

- Hostname/address: you can use any of these
  - Service: `mongo-sh-hid.demo`
  - Pod IP: (`$ kubectl get po -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-hid-mongos -o yaml | grep podIP`)
- Port: `27017`
- Username: Run following command to get _username_,

  ```bash
  $ kubectl get secrets -n demo mongo-sh-hid-auth -o jsonpath='{.data.\username}' | base64 -d
  root
  ```

- Password: Run the following command to get _password_,

  ```bash
  $ kubectl get secrets -n demo mongo-sh-hid-auth -o jsonpath='{.data.\password}' | base64 -d
  6&UiN5;qq)Tnai=7
  ```

Now, you can connect to this database through [mongo-shell](https://docs.mongodb.com/v4.2/mongo/).

## Sharded Data

In this tutorial, we will insert sharded and unsharded document, and we will see if the data actually sharded across cluster or not.

```bash
$ kubectl get pod -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-hid-mongos
NAME                    READY   STATUS    RESTARTS   AGE
mongo-sh-hid-mongos-0   1/1     Running   0          6m38s
mongo-sh-hid-mongos-1   1/1     Running   0          6m20s

$ kubectl exec -it mongo-sh-hid-mongos-0 -n demo bash

mongodb@mongo-sh-mongos-0:/$ mongo admin -u root -p '6&UiN5;qq)Tnai=7'
Percona Server for MongoDB shell version v7.0.4-11
connecting to: mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("e6979884-81b0-41c9-9745-50654f6fb39b") }
Percona Server for MongoDB server version: v7.0.4-11
Welcome to the Percona Server for MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://www.percona.com/doc/percona-server-for-mongodb
Questions? Try the support group
	https://www.percona.com/forums/questions-discussions/percona-server-for-mongodb
mongos> 
```

To detect if the MongoDB instance that your client is connected to is mongos, use the isMaster command. When a client connects to a mongos, isMaster returns a document with a `msg` field that holds the string `isdbgrid`.

```bash
mongos> rs.isMaster()
{
	"ismaster" : true,
	"msg" : "isdbgrid",
	"maxBsonObjectSize" : 16777216,
	"maxMessageSizeBytes" : 48000000,
	"maxWriteBatchSize" : 100000,
	"localTime" : ISODate("2022-10-31T06:11:39.882Z"),
	"logicalSessionTimeoutMinutes" : 30,
	"connectionId" : 310,
	"maxWireVersion" : 9,
	"minWireVersion" : 0,
	"topologyVersion" : {
		"processId" : ObjectId("635f64d1716935915500369b"),
		"counter" : NumberLong(0)
	},
	"ok" : 1,
	"operationTime" : Timestamp(1667196696, 31),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1667196696, 31),
		"signature" : {
			"hash" : BinData(0,"q30+hpYp5vn4t5HCvUiw1LDfbTg="),
			"keyId" : NumberLong("7160552274547179543")
		}
	}
}
```

`mongo-sh-hid` Shard status,

```bash
mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("635f645bf391eaa4fdef2fba")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-hid-shard0-0.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard0-1.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard0-2.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1,  "tags" : [ "shard0" ] }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-hid-shard1-0.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard1-1.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard1-2.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1,  "tags" : [ "shard1" ] }
  active mongoses:
        "7.0.4-11" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                407 : Success
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	618
                                shard1	406
                        too many chunks to print, use verbose if you want to force print
        {  "_id" : "kubedb-system",  "primary" : "shard0",  "partitioned" : true,  "version" : {  "uuid" : UUID("987e328c-5675-49d7-81a1-25d99142cad1"),  "lastMod" : 1 } }
                kubedb-system.health-check
                        shard key: { "id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	3
                                shard1	1
                        { "id" : { "$minKey" : 1 } } -->> { "id" : 0 } on : shard0 Timestamp(2, 1) 
                        { "id" : 0 } -->> { "id" : 1 } on : shard0 Timestamp(1, 2) 
                        { "id" : 1 } -->> { "id" : 2 } on : shard1 Timestamp(2, 0) 
                        { "id" : 2 } -->> { "id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 4) 
                         tag: shard0  { "id" : 0 } -->> { "id" : 1 }
                         tag: shard1  { "id" : 1 } -->> { "id" : 2 }
```






As `sh.status()` command only shows the general members, if we want to assure that hidden-nodes have been added correctly we need to exec into any shard-pod & run `rs.conf()` command against the admin database. Open another terminal : 


```bash
kubectl exec -it -n demo pod/mongo-sh-hid-shard1-0 -- bash

root@mongo-sh-hid-shard0-1:/ mongo admin -u root -p '6&UiN5;qq)Tnai=7'
Defaulted container "mongodb" out of: mongodb, copy-config (init)
Percona Server for MongoDB shell version v7.0.4-11
connecting to: mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("86dadf16-fff2-4483-b3ee-1ca7fc94229f") }
Percona Server for MongoDB server version: v7.0.4-11
Welcome to the Percona Server for MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://www.percona.com/doc/percona-server-for-mongodb
Questions? Try the support group
	https://www.percona.com/forums/questions-discussions/percona-server-for-mongodb

shard1:PRIMARY> rs.conf()
{
	"_id" : "shard1",
	"version" : 6,
	"term" : 1,
	"protocolVersion" : NumberLong(1),
	"writeConcernMajorityJournalDefault" : false,
	"members" : [
		{
			"_id" : 0,
			"host" : "mongo-sh-hid-shard1-0.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",
			"arbiterOnly" : false,
			"buildIndexes" : true,
			"hidden" : false,
			"priority" : 1,
			"tags" : {
				
			},
			"slaveDelay" : NumberLong(0),
			"votes" : 1
		},
		{
			"_id" : 1,
			"host" : "mongo-sh-hid-shard1-1.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",
			"arbiterOnly" : false,
			"buildIndexes" : true,
			"hidden" : false,
			"priority" : 1,
			"tags" : {
				
			},
			"slaveDelay" : NumberLong(0),
			"votes" : 1
		},
		{
			"_id" : 2,
			"host" : "mongo-sh-hid-shard1-2.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",
			"arbiterOnly" : false,
			"buildIndexes" : true,
			"hidden" : false,
			"priority" : 1,
			"tags" : {
				
			},
			"slaveDelay" : NumberLong(0),
			"votes" : 1
		},
		{
			"_id" : 3,
			"host" : "mongo-sh-hid-shard1-hidden-0.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",
			"arbiterOnly" : false,
			"buildIndexes" : true,
			"hidden" : true,
			"priority" : 0,
			"tags" : {
				
			},
			"slaveDelay" : NumberLong(0),
			"votes" : 1
		},
		{
			"_id" : 4,
			"host" : "mongo-sh-hid-shard1-hidden-1.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",
			"arbiterOnly" : false,
			"buildIndexes" : true,
			"hidden" : true,
			"priority" : 0,
			"tags" : {
				
			},
			"slaveDelay" : NumberLong(0),
			"votes" : 1
		}
	],
	"settings" : {
		"chainingAllowed" : true,
		"heartbeatIntervalMillis" : 2000,
		"heartbeatTimeoutSecs" : 10,
		"electionTimeoutMillis" : 10000,
		"catchUpTimeoutMillis" : -1,
		"catchUpTakeoverDelayMillis" : 30000,
		"getLastErrorModes" : {
			
		},
		"getLastErrorDefaults" : {
			"w" : 1,
			"wtimeout" : 0
		},
		"replicaSetId" : ObjectId("635f645c4883e315f55b07b4")
	}
}
```

Enable sharding to collection `songs.list` and insert document. See [`sh.shardCollection(namespace, key, unique, options)`](https://docs.mongodb.com/manual/reference/method/sh.shardCollection/#sh.shardCollection) for details about `shardCollection` command.

```bash
mongos> sh.enableSharding("songs");
{
	"ok" : 1,
	"operationTime" : Timestamp(1667197117, 5),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1667197117, 5),
		"signature" : {
			"hash" : BinData(0,"PqbGBYWJBwAexJoFMwUEQ1Z+ezc="),
			"keyId" : NumberLong("7160552274547179543")
		}
	}
}


mongos> sh.shardCollection("songs.list", {"myfield": 1});
{
	"collectionsharded" : "songs.list",
	"collectionUUID" : UUID("ed9c0fec-d488-4a2f-b5ce-8b244676a5b4"),
	"ok" : 1,
	"operationTime" : Timestamp(1667197139, 14),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1667197139, 14),
		"signature" : {
			"hash" : BinData(0,"Cons7FRJzPPeysmanMLyNgJlwNk="),
			"keyId" : NumberLong("7160552274547179543")
		}
	}
}

mongos> use songs
switched to db songs

mongos> db.list.insert({"led zeppelin": "stairway to heaven", "slipknot": "psychosocial"});
WriteResult({ "nInserted" : 1 })

mongos> db.list.insert({"pink floyd": "us and them", "nirvana": "smells like teen spirit", "john lennon" : "imagine" });
WriteResult({ "nInserted" : 1 })

mongos> db.list.find()
{ "_id" : ObjectId("635f68e774b0bd92060ebeb6"), "led zeppelin" : "stairway to heaven", "slipknot" : "psychosocial" }
{ "_id" : ObjectId("635f692074b0bd92060ebeb7"), "pink floyd" : "us and them", "nirvana" : "smells like teen spirit", "john lennon" : "imagine" }

```

Run [`sh.status()`](https://docs.mongodb.com/manual/reference/method/sh.status/) to see whether the `songs` database has sharding enabled, and the primary shard for the `songs` database.

The Sharded Collection section `sh.status.databases.<collection>` provides information on the sharding details for sharded collection(s) (E.g. `songs.list`). For each sharded collection, the section displays the shard key, the number of chunks per shard(s), the distribution of documents across chunks, and the tag information, if any, for shard key range(s).

```bash
mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("635f645bf391eaa4fdef2fba")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-hid-shard0-0.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard0-1.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard0-2.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1,  "tags" : [ "shard0" ] }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-hid-shard1-0.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard1-1.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard1-2.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1,  "tags" : [ "shard1" ] }
  active mongoses:
        "7.0.4-11" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                513 : Success
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	512
                                shard1	512
                        too many chunks to print, use verbose if you want to force print
        {  "_id" : "kubedb-system",  "primary" : "shard0",  "partitioned" : true,  "version" : {  "uuid" : UUID("987e328c-5675-49d7-81a1-25d99142cad1"),  "lastMod" : 1 } }
                kubedb-system.health-check
                        shard key: { "id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	3
                                shard1	1
                        { "id" : { "$minKey" : 1 } } -->> { "id" : 0 } on : shard0 Timestamp(2, 1) 
                        { "id" : 0 } -->> { "id" : 1 } on : shard0 Timestamp(1, 2) 
                        { "id" : 1 } -->> { "id" : 2 } on : shard1 Timestamp(2, 0) 
                        { "id" : 2 } -->> { "id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 4) 
                         tag: shard0  { "id" : 0 } -->> { "id" : 1 }
                         tag: shard1  { "id" : 1 } -->> { "id" : 2 }
        {  "_id" : "songs",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("03c7f9c8-f30f-42a4-8505-7f58fb95d3f3"),  "lastMod" : 1 } }
                songs.list
                        shard key: { "myfield" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard1	1
                        { "myfield" : { "$minKey" : 1 } } -->> { "myfield" : { "$maxKey" : 1 } } on : shard1 Timestamp(1, 0) 

```

Now create another database where partiotioned is not applied and see how the data is stored.

```bash
mongos> use demo
switched to db demo

mongos> db.anothercollection.insert({"myfield": "ccc", "otherfield": "this is non sharded", "kube" : "db" });
WriteResult({ "nInserted" : 1 })

mongos> db.anothercollection.insert({"myfield": "aaa", "more": "field" });
WriteResult({ "nInserted" : 1 })


mongos> db.anothercollection.find()
{ "_id" : ObjectId("635f69c674b0bd92060ebeb8"), "myfield" : "ccc", "otherfield" : "this is non sharded", "kube" : "db" }
{ "_id" : ObjectId("635f69d574b0bd92060ebeb9"), "myfield" : "aaa", "more" : "field" }
```

Now, eventually `sh.status()`

```
mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("635f645bf391eaa4fdef2fba")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-hid-shard0-0.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard0-1.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard0-2.mongo-sh-hid-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1,  "tags" : [ "shard0" ] }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-hid-shard1-0.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard1-1.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-hid-shard1-2.mongo-sh-hid-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1,  "tags" : [ "shard1" ] }
  active mongoses:
        "7.0.4-11" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                513 : Success
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	512
                                shard1	512
                        too many chunks to print, use verbose if you want to force print
        {  "_id" : "demo",  "primary" : "shard1",  "partitioned" : false,  "version" : {  "uuid" : UUID("040d4dc2-232f-4cc8-bae0-11c79244a9a7"),  "lastMod" : 1 } }
        {  "_id" : "kubedb-system",  "primary" : "shard0",  "partitioned" : true,  "version" : {  "uuid" : UUID("987e328c-5675-49d7-81a1-25d99142cad1"),  "lastMod" : 1 } }
                kubedb-system.health-check
                        shard key: { "id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	3
                                shard1	1
                        { "id" : { "$minKey" : 1 } } -->> { "id" : 0 } on : shard0 Timestamp(2, 1) 
                        { "id" : 0 } -->> { "id" : 1 } on : shard0 Timestamp(1, 2) 
                        { "id" : 1 } -->> { "id" : 2 } on : shard1 Timestamp(2, 0) 
                        { "id" : 2 } -->> { "id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 4) 
                         tag: shard0  { "id" : 0 } -->> { "id" : 1 }
                         tag: shard1  { "id" : 1 } -->> { "id" : 2 }
        {  "_id" : "songs",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("03c7f9c8-f30f-42a4-8505-7f58fb95d3f3"),  "lastMod" : 1 } }
                songs.list
                        shard key: { "myfield" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard1	1
                        { "myfield" : { "$minKey" : 1 } } -->> { "myfield" : { "$maxKey" : 1 } } on : shard1 Timestamp(1, 0) 

```

Here, `demo` database is not partitioned and all collections under `demo` database are stored in it's primary shard, which is `shard0`.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo mg/mongo-sh-hid
kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/stash/overview/index.md) process of MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
