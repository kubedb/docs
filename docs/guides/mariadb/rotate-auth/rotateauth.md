---
title: Rotate Authentication MariaDB 
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-rotate-authentication
    name: Guide
    parent: guides-mariadb-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Rotate Authentication of MariaDB

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `MariaDB` user's authentication credentials using a `MariaDBOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential, and does not provide the secret details directly to the user.
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `password`, and then reference this Secret in the `MariaDBOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MariaDB](/docs/guides/mariadb/concepts/mariadb/index.md)
    - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/mariadb/index.md)

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Create a MariaDB database

KubeDB implements a `MariaDB` CRD to define the specification of a MariaDB database. Below is the `MariaDB` object created in this tutorial.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/quickstart/overview/examples/sample-mariadb-v1.yaml
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until sample-mariadb has status Ready. i.e,

```shell
$ kubectl get mariadb -n demo -w
NAME             VERSION   STATUS   AGE
sample-mariadb   10.5.23   Ready    30m
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get mariadb -n demo sample-mariadb -ojson | jq .spec.authSecret.name
"sample-mariadb-auth"
$ kubectl get secret -n demo sample-mariadb-auth -o=jsonpath='{.data.username}' | base64 -d
root⏎                                  
$ kubectl get secret -n demo sample-mariadb-auth -o=jsonpath='{.data.password}' | base64 -d
s)cJQ*iL8wHySpvT⏎                                                                                                                 
````
Now, you can exec into the pod `sample-mariadb` and connect to database using `username` and `password`
```shell
$ kubectl exec -it -n demo sample-mariadb-0 -- mariadb -u root --password='s)cJQ*iL8wHySpvT'
Defaulted container "mariadb" out of: mariadb, mariadb-init (init)
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 207
Server version: 10.5.23-MariaDB-1:10.5.23+maria~ubu2004 mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
+--------------------+
4 rows in set (0.000 sec)

MariaDB [(none)]> CREATE DATABASE Bharatnaityam;
Query OK, 1 row affected (0.000 sec)

MariaDB [(none)]> show databases;
+--------------------+
| Database           |
+--------------------+
| Bharatnaityam      |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
+--------------------+
5 rows in set (0.000 sec)


```
If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth MariaDBOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the `MariaDB` using operator generated, we have to create a 
`MariaDBOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `MariaDBOpsRequest` CRO that we
are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-mariadb
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on 
`sample-mariadb` database.
- `spec.type` specifies that we are performing `RotateAuth` on MariaDB.

Let's create the `MariaDBOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/rotate-auth/overview/examples/Mariadb-rotate-auth-generated.yaml
 mariadbopsrequest.ops.kubedb.com/mdops-rotate-auth-generated created
