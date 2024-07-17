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

This tutorial will show you how to create a 3 node PerconaXtraDB Cluster using KubeDB.

## Before You Begin

Before proceeding:

- Read [perconaxtradb galera cluster concept](/docs/guides/percona-xtradb/clustering/overview) to learn about PerconaXtraDB Group Replication.

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
apiVersion: kubedb.com/v1
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
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/clustering/galera-cluster/examples/demo-1.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Here,

- `spec.replicas` is the number of nodes in the cluster.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `PerconaXtraDB` objects using Kubernetes API. When a `PerconaXtraDB` object is created, KubeDB operator will create a new PetSet and a Service with the matching PerconaXtraDB object name. KubeDB operator will also create a governing service for the PetSet with the name `<perconaxtradb-object-name>-pods`.

```bash
$ kubectl get perconaxtradb -n demo sample-pxc -o yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"PerconaXtraDB","metadata":{"annotations":{},"name":"sample-pxc","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","version":"8.0.26"}}
  creationTimestamp: "2022-12-20T05:15:56Z"
  finalizers:
  - kubedb.com
  generation: 4
  name: sample-pxc
  namespace: demo
  resourceVersion: "8919"
  uid: 5202f646-1f14-4008-9034-cddd481a0ea3
spec:
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: sample-pxc-auth
  autoOps: {}
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: sample-pxc
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: perconaxtradbs.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: sample-pxc
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: perconaxtradbs.kubedb.com
              namespaces:
              - demo
              topologyKey: failure-domain.beta.kubernetes.io/zone
            weight: 50
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      securityContext:
        fsGroup: 1001
        runAsGroup: 1001
        runAsUser: 1001
      serviceAccountName: sample-pxc
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  systemUserSecrets:
    monitorUserSecret:
      name: sample-pxc-monitor
    replicationUserSecret:
      name: sample-pxc-replication
  deletionPolicy: WipeOut
  version: 8.0.26
status:
  conditions:
  - lastTransitionTime: "2022-12-20T05:15:56Z"
    message: 'The KubeDB operator has started the provisioning of PerconaXtraDB: demo/sample-pxc'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2022-12-20T05:17:30Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2022-12-20T05:19:02Z"
    message: database sample-pxc/demo is ready
    observedGeneration: 4
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2022-12-20T05:18:12Z"
    message: database sample-pxc/demo is accepting connection
    observedGeneration: 4
    reason: AcceptingConnection
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2022-12-20T05:19:07Z"
    message: 'The PerconaXtraDB: demo/sample-pxc is successfully provisioned.'
    observedGeneration: 4
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 4
  phase: Ready

$ kubectl get sts,svc,secret,pvc,pv,pod -n demo
NAME                          READY   AGE
petset.apps/sample-pxc   3/3     7m5s

NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/sample-pxc        ClusterIP   10.96.207.41   <none>        3306/TCP   7m11s
service/sample-pxc-pods   ClusterIP   None           <none>        3306/TCP   7m11s

NAME                            TYPE                                  DATA   AGE
secret/default-token-bbgjp      kubernetes.io/service-account-token   3      7m19s
secret/sample-pxc-auth          kubernetes.io/basic-auth              2      7m11s
secret/sample-pxc-monitor       kubernetes.io/basic-auth              2      7m11s
secret/sample-pxc-replication   kubernetes.io/basic-auth              2      7m11s
secret/sample-pxc-token-gbzg6   kubernetes.io/service-account-token   3      7m11s

NAME                                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-sample-pxc-0   Bound    pvc-cb4f41de-1ead-4124-98a7-e3e950c8d10f   1Gi        RWO            standard       7m5s
persistentvolumeclaim/data-sample-pxc-1   Bound    pvc-3c2887f5-5a7c-4df3-b7ca-34a6bcf91904   1Gi        RWO            standard       7m5s
persistentvolumeclaim/data-sample-pxc-2   Bound    pvc-521f81f1-6261-4252-a2a8-32bfe472000e   1Gi        RWO            standard       7m5s

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
persistentvolume/pvc-3c2887f5-5a7c-4df3-b7ca-34a6bcf91904   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-1   standard                7m3s
persistentvolume/pvc-521f81f1-6261-4252-a2a8-32bfe472000e   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-2   standard                7m1s
persistentvolume/pvc-cb4f41de-1ead-4124-98a7-e3e950c8d10f   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-0   standard                7m2s

NAME               READY   STATUS    RESTARTS   AGE
pod/sample-pxc-0   2/2     Running   0          7m5s
pod/sample-pxc-1   2/2     Running   0          7m5s
pod/sample-pxc-2   2/2     Running   0          7m5s

```

## Connect with PerconaXtraDB database

Once the database is in running state we can connect to each of three nodes. We will use login credentials `MYSQL_ROOT_USERNAME` and `MYSQL_ROOT_PASSWORD` saved as container's environment variable.

```bash
# First Node
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 133
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT 1; 
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.00 sec)

mysql> quit;
Bye


# Second Node
$ kubectl exec -it -n demo sample-pxc-1 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 123
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT 1; 
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.00 sec)

mysql> quit;
Bye


# Third Node
$ kubectl exec -it -n demo sample-pxc-2 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 139
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT 1;
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.00 sec)

mysql> quit;
Bye

```

## Check the Cluster Status

Now, we are ready to check newly created cluster status. Connect and run the following commands from any of the hosts and you will get the same result, that is the cluster size is three.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 231
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show status like 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.00 sec)

mysql> quit;
Bye

```

## Data Availability

In a PerconaXtraDB Galera Cluster, Each member can read and write. In this section, we will insert data from any nodes, and we will see whether we can get the data from every other members.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 260
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE playground;
Query OK, 1 row affected (0.01 sec)

mysql> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.02 sec)

mysql> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');
Query OK, 1 row affected (0.00 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> quit;
Bye
bash-4.4$ exit
exit

$ kubectl exec -it -n demo sample-pxc-2 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 253
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 4, 'red');
Query OK, 1 row affected (0.00 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  6 | slide |     4 | red   |
+----+-------+-------+-------+
2 rows in set (0.00 sec)

mysql> quit;
Bye
bash-4.4$ exit
exit

$ kubectl exec -it -n demo sample-pxc-2 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 283
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 4, 'red');
Query OK, 1 row affected (0.00 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  6 | slide |     4 | red   |
|  9 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.00 sec)

mysql> quit;
Bye
bash-4.4$ exit
exit
```

## Automatic Failover

To test automatic failover, we will force the one of three pods to restart and check if it can rejoin the cluster.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 332
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  6 | slide |     4 | red   |
|  9 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.00 sec)

mysql> quit;
Bye
bash-4.4$ exit
exit


# Forcefully delete Node 1
~ $ kubectl delete pod -n demo sample-pxc-0
pod "sample-pxc-0" deleted

# Wait for sample-pxc-0 to restart
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 49
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
|  6 | slide |     4 | red   |
|  9 | slide |     4 | red   |
+----+-------+-------+-------+
3 rows in set (0.00 sec)

# Check cluster size
mysql [(none)]> show status like 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.002 sec)

mysql> quit
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
