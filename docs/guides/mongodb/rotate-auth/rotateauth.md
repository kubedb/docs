---
title: Rotate Authentication MongoDB
menu:
  docs_{{ .version }}:
    identifier: mg-rotate-auth
    name: Rotate Authentication
    parent: mg-mongodb-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Rotate Authentication of MongoDB

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `MongoDB` user's authentication credentials using a `MongoDBOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential, and does not provide the secret details directly to the user.
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `username` and `password`, and then reference this Secret in the `MongoDBOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
    - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Create a MongoDB database
KubeDB implements a MongoDB CRD to define the specification of a MongoDB database.

You can apply this yaml file:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mgo-quickstart
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "rs1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Command:

```shell
$ kubectl apply -f mongobd.yaml 
mongodb.kubedb.com/mgo-quickstart created
```

Or, you can deploy by using command:

```shell
$  kubectl create -f https://github.com/kubedb/docs/raw/v2025.6.30/docs/examples/mongodb/quickstart/replicaset-v1alpha2.yaml
mongodb.kubedb.com/mgo-quickstart created
```

Now, wait until mgo-quickstart has status Ready. i.e,

```shell
$ kubectl get mg -n demo -w
NAME             VERSION   STATUS   AGE
mgo-quickstart   4.4.26    Ready      8m1s
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get mg -n demo mgo-quickstart -ojson | jq .spec.authSecret.name
"mgo-quickstart-auth"
$ kubectl get secret -n demo mgo-quickstart-auth -o=jsonpath='{.data.username}' | base64 -d
root⏎                                  
$ kubectl get secret -n demo mgo-quickstart-auth -o=jsonpath='{.data.password}' | base64 -d
eR*W_mz6bjyZxeiG⏎                                                                                                                
````
Now, you can exec into the pod `mgo-quickstart` and connect to database using `username` and `password`
```shell
$ kubectl exec -it -n demo mgo-quickstart-0 -- bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mgo-quickstart-0:/$ mongo -u root -p $MONGO_INITDB_ROOT_PASSWORD 
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("dcd7f912-93d0-4f24-843d-5e2cbecbb6e0") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://docs.mongodb.com/
Questions? Try the MongoDB Developer Community Forums
	https://community.mongodb.com
---
The server generated these startup warnings when booting: 
        2025-07-10T08:40:43.374+00:00: Using the XFS filesystem is strongly recommended with the WiredTiger storage engine. See http://dochub.mongodb.org/core/prodnotes-filesystem
---
rs1:SECONDARY> use Mohiniyattam
switched to db Mohiniyattam

```
If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth MongoDBOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Postgres using operator generated, we have to create a `MongoDBOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `MongoDBOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mgops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mgo-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `mgo-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on MongoDB.

Let's create the `MongoDBOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/v2025.6.30/docs/examples/mongodb/rotate-auth/mongodb-rotate-auth-generated.yaml
 mongodbopsrequest.ops.kubedb.com/mgops-rotate-auth-generated created
```
Let's wait for `MongoDBOpsrequest` to be `Successful`. Run the following command to watch `MongoDBOpsrequest` CRO
```shell
 $kubectl get mongodbopsrequest -n demo
