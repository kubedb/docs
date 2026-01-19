---
title: Reconfigure Standalone MSSQLServer Database
menu:
  docs_{{ .version }}:
    identifier: ms-reconfigure-standalone
    name: Standalone
    parent: ms-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure MSSQLServer Standalone Database

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a MSSQLServer standalone database.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/mssqlserver/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mssqlserver](/docs/examples/mssqlserver/reconfigure) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `MSSQLServer` standalone using a supported version by `KubeDB` operator. Then we are going to apply `MSSQLServerOpsRequest` to reconfigure its configuration.

### Prepare MSSQLServer Standalone Database

Now, we are going to deploy a `MSSQLServer` standalone database with version `2022-cu12`.

### Deploy MSSQLServer standalone 

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

Now, we will create `mssql.conf` file containing required configuration settings.

```ini
$ cat mssql.conf
[memory]
memorylimitmb = 2048
```
Here, `memorylimitmb` is set to `2048`, whereas the default value is `12280`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo ms-custom-config --from-file=./mssql.conf
secret/ms-custom-config created
```

In this section, we are going to create a MSSQLServer object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ms-standalone
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

Let's create the `MSSQLServer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure/ms-standalone.yaml
MSSQLServer.kubedb.com/ms-standalone created
```

Now, wait until `ms-standalone` has status `Ready`. i.e,

```bash
$ kubectl get ms -n demo
NAME            VERSION    STATUS    AGE
ms-standalone   4.4.26      Ready     23s
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a MSSQLServer instance,
```bash
$ kubectl get secrets -n demo ms-standalone-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secrets -n demo ms-standalone-auth -o jsonpath='{.data.\password}' | base64 -d
SERtEyH1RMMEsvE0
```

Now let's connect to the SQL Server instance and run internal command to check the configuration we have provided.

```bash
$ kubectl exec -it -n demo ms-standalone-0 -c mssql -- bash
mssql@ms-standalone-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
[memory]
memorylimitmb = 2048
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P SERtEyH1RMMEsvE0
1> SELECT physical_memory_kb / 1024 AS physical_memory_mb FROM sys.dm_os_sys_info;
2> go
physical_memory_mb  
--------------------
                2048

(1 rows affected)
1> 
```

As we can see from the configuration of running MSSQLServer, the value of `physical_memory_mb` has been set to `2048`.

### Reconfigure using new secret

Now we will reconfigure this database to set `memorylimitmb` to `2560`.

Now, we will edit the `mssql.conf` file containing required configuration settings.

```ini
$ cat mssql.conf
[memory]
memorylimitmb = 2560
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-custom-config --from-file=./mssql.conf
secret/new-custom-config created
```

#### Create MSSQLServerOpsRequest

Now, we will use this secret to replace the previous secret using a `MSSQLServerOpsRequest` CR. The `MSSQLServerOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-reconfigure-standalone
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ms-standalone
  configuration:
    secretName: new-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `ms-standalone` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.secretName` specifies the name of the new secret.
