---
title: FerretDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: fr-quickstart-quickstart
    name: Overview
    parent: fr-quickstart-ferretdb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# FerretDB QuickStart

This tutorial will show you how to use KubeDB to run a FerretDB database.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/ferretdb/quick-start.png">
</p>

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

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

## Find Available FerretDBVersion

When you have installed KubeDB, it has created `FerretDBVersion` CR for all supported FerretDB versions.

```bash
$ kubectl get ferretdbversions
NAME     VERSION   DB_IMAGE                                  DEPRECATED   AGE
1.18.0   1.18.0    ghcr.io/appscode-images/ferretdb:1.18.0                104m
1.23.0   1.23.0    ghcr.io/appscode-images/ferretdb:1.23.0                104m
1.24.0   1.24.0    ghcr.io/appscode-images/ferretdb:1.24.0                104m
2.0.0    2.0.0     ghcr.io/appscode-images/ferretdb:2.0.0                 5d4h
```

## Create a FerretDB database

FerretDB use Postgres as it's main backend. Currently, KubeDB supports Postgres backend as database engine for FerretDB. KubeDB operator will create and manage the backend Postgres for FerretDB

Below is the `FerretDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferret
  namespace: demo
spec:
  version: "2.0.0"
  authSecret:
    kind: Secret
    externallyManaged: false
  sslMode: disabled
  backend:
    storageType: Durable
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 500Mi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferredb/quickstart/ferretdb-internal.yaml
ferretdb.kubedb.com/ferret created
```

Here,

- `spec.version` is name of the FerretDBVersion CR where the docker images are specified. In this tutorial, a FerretDB 1.18.0 database is created.
- `spec.storageType` specifies the type of storage that will be used for FerretDB database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create FerretDB database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `FerretDB` CR or which resources KubeDB should keep or delete when you delete `FerretDB` CR. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy)

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `FerretDB` objects using Kubernetes api. When a `FerretDB` object is created, KubeDB operator will create a new PetSet and a Service with the matching FerretDB object name. KubeDB operator will also create a governing service for PetSets with the name `<ferretdb-name>-pods`.

KubeDB will create a Postgres database alongside with FerretDB for FerretDB's backend engine.

KubeDB operator sets the `status.phase` to Ready once the database is successfully provisioned and ready to use.

```bash
$ kubectl get fr -n demo
NAME     NAMESPACE   VERSION   STATUS   AGE
ferret   demo        2.0.0     Ready    5m10s

$ kubectl get pg -n demo
NAME                VERSION           STATUS   AGE
ferret-pg-backend   17.4-documentdb   Ready    6m10s
```

Let’s describe FerretDB object ferret

```bash
$ kubectl describe fr ferret -n demo
Name:         ferret
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         FerretDB
Metadata:
  Creation Timestamp:  2025-04-03T05:41:50Z
  Finalizers:
    kubedb.com
  Generation:        3
  Resource Version:  3098298
  UID:               81a859d5-8f1f-4475-a9f7-2b2a42d9e626
Spec:
  Auth Secret:
    Externally Managed:  false
    Name:                ferret-auth
  Deletion Policy:       WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Server:
    Primary:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  ferretdb
            Resources:
              Limits:
                Memory:  1Gi
              Requests:
                Cpu:     500m
                Memory:  1Gi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Group:     1000
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
  Ssl Mode:            disabled
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:  500Mi
  Storage Type:   Durable
  Version:        2.0.0
Status:
  Conditions:
    Last Transition Time:  2025-04-03T05:41:51Z
    Message:               The KubeDB operator has started the provisioning of FerretDB: demo/ferret
    Observed Generation:   2
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-04-03T05:42:43Z
    Message:               All replicas are ready for FerretDB demo/ferret
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-04-03T05:42:54Z
    Message:               The FerretDB: demo/ferret is accepting client requests.
    Observed Generation:   3
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-04-03T05:42:54Z
    Message:               The FerretDB: demo/ferret is ready.
    Observed Generation:   3
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-04-03T05:42:54Z
    Message:               The FerretDB: demo/ferret is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

```bash
$ kubectl get petset -n demo
NAME                AGE
ferret              2m26s
ferret-pg-backend   3m4s

$ kubectl get appbindings -n demo
NAME                TYPE                  VERSION   AGE
ferret              kubedb.com/ferretdb   2.0.0     6m6s
ferret-pg-backend   kubedb.com/postgres   17.4      6m54s

$ kubectl get pvc -n demo
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-ferret-pg-backend-0   Bound    pvc-3438b4f7-aeb0-46a6-a427-bdf593c9fb80   500Mi      RWO            local-path     <unset>                 8m30s
data-ferret-pg-backend-1   Bound    pvc-c2ece700-864c-49c6-81ad-87d2eadb0200   500Mi      RWO            local-path     <unset>                 8m13s
data-ferret-pg-backend-2   Bound    pvc-4c456891-c0ab-4452-82b9-453aa68bfc0c   500Mi      RWO            local-path     <unset>                 8m6s

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-3438b4f7-aeb0-46a6-a427-bdf593c9fb80   500Mi      RWO            Delete           Bound    demo/data-ferret-pg-backend-0   local-path     <unset>                          8m42s
pvc-4c456891-c0ab-4452-82b9-453aa68bfc0c   500Mi      RWO            Delete           Bound    demo/data-ferret-pg-backend-2   local-path     <unset>                          8m17s
pvc-c2ece700-864c-49c6-81ad-87d2eadb0200   500Mi      RWO            Delete           Bound    demo/data-ferret-pg-backend-1   local-path     <unset>                          8m23s

