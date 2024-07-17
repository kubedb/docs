---
title: MongoDB ReplicaSet with Hidden-node
menu:
  docs_{{ .version }}:
    identifier: mg-hidden-replicaset
    name: ReplicaSet with Hidden node
    parent: mg-hidden
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MongoDB ReplicaSet with Hidden-node

This tutorial will show you how to use KubeDB to run a MongoDB ReplicaSet with hidden-node.

## Before You Begin

Before proceeding:

- Read [mongodb hidden-node concept](/docs/guides/mongodb/hidden-node/concept.md) to get the concept about MongoDB Hidden node.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MongoDB ReplicaSet with Hidden-node

To deploy a MongoDB ReplicaSet, user have to specify `spec.replicaSet` option in `Mongodb` CRD.

The following is an example of a `Mongodb` object which creates MongoDB ReplicaSet of three members.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongo-rs-hid
  namespace: demo
spec:
  version: "percona-7.0.4"
  replicaSet:
    name: "replicaset"
  podTemplate:
    spec:
      containers:
      - name: mongo
        resources:
          requests:
            cpu: "600m"
            memory: "600Mi"
  replicas: 3
  storageEngine: inMemory
  storageType: Ephemeral
  ephemeralStorage:
    sizeLimit: "900Mi"
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
> Note: inMemory databases are only allowed for Percona variations of mongodb

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/hidden-node/replicaset.yaml
mongodb.kubedb.com/mongo-rs-hid created
```

Here,

- `spec.replicaSet` represents the configuration for replicaset.
  - `name` denotes the name of mongodb replicaset.
- `spec.keyFileSecret` (optional) is a secret name that contains keyfile (a random string)against `key.txt` key. Each mongod instances in the replica set and `shardTopology` uses the contents of the keyfile as the shared password for authenticating other members in the replicaset. Only mongod instances with the correct keyfile can join the replica set. _User can provide the `keyFileSecret` by creating a secret with key `key.txt`. See [here](https://docs.mongodb.com/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/#create-a-keyfile) to create the string for `keyFileSecret`._ If `keyFileSecret` is not given, KubeDB operator will generate a `keyFileSecret` itself.
- `spec.replicas` denotes the number of general members in `rs0` mongodb replicaset.
- `spec.podTemplate` denotes specifications of all the 3 general replicaset members.
- `spec.storageEngine` is set to inMemory, & `spec.storageType` to ephemeral.
- `spec.ephemeralStorage` holds the emptyDir volume specifications. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this ephemeral storage configuration.
- `spec.hidden` denotes hidden-node spec of the deployed MongoDB CRD. There are four fields under it : 
  - `spec.hidden.podTemplate` holds the hidden-node podSpec. `null` value of it, instructs kubedb operator to use  the default hidden-node podTemplate.
  - `spec.hidden.configSecret` is an optional field to provide custom configuration file for database (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise default configuration file will be used. 
  - `spec.hidden.replicas` holds the number of hidden-node in the replica set.
  - `spec.hidden.storage` specifies the StorageClass of PVC dynamically allocated to store data for these hidden-nodes. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.


KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create two new PetSets (one for replicas & one for hidden-nodes) and a Service with the matching MongoDB object name. This service will always point to the primary of the replicaset. KubeDB operator will also create a governing service for the pods of those two PetSets with the name `<mongodb-name>-pods`.

```bash
$ kubectl dba describe mg -n demo mongo-rs-hid
Name:               mongo-rs-hid
Namespace:          demo
CreationTimestamp:  Mon, 31 Oct 2022 11:03:50 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-rs-hid","namespace":"demo"},"spec":{"ephemeralStorage":...
Replicas:           3  total
Status:             Ready
StorageType:        Ephemeral
No volumes.
Paused:              false
Halted:              false
Termination Policy:  WipeOut

PetSet:          
  Name:               mongo-rs-hid
  CreationTimestamp:  Mon, 31 Oct 2022 11:03:50 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mongo-rs-hid
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
                        mongodb.kubedb.com/node.type=replica
  Annotations:        <none>
  Replicas:           824644499032 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

