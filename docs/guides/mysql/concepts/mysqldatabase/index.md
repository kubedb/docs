---
title: MySQLDatabase
menu:
  docs_{{ .version }}:
    identifier: mysqldatabase-concepts
    name: MySQLDatabase
    parent: guides-mysql-concepts
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQLDatabase

## What is MySQLDatabase ?
`MySQLDatabase` is a Kubernetes Custom Resource Definitions (CRD). It provides a declarative way of implementing multitenancy inside KubeDB provisioned MySQL server. You need to describe the target database, desired database configuration, the vault server reference for managing the user in a `MySQLDatabase` object, and the KubeDB Schema Manager operator will create Kubernetes objects in the desired state for you.

## MySQLDatabase Specification

As with all other Kubernetes objects, an `MySQLDatabase` needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `spec` section.

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MySQLDatabase
metadata:
  name: demo-schema
  namespace: demo
spec:
  database:
    serverRef:
      name: mysql-server
      namespace: dev
    config: 
      name: myDB
      characterSet: big5
      encryption: disable
      readOnly: 0
  vaultRef:
    name: vault
    namespace: dev
  accessPolicy:
    subjects:
      - kind: ServiceAccount
        name: "tester"
        namespace: "demo"
    defaultTTL: "10m"
  init: 
    initialized: false
    snapshot:
      repository:
        name: repository
        namespace: demo
    script: 
      scriptPath: "etc/config"
      configMap:
        name: scripter
  deletionPolicy: "Delete"
```



### spec.database

`spec.database` is a required field specifying the database server reference and the desired database configuration. You need to specify the following fields in `spec.database`,

 - `serverRef` refers to the mysql instance where the particular schema will be applied.
 - `config` defines the initial configuration of the desired database.

KubeDB accepts the following fields to set in `spec.database`:

 - serverRef:
   - name
   - namespace

 - config:
   - name
   - characterSet
   - encryption
   - readOnly


### spec.vaultRef

`spec.vaultRef` is a required field that specifies which KubeVault server to use for user management. You need to specify the following fields in `spec.vaultRef`,

- `name` specifies the name of the Vault server.
- `namespace` refers to the namespace where the Vault server is running.


### spec.accessPolicy

`spec.accessPolicy` is a required field that specifies the access permissions like which service account or cluster user have the access and for how long they can access through it. You need to specify the following fields in `spec.accessPolicy`,

- `subjects` refers to the user or service account which is allowed to access the credentials.
- `defaultTTL` specifies for how long the credential would be valid.

KubeDB accepts the following fields to set in `spec.accessPolicy`:

- subjects:
  - kind
  - name
  - namespace

- defaultTTL


### spec.init

`spec.init` is an optional field, containing the information of a script or a snapshot using which the database should be initialized during creation. You need to specify the following fields in `spec.init`,

- `script` refers to the information regarding the .sql file which should be used for initialization.
- `snapshot` carries information about the  repository and snapshot_id to initialize the database by restoring the snapshot. 

KubeDB accepts the following fields to set in `spec.init`:

- script:
  - `scriptPath` accepts a directory location at which the operator should mount the .sql file.
  - `volumeSource` this can be either secret or configmap. The referred volume source should carry the .sql file in it. 

- snapshot:
  - `repository` refers to the repository cr which carries necessary information about the snapshot location .
  - `snapshotId` refers to the specific snapshot which should be restored . 



### spec.deletionPolicy

`spec.deletionPolicy` is a required field that gives flexibility whether to `nullify` (reject) the delete operation or which resources KubeDB should keep or delete when you delete the CRD.



## Next Steps

- Learn about MySQL CRD [here](/docs/guides/mysql/concepts/database/index.md).
- Deploy your first MySQL database with KubeDB by following the guide [here](https://kubedb.com/docs/latest/guides/mysql/quickstart/).