---
title: Rotate Authentication PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-rotateauthentication
    name: Guide
    parent: pb-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of PgBouncer

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `PgBouncer` user's authentication credentials using a `PgBouncerOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential..
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `username` and `password`, and then reference this Secret in the `PgBouncerOpsRequest`.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

**We have designed this tutorial to demonstrate a production setup of KubeDB managed PgBouncer.**
**If you just want to try out KubeDB, you can bypass some of the safety features following the tips [here](/docs/guides/pgbouncer/quickstart/quickstart.md#tips-for-testing).**

> Note: Before you begin, you have to create `PgBouncer` CRD with the help of [this](/docs/guides/pgbouncer/quickstart/quickstart.md).

## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get pgbouncer -n demo pgbouncer-server -ojson | jq .spec.authSecret.name
"pgbouncer-server-auth"
$ kubectl get secrets -n demo pgbouncer-server-auth -o jsonpath='{.data.\username}' | base64 -d
pgbouncer⏎     
$ kubectl get secrets -n demo pgbouncer-server-auth -o jsonpath='{.data.\password}' | base64 -d
I*7SQB7)6~Kni8*X⏎                                    
````

### Connect with PgBouncer database using credentials

Here, we will connect to PgBouncer server from local-machine through port-forwarding.
We will connect to `pgbouncer-server` pod from local-machine using port-frowarding and it must be running in separate terminal.
```bash
$ kubectl port-forward svc/pgbouncer-server -n demo 9999:5432
Forwarding from 127.0.0.1:9999 -> 5432
Forwarding from [::1]:9999 -> 5432

```
Now, you can exec into the pod pgbouncer-server` and connect to database using `username` and `password`
```shell
$ kubectl exec -it -n demo pgbouncer-server-0 -- /bin/sh
/ $ cat /var/run/pgbouncer/secret/userlist
"pgbouncer" "md5f24d95a2a5c1ed1debe8c3e6f19ac7ec"
"postgres" "md5095f5936c7d03fbc4998320d3bf993c4"
/ $ exit
```
First, you have to have `PostgreSQL` to run these commands.
```shell
$ export PGPASSWORD='I*7SQB7)6~Kni8*X'
$ psql --host=localhost --port=9999 --username=pgbouncer -d pgbouncer
psql (16.9 (Ubuntu 16.9-0ubuntu0.24.04.1), server 1.18.0/bouncer)
WARNING: psql major version 16, server major version 1.18.
         Some psql features might not work.
Type "help" for help.
pgbouncer=# show databases;
  name    |          host           | port | database  | force_user | pool_size | min_pool_size | reserve_pool | pool_mode | max_connections | current_connections | paused | disabled 
-----------+-------------------------+------+-----------+------------+-----------+---------------+--------------+-----------+-----------------+---------------------+--------+----------
 pgbouncer |                         | 5432 | pgbouncer | pgbouncer  |         2 |             0 |            0 | statement |               1 |                   0 |      0 |        0
 postgres  | quick-postgres.demo.svc | 5432 | postgres  |            |        20 |             1 |            5 |           |               1 |                   1 |      0 |        0
(2 rows)

```

If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth PgBouncerOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the PgBouncer using operator generated, we have to create a `PgBouncerOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `PgBouncerOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: pgbouncer-server
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `pgbouncer-server` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on PgBouncer.

Let's create the `PgBouncerOpsRequest` CR we have shown above,
```shell
 $kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/rotate-auth/rotate-auth.yaml
 pgbounceropsrequest.ops.kubedb.com/pbops-rotate-auth-generated created
```
Let's wait for `PgBouncerOpsrequest` to be `Successful`. Run the following command to watch `PgBouncerOpsrequest` CRO
```shell
 $ kubectl get PgBouncerOpsRequest -n demo
