---
title: Run MSSQLServer with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: ms-configuration-using-podtemplate
    name: Customize PodTemplate
    parent: ms-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run MSSQLServer with Custom PodTemplate

KubeDB supports providing custom configuration for MSSQLServer via [PodTemplate](/docs/guides/mssqlserver/concepts/mssqlserver.md#specpodtemplate). This tutorial will show you how to use KubeDB to run a MSSQLServer database with custom configuration using PodTemplate.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/configuration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for MSSQLServer.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
  - annotations (pod's annotation)
- controller
  - annotations (petset's annotation)
- spec:
  - containers
  - volumes
  - podPlacementPolicy
  - serviceAccountName
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - livenessProbe
  - readinessProbe
  - lifecycle

Read about the fields in details in [PodTemplate concept](/docs/guides/mssqlserver/concepts/mssqlserver.md#specpodtemplate),

## CRD Configuration

Below is the YAML for the MSSQLServer created in this example. Here   
- [`spec.podTemplate.spec.containers[].name`](/docs/guides/mssqlserver/concepts/mssqlserver.md#specpodtemplatespeccontainersname) field is used to specify the name of the container.   
- [`spec.podTemplate.spec.containers[].env`](/docs/guides/mssqlserver/concepts/mssqlserver.md#specpodtemplatespeccontainersenv) field specifies the environment variables to pass to the MSSQLServer docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/r/microsoft/mssql-server).   
- [`spec.podTemplate.spec.containers[].resources`](/docs/guides/mssqlserver/concepts/mssqlserver.md#specpodtemplatespeccontainersresources) is an optional field. This can be used to request compute resources required by containers of the database pods.

Here we have set 
- [memory limit](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-mssql-conf?view=sql-server-ver16#memorylimit), The `MSSQL_MEMORY_LIMIT_MB` setting controls the amount of physical memory (in MB) available to SQL Server. The default is 80% of the physical memory, to prevent out-of-memory (OOM) conditions. The above configuration changes the memory available to SQL Server to 2.5 GB (2,560 MB).
- [SQL Server Locale](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-mssql-conf?view=sql-server-ver16#lcid), The language.lcid setting changes the SQL Server locale to any supported language identifier (LCID). The above example changes the locale to French (1036):

- [MSSQL_PID](https://mcr.microsoft.com/en-us/product/mssql/server/about#configuration:~:text=MSSQL_PID%20is%20the,documentation%20here.), This variable determines which `SQL Server edition` will run inside the container. The acceptable values for `MSSQL_PID` are:   
`Developer`: This will run the container using the Developer Edition (this is the default if no MSSQL_PID environment variable is supplied)    
`Express`: This will run the container using the Express Edition    
`Evaluation`: This will run the container using the Evaluation Edition    
`Standard`: This will run the container using the Standard Edition   
`Enterprise`: This will run the container using the Enterprise Edition   
`EnterpriseCore`: This will run the container using the Enterprise Edition Core   
`<valid product id>`: This will run the container with the edition that is associated with the PID

Now, create an Issuer/ClusterIssuer which will be used to generate the certificate used for TLS configurations.

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

### Create MSSQLServer CR with Custom Configuration using PodTemplate

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: custom-config-podtemplate
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  storageType: Durable
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: "Evaluation"
            - name: MSSQL_MEMORY_LIMIT_MB
              value: "2560"
            - name: MSSQL_LCID
              value: "1036"
          resources:
            requests:
              cpu: "500m"
              memory: "1.5Gi"
            limits:
              cpu: "3"
              memory: "6Gi"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/configuration/custom-config-podtemplate.yaml
mssqlserver.kubedb.com/custom-config-podtemplate created
```

Now, wait a few minutes. KubeDB operator will create necessary Petset, PVCs, Services, Secrets etc. If everything goes well, we will see that a pod with the name `custom-config-podtemplate-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                          READY   STATUS    RESTARTS   AGE
custom-config-podtemplate-0   1/1     Running   0          16m
```

Now, check if the database has started with the custom configuration we have provided.

```bash
$ kubectl get pod -n demo custom-config-podtemplate-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "3",
    "memory": "6Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}


$ kubectl get secrets -n demo custom-config-podtemplate-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secrets -n demo custom-config-podtemplate-auth -o jsonpath='{.data.\password}' | base64 -d
3K7lJibYg3y6ICXc

$ kubectl exec -it custom-config-podtemplate-0 -n demo -c mssql -- bash
mssql@custom-config-podtemplate-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P 3K7lJibYg3y6ICXc
1> SELECT physical_memory_kb / 1024 AS physical_memory_mb FROM sys.dm_os_sys_info;
2> go
physical_memory_mb  
--------------------
                2560

(1 rows affected)
1> SELECT default_language_name FROM sys.server_principals WHERE name = 'sa';
2> go
default_language_name                                                                                                           
------------------------------------------------------------------------
Français                                                                                                                        

(1 rows affected)
1> select @@version
2> go
--------------------
Microsoft SQL Server 2022 (RTM-CU12) (KB5033663) - 16.0.4115.5 (X64) 
	Mar  4 2024 08:56:10 
	Copyright (C) 2022 Microsoft Corporation
	Enterprise Evaluation Edition (64-bit) on Linux (Ubuntu 22.04.4 LTS) <X64>                                                                                          

(1 rows affected)
```

You can see that our desired configuration is applied successfully. 

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo ms/custom-config-podtemplate -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo ms/custom-config-podtemplate
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart MSSQLServer](/docs/guides/mssqlserver/quickstart/quickstart.md) with KubeDB Operator.
- [Backup and Restore](/docs/guides/mssqlserver/backup/overview/index.md) MSSQLServer databases using Stash.
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
