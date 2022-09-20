---
title: Deploy MySQLDatabase
menu:
  docs_{{ .version }}:
    identifier: deploy-mysqldatabase
    name: Deploy MySQLDatabase
    parent: guides-mysql-schema-manager
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Create Database with MySQL Schema Manager

This guide will show you how to create database with MySQL Schema Manager using `KubeDB Enterprise Operator`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB Enterprise Operator` in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/install/enterprise/).
- Install `KubeVault Enterprise Operator` in your cluster following the steps [here](https://kubevault.com/docs/latest/setup/install/enterprise/).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mysql/schema-manager/deploy-mysqldatabase/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/schema-manager/deploy-mysqldatabase/yamls) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

## Deploy MySQL Server and Vault Server 

Firstly, we are going to deploy a `MySQL Server` by using `KubeDB` operator. Also, we are deploying a `Vault Server` using `KubeVault` Operator.

### Deploy MySQL Server

In this section, we are going to deploy a MySQL Server. Let's deploy it using this following yaml,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.0.29"
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
  terminationPolicy: WipeOut
```

Here,

- `spec.version` is the name of the MySQLVersion CR. Here, we are using MySQL version `8.0.29`.
- `spec.storageType` specifies the type of storage that will be used for MySQL. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the MySQL using `EmptyDir` volume.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.
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


### Deploy Schema Manager

Here, we are going to deploy `Schema Manager` with the new `Namespace` that we have created above. Let's deploy it using this following yaml,

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MySQLDatabase
metadata:
  name: schema-manager
  namespace: demox
spec:
  database:
    serverRef:
      name: mysql-server
      namespace: demo
    config:
      name: demo_user
      characterSet: utf8
      encryption: disabled
      readOnly: 0
  vaultRef:
    name: vault
    namespace: demo
  accessPolicy:
    subjects:
      - kind: ServiceAccount
        name: "tester"
        namespace: "demox"
    defaultTTL: "5m"
  deletionPolicy: "Delete"
```
Here,

- `spec.database` is a required field specifying the database server reference and the desired database configuration.
- `spec.vaultRef` is a required field that specifies which KubeVault server to use for user management.
- `spec.accessPolicy` is a required field that specifies the access permissions like which service account or cluster user have the access and also for how long they can access through it.
- `spec.deletionPolicy` is a required field that gives flexibility whether to `nullify` (reject) the delete operation or which resources KubeDB should keep or delete when you delete the CRD.

Let’s save this yaml configuration into `schema-manager.yaml` and apply it,

```bash
$ kubectl apply -f schema-manager.yaml 
mysqldatabase.schema.kubedb.com/schema-manager created
```

Let's check the `STATUS` of `Schema Manager`,

```bash
$ kubectl get mysqldatabase -A
NAMESPACE   NAME             DB_SERVER      DB_NAME     STATUS    AGE
demox       schema-manager   mysql-server   demo_user   Current   27s
```
Here,

> In `STATUS` section, `Current` means that the current `Secret` of `Schema Manager` is vaild, and it will automatically `Expired` after it reaches the limit of `defaultTTL` that we've defined in the above yaml. 

Now, let's get the secret name from `schema-manager`, and get the login credentials for connecting to the database,

```bash
$ kubectl get mysqldatabase schema-manager -n demox -o=jsonpath='{.status.authSecret.name}'
schema-manager-mysql-req-o2j0jk

$ kubectl view-secret schema-manager-mysql-req-o2j0jk -n demox -a
password=bCfsp77bWztyZwH-i4F6
username=v-kubernetes-k8s.dc833e-txGUfwPa
```

### Insert Sample Data

Here, we are going to connect to the database with the login credentials and insert some sample data into it. 

```bash
$ kubectl exec -it mysql-server-0 -n demo -c mysql -- bash
bash-4.4# mysql --user='v-kubernetes-k8s.dc833e-txGUfwPa' --password='bCfsp77bWztyZwH-i4F6'

Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 287
Server version: 8.0.29 MySQL Community Server - GPL


mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| demo_user          |
| information_schema |
+--------------------+
2 rows in set (0.01 sec)

mysql> USE demo_user;
Database changed

mysql> CREATE TABLE random(name varchar(20));
Query OK, 0 rows affected (0.02 sec)