PetSet:          
  Name:               mongo-rs-hid-hidden
  CreationTimestamp:  Mon, 31 Oct 2022 11:04:50 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mongo-rs-hid
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
                        mongodb.kubedb.com/node.type=hidden
  Annotations:        <none>
  Replicas:           824646223576 desired | 2 total
  Pods Status:        2 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mongo-rs-hid
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mongo-rs-hid
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.197.33
  Port:         primary  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.53:27017

Service:        
  Name:         mongo-rs-hid-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mongo-rs-hid
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.53:27017,10.244.0.54:27017,10.244.0.55:27017 + 2 more...

Auth Secret:
  Name:         mongo-rs-hid-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mongo-rs-hid
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-rs-hid","namespace":"demo"},"spec":{"ephemeralStorage":{"sizeLimit":"900Mi"},"hidden":{"podTemplate":{"spec":{"resources":{"requests":{"cpu":"400m","memory":"400Mi"}}}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"2Gi"}},"storageClassName":"standard"}},"podTemplate":{"spec":{"resources":{"requests":{"cpu":"600m","memory":"600Mi"}}}},"replicaSet":{"name":"replicaset"},"replicas":3,"storageEngine":"inMemory","storageType":"Ephemeral","deletionPolicy":"WipeOut","version":"percona-7.0.4"}}

    Creation Timestamp:  2022-10-31T05:05:38Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mongo-rs-hid
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mongodbs.kubedb.com
    Name:                            mongo-rs-hid
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mongo-rs-hid
        Port:    27017
        Scheme:  mongodb
    Parameters:
      API Version:  config.kubedb.com/v1alpha1
      Kind:         MongoConfiguration
      Replica Sets:
        host-0:  replicaset/mongo-rs-hid-0.mongo-rs-hid-pods.demo.svc:27017,mongo-rs-hid-1.mongo-rs-hid-pods.demo.svc:27017,mongo-rs-hid-2.mongo-rs-hid-pods.demo.svc:27017,mongo-rs-hid-hidden-0.mongo-rs-hid-pods.demo.svc:27017,mongo-rs-hid-hidden-1.mongo-rs-hid-pods.demo.svc:27017
      Stash:
        Addon:
          Backup Task:
            Name:  mongodb-backup-4.4.6
          Restore Task:
            Name:  mongodb-restore-4.4.6
    Secret:
      Name:   mongo-rs-hid-auth
    Type:     kubedb.com/mongodb
    Version:  7.0.4

Events:
  Type    Reason        Age   From              Message
  ----    ------        ----  ----              -------
  Normal  PhaseChanged  12m   MongoDB operator  Phase changed from  to Provisioning.
  Normal  Successful    12m   MongoDB operator  Successfully created governing service
  Normal  Successful    12m   MongoDB operator  Successfully created Primary Service
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    11m   MongoDB operator  Successfully patched MongoDB
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid-hidden
  Normal  Successful    11m   MongoDB operator  Successfully patched MongoDB
  Normal  Successful    11m   MongoDB operator  Successfully created appbinding
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid-hidden
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  PhaseChanged  11m   MongoDB operator  Phase changed from Provisioning to Ready.
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    11m   MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid-hidden
  Normal  Successful    11m   MongoDB operator  Successfully patched MongoDB
  Normal  Successful    7m    MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    7m    MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid-hidden
  Normal  Successful    7m    MongoDB operator  Successfully patched MongoDB
  Normal  Successful    7m    MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    7m    MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid-hidden
  Normal  Successful    7m    MongoDB operator  Successfully patched MongoDB
  Normal  Successful    7m    MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid
  Normal  Successful    7m    MongoDB operator  Successfully patched PetSet demo/mongo-rs-hid-hidden
  Normal  Successful    7m    MongoDB operator  Successfully patched MongoDB



$ kubectl get petset -n demo
NAME                  READY   AGE
mongo-rs-hid          3/3     13m
mongo-rs-hid-hidden   2/2     12m


