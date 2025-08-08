---
title: Redis Rotate Authentication
menu:
  docs_{{ .version }}:
    identifier: rd-rotateauth
    name: Guide
    parent: rd-rotateauth-redis
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of Redis

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Redis` user's authentication credentials using a `RedisOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential..
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `username` and `password`, and then reference this Secret in the `RedisOpsRequest`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.Redis=true` to ensure Redis CRD installation.

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY       VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION              AGE
  standard (default)   rancher.io/local-path      Delete          WaitForFirstConsumer           false                      4h
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create namespace demo
  namespace/demo created

  $ kubectl get namespaces
  NAME          STATUS    AGE
  demo          Active    10s
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a Redisdatabase

KubeDB implements a `Redis` CRD to define the specification of a Redis server. Below is the `Redis` object created in this tutorial.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-quickstart
  namespace: demo
spec:
  version: 6.2.14
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/quickstart/demo-v1.yaml
redis.kubedb.com/redis-quickstart created
```

Now, wait until redis-quickstart has status Ready. i.e,

```shell
$ kubectl get rd -n demo -w
NAME               VERSION   STATUS   AGE
redis-quickstart   6.2.14    Ready    74m
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get rd -n demo redis-quickstart -ojson | jq .spec.authSecret.name
"redis-quickstart-auth"
$ kubectl get secret -n demo redis-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
default⏎                               
$ kubectl get secret -n demo redis-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
nAiZo)pGW1f!se*2⏎            
````

## Create RotateAuth RedisOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Redis using operator generated, we have to create a `RedisOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `RedisOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rdops-rotate-auth-generated
  namespace: pool
spec:
  type: RotateAuth
  databaseRef:
    name: redis-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `redis-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Redis.

Let's create the `RedisOpsRequest` CR we have shown above,
```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/redis/rotate-auth/Redis-rotate-auth-generated.yaml
redisopsrequest.ops.kubedb.com/rdops-rotate-auth-generated created
```
Let's wait for `RedisOpsrequest` to be `Successful`. Run the following command to watch `RedisOpsrequest` CRO
```shell
 $ kubectl get rdops -n demo -w
 NAME                          TYPE         STATUS       AGE
rdops-rotate-auth-generated   RotateAuth   Successful    45s
```
If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Redisopsrequest -n demo rdops-rotate-auth-generated 
Name:         rdops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T11:35:14Z
  Generation:          1
  Resource Version:    135839
  UID:                 1c5e54b0-0f7d-4942-b569-61bbdb4c3f01
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   redis-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-18T11:35:14Z
    Message:               Redis ops request has started to rotate auth for Redis
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-18T11:35:17Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-07-18T11:35:17Z
    Message:               Successfully Generated New Auth Secret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-18T11:35:17Z
    Message:               Successfully patched config secret
    Observed Generation:   1
    Reason:                patchedSecret
    Status:                True
    Type:                  patchedSecret
    Last Transition Time:  2025-07-18T11:35:19Z
    Message:               Successfully updated petsets
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-18T11:35:24Z
    Message:               evict pod; ConditionStatus:True; PodName:redis-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--redis-quickstart-0
    Last Transition Time:  2025-07-18T11:35:59Z
    Message:               is pod ready; ConditionStatus:True; PodName:redis-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--redis-quickstart-0
    Last Transition Time:  2025-07-18T11:35:59Z
    Message:               Successfully Restarted Pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-07-18T11:35:59Z
    Message:               Successfully Rotated Auth
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                           Age    From                         Message
  ----     ------                                                           ----   ----                         -------
  Normal   PauseDatabase                                                    4m12s  KubeDB Ops-manager Operator  Pausing Redis demo/redis-quickstart
  Warning  evict pod; ConditionStatus:True; PodName:redis-quickstart-0      4m2s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:redis-quickstart-0
  Warning  is pod ready; ConditionStatus:False; PodName:redis-quickstart-0  4m2s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False; PodName:redis-quickstart-0
  Warning  is pod ready; ConditionStatus:True; PodName:redis-quickstart-0   3m27s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:redis-quickstart-0
  Normal   RestartPods                                                      3m27s  KubeDB Ops-manager Operator  Successfully Restarted Pods
  Normal   ResumeDatabase                                                   3m27s  KubeDB Ops-manager Operator  Resuming Redis demo/redis-quickstart
  Normal   ResumeDatabase                                                   3m27s  KubeDB Ops-manager Operator  Successfully resumed Redis demo/redis-quickstart
  Normal   Successful                                                       3m27s  KubeDB Ops-manager Operator  Succesfully Rotated Auth for Redis

```
**Verify Auth is rotated**
```shell
$ kubectl get rd -n demo redis-quickstart -ojson | jq .spec.authSecret.name
"redis-quickstart-auth"
$ kubectl get secret -n demo redis-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
default⏎                         
$ kubectl get secret -n demo redis-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
jGPx0DDKaOb6hWAb⏎                                      
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo redis-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
default⏎       
$ kubectl get secret -n demo redis-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
nAiZo)pGW1f!se*2⏎                                               
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

