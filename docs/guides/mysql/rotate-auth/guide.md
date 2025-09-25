---
title: MySQL Rotateauth Guide
menu:
    docs_{{ .version }}:
        identifier: guides-mysql-rotate-auth-guide
        name: Guide
        parent: guides-mysql-rotate-auth
        weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of MySQL

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `MySQL`
user's authentication credentials using a `MySQLOpsRequest`. There are two ways to perform this
rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential and 
updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a secret of type
   `kubernetes.io/basic-auth` containing the desired `password` and then reference this secret in the
   `MySQLOpsRequest` CR.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY     VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete            WaitForFirstConsumer   false                  6h22m
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available MySQLVersion

When you have installed KubeDB, it has created `MySQLVersion` crd for all supported MySQL versions. Check it by using the following command,

```bash
$ kubectl get mysqlversion
NAME            VERSION   DISTRIBUTION   DB_IMAGE                                      DEPRECATED   AGE
5.7.42-debian   5.7.42    Official       ghcr.io/appscode-images/mysql:5.7.42-debian                12d
5.7.44          5.7.44    Official       ghcr.io/appscode-images/mysql:5.7.44-oracle                12d
8.0.31-innodb   8.0.31    MySQL          ghcr.io/appscode-images/mysql:8.0.31-oracle                12d
8.0.35          8.0.35    Official       ghcr.io/appscode-images/mysql:8.0.35-oracle                12d
8.0.36          8.0.36    Official       ghcr.io/appscode-images/mysql:8.0.36-debian                12d
8.1.0           8.1.0     Official       ghcr.io/appscode-images/mysql:8.1.0-oracle                 12d
8.2.0           8.2.0     Official       ghcr.io/appscode-images/mysql:8.2.0-oracle                 12d
8.4.2           8.4.2     Official       ghcr.io/appscode-images/mysql:8.4.2-oracle                 12d
8.4.3           8.4.3     Official       ghcr.io/appscode-images/mysql:8.4.3-oracle                 12d
9.0.1           9.0.1     Official       ghcr.io/appscode-images/mysql:9.0.1-oracle                 12d
9.1.0           9.1.0     Official       ghcr.io/appscode-images/mysql:9.1.0-oracle                 12d
```

## Create a Mysql Database

KubeDB implements a `MySQL` CRD to define the specification of a MySQL server. Below is the `MySQL`
object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-quickstart
  namespace: demo
spec:
  version: "9.1.0"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/quickstart/yamls/quickstart-v1.yaml
mysql.kubedb.com/mysql-quickstart created
```
Let's wait for `MySQL` status is `Ready`. Run the following command to watch `MySQL` CRO
```shell
$ kubectl get mysql -n demo -w
NAME             VERSION   STATUS   AGE
mysql-quickstart 10.5.23   Ready    30m
```
## Verify Authentication
The user can verify whether they are authorized by executing a query directly in the database. To
do this, the user needs `username` and `password` in order to connect to the database. Below is an
example showing how to retrieve the credentials from the secret.

````shell
$ kubectl get mysql -n demo mysql-quickstart -ojson | jq .spec.authSecret.name
"mysql-quickstart-auth"
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
root⏎                                
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
H04(Wn6AM_4r6)(k⏎                                       
````
Now, you can exec into the pod `mysql-quickstart-0` and connect to database using `username` and `password`
```bash
$  kubectl exec -it -n demo mysql-quickstart-0 -c mysql -- bash

bash-5.1$  mysql -uroot -p"H04(Wn6AM_4r6)(k"
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 169
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.01 sec)

```

If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth MySQLOpsRequest

#### 1. Using Operator Generated Credentials:

In order to rotate authentication to the MySQL using operator generated, we have to create a 
`MySQLOpsRequest` CR with `RotateAuth` type. Below is the YAML of the `MySQLOpsRequest` CRO that
we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mysql-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `mysql-quickstart` instance.
- `spec.type` specifies that we are performing `RotateAuth` on MySQL.

Let's create the `MySQLOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/mysql/rotate-auth/rotate-auth-generated.yaml
 MySQLOpsRequest.ops.kubedb.com/myops-rotate-auth-generated created
```
Let's wait for `MySQLOpsRequest` to be `Successful`. Run the following command to watch `MySQLOpsRequest` CRO
```shell
 $ kubectl get MySQLOpsRequest-n demo
