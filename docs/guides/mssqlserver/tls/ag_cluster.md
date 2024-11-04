---
title: SQL Server Availability Group TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: mssql-ag-tls-availability-group
    name: Availability Group (HA Cluster)
    parent: ms-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run SQL Server Availability Group with TLS/SSL Encryption

KubeDB supports providing TLS/SSL encryption for MSSQLServer. This tutorial will show you how to use KubeDB to run a MSSQLServer with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- Install [csi-driver-cacerts](https://github.com/kubeops/csi-driver-cacerts) which will be used to add self-signed ca certificates to the OS trusted certificate store (eg, /etc/ssl/certs/ca-certificates.crt)

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver/tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in MSSQLServer.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificates`
    - `clientTLS`


- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.

  - `apiGroup` is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` is the type of resource that is being referenced. KubeDB supports both `Issuer` and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.

- `clientTLS` This setting determines whether TLS (Transport Layer Security) is enabled for the MS SQL Server.
  - If set to `true`, the sql server will be provisioned with `TLS`, and you will need to install the [csi-driver-cacerts](https://github.com/kubeops/csi-driver-cacerts) which will be used to add self-signed ca certificates to the OS trusted certificate store (/etc/ssl/certs/ca-certificates.crt).
  - If set to `false`, TLS will not be enabled for SQL Server. However, the Issuer will still be used to configure a TLS-enabled WAL-G proxy server, which is necessary for performing SQL Server backup operations.

- `certificates` (optional) are a list of certificates used to configure the server and/or client certificate.

Read about the fields in details in [mssqlserver concept](/docs/guides/mssqlserver/concepts/mssqlserver.md).


## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in MSSQLServer. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you CA certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mssqlserver/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls mssqlserver-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

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

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/tls/issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

## TLS/SSL encryption in SQL Server Availability Group

Below is the YAML for MSSQLServer in Availability Group Mode.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-ag-tls
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
  internalAuth:
    endpointCert:
      issuerRef:
        apiGroup: cert-manager.io
        name: mssqlserver-ca-issuer
        kind: Issuer
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: true
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

### Deploy MSSQLServer in Availability Group Mode

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/tls/mssql-ag-tls.yaml
ms.kubedb.com/mssql-ag-tls created
```

Now, wait until `mssql-ag-tls` has status `Ready`. i.e,

```bash
$ watch kubectl get ms -n demo
Every 2.0s: kubectl get ms -n demo                           
NAME           VERSION     STATUS   AGE
mssql-ag-tls   2022-cu12   Ready    4m26s
```


### Verify TLS/SSL in MSSQLServer in Availability Group Mode

Now, connect to this database by exec into a pod and verify if `tls` has been set up as intended.

```bash
$ kubectl describe secret -n demo mssql-ag-tls-client-cert
Name:         mssql-ag-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mssql-ag-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mssqlservers.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: mssql-ag-tls-client-cert
              cert-manager.io/common-name: mssql
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: mssqlserver-ca-issuer
              cert-manager.io/subject-organizationalunits: client
              cert-manager.io/subject-organizations: kubedb
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
tls.crt:  1180 bytes
tls.key:  1675 bytes
ca.crt:   1164 bytes
```


Now, we can connect with tls to the mssqlserver and write some data

```bash
$ kubectl get secret -n demo mssql-ag-tls-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secret -n demo mssql-ag-tls-auth -o jsonpath='{.data.\password}' | base64 -d
Ng1DaJSNjZkgXXFX
```

```bash
$ kubectl exec -it -n demo mssql-ag-tls-0 -c mssql -- bash
mssql@mssql-ag-tls-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P Ng1DaJSNjZkgXXFX -N 
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
mssqlagtls                                                                                                            

(1 rows affected)
1> select replica_server_name from sys.availability_replicas;
2> go
replica_server_name                                                                                                                                                                                                                                             
-------------------------------------------------------------------------------------------
mssql-ag-tls-0                                                                                                                                                                                                                                                  
mssql-ag-tls-1                                                                                                                                                                                                                                                  
mssql-ag-tls-2   
(3 rows affected)
1> select database_name	from sys.availability_databases_cluster;
2> go
database_name                                                                                                                   
------------------------------------------------------------------------------------------
agdb1                                                                                                                           
agdb2                                                                                                                           

(2 rows affected)

```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo mssqlserver/mssql-ag-tls -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
mssqlserver.kubedb.com/mssql-ag-tls patched

$ kubectl delete -n demo mssqlserver mssql-ag-tls
mssqlserver.kubedb.com "mssql-ag-tls" deleted

$ kubectl delete issuer mssqlserver-ca-issuer
clusterissuer.cert-manager.io "mssqlserver-ca-issuer" deleted
```

## Next Steps

- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- [Backup and Restore](/docs/guides/mssqlserver/backup/overview/index.md) MSSQLServer databases using KubeStash. .
