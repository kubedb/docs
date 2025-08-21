---
title: Ignite Rotateauth Guide
menu:
docs_{{ .version }}:
identifier: ig-rotate-auth-guide
name: Guide
parent: ig-quickstart-ignite
weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of Ignite

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Ignite` user's authentication credentials using a `IgniteOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a Secret of type
   `kubernetes.io/basic-auth` containing the desired `password`, and then reference this Secret in the
   `IgniteOpsRequest` CR.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.Ignite=true` to ensure Ignite CRD installation.

- You should be familiar with the following `KubeDB` concepts:
    - [Ignite](/docs/guides/ignite/concepts/ignite.md)
    - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Ignite](/docs/examples/ignite/rotate-auth) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Find Available IgniteVersion

When you have installed KubeDB, it has created `IgniteVersion` crd for all supported Ignite versions. Check 0

```bash
$ kubectl get igniteversions
NAME        VERSION    DB_IMAGE                                            DEPRECATED   AGE
2.17.0      2.17.0     ghcr.io/appscode-images/ignite:2.17.0                            2h
```

## Create a Ignite server

KubeDB implements a `Ignite` CRD to define the specification of a Ignite server. Below is the `Ignite` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-quickstart
  namespace: demo
spec:
  replicas: 3
  version: 2.17.0
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/quickstart/demo.yaml
ignite.kubedb.com/ignite-quickstart created
```

## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get ms -n demo Ignite-quickstart -ojson | jq .spec.authSecret.name
"Ignite-quickstart-auth"
$ kubectl get secret -n demo Ignite-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
sa⏎                            
$ kubectl get secret -n demo Ignite-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
9ycCSYznZpZRxs9U⏎             
````
Now, you can exec into the pod `Ignite-quickstart-0` and connect to database using `username` and `password`
```bash
$ kubectl exec -it -n demo Ignite-quickstart-0 -c mssql -- bash
mssql@Ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "9ycCSYznZpZRxs9U"
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
1> exit
mssql@Ignite-quickstart-0:/$ exit
exit
⏎   
```

If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth IgniteOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Ignite using operator generated, we have to create a `IgniteOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `IgniteOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: msops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: Ignite-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `Ignite-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Ignite.

Let's create the `IgniteOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/ignite/rotate-auth/rotate-auth-generated.yaml
 Igniteopsrequest.ops.kubedb.com/msops-rotate-auth-generated created
```
Let's wait for `IgniteOpsrequest` to be `Successful`. Run the following command to watch `IgniteOpsrequest` CRO
```shell
 $ kubectl get Igniteopsrequest -n demo
 NAME                          TYPE         STATUS       AGE
 msops-rotate-auth-generated   RotateAuth   Successful   7m47s