```
Let's wait for `MariaDBOpsrequest` to be `Successful`. Run the following command to watch `MariaDBOpsrequest` CRO
```shell
$ kubectl get Mariadbopsrequest -n demo
NAME                          TYPE         STATUS       AGE
mdops-rotate-auth-generated   RotateAuth   Successful   6m28s
```
If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Mariadbopsrequest -n demo mdops-rotate-auth-generated
Name:         mdops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2025-07-11T06:43:21Z
  Generation:          1
  Resource Version:    635526
  UID:                 fdf5a041-9403-4c0b-a788-2adad326dd88
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   sample-mariadb
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-11T06:43:21Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/mdops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-07-11T06:43:24Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                patchedSecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-11T06:43:32Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-mariadb-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-mariadb-0
    Last Transition Time:  2025-07-11T06:43:32Z
    Message:               get pod; ConditionStatus:True; PodName:sample-mariadb-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-mariadb-0
    Last Transition Time:  2025-07-11T06:44:07Z
    Message:               Successfully restarted MariaDB pods for MariaDBOpsRequest: demo/mdops-rotate-auth-generated
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-11T06:44:12Z
    Message:               Successfully rotate MariaDB auth for MariaDBOpsRequest: demo/mdops-rotate-auth-generated
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-11T06:44:12Z
    Message:               Controller has successfully rotate MariaDB auth secret demo/mdops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                     Age    From                         Message
  ----     ------                                                     ----   ----                         -------
  Normal   Starting                                                   9m17s  KubeDB Ops-manager Operator  Start processing for MariaDBOpsRequest: demo/mdops-rotate-auth-generated
  Normal   Starting                                                   9m17s  KubeDB Ops-manager Operator  Pausing MariaDB databse: demo/sample-mariadb
  Normal   Successful                                                 9m17s  KubeDB Ops-manager Operator  Successfully paused MariaDB database: demo/sample-mariadb for MariaDBOpsRequest: mdops-rotate-auth-generated
  Normal   Starting                                                   9m6s   KubeDB Ops-manager Operator  Restarting Pod: demo/sample-mariadb-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-mariadb-0  9m6s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-mariadb-0
  Warning  get pod; ConditionStatus:True; PodName:sample-mariadb-0    9m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-mariadb-0
  Warning  get pod; ConditionStatus:True; PodName:sample-mariadb-0    9m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-mariadb-0
  Warning  get pod; ConditionStatus:True; PodName:sample-mariadb-0    8m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-mariadb-0
  Normal   Successful                                                 8m31s  KubeDB Ops-manager Operator  Successfully restarted MariaDB pods for MariaDBOpsRequest: demo/mdops-rotate-auth-generated
  Normal   Successful                                                 8m26s  KubeDB Ops-manager Operator  Successfully rotate MariaDB auth for MariaDBOpsRequest: demo/mdops-rotate-auth-generated
  Normal   Starting                                                   8m26s  KubeDB Ops-manager Operator  Resuming MariaDB database: demo/sample-mariadb
  Normal   Successful                                                 8m26s  KubeDB Ops-manager Operator  Successfully resumed MariaDB database: demo/sample-mariadb
  Normal   Successful                                                 8m26s  KubeDB Ops-manager Operator  Controller has successfully rotate MariaDB auth secret

```
**Verify Auth is rotated**
```shell
$ kubectl get mariadb -n demo sample-mariadb -ojson | jq .spec.authSecret.name
"sample-mariadb-auth"
$ kubectl get secret -n demo sample-mariadb-auth -o=jsonpath='{.data.username}' | base64 -d
root⏎                                                               
$ kubectl get secret -n demo sample-mariadb-auth -o=jsonpath='{.data.password}' | base64 -d
gTJJMdgpKy9U(Eqi⏎                      
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo sample-mariadb-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎                                                                                                          
$ kubectl get secret -n demo sample-mariadb-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
s)cJQ*iL8wHySpvT⏎                        
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

> Note: You cannot change the database `username`, but you can update the `password` while keeping the existing `username`.

```shell
$ kubectl create secret generic sample-mariadb-auth-user -n demo \
   --type=kubernetes.io/basic-auth \
   --from-literal=username=root \
   --from-literal=password=testpassword
secret/sample-mariadb-auth-user created
```
Now create a `MariaDBOpsRequest` with `RotateAuth` type. Below is the YAML of the `MariaDBOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-mariadb
  authentication:
    secretRef:
      name: sample-mariadb-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `sample-mariadb`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on `MariaDB`.
