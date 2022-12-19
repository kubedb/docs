---
title: PerconaXtraDB Galera Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-clustering-galeracluster
    name: PerconaXtraDB Galera Cluster Guide
    parent: guides-perconaxtradb-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - PerconaXtraDB Cluster

This tutorial will show you how to use KubeDB to provision a PerconaXtraDB replication group in single-primary mode.

## Before You Begin

Before proceeding:

- Read [perconaxtradb galera cluster concept](/docs/guides/perconaxtradb/clustering/overview) to learn about PerconaXtraDB Group Replication.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy PerconaXtraDB Cluster

The following is an example `PerconaXtraDB` object which creates a multi-master PerconaXtraDB group with three members.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/clustering/galera-cluster/examples/demo-1.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Here,

- `spec.replicas` is the number of nodes in the cluster.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `PerconaXtraDB` objects using Kubernetes API. When a `PerconaXtraDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching PerconaXtraDB object name. KubeDB operator will also create a governing service for the StatefulSet with the name `<perconaxtradb-object-name>-pods`.

```bash
$ kubectl get perconaxtradb -n demo sample-pxc -o yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"PerconaXtraDB","metadata":{"annotations":{},"name":"sample-pxc","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"WipeOut","version":"8.0.26"}}
  creationTimestamp: "2021-03-16T09:39:01Z"
  finalizers:
  - kubedb.com
  generation: 2
  managedFields:
    ...
  name: sample-pxc
  namespace: demo
spec:
  authSecret:
    name: sample-pxc-auth
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
  terminationPolicy: WipeOut
  version: 8.0.26
status:
  conditions:
  - lastTransitionTime: "2021-03-16T09:39:01Z"
    message: 'The KubeDB operator has started the provisioning of PerconaXtraDB: demo/sample-pxc'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2021-03-16T09:40:00Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2021-03-16T09:39:09Z"
    message: 'The PerconaXtraDB: demo/sample-pxc is accepting client requests.'
    observedGeneration: 2
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2021-03-16T09:39:50Z"
    message: 'The MySQL: demo/sample-pxc is ready.'
    observedGeneration: 2
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2021-03-16T09:40:00Z"
    message: 'The PerconaXtraDB: demo/sample-pxc is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready


$ kubectl get sts,svc,secret,pvc,pv,pod -n demo
NAME                              READY   AGE
statefulset.apps/sample-pxc   3/3     116m

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/sample-pxc        ClusterIP   10.97.162.171   <none>        3306/TCP   116m
service/sample-pxc-pods   ClusterIP   None            <none>        3306/TCP   116m

NAME                                TYPE                                  DATA   AGE
secret/default-token-696cj          kubernetes.io/service-account-token   3      121m
secret/sample-pxc-auth          kubernetes.io/basic-auth              2      116m
secret/sample-pxc-token-dk4dx   kubernetes.io/service-account-token   3      116m

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-sample-pxc-0   Bound    pvc-1e259abc-5937-421a-990c-b903a83d2d8a   1Gi        RWO            standard       116m
persistentvolumeclaim/data-sample-pxc-1   Bound    pvc-1d0b5bcd-2699-4b87-b57b-3072ddc1027f   1Gi        RWO            standard       116m
persistentvolumeclaim/data-sample-pxc-2   Bound    pvc-5b85a06e-17f5-487a-9150-e928f5cf4590   1Gi        RWO            standard       116m

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS   REASON   AGE
persistentvolume/pvc-1d0b5bcd-2699-4b87-b57b-3072ddc1027f   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-1   standard                116m
persistentvolume/pvc-1e259abc-5937-421a-990c-b903a83d2d8a   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-0   standard                116m
persistentvolume/pvc-5b85a06e-17f5-487a-9150-e928f5cf4590   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-2   standard                116m

NAME                   READY   STATUS    RESTARTS   AGE
pod/sample-pxc-0   1/1     Running   0          116m
pod/sample-pxc-1   1/1     Running   0          116m
pod/sample-pxc-2   1/1     Running   0          116m
```

## Connect with PerconaXtraDB database

Once the database is in running state we can conncet to each of three nodes. We will use login credentials `MYSQL_ROOT_USERNAME` and `MYSQL_ROOT_PASSWORD` saved as container's environment variable.

```bash
# First Node
$ kubectl exec -it -n demo sample-pxc-0 -- bash
root@sample-pxc-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 26
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]> SELECT 1; 
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