NAME                          TYPE         STATUS       AGE
myops-rotate-auth-generated   RotateAuth   Successful   82s
```
If we describe the `MySQLOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe MySQLOpsRequest-n demo myops-rotate-auth-generated
Name:         myops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2025-08-20T09:16:39Z
  Generation:          1
  Resource Version:    69442
  UID:                 fded850f-d5a9-46d0-b082-9f789bb68ac2
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mysql-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-20T09:16:39Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-08-20T09:16:42Z
    Message:               Successfully generated new credential
    Observed Generation:   1
    Reason:                patchedsecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-20T09:16:44Z
    Message:               Successfully updated MySQL petset
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-20T09:16:49Z
    Message:               evict pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod
    Last Transition Time:  2025-08-20T09:16:59Z
    Message:               is pod ready; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady
    Last Transition Time:  2025-08-20T09:16:59Z
    Message:               is join in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsJoinInCluster
    Last Transition Time:  2025-08-20T09:16:59Z
    Message:               Successfully started MySQL pods for MySQLOpsRequest: demo/myops-rotate-auth-generated 
    Observed Generation:   1
    Reason:                RestartPodsSucceeded
    Status:                True
    Type:                  Restart
    Last Transition Time:  2025-08-20T09:16:59Z
    Message:               Controller has successfully rotate MySQL auth secret demo/myops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age   From                         Message
  ----     ------                                    ----  ----                         -------
  Normal   Starting                                  27m   KubeDB Ops-manager Operator  Start processing for MySQLOpsRequest: demo/myops-rotate-auth-generated
  Normal   Starting                                  27m   KubeDB Ops-manager Operator  Pausing MySQL databse: demo/mysql-quickstart
  Normal   Successful                                27m   KubeDB Ops-manager Operator  Successfully paused MySQL database: demo/mysql-quickstart for MySQLOpsRequest: myops-rotate-auth-generated
  Normal   Starting                                  27m   KubeDB Ops-manager Operator  Restarting Pod: mysql-quickstart-0/demo
  Warning  evict pod; ConditionStatus:True           27m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True
  Warning  is pod ready; ConditionStatus:False       27m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False
  Warning  is pod ready; ConditionStatus:True        26m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True
  Warning  is join in cluster; ConditionStatus:True  26m   KubeDB Ops-manager Operator  is join in cluster; ConditionStatus:True
  Normal   Successful                                26m   KubeDB Ops-manager Operator  Successfully started MySQL pods for MySQLOpsRequest: demo/myops-rotate-auth-generated
  Normal   Starting                                  26m   KubeDB Ops-manager Operator  Resuming MySQL database: demo/mysql-quickstart
  Normal   Successful                                26m   KubeDB Ops-manager Operator  Successfully resumed MySQL database: demo/mysql-quickstart
  Normal   Successful                                26m   KubeDB Ops-manager Operator  Controller has successfully rotate MySQL auth secret

```
**Verify Auth is rotated**
```shell
$ kubectl get mysql -n demo mysql-quickstart -ojson | jq .spec.authSecret.name
"mysql-quickstart-auth"
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
root⏎                                    
$ kubectl get secret -n demo mysql-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
vYBjULhCEzPwe5xo⏎                                     
```
Let's verify if we can connect to the database using the new credentials.

```shell
$ kubectl exec -it -n demo mysql-quickstart-0 -c mysql -- bash

