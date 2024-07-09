---
title: MongoDB ReplicaSet Guide
menu:
  docs_{{ .version }}:
    identifier: mg-clustering-replicaset
    name: ReplicaSet Guide
    parent: mg-clustering-mongodb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MongoDB ReplicaSet

This tutorial will show you how to use KubeDB to run a MongoDB ReplicaSet.

## Before You Begin

Before proceeding:

- Read [mongodb replication concept](/docs/guides/mongodb/clustering/replication_concept.md) to learn about MongoDB Replica Set clustering.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MongoDB ReplicaSet

To deploy a MongoDB ReplicaSet, user have to specify `spec.replicaSet` option in `Mongodb` CRD.

The following is an example of a `Mongodb` object which creates MongoDB ReplicaSet of three members.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mgo-replicaset
  namespace: demo
spec:
  version: "4.4.26"
  replicas: 3
  replicaSet:
    name: rs0
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/clustering/replicaset.yaml
mongodb.kubedb.com/mgo-replicaset created
```

Here,

- `spec.replicaSet` represents the configuration for replicaset.
  - `name` denotes the name of mongodb replicaset.
- `spec.keyFileSecret` (optional) is a secret name that contains keyfile (a random string)against `key.txt` key. Each mongod instances in the replica set and `shardTopology` uses the contents of the keyfile as the shared password for authenticating other members in the replicaset. Only mongod instances with the correct keyfile can join the replica set. _User can provide the `keyFileSecret` by creating a secret with key `key.txt`. See [here](https://docs.mongodb.com/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/#create-a-keyfile) to create the string for `keyFileSecret`._ If `keyFileSecret` is not given, KubeDB operator will generate a `keyFileSecret` itself.
- `spec.replicas` denotes the number of members in `rs0` mongodb replicaset.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create a new PetSet and a Service with the matching MongoDB object name. This service will always point to the primary of the replicaset. KubeDB operator will also create a governing service for PetSets with the name `<mongodb-name>-pods`.

```bash
$ kubectl dba describe mg -n demo mgo-replicaset
Name:               mgo-replicaset
Namespace:          demo
CreationTimestamp:  Wed, 10 Feb 2021 11:05:06 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-replicaset","namespace":"demo"},"spec":{"replicaSet":{"na...
Replicas:           3  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  Delete

PetSet:          
  Name:               mgo-replicaset
  CreationTimestamp:  Wed, 10 Feb 2021 11:05:06 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mgo-replicaset
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:        <none>
  Replicas:           824637635032 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mgo-replicaset
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-replicaset
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           fd00:10:96::d5f5
  Port:         primary  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    [fd00:10:244::a]:27017

Service:        
  Name:         mgo-replicaset-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-replicaset
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    [fd00:10:244::a]:27017,[fd00:10:244::c]:27017,[fd00:10:244::e]:27017

Auth Secret:
  Name:         mgo-replicaset-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-replicaset
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
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-replicaset","namespace":"demo"},"spec":{"replicaSet":{"name":"rs0"},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"4.4.26"}}

    Creation Timestamp:  2021-02-10T05:07:10Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mgo-replicaset
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mongodbs.kubedb.com
    Name:                            mgo-replicaset
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mgo-replicaset
        Port:    27017
        Scheme:  mongodb
    Parameters:
      API Version:  config.kubedb.com/v1alpha1
      Kind:         MongoConfiguration
      Replica Sets:
        host-0:  rs0/mgo-replicaset-0.mgo-replicaset-pods.demo.svc,mgo-replicaset-1.mgo-replicaset-pods.demo.svc,mgo-replicaset-2.mgo-replicaset-pods.demo.svc
    Secret:
      Name:   mgo-replicaset-auth
    Type:     kubedb.com/mongodb
    Version:  4.4.26