PerconaXtraDB [(none)]> quit;
Bye


# Second Node
$ kubectl exec -it -n demo sample-pxc-1 -- bash
root@sample-pxc-1:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 94
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]> SELECT 1;
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

PerconaXtraDB [(none)]> quit;
Bye


# Third Node
$ kubectl exec -it -n demo sample-pxc-2 -- bash
root@sample-pxc-2:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 78
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]> SELECT 1;
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

PerconaXtraDB [(none)]> quit;
Bye
```

## Check the Cluster Status

Now, we are ready to check newly created cluster status. Connect and run the following commands from any of the hosts and you will get the same result, that is the cluster size is three.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
root@sample-pxc-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 137
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]> show status like 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.001 sec)
```

## Data Availability

In a PerconaXtraDB Galera Cluster, Each member can read and write. In this section, we will insert data from any nodes, and we will see whether we can get the data from every other members.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
root@sample-pxc-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 202
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]> CREATE DATABASE playground;
Query OK, 1 row affected (0.013 sec)

# Create table in Node 1
PerconaXtraDB [(none)]> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.053 sec)

# Insert sample data into Node 1
PerconaXtraDB [(none)]> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');
Query OK, 1 row affected (0.003 sec)

# Read data from Node 1
PerconaXtraDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.001 sec)

PerconaXtraDB [(none)]> quit;
Bye
root@sample-pxc-0:/ exit
exit
~ $ kubectl exec -it -n demo sample-pxc-1 -- bash
root@sample-pxc-1:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 209
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Read data from Node 2
PerconaXtraDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.001 sec)

#Insert data into node 2
PerconaXtraDB [(none)]> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 4, 'red');
Query OK, 1 row affected (0.032 sec)

# Read data from Node 2 after insertion
PerconaXtraDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
+----+-------+-------+-------+
2 rows in set (0.000 sec)

PerconaXtraDB [(none)]> quit;
Bye
root@sample-pxc-1:/ exit
exit
~ $ kubectl exec -it -n demo sample-pxc-2 -- bash
root@sample-pxc-2:/  mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 209
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Insert data into Node 3
PerconaXtraDB [(none)]> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 4, 'red');
Query OK, 1 row affected (0.005 sec)

# Read data from Node 3
PerconaXtraDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
|  6 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.000 sec)

PerconaXtraDB [(none)]> quit
Bye
root@sample-pxc-2:/# exit
exit
```

## Automatic Failover

To test automatic failover, we will force the one of three pods to restart and check if it can rejoin the cluster.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
kubectl exec -it -n demo sample-pxc-0 -- bash
root@sample-pxc-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 11
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Check current data
PerconaXtraDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
|  6 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.002 sec)

PerconaXtraDB [(none)]> quit;
Bye
root@sample-pxc-0:/ exit
exit

# Forcefully delete Node 1
~ $ kubectl delete pod -n demo sample-pxc-0
pod "sample-pxc-0" deleted

# Wait for sample-pxc-0 to restart
$ kubectl exec -it -n demo sample-pxc-0 -- bash
root@sample-pxc-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 10
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Check data after rejoining
PerconaXtraDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  5 | slide |     4 | red   |
|  6 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.002 sec)

# Check cluster size
PerconaXtraDB [(none)]> show status like 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.002 sec)

PerconaXtraDB [(none)]> quit
Bye

```

## Cleaning up

Clean what we created in this tutorial.

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
perconaxtradb.kubedb.com "sample-pxc" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
