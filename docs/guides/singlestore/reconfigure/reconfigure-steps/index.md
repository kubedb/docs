---
title: Reconfigure SingleStore Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-reconfigure-reconfigure-steps
    name: Cluster
    parent: guides-sdb-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure SingleStore Cluster Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a MySQL Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
- [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
- [SingleStore Cluster](/docs/guides/singlestore/clustering)
- [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
- [Reconfigure Overview](/docs/guides/singlestore/reconfigure/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

Now, we are going to deploy a  `SingleStore` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `SingleStoreOpsRequest` to reconfigure its configuration.

## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## Deploy SingleStore

At first, we will create `sdb-config.cnf` file containing required configuration settings.

```ini
$ cat sdb-config.cnf 
[server]
max_connections = 250
read_buffer_size = 122880

```

Here, `max_connections` is set to `250`, whereas the default value is `100000`. Likewise, `read_buffer_size` has the deafult value `131072`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo sdb-configuration --from-file=./sdb-config.cnf
secret/sdb-configuration created
```

In this section, we are going to create a SingleStore object specifying `spec.topology.aggreagtor.configSecret` field to apply this custom configuration. Below is the YAML of the `SingleStore` CR that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: custom-sdb
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      configSecret:
        name: sdb-configuration
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      configSecret:
        name: sdb-configuration
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure/reconfigure-steps/yamls/custom-sdb.yaml
singlestore.kubedb.com/custom-sdb created


Now, wait until `custom-sdb` has status `Ready`. i.e,

```bash
$ kubectl get pod -n demo
NAME                      READY   STATUS    RESTARTS   AGE
custom-sdb-aggregator-0   2/2     Running   0          94s
custom-sdb-aggregator-1   2/2     Running   0          88s
custom-sdb-leaf-0         2/2     Running   0          91s
custom-sdb-leaf-1         2/2     Running   0          86s

$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
custom-sdb   kubedb.com/v1alpha2   8.7.10    Ready    4m29s
```

We can see the database is in ready phase so it can accept conncetion.

Now, we will check if the database has started with the custom configuration we have provided.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# Connceting to the database
$ kubectl exec -it -n demo custom-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@custom-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 208
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
singlestore> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 250   |
+-----------------+-------+
1 row in set (0.00 sec)

# value of `read_buffer_size` is same as provided
singlestore> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 122880 |
+------------------+--------+
1 row in set (0.00 sec)

singlestore> exit
Bye

```

As we can see from the configuration of ready mysql, the value of `max_connections` has been set to `250` and `read_buffer_size` has been set to `122880`.

### Reconfigure using new config secret

Now we will reconfigure this database to set `max_connections` to `350` and `read_buffer_size` to `132880`.

Now, we will create new file `new-sdb-config.cnf` containing required configuration settings.

#### Create SingleStoreOpsRequest

Now, we will use this secret to replace the previous secret using a `SingleStoreOpsRequest` CR. The `SingleStoreOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-config
  namespace: demo
spec:
  type: Configuration
  databaseRef:
    name: custom-sdb
  configuration:
    aggregator:
      applyConfig:
        sdb-apply.cnf: |-
          max_connections = 550
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `sample-mysql` database.
- `spec.type` specifies that we are performing `Configuration` on our database.
- `spec.aggregator.configSecret.name` specifies the name of the new secret for aggregator nodes. You can also specifies `spec.leaf.configSecret.name` the name of the new secret for leaf nodes.

Let's create the `SinglestoreOpsRequest` CR we have shown above,

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
Server version: 8.0.35 MySQL Community Server - GPL

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
Server version: 8.0.35 MySQL Community Server - GPL

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
