---
title: RabbitMQ Rotateauth Guide
menu:
docs_{{ .version }}:
identifier: rm-rotateauth-guide
name: Guide
parent: rm-rotateauth
weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of RabbitMQ

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a 
`RabbitMQ` user's authentication credentials using a `RabbitMQOpsRequest`. There are two ways to 
perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential and 
updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a secret of type
   `kubernetes.io/basic-auth` containing the desired `password` and then reference this secret in 
the `RabbitMQOpsRequest` CR.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be 
configured to communicate with your cluster. If you do not already have a cluster, you can create 
one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB in your cluster following the steps [here](/docs/setup/README.md) and make 
sure install with helm command including `--set global.featureGates.RabbitMQ=true` to ensure 
RabbitMQ CRDs.

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run 
KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY     VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete            WaitForFirstConsumer   false                  6h22m
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this 
tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available RabbitMQVersion

When you have installed KubeDB, it has created `RabbitMQVersion` CR for all supported RabbitMQ versions. Check it by using the `kubectl get rabbitmqversions` command. You can also use `rmv` shorthand instead of `rabbitmqversions`.

```bash
$ kubectl get rabbitmqversion
NAME      VERSION   DB_IMAGE                                                     DEPRECATED   AGE
3.12.12   3.12.12   ghcr.io/appscode-images/rabbitmq:3.12.12-management-alpine                3h13m
3.13.2    3.13.2    ghcr.io/appscode-images/rabbitmq:3.13.2-management-alpine                 3h13m
4.0.4     4.0.4     ghcr.io/appscode-images/rabbitmq:4.0.4-management-alpine                  3h13m
```

## Create a RabbitMQ server

KubeDB implements a `RabbitMQ` CRD to define the specification of a RabbitMQ server. Below is the `RabbitMQ` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq
  namespace: demo
spec:
  deletionPolicy: Delete
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  version: 3.12.12
```

```bash
$ kubectl apply -f rabbit.yaml
RabbitMQ.kubedb.com/rabbitmq created
```

## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the secret.

````shell
$ kubectl get rm -n demo rabbitmq -ojson | jq .spec.authsecret.name
"rabbitmq-auth"
$ kubectl get secret -n demo rabbitmq-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎          
$ kubectl get secret -n demo rabbitmq-auth -o jsonpath='{.data.password}' | base64 -d
4TC.R7hXc1g;kA)P⏎                        
````
Now, you can exec into the pod `rabbitmq-0` and connect to database using `username` and `password`
```bash
$ kubectl exec -it -n demo rabbitmq-0 -c rabbitmq -- bash
rabbitmq-0:/$ rabbitmqadmin -u admin -p '4TC.R7hXc1g;kA)P' list queues
+---------------+----------+
|     name      | messages |
+---------------+----------+
| kubedb_system | 0        |
+---------------+----------+
rabbitmq-0:/$ exit
exit
```

If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth RabbitMQOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the RabbitMQ using operator generated, we have to create a `RabbitMQOpsRequest` CR with `RotateAuth` type. Below is the YAML of the `RabbitMQOpsRequest` CR that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rm-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: rabbitmq
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `rabbitmq` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on RabbitMQ.

Let's create the `RabbitMQOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/rabbitmq/rotate-auth/rotate-auth-generated.yaml
 RabbitMQopsrequest.ops.kubedb.com/rm-rotate-auth-generated created
```
Let's wait for `RabbitMQOpsrequest` to be `Successful`. Run the following command to watch `RabbitMQOpsrequest` CR
```shell
 $ kubectl get RabbitMQopsrequest -n demo
