---
title: mysql Rotateauth Guide
menu:
docs_{{ .version }}:
identifier: ig-rotateauth-guide
name: Guide
parent: ig-quickstart-mysql
weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of mysql

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `mysql` user's authentication credentials using a `mysqlOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a Secret of type
   `kubernetes.io/basic-auth` containing the desired `password`, and then reference this Secret in the
   `mysqlOpsRequest` CR.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.mysql=true` to ensure mysql CRD installation.

- You should be familiar with the following `KubeDB` concepts:
    - [mysql](/docs/guides/mysql/concepts/mysql.md)
    - [mysqlOpsRequest](/docs/guides/mysql/concepts/opsrequest.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mysql](/docs/examples/mysql/rotate-auth) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Find Available mysqlVersion

When you have installed KubeDB, it has created `mysqlVersion` crd for all supported mysql versions. Check 0

```bash
$ kubectl get mysqlversions
NAME        VERSION    DB_IMAGE                                            DEPRECATED   AGE
2.17.0      2.17.0     ghcr.io/appscode-images/mysql:2.17.0                            2h
```

## Create a mysql server

KubeDB implements a `mysql` CRD to define the specification of a mysql server. Below is the `mysql` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: mysql
metadata:
  name: mysql-quickstart
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/quickstart/demo.yaml
mysql.kubedb.com/mysql-quickstart created
```

## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get mysql -n demo mysql-quickstart -ojson | jq .spec.authSecret.name
"mysql-quickstart-auth"
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
mysql⏎                 
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
Z6WsA2kkHCUUPJua⏎                         
````
Now, you can exec into the pod `mysql-quickstart-0` and connect to database using `username` and `password`
```bash
$ kubectl exec -it -n demo mysql-quickstart-0 -c mssql -- bash
mssql@mysql-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "9ycCSYznZpZRxs9U"
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
mssql@mysql-quickstart-0:/$ exit
exit
⏎   
```

If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth mysqlOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the mysql using operator generated, we have to create a `mysqlOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `mysqlOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: mysqlOpsRequest
metadata:
  name: msops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mysql-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `mysql-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on mysql.

Let's create the `mysqlOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/mysql/rotate-auth/rotate-auth-generated.yaml
 mysqlopsrequest.ops.kubedb.com/msops-rotate-auth-generated created
```
Let's wait for `mysqlOpsrequest` to be `Successful`. Run the following command to watch `mysqlOpsrequest` CRO
```shell
 $ kubectl get mysqlopsrequest -n demo
 NAME                          TYPE         STATUS       AGE
 msops-rotate-auth-generated   RotateAuth   Successful   7m47s
```
If we describe the `mysqlOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe mysqlopsrequest -n demo msops-rotate-auth-generated
Name:         msops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         mysqlOpsRequest
Metadata:
  Creation Timestamp:  2025-07-15T10:59:19Z
  Generation:          1
  Resource Version:    748399
  UID:                 05d2adb0-9adb-497c-95ce-b278ca41dd70
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mysql-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-15T10:59:19Z
    Message:               mysql ops-request has started to rotate auth for mysql nodes
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
    Message:               successfully reconciled the mysql with updated credentials
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
    Message:               get pod; ConditionStatus:True; PodName:mysql-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mysql-quickstart-0
    Last Transition Time:  2025-07-15T11:00:35Z
    Message:               evict pod; ConditionStatus:True; PodName:mysql-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mysql-quickstart-0
    Last Transition Time:  2025-07-15T11:01:10Z
    Message:               check pod running; ConditionStatus:True; PodName:mysql-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mysql-quickstart-0
    Last Transition Time:  2025-07-15T11:01:15Z
    Message:               Successfully completed mysql auth rotate
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                      Age   From                         Message
  ----     ------                                                                      ----  ----                         -------
  Normal   Starting                                                                    19m   KubeDB Ops-manager Operator  Start processing for mysqlOpsRequest: demo/msops-rotate-auth-generated
  Normal   Starting                                                                    19m   KubeDB Ops-manager Operator  Pausing mysql database: demo/mysql-quickstart
  Normal   Successful                                                                  19m   KubeDB Ops-manager Operator  Successfully paused mysql database: demo/mysql-quickstart for mysqlOpsRequest: msops-rotate-auth-generated
  Normal   UpdatePetSets                                                               18m   KubeDB Ops-manager Operator  successfully reconciled the mysql with updated credentials
  Warning  get pod; ConditionStatus:True; PodName:mysql-quickstart-0             18m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mysql-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:mysql-quickstart-0           18m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mysql-quickstart-0
  Warning  check pod running; ConditionStatus:False; PodName:mysql-quickstart-0  18m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mysql-quickstart-0
  Warning  check pod running; ConditionStatus:True; PodName:mysql-quickstart-0   17m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mysql-quickstart-0
  Normal   RestartNodes                                                                17m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                    17m   KubeDB Ops-manager Operator  Resuming mysql database: demo/mysql-quickstart
  Normal   Successful                                                                  17m   KubeDB Ops-manager Operator  Successfully resumed mysql database: demo/mysql-quickstart for mysqlOpsRequest: msops-rotate-auth-generated
  Normal   UpdatePetSets                                                               17m   KubeDB Ops-manager Operator  successfully reconciled the mysql with updated credentials

