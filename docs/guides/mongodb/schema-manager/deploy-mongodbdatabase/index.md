---
title: Deploy MongoDBDatabase
menu:
  docs_{{ .version }}:
    identifier: deploy-mongodbdatabase
    name: Deploy MongoDBDatabase
    parent: mg-schema-manager
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Create Database with MongoDB Schema Manager

This guide will show you how to create database with MongoDB Schema Manager using `Schema Manager Operator`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB Enterprise Operator` in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/install/enterprise/).
- Install `KubeVault Enterprise Operator` in your cluster following the steps [here](https://kubevault.com/docs/latest/setup/install/enterprise/).

- You should be familiar with the following concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [MongoDBDatabase](/docs/guides/mongodb/concepts/mongodbdatabase.md)
  - [Schema Manager Overview](/docs/guides/mongodb/schema-manager/overview/index.md)
  - [Stash Overview](https://stash.run/docs/latest/concepts/what-is-stash/overview/)  
  - [KubeVault Overview](https://kubevault.com/docs/latest/concepts/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mongodb/schema-manager/deploy-mongodbdatabase/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mongodb/schema-manager/deploy-mongodbdatabase/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

## Deploy MongoDB Server and Vault Server 

Firstly, we are going to deploy a `MongoDB Server` by using `KubeDB` operator. Also, we are deploying a `Vault Server` using `KubeVault` Operator.

### Deploy MongoDB Server

In this section, we are going to deploy a MongoDB Server. Let’s deploy it using this following yaml,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongodb
  namespace: demo
spec:
  allowedSchemas:
    namespaces:
      from: Selector
      selector:
        matchExpressions:
        - {key: kubernetes.io/metadata.name, operator: In, values: [dev]}
    selector:
      matchLabels:
        "schema.kubedb.com": "mongo"
  version: "4.4.6"
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
  terminationPolicy: WipeOut
```

Here,

- `spec.version` is the name of the MongoDBVersion CR. Here, we are using MongoDB version `4.4.6`.
- `spec.storageType` specifies the type of storage that will be used for MongoDB. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the MongoDB using `EmptyDir` volume.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.allowedSchemas` specifies the namespace and selectors of allowed `Schema Manager`.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete the operation of MongoDB CR. *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specterminationpolicy).


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
  terminationPolicy: WipeOut
```

Here,

- `spec.version` is a required field that specifies the original version of Vault that has been used to build the docker image specified in `spec.vault.image` field.
- `spec.replicas` specifies the number of Vault nodes to deploy. It has to be a positive number.
- `spec.allowedSecretEngines` defines the types of Secret Engines & the Allowed namespaces from where a `SecretEngine` can be attached to the `VaultServer`.
- `spec.unsealer` is an optional field that specifies `Unsealer` configuration. `Unsealer` handles automatic initializing and unsealing of Vault.
- `spec.backend` is a required field that specifies the Vault backend storage configuration. KubeVault operator generates storage configuration according to this `spec.backend`.
- `spec.authMethods` is an optional field that specifies the list of auth methods to enable in Vault.
- `spec.terminationPolicy` is an optional field that gives flexibility whether to nullify(reject) the delete operation of VaultServer crd or which resources KubeVault operator should keep or delete when you delete VaultServer crd. 


Let’s save this yaml configuration into `vault.yaml` Then create the above `VaultServer` CR

```bash
$ kubectl apply -f vault.yaml
vaultserver.kubevault.com/vault created
```

### Create Separate Namespace For Schema Manager

In this section, we are going to create a new `Namespace` and we will only allow this namespace for our `Schema Manager`. Below is the YAML of the `Namespace` that we are going to create,

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dev
  labels:
    kubernetes.io/metadata.name: dev
```

Let’s save this yaml configuration into `namespace.yaml`. Then create the above `Namespace`,

```bash
$ kubectl apply -f namespace.yaml
namespace/dev created
```


### Deploy Schema Manager

Here, we are going to deploy `Schema Manager` with the new Namespace that we have created above. Let’s deploy it using this following yaml,

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MongoDBDatabase
metadata:
  name: mongodb-schema
  namespace: dev
  labels:
    "schema.kubedb.com": "mongo"
spec:
  database:
    serverRef:
      name: mongodb
      namespace: demo
    config:
      name: emptydb
  vaultRef:
    name: vault
    namespace: demo
  accessPolicy:
    subjects:
      - name: "saname"
        namespace: dev
        kind: "ServiceAccount"
        apiGroup: ""
    defaultTTL: "5m"
    maxTTL: "200h"
  deletionPolicy: Delete
```
Here,