NAME                          TYPE         STATUS       AGE
pbops-rotate-auth-generated   RotateAuth   Successful   4m34s
```
If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed.
```shell
$   kubectl describe PgBounceropsrequest -n demo pbops-rotate-auth-generated
Name:         pbops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T04:25:48Z
  Generation:          1
  Resource Version:    103172
  UID:                 5c5ac3bb-ee72-4d1d-b10c-e2951b4560b9
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   pgbouncer-server
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-18T04:25:48Z
    Message:               PgBouncer ops request has started to rotate auth for pgbouncer
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-18T04:25:51Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-18T04:26:32Z
    Message:               Successfully updated petsets rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-18T04:26:37Z
    Message:               get pod; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:26:37Z
    Message:               evict pod; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:26:42Z
    Message:               check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckReplicaFunc--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:26:42Z
    Message:               check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:27:02Z
    Message:               check pg bouncer running; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPgBouncerRunning--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:27:07Z
    Message:               Restart performed successfully in PgBouncer: demo/pgbouncer-server for PgBouncerOpsRequest: pbops-rotate-auth-generated
    Observed Generation:   1
    Reason:                RestartPodsSucceeded
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-07-18T04:27:07Z
    Message:               Controller has successfully completed  with RotateAuth of PgBouncerOpsRequest: demo/pbops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                       Age    From                         Message
  ----     ------                                                                       ----   ----                         -------
  Normal   Starting                                                                     5m51s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pgbouncer-server
  Normal   Successful                                                                   5m51s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pgbouncer-server for PgBouncerOpsRequest: pbops-rotate-auth-generated
  Normal   VersionUpdate                                                                5m48s  KubeDB Ops-manager Operator  Updating PetSets
  Warning  Failed                                                                       5m28s  KubeDB Ops-manager Operator  Operation cannot be fulfilled on pgbouncers.kubedb.com "pgbouncer-server": the object has been modified; please apply your changes to the latest version and try again
  Normal   VersionUpdate                                                                5m28s  KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                                5m7s   KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  get pod; ConditionStatus:True; PodName:pgbouncer-server-0                    5m2s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  evict pod; ConditionStatus:True; PodName:pgbouncer-server-0                  5m2s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0         4m57s  KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0            4m57s  KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pg bouncer running; ConditionStatus:False; PodName:pgbouncer-server-0  4m47s  KubeDB Ops-manager Operator  check pg bouncer running; ConditionStatus:False; PodName:pgbouncer-server-0
  Warning  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0         4m47s  KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0            4m47s  KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0         4m37s  KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0            4m37s  KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pg bouncer running; ConditionStatus:True; PodName:pgbouncer-server-0   4m37s  KubeDB Ops-manager Operator  check pg bouncer running; ConditionStatus:True; PodName:pgbouncer-server-0
  Normal   Successful                                                                   4m32s  KubeDB Ops-manager Operator  Restart performed successfully in PgBouncer: demo/pgbouncer-server for PgBouncerOpsRequest: pbops-rotate-auth-generated
```

**Verify Auth is rotated**
```shell
$ kubectl get PgBouncer -n demo pgbouncer-server -ojson | jq .spec.authSecret.name
"pgbouncer-server-auth"
$ kubectl get secret -n demo pgbouncer-server-auth -o=jsonpath='{.data.username}' | base64 -d
pgbouncer⏎      
$ kubectl get secrets -n demo pgbouncer-server-auth -o jsonpath='{.data.\password}' | base64 -d
Hc5nXhC403rvDGPf⏎                                                        
```
Also, there will be two more new keys in the secret that stores the previous credentials. The key is `authData.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo pgbouncer-server-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
pgbouncer⏎        
$ kubectl get secret -n demo pgbouncer-server-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
I*7SQB7)6~Kni8*X⏎                                                   
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

