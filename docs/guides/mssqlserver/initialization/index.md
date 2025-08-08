---
title: Initialize SQL Server using Script
menu:
  docs_{{ .version }}:
    identifier: ms-initialization
    name: Initialization Using Script
    parent: guides-mssqlserver
    weight: 41
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# Initialize Microsoft SQL Server using Script

This tutorial will show you how to use KubeDB to initialize a MSSQLServer database with `*.sql`, `*.sh` or `*.sql.gz` script.

In this tutorial, we will use .sql script stored in GitHub repository [kubedb/mssqlserver-init-scripts](https://github.com/kubedb/mssqlserver-init-scripts).

> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).


## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```
  
## Prepare Initialization Scripts

MSSQLServer supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use `init.sql` script from [mssqlserver-init-scripts](https://github.com/kubedb/mssqlserver-init-scripts) git repository to create a database named `mssql` and a table named `kubedb_init` in that database. 

We will use a ConfigMap as a script source. You can use any Kubernetes supported [volumes](https://kubernetes.io/docs/concepts/storage/volumes) as a script source.

At first, we will create a ConfigMap with `init.sql` file. Then, we will provide this ConfigMap as script source in `init.script` of the MSSQLServer CR spec.

Let's create a ConfigMap with the `init.sql` initialization script,

```bash
$ kubectl create configmap -n demo mssql-init-scripts \
--from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/mssqlserver-init-scripts/raw/master/init.sql)"
configmap/mssql-init-scripts created
```


## Deploy the Microsoft SQL Server database 

At first, we need to create an Issuer/ClusterIssuer which will be used to generate the certificate used for TLS configurations.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/standalone/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

### Deploy a Microsoft SQL Server database with Init-Script
KubeDB implements a `MSSQLServer` CRD to define the specification of a Microsoft SQL Server database. Below is the `MSSQLServer` object created in this tutorial.

Here, our issuer `mssqlserver-ca-issuer` is ready to deploy a `MSSQLServer`. Below is the YAML of SQL Server that we are going to create

<ul class="nav nav-tabs" id="definationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active  " id="st-tab" data-toggle="tab" href="#Standalone" role="tab" aria-controls="Standalone" aria-selected="true">Stand Alone</a>
  </li>

  <li class="nav-item">
    <a class="nav-link active" id="gr-tab" data-toggle="tab" href="#AvailabilityGroup" role="tab" aria-controls="AvailabilityGroup" aria-selected="false">Group Replication</a>
  </li>
</ul>


<div class="tab-content" id="definationTabContent">


  <div class="tab-pane fade show active" id="Standalone" role="tabpanel" aria-labelledby="gr-tab">

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ms-init
  namespace: demo
spec:
  version: "2022-cu19"
  storageType: Durable
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
  init:
    script:
      configMap:
        name: mssql-init-scripts
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/initialization/yamls/initializ-standalone.yaml
mssqlserver.kubedb.com/ms-init created
```
  </div>

  



  <div class="tab-pane fade" id="AvailabilityGroup" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ms-ag-init
  namespace: demo
spec:
  version: "2022-cu19"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb
        - mssql
      secondaryAccessMode: "All"
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
  init:
    script:
      configMap:
        name: mssql-init-scripts
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/initialization/yamls/initialize-ag-cluster.yaml
mssqlsever.kubedb.com/ms-ag-init created
```
  </div>

</div>





Here,
- `spec.init.script` specifies a script source used to initialize the database. The scripts will be executed alphabetically. In this tutorial, a sample .sql script from the git repository `https://github.com/kubedb/mssqlserver-init-scripts.git` is used to create a test database named `mssql`. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) instead of `ConfigMap`.  The \*.sql, \*sql.gz and/or \*.sh scripts that are stored inside the folder will be executed alphabetically. The scripts inside child folders will be skipped.

KubeDB operator watches for `MSSQLServer` objects using Kubernetes API. When a `MSSQLServer` object is created, KubeDB operator will create a PetSet and Services, Secrets, and other necessary resouces for this `MSSQLServer` Database.

```bash
$ kubectl dba describe ms -n demo ms-init
Name:         ms-init
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         MSSQLServer
Metadata:
  Creation Timestamp:  2025-08-06T10:35:28Z
  Finalizers:
    kubedb.com
  Generation:        3
  Resource Version:  224203
  UID:               0c2a7678-8086-4fc0-9d73-a232abf47135
Spec:
  Auth Secret:
    Active From:  2025-08-06T10:35:29Z
    Name:         ms-init-auth
  Auto Ops:
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Init:
    Script:
      Config Map:
        Name:  mssql-init-scripts
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Env:
          Name:   ACCEPT_EULA
          Value:  Y
          Name:   MSSQL_PID
          Value:  Evaluation
        Name:     mssql
        Resources:
          Limits:
            Memory:  2Gi
          Requests:
            Cpu:     1
            Memory:  1536Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Add:
              NET_BIND_SERVICE
            Drop:
              ALL
          Run As Group:     10001
          Run As Non Root:  true
          Run As User:      10001
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  mssql-init
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  256Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     10001
          Run As Non Root:  true
          Run As User:      10001
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:            10001
      Service Account Name:  ms-init
  Replicas:                  1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:  1Gi
  Storage Type:   Durable
  Tls:
    Certificates:
      Alias:        server
      Secret Name:  ms-init-server-cert
      Subject:
        Organizational Units:
          server
        Organizations:
          kubedb
      Alias:        client
      Secret Name:  ms-init-client-cert
      Subject:
        Organizational Units:
          client
        Organizations:
          kubedb
    Client TLS:  false
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mssqlserver-ca-issuer
  Version:        2022-cu19
Status:
  Conditions:
    Last Transition Time:  2025-08-06T10:35:28Z
    Message:               The KubeDB operator has started the provisioning of MSSQLServer: demo/ms-init
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-08-06T10:37:34Z
    Message:               All replicas are ready for MSSQLServer demo/ms-init
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-08-06T10:37:45Z
    Message:               database demo/ms-init is accepting connection
    Observed Generation:   3
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-08-06T10:37:45Z
    Message:               database demo/ms-init is ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-08-06T10:38:37Z
    Message:               The MSSQLServer: demo/ms-init is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. 

KubeDB operator has created a new Secret called `ms-init-auth`  for storing the password for MSSQLServer SA user.

```bash
$ kubectl view-secret -n demo ms-init-auth -a
password='9jtGBoona46wUYmL'
username='sa'
```


Let's connect ot the database pod and verify the `init.sql` script is executed successfully or not. 

```bash 
$ kubectl exec -it -n demo ms-init-0 -- bash
Defaulted container "mssql" out of: mssql, mssql-init (init)
mssql@ms-init-0:/$ cd init-database/
mssql@ms-init-0:/init-database$ ls
init.sql
mssql@ms-init-0:/init-database$ cat init.sql
-- 1) Create a database if it doesn't already exist
IF DB_ID(N'mssql') IS NULL
BEGIN
PRINT N'Creating database [mssql]...';
CREATE DATABASE [mssql];
END
GO

