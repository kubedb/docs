---
title: SQL Server Standalone Guide
menu:
  docs_{{ .version }}:
    identifier: ms-clustering-standalone
    name: Standalone Guide
    parent: ms-clustering
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Microsoft SQL Server Standalone

This tutorial will show you how to use KubeDB to run a Standalone SQL Server database.

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


> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver/standalone/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/standalone/) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).



## Deploy Microsoft SQL Server database

At first, we need to create an Issuer/ClusterIssuer which will be used to generate the certificate used for TLS configurations.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
-
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


### Deploy Standalone Microsoft SQL Server
KubeDB implements a `MSSQLServer` CRD to define the specification of a Microsoft SQL Server database. Below is the `MSSQLServer` object created in this tutorial.

Here, our issuer `mssqlserver-ca-issuer` is ready to deploy a `MSSQLServer`. Below is the YAML of SQL Server that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-standalone
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/standalone/mssqlserver-standalone.yaml
mssqlserver.kubedb.com/mssqlserver-standalone created
```

Here,

- `spec.version` is the name of the MSSQLServerVersion CR where the docker images are specified. In this tutorial, a MSSQLServer `2022-cu12` database is going to be created.
- `spec.storageType` specifies the type of storage that will be used for MSSQLServer database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MSSQLServer database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). Here `tls.clientTLS: false` means tls will not be enabled for SQL Server but the Issuer will be used to configure tls enabled wal-g proxy-server which is required for SQL Server backup operation.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MSSQLServer` CR or which resources KubeDB should keep or delete when you delete `MSSQLServer` CR. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy)

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in storage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MSSQLServer` objects using Kubernetes api. When a `MSSQLServer` object is created, KubeDB operator will create a new PetSet and a Service with the matching MSSQLServer object name. KubeDB operator will also create a governing service for PetSets with the name `<MSSQLServerName>-pods`, if one is not already present.

```bash
$ kubectl get petset -n demo mssqlserver-standalone
NAME                     AGE
mssqlserver-standalone   13m


$ kubectl get pvc -n demo
NAME                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mssqlserver-standalone-0   Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f12   1Gi        RWO            standard       15m


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                STORAGECLASS   REASON   AGE
pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f12   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-standalone-0   standard                15m


kubectl get service -n demo
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
mssqlserver-standalone        ClusterIP   10.96.128.61   <none>        1433/TCP   15m
mssqlserver-standalone-pods   ClusterIP   None           <none>        1433/TCP   15m

```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created and is able to accept client connections. Run the following command to see the modified MSSQLServer object:

```bash
$ kubectl get ms -n demo mssqlserver-standalone -o yaml
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MSSQLServer","metadata":{"annotations":{},"name":"mssqlserver-standalone","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","tls":{"clientTLS":false,"issuerRef":{"apiGroup":"cert-manager.io","kind":"Issuer","name":"mssqlserver-ca-issuer"}},"version":"2022-cu12"}}
  creationTimestamp: "2024-10-08T10:24:13Z"
  finalizers:
    - kubedb.com
  generation: 2
  name: mssqlserver-standalone
  namespace: demo
  resourceVersion: "235027"
  uid: 05a5d98e-b2c1-4f81-84d9-87257bd76f08
spec:
  authSecret:
    name: mssqlserver-standalone-auth
  coordinator:
    resources: {}
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
  replicas: 1
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
        secretName: mssqlserver-standalone-server-cert
        subject:
          organizationalUnits:
            - server
          organizations:
            - kubedb
      - alias: client
        secretName: mssqlserver-standalone-client-cert
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
  version: 2022-cu12
