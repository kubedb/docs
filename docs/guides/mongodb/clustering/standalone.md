---
title: MongoDB Standalone Guide
menu:
  docs_{{ .version }}:
    identifier: mg-clustering-standalone
    name: Standalone Guide
    parent: mg-clustering-mongodb
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MongoDB Standalone

This tutorial will show you how to use KubeDB to run a MongoDB Standalone.

## Before You Begin

Before proceeding:

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MongoDB Standalone

To deploy a MongoDB Standalone, user must have `spec.replicaSet` & `spec.shardTopology` options in `Mongodb` CRD to be set to nil. Arbiter & Hidden-node are also not supported for standalone mongoDB.

The following is an example of a `Mongodb` object which creates MongoDB Standalone database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-alone
  namespace: demo
spec:
  version: "4.2.3"
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "300m"
          memory: "400Mi"
  storage:
    resources:
      requests:
        storage: 500Mi
    storageClassName: standard
  terminationPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/clustering/standalone.yaml
mongodb.kubedb.com/mg-alone created
```

Here,

- `spec.version` is the version to be used for MongoDB.
- `spec.podTemplate` specifies the resources and other specifications of the pod. Have a look [here](/docs/guides/mongodb/concepts/mongodb.md#specpodtemplate) to know the other subfields of the podTemplate.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MongoDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `<mongodb-name>-pods`.

```bash
$ kubectl dba describe mg -n demo mg-alone
Name:               mg-alone
Namespace:          demo
CreationTimestamp:  Fri, 04 Nov 2022 10:30:07 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mg-alone","namespace":"demo"},"spec":{"podTemplate":{"spec":{...
Replicas:           1  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          500Mi
Paused:              false
Halted:              false
Termination Policy:  WipeOut

StatefulSet:          
  Name:               mg-alone
  CreationTimestamp:  Fri, 04 Nov 2022 10:30:07 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mg-alone
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:        <none>
  Replicas:           824638445048 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mg-alone
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mg-alone
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.47.157
  Port:         primary  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.25:27017

Service:        
  Name:         mg-alone-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mg-alone
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.25:27017

Auth Secret:
  Name:         mg-alone-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mg-alone
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
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mg-alone","namespace":"demo"},"spec":{"podTemplate":{"spec":{"resources":{"requests":{"cpu":"300m","memory":"400Mi"}}}},"storage":{"resources":{"requests":{"storage":"500Mi"}},"storageClassName":"standard"},"terminationPolicy":"WipeOut","version":"4.2.3"}}

    Creation Timestamp:  2022-11-04T04:30:14Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mg-alone
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mongodbs.kubedb.com
    Name:                            mg-alone
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mg-alone
        Port:    27017
        Scheme:  mongodb
    Parameters:
      API Version:  config.kubedb.com/v1alpha1
      Kind:         MongoConfiguration
      Stash:
        Addon:
          Backup Task:
            Name:  mongodb-backup-4.2.3
          Restore Task:
            Name:  mongodb-restore-4.2.3
    Secret:
      Name:   mg-alone-auth
    Type:     kubedb.com/mongodb
    Version:  4.2.3

Events:
  Type    Reason        Age   From              Message
  ----    ------        ----  ----              -------
  Normal  PhaseChanged  21s   MongoDB operator  Phase changed from  to Provisioning.
  Normal  Successful    21s   MongoDB operator  Successfully created governing service
  Normal  Successful    21s   MongoDB operator  Successfully created Primary Service
  Normal  Successful    14s   MongoDB operator  Successfully patched StatefulSet demo/mg-alone
  Normal  Successful    14s   MongoDB operator  Successfully patched MongoDB
  Normal  Successful    14s   MongoDB operator  Successfully created appbinding
  Normal  Successful    14s   MongoDB operator  Successfully patched MongoDB
  Normal  Successful    14s   MongoDB operator  Successfully patched StatefulSet demo/mg-alone
  Normal  Successful    4s    MongoDB operator  Successfully patched StatefulSet demo/mg-alone
  Normal  Successful    4s    MongoDB operator  Successfully patched MongoDB
  Normal  PhaseChanged  4s    MongoDB operator  Phase changed from Provisioning to Ready.
  Normal  Successful    4s    MongoDB operator  Successfully patched StatefulSet demo/mg-alone
  Normal  Successful    4s    MongoDB operator  Successfully patched MongoDB



