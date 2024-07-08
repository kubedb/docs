---
title: MySQL Group Replcation Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-clustering-group-replication
    name: MySQL Group Replication Guide
    parent: guides-mysql-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MySQL Group Replication

This tutorial will show you how to use KubeDB to provision a MySQL replication group in single-primary mode.

## Before You Begin

Before proceeding:

- Read [mysql group replication concept](/docs/guides/mysql/clustering/overview/index.md) to learn about MySQL Group Replication.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/clustering/group-replication/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/clustering/group-replication/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MySQL Cluster

To deploy a single primary MySQL replication group , specify `spec.topology` field in `MySQL` CRD.

The following is an example `MySQL` object which creates a MySQL group with three members (one is primary member and the two others are secondary members).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "8.0.35"
  replicas: 3
  topology:
    mode: GroupReplication
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/group-replication/yamls/group-replication.yaml
mysql.kubedb.com/my-group created
```

Here,

- `spec.topology` tells about the clustering configuration for MySQL.
- `spec.topology.mode` specifies the mode for MySQL cluster. Here we have used `GroupReplication` to tell the operator that we want to deploy a MySQL replication group.
- `spec.topology.group` contains group replication info.
- `spec.topology.group.name` the name for the group. It is a valid version 4 UUID.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, KubeDB operator will create a new PetSet and a Service with the matching MySQL object name. KubeDB operator will also create a governing service for the PetSet with the name `<mysql-object-name>-pods`.

```bash
$ kubectl dba describe my -n demo my-group
Name:               my-group
Namespace:          demo
CreationTimestamp:  Tue, 28 Jun 2022 17:54:10 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"my-group","namespace":"demo"},"spec":{"replicas":3,"storage":{"...
Replicas:           3  total
Status:             Provisioning
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  WipeOut

PetSet:          
  Name:               my-group
  CreationTimestamp:  Tue, 28 Jun 2022 17:54:10 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=my-group
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:        <none>
  Replicas:           824640792392 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         my-group
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.223.45
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.44:3306

Service:        
  Name:         my-group-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.44:3306,10.244.0.46:3306,10.244.0.48:3306

Service:        
  Name:         my-group-standby
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.70.224
  Port:         standby  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    <none>

Auth Secret:
  Name:         my-group-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"my-group","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","topology":{"group":{"name":"dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"},"mode":"GroupReplication"},"version":"8.0.35"}}

    Creation Timestamp:  2022-06-28T11:54:10Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    my-group
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysqls.kubedb.com
    Name:                            my-group
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    my-group
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(my-group.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  mysql-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  mysql-restore-8.0.21
    Secret:
      Name:   my-group-auth
    Type:     kubedb.com/mysql
    Version:  8.0.35

Events:
  Type     Reason      Age   From               Message
  ----     ------      ----  ----               -------
  Normal   Successful  1m    Kubedb operator  Successfully created governing service
  Normal   Successful  1m    Kubedb operator  Successfully created service for primary/standalone
  Normal   Successful  1m    Kubedb operator  Successfully created service for secondary replicas
  Normal   Successful  1m    Kubedb operator  Successfully created database auth secret
  Normal   Successful  1m    Kubedb operator  Successfully created PetSet
  Normal   Successful  1m    Kubedb operator  Successfully created MySQL
  Normal   Successful  1m    Kubedb operator  Successfully created appbinding


$ kubectl get statefulset -n demo
NAME       READY   AGE
my-group   3/3     3m47s

$ kubectl get pvc -n demo
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-my-group-0   Bound    pvc-4f8538f6-a6ce-4233-b533-8566852f5b98   1Gi        RWO            standard       4m16s
data-my-group-1   Bound    pvc-8823d3ad-d614-4172-89ac-c2284a17f502   1Gi        RWO            standard       4m11s
data-my-group-2   Bound    pvc-94f1c312-50e3-41e1-94a8-a820be0abc08   1Gi        RWO            standard       4m7s
s

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
pvc-4f8538f6-a6ce-4233-b533-8566852f5b98   1Gi        RWO            Delete           Bound    demo/data-my-group-0   standard                4m39s
pvc-8823d3ad-d614-4172-89ac-c2284a17f502   1Gi        RWO            Delete           Bound    demo/data-my-group-1   standard                4m35s
pvc-94f1c312-50e3-41e1-94a8-a820be0abc08   1Gi        RWO            Delete           Bound    demo/data-my-group-2   standard                4m31s

$ kubectl get service -n demo
NAME               TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
my-group           ClusterIP      10.96.223.45    <none>        3306/TCP       5m13s
my-group-pods      ClusterIP      None            <none>        3306/TCP       5m13s
my-group-standby   ClusterIP      10.96.70.224    <none>        3306/TCP       5m13s

```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified `MySQL` object:

```yaml
$ kubectl get  my -n demo my-group -o yaml | kubectl neat
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  authSecret:
    name: my-group-auth
  podTemplate:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: my-group
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: mysqls.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: my-group
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: mysqls.kubedb.com
              namespaces:
              - demo
              topologyKey: failure-domain.beta.kubernetes.io/zone
            weight: 50
      resources:
        limits:
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: my-group
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
  topology:
    group:
      name: dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b
    mode: GroupReplication
  version: 8.0.35
status:
  observedGeneration: 2$4213139756412538772
  phase: Running
```

## Connect with MySQL database

KubeDB operator has created a new Secret called `my-group-auth` **(format: {mysql-object-name}-auth)** for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **username** for MySQL superuser and a `password` key which contains the **password** for MySQL superuser.

If you want to use an existing secret please specify that when creating the MySQL object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, you can connect to this database from your terminal using the `mysql` user and password.

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
d)q2MVmJK$Oex=mW
```

The operator creates a group according to the newly created `MySQL` object. This group has 3 members (one primary and two secondary).

You can connect to any of these group members. In that case you just need to specify the host name of that member Pod (either PodIP or the fully-qualified-domain-name for that Pod using the governing service named `<mysql-object-name>-pods`) by `--host` flag.

```bash
# first list the mysql pods list
$ kubectl get pods -n demo -l app.kubernetes.io/instance=my-group
NAME         READY   STATUS    RESTARTS   AGE
my-group-0   2/2     Running   0          8m23s
my-group-1   2/2     Running   0          8m18s
my-group-2   2/2     Running   0          8m14s


# get the governing service
$ kubectl get service my-group-pods -n demo
NAME            TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
my-group-pods   ClusterIP   None         <none>        3306/TCP   8m49s

# list the pods with PodIP
$ kubectl get pods -n demo -l app.kubernetes.io/instance=my-group -o jsonpath='{range.items[*]}{.metadata.name} ........... {.status.podIP} ............ {.metadata.name}.my-group-pods.{.metadata.namespace}{"\\n"}{end}'
my-group-0 ........... 10.244.0.44 ............ my-group-0.my-group-pods.demo
my-group-1 ........... 10.244.0.46 ............ my-group-1.my-group-pods.demo
my-group-2 ........... 10.244.0.48 ............ my-group-2.my-group-pods.demo

```

Now you can connect to these database using the above info. Ignore the warning message. It is happening for using password in the command.

```bash
# connect to the 1st server
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 2nd server
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-1.my-group-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 3rd server
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-2.my-group-pods.demo -e "select 1;"
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
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-0.my-group-pods.demo -e "show status like '%primary%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+--------------------------------------+
| Variable_name                    | Value                                |
+----------------------------------+--------------------------------------+
| group_replication_primary_member | 1ace16b5-f6d9-11ec-9a26-9ae7d6def698 |
+----------------------------------+--------------------------------------+

```

The value **1ace16b5-f6d9-11ec-9a26-9ae7d6def698** in the above table is the ID of the primary member of the group.

```bash
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-0.my-group-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                       | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION | MEMBER_COMMUNICATION_STACK |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| group_replication_applier | 13aad5a5-f6d9-11ec-87bb-96e838330519 | my-group-2.my-group-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.35         | XCom                       |
| group_replication_applier | 1739589f-f6d9-11ec-956c-c2c213efafa8 | my-group-1.my-group-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.35         | XCom                       |
| group_replication_applier | 1ace16b5-f6d9-11ec-9a26-9ae7d6def698 | my-group-0.my-group-pods.demo.svc |        3306 | ONLINE       | PRIMARY     | 8.0.35         | XCom                       |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+

```

## Data Availability

In a MySQL group, only the primary member can write not the secondary. But you can read data from any member. In this tutorial, we will insert data from primary, and we will see whether we can get the data from any other member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# create a database on primary
$ kubectl exec -it -n demo my-group-0 -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo my-group-0 -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$  kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from primary
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```
In the previous step we have inserted into the primary pod. In the next step we will read from secondary pods to determine whether the data has been successfully copied to the secondary pods.
```bash
# read from secondary-1
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-1.my-group-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from secondary-2
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-2.my-group-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Write on Secondary Should Fail

Only, primary member preserves the write permission. No secondary can write data.

```bash
# try to write on secondary-1
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-1.my-group-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MySQL server is running with the --super-read-only option so it cannot execute this statement
command terminated with exit code 1

# try to write on secondary-2
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW'  --host=my-group-2.my-group-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MySQL server is running with the --super-read-only option so it cannot execute this statement
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
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "show status like '%primary%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+--------------------------------------+
| Variable_name                    | Value                                |
+----------------------------------+--------------------------------------+
| group_replication_primary_member |  1739589f-f6d9-11ec-956c-c2c213efafa8|
+----------------------------------+--------------------------------------+


# now check the group status
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                       | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION | MEMBER_COMMUNICATION_STACK |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| group_replication_applier | 13aad5a5-f6d9-11ec-87bb-96e838330519 | my-group-2.my-group-pods.demo.svc |        3306 | ONLINE       | SECONDARY     | 8.0.35         | XCom                       |
| group_replication_applier | 1739589f-f6d9-11ec-956c-c2c213efafa8 | my-group-1.my-group-pods.demo.svc |        3306 | ONLINE       | PRIMARY   | 8.0.35         | XCom                       |
| group_replication_applier | 1ace16b5-f6d9-11ec-9a26-9ae7d6def698 | my-group-0.my-group-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.35         | XCom                       |
+---------------------------+--------------------------------------+-----------------------------------+-------------+--------------+-------------+----------------+----------------------------+


# read data from new primary my-group-1.my-group-pods.demo
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-1.my-group-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```
Now Let's read the data from secondary pods to see if the data is consistant.
```bash
# read data from secondary-1 my-group-0.my-group-pods.demo
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-0.my-group-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

# read data from secondary-2 my-group-2.my-group-pods.demo
$ kubectl exec -it -n demo my-group-0 -c mysql -- mysql -u root --password='d)q2MVmJK$Oex=mW' --host=my-group-2.my-group-pods.demo -e "SELECT * FROM playground.equipment;"
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

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLDBVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
