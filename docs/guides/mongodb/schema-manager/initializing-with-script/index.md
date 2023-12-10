---
title: Initializing with Script
menu:
  docs_{{ .version }}:
    identifier: mg-initializing-with-script
    name: Initializing with Script
    parent: mg-schema-manager
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initializing with Script

This guide will show you how to to create database and initialize script with MongoDB `Schema Manager` using `Schema Manager Operator`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB Ops-manager operator` in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/install/stash/).
- Install `KubeVault Ops-manager operator` in your cluster following the steps [here](https://kubevault.com/docs/latest/setup/install/stash/).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mongodb/schema-manager/initializing-with-script/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mongodb/schema-manager/initializing-with-script/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

## Deploy MongoDB Server and Vault Server 

Here, we are going to deploy a `MongoDB Server` by using `KubeDB` operator. Also, we are deploying a `Vault Server` using `KubeVault` Operator.

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
  terminationPolicy: WipeOut
```

Here,

- `spec.version` is the name of the MongoDBVersion CR. Here, we are using MongoDB version `4.4.26`.
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


### Script with ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-script
  namespace: dev
data:
  init.js: |-
    use initdb;
    db.product.insert({"name" : "KubeDB"});
```

```bash
$ kubectl apply -f test-script.yaml 
configmap/test-script created
```


### Deploy Schema Manager Initialize with Script

Here, we are going to deploy `Schema Manager` with the new Namespace that we have created above. Let’s deploy it using this following yaml,

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MongoDBDatabase
metadata:
  name: sample-script
  namespace: dev
  labels:
    "schema.kubedb.com": "mongo"
spec:
  database:
    serverRef:
      name: mongodb
      namespace: demo
    config:
      name: initdb
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
  init:
    initialized: false
    script:
      scriptPath: "/etc/config"
      configMap:
        name: "test-script"
      podTemplate:
        spec:
          containers:
            - env:
              - name: "HAVE_A_TRY"
                value: "whoo! It works"
              name: cnt
              image: nginx
              command:
               - /bin/sh
               - -c
              args:
               - ls
  deletionPolicy: Delete
```
Here,

- `spec.database` is a required field specifying the database server reference and the desired database configuration.
- `spec.vaultRef` is a required field that specifies which KubeVault server to use for user management.
- `spec.accessPolicy` is a required field that specifies the access permissions like which service account or cluster user have the access and also for how long they can access through it.
- `spec.init` is an optional field, containing the information of a script or a snapshot using which the database should be initialized during creation.
- `spec.init.script` refers to the information regarding the .js file which should be used for initialization.
- `spec.init.script.scriptPath` accepts a directory location at which the operator should mount the .js file. 
- `spec.init.script.podTemplate` specifies pod-related details, like environment variables, arguments, images etc.
- `spec.deletionPolicy` is an optional field that gives flexibility whether to `nullify` (reject) the delete operation.

Let’s save this yaml configuration into `sample-script.yaml` and apply it,

```bash
$ kubectl apply -f sample-script.yaml 
mongodbdatabase.schema.kubedb.com/sample-script created
```

Let's check the `STATUS` of `Schema Manager`,

```bash
$ kubectl get mongodbdatabase -A
NAMESPACE   NAME            DB_SERVER   DB_NAME   STATUS    AGE
dev         sample-script   mongodb     initdb    Current   56s
```
Here,

> In `STATUS` section, `Current` means that the current `Secret` of `Schema Manager` is vaild, and it will automatically `Expired` after it reaches the limit of `defaultTTL` that we've defined in the above yaml. 

Now, let's get the secret name from `schema-manager`, and get the login credentials for connecting to the database,

```bash
$ kubectl get mongodbdatabase sample-script -n dev -o=jsonpath='{.status.authSecret.name}'
sample-script-mongo-req-98k0ch

$ kubectl view-secret -n dev sample-script-mongo-req-98k0ch -a
password=-e4v396GFjjjMgPPuU7q
username=v-kubernetes-demo-k8s-f7695915-1e-6sXNTvVpPDtueRQWvoyH-1662641233
```

### Verify Initialization

Here, we are going to connect to the database with the login credentials and verify the database initialization,

```bash
$ kubectl exec -it -n demo mongodb-0 -c mongodb -- bash
root@mongodb-0:/# mongo --authenticationDatabase=initdb --username='v-kubernetes-demo-k8s-f7695915-1e-6sXNTvVpPDtueRQWvoyH-1662641233' --password='-e4v396GFjjjMgPPuU7q' initdb
MongoDB shell version v4.4.26
...

replicaset:PRIMARY> show dbs
initdb  0.000GB

replicaset:PRIMARY> show collections
product

replicaset:PRIMARY> db.product.find()
{ "_id" : ObjectId("6319e46f950868e7b3476cdf"), "name" : "KubeDB" }

replicaset:PRIMARY> exit
bye
```

Now, Let's check the `STATUS` of `Schema Manager` again,

```bash
$ kubectl get mongodbdatabase -A
NAMESPACE   NAME            DB_SERVER   DB_NAME   STATUS    AGE
dev         sample-script   mongodb     initdb    Expired   6m
```

Here, we can see that the `STATUS` of the `schema-manager` is `Expired` because it's exceeded `defaultTTL: "5m"`, which means the current `Secret` of `Schema Manager` isn't vaild anymore. Now, if we try to connect and login with the credentials that we have acquired before from `schema-manager`, it won't work.

```bash
$ kubectl exec -it -n demo mongodb-0 -c mongodb -- bash
root@mongodb-0:/# mongo --authenticationDatabase=initdb --username='v-kubernetes-demo-k8s-f7695915-1e-6sXNTvVpPDtueRQWvoyH-1662641233' --password='-e4v396GFjjjMgPPuU7q' initdb
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
$ kubectl delete ns dev
$ kubectl delete ns demo
```


## Next Steps

- Detail concepts of [MongoDBDatabase object](/docs/guides/mongodb/concepts/mongodbdatabase.md).
- Go through the concepts of [KubeVault](https://kubevault.com/docs/latest/guides).
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).