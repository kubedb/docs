---
title: SQL Server Distributed Availability Group Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: ms-clustering-distrubuted-availability-group
    name: Distributed Availability Group Cluster
    parent: ms-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Microsoft SQL Server Distributed Availability Group (DAG) Cluster

This tutorial will show you how to use KubeDB to run a Microsoft SQL Server Distributed Availability Group (DAG) Cluster, which is ideal for disaster recovery scenarios across two different sites or Kubernetes clusters.

## Before You Begin

- You will need two separate Kubernetes clusters, two distinct environments that can communicate over the network. The kubectl command-line tool must be configured to communicate with your clusters.
- Each cluster must have KubeDB installed. Follow the steps [here](/docs/setup/README.md), ensuring you enable the MSSQLServer feature gate: `--set global.featureGates.MSSQLServer=true`.
- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. - Each cluster must have `cert-manager` installed. Follow the steps [here](https://cert-manager.io/docs/installation/kubernetes/).
- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in both clusters.
  ```bash
  $ kubectl get storageclasses
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  4h48m
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  # In Cluster 1
  kubectl create ns demo
  # In Cluster 2
  kubectl create ns demo
  ```

> Note: The YAML files used in this tutorial are stored in [docs/examples/mssqlserver/dag-cluster/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/dag-cluster/) folder in the [kubedb/docs](https://github.com/kubedb/docs) repository.


## Find Available Microsoft SQL Server Versions

When you have installed KubeDB, it has created `MSSQLServerVersion` CR for all supported Microsoft SQL Server versions. Check it by using the `kubectl get mssqlserverversions`. You can also use `msversion` shorthand instead of `mssqlserverversions`.

```bash
$ kubectl get msversion
NAME        VERSION   DB_IMAGE                                                DEPRECATED   AGE
2022-cu12   2022      mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04                161m
2022-cu14   2022      mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04                161m
2022-cu16   2022      mcr.microsoft.com/mssql/server:2022-CU16-ubuntu-22.04                161m
2022-cu19   2022      mcr.microsoft.com/mssql/server:2022-CU19-ubuntu-22.04                161m
```


> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver/dag-cluster/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/dag-cluster/) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).



## Deploy Microsoft SQL Server Distributed Availability Group Cluster 

The process involves deploying a primary Availability Group (AG) in the first cluster, exporting its critical credentials, and then deploying a secondary AG in the second cluster that uses those credentials to join the DAG.



### Create Issuer/ClusterIssuer on Both Clusters

First, create an `Issuer` in your primary cluster's `demo` namespace. This will be used to generate the necessary certificates for endpoint authentication, and even if TLS is not enabled for SQL Server. The issuer will be used to configure the TLS-enabled Wal-G proxy server, which is required for the SQL Server backup and restore operations.

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/dag-cluster/mssqlserver-ca-issuer.yaml
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





### Deploy Microsoft SQL Server Distributed AG's primary cluster (AG1) 
KubeDB implements a `MSSQLServer` CRD to define the specification of a Microsoft SQL Server database. Below is the `MSSQLServer` object created in this tutorial.
Here, our issuer `mssqlserver-ca-issuer` is ready to deploy a `MSSQLServer`.

Now, deploy the first `MSSQLServer` resource. This will act as the **primary** site of our Distributed AG. Note the `topology.mode` and the `distributedAG` block.

```yaml
# ag1.yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ag1
  namespace: demo
spec:
  version: "2022-cu16"
  replicas: 3
  topology:
    mode: DistributedAG
    availabilityGroup:
      databases:
        - agdb
      secondaryAccessMode: "All"
    distributedAG:
      self:
        role: Primary
        url: "10.2.0.236" # Replace with the reachable LoadBalancer IP/hostname of this AG
      remote:
        name: ag2
        url: "10.2.0.181" # Replace with the reachable LoadBalancer IP/hostname of the secondary AG
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
              value: Evaluation
  serviceTemplates:
    - alias: primary
      spec:
        type: LoadBalancer # Exposes the primary replica via a LoadBalancer
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

> **Note:** You must replace the `url` fields with the actual, mutually reachable IP addresses or hostnames of your `LoadBalancer` services. You may need to create placeholder services first to get these IPs.


### Create placeholder services first to get these IPs
```yaml
apiVersion: v1
kind: Service
metadata:
  name: ag1
  namespace: demo