$ kubectl get sts,svc,pvc,pv -n demo
NAME                        READY   AGE
statefulset.apps/mg-alone   1/1     65s

NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
service/mg-alone        ClusterIP   10.96.47.157   <none>        27017/TCP   65s
service/mg-alone-pods   ClusterIP   None           <none>        27017/TCP   65s

NAME                                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mg-alone-0   Bound    pvc-78328965-1210-4f7a-a508-2749b328a5ac   500Mi      RWO            standard       65s

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                     STORAGECLASS   REASON   AGE
persistentvolume/pvc-78328965-1210-4f7a-a508-2749b328a5ac   500Mi      RWO            Delete           Bound    demo/datadir-mg-alone-0   standard                62s

```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mg-alone -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mg-alone","namespace":"demo"},"spec":{"podTemplate":{"spec":{"resources":{"requests":{"cpu":"300m","memory":"400Mi"}}}},"storage":{"resources":{"requests":{"storage":"500Mi"}},"storageClassName":"standard"},"terminationPolicy":"WipeOut","version":"4.2.3"}}
  creationTimestamp: "2022-11-04T04:30:07Z"
  finalizers:
    - kubedb.com
  generation: 2
  name: mg-alone
  namespace: demo
  resourceVersion: "914996"
  uid: 55ece68c-8df6-4055-b463-1fcb119f0fb1
spec:
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: mg-alone-auth
  autoOps: {}
  coordinator:
    resources: {}
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
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
                    app.kubernetes.io/instance: mg-alone
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                namespaces:
                  - demo
                topologyKey: kubernetes.io/hostname
              weight: 100
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: mg-alone
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
        timeoutSeconds: 5
      resources:
        limits:
          memory: 400Mi
        requests:
          cpu: 300m
          memory: 400Mi
      serviceAccountName: mg-alone
  replicas: 1
  sslMode: disabled
  storage:
    resources:
      requests:
        storage: 500Mi
    storageClassName: standard
  storageEngine: wiredTiger
  storageType: Durable
  terminationPolicy: WipeOut
  version: 4.2.3
status:
  conditions:
    - lastTransitionTime: "2022-11-04T04:30:07Z"
      message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mg-alone'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2022-11-04T04:30:14Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2022-11-04T04:30:24Z"
      message: 'The MongoDB: demo/mg-alone is accepting client requests.'
      observedGeneration: 2
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2022-11-04T04:30:24Z"
      message: 'The MongoDB: demo/mg-alone is ready.'
      observedGeneration: 2
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2022-11-04T04:30:24Z"
      message: 'The MongoDB: demo/mg-alone is successfully provisioned.'
      observedGeneration: 2
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 2
  phase: Ready

```

Please note that KubeDB operator has created a new Secret called `mg-alone-auth` *(format: {mongodb-object-name}-auth)* for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the *username* for MongoDB superuser and a `password` key which contains the *password* for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

## Data Insertion

