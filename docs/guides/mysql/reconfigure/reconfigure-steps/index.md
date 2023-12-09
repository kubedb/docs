---
title: Reconfigure MySQL Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-reconfigure-reconfigure-steps
    name: Cluster
    parent: guides-mysql-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure MySQL Cluster Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a MySQL Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
- [MySQL](/docs/guides/mysql/concepts/database/index.md)
- [MySQL Cluster](/docs/guides/mysql/clustering)
- [MYSQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)
- [Reconfigure Overview](/docs/guides/mysql/reconfigure/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

Now, we are going to deploy a  `MySQL` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `MySQLOpsRequest` to reconfigure its configuration.

### Prepare MySQL Cluster

Now, we are going to deploy a `MySQL` Cluster database with version `8.0.31`.

### Deploy MySQL

At first, we will create `my-config.cnf` file containing required configuration settings.

```ini
$ cat my-config.cnf 
[mysqld]
max_connections = 200
read_buffer_size = 1048576
```

Here, `max_connections` is set to `200`, whereas the default value is `151`. Likewise, `read_buffer_size` has the deafult value `131072`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo my-configuration --from-file=./my-config.cnf
secret/my-configuration created
```

In this section, we are going to create a MySQL object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `MySQL` CR that we are going to create,

<ul class="nav nav-tabs" id="definationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="gr-tab" data-toggle="tab" href="#groupReplication" role="tab" aria-controls="groupReplication" aria-selected="true">Group Replication</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="ic-tab" data-toggle="tab" href="#innodbCluster" role="tab" aria-controls="innodbCluster" aria-selected="false">Innodb Cluster</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="sc-tab" data-toggle="tab" href="#semisync" role="tab" aria-controls="semisync" aria-selected="false">Semi sync </a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="st-tab" data-toggle="tab" href="#standAlone" role="tab" aria-controls="standAlone" aria-selected="false">Stand Alone</a>
  </li>
</ul>


<div class="tab-content" id="definationTabContent">
  <div class="tab-pane fade show active" id="groupReplication" role="tabpanel" aria-labelledby="gr-tab">

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.31"
  topology:
    mode: GroupReplication
  replicas: 3
  configSecret:
    name: my-configuration
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure/reconfigure-steps/yamls/group-replication.yaml
mysql.kubedb.com/sample-mysql created
```

  </div>

  <div class="tab-pane fade" id="innodbCluster" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.31-innodb"
  topology:
    mode: InnoDBCluster
    innoDBCluster:
      router:
        replicas: 1
  replicas: 3
  configSecret:
    name: my-configuration
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure/reconfigure-steps/yamls/innob-cluster.yaml
mysql.kubedb.com/sample-mysql created
```

  </div>

  <div class="tab-pane fade " id="semisync" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.31"
  topology:
    mode: SemiSync
    semiSync:
      sourceWaitForReplicaCount: 1
      sourceTimeout: 23h
      errantTransactionRecoveryPolicy: PseudoTransaction
  replicas: 3
  configSecret:
    name: my-configuration
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure/reconfigure-steps/yamls/semi-sync.yaml
mysql.kubedb.com/sample-mysql created
```

  </div>


  <div class="tab-pane fade" id="standAlone" role="tabpanel" aria-labelledby="st-tab">

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.31"
  configSecret:
    name: my-configuration
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure/reconfigure-steps/yamls/stand-alone.yaml
mysql.kubedb.com/sample-mysql created
```
  </div>

</div>


Now, wait until `sample-mysql` has status `Ready`. i.e,

```bash
$ kubectl get mysql -n demo
NAME           VERSION   STATUS   AGE
sample-mysql   8.0.31    Ready    5m49s
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a mysql instance,

```bash
$ kubectl get secrets -n demo sample-mysql-auth -o jsonpath='{.data.\username}' | base64 -d                                                                       
root

$ kubectl get secrets -n demo sample-mysql-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         
86TwLJ!2Kpq*vv1y
```

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mysql-0 -- bash
mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 112
Server version: 8.0.31 MySQL Community Server - GPL

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 200   |
+-----------------+-------+
1 row in set (0.00 sec)

mysql> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 1048576 |
+------------------+---------+
1 row in set (0.00 sec)

mysql> 

```

As we can see from the configuration of ready mysql, the value of `max_connections` has been set to `200` and `read_buffer_size` has been set to `1048576`.

### Reconfigure using new config secret

Now we will reconfigure this database to set `max_connections` to `250` and `read_buffer_size` to `122880`.

Now, we will create new file `new-my-config.cnf` containing required configuration settings.

```ini
$ cat new-my-config.cnf 
[mysqld]
max_connections = 250
read_buffer_size = 122880
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-my-configuration --from-file=./new-my-config.cnf
secret/new-my-configuration created
```

#### Create MySQLOpsRequest

Now, we will use this secret to replace the previous secret using a `MySQLOpsRequest` CR. The `MySQLOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-mysql
  configuration:   
    configSecret:
      name: new-my-configuration
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `sample-mysql` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `MySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure/reconfigure-steps/yamls/reconfigure-using-secret.yaml
mysqlopsrequest.ops.kubedb.com/myops-reconfigure-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `MySQL` object.

Let's wait for `MySQLOpsRequest` to be `Successful`.  Run the following command to watch `MySQLOpsRequest` CR,

```bash
$ kubectl get mysqlopsrequest --all-namespaces
NAMESPACE   NAME                       TYPE          STATUS       AGE
demo        myops-reconfigure-config   Reconfigure   Successful   3m8s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe mysqlopsrequest -n demo myops-reconfigure-config
Name:         myops-reconfigure-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2022-11-23T09:09:20Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:configuration:
          .:
          f:configSecret:
        f:databaseRef:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-11-23T09:09:20Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2022-11-23T09:09:20Z
  Resource Version:  786443
  UID:               253ff2e3-0647-4926-bfb9-ef44b3b8a31d
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-my-configuration
  Database Ref:
    Name:  sample-mysql
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2022-11-23T09:09:20Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-reconfigure-config
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-11-23T09:13:10Z
    Message:               Successfully reconfigured MySQL pod for MySQLOpsRequest: demo/myops-reconfigure-config 
    Observed Generation:   1
    Reason:                SuccessfullyDBReconfigured
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2022-11-23T09:13:10Z
    Message:               Controller has successfully reconfigure the MySQL demo/myops-reconfigure-config
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    30m   KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/myops-reconfigure-config
  Normal  Starting    30m   KubeDB Enterprise Operator  Pausing MySQL databse: demo/sample-mysql
  Normal  Successful  30m   KubeDB Enterprise Operator  Successfully paused MySQL database: demo/sample-mysql for MySQLOpsRequest: myops-reconfigure-config
  Normal  Starting    30m   KubeDB Enterprise Operator  Restarting Pod: sample-mysql-1/demo
  Normal  Starting    29m   KubeDB Enterprise Operator  Restarting Pod: sample-mysql-2/demo
  Normal  Starting    28m   KubeDB Enterprise Operator  Restarting Pod: sample-mysql-0/demo
  Normal  Successful  27m   KubeDB Enterprise Operator  Successfully reconfigured MySQL pod for MySQLOpsRequest: demo/myops-reconfigure-config
  Normal  Starting    27m   KubeDB Enterprise Operator  Reconfiguring MySQL
  Normal  Successful  27m   KubeDB Enterprise Operator  Successfully reconfigure the MySQL object
  Normal  Starting    27m   KubeDB Enterprise Operator  Resuming MySQL database: demo/sample-mysql
  Normal  Successful  27m   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/sample-mysql
  Normal  Successful  27m   KubeDB Enterprise Operator  Controller has Successfully reconfigure the of MySQL: demo/sample-mysql

```

Now let's connect to a mysql instance and run a mysql internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mysql-0 -- bash

bash-4.4# mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}

mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 279
Server version: 8.0.31 MySQL Community Server - GPL

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> 
mysql> 
mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 250   |
+-----------------+-------+
1 row in set (0.00 sec)

mysql> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 122880 |
+------------------+--------+
1 row in set (0.00 sec)

mysql> 

```

As we can see from the configuration has changed, the value of `max_connections` has been changed from `200` to `250` and and the `read_buffer_size` has been changed `1048576` to `122880`. So the reconfiguration of the database is successful.

### Remove Custom Configuration

We can also remove exisiting custom config using `MySQLOpsRequest`. Provide `true` to field `spec.configuration.removeCustomConfig` and make an Ops Request to remove existing custom configuration.

#### Create MySQLOpsRequest

Lets create an `MySQLOpsRequest` having `spec.configuration.removeCustomConfig` is equal `true`,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-mysql
  configuration:   
    removeCustomConfig: true
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `myops-reconfigure-remove` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.removeCustomConfig` is a bool field that should be `true` when you want to remove existing custom configuration.

Let's create the `MySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure/yamls/reconfigure-steps/reconfigure-remove.yaml
mysqlopsrequest.ops.kubedb.com/mdops-reconfigure-remove created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `MySQL` object.

Let's wait for `MySQLOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mysqlopsrequest --all-namespaces
NAMESPACE   NAME                       TYPE          STATUS       AGE
demo        mdops-reconfigure-remove   Reconfigure   Successful   2m1s
```

Now let's connect to a mysql instance and run a mysql internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mysql-0 -- bash
bash-4.4# mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 279
Server version: 8.0.31 MySQL Community Server - GPL

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> 
mysql> 
mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 151   |
+-----------------+-------+
1 row in set (0.00 sec)

mysql> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 131072 |
+------------------+--------+
1 row in set (0.00 sec)

mysql> 

```

As we can see from the configuration has changed to its default value. So removal of existing custom configuration using `MySQLOpsRequest` is successful.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mysql -n demo sample-mysql
$ kubectl delete mysqlopsrequest -n demo myops-reconfigure-config  mdops-reconfigure-remove
$ kubectl delete ns demo
```
