---
title: SQL Server Availability Group Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: ms-clustering-availability-group
    name: Availability Group (HA Cluster)
    parent: ms-clustering
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Microsoft SQL Server Availability Group Cluster 

This tutorial will show you how to use KubeDB to run Microsoft SQL Server Availability Group Cluster.

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
2022-cu12   2022      mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04                17h
2022-cu14   2022      mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04                17h
```


> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver/ag-cluster/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/ag-cluster/) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).



## Deploy Microsoft SQL Server Availability Group Cluster 

First, an issuer needs to be created, even if TLS is not enabled for SQL Server. The issuer will be used to configure the TLS-enabled Wal-G proxy server, which is required for the SQL Server backup and restore operations.

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

### Configuring Environment Variables for SQL Server on Linux
You can use environment variables to configure SQL Server on Linux containers.
When deploying `Microsoft SQL Server` on Linux using `containers`, you need to specify the `product edition` through the [MSSQL_PID](https://mcr.microsoft.com/en-us/product/mssql/server/about#configuration:~:text=MSSQL_PID%20is%20the,documentation%20here.) environment variable. This variable determines which `SQL Server edition` will run inside the container. The acceptable values for `MSSQL_PID` are:   
`Developer`: This will run the container using the Developer Edition (this is the default if no MSSQL_PID environment variable is supplied)    
`Express`: This will run the container using the Express Edition    
`Standard`: This will run the container using the Standard Edition   
`Enterprise`: This will run the container using the Enterprise Edition   
`EnterpriseCore`: This will run the container using the Enterprise Edition Core   
`<valid product id>`: This will run the container with the edition that is associated with the PID

`ACCEPT_EULA` confirms your acceptance of the [End-User Licensing Agreement](https://go.microsoft.com/fwlink/?linkid=857698).

For a complete list of environment variables that can be used, refer to the documentation [here](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-environment-variables?view=sql-server-2017).

Below is an example of how to configure the `MSSQL_PID` and `ACCEPT_EULA` environment variable in the KubeDB MSSQLServer Custom Resource Definition (CRD):
```bash
metadata:
  name: mssqlserver
  namespace: demo
spec:
  podTemplate:
    spec:
      containers:
      - name: mssql
        env:
        - name: ACCEPT_EULA
          value: "Y"
        - name: MSSQL_PID
          value: Enterprise
```
In this example, the SQL Server container will run the Enterprise Edition.

### Deploy Microsoft SQL Server Availability Group Cluster 
KubeDB implements a `MSSQLServer` CRD to define the specification of a Microsoft SQL Server database. Below is the `MSSQLServer` object created in this tutorial.

Here, our issuer `mssqlserver-ca-issuer` is ready to deploy a `MSSQLServer`. Below is the YAML of SQL Server that we are going to create,


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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/ag-cluster/mssqlserver-ag-cluster.yaml
mssqlserver.kubedb.com/mssqlserver-ag-cluster created
```

Here,

- `spec.version` is the name of the MSSQLServerVersion CR where the docker images are specified. In this tutorial, a MSSQLServer `2022-cu12` database is going to be created.
- `spec.replicas` denotes the number of replicas of the created availability group
- `spec.topology` specifies the mode `AvailabilityGroup` and the list of names of the databases that we want in our availability group. 
   KubeDB operator will create and add these databases to the created availability group automatically. User don't have to create, configure or add the database to the availability group manually. User can update this list later as well. 
- `spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). Here `tls.clientTLS: false` means tls will not be enabled for SQL Server but the Issuer will be used to configure tls enabled wal-g proxy-server which is required for SQL Server backup operation.
- `spec.storageType` specifies the type of storage that will be used for MSSQLServer database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MSSQLServer database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MSSQLServer` CR or which resources KubeDB should keep or delete when you delete `MSSQLServer` CR. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy)

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in storage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MSSQLServer` objects using Kubernetes api. When a `MSSQLServer` object is created, KubeDB operator will create a new PetSet and a Service with the matching MSSQLServer object name. KubeDB operator will also create a governing service for PetSets with the name `<MSSQLServerName>-pods`, if one is not already present.

Let's see the sql server resources that are created. 
```bash
$ kubectl get ms,petset,pod,svc,secret,issuer,pvc -n demo