status:
  conditions:
    - lastTransitionTime: "2024-10-08T10:24:13Z"
      message: 'The KubeDB operator has started the provisioning of MSSQL: demo/mssqlserver-standalone'
      observedGeneration: 2
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2024-10-08T10:25:14Z"
      message: All replicas are ready for MSSQLServer demo/mssqlserver-standalone
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2024-10-08T10:25:55Z"
      message: database demo/mssqlserver-standalone is accepting connection
      observedGeneration: 2
      reason: AcceptingConnection
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2024-10-08T10:25:55Z"
      message: database demo/mssqlserver-standalone is ready
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: Ready
    - lastTransitionTime: "2024-10-08T10:26:14Z"
      message: 'The MSSQL: demo/mssqlserver-standalone is successfully provisioned.'
      observedGeneration: 2
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready
```

## Connect with MSSQLServer database

KubeDB operator has created a new Secret called `mssqlserver-standalone-auth` *(format: {mssqlserver-object-name}-auth)* for storing the sa password for `mssqlserver`. This secret contains a `username` key which contains the *username* for MSSQLServer SA and a `password` key which contains the *password* for MSSQLServer SA user.

If you want to use an existing secret please specify that when creating the MSSQLServer object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `sa` as value of `username` and a strong password for the sa user. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, we need `username` and `password` to connect to this database from `kubectl exec` command. In this example  `mssqlserver-standalone-auth` secret holds username and password

```bash
$ kubectl get secret -n demo mssqlserver-standalone-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secret -n demo mssqlserver-standalone-auth -o jsonpath='{.data.\password}' | base64 -d
axgXHj4oRIVQ1ocK
```
We can exec into the pod `mssqlserver-standalone-0` using the following command:
```bash
kubectl exec -it -n demo mssqlserver-standalone-0 -c mssql -- bash
mssql@mssqlserver-standalone-0:/$
```

You can connect to the database using the `sqlcmd` utility, which comes with the mssql-tools package on Linux. To check the installed version of sqlcmd, run the following command:
```bash
mssql@mssqlserver-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd "-?"
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


Now, connect to the database using username and password
```bash
mssql@mssqlserver-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "axgXHj4oRIVQ1ocK"
1> select name from sys.databases
2> go
name                                                  
----------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                       
kubedb_system                                                                                                                   

(5 rows affected)
1> 
```



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
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: |
          {"apiVersion":"kubedb.com/v1alpha2","kind":"MSSQLServer","metadata":{"annotations":{},"name":"mssqlserver-standalone","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","tls":{"clientTLS":false,"issuerRef":{"apiGroup":"cert-manager.io","kind":"Issuer","name":"mssqlserver-ca-issuer"}},"version":"2022-cu12"}}
      creationTimestamp: "2024-10-08T10:24:44Z"
      generation: 1
      labels:
        app.kubernetes.io/component: database
        app.kubernetes.io/instance: mssqlserver-standalone
        app.kubernetes.io/managed-by: kubedb.com
        app.kubernetes.io/name: mssqlservers.kubedb.com
      name: mssqlserver-standalone
      namespace: demo
      ownerReferences:
        - apiVersion: kubedb.com/v1alpha2
          blockOwnerDeletion: true
          controller: true
          kind: MSSQLServer
          name: mssqlserver-standalone
          uid: 05a5d98e-b2c1-4f81-84d9-87257bd76f08
      resourceVersion: "234806"
      uid: 6a623112-9714-4f14-a1a2-d3f5a3b50df1
    spec:
      appRef:
        apiGroup: kubedb.com
        kind: MSSQLServer
        name: mssqlserver-standalone
        namespace: demo
      clientConfig:
        service:
          name: mssqlserver-standalone
          path: /
          port: 1433
          scheme: tcp
        url: tcp(mssqlserver-standalone.demo.svc:1433)/
      secret:
        name: mssqlserver-standalone-auth
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
$ kubectl patch -n demo ms mssqlserver-standalone -p '{"spec":{"deletionPolicy":"DoNotTerminate"}}' --type="merge"
mssqlserver.kubedb.com/mssqlserver-standalone patched

$ kubectl delete ms -n demo mssqlserver-standalone
The MSSQLServer "mssqlserver-standalone" is invalid: spec.deletionPolicy: Invalid value: "mssqlserver-standalone": Can not delete as deletionPolicy is set to "DoNotTerminate"
```

Now, run `kubectl patch -n demo ms mssqlserver-standalone -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"` to set `spec.deletionPolicy` to `Halt` (which deletes the mssqlserver object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.