- `spec.authentication.secretRef.name` specifies that we are using `sample-mariadb-auth-user` as `spec.authSecret.name` for authentication.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/rotate-auth/overview/examples/rotate-auth-user.yaml
mariadbopsrequest.ops.kubedb.com/mdops-rotate-auth-user created
```
Let’s wait for `MariaDBOpsRequest` to be Successful. Run the following command to watch `MariaDBOpsRequest` CRO:

```shell
$ kubectl get Mariadbopsrequest -n demo
NAME                          TYPE         STATUS       AGE
mdops-rotate-auth-generated   RotateAuth   Successful   100s
mdops-rotate-auth-user        RotateAuth   Successful   62s
```
We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed.
```shell
$  kubectl describe Mariadbopsrequest -n demo mdops-rotate-auth-user
Name:         mdops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2025-07-14T06:56:25Z
  Generation:          1
  Resource Version:    665963
  UID:                 6c30dc56-6ca9-4707-8b7c-5d4da6a1b585
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  sample-mariadb-auth-user
  Database Ref:
    Name:   sample-mariadb
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-14T06:56:25Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/mdops-rotate-auth-user
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-07-14T06:56:28Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                patchedSecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-14T06:56:35Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-mariadb-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-mariadb-0
    Last Transition Time:  2025-07-14T06:56:35Z
    Message:               get pod; ConditionStatus:True; PodName:sample-mariadb-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-mariadb-0
    Last Transition Time:  2025-07-14T06:56:40Z
    Message:               Successfully restarted MariaDB pods for MariaDBOpsRequest: demo/mdops-rotate-auth-user
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-14T06:56:45Z
    Message:               Successfully rotate MariaDB auth for MariaDBOpsRequest: demo/mdops-rotate-auth-user
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-14T06:56:45Z
    Message:               Controller has successfully rotate MariaDB auth secret demo/mdops-rotate-auth-user
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                     Age   From                         Message
  ----     ------                                                     ----  ----                         -------
  Normal   Starting                                                   83s   KubeDB Ops-manager Operator  Start processing for MariaDBOpsRequest: demo/mdops-rotate-auth-user
  Normal   Starting                                                   83s   KubeDB Ops-manager Operator  Pausing MariaDB databse: demo/sample-mariadb
  Normal   Successful                                                 83s   KubeDB Ops-manager Operator  Successfully paused MariaDB database: demo/sample-mariadb for MariaDBOpsRequest: mdops-rotate-auth-user
  Normal   Starting                                                   73s   KubeDB Ops-manager Operator  Restarting Pod: demo/sample-mariadb-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-mariadb-0  73s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-mariadb-0
  Warning  get pod; ConditionStatus:True; PodName:sample-mariadb-0    73s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-mariadb-0
  Warning  get pod; ConditionStatus:True; PodName:sample-mariadb-0    68s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-mariadb-0
  Normal   Successful                                                 68s   KubeDB Ops-manager Operator  Successfully restarted MariaDB pods for MariaDBOpsRequest: demo/mdops-rotate-auth-user
  Normal   Successful                                                 63s   KubeDB Ops-manager Operator  Successfully rotate MariaDB auth for MariaDBOpsRequest: demo/mdops-rotate-auth-user
  Normal   Starting                                                   63s   KubeDB Ops-manager Operator  Resuming MariaDB database: demo/sample-mariadb
  Normal   Successful                                                 63s   KubeDB Ops-manager Operator  Successfully resumed MariaDB database: demo/sample-mariadb
  Normal   Successful                                                 63s   KubeDB Ops-manager Operator  Controller has successfully rotate MariaDB auth secret

```
**Verify auth is rotate**
```shell
$ kubectl get mariadb -n demo sample-mariadb -ojson | jq .spec.authSecret.name
"sample-mariadb-auth-user"
$ kubectl get secret -n demo sample-mariadb-auth-user -o=jsonpath='{.data.username}' | base64 -d
root⏎                                                                    
$ kubectl get secret -n demo sample-mariadb-auth-user -o=jsonpath='{.data.password}' | base64 -d
testpassword⏎                                                                                    
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo sample-mariadb-auth-user -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎                                                                                                          
$ kubectl get secret -n demo sample-mariadb-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
gTJJMdgpKy9U(Eqi⏎                                             
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Mariadbopsrequest mdops-rotate-auth-generated mdops-rotate-auth-user -n demo
mariadbopsrequest.ops.kubedb.com "mdops-rotate-auth-generated" deleted
mariadbopsrequest.ops.kubedb.com "mdops-rotate-auth-user" deleted
$ kubectl delete secret -n demo  sample-mariadb-auth-user
secret "sample-mariadb-auth-user" deleted
$ kubectl delete secret -n demo  sample-mariadb-auth
secret "sample-mariadb-auth" deleted
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).