NAME                                            VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssqlserver-ag-cluster   2022-cu12   Ready    178m

NAME                                                  AGE
petset.apps.k8s.appscode.com/mssqlserver-ag-cluster   177m

NAME                           READY   STATUS    RESTARTS   AGE
pod/mssqlserver-ag-cluster-0   2/2     Running   0          177m
pod/mssqlserver-ag-cluster-1   2/2     Running   0          177m
pod/mssqlserver-ag-cluster-2   2/2     Running   0          177m

NAME                                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/mssqlserver-ag-cluster             ClusterIP   10.96.60.223   <none>        1433/TCP   178m
service/mssqlserver-ag-cluster-pods        ClusterIP   None           <none>        1433/TCP   178m
service/mssqlserver-ag-cluster-secondary   ClusterIP   10.96.9.35     <none>        1433/TCP   178m

NAME                                          TYPE                       DATA   AGE
secret/mssqlserver-ag-cluster-auth            kubernetes.io/basic-auth   2      178m
secret/mssqlserver-ag-cluster-client-cert     kubernetes.io/tls          3      178m
secret/mssqlserver-ag-cluster-config          Opaque                     1      178m
secret/mssqlserver-ag-cluster-dbm-login       kubernetes.io/basic-auth   1      178m
secret/mssqlserver-ag-cluster-endpoint-cert   kubernetes.io/tls          3      178m
secret/mssqlserver-ag-cluster-master-key      kubernetes.io/basic-auth   1      178m
secret/mssqlserver-ag-cluster-server-cert     kubernetes.io/tls          3      178m

NAME                                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mssqlserver-ag-cluster-0   Bound    pvc-33ae1829-c559-407b-a148-1792c22b52a6   1Gi        RWO            standard       177m
persistentvolumeclaim/data-mssqlserver-ag-cluster-1   Bound    pvc-b697b7ad-8348-431f-b2c7-01620bec4f8d   1Gi        RWO            standard       177m
persistentvolumeclaim/data-mssqlserver-ag-cluster-2   Bound    pvc-b486a79c-a8ae-449a-bc15-74491f062573   1Gi        RWO            standard       177m
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created and is able to accept client connections. Run the following command to see the modified MSSQLServer object:

```bash
$ kubectl get ms -n demo mssqlserver-ag-cluster -o yaml
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  creationTimestamp: "2024-10-08T10:56:30Z"
  finalizers:
    - kubedb.com
  generation: 2
  name: mssqlserver-ag-cluster
  namespace: demo
  resourceVersion: "238146"
  uid: 0d383530-bbb5-442f-bef0-562304539f98
spec:
  authSecret:
    name: mssqlserver-ag-cluster-auth
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
        - name: mssql
          resources:
            limits:
              memory: 4Gi
            requests:
              cpu: 500m
              memory: 4Gi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              add:
                - NET_BIND_SERVICE
              drop:
                - ALL
            runAsGroup: 10001
            runAsNonRoot: true
            runAsUser: 10001
            seccompProfile:
              type: RuntimeDefault
        - name: mssql-coordinator
          resources:
            limits:
              memory: 256Mi
            requests:
              cpu: 200m
              memory: 256Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsGroup: 10001
            runAsNonRoot: true
            runAsUser: 10001
            seccompProfile:
              type: RuntimeDefault
      initContainers:
        - name: mssql-init
          resources:
            limits:
              memory: 512Mi
            requests:
              cpu: 200m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsGroup: 10001
            runAsNonRoot: true
            runAsUser: 10001
            seccompProfile:
              type: RuntimeDefault
      podPlacementPolicy:
        name: default
      securityContext:
        fsGroup: 10001
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  tls:
    certificates:
      - alias: server
        secretName: mssqlserver-ag-cluster-server-cert
        subject:
          organizationalUnits:
            - server
          organizations:
            - kubedb
      - alias: client
        secretName: mssqlserver-ag-cluster-client-cert
        subject:
          organizationalUnits:
            - client
          organizations:
            - kubedb
    clientTLS: false
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssqlserver-ca-issuer
  topology:
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
    mode: AvailabilityGroup
  version: 2022-cu12
status:
  conditions:
    - lastTransitionTime: "2024-10-08T10:56:30Z"
      message: 'The KubeDB operator has started the provisioning of MSSQL: demo/mssqlserver-ag-cluster'
      observedGeneration: 2
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2024-10-08T10:57:55Z"
      message: All replicas are ready for MSSQLServer demo/mssqlserver-ag-cluster
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2024-10-08T10:58:14Z"
      message: Primary replica is ready for MSSQL demo/mssqlserver-ag-cluster
      observedGeneration: 2
      reason: AvailabilyGroupCreatedInPrimary
      status: "True"
      type: AGPrimaryReplicaReady
    - lastTransitionTime: "2024-10-08T10:58:48Z"
      message: database demo/mssqlserver-ag-cluster is accepting connection
      observedGeneration: 2
      reason: AcceptingConnection
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2024-10-08T10:58:48Z"
      message: database demo/mssqlserver-ag-cluster is ready
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: Ready
    - lastTransitionTime: "2024-10-08T10:59:18Z"
      message: 'The MSSQL: demo/mssqlserver-ag-cluster is successfully provisioned.'
      observedGeneration: 2
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 1
  phase: Ready
```

## Connect with MSSQLServer database

KubeDB operator has created a new Secret called `mssqlserver-ag-cluster-auth` *(format: {mssqlserver-object-name}-auth)* for storing the sa password for `sql server`. This secret contains a `username` key which contains the *username* for MSSQLServer SA and a `password` key which contains the *password* for MSSQLServer SA user.

