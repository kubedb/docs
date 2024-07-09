---
title: Initializing with Script
menu:
  docs_{{ .version }}:
    identifier: mysql-initializing-with-script
    name: Initializing with Script
    parent: guides-mysql-schema-manager
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initializing with Script

This guide will show you how to to create database and initialize Script with MySQL `Schema Manager` using KubeDB Ops Manager.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeVault` in your cluster following the steps [here](https://kubevault.com/docs/latest/setup/install/kubevault/).

- You should be familiar with the following `KubeDB` and `KubeVault` concepts:
  - [MySQL](/docs/guides/mysql/concepts/database/index.md)
  - [MySQLDatabase](/docs/guides/mysql/concepts/mysqldatabase/index.md)
  - [Schema Manager Overview](/docs/guides/mysql/schema-manager/overview/index.md)
  - [KubeVault Overview](https://kubevault.com/docs/latest/concepts/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mysql/schema-manager/initializing-with-script/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/schema-manager/initializing-with-script/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

## Deploy MySQL Server and Vault Server 

Here, we are going to deploy a `MySQL Server` by using `KubeDB` operator. Also, we are deploying a `Vault Server` using `KubeVault` Operator.

### Deploy MySQL Server

In this section, we are going to deploy a MySQL Server. Let's deploy it using this following yaml,

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.0.35"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 200Mi
  allowedSchemas:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          app: schemaManager
  deletionPolicy: WipeOut
```

Here,

- `spec.version` is the name of the MySQLVersion CR. Here, we are using MySQL version `8.0.35`.
- `spec.storageType` specifies the type of storage that will be used for MySQL. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the MySQL using `EmptyDir` volume.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.allowedSchemas` specifies the namespace of allowed `Schema Manager`.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete the operation of MySQL CR. *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy).


Let’s save this yaml configuration into `mysql-server.yaml` Then create the above `MySQL` CR

```bash
$ kubectl apply -f mysql-server.yaml
mysql.kubedb.com/mysql-server created
```

### Deploy Vault Server

In this section, we are going to deploy a Vault Server. Let's deploy it using this following yaml,

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: 1.9.2
  replicas: 1
  allowedSecretEngines:
    namespaces:
      from: All
    secretEngines:
      - mysql
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

In this section, we are going to create a new `Namespace` and we will only allow this namespace for our `Schema Manager`. Let's deploy it using this following yaml,

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: demox
  labels:
    app: schemaManager
```

Let’s save this yaml configuration into `namespace.yaml` Then create the above `Namespace`

```bash
$ kubectl apply -f namespace.yaml
namespace/demox created
```

### SQL Script with ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scripter
  namespace:  demox
data:
  script.sql: |-
    use demo_script;
    create table Product(Name varchar(50),Title varchar(50));
    insert into Product(Name,Title) value('KubeDB','Database Management Solution');
    insert into Product(Name,Title) value('Stash','Backup and Recovery Solution');
```

```bash
$ kubectl apply -f configmap.yaml 
configmap/scripter created

```


### Deploy Schema Manager Initialize with Script

Here, we are going to deploy `Schema Manager` with the new `Namespace` that we have created above. Let's deploy it using this following yaml,

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MySQLDatabase
metadata:
  name: schema-script
  namespace: demox
spec:
  database:
    serverRef:
      name: mysql-server
      namespace: demo
    config: 
      name: demo_script
  vaultRef:
    name: vault
    namespace: demo
  accessPolicy:
    subjects:
      - kind: ServiceAccount
        name: "script-tester"
        namespace: "demox"
    defaultTTL: "5m"
  init: 
    initialized: false
    script: 
      scriptPath: "etc/config"
      configMap:
        name: scripter
  deletionPolicy: "Delete"
```
Here,

