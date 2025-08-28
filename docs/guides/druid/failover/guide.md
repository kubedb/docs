---
title: Failover & Disaster Recovery Overview Microsoft SQL Server
menu:
  docs_{{ .version }}:
    identifier: ms-failover-disaster-recovery
    name: Overview
    parent: mssqlserver-fdr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ensuring Rock-Solid MSSQLServer Uptime

## High Availability with KubeDB: Auto-Failover and Resilient Druid Clusters
In today’s fast-paced world of real-time analytics, even a brief disruption in query serving or 
ingestion pipelines can have serious consequences for business operations. For teams running `Druid`
on Kubernetes, ensuring high availability and continuous data processing is essential. That’s where 
`KubeDB` comes in—a cloud-native database management solution built specifically for Kubernetes.

One of KubeDB’s key strengths is its native support for `High Availability (HA)` and `automated failover`
for `Druid clusters`. The KubeDB operator continuously monitors the health of your `Druid` services—such 
as the **Router, Broker, Coordinator, Overlord, Historical, MiddleManager, and Peon** nodes—ensuring
workloads remain online. If a pod or node running a critical Druid service fails, KubeDB automatically
replaces or reschedules it, allowing traffic to seamlessly fail over to healthy instances with minimal
disruption.


This article explores how KubeDB ensures automated failover for Apache Druid. You’ll learn how to deploy
a highly available Druid cluster on Kubernetes using KubeDB, and how the system handles failure scenarios
with self-healing and auto-recovery mechanisms.