NAME                       TYPE         STATUS       AGE
rm-rotate-auth-generated   RotateAuth   Successful   3m14s
```
If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe RabbitMQopsrequest -n demo rm-rotate-auth-generated
Name:         rm-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
Metadata:
  Creation Timestamp:  2025-08-15T14:52:14Z
  Generation:          1
  Resource Version:    36322
  UID:                 547ceec2-492e-4c11-a432-a15e849dbd8f
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   rabbitmq
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-15T14:52:14Z
    Message:               rabbitmq ops request has started to rotate auth for rmq nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-15T14:52:22Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-15T14:52:22Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-15T14:52:29Z
    Message:               successfully reconciled the rabbitmq with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-15T14:52:34Z
    Message:               get pod; ConditionStatus:True; PodName:rabbitmq-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rabbitmq-0
    Last Transition Time:  2025-08-15T14:52:34Z
    Message:               evict pod; ConditionStatus:True; PodName:rabbitmq-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rabbitmq-0
    Last Transition Time:  2025-08-15T14:52:39Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-15T14:52:44Z
    Message:               running pod; ConditionStatus:True; PodName:rabbitmq-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--rabbitmq-0
    Last Transition Time:  2025-08-15T14:52:49Z
    Message:               get pod; ConditionStatus:True; PodName:rabbitmq-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rabbitmq-1
    Last Transition Time:  2025-08-15T14:52:49Z
    Message:               evict pod; ConditionStatus:True; PodName:rabbitmq-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rabbitmq-1
    Last Transition Time:  2025-08-15T14:52:59Z
    Message:               running pod; ConditionStatus:True; PodName:rabbitmq-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--rabbitmq-1
    Last Transition Time:  2025-08-15T14:53:04Z
    Message:               get pod; ConditionStatus:True; PodName:rabbitmq-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rabbitmq-2
    Last Transition Time:  2025-08-15T14:53:04Z
    Message:               evict pod; ConditionStatus:True; PodName:rabbitmq-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rabbitmq-2
    Last Transition Time:  2025-08-15T14:53:14Z
    Message:               running pod; ConditionStatus:True; PodName:rabbitmq-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--rabbitmq-2
    Last Transition Time:  2025-08-15T14:53:19Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-15T14:53:19Z
    Message:               Successfuly completed reconfigure rmq
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                 Age    From                         Message
  ----     ------                                                 ----   ----                         -------
  Normal   Starting                                               3m50s  KubeDB Ops-manager Operator  Start processing for RabbitMQOpsRequest: demo/rm-rotate-auth-generated
  Normal   Starting                                               3m50s  KubeDB Ops-manager Operator  Pausing RabbitMQ databse: demo/rabbitmq
  Normal   Successful                                             3m50s  KubeDB Ops-manager Operator  Successfully paused RabbitMQ database: demo/rabbitmq for RabbitMQOpsRequest: rm-rotate-auth-generated
  Warning  reconcile; ConditionStatus:True                        3m42s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        3m42s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        3m42s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        3m42s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdateCredential                                       3m42s  KubeDB Ops-manager Operator  Successfully generated new credentials
  Warning  reconcile; ConditionStatus:True                        3m37s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        3m37s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        3m35s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                          3m35s  KubeDB Ops-manager Operator  successfully reconciled the rabbitmq with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:rabbitmq-0      3m30s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rabbitmq-0
  Warning  evict pod; ConditionStatus:True; PodName:rabbitmq-0    3m30s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rabbitmq-0
  Warning  running pod; ConditionStatus:False                     3m25s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:rabbitmq-0  3m20s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:rabbitmq-0
  Warning  get pod; ConditionStatus:True; PodName:rabbitmq-1      3m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rabbitmq-1
  Warning  evict pod; ConditionStatus:True; PodName:rabbitmq-1    3m15s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rabbitmq-1
  Warning  running pod; ConditionStatus:False                     3m10s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:rabbitmq-1  3m5s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:rabbitmq-1
  Warning  get pod; ConditionStatus:True; PodName:rabbitmq-2      3m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rabbitmq-2
  Warning  evict pod; ConditionStatus:True; PodName:rabbitmq-2    3m     KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rabbitmq-2
  Warning  running pod; ConditionStatus:False                     2m55s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:rabbitmq-2  2m50s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:rabbitmq-2
  Normal   RestartNodes                                           2m45s  KubeDB Ops-manager Operator  Successfully restarted all nodes

```
**Verify Auth is rotated**
```shell
$ kubectl get rm -n demo rabbitmq -ojson | jq .spec.authsecret.name
"rabbitmq-auth"
$ kubectl get secret -n demo rabbitmq-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎                                        
$ kubectl get secret -n demo rabbitmq-auth -o jsonpath='{.data.password}' | base64 -d
tB7;0ATxvhxeau15⏎                                            
```
Let's verify if we can connect to the database using the new credentials.

