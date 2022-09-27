---
title: MongoDBDatabase
menu:
  docs_{{ .version }}:
    identifier: mongodbdatabase-concepts
    name: MongoDBDatabase
    parent: mg-concepts-mongodb
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDBDatabase

## What is MongoDBDatabase?

`MongoDBDatabase` is a Kubernetes Custom Resource Definitions (CRD). It provides a declarative way of implementing multitenancy inside KubeDB provisioned MongoDB server. You need to describe the target database, desired database configuration, the vault server reference for managing the user in a `MongoDBDatabase` object, and the KubeDB Schema Manager operator will create Kubernetes objects in the desired state for you.

## MongoDBDatabase Specification

As with all other Kubernetes objects, an `MongoDBDatabase` needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `spec` section.

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MongoDBDatabase
metadata:
  name: demo-schema
  namespace: demo
spec:
  database:
    serverRef:
      name: mongodb-server
      namespace: dev
    config: 
      name: myDB
  vaultRef:
    name: vault
    namespace: dev
  accessPolicy:
    subjects:
      - kind: ServiceAccount
        name: "tester"
        namespace: "demo"
    defaultTTL: "10m"
    maxTTL: "200h"
  init: 
    snapshot:
      repository:
        name: repository
        namespace: demo
    script: 
      scriptPath: "etc/config"
      configMap:
        name: scripter
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
  deletionPolicy: "Delete"
```



### spec.database

`spec.database` is a required field specifying the database server reference and the desired database configuration. You need to specify the following fields in `spec.database`,

 - `serverRef` refers to the mongodb instance where the particular schema will be applied.
 - `config` defines the initial configuration of the desired database.

KubeDB accepts the following fields to set in `spec.database`:

 - serverRef:
   - name
   - namespace

 - config:
   - name


### spec.vaultRef

`spec.vaultRef` is a required field that specifies which KubeVault server to use for user management. You need to specify the following fields in `spec.vaultRef`,

- `name` specifies the name of the Vault server.
- `namespace` refers to the namespace where the Vault server is running.


### spec.accessPolicy

`spec.accessPolicy` is a required field that specifies the access permissions like which service account or cluster user have the access and for how long they can access through it. You need to specify the following fields in `spec.accessPolicy`,

- `subjects` refers to the user or service account which is allowed to access the credentials.
- `defaultTTL` specifies for how long the credential would be valid.
- `maxTTL`  specifies the maximum time-to-live for the credentials associated with this role.

KubeDB accepts the following fields to set in `spec.accessPolicy`:

- subjects:
  - kind
  - name
  - namespace

- defaultTTL

- maxTTL


### spec.init

`spec.init` is an optional field, containing the information of a script or a snapshot using which the database should be initialized during creation. You can only specify either script or snapshot fields in `spec.init`,

- `script` refers to the information regarding the .js file which should be used for initialization.
- `snapshot` carries information about the  repository and snapshot_id to initialize the database by restoring the snapshot. 

KubeDB accepts the following fields to set in `spec.init`:

- script:
  - `scriptPath` accepts a directory location at which the operator should mount the .js file.
  - `volumeSource` can be `secret`, `configmap`, `emptyDir`, `nfs`, `persistantVolumeClaim`, `hostPath` etc. The referred volume source should carry the .js file in it. 
  - `podTemplate `specifies pod-related details, like environment variables, arguments, images etc. & lastly there are several `volumeSources` as the subfield of `spec.init.script` like `configMap`, `secret` etc. which actually holds the Mongodb script file.

- snapshot:
  - `repository` refers to the repository cr which carries necessary information about the snapshot location .
  - `snapshotId` refers to the specific snapshot which should be restored . 



### spec.deletionPolicy

`spec.deletionPolicy` is an optional field that gives flexibility whether to `nullify` (reject) the delete operation.


## Next Steps

- Learn about [MongoDB CRD](/docs/guides/mongodb/concepts/mongodb.md)
- Deploy your first MongoDB database with KubeDB by following the guide [here](https://kubedb.com/docs/latest/guides/mongodb/quickstart/quickstart/).