$ kubectl get service -n demo
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
ferret                      ClusterIP   10.43.63.55    <none>        27017/TCP                    4m28s
ferret-pg-backend           ClusterIP   10.43.156.60   <none>        5432/TCP,2379/TCP            4m28s
ferret-pg-backend-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   4m28s
ferret-pg-backend-standby   ClusterIP   10.43.128.95   <none>        5432/TCP                     4m28s
ferret-pods                 ClusterIP   None           <none>        27017/TCP                    4m28s
```

Run the following command to see the modified FerretDB object:

```yaml
$ kubectl get fr ferret -n demo -oyaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"FerretDB","metadata":{"annotations":{},"name":"ferret","namespace":"demo"},"spec":{"authSecret":{"externallyManaged":false},"deletionPolicy":"WipeOut","sslMode":"disabled","storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"500Mi"}}},"storageType":"Durable","version":"2.0.0"}}
  creationTimestamp: "2025-04-03T05:41:50Z"
  finalizers:
    - kubedb.com
  generation: 3
  name: ferret
  namespace: demo
  resourceVersion: "3098298"
  uid: 81a859d5-8f1f-4475-a9f7-2b2a42d9e626
spec:
  authSecret:
    kind: Secret
    externallyManaged: false
    name: ferret-auth
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  server:
    primary:
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          containers:
            - name: ferretdb
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 1Gi
              securityContext:
                allowPrivilegeEscalation: false
                capabilities:
                  drop:
                    - ALL
                runAsGroup: 1000
                runAsNonRoot: true
                runAsUser: 1000
                seccompProfile:
                  type: RuntimeDefault
          podPlacementPolicy:
            name: default
          securityContext:
            fsGroup: 1000
      replicas: 1
  sslMode: disabled
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  storageType: Durable
  version: 2.0.0
status:
  conditions:
    - lastTransitionTime: "2025-04-03T05:41:51Z"
      message: 'The KubeDB operator has started the provisioning of FerretDB: demo/ferret'
      observedGeneration: 2
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2025-04-03T05:42:43Z"
      message: All replicas are ready for FerretDB demo/ferret
      observedGeneration: 3
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2025-04-03T05:42:54Z"
      message: 'The FerretDB: demo/ferret is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2025-04-03T05:42:54Z"
      message: 'The FerretDB: demo/ferret is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2025-04-03T05:42:54Z"
      message: 'The FerretDB: demo/ferret is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `ferret-auth` *(format: {ferretdb-object-name}-auth)* for storing the password for `postgres` superuser. This secret contains a `username` key which contains the *username* for FerretDB superuser and a `password` key which contains the *password* for FerretDB superuser.

If you want to use custom or existing secret please specify that when creating the FerretDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

Now, you can connect to this database by port-forwarding primary service `ferret` and connecting with [mongo-shell](https://www.mongodb.com/try/download/shell) locally

```bash
$ kubectl get secrets -n demo ferret-auth -o jsonpath='{.data.\username}' | base64 -d
postgres
$ kubectl get secrets -n demo ferret-auth -o jsonpath='{.data.\\password}' | base64 -d
UxV5a35kURSFE(;5

$ kubectl port-forward svc/ferret -n demo 27017
Forwarding from 127.0.0.1:27017 -> 27017
Forwarding from [::1]:27017 -> 27017
Handling connection for 27017
Handling connection for 27017
```

Now in another terminal

```bash
$ mongosh 'mongodb://postgres:UxV5a35kURSFE(;5@localhost:27017/ferretdb'
Current Mongosh Log ID:	67ee22bbd9c3422c286b140a
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.4.2
Using MongoDB:		7.0.77
Using Mongosh:		2.4.2

For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/

------
   The server generated these startup warnings when booting
   2025-04-03T05:55:07.528Z: Powered by FerretDB v2.0.0-1-g7fb2c9a8 and DocumentDB 0.102.0 (PostgreSQL 17.4).
   2025-04-03T05:55:07.528Z: Please star 🌟 us on GitHub: https://github.com/FerretDB/FerretDB and https://github.com/microsoft/documentdb.
   2025-04-03T05:55:07.528Z: The telemetry state is undecided. Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.
------

ferretdb> 


ferretdb> show dbs
kubedb_system  80.00 KiB

ferretdb> use mydb
switched to db mydb

mydb> db.movies.insertOne({"top gun": "maverick"})
{
  acknowledged: true,
  insertedId: ObjectId('65efeee6a3347fff66d04c71')
}

mydb> db.movies.find()
[
  { _id: ObjectId('65efeee6a3347fff66d04c71'), 'top gun': 'maverick' }
]

mydb> show dbs
kubedb_system  80.00 KiB
mydb           80.00 KiB

mydb> exit
```
All these data inside FerretDB is also storing inside `ferret-pg-backend` Postgres.

## Cleaning up

If you don't set the deletionPolicy, then the kubeDB set the DeletionPolicy to `WipeOut` by-default for `FerretDB`.

### WipeOut
If you want to clean up each of the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete -n demo fr/ferret

$ kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend using `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.