spec:
  ports:
    - name: primary
      port: 1433
      protocol: TCP
      targetPort: db
    - name: mirror
      port: 5022
      protocol: TCP
      targetPort: mirror
  selector:
    app.kubernetes.io/instance: ag1
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mssqlservers.kubedb.com
    kubedb.com/role: primary
  type: LoadBalancer
```

Deploy `ag1` primary service to your first cluster:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/dag-cluster/ag1-primary-svc.yaml
service/ag1 created created

$ kubectl get svc -n demo ag1
NAME   TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)                         AGE
ag1    LoadBalancer   10.43.117.2   10.2.0.236    1433:31485/TCP,5022:32511/TCP   122m
```

Deploy `ag1` to your first cluster:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/dag-cluster/ag1.yaml
mssqlserver.kubedb.com/ag1 created
```



Here,

- `spec.version` is the name of the MSSQLServerVersion CR where the docker images are specified. In this tutorial, a MSSQLServer `2022-cu16` database is going to be created.
- `spec.replicas` denotes the number of replicas of the local availability group
- `spec.topology` specifies the mode `DistributedAG` and the list of names of the databases that we want in our availability group. 
   KubeDB operator will create and add these databases to the created availability group automatically. Users don't have to create, configure or add the database to the availability group manually. Users can update this list later as well. 
- `spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). Here `tls.clientTLS: false` means tls will not be enabled for SQL Server but the Issuer will be used to configure tls enabled wal-g proxy-server which is required for SQL Server backup operation.
- `spec.storageType` specifies the type of storage that will be used for MSSQLServer database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MSSQLServer database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MSSQLServer` CR or which resources KubeDB should keep or delete when you delete `MSSQLServer` CR. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy)

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in storage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MSSQLServer` objects using Kubernetes api. When a `MSSQLServer` object is created, KubeDB operator will create a new PetSet and a Service with the matching MSSQLServer object name. KubeDB operator will also create a governing service for PetSets with the name `<MSSQLServerName>-pods`, if one is not already present.


Wait for `ag1` to become `Ready`:
```bash
# In Cluster 1
kubectl get mssqlserver -n demo -w

# Let's see all the SQL server resources that are created. 
```bash
$ kubectl get ms,petset,pod,svc,secret,issuer,pvc -n demo
NAME                         VERSION     STATUS   AGE
mssqlserver.kubedb.com/ag1   2022-cu16   Ready    6m10s

NAME                               AGE
petset.apps.k8s.appscode.com/ag1   5m6s

NAME        READY   STATUS    RESTARTS   AGE
pod/ag1-0   2/2     Running   0          5m5s
pod/ag1-1   2/2     Running   0          3m41s
pod/ag1-2   2/2     Running   0          3m20s

NAME                    TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
service/ag1             LoadBalancer   10.43.117.2     10.2.0.236    1433:31485/TCP,5022:32511/TCP   7m57s
service/ag1-pods        ClusterIP      None            <none>        1433/TCP,5022/TCP               6m10s
service/ag1-secondary   ClusterIP      10.43.169.148   <none>        1433/TCP                        6m10s

NAME                       TYPE                       DATA   AGE
secret/ag1-auth            kubernetes.io/basic-auth   2      6m10s
secret/ag1-client-cert     kubernetes.io/tls          3      6m10s
secret/ag1-config          Opaque                     1      6m8s
secret/ag1-dbm-login       kubernetes.io/basic-auth   1      6m8s
secret/ag1-endpoint-cert   kubernetes.io/tls          3      6m10s
secret/ag1-master-key      kubernetes.io/basic-auth   1      6m8s
secret/ag1-server-cert     kubernetes.io/tls          3      6m8s
secret/mssqlserver-ca      kubernetes.io/tls          2      8m50s

NAME                                           READY   AGE
issuer.cert-manager.io/mssqlserver-ca-issuer   True    8m39s

NAME                               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-ag1-0   Bound    pvc-2ef5b468-faf1-4bfe-af04-1a9c674a38b7   1Gi        RWO            local-path     <unset>                 5m5s
persistentvolumeclaim/data-ag1-1   Bound    pvc-8bdba861-5464-4e8b-a070-6ef21fc737bc   1Gi        RWO            local-path     <unset>                 3m41s
persistentvolumeclaim/data-ag1-2   Bound    pvc-a8b5c4b7-684e-42c8-8d72-18271801d06a   1Gi        RWO            local-path     <unset>                 3m20s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created and is able to accept client connections. Run the following command to see the modified MSSQLServer object:

