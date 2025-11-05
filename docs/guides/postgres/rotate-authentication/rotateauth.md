---
title: Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: pg-rotate-auth-details
    name: Guide
    parent: pg-rotate-authentication
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Rotate Authentication of PostgerSQL

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Postgres` user's authentication credentials using a `PostgresOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential..
2. **User Defined:** The user can create their own credentials by defining a secret of type `kubernetes.io/basic-auth` containing the desired `password` and then reference this secret in the `PostgresOpsRequest`.

## Before You Begin 

- You should be familiar with the following `KubeDB` concepts:
  - [PostgreSQL](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  
- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```
  
## Create a PostgreSQL database
KubeDB implements a Postgres CRD to define the specification of a PostgreSQL database.

You can apply this yaml file:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: quick-postgres
  namespace: demo
spec:
  version: "13.13"
  storageType: Durable
  storage:
     storageClassName: "standard"
     accessModes:
       - ReadWriteOnce
     resources:
       requests:
         storage: 1Gi
  deletionPolicy: DoNotTerminate
```

Command:

```shell
$ kubectl apply -f postgres.yaml 
postgres.kubedb.com/quick-postgres created
```

Or, you can deploy by using command:

```shell
$ kubectl create -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/postgres/quickstart/quick-postgres-v1.yaml
postgres.kubedb.com/quick-postgres created
```

Now, wait until quick-postgres has status Ready. i.e,

```shell
$ kubectl get pg -n demo -w
NAME             VERSION   STATUS   AGE
quick-postgres   13.13     Ready    7m36s
```
## Verify authentication 
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the secret.

````shell
$ kubectl get pg -n demo quick-postgres -ojson | jq .spec.authSecret.name
"quick-postgres-auth"
$ kubectl get secret -n demo quick-postgres-auth -o jsonpath='{.data.username}' | base64 -d
postgres
$ kubectl get secret -n demo quick-postgres-auth -o jsonpath='{.data.password}' | base64 -d
yFj_WnVA9rxfQlLt
````
Now, you can exec into the pod `quick-postgres-0` and connect to database using `username` and `password`
```shell
$ kubectl exec -it -n demo quick-postgres-0 -- bash 
Defaulted container "postgres" out of: postgres, postgres-init-container (init)
 
quick-postgres-0:/$ PGPASSWORD=yFj_WnVA9rxfQlLt psql -U postgres -d postgres -p 5432 -h quick-postgres.demo.svc
psql (13.13)
Type "help" for help.
postgres=# \dt
               List of relations
 Schema |        Name        | Type  |  Owner   
--------+--------------------+-------+----------
 public | kubedb_write_check | table | postgres
 (1 row)
```
If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth PostgresOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Postgres using operator generated, we have to create a `PostgresOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `PostgresOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
    name: pgops-rotate-auth-generated
    namespace: demo
spec:
    type: RotateAuth
    databaseRef:
      name: quick-postgres
    timeout: 5m
    apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `quick-postgres` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Postgres.

Let's create the `PostgresOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/postgres/rotate-auth/postgres-rotate-auth-generated.yaml
 postgresopsrequest.ops.kubedb.com/pgops-rotate-auth-generated created
```
Let's wait for `PostgresOpsrequest` to be `Successful`. Run the following command to watch `PostgresOpsrequest` CRO
```shell
 $ kubectl get postgresopsrequest -n demo
 NAME                          TYPE         STATUS       AGE
 pgops-rotate-auth-generated   RotateAuth   Successful   7m47s
```
If we describe the `PostgresOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe postgresopsrequest -n demo pgops-rotate-auth-generated 
 Name:         pgops-rotate-auth-generated
 Namespace:    demo
 Labels:       <none>
 Annotations:  <none>
 API Version:  ops.kubedb.com/v1alpha1
 Kind:         PostgresOpsRequest
 Metadata:
   Creation Timestamp:  2025-07-08T11:24:10Z
   Generation:          1
   Resource Version:    546623
   UID:                 97a07133-c98e-457c-9249-85c0c690a82e
 Spec:
   Apply:  IfReady
   Database Ref:
     Name:   quick-postgres
   Timeout:  5m
   Type:     RotateAuth
 Status:
   Conditions:
     Last Transition Time:  2025-07-08T11:24:10Z
     Message:               Postgres ops request has started to rotate auth for postgres
     Observed Generation:   1
     Reason:                RotateAuth
     Status:                True
     Type:                  RotateAuth
     Last Transition Time:  2025-07-08T11:24:13Z
     Message:               Successfully generated new credentials
     Observed Generation:   1
     Reason:                UpdateCredential
     Status:                True
     Type:                  UpdateCredential
     Last Transition Time:  2025-07-08T11:24:15Z
     Message:               Successfully updated petsets for rotate auth type
     Observed Generation:   1
     Reason:                UpdatePetSets
     Status:                True
     Type:                  UpdatePetSets
     Last Transition Time:  2025-07-08T11:24:55Z
     Message:               Successfully restarted all the nodes
     Observed Generation:   1
     Reason:                RestartNodes
     Status:                True
     Type:                  RestartNodes
     Last Transition Time:  2025-07-08T11:24:20Z
     Message:               evict pod; ConditionStatus:True
     Observed Generation:   1
     Status:                True
     Type:                  EvictPod
     Last Transition Time:  2025-07-08T11:24:20Z
     Message:               check pod ready; ConditionStatus:False; PodName:quick-postgres-0
     Observed Generation:   1
     Status:                False
     Type:                  CheckPodReady--quick-postgres-0
     Last Transition Time:  2025-07-08T11:24:55Z
     Message:               check pod ready; ConditionStatus:True
     Observed Generation:   1
     Status:                True
     Type:                  CheckPodReady
     Last Transition Time:  2025-07-08T11:24:56Z
     Message:               Successfully Rotated Postgres Auth secret
     Observed Generation:   1
     Reason:                Successful
     Status:                True
     Type:                  Successful
   Observed Generation:     1
   Phase:                   Successful
 Events:
   Type     Reason                                                            Age   From                         Message
   ----     ------                                                            ----  ----                         -------
   Normal   PauseDatabase                                                     20m   KubeDB Ops-manager Operator  Pausing Postgres demo/quick-postgres
   Normal   PauseDatabase                                                     20m   KubeDB Ops-manager Operator  Successfully paused Postgres demo/quick-postgres
   Normal   VersionUpdate                                                     20m   KubeDB Ops-manager Operator  Updating PetSets
   Normal   VersionUpdate                                                     20m   KubeDB Ops-manager Operator  Successfully Updated PetSets
   Warning  evict pod; ConditionStatus:True                                   20m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True
   Warning  check pod ready; ConditionStatus:False; PodName:quick-postgres-0  20m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:quick-postgres-0
   Warning  check pod ready; ConditionStatus:True                             19m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True
   Normal   RestartNodes                                                      19m   KubeDB Ops-manager Operator  Successfully restarted all the nodes
   Normal   ResumeDatabase                                                    19m   KubeDB Ops-manager Operator  Resuming PostgreSQL demo/quick-postgres
   Normal   ResumeDatabase                                                    19m   KubeDB Ops-manager Operator  Successfully resumed PostgreSQL demo/quick-postgres
   Normal   Successful                                                        19m   KubeDB Ops-manager Operator  Successfully Rotated Postgres Auth secret for demo/quick-postgres
  
```
**Verify Auth is rotated**
```shell
$ kubectl get pg -n demo quick-postgres -ojson | jq .spec.authSecret.name
"quick-postgres-auth"
$ kubectl get secret -n demo quick-postgres-auth -o=jsonpath='{.data.username}' | base64 -d
 postgres
$ kubectl get secret -n demo quick-postgres-auth -o jsonpath='{.data.password}' | base64 -d
 zGB9GF!NXwI.2HP9⏎                       
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo quick-postgres-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
postgres
$ kubectl get secret -n demo quick-postgres-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
yFj_WnVA9rxfQlLt                           
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,
> **Note:** Can not change the username while rotating authentication. The username must be same as 'postgres' which is the current username of the database.

```shell
$ kubectl create secret generic quick-postgres-user-auth -n demo \
                                                --type=kubernetes.io/basic-auth \
                                                --from-literal=username=postgres \
                                                --from-literal=password=postgres-secret
 secret/quick-postgres-user-auth created
```
Now create a `PostgresOpsRequest` with `RotateAuth` type. Below is the YAML of the `PostgresOpsRequest` 
that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pgops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: quick-postgres
  authentication:
    secretRef:
      kind: Secret
      name: quick-postgres-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `quick-postgres`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on postgres.
- `spec.authentication.secretRef.name` specifies that we are using `quick-postgres-user-auth` as `spec.authsecret.name` for authentication.

Let's create the `PostgresOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/postgres/rotate-auth/rotate-auth-user.yaml
postgresopsrequest.ops.kubedb.com/pgops-rotate-auth-user created
```
Let’s wait for `PostgresOpsRequest` to be Successful. Run the following command to watch `PostgresOpsRequest` CRO:

```shell
$ kubectl get postgresopsrequest -n demo
NAME                          TYPE         STATUS       AGE
pgops-rotate-auth-generated   RotateAuth   Successful   19h
pgops-rotate-auth-user        RotateAuth   Successful   7m44s
```
We can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe postgresopsrequest -n demo pgops-rotate-auth-user 
Name:         pgops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
Metadata:
  Creation Timestamp:  2025-07-09T06:45:44Z
  Generation:          1
  Resource Version:    562328
  UID:                 d25c3d36-cc15-4c82-8fe4-64e5ffc1467c
Spec:
  Apply:  IfReady
  Authentication:
    secret Ref:
      Name:  quick-postgres-user-auth
  Database Ref:
    Name:   quick-postgres
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-09T06:45:44Z
    Message:               Postgres ops request has started to rotate auth for postgres
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-09T06:45:47Z
    Message:               Successfully referenced the user provided authsecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-09T06:45:50Z
    Message:               Successfully updated petsets for rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-09T06:46:30Z
    Message:               Successfully restarted all the nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-09T06:45:55Z
    Message:               evict pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod
    Last Transition Time:  2025-07-09T06:45:55Z
    Message:               check pod ready; ConditionStatus:False; PodName:quick-postgres-0
    Observed Generation:   1
    Status:                False
    Type:                  CheckPodReady--quick-postgres-0
    Last Transition Time:  2025-07-09T06:46:30Z
    Message:               check pod ready; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady
    Last Transition Time:  2025-07-09T06:46:30Z
    Message:               Successfully Rotated Postgres Auth secret
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                            Age    From                         Message
  ----     ------                                                            ----   ----                         -------
  Normal   PauseDatabase                                                     10m    KubeDB Ops-manager Operator  Pausing Postgres demo/quick-postgres
  Normal   PauseDatabase                                                     10m    KubeDB Ops-manager Operator  Successfully paused Postgres demo/quick-postgres
  Normal   VersionUpdate                                                     10m    KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                     10m    KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  evict pod; ConditionStatus:True                                   10m    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True
  Warning  check pod ready; ConditionStatus:False; PodName:quick-postgres-0  10m    KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:quick-postgres-0
  Warning  check pod ready; ConditionStatus:True                             9m58s  KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True
  Normal   RestartNodes                                                      9m58s  KubeDB Ops-manager Operator  Successfully restarted all the nodes
  Normal   ResumeDatabase                                                    9m58s  KubeDB Ops-manager Operator  Resuming PostgreSQL demo/quick-postgres
  Normal   ResumeDatabase                                                    9m58s  KubeDB Ops-manager Operator  Successfully resumed PostgreSQL demo/quick-postgres
  Normal   Successful                                                        9m58s  KubeDB Ops-manager Operator  Successfully Rotated Postgres Auth secret for demo/quick-postgres

```
**Verify auth is rotate**
```shell
$ kubectl get pg -n demo quick-postgres -ojson | jq .spec.authSecret.name
"quick-postgres-user-auth"
$ kubectl get secret -n demo quick-postgres-user-auth-new -o=jsonpath='{.data.username}' | base64 -d
postgres                                        
$ kubectl get secret -n demo quick-postgres-user-auth-new -o=jsonpath='{.data.password}' | base64 -d
postgres-secret                                                                
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo quick-postgres-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
postgres                                                                                    
$ kubectl get secret -n demo quick-postgres-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
zGB9GF!NXwI.2HP9 
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete postgresopsrequest pgops-rotate-auth-generated pgops-rotate-auth-user -n demo
postgresopsrequest.ops.kubedb.com "pgops-rotate-auth-generated" "pgops-rotate-auth-user" deleted
$ kubectl delete secret -n demo  quick-postgres-user-auth
secret "quick-postgres-user-auth" deleted
$ kubectl delete secret -n demo  quick-postgres-auth
secret "quick-postgres-auth" deleted
```


## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/stash/overview/index.md) PostgreSQL database using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Learn about [custom PostgresVersions](/docs/guides/postgres/custom-versions/setup.md).
- Want to setup PostgreSQL cluster? Check how to [configure Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy PostgreSQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
