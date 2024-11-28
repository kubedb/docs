---
title: MongoDB Sharding Guide
menu:
  docs_{{ .version }}:
    identifier: mg-clustering-sharding
    name: Sharding Guide
    parent: mg-clustering-mongodb
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Sharding

This tutorial will show you how to use KubeDB to run a sharded MongoDB cluster.

## Before You Begin

Before proceeding:

- Read [mongodb sharding concept](/docs/guides/mongodb/clustering/sharding_concept.md) to learn about MongoDB Sharding clustering.

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

The following is an example of a `Mongodb` object which creates MongoDB Sharding of three members.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongo-sh
  namespace: demo
spec:
  version: 4.4.26
  shardTopology:
    configServer:
      replicas: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 3
      shards: 2
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/clustering/mongo-sharding.yaml
mongodb.kubedb.com/mongo-sh created
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

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create some new PetSets : 1 for mongos, 1 for configServer & 1 for each of the shards. It creates a primary Service with the matching MongoDB object name. KubeDB operator will also create governing services for PetSets with the name `<mongodb-name>-<node-type>-pods`.

MongoDB `mongo-sh` state,

```bash
$ kubectl get mg -n demo
NAME       VERSION   STATUS    AGE
mongo-sh   4.4.26     Ready     9m41s
```

All the types of nodes `Shard`, `ConfigServer` & `Mongos` are deployed as petset.

```bash
$ kubectl get petset -n demo
NAME                 READY   AGE
mongo-sh-configsvr   3/3     11m
mongo-sh-mongos      3/3     8m41s
mongo-sh-shard0      3/3     10m
mongo-sh-shard1      3/3     8m59s
```

All PVCs and PVs for MongoDB `mongo-sh`,

```bash
$ kubectl get pvc -n demo
NAME                           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
datadir-mongo-sh-configsvr-0   Bound    pvc-1db4185e-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       16m
datadir-mongo-sh-configsvr-1   Bound    pvc-330cc6ee-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       16m
datadir-mongo-sh-configsvr-2   Bound    pvc-3db2d3f5-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       15m
datadir-mongo-sh-shard0-0      Bound    pvc-49b7cc3b-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       15m
datadir-mongo-sh-shard0-1      Bound    pvc-5b781770-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       15m
datadir-mongo-sh-shard0-2      Bound    pvc-6ba3263e-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       14m
datadir-mongo-sh-shard1-0      Bound    pvc-75feb227-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       14m
datadir-mongo-sh-shard1-1      Bound    pvc-89bb7bb3-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       13m
datadir-mongo-sh-shard1-2      Bound    pvc-98c96ae4-6a5f-11e9-a871-080027a851ba   1Gi        RWO            standard       13m


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS   REASON   AGE
pvc-1db4185e-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-configsvr-0   standard                17m
pvc-330cc6ee-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-configsvr-1   standard                16m
pvc-3db2d3f5-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-configsvr-2   standard                16m
pvc-49b7cc3b-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-shard0-0      standard                16m
pvc-5b781770-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-shard0-1      standard                15m
pvc-6ba3263e-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-shard0-2      standard                15m
pvc-75feb227-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-shard1-0      standard                14m
pvc-89bb7bb3-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-shard1-1      standard                14m
pvc-98c96ae4-6a5f-11e9-a871-080027a851ba   1Gi        RWO            Delete           Bound    demo/datadir-mongo-sh-shard1-2      standard                13m
```

Services created for MongoDB `mongo-sh`