```shell
$ kubectl exec -it -n demo rabbitmq-0 -c rabbitmq -- bash

rabbitmq-0:/$ rabbitmqadmin -u admin -p 'tB7;0ATxvhxeau15' list queues
+---------------+----------+
|     name      | messages |
+---------------+----------+
| kubedb_system | 0        |
+---------------+----------+
rabbitmq-0:/$ 

```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo rabbitmq-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
admin⏎                                  
$ kubectl get secret -n demo rabbitmq-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
4TC.R7hXc1g;kA)P⏎                                             
```
Now verify whether the previous credential is workable or not 

```shell
$ kubectl exec -it -n demo rabbitmq-0 -c rabbitmq -- bash

rabbitmq-0:/$ rabbitmqadmin -u admin -p '4TC.R7hXc1g;kA)P' list queues
*** Access refused: /api/queues?columns=name,messages
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

```shell
$ kubectl create secret generic rm-auth-user -n demo \
                                               --type=kubernetes.io/basic-auth \
                                               --from-literal=username=rabbit \
                                               --from-literal=password=RabbitMQ2
secret/rm-auth-user created

```
Now create a `RabbitMQOpsRequest` with `RotateAuth` type. Below is the YAML of the `RabbitMQOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rmops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: rabbitmq
  authentication:
    secretRef:
      name: rm-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `rabbitmq`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on RabbitMQ.
- `spec.authentication.secretRef.name` specifies that we used `rm-auth-user` for database authentication.


Let's create the `RabbitMQOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/rabbitmq/rotate-auth/rotate-auth-user.yaml
RabbitMQopsrequest.ops.kubedb.com/rmops-rotate-auth-user created
```
Let’s wait for `RabbitMQOpsRequest` to be Successful. Run the following command to watch `RabbitMQOpsRequest` CR:

```shell
$ kubectl get RabbitMQopsrequest -n demo
NAME                       TYPE         STATUS       AGE
rm-rotate-auth-generated   RotateAuth   Successful   28m
rmops-rotate-auth-user     RotateAuth   Successful   80s
```
We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe RabbitMQopsrequest -n demo rmops-rotate-auth-user 
Name:         rmops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
Metadata:
  Creation Timestamp:  2025-08-15T15:19:14Z
  Generation:          1
  Resource Version:    37048
  UID:                 8bcf2459-3bc5-41f5-9ca4-7ccebdfd38bc
Spec:
  Apply:  IfReady
  Authentication:
    secret Ref:
      Name:  rm-auth-user
  Database Ref:
    Name:   rabbitmq
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-15T15:19:24Z
    Message:               rabbitmq ops request has started to rotate auth for rmq nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-15T15:19:29Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-15T15:19:29Z
    Message:               Successfully referenced the user provided authsecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-15T15:19:37Z
    Message:               successfully reconciled the rabbitmq with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-15T15:19:42Z
    Message:               get pod; ConditionStatus:True; PodName:rabbitmq-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rabbitmq-0
    Last Transition Time:  2025-08-15T15:19:42Z
    Message:               evict pod; ConditionStatus:True; PodName:rabbitmq-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rabbitmq-0
    Last Transition Time:  2025-08-15T15:19:47Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-15T15:19:52Z
    Message:               running pod; ConditionStatus:True; PodName:rabbitmq-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--rabbitmq-0
    Last Transition Time:  2025-08-15T15:19:57Z
    Message:               get pod; ConditionStatus:True; PodName:rabbitmq-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rabbitmq-1
    Last Transition Time:  2025-08-15T15:19:57Z
    Message:               evict pod; ConditionStatus:True; PodName:rabbitmq-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rabbitmq-1
    Last Transition Time:  2025-08-15T15:20:07Z
    Message:               running pod; ConditionStatus:True; PodName:rabbitmq-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--rabbitmq-1
    Last Transition Time:  2025-08-15T15:20:12Z
    Message:               get pod; ConditionStatus:True; PodName:rabbitmq-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rabbitmq-2
    Last Transition Time:  2025-08-15T15:20:12Z
    Message:               evict pod; ConditionStatus:True; PodName:rabbitmq-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rabbitmq-2
    Last Transition Time:  2025-08-15T15:20:22Z
    Message:               running pod; ConditionStatus:True; PodName:rabbitmq-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--rabbitmq-2
    Last Transition Time:  2025-08-15T15:20:27Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-15T15:20:27Z
    Message:               Successfuly completed reconfigure rmq
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                 Age   From                         Message
  ----     ------                                                 ----  ----                         -------
  Normal   Starting                                               105s  KubeDB Ops-manager Operator  Start processing for RabbitMQOpsRequest: demo/rmops-rotate-auth-user
  Warning  reconcile; ConditionStatus:True                        100s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        100s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        100s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        100s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        100s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        100s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdateCredential                                       100s  KubeDB Ops-manager Operator  Successfully referenced the user provided authsecret
  Warning  reconcile; ConditionStatus:True                        95s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        95s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        92s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                          92s   KubeDB Ops-manager Operator  successfully reconciled the rabbitmq with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:rabbitmq-0      87s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rabbitmq-0
  Warning  evict pod; ConditionStatus:True; PodName:rabbitmq-0    87s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rabbitmq-0
  Warning  running pod; ConditionStatus:False                     82s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:rabbitmq-0  77s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:rabbitmq-0
  Warning  get pod; ConditionStatus:True; PodName:rabbitmq-1      72s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rabbitmq-1
  Warning  evict pod; ConditionStatus:True; PodName:rabbitmq-1    72s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rabbitmq-1
  Warning  running pod; ConditionStatus:False                     67s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:rabbitmq-1  62s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:rabbitmq-1
  Warning  get pod; ConditionStatus:True; PodName:rabbitmq-2      57s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rabbitmq-2
  Warning  evict pod; ConditionStatus:True; PodName:rabbitmq-2    57s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rabbitmq-2
  Warning  running pod; ConditionStatus:False                     52s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:rabbitmq-2  47s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:rabbitmq-2
  Normal   RestartNodes                                           42s   KubeDB Ops-manager Operator  Successfully restarted all nodes

```
**Verify auth is rotate**
```shell
$ kubectl get rm -n demo rabbitmq -ojson | jq .spec.authsecret.name
"rm-auth-user"
$ kubectl get secret -n demo rm-auth-user -o=jsonpath='{.data.username}' | base64 -d
rabbit⏎                              
$ kubectl get secret -n demo rm-auth-user -o=jsonpath='{.data.password}' | base64 -d
RabbitMQ2⏎                                                                                         
```