$ kubectl get pvc -n demo
NAME                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
datadir-mongo-rs-hid-hidden-0   Bound    pvc-e8c2a3b3-0c47-453f-8a5a-40d7dcb5b4d7   2Gi        RWO            standard       13m
datadir-mongo-rs-hid-hidden-1   Bound    pvc-7b752799-b6b9-43cf-9aa7-d39a2577216c   2Gi        RWO            standard       13m


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                STORAGECLASS   REASON   AGE
pvc-7b752799-b6b9-43cf-9aa7-d39a2577216c   2Gi        RWO            Delete           Bound    demo/datadir-mongo-rs-hid-hidden-1   standard                13m
pvc-e8c2a3b3-0c47-453f-8a5a-40d7dcb5b4d7   2Gi        RWO            Delete           Bound    demo/datadir-mongo-rs-hid-hidden-0   standard                13m


$ kubectl get service -n demo
NAME                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mongo-rs-hid        ClusterIP   10.96.197.33   <none>        27017/TCP   14m
mongo-rs-hid-pods   ClusterIP   None           <none>        27017/TCP   14m
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mongo-rs-hid -o yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-rs-hid","namespace":"demo"},"spec":{"ephemeralStorage":{"sizeLimit":"900Mi"},"hidden":{"podTemplate":{"spec":{"resources":{"requests":{"cpu":"400m","memory":"400Mi"}}}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"2Gi"}},"storageClassName":"standard"}},"podTemplate":{"spec":{"resources":{"requests":{"cpu":"600m","memory":"600Mi"}}}},"replicaSet":{"name":"replicaset"},"replicas":3,"storageEngine":"inMemory","storageType":"Ephemeral","deletionPolicy":"WipeOut","version":"percona-7.0.4"}}
  creationTimestamp: "2022-10-31T05:03:50Z"
  finalizers:
    - kubedb.com
  generation: 3
  name: mongo-rs-hid
  namespace: demo
  resourceVersion: "716264"
  uid: 428fa2bd-db5a-4bf5-a4ad-174fd0d7ade2
spec:
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: mongo-rs-hid-auth
  autoOps: {}
  clusterAuthMode: keyFile
  ephemeralStorage:
    sizeLimit: 900Mi
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  hidden:
    podTemplate:
      controller: {}
      metadata: {}
      spec:
        affinity:
          podAntiAffinity:
            preferredDuringSchedulingIgnoredDuringExecution:
              - podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      app.kubernetes.io/instance: mongo-rs-hid
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                  namespaces:
                    - demo
                  topologyKey: kubernetes.io/hostname
                weight: 100
              - podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      app.kubernetes.io/instance: mongo-rs-hid
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                  namespaces:
                    - demo
                  topologyKey: failure-domain.beta.kubernetes.io/zone
                weight: 50
        livenessProbe:
          exec:
            command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
              --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
              --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then \n
              \         exit 0\n        fi\n        exit 1"
          failureThreshold: 3
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 5
        readinessProbe:
          exec:
            command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
              --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
              --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then \n
              \         exit 0\n        fi\n        exit 1"
          failureThreshold: 3
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
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
    name: mongo-rs-hid-key
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: mongo-rs-hid
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                namespaces:
                  - demo
                topologyKey: kubernetes.io/hostname
              weight: 100
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: mongo-rs-hid
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                namespaces:
                  - demo
                topologyKey: failure-domain.beta.kubernetes.io/zone
              weight: 50
      livenessProbe:
        exec:
          command:
            - bash
            - -c
            - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
            --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
            --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then \n
            \         exit 0\n        fi\n        exit 1"
        failureThreshold: 3
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 5
      readinessProbe:
        exec:
          command:
            - bash
            - -c
            - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
            --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
            --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then \n
            \         exit 0\n        fi\n        exit 1"
        failureThreshold: 3
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 1
      resources:
        limits:
          memory: 600Mi
        requests:
          cpu: 600m
          memory: 600Mi
      serviceAccountName: mongo-rs-hid
  replicaSet:
    name: replicaset
  replicas: 3
  sslMode: disabled
  storageEngine: inMemory
  storageType: Ephemeral
  deletionPolicy: WipeOut
  version: percona-7.0.4