```bash
$ kubectl get svc -n demo
NAME                     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
mongo-sh                 ClusterIP   10.108.188.201   <none>        27017/TCP   18m
mongo-sh-configsvr-pods  ClusterIP   None             <none>        27017/TCP   18m
mongo-sh-mongos-pods     ClusterIP   None             <none>        27017/TCP   18m
mongo-sh-shard0-pods     ClusterIP   None             <none>        27017/TCP   18m
mongo-sh-shard1-pods     ClusterIP   None             <none>        27017/TCP   18m
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. It has also defaulted some field of crd object. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mongo-sh -o yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-sh","namespace":"demo"},"spec":{"shardTopology":{"configServer":{"replicas":3,"storage":{"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"}},"mongos":{"replicas":2},"shard":{"replicas":3,"shards":2,"storage":{"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"}}},"version":"4.4.26"}}
  creationTimestamp: "2021-02-10T12:57:03Z"
  finalizers:
    - kubedb.com
  generation: 3
  managedFields:
    - apiVersion: kubedb.com/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:kubectl.kubernetes.io/last-applied-configuration: {}
        f:spec:
          .: {}
          f:shardTopology:
            .: {}
            f:configServer:
              .: {}
              f:replicas: {}
              f:storage:
                .: {}
                f:resources:
                  .: {}
                  f:requests:
                    .: {}
                    f:storage: {}
                f:storageClassName: {}
            f:mongos:
              .: {}
              f:replicas: {}
            f:shard:
              .: {}
              f:replicas: {}
              f:shards: {}
              f:storage:
                .: {}
                f:resources:
                  .: {}
                  f:requests:
                    .: {}
                    f:storage: {}
                f:storageClassName: {}
          f:version: {}
      manager: kubectl-client-side-apply
      operation: Update
      time: "2021-02-10T12:57:03Z"
    - apiVersion: kubedb.com/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:finalizers: {}
        f:spec:
          f:authSecret:
            .: {}
            f:name: {}
          f:keyFileSecret:
            .: {}
            f:name: {}
        f:status:
          .: {}
          f:conditions: {}
          f:observedGeneration: {}
          f:phase: {}
      manager: mg-operator
      operation: Update
      time: "2021-02-10T12:57:03Z"
  name: mongo-sh
  namespace: demo
  resourceVersion: "152268"
  uid: 8522c8c1-344b-4824-9061-47031b88f1fa
spec:
  authSecret:
    name: mongo-sh-auth
  clusterAuthMode: keyFile
  keyFileSecret:
    name: mongo-sh-key
  shardTopology:
    configServer:
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
          serviceAccountName: mongo-sh
      replicas: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
          serviceAccountName: mongo-sh
      replicas: 2
    shard:
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
          serviceAccountName: mongo-sh
      replicas: 3
      shards: 2
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  sslMode: disabled
  storageEngine: wiredTiger
  storageType: Durable
  deletionPolicy: Delete
  version: 4.4.26
status:
  conditions:
    - lastTransitionTime: "2021-02-10T12:57:03Z"
      message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mongo-sh'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2021-02-10T13:09:44Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2021-02-10T12:59:33Z"
      message: 'The MongoDB: demo/mongo-sh is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2021-02-10T12:59:33Z"
      message: 'The MongoDB: demo/mongo-sh is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2021-02-10T12:59:51Z"
      message: 'The MongoDB: demo/mongo-sh is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 3
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `mongo-sh-auth` _(format: {mongodb-object-name}-auth)_ for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the _username_ for MongoDB superuser and a `password` key which contains the _password_ for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

## Connection Information

- Hostname/address: you can use any of these
  - Service: `mongo-sh.demo`
  - Pod IP: (`$ kubectl get po -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-mongos -o yaml | grep podIP`)
- Port: `27017`
- Username: Run following command to get _username_,

  ```bash
  $ kubectl get secrets -n demo mongo-sh-auth -o jsonpath='{.data.\username}' | base64 -d
  root
  ```

- Password: Run the following command to get _password_,

  ```bash
  $ kubectl get secrets -n demo mongo-sh-auth -o jsonpath='{.data.\password}' | base64 -d
  7QiqLcuSCmZ8PU5a
  ```

Now, you can connect to this database through [mongo-shell](https://docs.mongodb.com/v4.2/mongo/).

## Sharded Data

In this tutorial, we will insert sharded and unsharded document, and we will see if the data actually sharded across cluster or not.

```bash
$ kubectl get po -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-mongos
NAME                READY   STATUS    RESTARTS   AGE
mongo-sh-mongos-0   1/1     Running   0          49m
mongo-sh-mongos-1   1/1     Running   0          49m

