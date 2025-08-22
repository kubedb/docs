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

## High Availability with KubeDB: Auto-Failover and Disaster Recovery
In today’s data-driven landscape, database downtime is more than just an inconvenience,
it can lead to serious business disruptions. For teams deploying stateful applications on Kubernetes,
ensuring the high availability and resiliency of MSSQLServer is critical. That’s where KubeDB comes 
in a cloud-native database management solution purpose built for Kubernetes.

One of the standout features of KubeDB is its native support for High Availability (HA) and 
automated failover for MSSQLServer. The KubeDB operator works in tandem with a dedicated database 
sidecar to monitor the health of your MSSQLServer cluster in real time. In the event of a node or
leader failure, the operator automatically initiates a failover process, promoting a healthy secondary
replica to take over with minimal disruption.

This article explores how KubeDB handles automated failover for MSSQLServer. You’ll learn how to 
deploy an Availability Group cluster on Kubernetes using KubeDB and then simulate a failure scenario to observe its 
self-healing and auto-recovery mechanisms in action.

By the end of this guide, you’ll gain a deeper understanding of how KubeDB ensures that your 
MSSQLServer workloads remain highly available—even in the face of failure.

> You will see how fast the failover happens when it's truly necessary. Failover in KubeDB-managed
MSSQLServer will generally happen within 2–10 seconds depending on your cluster networking. There is
an exception scenario that we discussed later in this doc where failover might take a bit longer up
to 45 seconds. But that is a bit rare though.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).


- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  5d20h
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available Microsoft SQL Server Versions

When you have installed KubeDB, it has created `MSSQLServerVersion` CR for all supported Microsoft SQL Server versions. Check it by using the `kubectl get mssqlserverversions`. You can also use `msversion` shorthand instead of `mssqlserverversions`.

```bash
$ kubectl get msversion
NAME        VERSION   DB_IMAGE                                                DEPRECATED   AGE
2022-cu12   2022      mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04                9m38s
2022-cu14   2022      mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04                9m38s
2022-cu16   2022      mcr.microsoft.com/mssql/server:2022-CU16-ubuntu-22.04                9m38s
2022-cu19   2022      mcr.microsoft.com/mssql/server:2022-CU19-ubuntu-22.04                9m38s

```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver/ag-cluster/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/ag-cluster/) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Microsoft SQL Server Availability Group Cluster

First, an issuer needs to be created, even if TLS is not enabled for SQL Server. The issuer will be 
used to configure the TLS-enabled Wal-G proxy server, which is required for the SQL Server backup 
and restore operations.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
- Create a secret using the certificate files we have just generated,
```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```
Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/ag-cluster/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

### Step 1: Create a High-Availability MSSQLServer Cluster

First, we need to deploy a MSSQLServer cluster configured for High Availability.
Unlike a Standalone instance, a Availability Group cluster consists of a primary pod
and one or more secondary pods that are ready to take over if the leader
fails.

Save the following YAML  mssqlserver-ag-cluster.yaml. This manifest
defines a 3-node MSSQLServer Availability Group cluster.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-ag-cluster
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      secondaryAccessMode: All
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation # Change it 
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

Now, create the namespace and apply the manifest:

```shell
# Create the namespace if it doesn't exist
$ kubectl create ns demo

# Apply the manifest to deploy the cluster
$ kubectl apply -f mssqlserver-ag-cluster.yaml
mssqlserver.kubedb.com/mssqlserver-ag-cluster created
```

You can monitor the status until all pods are ready:
```shell
watch kubectl get ms,petset,pods -n demo
```
See the database is ready.

```shell
$ kubectl get ms,petset,pods -n demo
NAME                                            VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssqlserver-ag-cluster   2022-cu16   Ready    11m

NAME                                                  AGE
petset.apps.k8s.appscode.com/mssqlserver-ag-cluster   10m

NAME                           READY   STATUS    RESTARTS   AGE
pod/mssqlserver-ag-cluster-0   2/2     Running   0          10m
pod/mssqlserver-ag-cluster-1   2/2     Running   0          8m47s
pod/mssqlserver-ag-cluster-2   2/2     Running   0          8m40s

```

Inspect who is primary and who is secondary.

