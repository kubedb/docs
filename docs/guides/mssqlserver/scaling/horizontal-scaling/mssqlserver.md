---
title: Horizontal Scaling MSSQLServer Cluster
menu:
  docs_{{ .version }}:
    identifier: ms-scaling-horizontal-guide
    name: Scale Horizontally
    parent: ms-scaling-horizontal
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale MSSQLServer Cluster

This guide will show you how to use `KubeDB` Ops Manager to increase/decrease the number of replicas of a `MSSQLServer` Cluster.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/mssqlserver/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mssqlserver/scaling/horizontal-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/scaling/horizontal-scaling) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

### Apply Horizontal Scaling on MSSQLServer Cluster

Here, we are going to deploy a `MSSQLServer` Cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

#### Prepare Cluster

At first, we are going to deploy a Cluster server with 2 replicas. Then, we are going to add two additional replicas through horizontal scaling. Finally, we will remove 1 replica from the cluster again via horizontal scaling.

**Find supported MSSQLServer Version:**

When you have installed `KubeDB`, it has created `MSSQLServerVersion` CR for all supported `MSSQLServer` versions. Let's check the supported MSSQLServer versions,

```bash
$ kubectl get mssqlserverversion
NAME        VERSION   DB_IMAGE                                                DEPRECATED   AGE
2022-cu12   2022      mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04                176m
2022-cu14   2022      mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04                176m
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MSSQLServer`. You can use any non-deprecated version. Here, we are going to create a MSSQLServer Cluster using `MSSQLServer` `2022-cu12`.

**Deploy MSSQLServer Cluster:**


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

In this section, we are going to deploy a MSSQLServer Cluster with 2 replicas. Then, in the next section we will scale up the cluster using horizontal scaling. Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-ag-cluster
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 2
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
          resources:
            requests:
              cpu: "500m"
              memory: "1.5Gi"
            limits:
              cpu: 1
              memory: "2Gi"
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

Let's create the `MSSQLServer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/scaling/horizontal-scaling/mssql-ag-cluster.yaml
mssqlserver.kubedb.com/mssql-ag-cluster created
```

**Wait for the cluster to be ready:**

`KubeDB` operator watches for `MSSQLServer` objects using Kubernetes API. When a `MSSQLServer` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc. A secret called `mssql-ag-cluster-auth` (format: <em>{mssqlserver-object-name}-auth</em>) will be created storing the password for mssqlserver superuser.
Now, watch `MSSQLServer` is going to `Running` state and also watch `PetSet` and its pod is created and going to `Running` state,

```bash
$ watch kubectl get ms,petset,pods -n demo
Every 2.0s: kubectl get ms,petset,pods -n demo   

NAME                                      VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-ag-cluster   2022-cu12   Ready    2m52s

NAME                                            AGE
petset.apps.k8s.appscode.com/mssql-ag-cluster   2m11s

NAME                     READY   STATUS    RESTARTS   AGE
pod/mssql-ag-cluster-0   2/2     Running   0          2m11s
pod/mssql-ag-cluster-1   2/2     Running   0          2m6s

```

Let's verify that the PetSet's pods have created the availability group cluster successfully,

```bash
$ kubectl get secrets -n demo mssql-ag-cluster-auth -o jsonpath='{.data.\username}' | base64 -d
sa
$ kubectl get secrets -n demo mssql-ag-cluster-auth -o jsonpath='{.data.\password}' | base64 -d
123KKxgOXuOkP206
```

Now, connect to the database using username and password, check the name of the created availability group, replicas of the availability group and see if databases are added to the availability group.
```bash
$ kubectl exec -it -n demo mssql-ag-cluster-0 -c mssql -- bash
mssql@mssql-ag-cluster-2:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "123KKxgOXuOkP206"
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
mssqlagcluster                                                                                                            