```
**Verify Auth is rotated**
```shell
$ kubectl get ms -n demo mysql-quickstart -ojson | jq .spec.authSecret.name
"mysql-quickstart-auth"
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
sa⏎  
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
zTBVvzgoEb2qUe3X⏎                          
```
Let's verify if we can connect to the database using the new credentials.

```shell
$ kubectl exec -it -n demo mysql-quickstart-0 -c mssql -- bash
mssql@mysql-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "zTBVvzgoEb2qUe3X"
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
mssql@mysql-quickstart-0:/$ exit
exit
⏎   
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo mysql-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
sa⏎                                                    
$ kubectl get secret -n demo mysql-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
9ycCSYznZpZRxs9U⏎                              
```
Let's confirm that the previous credentials no longer work.
```shell
kubectl exec -it -n demo mysql-quickstart-0 -c mssql -- bash
mssql@mysql-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "9ycCSYznZpZRxs9U"
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Login failed for user 'sa'..
mssql@mysql-quickstart-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,
> Note: The `username` must be fixed as `sa`. The `password` must include uppercase letters, lowercase letters, and numbers
```shell
$ kubectl create secret generic mysql-quickstart-auth-user -n demo \
                                               --type=kubernetes.io/basic-auth \
                                               --from-literal=username=sa \
                                               --from-literal=password=mysql2
secret/mysql-quickstart-auth-user created

```
Now create a `mysqlOpsRequest` with `RotateAuth` type. Below is the YAML of the `mysqlOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: mysqlOpsRequest
metadata:
  name: msops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mysql-quickstart
  authentication:
    secretRef:
      name: mysql-quickstart-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `mysql-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on mysql.
- `spec.authentication.secretRef.name` specifies that we want to use `mysql-quickstart-auth` for database authentication.


