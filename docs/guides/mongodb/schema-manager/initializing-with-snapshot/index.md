---
title: Initializing with Snapshot
menu:
  docs_{{ .version }}:
    identifier: initializing-with-snapshot
    name: Initializing with Snapshot
    parent: mg-schema-manager
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initializing with Snapshot

This guide will show you how to create database and initialize snapshot with MongoDB `Schema Manager` using `Schema Manager Operator`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/install/kubedb/).
- Install `KubeVault` in your cluster following the steps [here](https://kubevault.com/docs/latest/setup/install/kubevault/).

- You should be familiar with the following concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [MongoDBDatabase](/docs/guides/mongodb/concepts/mongodbdatabase.md)
  - [Schema Manager Overview](/docs/guides/mongodb/schema-manager/overview/index.md)
  - [Stash Overview](https://stash.run/docs/latest/concepts/what-is-stash/overview/)  
  - [KubeVault Overview](https://kubevault.com/docs/latest/concepts/overview/)

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mongodb/schema-manager/initializing-with-snapshot/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mongodb/schema-manager/initializing-with-snapshot/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.


### Create Namespace

We are going to create two different namespaces, in `db` namespace we will deploy MongoDB and Vault Server and in `demo` namespacae we will deploy `Schema Manager`. Let’s create those namespace using the following yaml,

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: db
  labels:
    kubernetes.io/metadata.name: db
---
apiVersion: v1
kind: Namespace
metadata:
  name: demo
  labels:
    kubernetes.io/metadata.name: demo
```
Let’s save this yaml configuration into `namespace.yaml` Then create those above namespaces.

```bash
$ kubectl apply -f namespace.yaml
namespace/db created
namespace/demo created
```

## Deploy MongoDB Server and Vault Server 

Here, we are going to deploy a `MongoDB Server` by using `KubeDB` operator. Also, we are deploying a `Vault Server` using `KubeVault` Operator.

### Deploy MongoDB Server

In this section, we are going to deploy a MongoDB Server. Let’s deploy it using this following yaml,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongodb
  namespace: db
spec:
  allowedSchemas:
    namespaces:
      from: All
  version: "4.4.26"
  replicaSet:
    name: "replicaset"
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "100m"
          memory: "100Mi"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 100Mi
  deletionPolicy: WipeOut
```

Here,

- `spec.version` is the name of the MongoDBVersion CR. Here, we are using MongoDB version `4.4.26`.
- `spec.storageType` specifies the type of storage that will be used for MongoDB. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the MongoDB using `EmptyDir` volume.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.allowedSchemas` specifies the namespace and selectors of allowed `Schema Manager`.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete the operation of MongoDB CR. *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specdeletionpolicy).

Let’s save this yaml configuration into `mongodb.yaml` Then create the above `MongoDB` CR

```bash
$ kubectl apply -f mongodb.yaml 
mongodb.kubedb.com/mongodb created
```

### Deploy Vault Server

In this section, we are going to deploy a Vault Server. Let’s deploy it using this following yaml,

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: 1.8.2
  replicas: 3
  allowedSecretEngines:
    namespaces:
      from: All
    secretEngines:
      - mongodb
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  backend:
    raft:
      path: "/vault/data"
      storage:
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  authMethods:
    - type: kubernetes
      path: kubernetes
  deletionPolicy: WipeOut
```

Here,

- `spec.version` is a required field that specifies the original version of Vault that has been used to build the docker image specified in `spec.vault.image` field.
- `spec.replicas` specifies the number of Vault nodes to deploy. It has to be a positive number.
- `spec.allowedSecretEngines` defines the types of Secret Engines & the Allowed namespaces from where a `SecretEngine` can be attached to the `VaultServer`.
- `spec.unsealer` is an optional field that specifies `Unsealer` configuration. `Unsealer` handles automatic initializing and unsealing of Vault.
- `spec.backend` is a required field that specifies the Vault backend storage configuration. KubeVault operator generates storage configuration according to this `spec.backend`.
- `spec.authMethods` is an optional field that specifies the list of auth methods to enable in Vault.
- `spec.deletionPolicy` is an optional field that gives flexibility whether to nullify(reject) the delete operation of VaultServer crd or which resources KubeVault operator should keep or delete when you delete VaultServer crd. 

Let’s save this yaml configuration into `vault.yaml` Then create the above `VaultServer` CR

```bash
$ kubectl apply -f vault.yaml
vaultserver.kubevault.com/vault created
```


### Create Repository Secret

Here, we are using local backend for storing data snapshots. It can be a cloud storage like GCS bucket, AWS S3, Azure Blob Storage, NFS etc. or a Kubernetes native resources like HostPath, PersistentVolumeClaim etc. For more information check [HERE](https://stash.run/docs/latest/guides/backends/overview/)

Let's, create a Secret for our Repository,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo repo-secret --from-file=./RESTIC_PASSWORD
secret/repo-secret created
```

Let’s save this yaml configuration into `repo-secret.yaml` Then create the secret,

```bash
$ kubectl apply -f repo-secret.yaml 
secret/repo-secret created
```


### Create Repository

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: repo
  namespace: demo
spec:
  backend:
    local:
      mountPath: /hello
      persistentVolumeClaim:
        claimName: snapshot-pvc
    storageSecretName: repo-secret
  usagePolicy:
    allowedNamespaces:
      from: All
```
This repository CRO specifies the `repo-secret` that we've created before and specifies the name and path to the local storage `PVC`. 

> Note: Here, we are using local storage `PVC`. My `PVC` name is `snapshot-pvc`. Don’t forget to change `backend.local.persistentVolumeClaim.claimName` to your `PVC` name.

Let’s save this yaml configuration into `repo.yaml` Lets create the repository,

```bash
$ kubectl apply -f repo.yaml 
repository.stash.appscode.com/repo created
```

After creating the repository we've backed up one of our MongoDB database with some sample data via Stash. So, now our repository contains some sample data inside it.


### Configure Snapshot Restore

Now, We are going to create a ServiceAccount, ClusterRole and ClusterRoleBinding. Stash does not grant necessary RBAC permissions to the restore job for taking restore from a different namespace. In this case, we have to provide the RBAC permissions manually. This helps to prevent unauthorized namespaces from getting access to a database via Stash. You can configure this process through this [Documentation](https://stash.run/docs/latest/guides/managed-backup/dedicated-backup-namespace/#configure-restore)

### Deploy Schema Manager Initialize with Snapshot

Here, we are going to deploy `Schema Manager` with the `demo` namespace that we have created above. Let’s deploy it using the following yaml,

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MongoDBDatabase
metadata:
  name: schema-restore
  namespace: demo
spec:
  database:
    serverRef:
      name: mongodb
      namespace: db
    config:
      name: products
  vaultRef:
    name: vault
    namespace: demo
  accessPolicy:
    subjects:
      - name: "saname"
        namespace: db
        kind: "ServiceAccount"
        apiGroup: ""
    defaultTTL: "5m"
    maxTTL: "200h"
  init:
    initialized: false
    snapshot:
      repository:
        name: repo
        namespace: demo
  deletionPolicy: Delete
```

Here,

- `spec.database` is a required field specifying the database server reference and the desired database configuration.
- `spec.vaultRef` is a required field that specifies which KubeVault server to use for user management.
- `spec.accessPolicy` is a required field that specifies the access permissions like which service account or cluster user have the access and also for how long they can access through it.
- `spec.init` is an optional field, containing the information of a script or a snapshot using which the database should be initialized during creation.
- `spec.deletionPolicy` is an optional field that gives flexibility whether to `nullify` (reject) the delete operation.

Let’s save this yaml configuration into `schema-restore.yaml` and apply it,

```bash
$ kubectl apply -f schema-restore.yaml 
mongodbdatabase.schema.kubedb.com/schema-restore created

```

Let's check the `STATUS` of `Schema Manager`,

```bash
$ kubectl get mongodbdatabase -A
NAMESPACE   NAME              DB_SERVER   DB_NAME    STATUS    AGE
demo        schema-restore    mongodb     products   Current   56s
```
Here,

> In `STATUS` section, `Current` means that the current `Secret` of `Schema Manager` is vaild, and it will automatically `Expired` after it reaches the limit of `defaultTTL` that we've defined in the above yaml. 

Also, check the `STATUS` of `restoresession`

```bash
$ kubectl get restoresession -n demo
NAME                      REPOSITORY   PHASE       DURATION   AGE
schema-restore-mongo-rs   repo         Succeeded   5s         21s
```


Now, let's get the secret name from `schema-manager`, and the login credentials for connecting to the database,

```bash
$ kubectl get mongodbdatabase schema-restore -n demo -o=jsonpath='{.status.authSecret.name}'
schema-restore-mongo-req-98k0ch

$ kubectl view-secret -n demo schema-restore-mongo-req-98k0ch -a
password=6ykdBljJ7D8agXeoSp-f
username=v-kubernetes-demo-k8s-f7695915-1e-2zXmduPS89LfvW6tr5Bw-1662639843
```

### Verify Initialization

Here, we are going to connect to the database with the login credentials and verify the database initialization,

```bash
$ kubectl exec -it -n demo mongodb-0 -c mongodb -- bash
root@mongodb-0:/# mongo --authenticationDatabase=products --username='v-kubernetes-demo-k8s-f7695915-1e-2zXmduPS89LfvW6tr5Bw-1662639843' --password='6ykdBljJ7D8agXeoSp-f' products
MongoDB shell version v4.4.26
...

replicaset:PRIMARY> show dbs
products       0.000GB

replicaset:PRIMARY> show collections
products

replicaset:PRIMARY> db.products.find()
{ "_id" : ObjectId("631b3139187d1588626fb80b"), "name" : "kubedb" }

replicaset:PRIMARY> exit
bye

```

Now, Let's check the `STATUS` of `Schema Manager` again,

```bash
$ kubectl get mongodbdatabase -A
NAMESPACE   NAME              DB_SERVER   DB_NAME    STATUS    AGE
demo        schema-restore    mongodb     products   Expired   7m
```

Here, we can see that the `STATUS` of the `schema-manager` is `Expired` because it's exceeded `defaultTTL: "5m"`, which means the current `Secret` of `Schema Manager` isn't vaild anymore. Now, if we try to connect and login with the credentials that we have acquired before from `schema-manager`, it won't work.


```bash
$ kubectl exec -it -n demo mongodb-0 -c mongodb -- bash
root@mongodb-0:/# mongo --authenticationDatabase=products --username='v-kubernetes-demo-k8s-f7695915-1e-2zXmduPS89LfvW6tr5Bw-1662639843' --password='6ykdBljJ7D8agXeoSp-f' products
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/initdb?authSource=initdb&compressors=disabled&gssapiServiceName=mongodb
Error: Authentication failed. :
connect@src/mongo/shell/mongo.js:374:17
@(connect):2:6
exception: connect failed
exiting with code 1
root@mongodb-0:/# exit
exit
```
> We can't connect to the database with the login credentials, which is `Expired`. We will not be able to access the database even though we're in the middle of a connected session. 



## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete ns db
$ kubectl delete ns demo
```


## Next Steps

- Detail concepts of [MongoDBDatabase object](/docs/guides/mongodb/concepts/mongodbdatabase.md).
- Go through the concepts of [KubeVault](https://kubevault.com/docs/latest/guides).
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).