```
If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Igniteopsrequest -n demo msops-rotate-auth-generated
Name:         msops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2025-07-15T10:59:19Z
  Generation:          1
  Resource Version:    748399
  UID:                 05d2adb0-9adb-497c-95ce-b278ca41dd70
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   Ignite-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-15T10:59:19Z
    Message:               Ignite ops-request has started to rotate auth for Ignite nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-15T10:59:22Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-07-15T10:59:22Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-15T11:00:29Z
    Message:               successfully reconciled the Ignite with updated credentials
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-15T11:01:15Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-15T11:00:35Z
    Message:               get pod; ConditionStatus:True; PodName:Ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--Ignite-quickstart-0
    Last Transition Time:  2025-07-15T11:00:35Z
    Message:               evict pod; ConditionStatus:True; PodName:Ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--Ignite-quickstart-0
    Last Transition Time:  2025-07-15T11:01:10Z
    Message:               check pod running; ConditionStatus:True; PodName:Ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--Ignite-quickstart-0
    Last Transition Time:  2025-07-15T11:01:15Z
    Message:               Successfully completed Ignite auth rotate
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                      Age   From                         Message
  ----     ------                                                                      ----  ----                         -------
  Normal   Starting                                                                    19m   KubeDB Ops-manager Operator  Start processing for IgniteOpsRequest: demo/msops-rotate-auth-generated
  Normal   Starting                                                                    19m   KubeDB Ops-manager Operator  Pausing Ignite database: demo/Ignite-quickstart
  Normal   Successful                                                                  19m   KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/Ignite-quickstart for IgniteOpsRequest: msops-rotate-auth-generated
  Normal   UpdatePetSets                                                               18m   KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated credentials
  Warning  get pod; ConditionStatus:True; PodName:Ignite-quickstart-0             18m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:Ignite-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:Ignite-quickstart-0           18m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:Ignite-quickstart-0
  Warning  check pod running; ConditionStatus:False; PodName:Ignite-quickstart-0  18m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:Ignite-quickstart-0
  Warning  check pod running; ConditionStatus:True; PodName:Ignite-quickstart-0   17m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:Ignite-quickstart-0
  Normal   RestartNodes                                                                17m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                    17m   KubeDB Ops-manager Operator  Resuming Ignite database: demo/Ignite-quickstart
  Normal   Successful                                                                  17m   KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/Ignite-quickstart for IgniteOpsRequest: msops-rotate-auth-generated
  Normal   UpdatePetSets                                                               17m   KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated credentials

```
**Verify Auth is rotated**
```shell
$ kubectl get ms -n demo Ignite-quickstart -ojson | jq .spec.authSecret.name
"Ignite-quickstart-auth"
$ kubectl get secret -n demo Ignite-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
sa⏎  
$ kubectl get secret -n demo Ignite-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
zTBVvzgoEb2qUe3X⏎                          
```
Let's verify if we can connect to the database using the new credentials.

```shell
$ kubectl exec -it -n demo Ignite-quickstart-0 -c mssql -- bash
mssql@Ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "zTBVvzgoEb2qUe3X"
1> select name from sys.databases
2> go
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
kubedb_system                                                                                                                   

(5 rows affected)
1> exit
mssql@Ignite-quickstart-0:/$ exit
exit
⏎   
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo Ignite-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
sa⏎                                                    
$ kubectl get secret -n demo Ignite-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
9ycCSYznZpZRxs9U⏎                              
```
Let's confirm that the previous credentials no longer work.
```shell
kubectl exec -it -n demo Ignite-quickstart-0 -c mssql -- bash
mssql@Ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "9ycCSYznZpZRxs9U"
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Login failed for user 'sa'..
mssql@Ignite-quickstart-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,
> Note: The `username` must be fixed as `sa`. The `password` must include uppercase letters, lowercase letters, and numbers
```shell
$ kubectl create secret generic Ignite-quickstart-auth-user -n demo \
                                               --type=kubernetes.io/basic-auth \
                                               --from-literal=username=sa \
                                               --from-literal=password=Ignite2
secret/Ignite-quickstart-auth-user created

