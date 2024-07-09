---
title: MongoDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: mg-quickstart-quickstart
    name: Overview
    parent: mg-quickstart-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB QuickStart

This tutorial will show you how to use KubeDB to run a MongoDB database.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/mongodb/quick-start.png">
</p>

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  2m5s

  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available MongoDBVersion

When you have installed KubeDB, it has created `MongoDBVersion` crd for all supported MongoDB versions. Check 0

```bash
$ kubectl get mongodbversions
NAME             VERSION   DISTRIBUTION   DB_IMAGE                                 DEPRECATED   AGE
3.4.17-v1        3.4.17    Official       mongo:3.4.17                                          68s
3.4.22-v1        3.4.22    Official       mongo:3.4.22                                          68s
3.6.13-v1        3.6.13    Official       mongo:3.6.13                                          68s
4.4.26         3.6.8     Official       mongo:3.6.8                                           68s
4.0.11-v1        4.0.11    Official       mongo:4.0.11                                          68s
4.0.3-v1         4.0.3     Official       mongo:4.0.3                                           68s
4.4.26         4.0.5     Official       mongo:4.0.5                                           68s
4.4.26        4.1.13    Official       mongo:4.1.13                                          68s
4.1.4-v1         4.1.4     Official       mongo:4.1.4                                           68s
4.1.7-v3         4.1.7     Official       mongo:4.1.7                                           68s
4.4.26            4.4.26     Official       mongo:4.4.26                                           68s
4.4.26            4.4.26     Official       mongo:4.4.26                                           68s
5.0.2            5.0.2     Official       mongo:5.0.2                                           68s
5.0.3            5.0.3     Official       mongo:5.0.3                                           68s
percona-3.6.18   3.6.18    Percona        percona/percona-server-mongodb:3.6.18                 68s
percona-4.0.10   4.0.10    Percona        percona/percona-server-mongodb:4.0.10                 68s
percona-4.2.7    4.2.7     Percona        percona/percona-server-mongodb:4.2.7-7                68s
percona-4.4.10   4.4.10    Percona        percona/percona-server-mongodb:4.4.10                 68s

```

## Create a MongoDB database

KubeDB implements a `MongoDB` CRD to define the specification of a MongoDB database. Below is the `MongoDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mgo-quickstart
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
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/quickstart/replicaset.yaml
mongodb.kubedb.com/mgo-quickstart created
```

Here,

- `spec.version` is name of the MongoDBVersion crd where the docker images are specified. In this tutorial, a MongoDB 4.4.26 database is created.
- `spec.storageType` specifies the type of storage that will be used for MongoDB database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MongoDB database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MongoDB` crd or which resources KubeDB should keep or delete when you delete `MongoDB` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy)
- `spec.replicaSet` denotes the name of the mongodb replica-set structure.
- `spec.replicas` denotes the number of replicas in the replica-set.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MongoDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `<mongodb-name>-pods`.

```bash
$ kubectl dba describe mg -n demo mgo-quickstart
Name:               mgo-quickstart
Namespace:          demo
CreationTimestamp:  Mon, 13 Jun 2022 18:01:55 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-quickstart","namespace":"demo"},"spec":{"replicaSet":{"na...
Replicas:           3  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  DoNotTerminate

StatefulSet:          
  Name:               mgo-quickstart
  CreationTimestamp:  Mon, 13 Jun 2022 18:01:55 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mgo-quickstart
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:        <none>
  Replicas:           824645483384 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mgo-quickstart
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.20.114
  Port:         primary  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.12:27017

Service:        
  Name:         mgo-quickstart-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.12:27017,10.244.0.14:27017,10.244.0.16:27017

Auth Secret:
  Name:         mgo-quickstart-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-quickstart
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
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-quickstart","namespace":"demo"},"spec":{"replicaSet":{"name":"rs1"},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"DoNotTerminate","version":"4.4.26"}}

    Creation Timestamp:  2022-06-13T12:01:55Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mgo-quickstart
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mongodbs.kubedb.com
    Name:                            mgo-quickstart
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mgo-quickstart
        Port:    27017
        Scheme:  mongodb
    Parameters:
      API Version:  config.kubedb.com/v1alpha1
      Kind:         MongoConfiguration
      Replica Sets:
        host-0:  rs1/mgo-quickstart-0.mgo-quickstart-pods.demo.svc:27017,mgo-quickstart-1.mgo-quickstart-pods.demo.svc:27017,mgo-quickstart-2.mgo-quickstart-pods.demo.svc:27017
      Stash:
        Addon:
          Backup Task:
            Name:  mongodb-backup-4.4.6
          Restore Task:
            Name:  mongodb-restore-4.4.6
    Secret:
      Name:   mgo-quickstart-auth
    Type:     kubedb.com/mongodb
    Version:  4.4.26

