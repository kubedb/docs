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
2. **User Defined:** The user can create their own credentials by defining a secret of type
   `kubernetes.io/basic-auth` containing the desired `password` and then reference this secret in the
   `IgniteOpsRequest` CR.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.Ignite=true` to ensure Ignite CRD installation.

- You should be familiar with the following `KubeDB` concepts:
    - [Ignite](/docs/guides/ignite/concepts/ignite.md)


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
```shell
$ kubectl get Ignite -n demo -w
NAME                TYPE                  VERSION   STATUS   AGE
ignite-quickstart   kubedb.com/v1alpha2   2.17.0    Ready    4m42s
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the secret.

````shell
$ kubectl get ignite -n demo ignite-quickstart -ojson | jq .spec.authSecret.name
"ignite-quickstart-auth"
$ kubectl get secret -n demo ignite-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
ignite⏎                      
$ kubectl get secret -n demo ignite-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
t5Z506kh2xOPbcZV⏎                     
````
Now, you can exec into the pod `ignite-quickstart-0` and connect to database using `username` and `password`
```bash
$  kubectl exec -it -n demo ignite-quickstart-0 -c ignite -- bash
[ignite@ignite-quickstart-0 ignite]$ apache-ignite/bin/sqlline.sh -u jdbc:ignite:thin://127.0.0.1/ -n ignite -p 't5Z506kh2xOPbcZV'
WARNING: Unknown module: jdk.internal.jvmstat specified to --add-exports
WARNING: An illegal reflective access operation has occurred
WARNING: Illegal reflective access by org.apache.ignite.internal.util.GridUnsafe$2 (file:/opt/ignite/apache-ignite/libs/ignite-core-2.17.0.jar) to field java.nio.Buffer.address
WARNING: Please consider reporting this to the maintainers of org.apache.ignite.internal.util.GridUnsafe$2
WARNING: Use --illegal-access=warn to enable warnings of further illegal reflective access operations
WARNING: All illegal access operations will be denied in a future release
Transaction isolation level TRANSACTION_REPEATABLE_READ is not supported. Default (TRANSACTION_NONE) will be used instead.
sqlline version 1.9.0

0: jdbc:ignite:thin://127.0.0.1/> CREATE TABLE City (id LONG PRIMARY KEY, name VARCHAR);
No rows affected (0.1 seconds)
0: jdbc:ignite:thin://127.0.0.1/> INSERT INTO City (id, name) VALUES (1, 'Forest Hill');
1 row affected (0.043 seconds)
0: jdbc:ignite:thin://127.0.0.1/> INSERT INTO City (id, name) VALUES (2, 'Denver');
1 row affected (0.003 seconds)
0: jdbc:ignite:thin://127.0.0.1/>  SELECT * FROM City;
+----+-------------+
| ID |    NAME     |
+----+-------------+
| 1  | Forest Hill |
| 2  | Denver      |
+----+-------------+
2 rows selected (0.048 seconds)

```

If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth IgniteOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Ignite using operator generated, we have to create a `IgniteOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `IgniteOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: ignite-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `ignite-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Ignite.

Let's create the `IgniteOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/ignite/rotate-auth/rotate-auth-generated.yaml
 IgniteOpsRequest.ops.kubedb.com/igops-rotate-auth-generated created
```
Let's wait for `IgniteOpsRequest` to be `Successful`. Run the following command to watch `IgniteOpsRequest` CRO
```shell
 $ kubectl get IgniteOpsRequest -n demo
 NAME                          TYPE         STATUS       AGE
igops-rotate-auth-generated   RotateAuth   Successful    66s