mysql> INSERT INTO random(name) value('KubeDB');
Query OK, 1 row affected (0.00 sec)

mysql> INSERT INTO random(name) value('KubeVault');
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM random;
+-----------+
| name      |
+-----------+
| KubeDB    |
| KubeVault |
+-----------+
2 rows in set (0.00 sec)

mysql> exit
Bye
```


Now, Let's check the `STATUS` of `Schema Manager` again,

```bash
$ kubectl get mysqldatabase -A
NAMESPACE   NAME             DB_SERVER      DB_NAME     STATUS    AGE
demox       schema-manager   mysql-server   demo_user   Expired   5m35s
```

Here, we can see that the `STATUS` of the `schema-manager` is `Expired` because it's exceeded `defaultTTL: "5m"`, which means the current `Secret` of `Schema Manager` isn't vaild anymore. Now, if we try to connect and login with the credentials that we have acquired before from `schema-manager`, it won't work.

```bash
$ kubectl exec -it mysql-server-0 -n demo -c mysql -- bash
bash-4.4# mysql --user='v-kubernetes-k8s.dc833e-txGUfwPa' --password='bCfsp77bWztyZwH-i4F6'
ERROR 1045 (28000): Access denied for user 'v-kubernetes-k8s.dc833e-txGUfwPa'@'localhost' (using password: YES)

mysql> exit
Bye
```
> We can't connect to the database with the login credentials, which is `Expired`. We will not be able to access the database even though we're in the middle of a connected session. 

## Alter Database

In this section, we are going to alter database by changing some characteristics of our database. For this demonstration, We have to logged in as a database admin.

```bash
$ kubectl exec -it mysql-server-0 -n demo -c mysql -- bash
bash-4.4# mysql -uroot -p$MYSQL_ROOT_PASSWORD

Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 2358
Server version: 8.0.29 MySQL Community Server - GPL

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| demo_user          |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.00 sec)

# Check the existing characteristics

mysql> SHOW CREATE DATABASE demo_user;
+-----------+----------------------------------------------------------------------------------------------------------+
| Database  | Create Database                                                                                          |
+-----------+----------------------------------------------------------------------------------------------------------+
| demo_user | CREATE DATABASE `demo_user` /*!40100 DEFAULT CHARACTER SET utf8mb3 */ /*!80016 DEFAULT ENCRYPTION='N' */ |
+-----------+----------------------------------------------------------------------------------------------------------+
1 row in set (0.00 sec)

mysql> exit
bye
```

Let's, change the `spec.database.config.characterSet` to `big5`.

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MySQLDatabase
metadata:
  name: schema-manager
  namespace: demox
spec:
  database:
    serverRef:
      name: mysql-server
      namespace: demo
    config:
      name: demo_user
      characterSet: big5
      encryption: disabled
      readOnly: 0
  vaultRef:
    name: vault
    namespace: demo
  accessPolicy:
    subjects:
      - kind: ServiceAccount
        name: "tester"
        namespace: "demox"
    defaultTTL: "5m"
  deletionPolicy: "Delete"
```

Save this yaml configuration and apply it,

```bash
$ kubectl apply -f schema-manager.yaml
mysqldatabase.schema.kubedb.com/schema-manager configured
```

Now, let's check the modified characteristics of our database.

```bash
$ kubectl exec -it mysql-server-0 -n demo -c mysql -- bash
bash-4.4# mysql -uroot -p$MYSQL_ROOT_PASSWORD

Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 2358
Server version: 8.0.29 MySQL Community Server - GPL

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| demo_user          |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.00 sec)

# Check the existing characteristics

mysql> SHOW CREATE DATABASE demo_user;
+-----------+-------------------------------------------------------------------------------------------------------+
| Database  | Create Database                                                                                       |
+-----------+-------------------------------------------------------------------------------------------------------+
| demo_user | CREATE DATABASE `demo_user` /*!40100 DEFAULT CHARACTER SET big5 */ /*!80016 DEFAULT ENCRYPTION='N' */ |
+-----------+-------------------------------------------------------------------------------------------------------+
1 row in set (0.00 sec)
```
Here, we can see that the `spec.database.config.characterSet` is changed to `big5`. So, our database altering has been successful. 

> Note: When the Schema Manager is deleted, the associated database and user will also be deleted.

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