- `spec.database` is a required field specifying the database server reference and the desired database configuration.
- `spec.vaultRef` is a required field that specifies which KubeVault server to use for user management.
- `spec.accessPolicy` is a required field that specifies the access permissions like which service account or cluster user have the access and also for how long they can access through it.
- `spec.init` is an optional field, containing the information of a script or a snapshot using which the database should be initialized during creation.
- `spec.deletionPolicy` is a required field that gives flexibility whether to `nullify` (reject) the delete operation or which resources KubeDB should keep or delete when you delete the CRD.

Let’s save this yaml configuration into `schema-manager.yaml` and apply it,

```bash
$ kubectl apply -f schema-script.yaml
mysqldatabase.schema.kubedb.com/schema-script created
```

Let's check the `STATUS` of `Schema Manager`,

```bash
$ kubectl get mysqldatabase -A
NAMESPACE   NAME            DB_SERVER      DB_NAME       STATUS    AGE
demox       schema-script   mysql-server   demo_script   Current   21s
```
Here,

> In `STATUS` section, `Current` means that the current `Secret` of `Schema Manager` is vaild, and it will automatically `Expired` after it reaches the limit of `defaultTTL` that we've defined in the above yaml. 

Now, let's get the secret name from `schema-manager`, and get the login credentials for connecting to the database,

```bash
$ kubectl get mysqldatabase schema-script -n demox -o=jsonpath='{.status.authSecret.name}'
schema-script-mysql-req-s85fuw

$ kubectl view-secret schema-script-mysql-req-s85fuw -n demox -a
password=DueiiR-JyGpa3rejG2Zd
username=v-kubernetes-k8s.dc833e-yb9r7uhs
```

### Verify Initialization

Here, we are going to connect to the database with the login credentials and verify the database initialization, 

```bash
$ kubectl exec -it mysql-server-0 -n demo -c mysql -- bash
bash-4.4# mysql --user='v-kubernetes-k8s.dc833e-yb9r7uhs' --password='DueiiR-JyGpa3rejG2Zd'

Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 287
Server version: 8.0.35 MySQL Community Server - GPL

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| demo_script        |
| information_schema |
+--------------------+
2 rows in set (0.00 sec)

mysql> USE demo_script;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
mysql> SHOW TABLES;
+-----------------------+
| Tables_in_demo_script |
+-----------------------+
| Product               |
+-----------------------+
1 row in set (0.00 sec)

mysql> SELECT * FROM Product;
+--------+------------------------------+
| Name   | Title                        |
+--------+------------------------------+
| KubeDB | Database Management Solution |
| Stash  | Backup and Recovery Solution |
+--------+------------------------------+
2 rows in set (0.00 sec)

mysql> exit
Bye
```


Now, Let's check the `STATUS` of `Schema Manager` again,

```bash
$ kubectl get mysqldatabase -A
NAMESPACE   NAME            DB_SERVER      DB_NAME       STATUS    AGE
demox       schema-script   mysql-server   demo_script   Expired   5m27s
```

Here, we can see that the `STATUS` of the `schema-manager` is `Expired` because it's exceeded `defaultTTL: "5m"`, which means the current `Secret` of `Schema Manager` isn't vaild anymore. Now, if we try to connect and login with the credentials that we have acquired before from `schema-manager`, it won't work.

```bash
$ kubectl exec -it mysql-server-0 -n demo -c mysql -- bash
bash-4.4# mysql --user='v-kubernetes-k8s.dc833e-yb9r7uhs' --password='DueiiR-JyGpa3rejG2Zd'
ERROR 1045 (28000): Access denied for user 'v-kubernetes-k8s.dc833e-txGUfwPa'@'localhost' (using password: YES)

mysql> exit
Bye
```
> We can't connect to the database with the login credentials, which is `Expired`. We will not be able to access the database even though we're in the middle of a connected session. 



## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete ns demox 
$ kubectl delete ns demo
```


## Next Steps

- Detail concepts of [MySQLDatabase object](/docs/guides/mysql/concepts/mysqldatabase/index.md).
- Go through the concepts of [KubeVault](https://kubevault.com/docs/latest/guides).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLVersion object](/docs/guides/mysql/concepts/catalog/index.md).