$ kubectl exec -it mongo-sh-mongos-0 -n demo bash

mongodb@mongo-sh-mongos-0:/$ mongo admin -u root -p 7QiqLcuSCmZ8PU5a
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("8b7abf57-09e4-4e30-b4a0-a37ebf065e8f") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	http://docs.mongodb.org/
Questions? Try the support group
	http://groups.google.com/group/mongodb-user
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
	"localTime" : ISODate("2021-02-10T13:37:24.140Z"),
	"logicalSessionTimeoutMinutes" : 30,
	"connectionId" : 803,
	"maxWireVersion" : 8,
	"minWireVersion" : 0,
	"ok" : 1,
	"operationTime" : Timestamp(1612964237, 2),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1612964237, 2),
		"signature" : {
			"hash" : BinData(0,"5ugX3jIC+sVDtYjxGWP5SCI7QSE="),
			"keyId" : NumberLong("6927618399740624913")
		}
	}
}
```

`mongo-sh` Shard status,

```bash
mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("6023d83b8df2b687ecfade84")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-shard0-0.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-1.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-2.mongo-sh-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-shard1-0.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-1.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-2.mongo-sh-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                No recent migrations
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
```

Shard collection `test.testcoll` and insert document. See [`sh.shardCollection(namespace, key, unique, options)`](https://docs.mongodb.com/manual/reference/method/sh.shardCollection/#sh.shardCollection) for details about `shardCollection` command.

```bash
mongos> sh.enableSharding("test");
{
	"ok" : 1,
	"operationTime" : Timestamp(1612964293, 5),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1612964293, 5),
		"signature" : {
			"hash" : BinData(0,"DJbXhWUbiTQCWvlWgTTW/vlH3LE="),
			"keyId" : NumberLong("6927618399740624913")
		}
	}
}

mongos> sh.shardCollection("test.testcoll", {"myfield": 1});
{
	"collectionsharded" : "test.testcoll",
	"collectionUUID" : UUID("f2617eb1-8f61-47dd-af58-73f5fe4ea2c0"),
	"ok" : 1,
	"operationTime" : Timestamp(1612964314, 14),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1612964314, 14),
		"signature" : {
			"hash" : BinData(0,"CZzOATrFeADxMkGTWbX85Olkc2Q="),
			"keyId" : NumberLong("6927618399740624913")
		}
	}
}

mongos> use test;
switched to db test

mongos> db.testcoll.insert({"myfield": "a", "otherfield": "b"});
WriteResult({ "nInserted" : 1 })

mongos> db.testcoll.insert({"myfield": "c", "otherfield": "d", "kube" : "db" });
WriteResult({ "nInserted" : 1 })

mongos> db.testcoll.find();
{ "_id" : ObjectId("5cc6d6f656a9ddd30be2c12a"), "myfield" : "a", "otherfield" : "b" }
{ "_id" : ObjectId("5cc6d71e56a9ddd30be2c12b"), "myfield" : "c", "otherfield" : "d", "kube" : "db" }

```

Run [`sh.status()`](https://docs.mongodb.com/manual/reference/method/sh.status/) to see whether the `test` database has sharding enabled, and the primary shard for the `test` database.

The Sharded Collection section `sh.status.databases.<collection>` provides information on the sharding details for sharded collection(s) (E.g. `test.testcoll`). For each sharded collection, the section displays the shard key, the number of chunks per shard(s), the distribution of documents across chunks, and the tag information, if any, for shard key range(s).

```bash
mongos> sh.status();
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("6023d83b8df2b687ecfade84")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-shard0-0.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-1.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-2.mongo-sh-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-shard1-0.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-1.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-2.mongo-sh-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                No recent migrations
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
        {  "_id" : "test",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("2a39d8c7-c731-46af-84c3-bf04ba10ac82"),  "lastMod" : 1 } }
                test.testcoll
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