Let's verify if we can connect to the database using the new credentials.
```shell
$  kubectl exec -it -n demo rabbitmq-0 -c rabbitmq -- bash

rabbitmq-0:/$ rabbitmqadmin -u rabbit -p 'RabbitMQ2' list queues
+---------------+----------+
|     name      | messages |
+---------------+----------+
| kubedb_system | 0        |
+---------------+----------+

```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo rm-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
tB7;0ATxvhxeau15⏎           
$ kubectl get secret -n demo rm-auth-user -o go-template='{{ index .data "username.prev" }}' | base64 -d
admin⏎                           
```
Let's confirm that the previous credentials no longer work.
```shell
$ kubectl exec -it -n demo rabbitmq-0 -c rabbitmq -- bash

rabbitmq-0:/$ rabbitmqadmin -u admin -p 'tB7;0ATxvhxeau15' list queues
*** Access refused: /api/queues?columns=name,messages

```
The above output shows that the credential has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete RabbitMQopsrequest rm-rotate-auth-generated rmops-rotate-auth-user -n demo
RabbitMQopsrequest.ops.kubedb.com "rm-rotate-auth-generated" "rmops-rotate-auth-user" deleted
$ kubectl delete secret -n rm-auth-user
secret "rm-auth-user" deleted
$ kubectl delete secret -n demo   rabbitmq-auth 
secret "rabbitmq-auth " deleted

```

## Next Steps


- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