Events:
  Type    Reason      Age   From             Message
  ----    ------      ----  ----             -------
  Normal  Successful  3m    KubeDB Operator  Successfully created governing service
  Normal  Successful  3m    KubeDB Operator  Successfully created Primary Service
  Normal  Successful  3m    KubeDB Operator  Successfully created appbinding
```

```bash
$ kubectl get statefulset -n demo
NAME             READY   AGE
mgo-quickstart   3/3     3m36s

$ kubectl get pvc -n demo
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
datadir-mgo-quickstart-0   Bound    pvc-18c3c456-c9a9-40b2-bec8-4302cc0aeccc   1Gi        RWO            standard       3m56s
datadir-mgo-quickstart-1   Bound    pvc-7ac4c470-8fa7-47a9-b118-2ac20f01186d   1Gi        RWO            standard       104s
datadir-mgo-quickstart-2   Bound    pvc-2e6dfb71-056b-4186-927d-855db35d0014   1Gi        RWO            standard       77s

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   REASON   AGE
pvc-18c3c456-c9a9-40b2-bec8-4302cc0aeccc   1Gi        RWO            Delete           Bound    demo/datadir-mgo-quickstart-0   standard                4m8s
pvc-2e6dfb71-056b-4186-927d-855db35d0014   1Gi        RWO            Delete           Bound    demo/datadir-mgo-quickstart-2   standard                90s
pvc-7ac4c470-8fa7-47a9-b118-2ac20f01186d   1Gi        RWO            Delete           Bound    demo/datadir-mgo-quickstart-1   standard                117s

$ kubectl get service -n demo
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mgo-quickstart        ClusterIP   10.96.20.114   <none>        27017/TCP   4m25s
mgo-quickstart-pods   ClusterIP   None           <none>        27017/TCP   4m25s

```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mgo-quickstart -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-quickstart","namespace":"demo"},"spec":{"replicaSet":{"name":"rs1"},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"DoNotTerminate","version":"4.4.26"}}
  creationTimestamp: "2022-06-13T12:01:55Z"
  finalizers:
    - kubedb.com
  generation: 3
  name: mgo-quickstart
  namespace: demo
  resourceVersion: "2069"
  uid: 197705bd-1558-4c01-aaac-c452d6972433
spec:
  allowedSchemas:
    namespaces:
      from: Same
  arbiter: null
  authSecret:
    name: mgo-quickstart-auth
  clusterAuthMode: keyFile
  coordinator:
    resources: {}
  keyFileSecret:
    name: mgo-quickstart-key
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
                    app.kubernetes.io/instance: mgo-quickstart
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                namespaces:
                  - demo
                topologyKey: kubernetes.io/hostname
              weight: 100
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: mgo-quickstart
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
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: mgo-quickstart
  replicaSet:
    name: rs1
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
    - lastTransitionTime: "2022-06-13T12:01:55Z"
      message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mgo-quickstart'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2022-06-13T12:04:58Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2022-06-13T12:03:35Z"
      message: 'The MongoDB: demo/mgo-quickstart is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2022-06-13T12:03:35Z"
      message: 'The MongoDB: demo/mgo-quickstart is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2022-06-13T12:04:58Z"
      message: 'The MongoDB: demo/mgo-quickstart is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 3
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `mgo-quickstart-auth` *(format: {mongodb-object-name}-auth)* for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the *username* for MongoDB superuser and a `password` key which contains the *password* for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

Now, you can connect to this database through [mongo-shell](https://docs.mongodb.com/v3.4/mongo/). In this tutorial, we are connecting to the MongoDB server from inside the pod.

```bash
$ kubectl get secrets -n demo mgo-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mgo-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
CaM8v9LmmSGB~&hj

$ kubectl exec -it mgo-quickstart-0 -n demo sh

> mongo admin

rs1:PRIMARY> db.auth("root","CaM8v9LmmSGB~&hj")
1


rs1:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

rs1:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("1e460a23-705d-47a4-b80a-9d2fb947e915"),
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


rs1:PRIMARY> use mydb
switched to db mydb

rs1:PRIMARY> db.movies.insertOne({"top gun": "maverick"})
{
	"acknowledged" : true,
	"insertedId" : ObjectId("62a72949198bad2c983d6611")
}

rs1:PRIMARY> db.movies.find()
{ "_id" : ObjectId("62a72949198bad2c983d6611"), "top gun" : "maverick" }

> exit
bye
```

# Database DeletionPolicy
This field is used to regulate the deletion process of the related resources when mongodb object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

## DoNotTerminate Property

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete mg mgo-quickstart -n demo
Error from server (BadRequest): admission webhook "mongodbwebhook.validators.kubedb.com" denied the request: mongodb "demo/mgo-quickstart" can't be terminated. To delete, change spec.deletionPolicy
```

## Halt Database