-- 2) Switch context
USE [mssql];
GO

-- 3) Drop the table if it already exists
IF OBJECT_ID(N'dbo.kubedb_init', N'U') IS NOT NULL
BEGIN
PRINT N'Dropping existing table [dbo.kubedb_init]...';
DROP TABLE dbo.kubedb_init;
END
GO

-- 4) Create the table with an IDENTITY primary key
PRINT N'Creating table [dbo.kubedb_init]...';
CREATE TABLE dbo.kubedb_init (
id   BIGINT             IDENTITY(1,1) NOT NULL CONSTRAINT PK_kubedb_init PRIMARY KEY,
name NVARCHAR(255)      NULL,
created_at DATETIME2    NOT NULL DEFAULT SYSUTCDATETIME()
);
GO

-- 5) Seed it with some test data
PRINT N'Inserting sample rows...';
INSERT INTO dbo.kubedb_init (name) VALUES
(N'name1'),
(N'name2'),
(N'name3'),
(N'name4'),
(N'name5'),
(N'name6'),
(N'name7'),
(N'name8');
GO

-- 6) Confirmation
DECLARE @cnt INT = (SELECT COUNT(*) FROM dbo.kubedb_init);
PRINT N'Inserted ' + CAST(@cnt AS NVARCHAR(10)) + N' rows into dbo.kubedb_init.';




mssql@ms-init-0:/init-database$ /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -No
1> select name from sys.databases;
2> go
name
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
mssql                                                                                                                           
kubedb_system

(6 rows affected)
1> use mssql;
2> go
Changed database context to 'mssql'.
1> select * from kubedb_init;
2> go
id                   name                                                                                                                                                                                                                                                            created_at
-------------------- --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- --------------------------------------
                   1 name1                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412
                   2 name2                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412
                   3 name3                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412
                   4 name4                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412
                   5 name5                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412
                   6 name6                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412
                   7 name7                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412
                   8 name8                                                                                                                                                                                                                                                                      2025-08-06 10:36:49.5645412

(8 rows affected)
```



As you can see here, the initial script has successfully created a `mssql` database and a table named `kubedb_init` in `mssql` database and inserted 8 rows of data into that table successfully.

## Cleaning up

To clean up the kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mssqlserver/ms-init -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mssqlserver/ms-init
kubectl delete ns demo
```

## Next Steps

- Initialize [MSSQLServer with Script](/docs/guides/mssqlserver/initialization/index.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
