---
title: Reconfigure PerconaXtraDB Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-reconfigure-cluster
    name: Cluster
    parent: guides-perconaxtradb-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfigure PerconaXtraDB Cluster Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a PerconaXtraDB Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/perconaxtradb/concepts/perconaxtradb)
  - [PerconaXtraDB Cluster](/docs/guides/perconaxtradb/clustering/galera-cluster)
  - [PerconaXtraDBOpsRequest](/docs/guides/perconaxtradb/concepts/opsrequest)
  - [Reconfigure Overview](/docs/guides/perconaxtradb/reconfigure/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

Now, we are going to deploy a  `PerconaXtraDB` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `PerconaXtraDBOpsRequest` to reconfigure its configuration.

### Prepare PerconaXtraDB Cluster

Now, we are going to deploy a `PerconaXtraDB` Cluster database with version `10.6.4`.

### Deploy PerconaXtraDB

At first, we will create `px-config.cnf` file containing required configuration settings.

```ini
$ cat px-config.cnf 
[mysqld]
max_connections = 200
read_buffer_size = 1048576
```

Here, `max_connections` is set to `200`, whereas the default value is `151`. Likewise, `read_buffer_size` has the deafult value `131072`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo px-configuration --from-file=./px-config.cnf
secret/px-configuration created
```

In this section, we are going to create a PerconaXtraDB object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
  configSecret:
    name: px-configuration
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

Let's create the `PerconaXtraDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure/cluster/examples/sample-pxc-config.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo 
NAME             VERSION   STATUS   AGE
sample-pxc       8.0.26    Ready    71s
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a perconaxtradb instance,

```bash
$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\username}' | base64 -d                                                                       
root

$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         
nrKuxni0wDSMrgwy
```

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3699
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

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


# value of `read_buffer_size` is same as provided 
mysql> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 1048576 |
+------------------+--------+
1 row in set (0.00 sec)


mysql> exit
Bye
```

As we can see from the configuration of ready perconaxtradb, the value of `max_connections` has been set to `200` and `read_buffer_size` has been set to `1048576`.

### Reconfigure using new config secret

Now we will reconfigure this database to set `max_connections` to `250` and `read_buffer_size` to `122880`.

Now, we will create new file `new-px-config.cnf` containing required configuration settings.

```ini
$ cat new-px-config.cnf 
[mysqld]
max_connections = 250
read_buffer_size = 122880
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-px-configuration --from-file=./new-px-config.cnf
secret/new-px-configuration created
```

#### Create PerconaXtraDBOpsRequest

Now, we will use this secret to replace the previous secret using a `PerconaXtraDBOpsRequest` CR. The `PerconaXtraDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-pxc
  configuration:   
    configSecret:
      name: new-px-configuration
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `sample-pxc` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure/cluster/examples/reconfigure-using-secret.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-reconfigure-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `PerconaXtraDB` object.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ kubectl get perconaxtradbopsrequest --all-namespaces
NAMESPACE   NAME                       TYPE          STATUS       AGE
demo        pxops-reconfigure-config   Reconfigure   Successful   3m8s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe perconaxtradbopsrequest -n demo pxops-reconfigure-config
Name:         pxops-reconfigure-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PerconaXtraDBOpsRequest
Metadata:
  Creation Timestamp:  2022-06-10T04:43:50Z
  Generation:          1
  Resource Version:  1123451
  UID:               27a73fc6-1d25-4019-8975-f7d4daf782b7
Spec:
  Configuration:
    Config Secret:
      Name:  new-px-configuration
  Database Ref:
    Name:  sample-pxc
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2022-06-10T04:43:50Z
    Message:               Controller has started to Progress the PerconaXtraDBOpsRequest: demo/pxops-reconfigure-config
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-06-10T04:47:25Z
    Message:               Successfully restarted PerconaXtraDB pods for PerconaXtraDBOpsRequest: demo/pxops-reconfigure-config
    Observed Generation:   1
    Reason:                SuccessfullyRestatedStatefulSet
    Status:                True
    Type:                  RestartStatefulSetPods
    Last Transition Time:  2022-06-10T04:47:30Z
    Message:               Successfully reconfigured PerconaXtraDB for PerconaXtraDBOpsRequest: demo/pxops-reconfigure-config
    Observed Generation:   1
    Reason:                SuccessfullyDBReconfigured
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-06-10T04:47:30Z
    Message:               Controller has successfully reconfigure the PerconaXtraDB demo/pxops-reconfigure-config
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful

```

Now let's connect to a perconaxtradb instance and run a perconaxtradb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3699
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 250   |
+-----------------+-------+
1 row in set (0.00 sec)


# value of `read_buffer_size` is same as provided 
mysql> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 122880 |
+------------------+--------+
1 row in set (0.00 sec)


mysql> exit
Bye
```


As we can see from the configuration has changed, the value of `max_connections` has been changed from `200` to `250` and and the `read_buffer_size` has been changed `1048576` to `122880`. So the reconfiguration of the database is successful.


### Reconfigure Existing Config Secret

Now, we will create a new `PerconaXtraDBOpsRequest` to reconfigure our existing secret `new-px-configuration` by modifying our `new-px-config.cnf` file using `applyConfig`. The `PerconaXtraDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-reconfigure-apply-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-pxc
  configuration:   
    applyConfig:
      new-px-config.cnf: |
        [mysqld]
        max_connections = 230
        read_buffer_size = 1064960
      innodb-config.cnf: |
        [mysqld]
        innodb_log_buffer_size = 17408000
```
> Note: You can modify multiple fields of your current configuration using `applyConfig`. If you don't have any secrets then `applyConfig` will create a secret for you. Here, we modified value of our two existing fields which are `max_connections` and `read_buffer_size` also, we modified a new field `innodb_log_buffer_size` of our configuration. 

Here,
- `spec.databaseRef.name` specifies that we are reconfiguring `sample-pxc` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` contains the configuration of existing or newly created secret.

Before applying this yaml we are going to check the existing value of our new field,

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3699
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like 'innodb_log_buffer_size';
+------------------------+----------+
| Variable_name          | Value    |
+------------------------+----------+
| innodb_log_buffer_size | 16777216 |
+------------------------+----------+
1 row in set (0.00 sec)

PerconaXtraDB [(none)]> exit
Bye
```

16777216

Here, we can see the default value for `innodb_log_buffer_size` is `16777216`. 

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure/cluster/examples/pxops-reconfigure-apply-config.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-reconfigure-apply-config created
```


#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `PerconaXtraDB` object.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ kubectl get perconaxtradbopsrequest pxops-reconfigure-apply-config -n demo
NAME                             TYPE          STATUS       AGE
pxops-reconfigure-apply-config   Reconfigure   Successful   4m59s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe perconaxtradbopsrequest -n demo pxops-reconfigure-apply-config
Name:         pxops-reconfigure-apply-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PerconaXtraDBOpsRequest
Metadata:
  Creation Timestamp:  2022-06-10T09:13:49Z
  Generation:          1
  Resource Version:  14120
  UID:               eb8d5df5-a0ce-4011-890c-c18c0200b5ac
Spec:
  Configuration:
    Apply Config:
      innodb-config.cnf:  [mysqld]
innodb_log_buffer_size = 17408000

      new-px-config.cnf:  [mysqld]
max_connections = 230
read_buffer_size = 1064960

  Database Ref:
    Name:  sample-pxc
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2022-06-10T09:13:49Z
    Message:               Controller has started to Progress the PerconaXtraDBOpsRequest: demo/pxops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-06-10T09:13:49Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareSecureCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2022-06-10T09:17:24Z
    Message:               Successfully restarted PerconaXtraDB pods for PerconaXtraDBOpsRequest: demo/pxops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                SuccessfullyRestatedStatefulSet
    Status:                True
    Type:                  RestartStatefulSetPods
    Last Transition Time:  2022-06-10T09:17:29Z
    Message:               Successfully reconfigured PerconaXtraDB for PerconaXtraDBOpsRequest: demo/pxops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                SuccessfullyDBReconfigured
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-06-10T09:17:29Z
    Message:               Controller has successfully reconfigure the PerconaXtraDB demo/pxops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
```

Now let's connect to a perconaxtradb instance and run a perconaxtradb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3699
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
# value of `max_conncetions` is same as provided 
mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 230   |
+-----------------+-------+
1 row in set (0.001 sec)

# value of `read_buffer_size` is same as provided
mysql> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 1064960 |
+------------------+---------+
1 row in set (0.001 sec)

# value of `innodb_log_buffer_size` is same as provided
mysql> show variables like 'innodb_log_buffer_size';
+------------------------+----------+
| Variable_name          | Value    |
+------------------------+----------+
| innodb_log_buffer_size | 17408000 |
+------------------------+----------+
1 row in set (0.001 sec)

mysql> exit
Bye
```

As we can see from above the configuration has been changed, the value of `max_connections` has been changed from `250` to `230` and the `read_buffer_size` has been changed `122880` to `1064960` also, `innodb_log_buffer_size` has been changed from `16777216` to `17408000`. So the reconfiguration of the `sample-pxc` database is successful.


### Remove Custom Configuration

We can also remove existing custom config using `PerconaXtraDBOpsRequest`. Provide `true` to field `spec.configuration.removeCustomConfig` and make an Ops Request to remove existing custom configuration.

#### Create PerconaXtraDBOpsRequest

Lets create an `PerconaXtraDBOpsRequest` having `spec.configuration.removeCustomConfig` is equal `true`,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-pxc
  configuration:   
    removeCustomConfig: true
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `pxops-reconfigure-remove` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.removeCustomConfig` is a bool field that should be `true` when you want to remove existing custom configuration.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure/cluster/examples/reconfigure-remove.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-reconfigure-remove created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `PerconaXtraDB` object.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ kubectl get perconaxtradbopsrequest --all-namespaces
NAMESPACE   NAME                       TYPE          STATUS       AGE
demo        pxops-reconfigure-remove   Reconfigure   Successful   2m1s
```

Now let's connect to a perconaxtradb instance and run a perconaxtradb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3699
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_connections` is default
PerconaXtraDB [(none)]> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 151   |
+-----------------+-------+
1 row in set (0.001 sec)

# value of `read_buffer_size` is default
PerconaXtraDB [(none)]> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 131072  |
+------------------+---------+
1 row in set (0.001 sec)

# value of `innodb_log_buffer_size` is default
PerconaXtraDB [(none)]> show variables like 'innodb_log_buffer_size';
+------------------------+----------+
| Variable_name          | Value    |
+------------------------+----------+
| innodb_log_buffer_size | 16777216 |
+------------------------+----------+
1 row in set (0.001 sec)

PerconaXtraDB [(none)]> exit
Bye
```

As we can see from the configuration has changed to its default value. So removal of existing custom configuration using `PerconaXtraDBOpsRequest` is successful.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
$ kubectl delete perconaxtradbopsrequest -n demo pxops-reconfigure-config pxops-reconfigure-apply-config pxops-reconfigure-remove
$ kubectl delete ns demo
```