```shell
# you can inspect who is primary
# and who is secondary like below

$ kubectl get pods -n demo --show-labels | grep role
mssqlserver-ag-cluster-0   2/2     Running   0          12m   app.kubernetes.io/component=database,app.kubernetes.io/instance=mssqlserver-ag-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mssqlservers.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=mssqlserver-ag-cluster-5c944b9596,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=mssqlserver-ag-cluster-0
mssqlserver-ag-cluster-1   2/2     Running   0          11m   app.kubernetes.io/component=database,app.kubernetes.io/instance=mssqlserver-ag-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mssqlservers.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=mssqlserver-ag-cluster-5c944b9596,kubedb.com/role=secondary,statefulset.kubernetes.io/pod-name=mssqlserver-ag-cluster-1
mssqlserver-ag-cluster-2   2/2     Running   0          10m   app.kubernetes.io/component=database,app.kubernetes.io/instance=mssqlserver-ag-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mssqlservers.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=mssqlserver-ag-cluster-5c944b9596,kubedb.com/role=secondary,statefulset.kubernetes.io/pod-name=mssqlserver-ag-cluster-2

```
The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=secondary` are the secondaries.


Let's create a table in the primary.

```shell
# find the primary pod
$ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1 }'
mssqlserver-ag-cluster-0
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

Verify the table creation in secondary's.

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
mssqlserver-ag-cluster-1 secondary
mssqlserver-ag-cluster-2 secondary

```

#### Case 1: Delete the current primary

Let's delete the current primary and see how the role change happens almost immediately.

```shell
$ kubectl delete pods -n demo mssqlserver-ag-cluster-0 
pod "mssqlserver-ag-cluster-0" deleted
```
```shell
mssqlserver-ag-cluster-0 
mssqlserver-ag-cluster-1 secondary
mssqlserver-ag-cluster-2 primary
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
now assigned to `secondary role`.

```shell
mssqlserver-ag-cluster-0 secondary
mssqlserver-ag-cluster-1 secondary
mssqlserver-ag-cluster-2 primary

```

Let's check if the secondary(`mssqlserver-ag-cluster-0`) got the updated data from new primary `mssqlserver-ag-cluster-2`.

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

#### Case 2: Delete the current primary and one secondary
```shell
$ kubectl delete pods -n demo mssqlserver-ag-cluster-1 mssqlserver-ag-cluster-2
pod "mssqlserver-ag-cluster-1" deleted
pod "mssqlserver-ag-cluster-2" deleted
```
Again we can see the failover happened pretty quickly.
```shell
mssqlserver-ag-cluster-0 secondary
mssqlserver-ag-cluster-1 
mssqlserver-ag-cluster-2
```

After 10-30 second, the deleted pods will be back and will have its role.

```shell
mssqlserver-ag-cluster-0 primary
mssqlserver-ag-cluster-1 secondary
mssqlserver-ag-cluster-2 secondary
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

Let's delete both of the secondary's.

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
mssqlserver-ag-cluster-1 secondary
mssqlserver-ag-cluster-2 secondary

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
403818D7-CCD6-4EE6-B24C-A61DF3992B1D BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY               NULL                    NULL                         1                  CONNECTED                                                   NULL         NULL                                    2 HEALTHY                                                                           NULL        NULL                                      NULL                        NULL                                        NULL
2F227F4D-29CA-4273-B223-1A54EEB71EFF BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY               NULL                    NULL                         1                  CONNECTED                                                   NULL         NULL                                    2 HEALTHY                                                                           NULL        NULL                                      NULL                        NULL                                        NULL

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
mssqlserver-ag-cluster-1 secondary
mssqlserver-ag-cluster-2 secondary

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
403818D7-CCD6-4EE6-B24C-A61DF3992B1D BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY                                                                 NULL NULL                                                                       1 CONNECTED                                                               NULL NULL                                                                              2 HEALTHY                                                                           NULL NULL                              NULL                          NULL                                        NULL
2F227F4D-29CA-4273-B223-1A54EEB71EFF BBCC64C9-E0E3-5985-6F01-884248E3DDC6        0    2 SECONDARY                                                                 NULL NULL                                                                       1 CONNECTED                                                               NULL NULL                                                                              2 HEALTHY                                                                           NULL NULL                              NULL                          NULL                                        NULL

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
    MSSQLServer: 20Gi # expanded resource
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