- Have a look [here](/docs/guides/mssqlserver/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure/mops-reconfigure-standalone.yaml
MSSQLServeropsrequest.ops.kubedb.com/mops-reconfigure-standalone created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `MSSQLServer` object.

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CR,

```bash
$ watch kubectl get MSSQLServeropsrequest -n demo
Every 2.0s: kubectl get MSSQLServeropsrequest -n demo

NAME                           TYPE          STATUS       AGE
msops-reconfigure-standalone   Reconfigure   Successful   2m42s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe MSSQLServeropsrequest -n demo mops-reconfigure-standalone
Name:         msops-reconfigure-standalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-07T12:46:35Z
  Generation:          1
  Resource Version:    160710
  UID:                 a3859b47-575c-40a5-84d4-38fb9f37a8ef
Spec:
  Apply:  IfReady
  Configuration:
    Secret Name:  new-custom-config
  Database Ref:
    Name:   ms-standalone
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-11-07T12:46:35Z
    Message:               MSSQLServer ops-request has started to reconfigure MSSQLServer nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-11-07T12:46:49Z
    Message:               successfully reconciled the mssqlserver with new configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-07T12:46:54Z
    Message:               get pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ms-standalone-0
    Last Transition Time:  2024-11-07T12:46:54Z
    Message:               evict pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ms-standalone-0
    Last Transition Time:  2024-11-07T12:47:34Z
    Message:               check pod running; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ms-standalone-0
    Last Transition Time:  2024-11-07T12:47:39Z
    Message:               Successfully Restarted Pods after reconfiguration
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-07T12:47:39Z
    Message:               Successfully completed reconfiguring for MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           3m45s  KubeDB Ops-manager Operator  Start processing for MSSQLServerOpsRequest: demo/msops-reconfigure-standalone
  Normal   Starting                                                           3m45s  KubeDB Ops-manager Operator  Pausing MSSQLServer database: demo/ms-standalone
  Normal   Successful                                                         3m45s  KubeDB Ops-manager Operator  Successfully paused MSSQLServer database: demo/ms-standalone for MSSQLServerOpsRequest: msops-reconfigure-standalone
  Normal   UpdatePetSets                                                      3m31s  KubeDB Ops-manager Operator  successfully reconciled the mssqlserver with new configuration
  Warning  get pod; ConditionStatus:True; PodName:ms-standalone-0             3m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ms-standalone-0
  Warning  evict pod; ConditionStatus:True; PodName:ms-standalone-0           3m26s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ms-standalone-0
  Warning  check pod running; ConditionStatus:False; PodName:ms-standalone-0  3m21s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:ms-standalone-0
  Warning  check pod running; ConditionStatus:True; PodName:ms-standalone-0   2m46s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ms-standalone-0
  Normal   RestartPods                                                        2m41s  KubeDB Ops-manager Operator  Successfully Restarted Pods after reconfiguration
  Normal   Starting                                                           2m41s  KubeDB Ops-manager Operator  Resuming MSSQLServer database: demo/ms-standalone
  Normal   Successful                                                         2m41s  KubeDB Ops-manager Operator  Successfully resumed MSSQLServer database: demo/ms-standalone for MSSQLServerOpsRequest: msops-reconfigure-standalone
```

Now let's connect to SQL Server instance and run a internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo ms-standalone-0 -c mssql -- bash
mssql@ms-standalone-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
[memory]
memorylimitmb = 2560
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P SERtEyH1RMMEsvE0
1> SELECT physical_memory_kb / 1024 AS physical_memory_mb FROM sys.dm_os_sys_info;
2> go
physical_memory_mb  
--------------------
                2560

(1 rows affected)
1> 
```

As we can see from the configuration of running SQL Server, the value of `physical_memory_mb` has been changed from `2048` to `2560`. So the reconfiguration of the database is successful.


### Reconfigure using apply config

Now we will reconfigure this database again to set `memorylimitmb` to `3072`. This time we won't use a new secret. We will use the `applyConfig` field of the `MSSQLServerOpsRequest`. This will merge the new config in the existing secret.

#### Create MSSQLServerOpsRequest

Now, we will use the new configuration in the `applyConfig` field in the `MSSQLServerOpsRequest` CR. The `MSSQLServerOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-reconfigure-standalone-apply
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ms-standalone
  configuration:
    applyConfig:
      mssql.conf: |-
        [memory]
        memorylimitmb = 3072
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `ms-standalone` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure/msops-reconfigure-standalone-apply.yaml
MSSQLServeropsrequest.ops.kubedb.com/msops-reconfigure-standalone-apply created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CR,

```bash
$ watch kubectl get MSSQLServeropsrequest -n demo
Every 2.0s: kubectl get MSSQLServeropsrequest -n demo

NAME                                 TYPE          STATUS       AGE
msops-reconfigure-standalone-apply   Reconfigure   Successful   2m2s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe MSSQLServeropsrequest -n demo msops-reconfigure-standalone-apply
Name:         msops-reconfigure-standalone-apply
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-07T12:57:43Z
  Generation:          1
  Resource Version:    161738
  UID:                 e9da170c-5d6b-41fd-ae05-b39c6c5f3029
Spec:
  Apply:  IfReady
  Configuration:
    Apply Config:
      mssql.conf:  [memory]
memorylimitmb = 3072
  Database Ref:
    Name:   ms-standalone
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-11-07T12:57:43Z
    Message:               MSSQLServer ops-request has started to reconfigure MSSQLServer nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-11-07T12:57:46Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2024-11-07T12:57:51Z
    Message:               successfully reconciled the mssqlserver with new configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-07T12:57:56Z
    Message:               get pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ms-standalone-0
    Last Transition Time:  2024-11-07T12:57:56Z
    Message:               evict pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ms-standalone-0
    Last Transition Time:  2024-11-07T12:58:31Z
    Message:               check pod running; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ms-standalone-0
    Last Transition Time:  2024-11-07T12:58:36Z
    Message:               Successfully Restarted Pods after reconfiguration
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-07T12:58:36Z
    Message:               Successfully completed reconfiguring for MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           2m40s  KubeDB Ops-manager Operator  Start processing for MSSQLServerOpsRequest: demo/msops-reconfigure-standalone-apply
  Normal   Starting                                                           2m40s  KubeDB Ops-manager Operator  Pausing MSSQLServer database: demo/ms-standalone
  Normal   Successful                                                         2m40s  KubeDB Ops-manager Operator  Successfully paused MSSQLServer database: demo/ms-standalone for MSSQLServerOpsRequest: msops-reconfigure-standalone-apply
  Normal   UpdatePetSets                                                      2m32s  KubeDB Ops-manager Operator  successfully reconciled the mssqlserver with new configuration
  Warning  get pod; ConditionStatus:True; PodName:ms-standalone-0             2m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ms-standalone-0
  Warning  evict pod; ConditionStatus:True; PodName:ms-standalone-0           2m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ms-standalone-0
  Warning  check pod running; ConditionStatus:False; PodName:ms-standalone-0  2m22s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:ms-standalone-0
  Warning  check pod running; ConditionStatus:True; PodName:ms-standalone-0   112s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ms-standalone-0
  Normal   RestartPods                                                        107s   KubeDB Ops-manager Operator  Successfully Restarted Pods after reconfiguration
  Normal   Starting                                                           107s   KubeDB Ops-manager Operator  Resuming MSSQLServer database: demo/ms-standalone
  Normal   Successful                                                         107s   KubeDB Ops-manager Operator  Successfully resumed MSSQLServer database: demo/ms-standalone for MSSQLServerOpsRequest: msops-reconfigure-standalone-apply
```

Now let's connect to the SQL Server instance and run a internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo ms-standalone-0 -c mssql -- bash
mssql@ms-standalone-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
[memory]
memorylimitmb = 3072
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P SERtEyH1RMMEsvE0
1> SELECT physical_memory_kb / 1024 AS physical_memory_mb FROM sys.dm_os_sys_info;
2> go
physical_memory_mb  
--------------------
                3072

(1 rows affected)
1> 
```

As we can see from the configuration of running SQL Server, the value of `physical_memory_mb` has been changed from `2560` to `3072`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ms -n demo ms-standalone
kubectl delete MSSQLServeropsrequest -n demo mops-reconfigure-standalone msops-reconfigure-standalone-apply
```