Now, you can connect to this database through [mg-alone](https://docs.mongodb.com/v3.4/mongo/). In this tutorial, we will insert document on the primary member, and we will see if the data becomes available on secondary members.

At first, insert data inside primary member `rs0:PRIMARY`.

```bash
$ kubectl get secrets -n demo mg-alone-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mg-alone-auth -o jsonpath='{.data.\password}' | base64 -d
5O4R2ze2bWXcWsdP

$ kubectl exec -it mg-alone-0 -n demo bash

mongodb@mg-alone-0:/$ mongo admin -u root -p 5O4R2ze2bWXcWsdP
MongoDB shell version v4.2.3
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.2.3
Welcome to the MongoDB shell.

> rs.isMaster()
{
	"ismaster" : true,
	"maxBsonObjectSize" : 16777216,
	"maxMessageSizeBytes" : 48000000,
	"maxWriteBatchSize" : 100000,
	"localTime" : ISODate("2022-11-04T04:45:33.151Z"),
	"logicalSessionTimeoutMinutes" : 30,
	"connectionId" : 447,
	"minWireVersion" : 0,
	"maxWireVersion" : 8,
	"readOnly" : false,
	"ok" : 1
}

> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB


> use admin
switched to db admin
> show users
{
	"_id" : "admin.root",
	"userId" : UUID("bd711827-8d7e-4c7c-b9d7-ddb27869b9fb"),
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



> use newdb
switched to db newdb
> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })
> db.movie.find().pretty()
{ "_id" : ObjectId("6364996b3bdf351ff67cc7a8"), "name" : "batman" }

> exit
bye
```

## Data availability
As this is a standalone database which doesn't have multiple replicas, It offers no redundancy & high availability of data. All the data are stored in one place, & deleting that will occur in data lost.

## Halt Database

When [TerminationPolicy](/docs/guides/mongodb/concepts/mongodb.md#specterminationpolicy) is set to halt, and you delete the mongodb object, the KubeDB operator will delete the StatefulSet and its pods but leaves the PVCs, secrets and database backup (snapshots) intact. Learn details of all `TerminationPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specterminationpolicy).

You can also keep the mongodb object and halt the database to resume it again later. If you halt the database, the kubedb will delete the statefulsets and services but will keep the mongodb object, pvcs, secrets and backup (snapshots).

To halt the database, first you have to set the terminationPolicy to `Halt` in existing database. You can use the below command to set the terminationPolicy to `Halt`, if it is not already set.

```bash
$ kubectl patch -n demo mg/mg-alone -p '{"spec":{"terminationPolicy":"Halt"}}' --type="merge"
mongodb.kubedb.com/mg-alone patched
```

Then, you have to set the `spec.halted` as true to set the database in a `Halted` state. You can use the below command.

```bash
$ kubectl patch -n demo mg/mg-alone -p '{"spec":{"halted":true}}' --type="merge"
mongodb.kubedb.com/mg-alone patched
```

After that, kubedb will delete the statefulsets and services and you can see the database Phase as `Halted`.

Now, you can run the following command to get all mongodb resources in demo namespaces,

```bash
$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME                          VERSION   STATUS   AGE
mongodb.kubedb.com/mg-alone   4.2.3     Halted   2m4s

NAME                   TYPE                       DATA   AGE
secret/mg-alone-auth   kubernetes.io/basic-auth   2      2m4s
secret/mongo-ca        kubernetes.io/tls          2      15d

NAME                                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mg-alone-0   Bound    pvc-a1a873a6-4f6d-42eb-a38f-83d36fc44e1a   500Mi      RWO            standard       2m4s

```


## Resume Halted Database

Now, to resume the database, i.e. to get the same database setup back again, you have to set the the `spec.halted` as false. You can use the below command.

```bash
$ kubectl patch -n demo mg/mg-alone -p '{"spec":{"halted":false}}' --type="merge"
mongodb.kubedb.com/mg-alone patched
```

When the database is resumed successfully, you can see the database Status is set to `Ready`.

```bash
$ kubectl get mg -n demo
NAME             VERSION   STATUS    AGE
mg-alone   4.2.3     Ready     6m27s
```

Now, If you again exec into the primary `pod` and look for previous data, you will see that, all the data persists.

```bash
$ kubectl exec -it mg-alone-1 -n demo bash

mongodb@mg-alone-1:/$ mongo admin -u root -p 7QiqLcuSCmZ8PU5a

> use newdb
switched to db newdb
> show collections
movie
> db.movie.find()
{ "_id" : ObjectId("6364af93b1ae8e7a8467058a"), "name" : "batman" }

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mg-alone -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mg-alone

kubectl delete ns demo
```

## Next Steps

- Deploy MongoDB [ReplicaSet](/docs/guides/mongodb/clustering/replication_concept.md)
- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md) process of MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