(1 rows affected)
1> select replica_server_name from sys.availability_replicas;
2> go
replica_server_name                                                                                                                                                                                                                                             
-------------------------------------------------------------------------------------------
mssql-ag-cluster-0                                                                                                                                                                                                                                              
mssql-ag-cluster-1                                                                                                                    
(3 rows affected)
1> select database_name	from sys.availability_databases_cluster;
2> go
database_name                                                                                                                   
------------------------------------------------------------------------------------------
agdb1                                                                                                                           
agdb2                                                                                                                           

(2 rows affected)

```


So, we can see that our cluster has 2 replicas. Now, we are ready to apply the horizontal scale to this MSSQLServer cluster.

#### Scale Up

Here, we are going to add 1 replica in our Cluster using horizontal scaling.

**Create MSSQLServerOpsRequest:**

To scale up your cluster, you have to create a `MSSQLServerOpsRequest` CR with your desired number of replicas after scaling. Below is the YAML of the `MSSQLServerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: ms-scale-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mssql-ag-cluster
  horizontalScaling:
    replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mssql-ag-cluster`.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the expected number of replicas after the scaling.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/scaling/horizontal-scaling/msops-hscale-up.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-hscale-up created
```

**Verify Scale-Up Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale up the PetSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `MSSQLServer` object.

First, we will wait for `MSSQLServerOpsRequest` to be successful. Run the following command to watch `MSSQLServerOpsRequest` cr,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo msops-hscale-up
Every 2.0s: kubectl get mssqlserveropsrequest -n demo msops-hscale-up                 

NAME              TYPE                STATUS       AGE
msops-hscale-up   HorizontalScaling   Successful   76s

```

You can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest`, we will see that the `MSSQLServer` cluster is scaled up.

```bash
kubectl describe mssqlserveropsrequest -n demo msops-hscale-up
Name:         msops-hscale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-10-24T15:09:36Z
  Generation:          1
  Resource Version:    752963
  UID:                 43193e49-8461-4e14-b1c1-7aaa33d0251a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mssql-ag-cluster
  Horizontal Scaling:
    Replicas:  3
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-10-24T15:09:36Z
    Message:               MSSQLServer ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-10-24T15:09:39Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-24T15:10:29Z
    Message:               Successfully Scaled Up Node
    Observed Generation:   1
    Reason:                HorizontalScaleUp
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2024-10-24T15:09:44Z
    Message:               get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetCurrentLeader--mssql-ag-cluster-0
    Last Transition Time:  2024-10-24T15:09:44Z
    Message:               get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetRaftNode--mssql-ag-cluster-0
    Last Transition Time:  2024-10-24T15:09:44Z
    Message:               add raft node; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  AddRaftNode--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T15:09:49Z
    Message:               patch petset; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T15:09:49Z
    Message:               mssql-ag-cluster already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-10-24T15:09:59Z
    Message:               is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T15:10:19Z
    Message:               is mssql running; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsMssqlRunning
    Last Transition Time:  2024-10-24T15:10:24Z
    Message:               ensure replica join; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  EnsureReplicaJoin
    Last Transition Time:  2024-10-24T15:10:34Z
    Message:               successfully reconciled the MSSQLServer with modified replicas
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-24T15:10:35Z
    Message:               Successfully updated MSSQLServer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-10-24T15:10:35Z
    Message:               Successfully completed the HorizontalScaling for MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age    From                         Message
  ----     ------                                                                ----   ----                         -------
  Normal   Starting                                                              2m22s  KubeDB Ops-manager Operator  Start processing for MSSQLServerOpsRequest: demo/msops-hscale-up
  Normal   Starting                                                              2m22s  KubeDB Ops-manager Operator  Pausing MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                            2m22s  KubeDB Ops-manager Operator  Successfully paused MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: msops-hscale-up
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  2m14s  KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0       2m14s  KubeDB Ops-manager Operator  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  add raft node; ConditionStatus:True; PodName:mssql-ag-cluster-2       2m14s  KubeDB Ops-manager Operator  add raft node; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  2m9s   KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0       2m9s   KubeDB Ops-manager Operator  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  patch petset; ConditionStatus:True; PodName:mssql-ag-cluster-2        2m9s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  2m4s   KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  is pod ready; ConditionStatus:False; PodName:mssql-ag-cluster-2       2m4s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False; PodName:mssql-ag-cluster-2
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  119s   KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2        119s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  is mssql running; ConditionStatus:False                               109s   KubeDB Ops-manager Operator  is mssql running; ConditionStatus:False
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  109s   KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2        109s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  99s    KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2        99s    KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  is mssql running; ConditionStatus:True                                99s    KubeDB Ops-manager Operator  is mssql running; ConditionStatus:True
  Warning  ensure replica join; ConditionStatus:False                            98s    KubeDB Ops-manager Operator  ensure replica join; ConditionStatus:False
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  94s    KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2        94s    KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  is mssql running; ConditionStatus:True                                94s    KubeDB Ops-manager Operator  is mssql running; ConditionStatus:True
  Warning  ensure replica join; ConditionStatus:True                             94s    KubeDB Ops-manager Operator  ensure replica join; ConditionStatus:True
  Warning  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0  89s    KubeDB Ops-manager Operator  get current leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Normal   HorizontalScaleUp                                                     89s    KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdatePetSets                                                         84s    KubeDB Ops-manager Operator  successfully reconciled the MSSQLServer with modified replicas
  Normal   UpdateDatabase                                                        83s    KubeDB Ops-manager Operator  Successfully updated MSSQLServer
  Normal   Starting                                                              83s    KubeDB Ops-manager Operator  Resuming MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                            83s    KubeDB Ops-manager Operator  Successfully resumed MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: msops-hscale-up
  Normal   UpdateDatabase                                                        83s    KubeDB Ops-manager Operator  Successfully updated MSSQLServer
```