When [DeletionPolicy](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy) is set to halt, and you delete the mongodb object, the KubeDB operator will delete the StatefulSet and its pods but leaves the PVCs, secrets and database backup (snapshots) intact. Learn details of all `DeletionPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy).

You can also keep the mongodb object and halt the database to resume it again later. If you halt the database, the kubedb operator will delete the statefulsets and services but will keep the mongodb object, pvcs, secrets and backup (snapshots).

To halt the database, first you have to set the deletionPolicy to `Halt` in existing database. You can use the below command to set the deletionPolicy to `Halt`, if it is not already set.

```bash
$ kubectl patch -n demo mg/mgo-quickstart -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"
mongodb.kubedb.com/mgo-quickstart patched
```

Then, you have to set the `spec.halted` as true to set the database in a `Halted` state. You can use the below command.

```bash
$ kubectl patch -n demo mg/mgo-quickstart -p '{"spec":{"halted":true}}' --type="merge"
mongodb.kubedb.com/mgo-quickstart patched
```

After that, kubedb will delete the statefulsets and services and you can see the database Phase as `Halted`.

Now, you can run the following command to get all mongodb resources in demo namespaces,

```bash
$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME                                VERSION   STATUS   AGE
mongodb.kubedb.com/mgo-quickstart   4.4.26     Halted   12m

NAME                         TYPE                                  DATA   AGE
secret/default-token-swg6h   kubernetes.io/service-account-token   3      12m
secret/mgo-quickstart-auth   Opaque                                2      12m
secret/mgo-quickstart-key    Opaque                                1      12m

NAME                                             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mgo-quickstart-0   Bound    pvc-18c3c456-c9a9-40b2-bec8-4302cc0aeccc   1Gi        RWO            standard       12m
persistentvolumeclaim/datadir-mgo-quickstart-1   Bound    pvc-7ac4c470-8fa7-47a9-b118-2ac20f01186d   1Gi        RWO            standard       9m57s
persistentvolumeclaim/datadir-mgo-quickstart-2   Bound    pvc-2e6dfb71-056b-4186-927d-855db35d0014   1Gi        RWO            standard       9m30s
```


## Resume Halted Database

Now, to resume the database, i.e. to get the same database setup back again, you have to set the the `spec.halted` as false. You can use the below command.

```bash
$ kubectl patch -n demo mg/mgo-quickstart -p '{"spec":{"halted":false}}' --type="merge"
mongodb.kubedb.com/mgo-quickstart patched
```

When the database is resumed successfully, you can see the database Status is set to `Ready`.

```bash
$ kubectl get mg -n demo
NAME             VERSION   STATUS   AGE
mgo-quickstart   4.4.26     Ready    13m
```

Now, If you again exec into the `pod` and look for previous data, you will see that, all the data persists.

```bash
$ kubectl exec -it mgo-quickstart-0 -n demo bash

mongodb@mgo-quickstart-0:/$ mongo admin -u root -p CaM8v9LmmSGB~&hj
rs1:SECONDARY> use mydb
switched to db mydb

rs1:SECONDARY> rs.slaveOk()
WARNING: slaveOk() is deprecated and may be removed in the next major release. Please use secondaryOk() instead.

rs1:SECONDARY> db.movies.find()
{ "_id" : ObjectId("62a72949198bad2c983d6611"), "top gun" : "maverick" }

```


## Cleaning up

If you don't set the deletionPolicy, then the kubeDB set the DeletionPolicy to `Delete` by-default.

### Delete
If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken snapshots and secrets then you might want to set the mongodb object deletionPolicy to Delete. In this setting, StatefulSet and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the DeletionPolicy is set to Delete and the mongodb object is deleted, the KubeDB operator will delete the StatefulSet and its pods along with PVCs but leaves the secret and database backup data(snapshots) intact.

```bash
$ kubectl patch -n demo mg/mgo-quickstart -p '{"spec":{"deletionPolicy":"Delete"}}' --type="merge"
kubectl delete -n demo mg/mgo-quickstart

$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME                         TYPE                                  DATA   AGE
secret/default-token-swg6h   kubernetes.io/service-account-token   3      27m
secret/mgo-quickstart-auth   Opaque                                2      27m
secret/mgo-quickstart-key    Opaque                                1      27m

$ kubectl delete ns demo
```

### WipeOut
But if you want to cleanup each of the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo mg/mgo-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo mg/mgo-quickstart

$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME              TYPE              DATA   AGE

$ kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend using `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.

2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to resume database. So, we have `Halt` option which preserves all your `PVCs`, `Secrets`, `Snapshots` etc. If you don't want to resume database, you can just use `spec.deletionPolicy: WipeOut`. It will delete everything created by KubeDB for a particular MongoDB crd when you delete the mongodb object. For more details about termination policy, please visit [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy).

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md) MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