If you want to use an existing secret please specify that when creating the MSSQLServer object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `sa` as value of `username` and a strong password for the sa user. For more details see [here](/docs/guides/mssqlserver/concepts/mssqlserver.md#specauthsecret).

Now, we need `username` and `password` to connect to this database from `kubectl exec` command. In this example  `mssqlserver-ag-cluster-auth` secret holds username and password

```bash
$ kubectl get secret -n demo mssqlserver-ag-cluster-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secret -n demo mssqlserver-ag-cluster-auth -o jsonpath='{.data.\password}' | base64 -d
wFKDGnWgFP5Rdv92
```
We can exec into the pod `mssqlserver-ag-cluster-0` using the following command:
```bash
kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$
```

You can connect to the database using the `sqlcmd` utility, which comes with the mssql-tools package on Linux. To check the installed version of sqlcmd, run the following command:
```bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd "-?"
Microsoft (R) SQL Server Command Line Tool
Version 17.10.0001.1 Linux
Copyright (C) 2017 Microsoft Corporation. All rights reserved.

usage: sqlcmd            [-U login id]          [-P password]
  [-S server or Dsn if -D is provided] 
  [-H hostname]          [-E trusted connection]
  [-N Encrypt Connection][-C Trust Server Certificate]
  [-d use database name] [-l login timeout]     [-t query timeout]
  [-h headers]           [-s colseparator]      [-w screen width]
  [-a packetsize]        [-e echo input]        [-I Enable Quoted Identifiers]
  [-c cmdend]
  [-q "cmdline query"]   [-Q "cmdline query" and exit]
  [-m errorlevel]        [-V severitylevel]     [-W remove trailing spaces]
  [-u unicode output]    [-r[0|1] msgs to stderr]
  [-i inputfile]         [-o outputfile]
  [-k[1|2] remove[replace] control characters]
  [-y variable length type display width]
  [-Y fixed length type display width]
  [-p[1] print statistics[colon format]]
  [-R use client regional setting]
  [-K application intent]
  [-M multisubnet failover]
  [-b On error batch abort]
  [-D Dsn flag, indicate -S is Dsn] 
  [-X[1] disable commands, startup script, environment variables [and exit]]
  [-x disable variable substitution]
  [-g enable column encryption]
  [-G use Azure Active Directory for authentication]
  [-? show syntax summary]
```


Now, connect to the database using username and password, check the name of the created availability group, replicas of the availability group and see if databases are added to the availability group.
```bash
$ kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "wFKDGnWgFP5Rdv92"
1> select name from sys.databases
2> go
name                                                  
----------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
agdb1                                                                                                                           
agdb2                                                                                                                           
kubedb_system                                                                                                                   

(5 rows affected)
1> SELECT name FROM sys.availability_groups
2> go
name                                                                                                                            
----------------------------------------------------------------------------
mssqlserveragcluster                                                                                                            

(1 rows affected)
1> select replica_server_name from sys.availability_replicas;
2> go
replica_server_name                                                                                                                                                                                                                                             
-------------------------------------------------------------------------------------------
mssqlserver-ag-cluster-0                                                                                                                                                                                                                                        
mssqlserver-ag-cluster-1                                                                                                                                                                                                                                        
mssqlserver-ag-cluster-2      
(3 rows affected)
1> select database_name	from sys.availability_databases_cluster;
2> go
database_name                                                                                                                   
------------------------------------------------------------------------------------------
agdb1                                                                                                                           
agdb2                                                                                                                           

(2 rows affected)

```


Now, to check the redundancy and data availability in secondary members. Let's insert some data into the primary database of sql server availability group and see if data replication is working fine. First we have to determine the primary replica, as data writes are only permitted on the primary node.


```bash
$ kubectl get pods -n demo --selector=app.kubernetes.io/instance=mssqlserver-ag-cluster -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\n"}{end}'
mssqlserver-ag-cluster-0	primary
mssqlserver-ag-cluster-1	secondary
mssqlserver-ag-cluster-2	secondary
```

From the output above, we can see that mssqlserver-ag-cluster-0 is the primary node. To insert data, log into the primary MSSQLServer pod. Use the following command,

```bash
$ kubectl exec -it mssqlserver-ag-cluster-0 -c mssql -n demo -- bash
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "wFKDGnWgFP5Rdv92"
1> SELECT database_name FROM sys.availability_databases_cluster
2> go
database_name                                                                                                                   
----------------------------------------------------------------------------------------------
agdb1                                                                                                                           
agdb2                                                                                                                           

(2 rows affected)
1> use agdb1
2> go
Changed database context to 'agdb1'.
1> CREATE TABLE Data (ID INT, NAME NVARCHAR(255), AGE INT);
2> go
1> INSERT INTO Data(ID, Name, Age) VALUES (1, 'John Doe', 25), (2, 'Jane Smith', 30);                                                                                              
2> go
(2 rows affected)
1> Select * from data
2> go
ID    NAME                             AGE                                  
--------------------------------------------------------
1     John Doe                         25                                                                
2     Jane Smith                       30

(2 rows affected)
1> 
```
Now, Let's verify that the data inserted into the primary node has been replicated to the secondary nodes.
### Access the inserted data from secondaries
Access the secondary node (Node 2) to verify that the data is present.

```bash
$ kubectl exec -it mssqlserver-ag-cluster-1 -c mssql -n demo -- bash
mssql@mssqlserver-ag-cluster-1:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "wFKDGnWgFP5Rdv92"
1> SELECT database_name FROM sys.availability_databases_cluster
2> go
database_name                                                                                                                   
-------------------------------------------------
agdb1                                                                                                                           
agdb2                                                                                                                           

(2 rows affected)
1> use agdb1
2> go
Changed database context to 'agdb1'.
1> select * from data
2> go
ID    NAME                             AGE                                  
--------------------------------------------------------
1     John Doe                         25                                                                
2     Jane Smith                       30

(2 rows affected)
1> 
```


Now access the secondary node (Node 3) to verify that the data is present.
```bash
$ kubectl exec -it mssqlserver-ag-cluster-2 -c mssql -n demo -- bash
mssql@mssqlserver-ag-cluster-2:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "wFKDGnWgFP5Rdv92"
1> SELECT database_name FROM sys.availability_databases_cluster
2> go
database_name                                                                                                                   
-------------------------------------------------
agdb1                                                                                                                           
agdb2                                                                                                                           

(2 rows affected)
1> use agdb1
2> go
Changed database context to 'agdb1'.
1> select * from data
2> go
ID    NAME                             AGE                                  
--------------------------------------------------------
1     John Doe                         25                                                                
2     Jane Smith                       30

(2 rows affected)
1> 
```


## Automatic Failover
To test automatic failover, we will force the primary member to restart. As the primary member (pod) becomes unavailable, the rest of the members will elect a primary member by election.

```bash
$ kubectl get pods -n demo 
NAME                       READY   STATUS    RESTARTS   AGE
mssqlserver-ag-cluster-0   2/2     Running   0          129m
mssqlserver-ag-cluster-1   2/2     Running   0          129m
mssqlserver-ag-cluster-2   2/2     Running   0          129m

$ kubectl delete pod -n demo mssqlserver-ag-cluster-0 
pod "mssqlserver-ag-cluster-0" deleted

$ kubectl get pods -n demo
NAME                       READY   STATUS    RESTARTS   AGE
mssqlserver-ag-cluster-0   2/2     Running   0          7s
mssqlserver-ag-cluster-1   2/2     Running   0          130m
mssqlserver-ag-cluster-2   2/2     Running   0          130m
```


Now find the new primary pod by running this command. 
```bash
$ kubectl get pods -n demo --selector=app.kubernetes.io/instance=mssqlserver-ag-cluster -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\n"}{end}'
mssqlserver-ag-cluster-0	
mssqlserver-ag-cluster-1	primary
mssqlserver-ag-cluster-2	secondary
```

We can see that, the primary node is now is `mssqlserver-ag-cluster-1`. The old primary pod `mssqlserver-ag-cluster-0` role is still pending. It will be set when old primary joins with the new primary as secondary.  
Lets exec into this new primary and see the availability replica role.
```bash
$ kubectl exec -it mssqlserver-ag-cluster-1 -c mssql -n demo -- bash
mssql@mssqlserver-ag-cluster-1:/$ opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "wFKDGnWgFP5Rdv92"
1> SELECT ar.replica_server_name, ars.role_desc
2> FROM sys.dm_hadr_availability_replica_states ars
3> INNER JOIN sys.availability_replicas ar ON ars.replica_id = ar.replica_id;
4> go
replica_server_name                                                                                                                                                                                                                                              role_desc                                                   
-------------------------------------------------------------------------------------------------------
mssqlserver-ag-cluster-0                                                                                                                                                                                                                                         SECONDARY                                                   
mssqlserver-ag-cluster-1                                                                                                                                                                                                                                         PRIMARY                                                     
mssqlserver-ag-cluster-2                                                                                                                                                                                                                                         SECONDARY                                                   

(3 rows affected)

```

We can see that new primary is `mssqlserver-ag-cluster-1` and the old primary `mssqlserver-ag-cluster-0` joined the availability group cluster as secondary. MSSQLServer status is `Ready` now. We can see the updated pod labels. 
```bash
$ kubectl get pods -n demo --selector=app.kubernetes.io/instance=mssqlserver-ag-cluster -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\n"}{end}'
mssqlserver-ag-cluster-0	secondary
mssqlserver-ag-cluster-1	primary
mssqlserver-ag-cluster-2	secondary
````


Run the following command to see the created appbinding object:
```bash
$ kubectl get appbinding -n demo -oyaml
```

```yaml
apiVersion: v1
items:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    kind: AppBinding
    metadata:
      creationTimestamp: "2024-10-08T10:57:14Z"
      generation: 1
      labels:
        app.kubernetes.io/component: database
        app.kubernetes.io/instance: mssqlserver-ag-cluster
        app.kubernetes.io/managed-by: kubedb.com
        app.kubernetes.io/name: mssqlservers.kubedb.com
      name: mssqlserver-ag-cluster
      namespace: demo
      ownerReferences:
        - apiVersion: kubedb.com/v1alpha2
          blockOwnerDeletion: true
          controller: true
          kind: MSSQLServer
          name: mssqlserver-ag-cluster
          uid: 0d383530-bbb5-442f-bef0-562304539f98
      resourceVersion: "237817"
      uid: d7f2f0fc-251f-4ea4-824b-6eb3c96897b4
    spec:
      appRef:
        apiGroup: kubedb.com
        kind: MSSQLServer
        name: mssqlserver-ag-cluster
        namespace: demo
      clientConfig:
        service:
          name: mssqlserver-ag-cluster
          path: /
          port: 1433
          scheme: tcp
      secret:
        name: mssqlserver-ag-cluster-auth
      type: kubedb.com/mssqlserver
      version: "2022"
kind: List
metadata:
  resourceVersion: ""
```

You can use this appbinding to connect with the mssql server from external



## Database DeletionPolicy

This field is used to regulate the deletion process of the related resources when `MSSQLServer` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl patch -n demo ms mssqlserver-ag-cluster -p '{"spec":{"deletionPolicy":"DoNotTerminate"}}' --type="merge"
mssqlserver.kubedb.com/mssqlserver-ag-cluster patched

$ kubectl delete ms -n demo mssqlserver-ag-cluster
The MSSQLServer "mssqlserver-ag-cluster" is invalid: spec.deletionPolicy: Invalid value: "mssqlserver-ag-cluster": Can not delete as deletionPolicy is set to "DoNotTerminate"
```

Now, run `kubectl patch -n demo ms mssqlserver-ag-cluster -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"` to set `spec.deletionPolicy` to `Halt` (which deletes the mssqlserver object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.

Learn details of all `DeletionPolicy` [here](/docs/guides/mssqlserver/concepts/mssqlserver.md#specdeletionpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `MSSQLServer` object `deletionPolicy` to `Halt`.

When the [DeletionPolicy](/docs/guides/mssqlserver/concepts/mssqlserver.md#specdeletionpolicy) is set to `halt` and the MSSQLServer object is deleted, the KubeDB operator will delete the PetSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `deletionPolicy` to `halt` in existing database using `patch` command for testing.

At first, run `kubectl patch -n demo ms mssqlserver-ag-cluster -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"`. Then delete the mssqlserver object,

```bash
$ kubectl delete ms -n demo mssqlserver-ag-cluster
mssqlserver.kubedb.com "mssqlserver-ag-cluster" deleted
```

Now, run the following command to get mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get ms,petset,pod,svc,secret,pvc -n demo 
NAME                                          TYPE                       DATA   AGE
secret/mssqlserver-ag-cluster-auth            kubernetes.io/basic-auth   2      3h6m
secret/mssqlserver-ag-cluster-client-cert     kubernetes.io/tls          3      3h6m
secret/mssqlserver-ag-cluster-config          Opaque                     1      3h6m
secret/mssqlserver-ag-cluster-dbm-login       kubernetes.io/basic-auth   1      3h6m
secret/mssqlserver-ag-cluster-endpoint-cert   kubernetes.io/tls          3      3h6m
secret/mssqlserver-ag-cluster-master-key      kubernetes.io/basic-auth   1      3h6m
secret/mssqlserver-ag-cluster-server-cert     kubernetes.io/tls          3      3h6m
secret/mssqlserver-ca                         kubernetes.io/tls          2      3h8m

NAME                                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mssqlserver-ag-cluster-0   Bound    pvc-33ae1829-c559-407b-a148-1792c22b52a6   1Gi        RWO            standard       3h6m
persistentvolumeclaim/data-mssqlserver-ag-cluster-1   Bound    pvc-b697b7ad-8348-431f-b2c7-01620bec4f8d   1Gi        RWO            standard       3h6m
persistentvolumeclaim/data-mssqlserver-ag-cluster-2   Bound    pvc-b486a79c-a8ae-449a-bc15-74491f062573   1Gi        RWO            standard       3h5m
```

From the above output, you can see that all mssqlserver resources(`MSSQLServer`, `PetSet`, `Pod`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your mssqlserver again using these resources.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `MSSQLServer` object `deletionPolicy` to `Delete`. In this setting, `PetSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [DeletionPolicy](/docs/guides/mssqlserver/concepts/mssqlserver.md#specdeletionpolicy) is set to `Delete` and the MSSQLServer object is deleted, the KubeDB operator will delete the PetSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `deletionPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl patch -n demo ms mssqlserver-ag-cluster -p '{"spec":{"deletionPolicy":"Delete"}}' --type="merge"
mssqlserver.kubedb.com/mssqlserver-ag-cluster patched
$ kubectl delete ms -n demo mssqlserver-ag-cluster 
mssqlserver.kubedb.com "mssqlserver-ag-cluster" deleted
```

Now, run the following command to get all mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get ms,petset,pod,svc,secret,pvc -n demo 
NAME                                          TYPE                       DATA   AGE
secret/mssqlserver-ag-cluster-auth            kubernetes.io/basic-auth   2      3h6m
secret/mssqlserver-ag-cluster-client-cert     kubernetes.io/tls          3      3h6m
secret/mssqlserver-ag-cluster-config          Opaque                     1      3h6m
secret/mssqlserver-ag-cluster-dbm-login       kubernetes.io/basic-auth   1      3h6m
secret/mssqlserver-ag-cluster-endpoint-cert   kubernetes.io/tls          3      3h6m
secret/mssqlserver-ag-cluster-master-key      kubernetes.io/basic-auth   1      3h6m
secret/mssqlserver-ag-cluster-server-cert     kubernetes.io/tls          3      3h6m
secret/mssqlserver-ca                         kubernetes.io/tls          2      3h8m
```

From the above output, you can see that all mssqlserver resources(`MSSQLServer`, `PetSet`, `Pod`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your mssqlserver using `snapshots`(if previously taken) and `Secrets`.

>If you don't set the deletionPolicy then the kubeDB set the DeletionPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `MSSQLServer` database and relevant resources without any tracking by setting `deletionPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `MSSQLServer` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `deletionPolicy` is set to `WipeOut`.

Let's set `deletionPolicy` set to `WipeOut` and delete the database using the following command:

```yaml
$ kubectl patch -n demo ms mssqlserver-ag-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
mssqlserver.kubedb.com/mssqlserver-ag-cluster patched

$ kubectl delete ms -n demo mssqlserver-ag-cluster
mssqlserver.kubedb.com "mssqlserver-ag-cluster" deleted
```

Now, run the following command to get all mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get ms,petset,pod,svc,secret,pvc -n demo 
NAME                       TYPE                       DATA   AGE
secret/mssqlserver-ca      kubernetes.io/tls          2      3h8m
```

From the above output, you can see that all mssqlserver resources are deleted. there is no option to recreate/reinitialize your database if `deletionPolicy` is set to `WipeOut`.

>Be careful when you set the `deletionPolicy` to `WipeOut`. Because there is no option to trace the database resources if once deleted the database.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mssqlserver/mssqlserver-ag-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete mssqlserver -n demo mssqlserver-ag-cluster
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular MSSQLServer crd when you delete the crd. For more details about deletion policy, please visit [here](/docs/guides/mssqlserver/concepts/mssqlserver.md#specdeletionpolicy).

## Next Steps

- Learn about [backup and restore](/docs/guides/mssqlserver/backup/overview/index.md) SQL Server using KubeStash.
- Want to set up SQL Server Availability Group clusters? Check how to [Configure SQL Server Availability Group Cluster](/docs/guides/mssqlserver/clustering/ag_cluster.md)
- Detail concepts of [MSSQLServer Object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).