```bash
$ kubectl get ms -n demo ag1 -o yaml
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  creationTimestamp: "2025-07-04T07:01:39Z"
  finalizers:
    - kubedb.com
  generation: 3
  name: ag1
  namespace: demo
  resourceVersion: "79870"
  uid: d371db52-17c6-49b3-8318-67bc1dae344c
spec:
  authSecret:
    activeFrom: "2025-07-04T07:01:39Z"
    name: ag1-auth
  autoOps: {}
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
        - env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation
          name: mssql
          resources:
            limits:
              memory: 2Gi
            requests:
              cpu: "1"
              memory: 1536Mi
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
          resources: {}
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
      podPlacementPolicy:
        name: default
      securityContext:
        fsGroup: 10001
      serviceAccountName: ag1
  replicas: 3
  serviceTemplates:
    - alias: primary
      metadata: {}
      spec:
        type: LoadBalancer
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  storageType: Durable
  tls:
    certificates:
      - alias: server
        secretName: ag1-server-cert
        subject:
          organizationalUnits:
            - server
          organizations:
            - kubedb
      - alias: client
        secretName: ag1-client-cert
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
        - agdb
      leaderElection:
        electionTick: 10
        heartbeatTick: 1
        period: 300ms
        transferLeadershipInterval: 1s
        transferLeadershipTimeout: 1m0s
      secondaryAccessMode: All
    distributedAG:
      remote:
        name: ag2
        url: 10.2.0.181
      self:
        role: Primary
        url: 10.2.0.236
    mode: DistributedAG
  version: 2022-cu16
status:
  conditions:
    - lastTransitionTime: "2025-07-04T07:01:39Z"
      message: 'The KubeDB operator has started the provisioning of MSSQLServer: demo/ag1'
      observedGeneration: 1
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2025-07-04T07:04:50Z"
      message: All replicas are ready for MSSQLServer demo/ag1
      observedGeneration: 3
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2025-07-04T07:04:43Z"
      message: Primary replica is ready for MSSQL demo/ag1
      observedGeneration: 3
      reason: AvailabilyGroupCreatedInPrimary
      status: "True"
      type: AGPrimaryReplicaReady
    - lastTransitionTime: "2025-07-04T07:05:00Z"
      message: database demo/ag1 is accepting connection
      observedGeneration: 3
      reason: AcceptingConnection
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2025-07-04T07:05:00Z"
      message: database demo/ag1 is ready
      observedGeneration: 3
      reason: AllReplicasReady
      status: "True"
      type: Ready
    - lastTransitionTime: "2025-07-04T07:05:52Z"
      message: 'The MSSQLServer: demo/ag1 is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready
```


## Configure Shared Credentials

A DAG requires that both participating AGs share identical endpoint credentials. KubeDB simplifies this with a CLI command.

**Generate Configuration from Primary:**

Run the `kubectl-dba` command against your primary `MSSQLServer` instance (`ag1`). This extracts the required secrets into a YAML file.
```bash
# In Cluster 1 context
$ kubectl-dba mssql dag-config ag1 -n demo
```
Generating DAG configuration for MSSQLServer 'ag1' in namespace 'demo'...
- Fetching secret 'ag1-dbm-login'...
- Fetching secret 'ag1-master-key'...
- Fetching secret 'ag1-endpoint-cert'...
- Fetching AppBinding 'ag1'...
  Successfully generated DAG configuration.
  Apply this file in your remote cluster: kubectl apply -f ./ag1-dag-config.yaml



**Apply Configuration to Secondary Cluster:**
Apply the generated manifest to your second cluster. This creates the identical secrets needed for authentication.
```bash
# In Cluster 2 context
kubectl apply -f ./ag1-dag-config.yaml
```
secret/ag1-dbm-login created
secret/ag1-master-key created
secret/ag1-endpoint-cert created
appbinding.appcatalog.appscode.com/ag1 created




## Deploy the Secondary Availability Group (ag2)

Now, deploy the second `MSSQLServer` resource in your second cluster. This will act as the **secondary** site of our DAG.
Notice that `spec.topology.availabilityGroup.databases` is empty, and we now reference the secrets we just created.