Now, we are going to verify whether the number of replicas has increased to meet up the desired state. So let's check the new pods coordinator container's logs to see if this is joined in the cluster as new replica.

```bash
$ kubectl logs -f -n demo mssql-ag-cluster-2 -c mssql-coordinator
raft2024/10/24 15:09:55 INFO: 3 switched to configuration voters=(1 2 3)
raft2024/10/24 15:09:55 INFO: 3 switched to configuration voters=(1 2 3)
raft2024/10/24 15:09:55 INFO: 3 switched to configuration voters=(1 2 3)
raft2024/10/24 15:09:55 INFO: 3 [term: 1] received a MsgHeartbeat message with higher term from 1 [term: 3]
raft2024/10/24 15:09:55 INFO: 3 became follower at term 3
raft2024/10/24 15:09:55 INFO: raft.node: 3 elected leader 1 at term 3
I1024 15:09:56.855261       1 mssql.go:94] new elected primary is :mssql-ag-cluster-0.
I1024 15:09:56.864197       1 mssql.go:120] New primary is ready to accept connections...
I1024 15:09:56.864213       1 mssql.go:171] lastLeaderId : 0,      currentLeaderId : 1
I1024 15:09:56.864230       1 on_leader_change.go:47] New Leader elected.
I1024 15:09:56.864237       1 on_leader_change.go:82] This pod is now a secondary according to raft
I1024 15:09:56.864243       1 on_leader_change.go:100] instance demo/mssql-ag-cluster-2 running according to the role
I1024 15:09:56.864317       1 utils.go:219] /scripts/run_signal.txt file created successfully
E1024 15:09:56.935767       1 exec_utils.go:65] Error while trying to get process output from the pod. Error: could not execute: command terminated with exit code 1
I1024 15:09:56.935794       1 on_leader_change.go:110] mssql is not ready yet
I1024 15:10:07.980792       1 on_leader_change.go:110] mssql is not ready yet
I1024 15:10:18.049036       1 on_leader_change.go:110] mssql is not ready yet
I1024 15:10:18.116939       1 on_leader_change.go:118] mssql is ready now
I1024 15:10:18.127315       1 ag_status.go:43] No Availability Group found
I1024 15:10:18.127336       1 ag.go:79] Joining  Availability Group... 
I1024 15:10:24.638144       1 on_leader_change.go:94] Successfully patched label of demo/mssql-ag-cluster-2 to secondary
I1024 15:10:24.650611       1 health.go:50] Sequence Number updated. new sequenceNumber = 4294967322, previous sequenceNumber = 0
I1024 15:10:24.650632       1 health.go:51] 1:1A (4294967322)
```


