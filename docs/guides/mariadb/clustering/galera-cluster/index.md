---
title: MariaDB Galera Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-clustering-galeracluster
    name: MariaDB Galera Cluster Guide
    parent: guides-mariadb-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MariaDB Cluster

This tutorial will show you how to use KubeDB to provision a MariaDB replication group in single-primary mode.

## Before You Begin

Before proceeding:

- Read [mariadb galera cluster concept](/docs/guides/mariadb/clustering/overview) to learn about MariaDB Group Replication.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MariaDB Cluster

The following is an example `MariaDB` object which creates a multi-master MariaDB group with three members.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/clustering/galera-cluster/examples/demo-1.yaml
mariadb.kubedb.com/sample-mariadb created
```

Here,

- `spec.replicas` is the number of nodes in the cluster.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `MariaDB` objects using Kubernetes API. When a `MariaDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MariaDB object name. KubeDB operator will also create a governing service for the StatefulSet with the name `<mariadb-object-name>-pods`.

```bash
$ kubectl get mariadb -n demo sample-mariadb -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"sample-mariadb","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","version":"10.5.23"}}
  creationTimestamp: "2021-03-16T09:39:01Z"
  finalizers:
  - kubedb.com
  generation: 2
  managedFields:
    ...
  name: sample-mariadb
  namespace: demo
spec:
  authSecret:
    name: sample-mariadb-auth
  podTemplate:
    ...
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
  version: 10.5.23
status:
  conditions:
  - lastTransitionTime: "2021-03-16T09:39:01Z"
    message: 'The KubeDB operator has started the provisioning of MariaDB: demo/sample-mariadb'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2021-03-16T09:40:00Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2021-03-16T09:39:09Z"
    message: 'The MariaDB: demo/sample-mariadb is accepting client requests.'
    observedGeneration: 2
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2021-03-16T09:39:50Z"
    message: 'The MySQL: demo/sample-mariadb is ready.'
    observedGeneration: 2
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2021-03-16T09:40:00Z"
    message: 'The MariaDB: demo/sample-mariadb is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready


$ kubectl get sts,svc,secret,pvc,pv,pod -n demo
NAME                              READY   AGE
petset.apps/sample-mariadb   3/3     116m

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/sample-mariadb        ClusterIP   10.97.162.171   <none>        3306/TCP   116m
service/sample-mariadb-pods   ClusterIP   None            <none>        3306/TCP   116m

NAME                                TYPE                                  DATA   AGE
secret/default-token-696cj          kubernetes.io/service-account-token   3      121m
secret/sample-mariadb-auth          kubernetes.io/basic-auth              2      116m
secret/sample-mariadb-token-dk4dx   kubernetes.io/service-account-token   3      116m

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-sample-mariadb-0   Bound    pvc-1e259abc-5937-421a-990c-b903a83d2d8a   1Gi        RWO            standard       116m
persistentvolumeclaim/data-sample-mariadb-1   Bound    pvc-1d0b5bcd-2699-4b87-b57b-3072ddc1027f   1Gi        RWO            standard       116m
persistentvolumeclaim/data-sample-mariadb-2   Bound    pvc-5b85a06e-17f5-487a-9150-e928f5cf4590   1Gi        RWO            standard       116m

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS   REASON   AGE
persistentvolume/pvc-1d0b5bcd-2699-4b87-b57b-3072ddc1027f   1Gi        RWO            Delete           Bound    demo/data-sample-mariadb-1   standard                116m
persistentvolume/pvc-1e259abc-5937-421a-990c-b903a83d2d8a   1Gi        RWO            Delete           Bound    demo/data-sample-mariadb-0   standard                116m
persistentvolume/pvc-5b85a06e-17f5-487a-9150-e928f5cf4590   1Gi        RWO            Delete           Bound    demo/data-sample-mariadb-2   standard                116m

NAME                   READY   STATUS    RESTARTS   AGE
pod/sample-mariadb-0   1/1     Running   0          116m
pod/sample-mariadb-1   1/1     Running   0          116m
pod/sample-mariadb-2   1/1     Running   0          116m
```

## Connect with MariaDB database

Once the database is in running state we can conncet to each of three nodes. We will use login credentials `MYSQL_ROOT_USERNAME` and `MYSQL_ROOT_PASSWORD` saved as container's environment variable.

```bash
# First Node
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 26
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SELECT 1; 
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

MariaDB [(none)]> quit;
Bye


# Second Node
$ kubectl exec -it -n demo sample-mariadb-1 -- bash
root@sample-mariadb-1:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 94
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SELECT 1;
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

MariaDB [(none)]> quit;
Bye


# Third Node
$ kubectl exec -it -n demo sample-mariadb-2 -- bash
root@sample-mariadb-2:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 78
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SELECT 1;
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

MariaDB [(none)]> quit;
Bye
```

## Check the Cluster Status

Now, we are ready to check newly created cluster status. Connect and run the following commands from any of the hosts and you will get the same result, that is the cluster size is three.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 137
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> show status like 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.001 sec)
```

## Data Availability

In a MariaDB Galera Cluster, Each member can read and write. In this section, we will insert data from any nodes, and we will see whether we can get the data from every other members.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 202
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE playground;
Query OK, 1 row affected (0.013 sec)

# Create table in Node 1
MariaDB [(none)]> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.053 sec)

# Insert sample data into Node 1
MariaDB [(none)]> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');
Query OK, 1 row affected (0.003 sec)

# Read data from Node 1
MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.001 sec)

MariaDB [(none)]> quit;
Bye
root@sample-mariadb-0:/ exit
exit
~ $ kubectl exec -it -n demo sample-mariadb-1 -- bash
root@sample-mariadb-1:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 209
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Read data from Node 2
MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.001 sec)

#Insert data into node 2
MariaDB [(none)]> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 4, 'red');
Query OK, 1 row affected (0.032 sec)

# Read data from Node 2 after insertion
MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
+----+-------+-------+-------+
2 rows in set (0.000 sec)

MariaDB [(none)]> quit;
Bye
root@sample-mariadb-1:/ exit
exit
~ $ kubectl exec -it -n demo sample-mariadb-2 -- bash
root@sample-mariadb-2:/  mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 209
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Insert data into Node 3
MariaDB [(none)]> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 4, 'red');
Query OK, 1 row affected (0.005 sec)

# Read data from Node 3
MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
|  6 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.000 sec)

MariaDB [(none)]> quit
Bye
root@sample-mariadb-2:/# exit
exit
```

## Automatic Failover

To test automatic failover, we will force the one of three pods to restart and check if it can rejoin the cluster.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 11
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Check current data
MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
|  6 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.002 sec)

MariaDB [(none)]> quit;
Bye
root@sample-mariadb-0:/ exit
exit

# Forcefully delete Node 1
~ $ kubectl delete pod -n demo sample-mariadb-0
pod "sample-mariadb-0" deleted

# Wait for sample-mariadb-0 to restart
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 10
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Check data after rejoining
MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
|  6 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.002 sec)

# Check cluster size
MariaDB [(none)]> show status like 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.002 sec)

MariaDB [(none)]> quit
Bye

```

## Cleaning up

Clean what we created in this tutorial.

```bash
$ kubectl delete mariadb -n demo sample-mariadb
mariadb.kubedb.com "sample-mariadb" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