```yaml
# ag2.yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ag2
  namespace: demo
spec:
  version: "2022-cu16"
  replicas: 3
  topology:
    mode: DistributedAG
    availabilityGroup:
      # Databases field must be empty for the secondary AG.
      secondaryAccessMode: "All"
      # Reference the secrets you copied from the primary cluster.
      loginSecretName: ag1-dbm-login
      masterKeySecretName: ag1-master-key
      endpointCertSecretName: ag1-endpoint-cert
    distributedAG:
      self:
        role: Secondary
        url: "10.2.0.181" # Replace with the reachable LoadBalancer IP/hostname of this AG
      remote:
        name: ag1
        url: "10.2.0.236" # Replace with the reachable LoadBalancer IP/hostname of the primary AG
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer # An issuer with this name must also exist in the secondary cluster
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
              value: Evaluation
  serviceTemplates:
  - alias: primary
    spec:
      type: LoadBalancer
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```


> **Note:** You must replace the `url` fields with the actual, mutually reachable IP addresses or hostnames of your `LoadBalancer` services. You may need to create placeholder services first to get these IPs.


### Create placeholder services first to get these IPs
```yaml
apiVersion: v1
kind: Service
metadata:
  name: ag2
  namespace: demo
spec:
  ports:
    - name: primary
      port: 1433
      protocol: TCP
      targetPort: db
    - name: mirror
      port: 5022
      protocol: TCP
      targetPort: mirror
  selector:
    app.kubernetes.io/instance: ag2
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mssqlservers.kubedb.com
    kubedb.com/role: primary
  type: LoadBalancer
```

Deploy `ag2` primary service to your first cluster:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/dag-cluster/ag1-primary-svc.yaml
service/ag1 created created

$ kubectl get svc -n demo ag2 
NAME   TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
ag2    LoadBalancer   10.43.169.101   10.2.0.181    1433:31686/TCP,5022:32633/TCP   3m10s
```



Deploy `ag2` to your second cluster:

```bash
# In Cluster 2
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/dag-cluster/ag2.yaml
mssqlserver.kubedb.com/ag2 created
```



Once both `ag1` and `ag2` are `Ready`, KubeDB has established the Distributed AG.


## Verify Data Replication

**Insert data into the primary site (`ag1`):**


Let's insert some data into the primary database of SQL server distributed availability group and see if data replication is working fine. First, we have to determine the primary replica, as data writes are only permitted on the primary node.


```bash
$ kubectl get pods -n demo --selector=app.kubernetes.io/instance=ag1 -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\n"}{end}'
ag1-0	primary
ag1-1	secondary
ag1-2	secondary
```

From the output above, we can see that ag1-0 is the primary node. Which is actually the 'GLOBAL PRIMARY' of our 'Distributed Availability Group.'


```bash
# In Cluster 1
kubectl exec -it ag1-0 -c mssql -n demo -- bash
# See AG, DAG, and database status 
mssql@ag1-0:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
select database_name from sys.availability_databases_cluster;
SELECT name FROM sys.availability_groups;
SELECT replica_server_name FROM sys.availability_replicas;
SELECT is_local, role_desc, replica_id, group_id, synchronization_health_desc, connected_state_desc, operational_state_desc from sys.dm_hadr_availability_replica_states
"


database_name                                                                                                                   
--------------------------------------------------------------------------------------------------------------------------------
agdb                                                                                                                            

(1 rows affected)
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
ag1                                                                                                                             
dag                                                                                                                             

(2 rows affected)
replica_server_name                                                                                                                                                                                                                                             
----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
ag1-0                                                                                                                                                                                                                                                           
ag1-1                                                                                                                                                                                                                                                           
ag1-2                                                                                                                                                                                                                                                           
ag1                                                                                                                                                                                                                                                             
ag2                                                                                                                                                                                                                                                             