```
If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe IgniteOpsRequest -n demo igops-rotate-auth-generated
Name:         igops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2025-08-26T08:51:53Z
  Generation:          1
  Resource Version:    460277
  UID:                 5f94d6ad-5486-4c33-b0bc-217116b5268d
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   ignite-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-26T08:51:53Z
    Message:               Ignite ops-request has started to rotate auth for Ignite nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-26T08:51:57Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-26T08:52:04Z
    Message:               successfully reconciled the Ignite with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-26T08:52:59Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-26T08:52:09Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-0
    Last Transition Time:  2025-08-26T08:52:09Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-0
    Last Transition Time:  2025-08-26T08:52:14Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-26T08:52:54Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-0
    Last Transition Time:  2025-08-26T08:52:59Z
    Message:               Successfully completed reconfigure Ignite
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age    From                         Message
  ----     ------                                                          ----   ----                         -------
  Normal   Starting                                                        9m10s  KubeDB Ops-manager Operator  Start processing for IgniteOpsRequest: demo/igops-rotate-auth-generated
  Normal   Starting                                                        9m10s  KubeDB Ops-manager Operator  Pausing Ignite databse: demo/ignite-quickstart
  Normal   Successful                                                      9m10s  KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-generated
  Normal   UpdatePetSets                                                   8m59s  KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated version
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-0      8m54s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0    8m54s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  running pod; ConditionStatus:False                              8m49s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-0  8m9s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Normal   RestartNodes                                                    8m4s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                        8m4s   KubeDB Ops-manager Operator  Resuming Ignite database: demo/ignite-quickstart
  Normal   Successful                                                      8m4s   KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-generated

```
**Verify Auth is rotated**
```shell
$ kubectl get ignite -n demo ignite-quickstart -ojson | jq .spec.authSecret.name
"ignite-quickstart-auth"
$ kubectl get secret -n demo ignite-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
ignite⏎    
$ kubectl get secret -n demo ignite-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
0wFNGcZK54y9OnGX⏎                                        
```
Let's verify if we can connect to the database using the new credentials.

```shell
$ kubectl exec -it -n demo ignite-quickstart-0 -c ignite -- bash
[ignite@ignite-quickstart-0 ignite]$ apache-ignite/bin/sqlline.sh -u jdbc:ignite:thin://127.0.0.1/ -n ignite -p '0wFNGcZK54y9OnGX'
WARNING: Unknown module: jdk.internal.jvmstat specified to --add-exports
WARNING: An illegal reflective access operation has occurred
WARNING: Illegal reflective access by org.apache.ignite.internal.util.GridUnsafe$2 (file:/opt/ignite/apache-ignite/libs/ignite-core-2.17.0.jar) to field java.nio.Buffer.address
WARNING: Please consider reporting this to the maintainers of org.apache.ignite.internal.util.GridUnsafe$2
WARNING: Use --illegal-access=warn to enable warnings of further illegal reflective access operations
WARNING: All illegal access operations will be denied in a future release
Transaction isolation level TRANSACTION_REPEATABLE_READ is not supported. Default (TRANSACTION_NONE) will be used instead.
sqlline version 1.9.0
0: jdbc:ignite:thin://127.0.0.1/>  SELECT * FROM City;
+----+-------------+
| ID |    NAME     |
+----+-------------+
| 1  | Forest Hill |
| 2  | Denver      |
+----+-------------+
2 rows selected (0.142 seconds)

```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo ignite-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
ignite⏎                                                    
$ kubectl get secret -n demo ignite-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
9ycCSYznZpZRxs9U⏎                              
```
Let's confirm that the previous credentials no longer work.
```shell
kubectl exec -it -n demo ignite-quickstart-0 -c mssql -- bash
mssql@ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "9ycCSYznZpZRxs9U"
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Login failed for user 'sa'..
mssql@ignite-quickstart-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,
> Note: The `username` must be fixed as `sa`. The `password` must include uppercase letters, lowercase letters, and numbers
```shell
$ kubectl create secret generic ignite-quickstart-auth-user -n demo \
                                               --type=kubernetes.io/basic-auth \
                                               --from-literal=username=ignite \
                                               --from-literal=password=Ignite2
secret/ignite-quickstart-auth-user created

```
Now create a `IgniteOpsRequest` with `RotateAuth` type. Below is the YAML of the `IgniteOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: ignite-quickstart
  authentication:
    secretRef:
      name: ignite-quickstart-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `ignite-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Ignite.
- `spec.authentication.secretRef.name` specifies that we want to use `ignite-quickstart-auth` for database authentication.


