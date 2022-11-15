---
title: MySQL Innodb Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-clustering-innodb-cluster
    name: MySQL Innodb Cluster Guide
    parent: guides-mysql-clustering
    weight: 22
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MySQL Innodb Cluster

This tutorial will show you how to use KubeDB to provision a MySQL Innodb cluster single-primary mode.

## Before You Begin

Before proceeding:
- Innodb cluster itself use mysql group replication under the hood 
- Read [mysql group replication concept](/docs/guides/mysql/clustering/overview/index.md) to learn about MySQL Group Replication.
- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```



## Deploy MySQL Innodb Cluster

To deploy a MySQL Innodb cluster, specify `spec.topology` field in `MySQL` CRD.

The following is an example `MySQL` object which creates a MySQL Innodb cluster with three members (one is primary member and the two others are secondary members).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: innodb
  namespace: demo
spec:
  version: "8.0.31-innodb"
  replicas: 3
  topology:
    mode: InnoDBCluster
    innoDBCluster:
      router:
        replicas: 1
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/innodb-cluster/yamls/innodb.yaml
mysql.kubedb.com/innodb created
```

Here,

- `spec.topology` tells about the clustering configuration for MySQL.
- `spec.topology.mode` specifies the mode for MySQL cluster. Here we have used `InnoDBCluster` to tell the operator that we want to deploy a MySQL Innodb Cluster.
- `spec.topology.innoDBCluster` contains the InnodbCluster info.Innodb cluster comes with a router as a load balancer
- `spec.topology.Router.replica` is for the number of replica fo innodb cluster router.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MySQL object name. KubeDB operator will also create a governing service for the StatefulSet with the name `<mysql-object-name>-pods`.

```bash
$ kubectl dba describe my -n demo innodb
Name:               innodb
Namespace:          demo
CreationTimestamp:  Tue, 15 Nov 2022 15:14:42 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"innodb","namespace":"demo"},"spec":{"replicas":3,"storage":{"ac...
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

StatefulSet:          
  Name:               innodb
  CreationTimestamp:  Tue, 15 Nov 2022 15:14:42 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=innodb
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysqls.kubedb.com
                        mysql.kubedb.com/component=database
  Annotations:        <none>
  Replicas:           824641134776 desired | 1 total
  Pods Status:        0 Running / 1 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         innodb
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=innodb
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
                  mysql.kubedb.com/component=database
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.244.213
  Port:         primary  3306/TCP
  TargetPort:   rw/TCP
  Endpoints:    

Service:        
  Name:         innodb-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=innodb
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
                  mysql.kubedb.com/component=database
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.26:3306

Service:        
  Name:         innodb-standby
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=innodb
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
                  mysql.kubedb.com/component=database
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.146.147
  Port:         standby  3306/TCP
  TargetPort:   ro/TCP
  Endpoints:    

Auth Secret:
  Name:         innodb-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=innodb
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
                  mysql.kubedb.com/component=database
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"innodb","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"WipeOut","topology":{"innoDBCluster":{"router":{"replicas":1}},"mode":"InnoDBCluster"},"version":"8.0.31-innodb"}}

    Creation Timestamp:  2022-11-15T09:14:42Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    innodb
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysqls.kubedb.com
      mysql.kubedb.com/component:    database
    Name:                            innodb
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    innodb
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(innodb.demo.svc:3306)/
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
      Name:   innodb-auth
    Type:     kubedb.com/mysql
    Version:  8.0.31

Events:
  Type    Reason         Age   From            Message
  ----    ------         ----  ----            -------
  Normal  Phase Changed  27s   MySQL operator  phase changed from  to Provisioning reason:
  Normal  Successful     27s   MySQL operator  Successfully created governing service
  Normal  Successful     27s   MySQL operator  Successfully created service for primary/standalone
  Normal  Successful     27s   MySQL operator  Successfully created service for secondary replicas
  Normal  Successful     27s   MySQL operator  Successfully created database auth secret
  Normal  Successful     27s   MySQL operator  Successfully created StatefulSet
  Normal  Successful     27s   MySQL operator  successfully patched created StatefulSet innodb-router
  Normal  Successful     27s   MySQL operator  Successfully created MySQL
  Normal  Successful     27s   MySQL operator  Successfully created appbinding

$ kubectl get statefulset -n demo
NAME       READY   AGE
NAME            READY   AGE
innodb          3/3     2m17s
innodb-router   1/1     2m17s

$ kubectl get pvc -n demo
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-innodb-0    Bound    pvc-6f7f8ebd-0b56-45fb-b91a-fe133bfae594   1Gi        RWO            standard       2m47s
data-innodb-1    Bound    pvc-16f9d6df-ce46-49da-9720-415d7f7d8b69   1Gi        RWO            standard       113s
data-innodb-2    Bound    pvc-8cfcb761-eb63-4a12-bc7e-5d86f727330e   1Gi        RWO            standard       88s

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
pvc-6f7f8ebd-0b56-45fb-b91a-fe133bfae594   1Gi        RWO            Delete           Bound    demo/data-innodb-0    standard                3m50s
pvc-16f9d6df-ce46-49da-9720-415d7f7d8b69   1Gi        RWO            Delete           Bound    demo/data-innodb-1    standard                2m38s
pvc-8cfcb761-eb63-4a12-bc7e-5d86f727330e   1Gi        RWO            Delete           Bound    demo/data-innodb-2    standard                2m32s


$ kubectl get service -n demo
NAME               TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
innodb           ClusterIP   10.96.244.213   <none>        3306/TCP   5m23s
innodb-pods      ClusterIP   None            <none>        3306/TCP   5m23s
innodb-standby   ClusterIP   10.96.146.147   <none>        3306/TCP   5m23s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified `MySQL` object:

```yaml
$ kubectl get  my -n demo innodb -o yaml | kubectl neat
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: innodb
  namespace: demo