(5 rows affected)
is_local role_desc                                                    replica_id                           group_id                             synchronization_health_desc                                  connected_state_desc                                         operational_state_desc                                      
-------- ------------------------------------------------------------ ------------------------------------ ------------------------------------ ------------------------------------------------------------ ------------------------------------------------------------ ------------------------------------------------------------
       1 PRIMARY                                                      AB19A923-FDFB-436D-9B16-4556961CF015 BE9BE8C9-6E17-1132-BFBA-8B7D2C28AFDB HEALTHY                                                      CONNECTED                                                    ONLINE                                                      
       0 SECONDARY                                                    DD8A151D-E851-4D62-8E04-DB4224B2A5A7 BE9BE8C9-6E17-1132-BFBA-8B7D2C28AFDB HEALTHY                                                      CONNECTED                                                    NULL                                                        
       0 SECONDARY                                                    1A5379EF-4058-4F34-A091-D2F18AD05FAB BE9BE8C9-6E17-1132-BFBA-8B7D2C28AFDB HEALTHY                                                      CONNECTED                                                    NULL                                                        
       1 PRIMARY                                                      6CD38135-9FFF-24A2-9401-E9833DBDC2D1 6BC05A51-AA36-A196-09BD-481D7A0973C0 HEALTHY                                                      CONNECTED                                                    ONLINE                                                      
       0 SECONDARY                                                    0EAC444F-1CF1-8D21-0178-B43D2842ACF5 6BC05A51-AA36-A196-09BD-481D7A0973C0 HEALTHY                                                      CONNECTED                                                    NULL                                                        

(5 rows affected)


mssql@ag1-0:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
use agdb;
CREATE TABLE test_table (id INT, name NVARCHAR(50));
INSERT INTO test_table VALUES (1, 'DAG Setup'); 
SELECT * FROM test_table;
"
Changed database context to 'agdb'.

(1 rows affected)
id          name                                              
----------- --------------------------------------------------
          1 DAG Setup                                         

(1 rows affected)
 ```


**Verify the data exists on the secondary site (`ag2`):**


```bash
# In Cluster 2
kubectl exec -it ag2-0 -n demo -c mssql -- bash 
# See AG, DAG, and database status 
mssql@ag2-0:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
select database_name from sys.availability_databases_cluster;
SELECT name FROM sys.availability_groups;
SELECT replica_server_name FROM sys.availability_replicas;
SELECT is_local, role_desc, replica_id, group_id, synchronization_health_desc, connected_state_desc, operational_state_desc from sys.dm_hadr_availability_replica_states
"
database_name                                                                                                                   
--------------------------------------------------------------------------------------------------------------------------------
agdb                                                                                                                            

(1 rows affected)
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
ag2                                                                                                                             
dag                                                                                                                             

(2 rows affected)
replica_server_name                                                                                                                                                                                                                                             
----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
ag2-0                                                                                                                                                                                                                                                           
ag2-1                                                                                                                                                                                                                                                           
ag2-2                                                                                                                                                                                                                                                           
ag1                                                                                                                                                                                                                                                             
ag2                                                                                                                                                                                                                                                             

(5 rows affected)
is_local role_desc                                                    replica_id                           group_id                             synchronization_health_desc                                  connected_state_desc                                         operational_state_desc                                      
-------- ------------------------------------------------------------ ------------------------------------ ------------------------------------ ------------------------------------------------------------ ------------------------------------------------------------ ------------------------------------------------------------
       1 PRIMARY                                                      C3411DB7-8A80-41F3-A6D4-FCCCCC85D8ED 04539685-21FA-DFF2-D990-B45A6BCDD4CD HEALTHY                                                      CONNECTED                                                    ONLINE                                                      
       0 SECONDARY                                                    E989DD35-3F99-4E72-89A9-D21ACE041099 04539685-21FA-DFF2-D990-B45A6BCDD4CD HEALTHY                                                      CONNECTED                                                    NULL                                                        
       0 SECONDARY                                                    FC54D6DC-88C3-4107-A996-7B2E9C0C07B1 04539685-21FA-DFF2-D990-B45A6BCDD4CD HEALTHY                                                      CONNECTED                                                    NULL                                                        
       1 SECONDARY                                                    0EAC444F-1CF1-8D21-0178-B43D2842ACF5 6BC05A51-AA36-A196-09BD-481D7A0973C0 HEALTHY                                                      CONNECTED                                                    ONLINE                                                      

(4 rows affected)
mssql@ag2-0:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
use agdb;
SELECT * FROM test_table;
"
Changed database context to 'agdb'.
id          name                                              
----------- --------------------------------------------------
          1 DAG Setup                                         

(1 rows affected)
```

You should see the "1 DAG Setup" row, confirming that data is replicating correctly.


You can check on other nodes of ag1 and ag2, you should get the desired data, and confirm data replicating correcly. 

```bash
kubectl exec -it ag1-2 -n demo -c mssql -- bash
mssql@ag1-2:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
use agdb;
SELECT * FROM test_table;
"
Changed database context to 'agdb'.
id          name                                              
----------- --------------------------------------------------
          1 DAG Setup                                         