```
Now create a `IgniteOpsRequest` with `RotateAuth` type. Below is the YAML of the `IgniteOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: msops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: Ignite-quickstart
  authentication:
    secretRef:
      name: Ignite-quickstart-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `Ignite-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Ignite.
- `spec.authentication.secretRef.name` specifies that we want to use `Ignite-quickstart-auth` for database authentication.


Let's create the `IgniteOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/ignite/rotate-auth/rotate-auth-user.yaml
Igniteopsrequest.ops.kubedb.com/msops-rotate-auth-user created
```
Let’s wait for `IgniteOpsRequest` to be Successful. Run the following command to watch `IgniteOpsRequest` CRO:

```shell
$ kubectl get Igniteopsrequest -n demo
NAME                          TYPE         STATUS       AGE
msops-rotate-auth-generated   RotateAuth   Successful   19h
msops-rotate-auth-user        RotateAuth   Successful   7m44s
```
We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Igniteopsrequest -n demo msops-rotate-auth-user 
Name:         msops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2025-07-15T11:03:16Z
  Generation:          1
  Resource Version:    748850
  UID:                 9928916d-ba08-4d3a-a5a9-afa3cea4ee90
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  Ignite-quickstart-auth
  Database Ref:
    Name:   Ignite-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-15T11:03:16Z
    Message:               Ignite ops-request has started to rotate auth for Ignite nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-15T11:03:19Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-07-15T11:03:19Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-15T11:04:26Z
    Message:               successfully reconciled the Ignite with updated credentials
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-15T11:05:11Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-15T11:04:31Z
    Message:               get pod; ConditionStatus:True; PodName:Ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--Ignite-quickstart-0
    Last Transition Time:  2025-07-15T11:04:31Z
    Message:               evict pod; ConditionStatus:True; PodName:Ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--Ignite-quickstart-0
    Last Transition Time:  2025-07-15T11:05:06Z
    Message:               check pod running; ConditionStatus:True; PodName:Ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--Ignite-quickstart-0
    Last Transition Time:  2025-07-15T11:05:11Z
    Message:               Successfully completed Ignite auth rotate
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                      Age   From                         Message
  ----     ------                                                                      ----  ----                         -------
  Normal   Starting                                                                    16m   KubeDB Ops-manager Operator  Start processing for IgniteOpsRequest: demo/msops-rotate-auth-user
  Normal   Starting                                                                    16m   KubeDB Ops-manager Operator  Pausing Ignite database: demo/Ignite-quickstart
  Normal   Successful                                                                  16m   KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/Ignite-quickstart for IgniteOpsRequest: msops-rotate-auth-user
  Normal   UpdatePetSets                                                               15m   KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated credentials
  Warning  get pod; ConditionStatus:True; PodName:Ignite-quickstart-0             15m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:Ignite-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:Ignite-quickstart-0           15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:Ignite-quickstart-0
  Warning  check pod running; ConditionStatus:False; PodName:Ignite-quickstart-0  15m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:Ignite-quickstart-0
  Warning  check pod running; ConditionStatus:True; PodName:Ignite-quickstart-0   14m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:Ignite-quickstart-0
  Normal   RestartNodes                                                                14m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                    14m   KubeDB Ops-manager Operator  Resuming Ignite database: demo/Ignite-quickstart
  Normal   Successful                                                                  14m   KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/Ignite-quickstart for IgniteOpsRequest: msops-rotate-auth-user

```
**Verify auth is rotate**
```shell
$ kubectl get ms -n demo Ignite-quickstart -ojson | jq .spec.authSecret.name
"Ignite-quickstart-auth "
$ kubectl get secret -n demo Ignite-quickstart-auth -o=jsonpath='{.data.username}' | base64 -d
sa⏎                                      
$ kubectl get secret -n demo Ignite-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
Ignite2⏎                                                                  
```

Let's verify if we can connect to the database using the new credentials.
```shell
$ kubectl exec -it -n demo Ignite-quickstart-0 -c mssql -- bash
mssql@Ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "Ignite2"
1> SELECT name FROM sys.databases 
2> go
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
kubedb_system                                                                                                                   

(5 rows affected)
1> 
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo Ignite-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
sa⏎                                                                                         
$ kubectl get secret -n demo quick-Ignite-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
zTBVvzgoEb2qUe3X⏎ 
```
Let's confirm that the previous credentials no longer work.
```shell
kubectl exec -it -n demo Ignite-quickstart-0 -c mssql -- bash
mssql@Ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "zTBVvzgoEb2qUe3X"
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Login failed for user 'sa'..
mssql@Ignite-quickstart-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Igniteopsrequest msops-rotate-auth-generated msops-rotate-auth-user -n demo
Igniteopsrequest.ops.kubedb.com "msops-rotate-auth-generated" "msops-rotate-auth-user" deleted
$ kubectl delete secret -n demoquick-Ignite-user-auth
secret "quick-Ignite-user-auth" deleted
$ kubectl delete secret -n demo   Ignite-quickstart-auth 
secret "Ignite-quickstart-auth " deleted

```

## Next Steps

- Learn about [backup and restore](/docs/guides/Ignite/backup/overview/index.md) SQL Server using KubeStash.
- Want to set up SQL Server Availability Group clusters? Check how to [Configure SQL Server Availability Gruop Cluster](/docs/guides/Ignite/clustering/ag_cluster.md)
- Detail concepts of [Ignite object](/docs/guides/ignite/concepts/Ignite.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