Let's create the `IgniteOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/ignite/rotate-auth/rotate-auth-user.yaml
IgniteOpsRequest.ops.kubedb.com/igops-rotate-auth-user created
```
Let’s wait for `IgniteOpsRequest` to be Successful. Run the following command to watch `IgniteOpsRequest` CRO:

```shell
$ kubectl get IgniteOpsRequest -n demo
NAME                          TYPE         STATUS       AGE
igops-rotate-auth-generated   RotateAuth   Successful   19h
igops-rotate-auth-user        RotateAuth   Successful   7m44s
```
We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed.
```shell
$  kubectl describe IgniteOpsRequest -n demo igops-rotate-auth-user
Name:         igops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2025-09-08T11:39:51Z
  Generation:          1
  Resource Version:    212371
  UID:                 a5e5b870-cccd-40a5-8eb8-37bec466eb57
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  ignite-quickstart-auth-user
  Database Ref:
    Name:   ignite-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-09-08T11:39:51Z
    Message:               Ignite ops-request has started to rotate auth for Ignite nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-09-08T11:39:55Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-09-08T11:40:07Z
    Message:               successfully reconciled the Ignite with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-09-08T11:41:02Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-09-08T11:40:12Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-0
    Last Transition Time:  2025-09-08T11:40:12Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-0
    Last Transition Time:  2025-09-08T11:40:17Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-09-08T11:40:57Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-0
    Last Transition Time:  2025-09-08T11:41:02Z
    Message:               Successfully completed reconfigure Ignite
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age    From                         Message
  ----     ------                                                          ----   ----                         -------
  Normal   Starting                                                        5m31s  KubeDB Ops-manager Operator  Start processing for IgniteOpsRequest: demo/igops-rotate-auth-user
  Normal   Starting                                                        5m30s  KubeDB Ops-manager Operator  Pausing Ignite databse: demo/ignite-quickstart
  Normal   Successful                                                      5m30s  KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-user
  Normal   UpdatePetSets                                                   5m15s  KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated version
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-0      5m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0    5m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  running pod; ConditionStatus:False                              5m5s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-0  4m25s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Normal   RestartNodes                                                    4m20s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                        4m20s  KubeDB Ops-manager Operator  Resuming Ignite database: demo/ignite-quickstart
  Normal   Successful                                                      4m20s  KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-user

```
**Verify auth is rotate**
```shell
$ kubectl get ignite -n demo ignite-quickstart -ojson | jq .spec.authSecret.name
"ignite-quickstart-auth-user"
$ kubectl get secret -n demo ignite-quickstart-auth-user -o jsonpath='{.data.username}' | base64 -d
ignite⏎                                                               
$ kubectl get secret -n demo ignite-quickstart-auth-user -o jsonpath='{.data.password}' | base64 -d
Ignite2⏎                                                                                       
```

Let's verify if we can connect to the database using the new credentials.
```shell
$ kubectl exec -it -n demo ignite-quickstart-0 -c mssql -- bash
mssql@ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "Ignite2"
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
$ kubectl get secret -n demo ignite-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
sa⏎                                                                                         
$ kubectl get secret -n demo quick-Ignite-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
zTBVvzgoEb2qUe3X⏎ 
```
Let's confirm that the previous credentials no longer work.
```shell
kubectl exec -it -n demo ignite-quickstart-0 -c mssql -- bash
mssql@ignite-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "zTBVvzgoEb2qUe3X"
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Login failed for user 'sa'..
mssql@ignite-quickstart-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete IgniteOpsRequest igops-rotate-auth-generated igops-rotate-auth-user -n demo
IgniteOpsRequest.ops.kubedb.com "igops-rotate-auth-generated" "igops-rotate-auth-user" deleted
$ kubectl delete secret -n demoquick-Ignite-user-auth
secret "quick-Ignite-user-auth" deleted
$ kubectl delete secret -n demo   ignite-quickstart-auth 
secret "ignite-quickstart-auth " deleted

```

## Next Steps

- Learn about [backup and restore](/docs/guides/Ignite/backup/overview/index.md) SQL Server using KubeStash.
- Want to set up SQL Server Availability Group clusters? Check how to [Configure SQL Server Availability Gruop Cluster](/docs/guides/Ignite/clustering/ag_cluster.md)
- Detail concepts of [Ignite object](/docs/guides/ignite/concepts/Ignite.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