(1 rows affected)



kubectl exec -it ag2-2 -n demo -c mssql -- bash
mssql@ag2-2:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q " 
use agdb;
SELECT * FROM test_table;
"
Changed database context to 'agdb'.
id          name                                              
----------- --------------------------------------------------
          1 DAG Setup                                         

(1 rows affected)
```


## Perform In Cluster Failover: Automatically handled by KubeDB

To test automatic failover, we will force the primary member to restart. As the primary member (pod) becomes unavailable, the rest of the members will elect a primary member by election.

```bash
$ kubectl get pods -n demo 
NAME    READY   STATUS    RESTARTS   AGE
ag1-0   2/2     Running   0          157m
ag1-1   2/2     Running   0          155m
ag1-2   2/2     Running   0          155m


$ kubectl delete pod -n demo ag1-0 
pod "ag1-0" deleted

$ kubectl get pods -n demo
NAME    READY   STATUS    RESTARTS   AGE
ag1-0   2/2     Running   0          18s
ag1-1   2/2     Running   0          158m
ag1-2   2/2     Running   0          157m
```


Now find the new primary pod by running this command.
```bash
$ kubectl get pods -n demo --selector=app.kubernetes.io/instance=ag1 -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\n"}{end}'
ag1-0	
ag1-1	secondary
ag1-2	primary
```

We can see that, the primary node is now is `ag1-2`. The old primary pod `ag1-0` role will be set when old primary joins with the new primary as secondary.  
Lets exec into this new primary and see the availability replica's role.
```bash
$ kubectl exec -it ag1-2 -c mssql -n demo -- bash
mssql@ag1-2:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q " 
SELECT ar.replica_server_name, ars.role_desc
FROM sys.dm_hadr_availability_replica_states ars
INNER JOIN sys.availability_replicas ar ON ars.replica_id = ar.replica_id;
go
"
replica_server_name                                                                                                                                                                                                                                              role_desc                                                   
---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- ------------------------------------------------------------
ag1-0                                                                                                                                                                                                                                                            SECONDARY                                                   
ag1-1                                                                                                                                                                                                                                                            SECONDARY                                                   
ag1-2                                                                                                                                                                                                                                                            PRIMARY                                                     
ag1                                                                                                                                                                                                                                                              PRIMARY                                                     
ag2                                                                                                                                                                                                                                                              SECONDARY                                                   

(5 rows affected)
```

We can see that new primary is `ag1-2` and the old primary `ag1-0` joined the availability group cluster as secondary. MSSQLServer status is `Ready` now. We can see the updated pod labels.
```bash
$ kubectl get pods -n demo --selector=app.kubernetes.io/instance=ag1 -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\n"}{end}'
mssqlserver-ag-cluster-0	secondary
mssqlserver-ag-cluster-1	primary
mssqlserver-ag-cluster-2	secondary
````




## Performing Data Center Failover: Change DAG replica's primary role from ag1 to ag2

The following steps show how to fail over the primary role from `ag1` to `ag2`.

#### 1. Prepare for Zero Data Loss (If `ag1` is Online)