spec:
  authSecret:
    name: innodb-auth
  podTemplate:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: innodb
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: mysqls.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: innodb
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
      serviceAccountName: innodb
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
  topology:
    innoDBCluster:
      mode: Single-Primary
      router:
        replicas: 1
    mode: InnoDBCluster
  version: 8.0.31-innodb
status:
  observedGeneration: 2
  phase: Running
```

## Connect with MySQL database

KubeDB operator has created a new Secret called `innodb-auth` **(format: {mysql-object-name}-auth)** for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **username** for MySQL superuser and a `password` key which contains the **password** for MySQL superuser.

If you want to use an existing secret please specify that when creating the MySQL object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, you can connect to this database from your terminal using the `mysql` user and password.

```bash
$ kubectl get secrets -n demo innodb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo innodb-auth -o jsonpath='{.data.\password}' | base64 -d
ny5jSirIzVtWDcZ7
```

The operator creates a cluster according to the newly created `MySQL` object. This group has 3 members (one primary and two secondary).

You can connect to any of these cluster members. In that case you just need to specify the host name of that member Pod (either PodIP or the fully-qualified-domain-name for that Pod using the governing service named `<mysql-object-name>-pods`) by `--host` flag.

```bash
# first list the mysql pods list
$ kubectl get pods -n demo -l app.kubernetes.io/instance=innodb
NAME         READY   STATUS    RESTARTS   AGE
innodb-0          2/2     Running   0          15m
innodb-1          2/2     Running   0          14m
innodb-2          2/2     Running   0          14m
innodb-router-0   1/1     Running   0          15m



# get the governing service
$ kubectl get service innodb-pods -n demo
NAME          TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
innodb-pods   ClusterIP   None         <none>        3306/TCP   16m

# list the pods with PodIP
$ kubectl get pods -n demo -l app.kubernetes.io/instance=innodb -o jsonpath='{range.items[*]}{.metadata.name} ........... {.status.podIP} ............ {.metadata.name}.innodb-pods.{.metadata.namespace}{"\\n"}{end}'
innodb-0 ........... 10.244.0.26 ............ innodb-0.innodb-pods.demo
innodb-1 ........... 10.244.0.28 ............ innodb-1.innodb-pods.demo
innodb-2 ........... 10.244.0.30 ............ innodb-2.innodb-pods.demo