```shell
$ kubectl create secret generic redis-quickstart-user-auth -n demo \
                                                --type=kubernetes.io/basic-auth \
                                                --from-literal=username=admin \
                                                --from-literal=password=Redis-secret
 secret/redis-quickstart-user-auth created
```
Now create a `RedisOpsRequest` with `RotateAuth` type. Below is the YAML of the `RedisOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rdops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: redis-quickstart
  authentication:
    secretRef:
      name: redis-quickstart-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `redis-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Redis.
- `spec.authentication.secretRef.name` specifies that we are using `redis-quickstart-user-auth` as `spec.authSecret.name` for authentication.

Let's create the `RedisOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/Redis/rotate-auth/rotate-auth-user.yaml
redisopsrequest.ops.kubedb.com/rdops-rotate-auth-user created
```
Let’s wait for `RedisOpsRequest` to be Successful. Run the following command to watch `RedisOpsRequest` CRO:

```shell
$ kubectl get rdops -n demo -w
NAME                          TYPE         STATUS       AGE
rdops-rotate-auth-generated   RotateAuth   Successful   6m39s
rdops-rotate-auth-user        RotateAuth   Successful   46s
```
We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Redisopsrequest -n demo rdops-rotate-auth-user
Name:         rdops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T11:41:07Z
  Generation:          1
  Resource Version:    136344
  UID:                 aa4f7fa3-4047-452f-ad1b-d39ef738c0f2
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  redis-quickstart-user-auth
  Database Ref:
    Name:   redis-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-18T11:41:07Z
    Message:               Redis ops request has started to rotate auth for Redis
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-18T11:41:10Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-07-18T11:41:10Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-18T11:41:10Z
    Message:               Successfully patched config secret
    Observed Generation:   1
    Reason:                patchedSecret
    Status:                True
    Type:                  patchedSecret
    Last Transition Time:  2025-07-18T11:41:12Z
    Message:               Successfully updated petsets
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-18T11:41:17Z
    Message:               evict pod; ConditionStatus:True; PodName:redis-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--redis-quickstart-0
    Last Transition Time:  2025-07-18T11:41:52Z
    Message:               is pod ready; ConditionStatus:True; PodName:redis-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--redis-quickstart-0
    Last Transition Time:  2025-07-18T11:41:52Z
    Message:               Successfully Restarted Pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-07-18T11:41:53Z
    Message:               Successfully Rotated Auth
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                           Age   From                         Message
  ----     ------                                                           ----  ----                         -------
  Normal   PauseDatabase                                                    115s  KubeDB Ops-manager Operator  Pausing Redis demo/redis-quickstart
  Warning  evict pod; ConditionStatus:True; PodName:redis-quickstart-0      105s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:redis-quickstart-0
  Warning  is pod ready; ConditionStatus:False; PodName:redis-quickstart-0  105s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False; PodName:redis-quickstart-0
  Warning  is pod ready; ConditionStatus:True; PodName:redis-quickstart-0   70s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:redis-quickstart-0
  Normal   RestartPods                                                      70s   KubeDB Ops-manager Operator  Successfully Restarted Pods
  Normal   ResumeDatabase                                                   69s   KubeDB Ops-manager Operator  Resuming Redis demo/redis-quickstart
  Normal   ResumeDatabase                                                   69s   KubeDB Ops-manager Operator  Successfully resumed Redis demo/redis-quickstart
  Normal   Successful                                                       69s   KubeDB Ops-manager Operator  Succesfully Rotated Auth for Redis

```
**Verify auth is rotate**
```shell
$ kubectl get rd -n demo redis-quickstart -ojson | jq .spec.authSecret.name
"redis-quickstart-user-auth"
$kubectl get secret -n demo redis-quickstart-user-auth -o=jsonpath='{.data.username}' | base64 -d
admin⏎           
$ kubectl get secret -n demo redis-quickstart-user-auth -o=jsonpath='{.data.password}' | base64 -d
Redis-secret⏎                                                                                 
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo redis-quickstart-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
default⏎        
$ kubectl get secret -n demo redis-quickstart-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
jGPx0DDKaOb6hWAb⏎              
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Redisopsrequest rdops-rotate-auth-generated rdops-rotate-auth-user -n demo
Redisopsrequest.ops.kubedb.com "rdops-rotate-auth-generated" "rdops-rotate-auth-user" deleted
$ kubectl delete secret -n demo  redis-quickstart-user-auth
secret "redis-quickstart-user-auth" deleted
$ kubectl delete secret -n demo  redis-quickstart-auth
secret "redis-quickstart-auth" deleted
```


## Next Steps

- Learn how to use KubeDB to run a Redis server [here](/docs/guides/redis/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).