Events:
  Type    Reason      Age   From              Message
  ----    ------      ----  ----              -------
  Normal  Successful  12m   MongoDB operator  Successfully created stats service
  Normal  Successful  12m   MongoDB operator  Successfully created Service
  Normal  Successful  12m   MongoDB operator  Successfully  stats service
  Normal  Successful  12m   MongoDB operator  Successfully  stats service
  Normal  Successful  11m   MongoDB operator  Successfully  stats service
  Normal  Successful  11m   MongoDB operator  Successfully  stats service
  Normal  Successful  11m   MongoDB operator  Successfully  stats service
  Normal  Successful  11m   MongoDB operator  Successfully  stats service
  Normal  Successful  10m   MongoDB operator  Successfully  stats service
  Normal  Successful  10m   MongoDB operator  Successfully  stats service
  Normal  Successful  10m   MongoDB operator  Successfully  stats service
  Normal  Successful  10m   MongoDB operator  Successfully  stats service
  Normal  Successful  10m   MongoDB operator  Successfully  stats service
  Normal  Successful  10m   MongoDB operator  Successfully patched PetSet demo/mgo-replicaset
  Normal  Successful  10m   MongoDB operator  Successfully patched MongoDB
  Normal  Successful  10m   MongoDB operator  Successfully created appbinding
  Normal  Successful  10m   MongoDB operator  Successfully  stats service
  Normal  Successful  10m   MongoDB operator  Successfully patched PetSet demo/mgo-replicaset
  Normal  Successful  10m   MongoDB operator  Successfully patched MongoDB


$ kubectl get petset -n demo
NAME             READY   AGE
mgo-replicaset   3/3     105s

$ kubectl get pvc -n demo
NAME                       STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
datadir-mgo-replicaset-0   Bound     pvc-597784c9-c093-11e8-b4a9-0800272618ed   1Gi        RWO            standard       1h
datadir-mgo-replicaset-1   Bound     pvc-8ca7a9d9-c093-11e8-b4a9-0800272618ed   1Gi        RWO            standard       1h
datadir-mgo-replicaset-2   Bound     pvc-b7d8a624-c093-11e8-b4a9-0800272618ed   1Gi        RWO            standard       1h

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM                           STORAGECLASS   REASON    AGE
pvc-597784c9-c093-11e8-b4a9-0800272618ed   1Gi        RWO            Delete           Bound     demo/datadir-mgo-replicaset-0   standard                 1h
pvc-8ca7a9d9-c093-11e8-b4a9-0800272618ed   1Gi        RWO            Delete           Bound     demo/datadir-mgo-replicaset-1   standard                 1h
pvc-b7d8a624-c093-11e8-b4a9-0800272618ed   1Gi        RWO            Delete           Bound     demo/datadir-mgo-replicaset-2   standard                 1h

$ kubectl get service -n demo
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
mgo-replicaset       ClusterIP   10.97.174.220   <none>        27017/TCP   119s
mgo-replicaset-pods  ClusterIP   None            <none>        27017/TCP   119s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mgo-replicaset -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-replicaset","namespace":"demo"},"spec":{"replicaSet":{"name":"rs0"},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"4.4.26"}}
  creationTimestamp: "2021-02-11T04:29:29Z"
  finalizers:
    - kubedb.com
  generation: 3
  managedFields:
    - apiVersion: kubedb.com/v1alpha2
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:kubectl.kubernetes.io/last-applied-configuration: {}
        f:spec:
          .: {}
          f:replicaSet:
            .: {}
            f:name: {}
          f:replicas: {}
          f:storage:
            .: {}
            f:accessModes: {}
            f:resources:
              .: {}
              f:requests:
                .: {}
                f:storage: {}
            f:storageClassName: {}
          f:version: {}
      manager: kubectl-client-side-apply
      operation: Update
      time: "2021-02-11T04:29:29Z"
    - apiVersion: kubedb.com/v1alpha2
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
      time: "2021-02-11T04:29:29Z"
  name: mgo-replicaset
  namespace: demo
  resourceVersion: "191685"
  uid: 1cc92de5-441e-42ac-8321-459d7a955af2