Now, connect to the database, check updated configurations of the availability group cluster. 
```bash
$ kubectl exec -it -n demo mssql-ag-cluster-2 -c mssql -- bash
mssql@mssql-ag-cluster-2:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "123KKxgOXuOkP206"
1> SELECT name FROM sys.availability_groups
2> go
name                                                                                                                            
----------------------------------------------------------------------------
mssqlagcluster                                                                                                            

(1 rows affected)
1> select replica_server_name from sys.availability_replicas;
2> go
replica_server_name                                                                                                                                                                                                                                             
-------------------------------------------------------------------------------------------
mssql-ag-cluster-0                                                                                                                                                                                                                                              
mssql-ag-cluster-1 

mssql-ag-cluster-2
                                                                                                                   
(3 rows affected)
1> select database_name	from sys.availability_databases_cluster;
2> go
database_name                                                                                                                   
------------------------------------------------------------------------------------------
agdb1                                                                                                                           
agdb2                                                                                                                           

(2 rows affected)
```

#### Scale Down

Here, we are going to remove 1 replica from our cluster using horizontal scaling.

**Create MSSQLServerOpsRequest:**

To scale down your cluster, you have to create a `MSSQLServerOpsRequest` CR with your desired number of replicas after scaling. Below is the YAML of the `MSSQLServerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mssql-ag-cluster
  horizontalScaling:
    replicas: 2
```

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/scaling/horizontal-scaling/msops-hscale-down.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-hscale-down created
```

**Verify Scale-down Succeeded:**

If everything goes well, `KubeDB` Ops Manager will scale down the PetSet's `Pod`. After the scaling process is completed successfully, the `KubeDB` Ops Manager updates the replicas of the `MSSQLServer` object.

Now, we will wait for `MSSQLServerOpsRequest` to be successful. Run the following command to watch `MSSQLServerOpsRequest` cr,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo msops-hscale-down
Every 2.0s: kubectl get mssqlserveropsrequest -n demo msops-hscale-down

NAME                TYPE                STATUS       AGE
msops-hscale-down   HorizontalScaling   Successful   98s
```

