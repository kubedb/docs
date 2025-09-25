---
title: Rotate Authentication PerconaXtraDB
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-rotateauthentiocation
    name: Guide
    parent: guides-perconaxtradb-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of PerconaXtraDB

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `PerconaXtraDB` user's authentication credentials using a `PerconaXtraDBOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential..
2. **User Defined:** The user can create their own credentials by defining a secret of type `kubernetes.io/basic-auth` containing the desired `username` and `password` and then reference this secret in the `PerconaXtraDBOpsRequest`.

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

```
$ kubectl create ns demo
namespace/demo created
```

## Create a PerconaXtraDB database

KubeDB implements a `PerconaXtraDB` CRD to define the specification of a PerconaXtraDB database. Below is the `PerconaXtraDB` object created in this tutorial.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.40"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/quickstart/overview/examples/sample-pxc-v1.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until sample-pxc has status Ready. i.e,

```shell
$  kubectl get perconaxtradb -n demo
NAME         VERSION   STATUS   AGE
sample-pxc   8.0.40    Ready    43m
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the secret.

````shell
$ kubectl get PerconaXtraDB -n demo sample-pxc -ojson | jq .spec.authsecret.name
"sample-pxc-auth"
$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\username}' | base64 -d
root⏎                                                                                                 
$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\password}' | base64 -d
Q!IsZ7.NXM.ZIxvT⏎                       
````

### Connect with PerconaXtraDB database using credentials

Now, you can connect to this database using `telnet`.
Here, we will connect to PerconaXtraDB server from local-machine through port-forwarding.
We will connect to `sample-pxc-0` pod from local-machine using port-frowarding and it must be running in separate terminal.
```bash
$ kubectl port-forward -n demo sample-pxc-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
```
Now, you can exec into the pod `sample-pxc` and connect to database using `username` and `password`
```shell
kubectl exec -it -n demo sample-pxc-0 -- mysql -u root --password='Q!IsZ7.NXM.ZIxvT'
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 970
Server version: 8.0.40-31.1 Percona XtraDB Cluster (GPL), Release rel31, Revision 4b32153, WSREP version 26.1.4.3

Copyright (c) 2009-2024 Percona LLC and/or its affiliates
Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
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
mysql> CREATE DATABASE odissi;
Query OK, 1 row affected (0.03 sec)

mysql> EXIT
Bye

```
If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth PerconaXtraDBOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the PerconaXtraDB using operator generated, we have to create a `PerconaXtraDBOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `PerconaXtraDBOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-pxc
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `sample-pxc` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on PerconaXtraDB.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/perconaXtraDB/rotate-auth/PerconaXtraDB-rotate-auth-generated.yaml
 PerconaXtraDBopsrequest.ops.kubedb.com/pxops-rotate-auth-generated created
```
Let's wait for `PerconaXtraDBOpsrequest` to be `Successful`. Run the following command to watch `PerconaXtraDBOpsrequest` CRO
```shell
 $ kubectl get PerconaXtraDBopsrequest -n demo
NAME                          TYPE         STATUS       AGE
pxops-rotate-auth-generated   RotateAuth   Successful   6m44s
```
If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe PerconaXtraDBopsrequest -n demo pxops-rotate-auth-generated
Name:         pxops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PerconaXtraDBOpsRequest
Metadata:
  Creation Timestamp:  2025-07-21T09:26:47Z
  Generation:          1
  Resource Version:    179861
  UID:                 6a6f2d74-818f-462c-8998-03e3fd9b157e
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   sample-pxc
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-21T09:26:47Z
    Message:               Controller has started to Progress the PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-07-21T09:26:50Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                patchedsecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-21T09:26:59Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-pxc-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-pxc-0
    Last Transition Time:  2025-07-21T09:26:59Z
    Message:               get pod; ConditionStatus:True; PodName:sample-pxc-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-pxc-0
    Last Transition Time:  2025-07-21T09:28:09Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-pxc-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-pxc-1
    Last Transition Time:  2025-07-21T09:28:09Z
    Message:               get pod; ConditionStatus:True; PodName:sample-pxc-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-pxc-1
    Last Transition Time:  2025-07-21T09:29:19Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-pxc-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-pxc-2
    Last Transition Time:  2025-07-21T09:29:19Z
    Message:               get pod; ConditionStatus:True; PodName:sample-pxc-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-pxc-2
    Last Transition Time:  2025-07-21T09:30:29Z
    Message:               Successfully restarted PerconaXtraDB pods for PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-generated
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-21T09:30:34Z
    Message:               Successfully rotate PerconaXtraDB auth for PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-generated
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-21T09:30:34Z
    Message:               Controller has successfully rotate PerconaXtraDB auth secret demo/pxops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>

