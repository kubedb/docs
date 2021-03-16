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

# KubeDB - MariaDB Group Replication

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
  version: "10.5.8"
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
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"sample-mariadb","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"WipeOut","version":"10.5.8"}}
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
  terminationPolicy: WipeOut
  version: 10.5.8
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
statefulset.apps/sample-mariadb   3/3     116m

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

Once the database pod is in `Running` state, verify that all 3 nodes joined the cluster. We will conncet into the first database using the creadentials `MYSQL_ROOT_USERNAME` and `MYSQL_ROOT_PASSWORD` saved as container's enoviroment variables.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 26
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> show status like 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.001 sec)

MariaDB [(none)]> quit;
Bye
```

From the above log, we can see that 3 nodes cluster is created and ready to accept connections.

## Connect with MariaDB database

KubeDB operator has created a new Secret called `my-group-auth` **(format: {mysql-object-name}-auth)** for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **username** for MariaDB superuser and a `password` key which contains the **password** for MariaDB superuser.

If you want to use an existing secret please specify that when creating the MariaDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/mysql.md#specdatabasesecret).

Now, you can connect to this database from your terminal using the `mysql` user and password.

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
dlNiQpjULZvEqo3B
```

The operator creates a group according to the newly created `MariaDB` object. This group has 3 members (one primary and two secondary).

You can connect to any of these group members. In that case you just need to specify the host name of that member Pod (either PodIP or the fully-qualified-domain-name for that Pod using the governing service named `<mysql-object-name>-gvr`) by `--host` flag.

```bash
# first list the mysql pods list
$ kubectl get pods -n demo -l app.kubernetes.io/instance=my-group
NAME         READY   STATUS    RESTARTS   AGE
my-group-0   2/2     Running   1          19m
my-group-1   2/2     Running   0          15m
my-group-2   2/2     Running   1          12m


# get the governing service
$ kubectl get service my-group-gvr -n demo
NAME           TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
my-group-gvr   ClusterIP   None         <none>        3306/TCP   137m

# list the pods with PodIP
$ kubectl get pods -n demo -l app.kubernetes.io/instance=my-group -o jsonpath='{range.items[*]}{.metadata.name} ........... {.status.podIP} ............ {.metadata.name}.my-group-gvr.{.metadata.namespace}{"\\n"}{end}'
my-group-0 ........... 172.17.0.5 ............ my-group-0.my-group-gvr.demo
my-group-1 ........... 172.17.0.6 ............ my-group-1.my-group-gvr.demo
my-group-2 ........... 172.17.0.7 ............ my-group-2.my-group-gvr.demo
```

Now you can connect to these database using the above info. Ignore the warning message. It is happening for using password in the command.

```bash
# connect to the 1st server
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 2nd server
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-1.my-group-gvr.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 3rd server
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-2.my-group-gvr.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+
```

## Check the Group Status

Now, you are ready to check newly created group status. Connect and run the following commands from any of the hosts and you will get the same results.

```bash
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "show status like '%primary%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+--------------------------------------+
| Variable_name                    | Value                                |
+----------------------------------+--------------------------------------+
| group_replication_primary_member | fc4a4935-e6bf-11ea-bf42-9a7560d22b8f |
+----------------------------------+--------------------------------------+
```

The value **37ed2c72-680a-11e9-8ac3-0242ac110005** in the above table means the ID of the primary member of the group.

```bash
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=MpPhZ9xbVlxvoC4d --host=my-group-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
| group_replication_applier | 768c83e2-e6c0-11ea-bfa5-5a3f7e8f0c23 | my-group-1.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.21         |
| group_replication_applier | 9ea36443-e6c0-11ea-b4d9-beec0709d261 | my-group-2.my-group-gvr.demo |        3306 | ONLINE       | SECONDARY   | 8.0.21         |
| group_replication_applier | fc4a4935-e6bf-11ea-bf42-9a7560d22b8f | my-group-0.my-group-gvr.demo |        3306 | ONLINE       | PRIMARY     | 8.0.21         |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+-------------+----------------+
```

## Data Availability

In a MariaDB group, only the primary member can write not the secondary. But you can read data from any member. In this tutorial, we will insert data from primary, and we will see whether we can get the data from any other members.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# create a database on primary
$ kubectl exec -it -n demo my-group-0 -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo my-group-0 -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$  kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from primary
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from secondary-1
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-1.my-group-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from secondary-2
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-2.my-group-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Write on Secondary Should Fail

Only, primary member preserves the write permission. No secondary can write data.

```bash
# try to write on secondary-1
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-1.my-group-gvr.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MariaDB server is running with the --super-read-only option so it cannot execute this statement
command terminated with exit code 1

# try to write on secondary-2
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-2.my-group-gvr.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MariaDB server is running with the --super-read-only option so it cannot execute this statement
command terminated with exit code 1
```

## Automatic Failover

To test automatic failover, we will force the primary Pod to restart. Since the primary member (`Pod`) becomes unavailable, the rest of the members will elect a new primary for these group. When the old primary comes back, it will join the group as a secondary member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# delete the primary Pod my-group-0
$ kubectl delete pod my-group-0 -n demo
pod "my-group-0" deleted

# check the new primary ID
kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "show status like '%primary%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+--------------------------------------+
| Variable_name                    | Value                                |
+----------------------------------+--------------------------------------+
| group_replication_primary_member | 4c97e5a4-680a-11e9-9f6b-0242ac110006 |
+----------------------------------+--------------------------------------+

# now check the gruop status
kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                  | MEMBER_PORT | MEMBER_STATE |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+
| group_replication_applier | 37ed2c72-680a-11e9-8ac3-0242ac110005 | my-group-0.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | 4c97e5a4-680a-11e9-9f6b-0242ac110006 | my-group-1.my-group-gvr.demo |        3306 | ONLINE       |
| group_replication_applier | 625714bc-680a-11e9-9e94-0242ac110007 | my-group-2.my-group-gvr.demo |        3306 | ONLINE       |
+---------------------------+--------------------------------------+------------------------------+-------------+--------------+

# read data from new primary my-group-1.my-group-gvr.demo
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-1.my-group-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+

# read data from secondary-1 my-group-0.my-group-gvr.demo
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-0.my-group-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+

# read data from secondary-2 my-group-2.my-group-gvr.demo
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password=dlNiQpjULZvEqo3B --host=my-group-2.my-group-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Cleaning up

Clean what you created in this tutorial.

```bash
kubectl delete -n demo my/my-group
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MariaDB object](/docs/guides/mysql/concepts/mysql.md).
- Detail concepts of [MariaDBDBVersion object](/docs/guides/mysql/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