spec:
  authSecret:
    name: mgo-replicaset-auth
  clusterAuthMode: keyFile
  keyFileSecret:
    name: mgo-replicaset-key
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
                    app.kubernetes.io/instance: mgo-replicaset
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                namespaces:
                  - demo
                topologyKey: kubernetes.io/hostname
              weight: 100
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: mgo-replicaset
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
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: mgo-replicaset
  replicaSet:
    name: rs0
  replicas: 3
  sslMode: disabled
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageEngine: wiredTiger
  storageType: Durable
  deletionPolicy: Delete
  version: 4.4.26
status:
  conditions:
    - lastTransitionTime: "2021-02-11T04:29:29Z"
      message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mgo-replicaset'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2021-02-11T04:31:22Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2021-02-11T04:31:11Z"
      message: 'The MongoDB: demo/mgo-replicaset is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2021-02-11T04:31:11Z"
      message: 'The MongoDB: demo/mgo-replicaset is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2021-02-11T04:31:22Z"
      message: 'The MongoDB: demo/mgo-replicaset is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 3
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `mgo-replicaset-auth` *(format: {mongodb-object-name}-auth)* for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the *username* for MongoDB superuser and a `password` key which contains the *password* for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

## Redundancy and Data Availability

Now, you can connect to this database through [mgo-replicaset](https://docs.mongodb.com/v3.4/mongo/). In this tutorial, we will insert document on the primary member, and we will see if the data becomes available on secondary members.

At first, insert data inside primary member `rs0:PRIMARY`.

```bash
$ kubectl get secrets -n demo mgo-replicaset-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mgo-replicaset-auth -o jsonpath='{.data.\password}' | base64 -d
5O4R2ze2bWXcWsdP

$ kubectl exec -it mgo-replicaset-0 -n demo bash

mongodb@mgo-replicaset-0:/$ mongo admin -u root -p 5O4R2ze2bWXcWsdP
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.

rs0:PRIMARY> > rs.isMaster().primary
mgo-replicaset-0.mgo-replicaset-gvr.demo.svc.cluster.local:27017

rs0:PRIMARY> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB

rs0:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("6b714456-2914-4ea0-9596-92249e8285a2"),
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


rs0:PRIMARY> use newdb
switched to db newdb

rs0:PRIMARY> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })

rs0:PRIMARY> db.movie.find().pretty()
{ "_id" : ObjectId("5b5efeea9d097ca0600694a3"), "name" : "batman" }

rs0:PRIMARY> exit
bye
```

Now, check the redundancy and data availability in secondary members.
We will exec in `mgo-replicaset-1`(which is secondary member right now) to check the data availability.

```bash
$ kubectl exec -it mgo-replicaset-1 -n demo bash
mongodb@mgo-replicaset-1:/$ mongo admin -u root -p 5O4R2ze2bWXcWsdP
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.

rs0:SECONDARY> rs.slaveOk()
rs0:SECONDARY> > show dbs
admin   0.000GB
config  0.000GB
local   0.000GB
newdb   0.000GB

rs0:SECONDARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("6b714456-2914-4ea0-9596-92249e8285a2"),
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

rs0:SECONDARY> use newdb
switched to db newdb

rs0:SECONDARY> db.movie.find().pretty()
{ "_id" : ObjectId("5b5efeea9d097ca0600694a3"), "name" : "batman" }

rs0:SECONDARY> exit
bye

```

## Automatic Failover

To test automatic failover, we will force the primary member to restart. As the primary member (`pod`) becomes unavailable, the rest of the members will elect a primary member by election.

```bash
$ kubectl get pods -n demo
NAME               READY     STATUS    RESTARTS   AGE
mgo-replicaset-0   1/1       Running   0          1h
mgo-replicaset-1   1/1       Running   0          1h
mgo-replicaset-2   1/1       Running   0          1h

$ kubectl delete pod -n demo mgo-replicaset-0
pod "mgo-replicaset-0" deleted

$ kubectl get pods -n demo
NAME               READY     STATUS        RESTARTS   AGE
mgo-replicaset-0   1/1       Terminating   0          1h
mgo-replicaset-1   1/1       Running       0          1h
mgo-replicaset-2   1/1       Running       0          1h

```

Now verify the automatic failover, Let's exec in `mgo-replicaset-1` pod,

```bash
$ kubectl exec -it mgo-replicaset-1 -n demo bash
mongodb@mgo-replicaset-1:/$ mongo admin -u root -p 5O4R2ze2bWXcWsdP
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.

rs0:SECONDARY> rs.isMaster().primary
mgo-replicaset-2.mgo-replicaset-gvr.demo.svc.cluster.local:27017

# Also verify, data persistency
rs0:SECONDARY> rs.slaveOk()
rs0:SECONDARY> > show dbs
admin   0.000GB
config  0.000GB
local   0.000GB
newdb   0.000GB

rs0:SECONDARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("6b714456-2914-4ea0-9596-92249e8285a2"),
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

rs0:SECONDARY> use newdb
switched to db newdb

rs0:SECONDARY> db.movie.find().pretty()
{ "_id" : ObjectId("5b5efeea9d097ca0600694a3"), "name" : "batman" }
```

## Halt Database

When [DeletionPolicy](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy) is set to halt, and you delete the mongodb object, the KubeDB operator will delete the PetSet and its pods but leaves the PVCs, secrets and database backup (snapshots) intact. Learn details of all `DeletionPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy).

You can also keep the mongodb object and halt the database to resume it again later. If you halt the database, the kubedb will delete the petsets and services but will keep the mongodb object, pvcs, secrets and backup (snapshots).

To halt the database, first you have to set the deletionPolicy to `Halt` in existing database. You can use the below command to set the deletionPolicy to `Halt`, if it is not already set.

```bash
$ kubectl patch -n demo mg/mgo-replicaset -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"
mongodb.kubedb.com/mgo-replicaset patched
```

Then, you have to set the `spec.halted` as true to set the database in a `Halted` state. You can use the below command.

```bash
$ kubectl patch -n demo mg/mgo-replicaset -p '{"spec":{"halted":true}}' --type="merge"
mongodb.kubedb.com/mgo-replicaset patched
```

After that, kubedb will delete the petsets and services and you can see the database Phase as `Halted`.

Now, you can run the following command to get all mongodb resources in demo namespaces,

```bash
$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME                                VERSION   STATUS   AGE
mongodb.kubedb.com/mgo-replicaset   4.4.26     Halted   9m43s

NAME                            TYPE                                  DATA   AGE
secret/default-token-x2zcl      kubernetes.io/service-account-token   3      47h
secret/mgo-replicaset-auth      Opaque                                2      23h
secret/mgo-replicaset-key       Opaque                                1      23h

NAME                                             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mgo-replicaset-0   Bound    pvc-816daa52-ee40-496f-a148-c75344a1b433   1Gi        RWO            standard       9m43s
persistentvolumeclaim/datadir-mgo-replicaset-1   Bound    pvc-e818bc86-ab3c-4ec5-901f-630aab6b814b   1Gi        RWO            standard       9m5s
persistentvolumeclaim/datadir-mgo-replicaset-2   Bound    pvc-5a50bce3-f85f-4157-be22-64dfc26e7517   1Gi        RWO            standard       8m25s
```


## Resume Halted Database

Now, to resume the database, i.e. to get the same database setup back again, you have to set the the `spec.halted` as false. You can use the below command.

```bash
$ kubectl patch -n demo mg/mgo-replicaset -p '{"spec":{"halted":false}}' --type="merge"
mongodb.kubedb.com/mgo-replicaset patched
```

When the database is resumed successfully, you can see the database Status is set to `Ready`.

```bash
$ kubectl get mg -n demo
NAME             VERSION   STATUS    AGE
mgo-replicaset   4.4.26     Ready     6m27s
```

Now, If you again exec into the primary `pod` and look for previous data, you will see that, all the data persists.

```bash
$ kubectl exec -it mgo-replicaset-1 -n demo bash

mongodb@mgo-replicaset-1:/$ mongo admin -u root -p 7QiqLcuSCmZ8PU5a

rs0:PRIMARY> use newdb
switched to db newdb

rs0:PRIMARY> db.movie.find()
{ "_id" : ObjectId("6024b3e47c614cd582c9bb44"), "name" : "batman" }
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mgo-replicaset -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mgo-replicaset

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md) process of MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