mongos> db.testcoll2.insert({"myfield": "ccc", "otherfield": "d", "kube" : "db" });
WriteResult({ "nInserted" : 1 })

mongos> db.testcoll2.insert({"myfield": "aaa", "otherfield": "d", "kube" : "db" });
WriteResult({ "nInserted" : 1 })


mongos> db.testcoll2.find()
{ "_id" : ObjectId("5cc6dc831b6d9b3cddc947ec"), "myfield" : "ccc", "otherfield" : "d", "kube" : "db" }
{ "_id" : ObjectId("5cc6dce71b6d9b3cddc947ed"), "myfield" : "aaa", "otherfield" : "d", "kube" : "db" }
```

Now, eventually `sh.status()`

```
mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("6023d83b8df2b687ecfade84")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-shard0-0.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-1.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-2.mongo-sh-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-shard1-0.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-1.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-2.mongo-sh-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                No recent migrations
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
        {  "_id" : "demo",  "primary" : "shard1",  "partitioned" : false,  "version" : {  "uuid" : UUID("93d077e0-2da0-4b68-a4d4-d23394b22ab2"),  "lastMod" : 1 } }
        {  "_id" : "test",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("2a39d8c7-c731-46af-84c3-bf04ba10ac82"),  "lastMod" : 1 } }
                test.testcoll
                        shard key: { "myfield" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard1	1
                        { "myfield" : { "$minKey" : 1 } } -->> { "myfield" : { "$maxKey" : 1 } } on : shard1 Timestamp(1, 0) 
```

Here, `demo` database is not partitioned and all collections under `demo` database are stored in it's primary shard, which is `shard1`.

## Halt Database

When [DeletionPolicy](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy) is set to halt, and you delete the mongodb object, the KubeDB operator will delete the PetSet and its pods but leaves the PVCs, secrets and database backup (snapshots) intact. Learn details of all `DeletionPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy).

You can also keep the mongodb object and halt the database to resume it again later. If you halt the database, the kubedb will delete the petsets and services but will keep the mongodb object, pvcs, secrets and backup (snapshots).

To halt the database, first you have to set the deletionPolicy to `Halt` in existing database. You can use the below command to set the deletionPolicy to `Halt`, if it is not already set.

```bash
$ kubectl patch -n demo mg/mongo-sh -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"
mongodb.kubedb.com/mongo-sh patched
```

Then, you have to set the `spec.halted` as true to set the database in a `Halted` state. You can use the below command.

```bash
$ kubectl patch -n demo mg/mongo-sh -p '{"spec":{"halted":true}}' --type="merge"
mongodb.kubedb.com/mongo-sh patched
```

After that, kubedb will delete the petsets and services and you can see the database Phase as `Halted`.

Now, you can run the following command to get all mongodb resources in demo namespaces,