```

Now you can connect to this database using the above info. Ignore the warning message. It is happening for using password in the command.

```bash
# connect to the 1st server
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 2nd server
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7'  --host=innodb-1.innodb-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 3rd server
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7'  --host=innodb-2.innodb-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+
```

## Check the Innodb Cluster status

The main advantage of innodb cluster is its comes with an admin shell from where you are able to call the mysql admin api and configure cluster and it provide some functionality wokring with the cluster.
Let's exec into one of the pod to see the cluster status.


```bash
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysqlsh -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo

 MySQL  innodb-0.innodb-pods.demo:33060+ ssl  JS > dba.getCluster().status()
{
    "clusterName": "innodb", 
    "defaultReplicaSet": {
        "name": "default", 
        "primary": "innodb-0.innodb-pods.demo.svc:3306", 
        "ssl": "REQUIRED", 
        "status": "OK", 
        "statusText": "Cluster is ONLINE and can tolerate up to ONE failure.", 
        "topology": {
            "innodb-0.innodb-pods.demo.svc:3306": {
                "address": "innodb-0.innodb-pods.demo.svc:3306", 
                "memberRole": "PRIMARY", 
                "mode": "R/W", 
                "readReplicas": {}, 
                "replicationLag": "applier_queue_applied", 
                "role": "HA", 
                "status": "ONLINE", 
                "version": "8.0.31"
            }, 
            "innodb-1.innodb-pods.demo.svc:3306": {
                "address": "innodb-1.innodb-pods.demo.svc:3306", 
                "memberRole": "SECONDARY", 
                "mode": "R/O", 
                "readReplicas": {}, 
                "replicationLag": "applier_queue_applied", 
                "role": "HA", 
                "status": "ONLINE", 
                "version": "8.0.31"
            }, 
            "innodb-2.innodb-pods.demo.svc:3306": {
                "address": "innodb-2.innodb-pods.demo.svc:3306", 
                "memberRole": "SECONDARY", 
                "mode": "R/O", 
                "readReplicas": {}, 
                "replicationLag": "applier_queue_applied", 
                "role": "HA", 
                "status": "ONLINE", 
                "version": "8.0.31"
            }
        }, 
        "topologyMode": "Single-Primary"
    }, 
    "groupInformationSourceMember": "innodb-0.innodb-pods.demo.svc:3306"
}


```

## Data Availability

In a MySQL Cluster, only the primary member can write not the secondary. But you can read data from any member. In this tutorial, we will insert data from primary, and we will see whether we can get the data from any other member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# create a database on primary
$ kubectl exec -it -n demo innodb-0 -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo innodb-0 -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$  kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from primary
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "SELECT * FROM playground.equipment;"
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
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7'  --host=innodb-1.innodb-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from secondary-2
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7'  --host=innodb-2.innodb-pods.demo -e "SELECT * FROM playground.equipment;"
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
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7'  --host=innodb-1.innodb-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MySQL server is running with the --super-read-only option so it cannot execute this statement
command terminated with exit code 1

# try to write on secondary-2
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7'  --host=innodb-2.innodb-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MySQL server is running with the --super-read-only option so it cannot execute this statement
command terminated with exit code 1
```

## Automatic Failover

To test automatic failover, we will force the primary Pod to restart. Since the primary member (`Pod`) becomes unavailable, the rest of the members will elect a new primary for the cluster. When the old primary comes back, it will join the cluster as a secondary member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# delete the primary Pod innodb-0
$ kubectl delete pod innodb-0  -n demo
pod "innodb-0" deleted

# check the new primary ID
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "show status like '%primary%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+--------------------------------------+
| Variable_name                    | Value                                |
+----------------------------------+--------------------------------------+
| group_replication_primary_member |  2b77185f-64c6-11ed-9621-e21f33a1cdb1|
+----------------------------------+--------------------------------------+


# now check the cluster  status  for underlying group replication
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "select * from performance_schema.replication_group_members"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+--------------------------------------+-------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                   | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION | MEMBER_COMMUNICATION_STACK |
+---------------------------+--------------------------------------+-------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| group_replication_applier | 294f333c-64c6-11ed-9893-468480005d43 | innodb-0.innodb-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.31         | MySQL                      |
| group_replication_applier | 2b77185f-64c6-11ed-9621-e21f33a1cdb1 | innodb-1.innodb-pods.demo.svc |        3306 | ONLINE       | PRIMARY     | 8.0.31         | MySQL                      |
| group_replication_applier | 2f0da15c-64c6-11ed-951a-fa8d12ce91a2 | innodb-2.innodb-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.31         | MySQL                      |
+---------------------------+--------------------------------------+-------------------------------+-------------+--------------+-------------+----------------+----------------------------+


# read data from new primary innodb-1.innodb-pods.demo
$ kubectl exec -it -n demo innodb-1 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-1.innodb-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```
Now Let's read the data from secondary pods to see if the data is consistent.
```bash
# read data from secondary-1 innodb-0.innodb-pods.demo
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-0.innodb-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

# read data from secondary-2 innodb-2.innodb-pods.demo
$ kubectl exec -it -n demo innodb-0 -c mysql -- mysql -u root --password='ny5jSirIzVtWDcZ7' --host=innodb-2.innodb-pods.demo -e "SELECT * FROM playground.equipment;"
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
kubectl delete -n demo my/innodb
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLDBVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