NAME                          TYPE         STATUS       AGE
mgops-rotate-auth-generated   RotateAuth   Successful   45m
```
If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe mongodbopsrequest -n demo mgops-rotate-auth-generated 
Name:         mgops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2025-07-10T08:39:47Z
  Generation:          1
  Resource Version:    607260
  UID:                 20c8ac77-20b8-45b0-b213-a3f8f06cc379
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mgo-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-10T08:39:47Z
    Message:               MongoDB ops request has started to rotate auth for mongodb
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-10T08:39:50Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-10T08:39:55Z
    Message:               Successfully updated petsets rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-10T08:40:00Z
    Message:               check is master; ConditionStatus:True; PodName:mgo-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckIsMaster--mgo-quickstart-1
    Last Transition Time:  2025-07-10T08:40:00Z
    Message:               evict pod; ConditionStatus:True; PodName:mgo-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mgo-quickstart-1
    Last Transition Time:  2025-07-10T08:40:15Z
    Message:               check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--mgo-quickstart-1
    Last Transition Time:  2025-07-10T08:40:15Z
    Message:               check is master; ConditionStatus:True; PodName:mgo-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckIsMaster--mgo-quickstart-2
    Last Transition Time:  2025-07-10T08:40:15Z
    Message:               evict pod; ConditionStatus:True; PodName:mgo-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mgo-quickstart-2
    Last Transition Time:  2025-07-10T08:40:40Z
    Message:               check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--mgo-quickstart-2
    Last Transition Time:  2025-07-10T08:40:40Z
    Message:               check is master; ConditionStatus:True; PodName:mgo-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckIsMaster--mgo-quickstart-0
    Last Transition Time:  2025-07-10T08:40:40Z
    Message:               step down; ConditionStatus:True; PodName:mgo-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  StepDown--mgo-quickstart-0
    Last Transition Time:  2025-07-10T08:40:40Z
    Message:               evict pod; ConditionStatus:True; PodName:mgo-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mgo-quickstart-0
    Last Transition Time:  2025-07-10T08:40:55Z
    Message:               check pod ready; ConditionStatus:True; PodName:mgo-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--mgo-quickstart-0
    Last Transition Time:  2025-07-10T08:40:55Z
    Message:               Successfully Restarted ReplicaSet nodes
    Observed Generation:   1
    Reason:                RestartReplicaSet
    Status:                True
    Type:                  RestartReplicaSet
    Last Transition Time:  2025-07-10T08:40:55Z
    Message:               Successfully Rotate Auth
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                            Age   From                         Message
  ----     ------                                                            ----  ----                         -------
  Normal   PauseDatabase                                                     44m   KubeDB Ops-manager Operator  Pausing MongoDB demo/mgo-quickstart
  Normal   PauseDatabase                                                     44m   KubeDB Ops-manager Operator  Successfully paused MongoDB demo/mgo-quickstart
  Normal   VersionUpdate                                                     44m   KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                     44m   KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  check is master; ConditionStatus:True; PodName:mgo-quickstart-1   44m   KubeDB Ops-manager Operator  check is master; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:mgo-quickstart-1         44m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:False; PodName:mgo-quickstart-1  44m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   44m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check is master; ConditionStatus:True; PodName:mgo-quickstart-2   44m   KubeDB Ops-manager Operator  check is master; ConditionStatus:True; PodName:mgo-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:mgo-quickstart-2         44m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mgo-quickstart-2
  Warning  check pod ready; ConditionStatus:False; PodName:mgo-quickstart-2  44m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:mgo-quickstart-2
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2
  Warning  check is master; ConditionStatus:True; PodName:mgo-quickstart-0   43m   KubeDB Ops-manager Operator  check is master; ConditionStatus:True; PodName:mgo-quickstart-0
  Warning  step down; ConditionStatus:True; PodName:mgo-quickstart-0         43m   KubeDB Ops-manager Operator  step down; ConditionStatus:True; PodName:mgo-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:mgo-quickstart-0         43m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mgo-quickstart-0
  Warning  check pod ready; ConditionStatus:False; PodName:mgo-quickstart-0  43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:mgo-quickstart-0
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-1
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-2
  Warning  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-0   43m   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:mgo-quickstart-0
  Normal   RestartReplicaSet                                                 43m   KubeDB Ops-manager Operator  Successfully Restarted ReplicaSet nodes
  Normal   ResumeDatabase                                                    43m   KubeDB Ops-manager Operator  Resuming MongoDB demo/mgo-quickstart
  Normal   ResumeDatabase                                                    43m   KubeDB Ops-manager Operator  Successfully resumed MongoDB demo/mgo-quickstart
  Normal   Successful                                                        43m   KubeDB Ops-manager Operator  Successfully Rotate Auth

```
**Verify Auth is rotated**
```shell
$ kubectl get mg -n demo mgo-quickstart -ojson | jq .spec.authSecret.name
"mgo-quickstart-auth"
$ kubectl get secret -n demo mgo-quickstart-auth -o=jsonpath='{.data.username}' | base64 -d
root⏎                                                               
$ kubectl get secret -n demo mgo-quickstart-auth -o=jsonpath='{.data.password}' | base64 -d
09wZM.)t8kpwKF5z⏎                      
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo mgo-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎                                                                                                          
$ kubectl get secret -n demo mgo-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
eR*W_mz6bjyZxeiG⏎                        
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

```shell
$  kubectl create secret generic quick-mg-user-auth -n demo \
                                                                                          --type=kubernetes.io/basic-auth \
                                                                                          --from-literal=username=mg-admin \
                                                                                          --from-literal=password=mongodb-secret
secret/quick-mg-user-auth created
```
Now create a `MongoDBOpsRequest` with `RotateAuth` type. Below is the YAML of the `MongoDBOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mgops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mgo-quickstart
  authentication:
    secretRef:
      name: quick-mg-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `mgo-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on postgres.
- `spec.authentication.secretRef.name` specifies that we are using `quick-mg-user-auth` as `spec.authSecret.name` for authentication.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/v2025.6.30/docs/examples/mongodb/rotate-auth/rotate-auth-user.yaml
mongodbopsrequest.ops.kubedb.com/mgops-rotate-auth-user created
```
Let’s wait for `MongoDBOpsRequest` to be Successful. Run the following command to watch `MongoDBOpsRequest` CRO:

```shell
$ kubectl get mongodbopsrequest -n demo
NAME                          TYPE         STATUS       AGE
mgops-rotate-auth-generated   RotateAuth   Successful   153m
mgops-rotate-auth-user        RotateAuth   Failed       59m
```
We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed.
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
    Secret Ref:
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
    Message:               Successfully referenced the user provided authSecret
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
    Message:               Successfully Rotated Postgres Auth Secret
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
  Normal   Successful                                                        9m58s  KubeDB Ops-manager Operator  Successfully Rotated Postgres Auth Secret for demo/quick-postgres

```
**Verify auth is rotate**
```shell
$ kubectl get pg -n demo mgo-quickstart -ojson | jq .spec.authSecret.name
"quick-mg-user-auth"
$ kubectl get secret -n demo quick-mg-user-auth -o=jsonpath='{.data.username}' | base64 -d
mg-admin⏎                                                                    
$ kubectl get secret -n demo quick-mg-user-auth -o=jsonpath='{.data.password}' | base64 -d
mongodb-secret⏎                                                                                    
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo quick-mg-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root⏎                                                                                                          
$ kubectl get secret -n demo quick-mg-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
09wZM.)t8kpwKF5z⏎                                             
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete mongodbopsrequest mgops-rotate-auth-generated mgops-rotate-auth-user -n demo
mongodbopsrequest.ops.kubedb.com "mgops-rotate-auth-generated" "mgops-rotate-auth-user" deleted
$ kubectl delete secret -n demo  quick-mg-user-auth
secret "quick-mg-user-auth" deleted
$ kubectl delete secret -n demo  mgo-quickstart-auth
secret "mgo-quickstart-auth" deleted
```


## Next Steps

- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