bash-5.1$  mysql -uroot -p"vYBjULhCEzPwe5xo"
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 49
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo mysql-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎                                                                                        
$ kubectl get secret -n demo mysql-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
H04(Wn6AM_4r6)(k⏎                                           
```
Let's confirm that the previous credentials no longer work.
```shell
kubectl exec -it -n demo mysql-quickstart-0 -c mysql -- bash
bash-5.1$ mysql -uroot -p"H04(Wn6AM_4r6)(k"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'root'@'localhost' (using password: YES)
bash-5.1$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2.Using User Created Credentials

At first, we need to create a secret with `kubernetes.io/basic-auth` type using custom password.
Below is the command to create a secret with `kubernetes.io/basic-auth` type,
> Note: The `username` must be fixed as `root`. 
```shell
$ kubectl create secret generic mysql-quickstart-auth-user -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=root \
                --from-literal=password=Mysql2
secret/mysql-quickstart-auth-user created
```
Now create a `MySQLOpsRequest` with `RotateAuth` type. Below is the YAML of the `MySQLOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mysql-quickstart
  authentication:
    secretRef:
      kind: Secret
      name: mysql-quickstart-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `mysql-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Mysql.
- `spec.authentication.secretRef.name` specifies that we use `mysql-quickstart-auth` for database authentication.


Let's create the `MySQLOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/mysql/rotate-auth/rotate-auth-user.yaml
MySQLOpsRequest.ops.kubedb.com/myops-rotate-auth-user created
```
Let’s wait for `MySQLOpsRequest` to be Successful. Run the following command to watch `MySQLOpsRequest` CRO:

```shell
$ kubectl get MySQLOpsRequest-n demo
NAME                          TYPE         STATUS       AGE
myops-rotate-auth-generated   RotateAuth   Successful   35m
myops-rotate-auth-user        RotateAuth   Successful   2m18s
```
We can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe MySQLOpsRequest-n demo myops-rotate-auth-user 
Name:         myops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2025-08-20T09:49:56Z
  Generation:          1
  Resource Version:    74129
  UID:                 88dd5aed-cd12-43d7-95d6-e0b7e726076f
Spec:
  Apply:  IfReady
  Authentication:
    secret Ref:
      Name:  mysql-quickstart-auth-user
  Database Ref:
    Name:   mysql-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-20T09:49:56Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-rotate-auth-user
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-08-20T09:49:59Z
    Message:               Successfully referenced the user provided authsecret
    Observed Generation:   1
    Reason:                patchedsecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-20T09:50:01Z
    Message:               Successfully updated MySQL petset
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-20T09:50:06Z
    Message:               evict pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod
    Last Transition Time:  2025-08-20T09:50:11Z
    Message:               is pod ready; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady
    Last Transition Time:  2025-08-20T09:50:41Z
    Message:               is join in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsJoinInCluster
    Last Transition Time:  2025-08-20T09:50:41Z
    Message:               Successfully started MySQL pods for MySQLOpsRequest: demo/myops-rotate-auth-user 
    Observed Generation:   1
    Reason:                RestartPodsSucceeded
    Status:                True
    Type:                  Restart
    Last Transition Time:  2025-08-20T09:50:41Z
    Message:               Controller has successfully rotate MySQL auth secret demo/myops-rotate-auth-user
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age    From                         Message
  ----     ------                                     ----   ----                         -------
  Normal   Starting                                   2m54s  KubeDB Ops-manager Operator  Start processing for MySQLOpsRequest: demo/myops-rotate-auth-user
  Normal   Starting                                   2m54s  KubeDB Ops-manager Operator  Pausing MySQL databse: demo/mysql-quickstart
  Normal   Successful                                 2m54s  KubeDB Ops-manager Operator  Successfully paused MySQL database: demo/mysql-quickstart for MySQLOpsRequest: myops-rotate-auth-user
  Normal   Starting                                   2m44s  KubeDB Ops-manager Operator  Restarting Pod: mysql-quickstart-0/demo
  Warning  evict pod; ConditionStatus:True            2m44s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True
  Warning  is pod ready; ConditionStatus:False        2m44s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False
  Warning  is pod ready; ConditionStatus:True         2m39s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True
  Warning  is join in cluster; ConditionStatus:False  2m9s   KubeDB Ops-manager Operator  is join in cluster; ConditionStatus:False
  Warning  is pod ready; ConditionStatus:True         2m9s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True
  Warning  is join in cluster; ConditionStatus:True   2m9s   KubeDB Ops-manager Operator  is join in cluster; ConditionStatus:True
  Normal   Successful                                 2m9s   KubeDB Ops-manager Operator  Successfully started MySQL pods for MySQLOpsRequest: demo/myops-rotate-auth-user
  Normal   Starting                                   2m9s   KubeDB Ops-manager Operator  Resuming MySQL database: demo/mysql-quickstart
  Normal   Successful                                 2m9s   KubeDB Ops-manager Operator  Successfully resumed MySQL database: demo/mysql-quickstart
  Normal   Successful                                 2m9s   KubeDB Ops-manager Operator  Controller has successfully rotate MySQL auth secret

```
**Verify auth is rotate**
```shell
$ kubectl get my -n demo mysql-quickstart -ojson | jq .spec.authSecret.name
"mysql-quickstart-auth-user"
$ kubectl get secret -n demo mysql-quickstart-auth-user -o=jsonpath='{.data.username}' | base64 -d
root⏎                                                                
$ kubectl get secret -n demo mysql-quickstart-auth-user -o=jsonpath='{.data.password}' | base64 -d
Mysql2⏎                                                                                       
```

Let's verify if we can connect to the database using the new credentials.
```shell
$ kubectl exec -it -n demo mysql-quickstart-0 -c mysql -- bash
bash-5.1$ mysql -uroot -p"Mysql2"
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 132
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.02 sec)

```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo mysql-quickstart-auth-user -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎           
$ kubectl get secret -n demo mysql-quickstart-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
vYBjULhCEzPwe5xo⏎             
```
Let's confirm that the previous credentials no longer work.
```shell
$ kubectl exec -it -n demo mysql-quickstart-0 -c mysql -- bash
bash-5.1$ mysql -uroot -p"vYBjULhCEzPwe5xo"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'root'@'localhost' (using password: YES)
bash-5.1$ 
```
The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Alternatively, you can delete individual resources by name. To do so, run:

```shell
$ kubectl delete MySQLOpsRequestmyops-rotate-auth-generated myops-rotate-auth-user -n demo
MySQLOpsRequest.ops.kubedb.com "myops-rotate-auth-generated" "myops-rotate-auth-user" deleted
$ kubectl delete secret -n demo mysql-quickstart-auth-user
secret "mysql-quickstart-auth-user" deleted
$ kubectl delete secret -n demo   mysql-quickstart-auth 
secret "mysql-quickstart-auth " deleted

```

## Next Steps

- Learn about [backup and restore](/docs/guides/mysql/backup/overview/index.md) SQL Server using KubeStash.
- Want to set up SQL Server Availability Group clusters? Check how to [Configure SQL Server Availability Gruop Cluster](/docs/guides/mysql/clustering/ag_cluster.md)
- Detail concepts of [Mysql object](/docs/guides/mysql/concepts/mysql.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