Let's create the `mysqlOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/mysql/rotate-auth/rotate-auth-user.yaml
mysqlopsrequest.ops.kubedb.com/msops-rotate-auth-user created
```
Let’s wait for `mysqlOpsRequest` to be Successful. Run the following command to watch `mysqlOpsRequest` CRO:

```shell
$ kubectl get mysqlopsrequest -n demo
NAME                          TYPE         STATUS       AGE
msops-rotate-auth-generated   RotateAuth   Successful   19h
msops-rotate-auth-user        RotateAuth   Successful   7m44s
```
We can see from the above output that the `mysqlOpsRequest` has succeeded. If we describe the `mysqlOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe mysqlopsrequest -n demo msops-rotate-auth-user 
Name:         msops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         mysqlOpsRequest
Metadata:
  Creation Timestamp:  2025-07-15T11:03:16Z
  Generation:          1
  Resource Version:    748850
  UID:                 9928916d-ba08-4d3a-a5a9-afa3cea4ee90
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  mysql-quickstart-auth
  Database Ref:
    Name:   mysql-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-15T11:03:16Z
    Message:               mysql ops-request has started to rotate auth for mysql nodes
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
    Message:               successfully reconciled the mysql with updated credentials
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
    Message:               get pod; ConditionStatus:True; PodName:mysql-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mysql-quickstart-0
    Last Transition Time:  2025-07-15T11:04:31Z
    Message:               evict pod; ConditionStatus:True; PodName:mysql-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mysql-quickstart-0
    Last Transition Time:  2025-07-15T11:05:06Z
    Message:               check pod running; ConditionStatus:True; PodName:mysql-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mysql-quickstart-0
    Last Transition Time:  2025-07-15T11:05:11Z
    Message:               Successfully completed mysql auth rotate
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                      Age   From                         Message
  ----     ------                                                                      ----  ----                         -------
  Normal   Starting                                                                    16m   KubeDB Ops-manager Operator  Start processing for mysqlOpsRequest: demo/msops-rotate-auth-user
  Normal   Starting                                                                    16m   KubeDB Ops-manager Operator  Pausing mysql database: demo/mysql-quickstart
  Normal   Successful                                                                  16m   KubeDB Ops-manager Operator  Successfully paused mysql database: demo/mysql-quickstart for mysqlOpsRequest: msops-rotate-auth-user
  Normal   UpdatePetSets                                                               15m   KubeDB Ops-manager Operator  successfully reconciled the mysql with updated credentials
  Warning  get pod; ConditionStatus:True; PodName:mysql-quickstart-0             15m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mysql-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:mysql-quickstart-0           15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mysql-quickstart-0
  Warning  check pod running; ConditionStatus:False; PodName:mysql-quickstart-0  15m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mysql-quickstart-0
  Warning  check pod running; ConditionStatus:True; PodName:mysql-quickstart-0   14m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mysql-quickstart-0
  Normal   RestartNodes                                                                14m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                    14m   KubeDB Ops-manager Operator  Resuming mysql database: demo/mysql-quickstart
  Normal   Successful                                                                  14m   KubeDB Ops-manager Operator  Successfully resumed mysql database: demo/mysql-quickstart for mysqlOpsRequest: msops-rotate-auth-user

```
**Verify auth is rotate**
```shell
$ kubectl get ms -n demo mysql-quickstart -ojson | jq .spec.authSecret.name
"mysql-quickstart-auth "
$ kubectl get secret -n demo mysql-quickstart-auth -o=jsonpath='{.data.username}' | base64 -d
sa⏎                                      
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
mysql2⏎                                                                  
```

Let's verify if we can connect to the database using the new credentials.
```shell
$ kubectl exec -it -n demo mysql-quickstart-0 -c mssql -- bash
mssql@mysql-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "mysql2"
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
$ kubectl get secret -n demo mysql-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
sa⏎                                                                                         
$ kubectl get secret -n demo quick-mysql-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
zTBVvzgoEb2qUe3X⏎ 
```
Let's confirm that the previous credentials no longer work.
```shell
kubectl exec -it -n demo mysql-quickstart-0 -c mssql -- bash
mssql@mysql-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "zTBVvzgoEb2qUe3X"
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Login failed for user 'sa'..
mssql@mysql-quickstart-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete mysqlopsrequest msops-rotate-auth-generated msops-rotate-auth-user -n demo
mysqlopsrequest.ops.kubedb.com "msops-rotate-auth-generated" "msops-rotate-auth-user" deleted
$ kubectl delete secret -n demoquick-mysql-user-auth
secret "quick-mysql-user-auth" deleted
$ kubectl delete secret -n demo   mysql-quickstart-auth 
secret "mysql-quickstart-auth " deleted

```

## Next Steps

- Learn about [backup and restore](/docs/guides/mysql/backup/overview/index.md) SQL Server using KubeStash.
- Want to set up SQL Server Availability Group clusters? Check how to [Configure SQL Server Availability Gruop Cluster](/docs/guides/mysql/clustering/ag_cluster.md)
- Detail concepts of [mysql object](/docs/guides/mysql/concepts/mysql.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