```shell
$  kubectl create secret generic quick-pb-user-auth -n demo \
        --type=kubernetes.io/basic-auth \
        --from-literal=username=user \
        --from-literal=password=PgBouncer2
secret/quick-pb-user-auth created
```
Now create a `PgBouncerOpsRequest` with `RotateAuth` type. Below is the YAML of the `PgBouncerOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: pgbouncer-server
  authentication:
    secretRef:
      name: quick-pb-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `pgbouncer-server`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on PgBouncer.
- `spec.authentication.secretRef.name` specifies that we are using `quick-pb-user-auth` as `spec.authSecret.name` for authentication.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```shell
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/rotate-auth/rotateauthuser.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-rotate-auth-user created
```
Let’s wait for `PgBouncerOpsRequest` to be Successful. Run the following command to watch `PgBouncerOpsRequest` CRO:

```shell
$ kubectl get PgBouncerOpsRequest -n demo
NAME                          TYPE         STATUS       AGE
pbops-rotate-auth-generated   RotateAuth   Successful   20m
pbops-rotate-auth-user        RotateAuth   Successful   6m16s
```
We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed.
```shell
$  kubectl describe PgBounceropsrequest -n demo pbops-rotate-auth-user
Name:         pbops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T04:39:59Z
  Generation:          1
  Resource Version:    104227
  UID:                 7a02957e-0144-4d6e-826b-8e27e04a6074
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  quick-pb-user-auth
  Database Ref:
    Name:   pgbouncer-server
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-18T04:39:59Z
    Message:               PgBouncer ops request has started to rotate auth for pgbouncer
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-18T04:39:59Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-18T04:40:00Z
    Message:               Successfully updated petsets rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-18T04:40:05Z
    Message:               get pod; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:40:05Z
    Message:               evict pod; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:40:10Z
    Message:               check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckReplicaFunc--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:40:10Z
    Message:               check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:40:30Z
    Message:               check pg bouncer running; ConditionStatus:True; PodName:pgbouncer-server-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPgBouncerRunning--pgbouncer-server-0
    Last Transition Time:  2025-07-18T04:40:35Z
    Message:               Restart performed successfully in PgBouncer: demo/pgbouncer-server for PgBouncerOpsRequest: pbops-rotate-auth-user
    Observed Generation:   1
    Reason:                RestartPodsSucceeded
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-07-18T04:40:45Z
    Message:               Controller has successfully completed  with RotateAuth of PgBouncerOpsRequest: demo/pbops-rotate-auth-user
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                       Age    From                         Message
  ----     ------                                                                       ----   ----                         -------
  Normal   VersionUpdate                                                                6m54s  KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                                6m53s  KubeDB Ops-manager Operator  Successfully Updated PetSets
  Normal   VersionUpdate                                                                6m53s  KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                                6m52s  KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  get pod; ConditionStatus:True; PodName:pgbouncer-server-0                    6m48s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  evict pod; ConditionStatus:True; PodName:pgbouncer-server-0                  6m48s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0         6m43s  KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0            6m43s  KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pg bouncer running; ConditionStatus:False; PodName:pgbouncer-server-0  6m33s  KubeDB Ops-manager Operator  check pg bouncer running; ConditionStatus:False; PodName:pgbouncer-server-0
  Warning  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0         6m33s  KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0            6m33s  KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0         6m23s  KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0            6m23s  KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pgbouncer-server-0
  Warning  check pg bouncer running; ConditionStatus:True; PodName:pgbouncer-server-0   6m23s  KubeDB Ops-manager Operator  check pg bouncer running; ConditionStatus:True; PodName:pgbouncer-server-0
  Normal   Successful                                                                   6m18s  KubeDB Ops-manager Operator  Restart performed successfully in PgBouncer: demo/pgbouncer-server for PgBouncerOpsRequest: pbops-rotate-auth-user
  
```
**Verify auth is rotate**
```shell
$  kubectl get PgBouncer -n demo pgbouncer-server -ojson | jq .spec.authSecret.name
"quick-pb-user-auth"
$ kubectl get secrets -n demo quick-pb-user-auth -o jsonpath='{.data.\username}' | base64 -d
user⏎                                                                                       
$ kubectl get secrets -n demo quick-pb-user-auth -o jsonpath='{.data.\password}' | base64 -d
PgBouncer2⏎                                                                                                                  
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$  kubectl get secret -n demo quick-pb-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
pgbouncer⏎                                                                                                                                      
$ kubectl get secret -n demo quick-pb-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
Hc5nXhC403rvDGPf⏎                                
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete PgBounceropsrequest pbops-rotate-auth-generated pbops-rotate-auth-user -n demo
PgBounceropsrequest.ops.kubedb.com "pbops-rotate-auth-generated" "pbops-rotate-auth-user" deleted
$ kubectl delete secret -n demoquick-pb-user-auth
secret "quick-pb-user-auth" deleted
$ kubectl delete secret -n demo   pgbouncer-server-auth
secret "pgbouncer-server-auth " deleted
```


## Next Steps

- Learn how to use KubeDB to run a PostgreSQL database [here](/docs/guides/postgres/README.md).
- Learn how to how to get started with PgBouncer [here](/docs/guides/pgbouncer/quickstart/quickstart.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