```

**Verify Auth is rotated**
```shell
$ kubectl get perconaxtradb -n demo sample-pxc -ojson | jq .spec.authsecret.name
"sample-pxc-auth"
$ kubectl get secret -n demo sample-pxc-auth -o=jsonpath='{.data.username}' | base64 -d
root⏎                         
$  kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\password}' | base64 -d
0o~37yrZq(363vDz⏎                                       
```
Also, there will be two more new keys in the secret that stores the previous credentials. The key is `authData.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo sample-pxc-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎   
$ kubectl get secret -n demo sample-pxc-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
Q!IsZ7.NXM.ZIxvT⏎                                         
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,
> Note: The `username` must be fixed as `root`. 
```shell
$  kubectl create secret generic quick-pcx-user-auth -n demo \
                                           --type=kubernetes.io/basic-auth \
                                           --from-literal=username=root \
                                           --from-literal=password=PerconaXtraDB2
secret/quick-pcx-user-auth created
```
Now create a `PerconaXtraDBOpsRequest` with `RotateAuth` type. Below is the YAML of the `PerconaXtraDBOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-pxc
  authentication:
    secretRef:
      kind: Secret
      name: quick-pcx-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `sample-pxc`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on PerconaXtraDB.
- `spec.authentication.secretRef.name` specifies that we are using `quick-pcx-user-auth` as `spec.authsecret.name` for authentication.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/perconaXtraDB/rotate-auth/rotate-auth-user.yaml
PerconaXtraDBopsrequest.ops.kubedb.com/pxops-rotate-auth-user created
```
Let’s wait for `PerconaXtraDBOpsRequest` to be Successful. Run the following command to watch `PerconaXtraDBOpsRequest` CRO:

```shell
$ kubectl get PerconaXtraDBopsrequest -n demo
NAME                          TYPE         STATUS       AGE
pxops-rotate-auth-generated   RotateAuth   Successful    55m
pxops-rotate-auth-user        RotateAuth   Successful    3m44s

```
We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe PerconaXtraDBopsrequest -n demo pxops-rotate-auth-user
Name:         pxops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PerconaXtraDBOpsRequest
Metadata:
  Creation Timestamp:  2025-07-21T10:25:50Z
  Generation:          1
  Resource Version:    184407
  UID:                 612937ae-eb86-440f-859e-16db902159f1
Spec:
  Apply:  IfReady
  Authentication:
    secret Ref:
      Name:  quick-pcx-user-auth
  Database Ref:
    Name:   sample-pxc
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-21T10:25:50Z
    Message:               Controller has started to Progress the PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-user
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-07-21T10:25:50Z
    Message:               Successfully referenced the user provided authsecret
    Observed Generation:   1
    Reason:                patchedsecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-21T10:25:59Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-pxc-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-pxc-0
    Last Transition Time:  2025-07-21T10:25:59Z
    Message:               get pod; ConditionStatus:True; PodName:sample-pxc-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-pxc-0
    Last Transition Time:  2025-07-21T10:27:09Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-pxc-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-pxc-1
    Last Transition Time:  2025-07-21T10:27:09Z
    Message:               get pod; ConditionStatus:True; PodName:sample-pxc-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-pxc-1
    Last Transition Time:  2025-07-21T10:28:19Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-pxc-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-pxc-2
    Last Transition Time:  2025-07-21T10:28:19Z
    Message:               get pod; ConditionStatus:True; PodName:sample-pxc-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--sample-pxc-2
    Last Transition Time:  2025-07-21T10:29:29Z
    Message:               Successfully restarted PerconaXtraDB pods for PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-user
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-21T10:29:34Z
    Message:               Successfully rotate PerconaXtraDB auth for PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-user
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-21T10:29:34Z
    Message:               Controller has successfully rotate PerconaXtraDB auth secret demo/pxops-rotate-auth-user
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                 Age   From                         Message
  ----     ------                                                 ----  ----                         -------
  Normal   Starting                                               39m   KubeDB Ops-manager Operator  Start processing for PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-user
  Normal   Starting                                               39m   KubeDB Ops-manager Operator  Restarting Pod: demo/sample-pxc-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-pxc-0  39m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    39m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Normal   Starting                                               38m   KubeDB Ops-manager Operator  Restarting Pod: demo/sample-pxc-1
  Warning  evict pod; ConditionStatus:True; PodName:sample-pxc-1  38m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    38m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Normal   Starting                                               37m   KubeDB Ops-manager Operator  Restarting Pod: demo/sample-pxc-2
  Warning  evict pod; ConditionStatus:True; PodName:sample-pxc-2  37m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    37m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-0
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-1
  Warning  get pod; ConditionStatus:True; PodName:sample-pxc-2    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:sample-pxc-2
  Normal   Successful                                             36m   KubeDB Ops-manager Operator  Successfully restarted PerconaXtraDB pods for PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-user
  Normal   Successful                                             36m   KubeDB Ops-manager Operator  Successfully rotate PerconaXtraDB auth for PerconaXtraDBOpsRequest: demo/pxops-rotate-auth-user
  Normal   Starting                                               36m   KubeDB Ops-manager Operator  Resuming PerconaXtraDB database: demo/sample-pxc
  Normal   Successful                                             36m   KubeDB Ops-manager Operator  Successfully resumed PerconaXtraDB database: demo/sample-pxc
  Normal   Successful                                             36m   KubeDB Ops-manager Operator  Controller has successfully rotate PerconaXtraDB auth secret

```
**Verify auth is rotate**
```shell
$ kubectl get perconaxtradb -n demo sample-pxc -ojson | jq .spec.authsecret.name
"quick-pcx-user-auth "
$ kubectl get secrets -n demo quick-pcx-user-auth -o jsonpath='{.data.\username}' | base64 -d
root⏎                                                                                     
$ kubectl get secrets -n demo quick-pcx-user-auth -o jsonpath='{.data.\password}' | base64 -d
PerconaXtraDB2⏎                                                                                                               
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo quick-pcx-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎                                                                                                                  
$ kubectl get secret -n demo quick-pcx-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
0o~37yrZq(363vDz⏎                     
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete PerconaXtraDBopsrequest pxops-rotate-auth-generated pxops-rotate-auth-user -n demo
PerconaXtraDBopsrequest.ops.kubedb.com "pxops-rotate-auth-generated" "pxops-rotate-auth-user" deleted
$ kubectl delete secret -n sample-pxc-auth
secret "sample-pxc-auth" deleted
$ kubectl delete secret -n demo   quick-pcx-user-auth 
secret "quick-pcx-user-auth " deleted
```