Learn details of all `DeletionPolicy` [here](/docs/guides/mssqlserver/concepts/mssqlserver.md#specdeletionpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `MSSQLServer` object `deletionPolicy` to `Halt`.

When the [DeletionPolicy](/docs/guides/mssqlserver/concepts/mssqlserver.md#specdeletionpolicy) is set to `halt` and the MSSQLServer object is deleted, the KubeDB operator will delete the PetSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `deletionPolicy` to `halt` in existing database using `patch` command for testing.

At first, run `kubectl patch -n demo ms mssqlserver-standalone -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"`. Then delete the mssqlserver object,

```bash
$ kubectl delete ms -n demo mssqlserver-standalone
mssqlserver.kubedb.com "mssqlserver-standalone" deleted
```

Now, run the following command to get mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get ms,petset,pod,svc,secret,pvc -n demo 
NAME                                        TYPE                       DATA   AGE
secret/mssqlserver-ca                       kubernetes.io/tls          2      40m
secret/mssqlserver-standalone-auth          kubernetes.io/basic-auth   2      30m
secret/mssqlserver-standalone-client-cert   kubernetes.io/tls          3      30m
secret/mssqlserver-standalone-config        Opaque                     1      30m
secret/mssqlserver-standalone-server-cert   kubernetes.io/tls          3      30m

NAME                                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mssqlserver-standalone-0   Bound    pvc-656e3bd1-65da-441c-851f-2ae076f8ebbd   1Gi        RWO            standard       29m
```

From the above output, you can see that all mssqlserver resources(`MSSQLServer`, `PetSet`, `Pod`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your mssqlserver again using these resources.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `MSSQLServer` object `deletionPolicy` to `Delete`. In this setting, `PetSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [DeletionPolicy](/docs/guides/mssqlserver/concepts/mssqlserver.md#specdeletionpolicy) is set to `Delete` and the MSSQLServer object is deleted, the KubeDB operator will delete the PetSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `deletionPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl patch -n demo ms mssqlserver-standalone -p '{"spec":{"deletionPolicy":"Delete"}}' --type="merge"
mssqlserver.kubedb.com/mssqlserver-standalone patched
$ kubectl delete ms -n demo mssqlserver-standalone 
mssqlserver.kubedb.com "mssqlserver-standalone" deleted
```

Now, run the following command to get all mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get ms,petset,pod,svc,secret,pvc -n demo 
NAME                                        TYPE                       DATA   AGE
secret/mssqlserver-ca                       kubernetes.io/tls          2      49m
secret/mssqlserver-standalone-auth          kubernetes.io/basic-auth   2      39m
secret/mssqlserver-standalone-client-cert   kubernetes.io/tls          3      39m
secret/mssqlserver-standalone-config        Opaque                     1      39m
secret/mssqlserver-standalone-server-cert   kubernetes.io/tls          3      39m
```

From the above output, you can see that all mssqlserver resources(`MSSQLServer`, `PetSet`, `Pod`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your mssqlserver using `snapshots`(if previously taken) and `Secrets`.

>If you don't set the deletionPolicy then the kubeDB set the DeletionPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `MSSQLServer` database and relevant resources without any tracking by setting `deletionPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `MSSQLServer` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `deletionPolicy` is set to `WipeOut`.

Let's set `deletionPolicy` set to `WipeOut` and delete the database using the following command:

```yaml
$ kubectl patch -n demo ms mssqlserver-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
mssqlserver.kubedb.com/mssqlserver-standalone patched

$ kubectl delete ms -n demo mssqlserver-standalone
mssqlserver.kubedb.com "mssqlserver-standalone" deleted
```

Now, run the following command to get all mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get ms,petset,pod,svc,secret,pvc -n demo 
NAME                       TYPE                DATA   AGE
secret/mssqlserver-ca      kubernetes.io/tls   2      53m
```

From the above output, you can see that all mssqlserver resources are deleted. there is no option to recreate/reinitialize your database if `deletionPolicy` is set to `WipeOut`.

>Be careful when you set the `deletionPolicy` to `WipeOut`. Because there is no option to trace the database resources if once deleted the database.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mssqlserver/mssqlserver-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete mssqlserver -n demo mssqlserver-standalone
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
- Want to set up SQL Server Availability Group clusters? Check how to [Configure SQL Server Availability Gruop Cluster](/docs/guides/mssqlserver/clustering/ag_cluster.md)
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).