To prevent data loss, switch the DAG to synchronous replication. Execute this command on the primary replicas of **both `ag1` and `ag2`** (the 'GLOBAL PRIMARY' and the 'FORWARDER'.

```bash
$ kubectl exec -it ag1-2 -c mssql -n demo -- bash

mssql@ag1-2:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
SELECT g.name, r.replica_server_name, r.availability_mode_desc, r.failover_mode_desc
FROM sys.availability_groups AS g JOIN sys.availability_replicas AS r ON g.group_id = r.group_id
"
name                                                                                                                             replica_server_name                                                                                                                                                                                                                                              availability_mode_desc                                       failover_mode_desc                                          
-------------------------------------------------------------------------------------------------------------------------------- ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- ------------------------------------------------------------ ------------------------------------------------------------
ag1                                                                                                                              ag1-0                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag1                                                                                                                              ag1-1                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag1                                                                                                                              ag1-2                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
dag                                                                                                                              ag1                                                                                                                                                                                                                                                              ASYNCHRONOUS_COMMIT                                          MANUAL                                                      
dag                                                                                                                              ag2                                                                                                                                                                                                                                                              ASYNCHRONOUS_COMMIT                                          MANUAL                                                      

(5 rows affected)




mssql@ag1-2:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
ALTER AVAILABILITY GROUP dag 
MODIFY AVAILABILITY GROUP ON 
  'ag1' WITH ( AVAILABILITY_MODE = SYNCHRONOUS_COMMIT ),
  'ag2' WITH ( AVAILABILITY_MODE = SYNCHRONOUS_COMMIT);
"

 /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
SELECT g.name, r.replica_server_name, r.availability_mode_desc, r.failover_mode_desc
FROM sys.availability_groups AS g JOIN sys.availability_replicas AS r ON g.group_id = r.group_id
"
name                                                                                                                             replica_server_name                                                                                                                                                                                                                                              availability_mode_desc                                       failover_mode_desc                                          
-------------------------------------------------------------------------------------------------------------------------------- ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- ------------------------------------------------------------ ------------------------------------------------------------
ag1                                                                                                                              ag1-0                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag1                                                                                                                              ag1-1                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag1                                                                                                                              ag1-2                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
dag                                                                                                                              ag1                                                                                                                                                                                                                                                              SYNCHRONOUS_COMMIT                                           MANUAL                                                      
dag                                                                                                                              ag2                                                                                                                                                                                                                                                              SYNCHRONOUS_COMMIT                                           MANUAL                                                      

(5 rows affected)
```



```bash
$ kubectl exec -it ag2-0 -c mssql -n demo -- bash
mssql@ag2-0:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
SELECT g.name, r.replica_server_name, r.availability_mode_desc, r.failover_mode_desc
FROM sys.availability_groups AS g JOIN sys.availability_replicas AS r ON g.group_id = r.group_id
"
name                                                                                                                             replica_server_name                                                                                                                                                                                                                                              availability_mode_desc                                       failover_mode_desc                                          
-------------------------------------------------------------------------------------------------------------------------------- ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- ------------------------------------------------------------ ------------------------------------------------------------
ag2                                                                                                                              ag2-0                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag2                                                                                                                              ag2-1                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag2                                                                                                                              ag2-2                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
dag                                                                                                                              ag1                                                                                                                                                                                                                                                              ASYNCHRONOUS_COMMIT                                          MANUAL                                                      
dag                                                                                                                              ag2                                                                                                                                                                                                                                                              ASYNCHRONOUS_COMMIT                                          MANUAL                                                      

(5 rows affected)
mssql@ag2-0:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
ALTER AVAILABILITY GROUP dag 
MODIFY AVAILABILITY GROUP ON 
  'ag1' WITH ( AVAILABILITY_MODE = SYNCHRONOUS_COMMIT ),
  'ag2' WITH ( AVAILABILITY_MODE = SYNCHRONOUS_COMMIT);
"
mssql@ag2-0:/$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No -Q "
SELECT g.name, r.replica_server_name, r.availability_mode_desc, r.failover_mode_desc
FROM sys.availability_groups AS g JOIN sys.availability_replicas AS r ON g.group_id = r.group_id
"
name                                                                                                                             replica_server_name                                                                                                                                                                                                                                              availability_mode_desc                                       failover_mode_desc                                          
-------------------------------------------------------------------------------------------------------------------------------- ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- ------------------------------------------------------------ ------------------------------------------------------------
ag2                                                                                                                              ag2-0                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag2                                                                                                                              ag2-1                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
ag2                                                                                                                              ag2-2                                                                                                                                                                                                                                                            SYNCHRONOUS_COMMIT                                           MANUAL                                                      
dag                                                                                                                              ag1                                                                                                                                                                                                                                                              SYNCHRONOUS_COMMIT                                           MANUAL                                                      
dag                                                                                                                              ag2                                                                                                                                                                                                                                                              SYNCHRONOUS_COMMIT                                           MANUAL                                                      

(5 rows affected)
```


## Cleaning up

> Be careful when you set the `deletionPolicy` to `WipeOut`. Because there is no option to trace the database resources if once deleted the database.


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
- Want to set up SQL Server Distributed Availability Group clusters? Check how to [Configure SQL Server Distributed Availability Group Cluster](/docs/guides/mssqlserver/clustering/ag_cluster.md)
- Detail concepts of [MSSQLServer Object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).