status:
  conditions:
    - lastTransitionTime: "2022-10-31T05:03:50Z"
      message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mongo-rs-hid'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2022-10-31T05:05:38Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2022-10-31T05:05:00Z"
      message: 'The MongoDB: demo/mongo-rs-hid is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2022-10-31T05:05:00Z"
      message: 'The MongoDB: demo/mongo-rs-hid is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2022-10-31T05:05:38Z"
      message: 'The MongoDB: demo/mongo-rs-hid is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 3
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `mongo-rs-hid-auth` *(format: {mongodb-object-name}-auth)* for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the *username* for MongoDB superuser and a `password` key which contains the *password* for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

## Redundancy and Data Availability

Now, you can connect to this database through [mongo-rs-hid](https://docs.mongodb.com/v3.4/mongo/). In this tutorial, we will insert document on the primary member, and we will see if the data becomes available on secondary members.

At first, insert data inside primary member `rs0:PRIMARY`.

```bash
$ kubectl get secrets -n demo mongo-rs-hid-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mongo-rs-hid-auth -o jsonpath='{.data.\password}' | base64 -d
OX4yb!IFm;~yAHkD

$ kubectl exec -it mongo-rs-hid-0 -n demo bash

bash-4.4$ mongo admin -u root -p 'OX4yb!IFm;~yAHkD'
Percona Server for MongoDB shell version v7.0.4-11
connecting to: mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("11890d64-37da-43dd-acb6-0f36a3678875") }
Percona Server for MongoDB server version: v7.0.4-11
Welcome to the Percona Server for MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://www.percona.com/doc/percona-server-for-mongodb
Questions? Try the support group
	https://www.percona.com/forums/questions-discussions/percona-server-for-mongodb
replicaset:PRIMARY> 
replicaset:PRIMARY> 
replicaset:PRIMARY> rs.status()
{
	"set" : "replicaset",
	"date" : ISODate("2022-10-31T05:25:19.148Z"),
	"myState" : 1,
	"term" : NumberLong(1),
	"syncSourceHost" : "",
	"syncSourceId" : -1,
	"heartbeatIntervalMillis" : NumberLong(2000),
	"majorityVoteCount" : 3,
	"writeMajorityCount" : 3,
	"votingMembersCount" : 5,
	"writableVotingMembersCount" : 5,
	"optimes" : {
		"lastCommittedOpTime" : {
			"ts" : Timestamp(1667193912, 1),
			"t" : NumberLong(1)
		},
		"lastCommittedWallTime" : ISODate("2022-10-31T05:25:12.590Z"),
		"readConcernMajorityOpTime" : {
			"ts" : Timestamp(1667193912, 1),
			"t" : NumberLong(1)
		},
		"readConcernMajorityWallTime" : ISODate("2022-10-31T05:25:12.590Z"),
		"appliedOpTime" : {
			"ts" : Timestamp(1667193912, 1),
			"t" : NumberLong(1)
		},
		"durableOpTime" : {
			"ts" : Timestamp(1667193912, 1),
			"t" : NumberLong(1)
		},
		"lastAppliedWallTime" : ISODate("2022-10-31T05:25:12.590Z"),
		"lastDurableWallTime" : ISODate("2022-10-31T05:25:12.590Z")
	},
	"lastStableRecoveryTimestamp" : Timestamp(1667193912, 1),
	"electionCandidateMetrics" : {
		"lastElectionReason" : "electionTimeout",
		"lastElectionDate" : ISODate("2022-10-31T05:04:02.548Z"),
		"electionTerm" : NumberLong(1),
		"lastCommittedOpTimeAtElection" : {
			"ts" : Timestamp(0, 0),
			"t" : NumberLong(-1)
		},
		"lastSeenOpTimeAtElection" : {
			"ts" : Timestamp(1667192642, 1),
			"t" : NumberLong(-1)
		},
		"numVotesNeeded" : 1,
		"priorityAtElection" : 1,
		"electionTimeoutMillis" : NumberLong(10000),
		"newTermStartDate" : ISODate("2022-10-31T05:04:02.552Z"),
		"wMajorityWriteAvailabilityDate" : ISODate("2022-10-31T05:04:02.553Z")
	},
	"members" : [
		{
			"_id" : 0,
			"name" : "mongo-rs-hid-0.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 1,
			"stateStr" : "PRIMARY",
			"uptime" : 1287,
			"optime" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-10-31T05:25:12Z"),
			"syncSourceHost" : "",
			"syncSourceId" : -1,
			"infoMessage" : "",
			"electionTime" : Timestamp(1667192642, 2),
			"electionDate" : ISODate("2022-10-31T05:04:02Z"),
			"configVersion" : 5,
			"configTerm" : 1,
			"self" : true,
			"lastHeartbeatMessage" : ""
		},
		{
			"_id" : 1,
			"name" : "mongo-rs-hid-1.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 1257,
			"optime" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-10-31T05:25:12Z"),
			"optimeDurableDate" : ISODate("2022-10-31T05:25:12Z"),
			"lastHeartbeat" : ISODate("2022-10-31T05:25:18.122Z"),
			"lastHeartbeatRecv" : ISODate("2022-10-31T05:25:18.120Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mongo-rs-hid-0.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"syncSourceId" : 0,
			"infoMessage" : "",
			"configVersion" : 5,
			"configTerm" : 1
		},
		{
			"_id" : 2,
			"name" : "mongo-rs-hid-2.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 1237,
			"optime" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-10-31T05:25:12Z"),
			"optimeDurableDate" : ISODate("2022-10-31T05:25:12Z"),
			"lastHeartbeat" : ISODate("2022-10-31T05:25:18.118Z"),
			"lastHeartbeatRecv" : ISODate("2022-10-31T05:25:18.119Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mongo-rs-hid-0.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"syncSourceId" : 0,
			"infoMessage" : "",
			"configVersion" : 5,
			"configTerm" : 1
		},
		{
			"_id" : 3,
			"name" : "mongo-rs-hid-hidden-0.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 1213,
			"optime" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-10-31T05:25:12Z"),
			"optimeDurableDate" : ISODate("2022-10-31T05:25:12Z"),
			"lastHeartbeat" : ISODate("2022-10-31T05:25:18.118Z"),
			"lastHeartbeatRecv" : ISODate("2022-10-31T05:25:18.119Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mongo-rs-hid-2.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"syncSourceId" : 2,
			"infoMessage" : "",
			"configVersion" : 5,
			"configTerm" : 1
		},
		{
			"_id" : 4,
			"name" : "mongo-rs-hid-hidden-1.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 1187,
			"optime" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1667193912, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-10-31T05:25:12Z"),
			"optimeDurableDate" : ISODate("2022-10-31T05:25:12Z"),
			"lastHeartbeat" : ISODate("2022-10-31T05:25:18.119Z"),
			"lastHeartbeatRecv" : ISODate("2022-10-31T05:25:18.008Z"),
			"pingMs" : NumberLong(0),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mongo-rs-hid-2.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
			"syncSourceId" : 2,
			"infoMessage" : "",
			"configVersion" : 5,
			"configTerm" : 1
		}
	],
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1667193912, 1),
		"signature" : {
			"hash" : BinData(0,"EPi/BjSrT3iN3lqSAFcCynAqlP0="),
			"keyId" : NumberLong("7160537873521836037")
		}
	},
	"operationTime" : Timestamp(1667193912, 1)
}
```

Here, Hidden-node's `statestr` is showing SECONDARY. If you want to see if they have been really added as hidden or not, you need to run `rs.conf()` command, look at the `hidden: true` specifications.

```shell
replicaset:PRIMARY> rs.conf()
{
	"_id" : "replicaset",
	"version" : 5,
	"term" : 1,
	"protocolVersion" : NumberLong(1),
	"writeConcernMajorityJournalDefault" : false,
	"members" : [
		{
			"_id" : 0,
			"host" : "mongo-rs-hid-0.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
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
			"host" : "mongo-rs-hid-1.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
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
			"host" : "mongo-rs-hid-2.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
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
			"host" : "mongo-rs-hid-hidden-0.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
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
			"host" : "mongo-rs-hid-hidden-1.mongo-rs-hid-pods.demo.svc.cluster.local:27017",
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
		"replicaSetId" : ObjectId("635f574270b72a363804832f")
```

```bash
replicaset:PRIMARY> rs.isMaster().primary
mongo-rs-hid-0.mongo-rs-hid-pods.demo.svc.cluster.local:27017

replicaset:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

replicaset:PRIMARY> use admin
switched to db admin
replicaset:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("5473e955-a97d-4c8f-a4fe-a82cbe7183f4"),
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



replicaset:PRIMARY> use mydb
switched to db mydb
replicaset:PRIMARY> db.songs.insert({"pink floyd": "shine on you crazy diamond"})
WriteResult({ "nInserted" : 1 })
replicaset:PRIMARY> db.songs.find().pretty()
{
	"_id" : ObjectId("635f5df01804db954f81276e"),
	"pink floyd" : "shine on you crazy diamond"
}

replicaset:PRIMARY> exit
bye
```

Now, check the redundancy and data availability in secondary members.
We will exec in `mongo-rs-hid-hidden-0`(which is a hidden node right now) to check the data availability.

```bash
$ kubectl exec -it mongo-rs-hid-hidden-0 -n demo bash
bash-4.4$ mongo admin -u root -p 'OX4yb!IFm;~yAHkD'
Percona Server for MongoDB server version: v7.0.4-11
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 7.0.4
Welcome to the MongoDB shell.

replicaset:SECONDARY> rs.slaveOk()
WARNING: slaveOk() is deprecated and may be removed in the next major release. Please use secondaryOk() instead.
replicaset:SECONDARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
mydb           0.000GB

replicaset:SECONDARY> use admin
switched to db admin
replicaset:SECONDARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("5473e955-a97d-4c8f-a4fe-a82cbe7183f4"),
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


replicaset:SECONDARY> use mydb
switched to db mydb

replicaset:SECONDARY> db.songs.find().pretty()
{
	"_id" : ObjectId("635f5df01804db954f81276e"),
	"pink floyd" : "shine on you crazy diamond"
}

rs0:SECONDARY> exit
bye

```

## Automatic Failover

To test automatic failover, we will force the primary member to restart. As the primary member (`pod`) becomes unavailable, the rest of the members will elect a primary member by election.

```bash
$ kubectl get pods -n demo
NAME                    READY   STATUS    RESTARTS   AGE
mongo-rs-hid-0          2/2     Running   0          34m
mongo-rs-hid-1          2/2     Running   0          33m
mongo-rs-hid-2          2/2     Running   0          33m
mongo-rs-hid-hidden-0   1/1     Running   0          33m
mongo-rs-hid-hidden-1   1/1     Running   0          32m

$ kubectl delete pod -n demo mongo-rs-hid-0
pod "mongo-rs-hid-0" deleted

$ kubectl get pods -n demo
NAME                    READY   STATUS        RESTARTS   AGE
mongo-rs-hid-0          2/2     Terminating   0          34m
mongo-rs-hid-1          2/2     Running       0          33m
mongo-rs-hid-2          2/2     Running       0          33m
mongo-rs-hid-hidden-0   1/1     Running       0          33m
mongo-rs-hid-hidden-1   1/1     Running       0          32m
```

Now verify the automatic failover, Let's exec in `mongo-rs-hid-0` pod,

```bash
$ kubectl exec -it mongo-rs-hid-0  -n demo bash
bash-4.4:/$ mongo admin -u root -p 'OX4yb!IFm;~yAHkD'
Percona Server for MongoDB server version: v7.0.4-11
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 7.0.4
Welcome to the MongoDB shell.

replicaset:SECONDARY> rs.isMaster().primary
mongo-rs-hid-1.mongo-rs-hid-pods.demo.svc.cluster.local:27017

# Also verify, data persistency
replicaset:SECONDARY> rs.slaveOk()
replicaset:SECONDARY> > show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
mydb           0.000GB

replicaset:SECONDARY> use mydb
switched to db mydb

replicaset:SECONDARY> db.songs.find().pretty()
{
	"_id" : ObjectId("635f5df01804db954f81276e"),
	"pink floyd" : "shine on you crazy diamond"
}
```
We could terminate the hidden-nodes also in a similar fashion, & check the automatic failover.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo mg/mongo-rs-hid
kubectl delete ns demo
```

## Next Steps

- Deploy MongoDB shard [with Hidden-node](/docs/guides/mongodb/hidden-node/sharding.md).
- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md) process of MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
