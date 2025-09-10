---
title: Failover & Disaster Recovery Overview Microsoft SQL Server
menu:
  docs_{{ .version }}:
    identifier: ms-failover-disaster-recovery
    name: Overview
    parent: FerretDB-fdr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ensuring Rock-Solid FerretDB Uptime

## High Availability with KubeDB: Auto-Failover and Disaster Recovery
In today’s data-driven landscape, database downtime is more than just an inconvenience,
it can lead to serious business disruptions. For teams deploying stateful applications on Kubernetes,
ensuring the high availability and resiliency of FerretDB is critical. That’s where KubeDB comes
in a cloud-native database management solution purpose built for Kubernetes.

One of the standout features of KubeDB is its native support for High Availability (HA) and
automated failover for FerretDB. The KubeDB operator works in tandem with a dedicated database
sidecar to monitor the health of your FerretDB cluster in real time. In the event of a node or
leader failure, the operator automatically initiates a failover process, promoting a healthy secondary
replica to take over with minimal disruption.

This article explores how KubeDB handles automated failover for FerretDB. You’ll learn how to
deploy an Availability Group cluster on Kubernetes using KubeDB and then simulate a failure scenario to observe its
self-healing and auto-recovery mechanisms in action.

By the end of this guide, you’ll gain a deeper understanding of how KubeDB ensures that your
FerretDB workloads remain highly available—even in the face of failure.

> You will see how fast the failover happens when it's truly necessary. Failover in KubeDB-managed
FerretDB will generally happen within 2–10 seconds depending on your cluster networking. There is
an exception scenario that we discussed later in this doc where failover might take a bit longer up
to 45 seconds. But that is a bit rare though.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  2m5s

  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available FerretDBVersion

When you have installed KubeDB, it has created `FerretDBVersion` CR for all supported FerretDB versions.

```bash
$ kubectl get ferretdbversions
NAME     VERSION   DB_IMAGE                                  DEPRECATED   AGE
1.18.0   1.18.0    ghcr.io/appscode-images/ferretdb:1.18.0                104m
1.23.0   1.23.0    ghcr.io/appscode-images/ferretdb:1.23.0                104m
1.24.0   1.24.0    ghcr.io/appscode-images/ferretdb:1.24.0                104m
2.0.0    2.0.0     ghcr.io/appscode-images/ferretdb:2.0.0                 5d4h
```

## Create a FerretDB database

FerretDB use Postgres as it's main backend. Currently, KubeDB supports Postgres backend as database engine for FerretDB. KubeDB operator will create and manage the backend Postgres for FerretDB

Below is the `FerretDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferret
  namespace: demo
spec:
  version: "2.0.0"
  backend:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
  deletionPolicy: WipeOut
  server:
    primary:
      replicas: 2
    secondary:
      replicas: 2
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/quickstart/ferretdb-internal.yaml
ferretdb.kubedb.com/ferret created
```

You can monitor the status until all pods are ready:
```shell
watch kubectl get ms,petset,pods -n demo
```
See the database is ready.

```shell
$ kubectl get ferretdb,petset,pods -n demo
NAME                         NAMESPACE   VERSION   STATUS   AGE
ferretdb.kubedb.com/ferret   demo        2.0.0     Ready    2m54s

NAME                                             AGE
petset.apps.k8s.appscode.com/ferret              2m1s
petset.apps.k8s.appscode.com/ferret-pg-backend   2m51s
petset.apps.k8s.appscode.com/ferret-secondary    2m1s

NAME                      READY   STATUS    RESTARTS   AGE
pod/ferret-0              1/1     Running   0          2m1s
pod/ferret-1              1/1     Running   0          2m
pod/ferret-pg-backend-0   2/2     Running   0          2m50s
pod/ferret-pg-backend-1   2/2     Running   0          2m43s
pod/ferret-pg-backend-2   2/2     Running   0          2m36s
pod/ferret-secondary-0    1/1     Running   0          2m1s
pod/ferret-secondary-1    1/1     Running   0          2m

```

Inspect who is `Primary` and who is `Secondary`.