By the end of this guide, you’ll gain a clear understanding of how KubeDB keeps your Druid workloads 
resilient—so ingestion pipelines continue, queries remain served, and coordinators maintain consistent
cluster state even in the face of unexpected failures.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured
to communicate with your cluster. If you do not already have a cluster, you can create one by using
[kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps 
[here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Druid=true` to ensure **Druid
CRD** and `--set global.featureGates.ZooKeeper=true` to ensure **ZooKeeper CRD** as Druid depends 
on ZooKeeper for external dependency  with helm command.

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [guides/druid/ha/overview/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/druid/ha/overview/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Apache Druid. If you just want to try out KubeDB, you can bypass some safety features following the tips [here](/docs/guides/druid/ha/guide/index.md#tips-for-testing).

## Find Available StorageClass

We will have to provide `StorageClass` in Druid CRD specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  14h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Find Available DruidVersion

When you install the KubeDB operator, it registers a CRD named [DruidVersion](/docs/guides/druid/concepts/druidversion.md). The installation process comes with a set of tested DruidVersion objects. Let's check available DruidVersions by,

```bash
$  kubectl get druidversion
NAME     VERSION   DB_IMAGE                               DEPRECATED   AGE
25.0.0   25.0.0    apache/druid:25.0.0                    true         8d
28.0.1   28.0.1    ghcr.io/appscode-images/druid:28.0.1                8d
30.0.0   30.0.0    ghcr.io/appscode-images/druid:30.0.0   true         8d
30.0.1   30.0.1    ghcr.io/appscode-images/druid:30.0.1                8d
31.0.0   31.0.0    ghcr.io/appscode-images/druid:31.0.0                8d

```

Notice the `DEPRECATED` column. Here, `true` means that this DruidVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated DruidVersion. You can also use the short from `drversion` to check available DruidVersions.

In this tutorial, we will use `28.0.1` DruidVersion CR to create a Druid cluster.

## Get External Dependencies Ready

### Deep Storage

One of the external dependency of Druid is deep storage where the segments are stored. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/ha/deep-storage-config.yaml
secret/deep-storage-config created
```

You can also use options like **Amazon S3**, **Google Cloud Storage**, **Azure Blob Storage** or **HDFS** and create a connection information `Secret` like this, and you are good to go.

### Metadata Storage

Druid uses the metadata store to house various metadata about the system, but not to store the actual data. The metadata store retains all metadata essential for a Druid cluster to work. **Apache Derby** is the default metadata store for Druid, however, it is not suitable for production. **MySQL** and **PostgreSQL** are more production suitable metadata stores.

Luckily, **PostgreSQL** and **MySQL** both are readily available in KubeDB as CRD and **KubeDB** operator will automatically create a **MySQL** cluster and create a database in it named `druid` by default.

If you choose to use  **PostgreSQL** as metadata storage, you can simply mention that in the `spec.metadataStorage.type` of the `Druid` CR and KubeDB operator will deploy a `PostgreSQL` cluster for druid to use.

### ZooKeeper

Apache Druid uses [Apache ZooKeeper](https://zookeeper.apache.org/) (ZK) for management of current cluster state i.e. internal service discovery, coordination, and leader election.

Fortunately, KubeDB also has support for **ZooKeeper** and **KubeDB** operator will automatically create a **ZooKeeper** cluster for druid to use.

## Create a Druid HA Cluster

The KubeDB operator implements a Druid CRD to define the specification of Druid.

The Druid instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-ha
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
  serviceTemplates:
    - alias: primary
      spec:
        type: LoadBalancer
        ports:
          - name: routers
            port: 8888
```

Let's create the Druid CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/ha/druid-with-monitoring.yaml
druid.kubedb.com/druid-ha created
```

See the database is ready.

```shell
$  kubectl get druid,petset,pods -n demo
NAME                        TYPE                  VERSION   STATUS   AGE
druid.kubedb.com/druid-ha   kubedb.com/v1alpha2   28.0.1    Ready    23h

NAME                                                   AGE
petset.apps.k8s.appscode.com/druid-ha-brokers          23h
petset.apps.k8s.appscode.com/druid-ha-coordinators     23h
petset.apps.k8s.appscode.com/druid-ha-historicals      23h
petset.apps.k8s.appscode.com/druid-ha-middlemanagers   23h
petset.apps.k8s.appscode.com/druid-ha-mysql-metadata   23h
petset.apps.k8s.appscode.com/druid-ha-routers          23h
petset.apps.k8s.appscode.com/druid-ha-zk               23h

NAME                            READY   STATUS    RESTARTS        AGE
pod/druid-ha-brokers-0          1/1     Running   1 (5h54m ago)   23h
pod/druid-ha-coordinators-0     1/1     Running   1 (5h54m ago)   23h
pod/druid-ha-historicals-0      1/1     Running   1 (5h54m ago)   23h
pod/druid-ha-middlemanagers-0   1/1     Running   1 (5h54m ago)   23h
pod/druid-ha-mysql-metadata-0   2/2     Running   2 (5h54m ago)   20h
pod/druid-ha-mysql-metadata-1   2/2     Running   3 (5h53m ago)   23h
pod/druid-ha-mysql-metadata-2   2/2     Running   2 (5h54m ago)   23h
pod/druid-ha-routers-0          1/1     Running   1 (5h54m ago)   23h
pod/druid-ha-routers-1          1/1     Running   1 (5h54m ago)   23h
pod/druid-ha-zk-0               1/1     Running   1 (5h54m ago)   23h
pod/druid-ha-zk-1               1/1     Running   2 (5h53m ago)   23h
pod/druid-ha-zk-2               1/1     Running   2 (5h53m ago)   23h
pod/myminio-default-0           2/2     Running   2 (5h54m ago)   23h

```

**Check the `kubedb.com/role` label to identify the primary and standby pods.**

```shell
kubectl get pods -n demo --show-labels | grep role
druid-ha-brokers-0          1/1     Running   1 (5h55m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-ha-brokers-69859bc946,kubedb.com/role=brokers,statefulset.kubernetes.io/pod-name=druid-ha-brokers-0
druid-ha-coordinators-0     1/1     Running   1 (5h55m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-ha-coordinators-85968c44,kubedb.com/role=coordinators,statefulset.kubernetes.io/pod-name=druid-ha-coordinators-0
druid-ha-historicals-0      1/1     Running   1 (5h55m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-ha-historicals-667fc997b6,kubedb.com/role=historicals,statefulset.kubernetes.io/pod-name=druid-ha-historicals-0
druid-ha-middlemanagers-0   1/1     Running   1 (5h55m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-ha-middlemanagers-7b57dd64fc,kubedb.com/role=middleManagers,statefulset.kubernetes.io/pod-name=druid-ha-middlemanagers-0
druid-ha-mysql-metadata-0   2/2     Running   2 (5h55m ago)   20h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha-mysql-metadata,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-ha-mysql-metadata-887b7f7b5,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=druid-ha-mysql-metadata-0
druid-ha-mysql-metadata-1   2/2     Running   3 (5h54m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha-mysql-metadata,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=druid-ha-mysql-metadata-887b7f7b5,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=druid-ha-mysql-metadata-1
druid-ha-mysql-metadata-2   2/2     Running   2 (5h55m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha-mysql-metadata,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=druid-ha-mysql-metadata-887b7f7b5,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=druid-ha-mysql-metadata-2
druid-ha-routers-0          1/1     Running   1 (5h55m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-ha-routers-6b896f8fbd,kubedb.com/role=routers,statefulset.kubernetes.io/pod-name=druid-ha-routers-0
druid-ha-routers-1          1/1     Running   1 (5h55m ago)   23h   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-ha,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=druid-ha-routers-6b896f8fbd,kubedb.com/role=routers,statefulset.kubernetes.io/pod-name=druid-ha-routers-1

```
The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=standby` are the secondaries.


Let's check whether you can use `Druid` perfectly  
```shell
# find the primary pod
$ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1 }'
druid-ha-mysql-metadata-2

$ kubectl get secret -n demo mssqlserver-ag-cluster-auth -o jsonpath='{.data.\username}' | base64 -d
sa⏎   
$ kubectl get secret -n demo mssqlserver-ag-cluster-auth -o jsonpath='{.data.\password}' | base64 -d
tZQpzrowQQ20xbCf⏎         
$ kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
1> select name from sys.databases
2> go
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
agdb1                                                                                                                           
agdb2                                                                                                                           
kubedb_system                                                                                                                   

(7 rows affected)
1> use agdb1
2> go 
Changed database context to 'agdb1'.
1> CREATE TABLE data (
2> id INT PRIMARY KEY,
3> name NVARCHAR(100),
4>  created_at DATETIME DEFAULT GETDATE()
5> );
6> go
1> INSERT INTO data (id, name) VALUES (1, 'Alice');
2> INSERT INTO data (id, name) VALUES (2, 'Bob');
3> go

(1 rows affected)

(1 rows affected)
1> SELECT * FROM data;
2> go
id          name                                                                                                 created_at             
----------- ---------------------------------------------------------------------------------------------------- -----------------------
          1 Alice                                                                                                2025-07-31 05:51:06.830
          2 Bob                                                                                                  2025-07-31 05:51:06.847

(2 rows affected)

```

Verify the table creation in standby's.

```shell
$ kubectl exec -it -n demo mssqlserver-ag-cluster-1 -c mssql -- bash
mssql@mssqlserver-ag-cluster-1:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
1> select name from sys.databases
2> go
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
agdb1                                                                                                                           
agdb2                                                                                                                           

(6 rows affected)
1> use agdb1
2> go
Changed database context to 'agdb1'.
1> SELECT * FROM data
2> go
id          name                                                                                                 created_at             
----------- ---------------------------------------------------------------------------------------------------- -----------------------
          1 Alice                                                                                                2025-07-31 05:51:06.830
          2 Bob                                                                                                  2025-07-31 05:51:06.847

(2 rows affected)

```
### Step 2: Simulating a Failover

Before simulating failover, let's discuss how we handle these failover scenarios in KubeDB-managed
MSSQLServer. We use sidecar container with all db pods, and inside that sidecar container,
we use [raft](https://raft.github.io/)protocol to detect the viable primary of the MSSQLServer
cluster. Raft will choose a db pod as a leader of the MSSQLServer cluster, we will check if that pod can really run as a leader. If everything is good with that chosen pod, we will run it as primary. This whole process of failover
generally takes less than 10 seconds to complete. So you can expect very rapid failover to ensure high availability of your MSSQLServer cluster.

Now current running primary is `mssqlserver-ag-cluster-0`. Let's open another terminal and run the command below.

```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"

```
It will show current ms cluster roles.
```shell
mssqlserver-ag-cluster-0 primary
mssqlserver-ag-cluster-1 standby
mssqlserver-ag-cluster-2 standby

```

#### Case 1: Delete the current primary

Let's delete the current primary and see how the role change happens almost immediately.

```shell
$ kubectl delete pods -n demo mssqlserver-ag-cluster-0 
pod "mssqlserver-ag-cluster-0" deleted
```
```shell
druid-ha-brokers-0 brokers
druid-ha-coordinators-0 coordinators
druid-ha-historicals-0 historicals
druid-ha-middlemanagers-0 middleManagers
druid-ha-mysql-metadata-0 primary
druid-ha-mysql-metadata-1 standby
druid-ha-mysql-metadata-2 primary
druid-ha-routers-0 routers
druid-ha-routers-1 routers
druid-ha-zk-0
druid-ha-zk-1
druid-ha-zk-2
myminio-default-0

```

You see almost immediately the failover happened. Here's what happened internally:

- Distributed raft algorithm implementation is running 24 * 7 in your each db sidecar. You can configure this behavior as shown below.
- As soon as `mssqlserver-ag-cluster-0` was being deleted and raft inside `mssqlserver-ag-cluster-0` senses the termination, it immediately switches the leadership to any other viable leader before termination.
- In our case, raft inside `mssqlserver-ag-cluster-2` got the leadership.
- Now this leader switch only means raft leader switch, not the **database leader switch(aka failover)** yet. So `mssqlserver-ag-cluster-2` still running as replica. It will be primary after the next step.
- Once raft sidecar inside `mssqlserver-ag-cluster-2` see it has become leader of the cluster, it initiates the database failover process and start running as primary.
- So, now `mssqlserver-ag-cluster-2` is running as primary.

Now we know how failover is done, let's check if the new primary `mssqlserver-ag-cluster-2` is working.

```shell
$ kubectl exec -it -n demo mssqlserver-ag-cluster-2 -c mssql -- bash
mssql@mssqlserver-ag-cluster-2:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
1> use agdb1
2> go
Changed database context to 'agdb1'.

1> CREATE Table data1
2> go
Msg 102, Level 15, State 1, Server mssqlserver-ag-cluster-2, Line 1
Incorrect syntax near 'data1'.
1> CREATE TABLE data1 (
2> id INT PRIMARY KEY,
3> name NVARCHAR(100),
4> );
5> go
1> SELECT name FROM sys.tables;
2> go
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
data                                                                                                                            
data1                                                                                                                           

(2 rows affected)

```


You will see the deleted pod (`mssqlserver-ag-cluster-0`) is brought back by the kubedb operator and it is
now assigned to `standby role`.

```shell
mssqlserver-ag-cluster-0 standby
mssqlserver-ag-cluster-1 standby
mssqlserver-ag-cluster-2 primary

```

Let's check if the standby(`mssqlserver-ag-cluster-0`) got the updated data from new primary `mssqlserver-ag-cluster-2`.

```shell
$ kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
1> use agdb1
2> go
Changed database context to 'agdb1'.
1> SELECT name FROM sys.tables;
2> go
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
data                                                                                                                            
data1                                                                                                                           

(2 rows affected)
1> CREATE TABLE data (id INT PRIMARY KEY, name NVARCHAR(100), created_at DATETIME DEFAULT GETDATE());
3> go
Msg 3906, Level 16, State 2, Server mssqlserver-ag-cluster-1, Line 1
Failed to update database "agdb1" because the database is read-only.

```

#### Case 2: Delete the current primary and one standby
```shell
$ kubectl delete pods -n demo mssqlserver-ag-cluster-1 mssqlserver-ag-cluster-2
pod "mssqlserver-ag-cluster-1" deleted
pod "mssqlserver-ag-cluster-2" deleted
```
Again we can see the failover happened pretty quickly.
```shell
mssqlserver-ag-cluster-0 standby
mssqlserver-ag-cluster-1 
mssqlserver-ag-cluster-2
```

After 10-30 second, the deleted pods will be back and will have its role.

```shell
mssqlserver-ag-cluster-0 primary
mssqlserver-ag-cluster-1 standby
mssqlserver-ag-cluster-2 standby
```

Let's validate the cluster state from new primary(`mssqlserver-ag-cluster-0`).

```shell
$ kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
1> use agdb1
2> go
Changed database context to 'agdb1'.

1> CREATE TABLE data2 (id INT PRIMARY KEY, name NVARCHAR(100), created_at DATETIME DEFAULT GETDATE());
2> go

```

#### Case3: Delete any of the replica's

Let's delete both of the standby's.

```shell
$ kubectl delete pods -n demo mssqlserver-ag-cluster-1 mssqlserver-ag-cluster-2
pod "mssqlserver-ag-cluster-1" deleted
pod "mssqlserver-ag-cluster-2" deleted

```
```shell
mssqlserver-ag-cluster-0 primary
mssqlserver-ag-cluster-1 
mssqlserver-ag-cluster-2

```

Shortly both of the pods will be back with its role.
```shell
mssqlserver-ag-cluster-0 primary
mssqlserver-ag-cluster-1 standby
mssqlserver-ag-cluster-2 standby

```
Let's verify cluster state.
```shell
$ kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf" 
1> use agdb1
2> go
Changed database context to 'agdb1'.

1> SELECT * FROM sys.dm_hadr_availability_replica_states;
2> go
replica_id                           group_id                             is_local role role_desc           operational_state  operational_state_desc           connected_state connected_state_desc                                         recovery_health recovery_health_desc         Synchronization_health synchronization_health_desc                                  last_connect_error_number last_connect_error_description      last_connect_error_timestamp write_lease_remaining_ticks current_configuration_commit_start_time_utc
------------------------------------ ------------------------------------ -------- ---- ------------------ ------------------  -------------------------------- --------------- ------------------------------------------------------------ --------------- ---------------------------- ---------------------- ------------------------------------------------------------ ------------------------- ----------------------------------- ---------------------------- --------------------------- -------------------------------------------
C4FADE0D-BC82-4D16-95E2-50AA6BE5BD8F BBCC64C9-E0E3-5985-6F01-884248E3DDC6        1    1 PRIMARY                 2                      ONLINE                        1                  CONNECTED                                                   1           ONLINE                                   2 HEALTHY                                                                           NULL        NULL                                      NULL                   9223372036854775807                              NULL
403818D7-CCD6-4EE6-B24C-A61DF3992B1D BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 standby               NULL                    NULL                         1                  CONNECTED                                                   NULL         NULL                                    2 HEALTHY                                                                           NULL        NULL                                      NULL                        NULL                                        NULL
2F227F4D-29CA-4273-B223-1A54EEB71EFF BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 standby               NULL                    NULL                         1                  CONNECTED                                                   NULL         NULL                                    2 HEALTHY                                                                           NULL        NULL                                      NULL                        NULL                                        NULL

(3 rows affected)

```

#### Case 4: Delete both primary and all replicas

Let's delete all the pods.

```shell
$ kubectl delete pods -n demo mssqlserver-ag-cluster-0 mssqlserver-ag-cluster-1 mssqlserver-ag-cluster-2
pod "mssqlserver-ag-cluster-0" deleted
pod "mssqlserver-ag-cluster-1" deleted
pod "mssqlserver-ag-cluster-2" deleted

```
```shell
mssqlserver-ag-cluster-0 
mssqlserver-ag-cluster-1
mssqlserver-ag-cluster-2
```

Within 20-30 second, all of the pod should be back.
```shell
mssqlserver-ag-cluster-0 primary
mssqlserver-ag-cluster-1 standby
mssqlserver-ag-cluster-2 standby

```
Let's verify the cluster state now.

```shell
$  kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf" 
1> use agdb1
2> go
1> SELECT * FROM sys.dm_hadr_availability_replica_states;
2> go
replica_id                           group_id                             is_local role role_desc                                                    operational_state operational_state_desc                                       connected_state connected_state_desc                                         recovery_health recovery_health_desc                                         synchronization_health synchronization_health_desc                                  last_connect_error_number last_connect_error_description   last_connect_error_timestamp write_lease_remaining_ticks current_configuration_commit_start_time_utc
------------------------------------ ------------------------------------ -------- ---- ------------------------------------------------------------ ----------------- ------------------------------------------------------------ --------------- ------------------------------------------------------------ --------------- ------------------------------------------------------------ ---------------------- ------------------------------------------------------------ ------------------------- -------------------------------- ---------------------------- --------------------------- -------------------------------------------
C4FADE0D-BC82-4D16-95E2-50AA6BE5BD8F BBCC64C9-E0E3-5985-6F01-884248E3DDC6        1    1 PRIMARY                                                                      2 ONLINE                                                                     1 CONNECTED                                                                  1 ONLINE                                                                            2 HEALTHY                                                                           NULL NULL                              NULL                         9223372036854775807                          NULL
403818D7-CCD6-4EE6-B24C-A61DF3992B1D BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 standby                                                                 NULL NULL                                                                       1 CONNECTED                                                               NULL NULL                                                                              2 HEALTHY                                                                           NULL NULL                              NULL                          NULL                                        NULL
2F227F4D-29CA-4273-B223-1A54EEB71EFF BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 standby                                                                 NULL NULL                                                                       1 CONNECTED                                                               NULL NULL                                                                              2 HEALTHY                                                                           NULL NULL                              NULL                          NULL                                        NULL

(3 rows affected)

```

> **We make sure the pod with highest lsn for all databases (you can think lsn as the highest data point
available in the databases) always run as primary, so if a case occur where the pod with highest lsn is being
terminated, we will not perform the failover until the highest lsn pod is back online.


### Disaster Scenario and Recovery

#### Scenario

You deploy a `MSSQLServer` database. The database was running fine. Someday, your database storage becomes full. As your MSSQLServer process can't write to the filesystem,
clients won't be able to connect to the database. Your database status will be `Not Ready`.

#### Recovery

In order to recover from this, you can create a `VolumeExpansion` `MSSQLServerOpsRequest` with expanded resource requests.
As soon as you create this, KubeDB will trigger the necessary steps to expand your volume based on your specifications on the `MSSQLServerOpsRequest` manifest. A sample `MSSQLServerOpsRequest` manifest for `VolumeExpansion` is given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-vol-exp-ha-demo
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: mssqlserver-ag-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Online # see the notes, your storageclass must support this mode
    mssqlserver: 20Gi # expanded resource
```


For more details, please check the full section [here](/docs/guides/mssqlserver/volume-expansion/overview.md).

> **Note**: There are two ways to update your volume: 1.Online 2.Offline. Which Mode to choose? <br>
It depends on your `StorageClass`. If your storageclass supports online volume expansion, you can go with it. Otherwise, you can go with `Offline` Volume Expansion.

## CleanUp

```shell
$ kubectl delete ms -n demo mssqlserver-ag-cluster
# Or, delete the demo
$ kubectl delete ns demo
```


## Next Steps

- Learn about [PITR](/docs/guides/mssqlserver/pitr/archiver.md)
- Learn about [backup and restore](/docs/guides/mssqlserver/backup/overview/index.md) MSSQLServer database using Stash.
- Want to setup MSSQLServer cluster? Check how to [configure Highly Available MSSQLServer Cluster](/docs/guides/mssqlserver/clustering/ag_cluster.md)
- Monitor your MSSQLServer database with KubeDB using [Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).