- `spec.database` is a required field specifying the database server reference and the desired database configuration.
- `spec.vaultRef` is a required field that specifies which KubeVault server to use for user management.
- `spec.accessPolicy` is a required field that specifies the access permissions like which service account or cluster user have the access and also for how long they can access through it.
- `spec.deletionPolicy` is an optional field that gives flexibility whether to `nullify` (reject) the delete operation.

Let’s save this yaml configuration into `mongodb-schema.yaml` and apply it,

```bash
$ kubectl apply -f mongodb-schema.yaml
mongodbdatabase.schema.kubedb.com/mongodb-schema created
```

Let's check the `STATUS` of `Schema Manager`,

```bash
$ kubectl get mongodbdatabase -A
NAMESPACE   NAME             DB_SERVER   DB_NAME   STATUS    AGE
dev         mongodb-schema   mongodb     emptydb   Current   54s

```
Here,

> In `STATUS` section, `Current` means that the current `Secret` of `Schema Manager` is vaild, and it will automatically `Expired` after it reaches the limit of `defaultTTL` that we've defined in the above yaml. 

Now, let's get the secret name from `schema-manager`, and get the login credentials for connecting to the database,

```bash
$ kubectl get mongodbdatabase mongodb-schema -n dev -o=jsonpath='{.status.authSecret.name}'
mongodb-schema-mongo-req-fybh8z

$ kubectl view-secret -n dev mongodb-schema-mongo-req-fybh8z -a
password=u-kDmBcMITz9dLrZ7cAL
username=v-kubernetes-demo-k8s-f7695915-1e-0NV83LXHuGMiittiObYE-1662635657
```

### Insert Sample Data

Here, we are going to connect to the database with the login credentials and insert some sample data into it. 

```bash
$ kubectl exec -it -n demo mongodb-0 -c mongodb -- bash
root@mongodb-0:/# mongo --authenticationDatabase=emptydb --username='v-kubernetes-demo-k8s-f7695915-1e-0NV83LXHuGMiittiObYE-1662635657' --password='u-kDmBcMITz9dLrZ7cAL' emptydb
MongoDB shell version v4.4.6
...

replicaset:PRIMARY> use emptydb
switched to db emptydb

replicaset:PRIMARY> db.product.insert({"name":"KubeDB"});
WriteResult({ "nInserted" : 1 })

replicaset:PRIMARY> db.product.find().pretty()
{ "_id" : ObjectId("6319cffeb0d19a8d717b4aee"), "name" : "KubeDB" }

replicaset:PRIMARY> exit
bye

```


Now, Let's check the `STATUS` of `Schema Manager` again,

```bash
$ kubectl get mongodbdatabase -A
NAMESPACE   NAME             DB_SERVER   DB_NAME   STATUS    AGE
dev         mongodb-schema   mongodb     emptydb   Expired   6m
```

Here, we can see that the `STATUS` of the `schema-manager` is `Expired` because it's exceeded `defaultTTL: "5m"`, which means the current `Secret` of `Schema Manager` isn't vaild anymore. Now, if we try to connect and login with the credentials that we have acquired before from `schema-manager`, it won't work.

```bash
$ kubectl exec -it -n demo mongodb-0 -c mongodb -- bash
root@mongodb-0:/# mongo --authenticationDatabase=emptydb --username='v-kubernetes-demo-k8s-f7695915-1e-0NV83LXHuGMiittiObYE-1662635657' --password='u-kDmBcMITz9dLrZ7cAL' emptydb
MongoDB shell version v4.4.6
connecting to: mongodb://127.0.0.1:27017/emptydb?authSource=emptydb&compressors=disabled&gssapiServiceName=mongodb
Error: Authentication failed. :
connect@src/mongo/shell/mongo.js:374:17
@(connect):2:6
exception: connect failed
exiting with code 1
root@mongodb-0:/# exit
exit
```
> Note: We can't connect to the database with the login credentials, which is `Expired`. We will not be able to access the database even though we're in the middle of a connected session. And when the `Schema Manager` is deleted, the associated database and user will also be deleted.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete ns dev
$ kubectl delete ns demo
```


## Next Steps

- Detail concepts of [MongoDBDatabase object](/docs/guides/mongodb/concepts/mongodbdatabase.md).
- Go through the concepts of [KubeVault](https://kubevault.com/docs/latest/guides).
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).