```shell

$  kubectl get pods -n demo --show-labels | grep role
ferret-pg-backend-0   2/2     Running   0          3m12s   app.kubernetes.io/component=database,app.kubernetes.io/instance=ferret-pg-backend,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=ferret-pg-backend-6766dfcdd7,kubedb-role=primary,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=ferret-pg-backend-0
ferret-pg-backend-1   2/2     Running   0          113s    app.kubernetes.io/component=database,app.kubernetes.io/instance=ferret-pg-backend,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=ferret-pg-backend-6766dfcdd7,kubedb-role=standby,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=ferret-pg-backend-1
ferret-pg-backend-2   2/2     Running   0          110s    app.kubernetes.io/component=database,app.kubernetes.io/instance=ferret-pg-backend,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=ferret-pg-backend-6766dfcdd7,kubedb-role=standby,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=ferret-pg-backend-2
```
The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=secondary` are the secondaries.


Let's create a table in the primary.

```shell
# find the primary pod
$ mongosh 'mongodb://postgres:ZkXB0haW_cCdFr*w@localhost:27017/ferretdb'
Current Mongosh Log ID:	68c00783d567e931f4ce5f46
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.5.8
Using MongoDB:		7.0.77
Using Mongosh:		2.5.8

For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/


To help improve our products, anonymous usage data is collected and sent to MongoDB periodically (https://www.mongodb.com/legal/privacy-policy).
You can opt-out by running the disableTelemetry() command.

------
   The server generated these startup warnings when booting
   2025-09-09T10:54:59.756Z: Powered by FerretDB v2.0.0-1-g7fb2c9a8 and DocumentDB 0.102.0 (PostgreSQL 17.4).
   2025-09-09T10:54:59.756Z: Please star 🌟 us on GitHub: https://github.com/FerretDB/FerretDB and https://github.com/microsoft/documentdb.
   2025-09-09T10:54:59.756Z: The telemetry state is undecided. Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.
------

ferretdb> show dbs
kubedb_system  0 B
ferretdb> use Kaktal
switched to db Kaktal
Kaktal> db.test.insertOne({ created: new Date() })
... 
{
  acknowledged: true,
  insertedId: ObjectId('68c00972d567e931f4ce5f47')
}
Kaktal> show dbs
... 
Kaktal         0 B
kubedb_system  0 B
Kaktal> 

```

Verify the table creation in secondary's.

```shell
$ kubectl exec -it -n demo FerretDB-ag-cluster-1 -c mssql -- bash
mssql@FerretDB-ag-cluster-1:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
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
FerretDB. We use sidecar container with all db pods, and inside that sidecar container,
we use [raft](https://raft.github.io/)protocol to detect the viable primary of the FerretDB
cluster. Raft will choose a db pod as a leader of the FerretDB cluster, we will check if that pod can really run as a leader. If everything is good with that chosen pod, we will run it as primary. This whole process of failover
generally takes less than 10 seconds to complete. So you can expect very rapid failover to ensure high availability of your FerretDB cluster.

Now current running primary is `FerretDB-ag-cluster-0`. Let's open another terminal and run the command below.

```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"

```
It will show current ms cluster roles.
```shell
FerretDB-ag-cluster-0 primary
FerretDB-ag-cluster-1 secondary
FerretDB-ag-cluster-2 secondary

```

#### Case 1: Delete the current primary

Let's delete the current primary and see how the role change happens almost immediately.

```shell
$ kubectl delete pods -n demo FerretDB-ag-cluster-0 
pod "FerretDB-ag-cluster-0" deleted
```
```shell
ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1 primary
ferret-pg-backend-2 standby
ferret-secondary-0
ferret-secondary-1

```

You see almost immediately the failover happened. Here's what happened internally:

- Distributed raft algorithm implementation is running 24 * 7 in your each db sidecar. You can configure this behavior as shown below.
- As soon as `FerretDB-ag-cluster-0` was being deleted and raft inside `FerretDB-ag-cluster-0` senses the termination, it immediately switches the leadership to any other viable leader before termination.
- In our case, raft inside `FerretDB-ag-cluster-2` got the leadership.
- Now this leader switch only means raft leader switch, not the **database leader switch(aka failover)** yet. So `FerretDB-ag-cluster-2` still running as replica. It will be primary after the next step.
- Once raft sidecar inside `FerretDB-ag-cluster-2` see it has become leader of the cluster, it initiates the database failover process and start running as primary.
- So, now `FerretDB-ag-cluster-2` is running as primary.

Now we know how failover is done, let's check if the new primary `FerretDB-ag-cluster-2` is working.

```shell
$ kubectl exec -it -n demo FerretDB-ag-cluster-2 -c mssql -- bash
mssql@FerretDB-ag-cluster-2:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
1> use agdb1
2> go
Changed database context to 'agdb1'.

1> CREATE Table data1
2> go
Msg 102, Level 15, State 1, Server FerretDB-ag-cluster-2, Line 1
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


You will see the deleted pod (`FerretDB-ag-cluster-0`) is brought back by the kubedb operator and it is
now assigned to `secondary role`.

```shell

ferret-0
ferret-1
ferret-pg-backend-0 standby
ferret-pg-backend-1 primary
ferret-pg-backend-2 standby
ferret-secondary-0
ferret-secondary-1

```

Let's check if the secondary(`FerretDB-ag-cluster-0`) got the updated data from new primary `FerretDB-ag-cluster-2`.

```shell
$ kubectl exec -it -n demo FerretDB-ag-cluster-0 -c mssql -- bash
mssql@FerretDB-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
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
Msg 3906, Level 16, State 2, Server FerretDB-ag-cluster-1, Line 1
Failed to update database "agdb1" because the database is read-only.

```

#### Case 2: Delete the current primary and one secondary
```shell
$ kubectl delete pods -n demo ferret-pg-backend-0 ferret-pg-backend-1
pod "ferret-pg-backend-0" deleted
pod "ferret-pg-backend-1" deleted

```
Again we can see the failover happened pretty quickly.
```shell

ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1 
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```

After 10-30 second, the deleted pods will be back and will have its role.

```shell
ferret-0
ferret-1
ferret-pg-backend-0 standby
ferret-pg-backend-1 standby
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```

Let's validate the cluster state from new primary(`FerretDB-ag-cluster-0`).

```shell
$ kubectl exec -it -n demo FerretDB-ag-cluster-0 -c mssql -- bash
mssql@FerretDB-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf"
1> use agdb1
2> go
Changed database context to 'agdb1'.

1> CREATE TABLE data2 (id INT PRIMARY KEY, name NVARCHAR(100), created_at DATETIME DEFAULT GETDATE());
2> go

```

#### Case3: Delete any of the replica's

Let's delete both of the secondary's.

```shell
$ kubectl delete pods -n demo ferret-pg-backend-0 ferret-pg-backend-1
pod "ferret-pg-backend-0" deleted
pod "ferret-pg-backend-1" deleted

```
```shell
ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```

Shortly both of the pods will be back with its role.
```shell
FerretDB-ag-cluster-0 primary
FerretDB-ag-cluster-1 secondary
FerretDB-ag-cluster-2 secondary

```
Let's verify cluster state.
```shell
$ kubectl exec -it -n demo FerretDB-ag-cluster-0 -c mssql -- bash
mssql@FerretDB-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf" 
1> use agdb1
2> go
Changed database context to 'agdb1'.

1> SELECT * FROM sys.dm_hadr_availability_replica_states;
2> go
replica_id                           group_id                             is_local role role_desc           operational_state  operational_state_desc           connected_state connected_state_desc                                         recovery_health recovery_health_desc         Synchronization_health synchronization_health_desc                                  last_connect_error_number last_connect_error_description      last_connect_error_timestamp write_lease_remaining_ticks current_configuration_commit_start_time_utc
------------------------------------ ------------------------------------ -------- ---- ------------------ ------------------  -------------------------------- --------------- ------------------------------------------------------------ --------------- ---------------------------- ---------------------- ------------------------------------------------------------ ------------------------- ----------------------------------- ---------------------------- --------------------------- -------------------------------------------
C4FADE0D-BC82-4D16-95E2-50AA6BE5BD8F BBCC64C9-E0E3-5985-6F01-884248E3DDC6        1    1 PRIMARY                 2                      ONLINE                        1                  CONNECTED                                                   1           ONLINE                                   2 HEALTHY                                                                           NULL        NULL                                      NULL                   9223372036854775807                              NULL
403818D7-CCD6-4EE6-B24C-A61DF3992B1D BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY               NULL                    NULL                         1                  CONNECTED                                                   NULL         NULL                                    2 HEALTHY                                                                           NULL        NULL                                      NULL                        NULL                                        NULL
2F227F4D-29CA-4273-B223-1A54EEB71EFF BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY               NULL                    NULL                         1                  CONNECTED                                                   NULL         NULL                                    2 HEALTHY                                                                           NULL        NULL                                      NULL                        NULL                                        NULL

(3 rows affected)

```

#### Case 4: Delete both primary and all replicas

Let's delete all the pods.

```shell
$ kubectl delete pods -n demo FerretDB-ag-cluster-0 FerretDB-ag-cluster-1 FerretDB-ag-cluster-2
pod "FerretDB-ag-cluster-0" deleted
pod "FerretDB-ag-cluster-1" deleted
pod "FerretDB-ag-cluster-2" deleted

```
```shell
ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1
ferret-pg-backend-2 
ferret-secondary-0
ferret-secondary-1

```

Within 20-30 second, all of the pod should be back.
```shell
ferret-0
ferret-1
ferret-pg-backend-0 standby
ferret-pg-backend-1 standby
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```
Let's verify the cluster state now.

```shell
$  kubectl exec -it -n demo FerretDB-ag-cluster-0 -c mssql -- bash
mssql@FerretDB-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "tZQpzrowQQ20xbCf" 
1> use agdb1
2> go
1> SELECT * FROM sys.dm_hadr_availability_replica_states;
2> go
replica_id                           group_id                             is_local role role_desc                                                    operational_state operational_state_desc                                       connected_state connected_state_desc                                         recovery_health recovery_health_desc                                         synchronization_health synchronization_health_desc                                  last_connect_error_number last_connect_error_description   last_connect_error_timestamp write_lease_remaining_ticks current_configuration_commit_start_time_utc
------------------------------------ ------------------------------------ -------- ---- ------------------------------------------------------------ ----------------- ------------------------------------------------------------ --------------- ------------------------------------------------------------ --------------- ------------------------------------------------------------ ---------------------- ------------------------------------------------------------ ------------------------- -------------------------------- ---------------------------- --------------------------- -------------------------------------------
C4FADE0D-BC82-4D16-95E2-50AA6BE5BD8F BBCC64C9-E0E3-5985-6F01-884248E3DDC6        1    1 PRIMARY                                                                      2 ONLINE                                                                     1 CONNECTED                                                                  1 ONLINE                                                                            2 HEALTHY                                                                           NULL NULL                              NULL                         9223372036854775807                          NULL
403818D7-CCD6-4EE6-B24C-A61DF3992B1D BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY                                                                 NULL NULL                                                                       1 CONNECTED                                                               NULL NULL                                                                              2 HEALTHY                                                                           NULL NULL                              NULL                          NULL                                        NULL
2F227F4D-29CA-4273-B223-1A54EEB71EFF BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY                                                                 NULL NULL                                                                       1 CONNECTED                                                               NULL NULL                                                                              2 HEALTHY                                                                           NULL NULL                              NULL                          NULL                                        NULL

(3 rows affected)

```

> **We make sure the pod with highest lsn for all databases (you can think lsn as the highest data point
available in the databases) always run as primary, so if a case occur where the pod with highest lsn is being
terminated, we will not perform the failover until the highest lsn pod is back online.


### Disaster Scenario and Recovery

#### Scenario

You deploy a `FerretDB` database. The database was running fine. Someday, your database storage becomes full. As your FerretDB process can't write to the filesystem,
clients won't be able to connect to the database. Your database status will be `Not Ready`.

#### Recovery

In order to recover from this, you can create a `VolumeExpansion` `FerretDBOpsRequest` with expanded resource requests.
As soon as you create this, KubeDB will trigger the necessary steps to expand your volume based on your specifications on the `FerretDBOpsRequest` manifest. A sample `FerretDBOpsRequest` manifest for `VolumeExpansion` is given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: msops-vol-exp-ha-demo
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: FerretDB-ag-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Online # see the notes, your storageclass must support this mode
    FerretDB: 20Gi # expanded resource
```


For more details, please check the full section [here](/docs/guides/FerretDB/volume-expansion/overview.md).

> **Note**: There are two ways to update your volume: 1.Online 2.Offline. Which Mode to choose? <br>
It depends on your `StorageClass`. If your storageclass supports online volume expansion, you can go with it. Otherwise, you can go with `Offline` Volume Expansion.

## CleanUp

```shell
$ kubectl delete ms -n demo FerretDB-ag-cluster
# Or, delete the demo
$ kubectl delete ns demo
```


## Next Steps

- Learn about [PITR](/docs/guides/FerretDB/pitr/archiver.md)
- Learn about [backup and restore](/docs/guides/FerretDB/backup/overview/index.md) FerretDB database using Stash.
- Want to setup FerretDB cluster? Check how to [configure Highly Available FerretDB Cluster](/docs/guides/FerretDB/clustering/ag_cluster.md)
- Monitor your FerretDB database with KubeDB using [Prometheus operator](/docs/guides/FerretDB/monitoring/using-prometheus-operator.md).
- Detail concepts of [FerretDB object](/docs/guides/FerretDB/concepts/FerretDB.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).