```bash
$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME                          VERSION   STATUS   AGE
mongodb.kubedb.com/mongo-sh   4.4.26     Halted   74m

NAME                            TYPE                                  DATA   AGE
secret/default-token-x2zcl      kubernetes.io/service-account-token   3      32h
secret/mongo-sh-auth            Opaque                                2      75m
secret/mongo-sh-key             Opaque                                1      75m

NAME                                                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mongo-sh-configsvr-0   Bound    pvc-9d1b3c01-fdce-45ab-b6f6-fc7bf9462e89   1Gi        RWO            standard       74m
persistentvolumeclaim/datadir-mongo-sh-configsvr-1   Bound    pvc-8e14fcea-ec15-4614-9ec5-21fdf3eb477c   1Gi        RWO            standard       74m
persistentvolumeclaim/datadir-mongo-sh-configsvr-2   Bound    pvc-b65665ce-f35b-4c4f-a7ac-410ad2dfa82d   1Gi        RWO            standard       73m
persistentvolumeclaim/datadir-mongo-sh-shard0-0      Bound    pvc-8fbfdd01-1ed1-4e3b-9a2a-0aa75911cbf0   1Gi        RWO            standard       74m
persistentvolumeclaim/datadir-mongo-sh-shard0-1      Bound    pvc-71d2b22b-2168-46d3-927c-d3ac92f22ebb   1Gi        RWO            standard       74m
persistentvolumeclaim/datadir-mongo-sh-shard0-2      Bound    pvc-82f83359-6e31-43e4-88b3-2555cb442ca0   1Gi        RWO            standard       73m
persistentvolumeclaim/datadir-mongo-sh-shard1-0      Bound    pvc-07ef7cd3-99b2-47de-b1bb-ef6c5606d92e   1Gi        RWO            standard       74m
persistentvolumeclaim/datadir-mongo-sh-shard1-1      Bound    pvc-ffa4b9a7-2492-4f18-be90-7950004e9efd   1Gi        RWO            standard       74m
persistentvolumeclaim/datadir-mongo-sh-shard1-2      Bound    pvc-4e75b90e-dac5-4431-a50e-2bc8dfcf481b   1Gi        RWO            standard       73m
```

From the above output, you can see that MongoDB object, PVCs, Secret are still there.

## Resume Halted Database

Now, to resume the database, i.e. to get the same database setup back again, you have to set the the `spec.halted` as false. You can use the below command.

```bash
$ kubectl patch -n demo mg/mongo-sh -p '{"spec":{"halted":false}}' --type="merge"
mongodb.kubedb.com/mongo-sh patched
```

When the database is resumed successfully, you can see the database Status is set to `Ready`.

```bash
$ kubectl get mg -n demo
NAME       VERSION   STATUS    AGE
mongo-sh   4.4.26     Ready     6m27s
```

Now, If you again exec into `pod` and look for previous data, you will see that, all the data persists.

```bash
$ kubectl get po -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-mongos
NAME                READY   STATUS    RESTARTS   AGE
mongo-sh-mongos-0   1/1     Running   0          3m52s
mongo-sh-mongos-1   1/1     Running   0          3m52s


$ kubectl exec -it mongo-sh-mongos-0 -n demo bash

mongodb@mongo-sh-mongos-0:/$ mongo admin -u root -p 7QiqLcuSCmZ8PU5a

mongos> use test;
switched to db test

mongos> db.testcoll.find();
{ "_id" : ObjectId("5cc6d6f656a9ddd30be2c12a"), "myfield" : "a", "otherfield" : "b" }
{ "_id" : ObjectId("5cc6d71e56a9ddd30be2c12b"), "myfield" : "c", "otherfield" : "d", "kube" : "db" }

mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("6023d83b8df2b687ecfade84")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-shard0-0.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-1.mongo-sh-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-shard0-2.mongo-sh-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-shard1-0.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-1.mongo-sh-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-shard1-2.mongo-sh-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  1
        Last reported error:  Could not find host matching read preference { mode: "primary" } for set shard0
        Time of Reported error:  Wed Feb 10 2021 14:16:04 GMT+0000 (UTC)
        Migration Results for the last 24 hours: 
                No recent migrations
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
        {  "_id" : "demo",  "primary" : "shard1",  "partitioned" : false,  "version" : {  "uuid" : UUID("93d077e0-2da0-4b68-a4d4-d23394b22ab2"),  "lastMod" : 1 } }
        {  "_id" : "test",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("2a39d8c7-c731-46af-84c3-bf04ba10ac82"),  "lastMod" : 1 } }
                test.testcoll
                        shard key: { "myfield" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard1	1
                        { "myfield" : { "$minKey" : 1 } } -->> { "myfield" : { "$maxKey" : 1 } } on : shard1 Timestamp(1, 0)
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mongo-sh -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mongo-sh

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
