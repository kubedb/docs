---
title: Reconfigure MariaDB Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-reconfigure-cluster
    name: Cluster
    parent: guides-mariadb-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfigure MariaDB Cluster Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a MariaDB Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDB Cluster](/docs/guides/mariadb/clustering/galera-cluster)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
  - [Reconfigure Overview](/docs/guides/mariadb/reconfigure/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

Now, we are going to deploy a  `MariaDB` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `MariaDBOpsRequest` to reconfigure its configuration.

### Prepare MariaDB Cluster

Now, we are going to deploy a `MariaDB` Cluster database with version `10.5.8`.

### Deploy MariaDB

At first, we will create `md-config.cnf` file containing required configuration settings.

```ini
$ cat md-config.cnf 
[mysqld]
max_connections = 200
read_buffer_size = 1048576
```

Here, `max_connections` is set to `200`, whereas the default value is `151`. Likewise, `read_buffer_size` has the deafult value `131072`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo md-configuration --from-file=./md-config.cnf
secret/md-configuration created
```

In this section, we are going to create a MariaDB object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `MariaDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.6.4"
  replicas: 3
  configSecret:
    name: md-configuration
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

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/reconfigure/cluster/examples/sample-mariadb-config.yaml
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo 
NAME             VERSION   STATUS   AGE
sample-mariadb   10.5.8    Ready    71s
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a mariadb instance,

```bash
$ kubectl get secrets -n demo sample-mariadb-auth -o jsonpath='{.data.\username}' | base64 -d                                                                       
root

$ kubectl get secrets -n demo sample-mariadb-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         
nrKuxni0wDSMrgwy
```

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 23
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
MariaDB [(none)]> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 200   |
+-----------------+-------+
1 row in set (0.001 sec)

# value of `read_buffer_size` is same as provided
MariaDB [(none)]> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 1048576 |
+------------------+---------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```

As we can see from the configuration of ready mariadb, the value of `max_connections` has been set to `200` and `read_buffer_size` has been set to `1048576`.

### Reconfigure using new config secret

Now we will reconfigure this database to set `max_connections` to `250` and `read_buffer_size` to `122880`.

Now, we will create new file `new-md-config.cnf` containing required configuration settings.

```ini
$ cat new-md-config.cnf 
[mysqld]
max_connections = 250
read_buffer_size = 122880
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-md-configuration --from-file=./new-md-config.cnf
secret/new-md-configuration created
```

#### Create MariaDBOpsRequest

Now, we will use this secret to replace the previous secret using a `MariaDBOpsRequest` CR. The `MariaDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-mariadb
  configuration:   
    configSecret:
      name: new-md-configuration
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `sample-mariadb` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/reconfigure/cluster/examples/reconfigure-using-secret.yaml
mariadbopsrequest.ops.kubedb.com/mdops-reconfigure-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `MariaDB` object.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest --all-namespaces
NAMESPACE   NAME                       TYPE          STATUS       AGE
demo        mdops-reconfigure-config   Reconfigure   Successful   3m8s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe mariadbopsrequest -n demo mdops-reconfigure-config
Name:         mdops-reconfigure-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2022-06-10T04:43:50Z
  Generation:          1
  Resource Version:  1123451
  UID:               27a73fc6-1d25-4019-8975-f7d4daf782b7
Spec:
  Configuration:
    Config Secret:
      Name:  new-md-configuration
  Database Ref:
    Name:  sample-mariadb
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2022-06-10T04:43:50Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/mdops-reconfigure-config
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-06-10T04:47:25Z
    Message:               Successfully restarted MariaDB pods for MariaDBOpsRequest: demo/mdops-reconfigure-config
    Observed Generation:   1
    Reason:                SuccessfullyRestatedStatefulSet
    Status:                True
    Type:                  RestartStatefulSetPods
    Last Transition Time:  2022-06-10T04:47:30Z
    Message:               Successfully reconfigured MariaDB for MariaDBOpsRequest: demo/mdops-reconfigure-config
    Observed Generation:   1
    Reason:                SuccessfullyDBReconfigured
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-06-10T04:47:30Z
    Message:               Controller has successfully reconfigure the MariaDB demo/mdops-reconfigure-config
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful

```

Now let's connect to a mariadb instance and run a mariadb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 23
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
MariaDB [(none)]> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 250   |
+-----------------+-------+
1 row in set (0.001 sec)

# value of `read_buffer_size` is same as provided
MariaDB [(none)]> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 122880  |
+------------------+---------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```

As we can see from the configuration has changed, the value of `max_connections` has been changed from `200` to `250` and so as the `read_buffer_size`. So the reconfiguration of the database is successful.


### Reconfigure Existing Config Secret

Now, we will create a new `MariaDBOpsRequest` to reconfigure our existing secret `new-md-configuration` by modifying our `new-md-config.cnf` file using `applyConfig`. The `MariaDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-reconfigure-apply-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-mariadb
  configuration:   
    applyConfig:
      new-md-config.cnf: |
        [mysqld]
        max_connections = 230
        read_buffer_size = 1068487
```
> Note: You can modify any field of your current configuration using `applyConfig`. If you don't have any secrets then `applyConfig` will create a secret for you.

Here,
- `spec.databaseRef.name` specifies that we are reconfiguring `sample-mariadb` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig.new-md-config.cnf` specifies configuration of the existing or new secret.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/reconfigure/cluster/examples/mdops-reconfigure-apply-config.yaml
mariadbopsrequest.ops.kubedb.com/mdops-reconfigure-apply-config created
```


#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `MariaDB` object.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest mdops-reconfigure-apply-config -n demo
NAME                             TYPE          STATUS       AGE
mdops-reconfigure-apply-config   Reconfigure   Successful   4m59s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe mariadbopsrequest -n demo mdops-reconfigure-apply-config
Name:         mdops-reconfigure-apply-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2022-06-10T06:33:01Z
  Generation:          1
  Resource Version:  1135158
  UID:               8b2bd634-9fed-49c1-91cf-a117afb18b32
Spec:
  Configuration:
    Apply Config:
      new-md-config.cnf:  [mysqld]
max_connections = 230
read_buffer_size = 1068487

  Database Ref:
    Name:  sample-mariadb
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2022-06-10T06:33:01Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/mdops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-06-10T06:33:01Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareSecureCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2022-06-10T06:36:36Z
    Message:               Successfully restarted MariaDB pods for MariaDBOpsRequest: demo/mdops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                SuccessfullyRestatedStatefulSet
    Status:                True
    Type:                  RestartStatefulSetPods
    Last Transition Time:  2022-06-10T06:36:41Z
    Message:               Successfully reconfigured MariaDB for MariaDBOpsRequest: demo/mdops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                SuccessfullyDBReconfigured
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-06-10T06:36:41Z
    Message:               Controller has successfully reconfigure the MariaDB demo/mdops-reconfigure-apply-config
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
```

Now let's connect to a mariadb instance and run a mariadb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 23
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
MariaDB [(none)]> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 230   |
+-----------------+-------+
1 row in set (0.001 sec)

# value of `read_buffer_size` is same as provided
MariaDB [(none)]> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 1064960 |
+------------------+---------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```

As we can see from above the configuration has changed, the value of `max_connections` has been changed from `250` to `230` and the `read_buffer_size` has been changed `122880` to `1064960`. So the reconfiguration of the `sample-mariadb` database is successful.


### Remove Custom Configuration

We can also remove exisiting custom config using `MariaDBOpsRequest`. Provide `true` to field `spec.configuration.removeCustomConfig` and make an Ops Request  to remove existing custom configuration.

#### Create MariaDBOpsRequest

Lets create an `MariaDBOpsRequest` having `spec.configuration.removeCustomConfig` is equal `true`,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-mariadb
  configuration:   
    removeCustomConfig: true
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mdops-reconfigure-remove` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.removeCustomConfig` is a bool field that should be `true` when you want to remove existing custom configuration.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/reconfigure/cluster/examples/reconfigure-remove.yaml
mariadbopsrequest.ops.kubedb.com/mdops-reconfigure-remove created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `MariaDB` object.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest --all-namespaces
NAMESPACE   NAME                       TYPE          STATUS       AGE
demo        mdops-reconfigure-remove   Reconfigure   Successful   3m8s
```

Now let's connect to a mariadb instance and run a mariadb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 23
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
MariaDB [(none)]> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 151   |
+-----------------+-------+
1 row in set (0.001 sec)

# value of `read_buffer_size` is same as provided
MariaDB [(none)]> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 131072  |
+------------------+---------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```

As we can see from the configuration has changed to its default value. So removal of existing custom configuration using `MariaDBOpsRequest` is successful.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
$ kubectl delete mariadbopsrequest -n demo mdops-reconfigure-config mdops-reconfigure-apply-config mdops-reconfigure-remove
$ kubectl delete ns demo
```