You can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest`, we shall see that the `MSSQLServer` cluster is scaled down.

```bash
$ kubectl describe  mssqlserveropsrequest -n demo msops-hscale-down
Name:         msops-hscale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-10-24T15:22:54Z
  Generation:          1
  Resource Version:    754237
  UID:                 c5dc6971-5f60-4736-992a-8fdf5a2911d9
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mssql-ag-cluster
  Horizontal Scaling:
    Replicas:  2
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-10-24T15:22:54Z
    Message:               MSSQLServer ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-10-24T15:23:06Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-24T15:24:06Z
    Message:               Successfully Scaled Down Node
    Observed Generation:   1
    Reason:                HorizontalScaleDown
    Status:                True
    Type:                  HorizontalScaleDown
    Last Transition Time:  2024-10-24T15:23:21Z
    Message:               get current raft leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetCurrentRaftLeader--mssql-ag-cluster-0
    Last Transition Time:  2024-10-24T15:23:11Z
    Message:               get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetRaftNode--mssql-ag-cluster-0
    Last Transition Time:  2024-10-24T15:23:11Z
    Message:               remove raft node; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  RemoveRaftNode--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T15:23:21Z
    Message:               patch petset; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T15:23:21Z
    Message:               mssql-ag-cluster already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-10-24T15:23:26Z
    Message:               get pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  GetPod
    Last Transition Time:  2024-10-24T15:23:56Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T15:23:56Z
    Message:               delete pvc; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T15:24:01Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-24T15:24:01Z
    Message:               ag node remove; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  AgNodeRemove
    Last Transition Time:  2024-10-24T15:24:11Z
    Message:               successfully reconciled the MSSQLServer with modified replicas
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-24T15:24:11Z
    Message:               Successfully updated MSSQLServer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-10-24T15:24:11Z
    Message:               Successfully completed the HorizontalScaling for MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                      Age   From                         Message
  ----     ------                                                                      ----  ----                         -------
  Normal   Starting                                                                    2m1s  KubeDB Ops-manager Operator  Start processing for MSSQLServerOpsRequest: demo/msops-hscale-down
  Normal   Starting                                                                    2m1s  KubeDB Ops-manager Operator  Pausing MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                                  2m1s  KubeDB Ops-manager Operator  Successfully paused MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: msops-hscale-down
  Warning  get current raft leader; ConditionStatus:True; PodName:mssql-ag-cluster-0   104s  KubeDB Ops-manager Operator  get current raft leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0             104s  KubeDB Ops-manager Operator  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  remove raft node; ConditionStatus:True; PodName:mssql-ag-cluster-2          104s  KubeDB Ops-manager Operator  remove raft node; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  get current raft leader; ConditionStatus:True; PodName:mssql-ag-cluster-0   94s   KubeDB Ops-manager Operator  get current raft leader; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0             94s   KubeDB Ops-manager Operator  get raft node; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  patch petset; ConditionStatus:True; PodName:mssql-ag-cluster-2              94s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2                   59s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  delete pvc; ConditionStatus:True; PodName:mssql-ag-cluster-2                59s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  get pvc; ConditionStatus:False                                              59s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2                   54s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  delete pvc; ConditionStatus:True; PodName:mssql-ag-cluster-2                54s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  get pvc; ConditionStatus:True                                               54s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  ag node remove; ConditionStatus:True                                        54s   KubeDB Ops-manager Operator  ag node remove; ConditionStatus:True
  Normal   HorizontalScaleDown                                                         49s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdatePetSets                                                               44s   KubeDB Ops-manager Operator  successfully reconciled the MSSQLServer with modified replicas
  Normal   UpdateDatabase                                                              44s   KubeDB Ops-manager Operator  Successfully updated MSSQLServer
  Normal   Starting                                                                    44s   KubeDB Ops-manager Operator  Resuming MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                                  44s   KubeDB Ops-manager Operator  Successfully resumed MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: msops-hscale-down
  Normal   UpdateDatabase                                                              44s   KubeDB Ops-manager Operator  Successfully updated MSSQLServer
```

Now, we are going to verify whether the number of replicas has decreased to meet up the desired state, Let's check, the mssqlserver status if it's ready then the scale-down is successful.

```bash
$ kubectl get ms,petset,pods -n demo
NAME                                      VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-ag-cluster   2022-cu12   Ready    39m

NAME                                            AGE
petset.apps.k8s.appscode.com/mssql-ag-cluster   38m

NAME                     READY   STATUS    RESTARTS   AGE
pod/mssql-ag-cluster-0   2/2     Running   0          38m
pod/mssql-ag-cluster-1   2/2     Running   0          38m
```


Now, connect to the database, check updated configurations of the availability group cluster.
```bash
$ kubectl exec -it -n demo mssql-ag-cluster-0 -c mssql -- bash
mssql@mssql-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "123KKxgOXuOkP206"
1> SELECT name FROM sys.availability_groups
2> go
name                                                                                                                            
----------------------------------------------------
mssqlagcluster                                                                                                                  

(1 rows affected)
1> select replica_server_name from sys.availability_replicas;
2> go
replica_server_name                                                                                                                                                                                                                                             
--------------------------------------
mssql-ag-cluster-0                                                                                                                                                                                                                                              
mssql-ag-cluster-1                                                                                                                                                                                                                                              

(2 rows affected)
```

You can see above that our `MSSQLServer` cluster now has a total of 2 replicas. It verifies that we have successfully scaled down.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ms -n demo mssql-ag-cluster
kubectl delete mssqlserveropsrequest -n demo msops-hscale-up
kubectl delete mssqlserveropsrequest -n demo msops-hscale-down
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```
