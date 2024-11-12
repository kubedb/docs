---
title: Run MSSQLServer with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: ms-configuration-config-file
    name: Config File
    parent: ms-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for MSSQLServer. This tutorial will show you how to use KubeDB to run SQL Server with custom configuration.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

SQL Server allows configuring database via configuration file. The default configuration file for SQL Server deployed by `KubeDB` can be found in `/var/opt/mssql/mssql.conf`. When SQL Server starts, it will look for  configuration file in `/var/opt/mssql/mssql.conf`. If configuration file exist, this configuration will overwrite the existing defaults.

> To learn available configuration option of SQL Server see [Configure SQL Server on Linux](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-mssql-conf?view=sql-server-ver16).

At first, you have to create a config file named `mssql.conf` with your desired configuration. Then you have to create a [secret](https://kubernetes.io/docs/concepts/configuration/secret/) using this file. Then specify this secret name in `spec.configSecret.name` section while creating MSSQLServer CR. 

KubeDB will create a secret named `{mssqlserver-name}-config` with configuration file contents as the value of the key `mssql.conf` and mount this secret into `/var/opt/mssql/` directory of the database pod. The secret named `{mssqlserver-name}-config` will contain your desired configurations with some default configurations.

In this tutorial, we will configure sql server via a custom config file.

## Custom Configuration

At first, create `mssql.conf` file containing required configuration settings.

```ini
$ cat mssql.conf
[network]
tlsprotocols = 1.2
forceencryption = 1

[language]
lcid = 1036

[memory]
memorylimitmb = 2304
```

Here we have set 
- [memory limit](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-mssql-conf?view=sql-server-ver16#memorylimit), The `memory.memorylimitmb` setting controls the amount of physical memory (in MB) available to SQL Server. The default is 80% of the physical memory, to prevent out-of-memory (OOM) conditions. The above configuration changes the memory available to SQL Server to 2.25 GB (2,304 MB).
- [SQL Server Locale](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-mssql-conf?view=sql-server-ver16#lcid), The language.lcid setting changes the SQL Server locale to any supported language identifier (LCID). The above example changes the locale to French (1036):
- [TLS](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-mssql-conf?view=sql-server-ver16#tls) The `network.forceencryption` If 1, then SQL Server forces all connections to be encrypted. By default, this option is 0. The `network.tlsprotocols`	A comma-separated list of which TLS protocols are allowed by SQL Server. SQL Server always attempts to negotiate the strongest allowed protocol. If a client doesn't support any allowed protocol, SQL Server rejects the connection attempt. For compatibility, all supported protocols are allowed by default (1.2, 1.1, 1.0). If your clients support TLS 1.2, Microsoft recommends allowing only TLS 1.2.



Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo ms-custom-config --from-file=./mssql.conf
secret/ms-custom-config created
```

Verify the secret has the configuration file.
```bash
$ kubectl get secret -n demo ms-custom-config -oyaml
```

```yaml
apiVersion: v1
data:
  mssql.conf: W25ldHdvcmtdCnRsc3Byb3RvY29scyA9IDEuMgpmb3JjZWVuY3J5cHRpb24gPSAxCgpbbGFuZ3VhZ2VdCmxjaWQgPSAxMDM2CgpbbWVtb3J5XQptZW1vcnlsaW1pdG1iID0gMjMwNA==
kind: Secret
metadata:
  creationTimestamp: "2024-10-16T06:12:28Z"
  name: ms-custom-config
  namespace: demo
  resourceVersion: "451820"
  uid: e7242e3a-d5dc-4705-a0f3-20b0ff0a59d3
type: Opaque
```



Now, we need to create an Issuer/ClusterIssuer which will be used to generate the certificate used for TLS configurations.

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



Now, create MSSQLServer CR specifying `spec.configSecret` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-custom-config
  namespace: demo
spec:
  version: "2022-cu12"
  configSecret:
    name: ms-custom-config
  replicas: 1
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/configuration/mssql-custom-config.yaml
mssqlserver.kubedb.com/mssql-custom-config created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secrets etc. If everything goes well, we will see that a pod with the name `mssql-custom-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo mssql-custom-config-0
NAME                    READY   STATUS    RESTARTS   AGE
mssql-custom-config-0   1/1     Running   0          94s
```

Now, we will check if the database has started with the custom configuration we have provided.

Now, Let's connect to the MSSQLServer from inside the pod.

```bash
$ kubectl get secrets -n demo mssql-custom-config-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secrets -n demo mssql-custom-config-auth -o jsonpath='{.data.\password}' | base64 -d
AqRe6WIuqwKXLaWc

$ kubectl exec -it mssql-custom-config-0 -n demo -c mssql -- bash
mssql@mssql-custom-config-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1036
[network]
tlsprotocols = 1.2
forceencryption = 1
[memory]
memorylimitmb = 2304
mssql@mssql-custom-config-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P AqRe6WIuqwKXLaWc
1> SELECT encrypt_option FROM sys.dm_exec_connections WHERE session_id = @@SPID;
2> go
encrypt_option                          
----------------------------------------
TRUE                                    

(1 rows affected)
1> SELECT default_language_name FROM sys.server_principals WHERE name = 'sa';
2> go
default_language_name                                                                                                           
-----------------------------------------------------------
Français                                                                                                                        

(1 rows affected)
1> SELECT physical_memory_kb / 1024 AS physical_memory_mb FROM sys.dm_os_sys_info;
2> go
physical_memory_mb  
--------------------
2304
(1 rows affected)
1> 
```


As we can see from the configuration of running sql server, the configuration given in the config secret has been set successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo ms/mssql-custom-config -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"

$ kubectl delete -n demo ms/mssql-custom-config
mssqlserver.kubedb.com "mssql-custom-config" deleted

$ kubectl delete -n demo secret ms-custom-config
mssqlserver.kubedb.com "mssql-custom-config" deleted

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mssqlserver/backup/overview/index.md) MSSQLServer databases using KubeStash.
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Detail concepts of [MSSQLServerVersion object](/docs/guides/mssqlserver/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
