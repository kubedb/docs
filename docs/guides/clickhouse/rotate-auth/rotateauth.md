---
title: ClickHouse Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: ch-rotate-auth-clickhouse
    name: ClickHouse RotateAuth Guide
    parent: ch-rotate-auth
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate ClickHouse Authentication

KubeDB supports rotating Authentication for existing ClickHouse via a ClickHouseOpsRequest. There are two ways to do that.
1. **Operator Generated**: User will not provide any secret. KubeDB operator will generate a random password and update the existing secret with that password.
2. **User Defined**: User can create a `kubernetes.io/basic-auth` type secret with `username` and  `password` and refers this to `ClickHouseOpsRequest`.

This tutorial will show you how to use KubeDB to rotate authentication credentials.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/clickhouse](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/clickhouse) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Create ClickHouse with Enabling Authentication

In this section, we are going to deploy a ClickHouse cluster with authentication enabled. In the next few sections we will rotate the authentication using `ClickHouseOpsRequest` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse
  namespace: demo
spec:
  version: 25.7.1
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/rotate-auth/clickhouse-cluster.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
$ kubectl get clickhouse -n demo -w
NAME         TYPE                  VERSION   STATUS   AGE
clickhouse   kubedb.com/v1alpha2   25.7.1    Ready    25h

```

We can verify from the above output that authentication is enabled for this cluster. By default, KubeDB operator create default credentials for the ClickHouse cluster. The default credentials are stored in a secret named `<clickhouse-name>-auth` in the same namespace as the ClickHouse cluster. You can find the secret by running the following command:

```bash
$ kubectl get secrets -n demo clickhouse-auth -o jsonpath='{.data.\username}' | base64 -d
admin                                                                     
$ kubectl get secrets -n demo clickhouse-auth -o jsonpath='{.data.\password}' | base64 -d
St9402lDFuk9LgDo
```

### Create RotateAuth ClickHouseOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the clickhouse using operator generated, we have to create a `ClickHouseOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `ClickHouseOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: clickhouse
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on clickhouse.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/rotate-auth/chops-rotate-auth-generated.yaml
clickhouseopsrequest.ops.kubedb.com/chops-rotate-auth-generated created
```

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CRO,

```bash
$ kubectl get clickhouseopsrequest -n demo 
NAME                          TYPE         STATUS       AGE
chops-rotate-auth-generated   RotateAuth   Successful   5m59s

```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe  chops -n demo chops-rotate-auth-generated 
Name:         chops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-09-11T10:57:34Z
  Generation:          1
  Resource Version:    389307
  UID:                 a11bfa35-c142-4a0b-bb35-b96b490ee444
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   clickhouse
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-09-11T10:57:34Z
    Message:               ClickHouse ops-request has started to rotate auth for clickhouse nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-09-11T10:57:37Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-09-11T10:57:47Z
    Message:               successfully reconciled the ClickHouse with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-09-11T10:57:42Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-09-11T10:58:07Z
    Message:               Successfully restarted all pods
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-09-11T10:57:52Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-0
    Last Transition Time:  2025-09-11T10:57:52Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-0
    Last Transition Time:  2025-09-11T10:57:57Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-09-11T10:58:07Z
    Message:               Successfully completed reconfigure clickhouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                 Age    From                         Message
  ----     ------                                                 ----   ----                         -------
  Normal   Starting                                               7m4s   KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-rotate-auth-generated
  Normal   Starting                                               7m4s   KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse
  Normal   Successful                                             7m4s   KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse for ClickHouseOpsRequest: chops-rotate-auth-generated
  Warning  reconcile; ConditionStatus:True                        6m56s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        6m56s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        6m51s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                          6m51s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with updated version
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-0    6m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-0  6m46s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-0
  Warning  running pod; ConditionStatus:False                     6m41s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Normal   RestartNodes                                           6m31s  KubeDB Ops-manager Operator  Successfully restarted all pods
  Normal   Starting                                               6m31s  KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse
  Normal   Successful                                             6m31s  KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse for ClickHouseOpsRequest: chops-rotate-auth-generated

```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
$ kubectl get ch -n demo clickhouse -ojson | jq .spec.authSecret.name
"clickhouse-auth"
$ kubectl get secrets -n demo clickhouse-auth -o jsonpath='{.data.\username}' | base64 -d
admin⏎                              
$ kubectl get secrets -n demo clickhouse-auth -o jsonpath='{.data.\password}' | base64 -d
sG0OKmIim3ZkfhpE⏎             
```
Now, you can exec into the pod `clickhouse-0` and connect to database using `username` and `password`
```bash
$ kubectl exec -it -n demo clickhouse-0 -c clickhouse -- bash
clickhouse@clickhouse-0:/$ clickhouse-client -uadmin --password="sG0OKmIim3ZkfhpE"
ClickHouse client version 25.7.1.3997 (official build).
Connecting to localhost:9000 as user admin.
Connected to ClickHouse server version 25.7.1.

Warnings:
 * Effective user of the process (clickhouse) does not match the owner of the data (root).
 * Delay accounting is not enabled, OSIOWaitMicroseconds will not be gathered. You can enable it using `sudo sh -c 'echo 1 > /proc/sys/kernel/task_delayacct'` or by using sysctl.

clickhouse-0.clickhouse-pods.demo.svc.cluster.local :) SHOW DATABASES

SHOW DATABASES

Query id: 4b9a02da-2e28-4a48-a235-d54df0534883

   ┌─name───────────────┐
1. │ INFORMATION_SCHEMA │
2. │ default            │
3. │ information_schema │
4. │ kubedb_system      │
5. │ system             │
   └────────────────────┘

5 rows in set. Elapsed: 0.001 sec. 

clickhouse-0.clickhouse-pods.demo.svc.cluster.local :) exit
Bye.
clickhouse@clickhouse-0:/$ exit
exit


```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo clickhouse-auth -o=jsonpath="{.data.password\.prev}" | base64 -d
w5MKkyQ1PMOOC7BO⏎                                                                                
$ kubectl get secret -n demo clickhouse-auth -o=jsonpath="{.data.username\.prev}" | base64 -d
admin⏎                       
```
Let's confirm that the previous credentials no longer work.
```shell
$ kubectl exec -it -n demo clickhouse-0 -c clickhouse -- bash
clickhouse@clickhouse-0:/$ clickhouse-client -uadmin --password="w5MKkyQ1PMOOC7BO"
ClickHouse client version 25.7.1.3997 (official build).
Connecting to localhost:9000 as user admin.
Code: 516. DB::Exception: Received from localhost:9000. DB::Exception: admin: Authentication failed: password is incorrect, or there is no user with such name.. (AUTHENTICATION_FAILED)

clickhouse@clickhouse-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.

#### 2. Using user created credentials

At first, we need to create a secret with `kubernetes.io/basic-auth` type using custom `username` and `password`. Below is the command to create a secret with `kubernetes.io/basic-auth` type,

```bash
$ kubectl create secret generic clickhouse-user-auth -n demo \
          --type=kubernetes.io/basic-auth \
          --from-literal=username=clickhouse \
          --from-literal=password=clickhouse-secret
secret/clickhouse-user-auth created
```

Now create a ClickHouse Ops Request with `RotateAuth` type. Below is the YAML of the `ClickHouseOpsRequest` that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: clickhouse
  authentication:
    secretRef:
      kind: Secret
      name: clickhouse-user-auth
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on clickhouse.
- `spec.authentication.secretRef.name` specifies that we are using `clickhouse-user-auth` secret for authentication.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/rotate-auth/chops-rotate-auth-user.yaml
clickhouseopsrequest.ops.kubedb.com/chops-rotate-auth-user created
```

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CRO,

```bash
$ kubectl get clickhouseopsrequest -n demo chops-rotate-auth-user 
NAME                     TYPE         STATUS       AGE
chops-rotate-auth-user   RotateAuth   Successful    4m43s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
$kubectl describe clickhouseopsrequest -n demo chops-rotate-auth-user 
Name:         chops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-09-11T11:32:42Z
  Generation:          2
  Resource Version:    390583
  UID:                 702e4b0f-fe1b-4283-b5a1-775086d84a28
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  clickhouse-user-auth
  Database Ref:
    Name:   clickhouse
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-09-11T11:36:53Z
    Message:               ClickHouse ops-request has started to rotate auth for clickhouse nodes
    Observed Generation:   2
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-09-11T11:36:56Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   2
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-09-11T11:37:05Z
    Message:               successfully reconciled the ClickHouse with updated version
    Observed Generation:   2
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-09-11T11:37:01Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   2
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-09-11T11:37:25Z
    Message:               Successfully restarted all pods
    Observed Generation:   2
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-09-11T11:37:10Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-0
    Observed Generation:   2
    Status:                True
    Type:                  GetPod--clickhouse-0
    Last Transition Time:  2025-09-11T11:37:10Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-0
    Observed Generation:   2
    Status:                True
    Type:                  EvictPod--clickhouse-0
    Last Transition Time:  2025-09-11T11:37:15Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   2
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-09-11T11:37:25Z
    Message:               Successfully completed reconfigure clickhouse
    Observed Generation:   2
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type     Reason                                                 Age   From                         Message
  ----     ------                                                 ----  ----                         -------
  Normal   Starting                                               117s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-rotate-auth-user
  Normal   Starting                                               117s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse
  Normal   Successful                                             117s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse for ClickHouseOpsRequest: chops-rotate-auth-user
  Warning  reconcile; ConditionStatus:True                        109s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        109s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                        105s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                          105s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with updated version
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-0    100s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-0  100s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-0
  Warning  running pod; ConditionStatus:False                     95s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Normal   RestartNodes                                           85s   KubeDB Ops-manager Operator  Successfully restarted all pods
  Normal   Starting                                               85s   KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse
  Normal   Successful                                             85s   KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse for ClickHouseOpsRequest: chops-rotate-auth-user

```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
$  kubectl get ch -n demo clickhouse -ojson | jq .spec.authSecret.name
"clickhouse-user-auth"
$ kubectl get secret -n demo clickhouse-user-auth -o=jsonpath='{.data.username}' | base64 -d
clickhouse⏎      
$ kubectl get secret -n demo clickhouse-user-auth -o=jsonpath='{.data.password}' | base64 -d
clickhouse-secret⏎              
```
Now, you can exec into the pod `clickhouse-0` and connect to database using `username` and `password`
```bash
$ kubectl exec -it -n demo clickhouse-0 -c clickhouse -- bash
clickhouse@clickhouse-0:/$ clickhouse-client -uclickhouse --password="clickhouse-secret"
ClickHouse client version 25.7.1.3997 (official build).
Connecting to localhost:9000 as user clickhouse.
Connected to ClickHouse server version 25.7.1.

Warnings:
 * Effective user of the process (clickhouse) does not match the owner of the data (root).
 * Delay accounting is not enabled, OSIOWaitMicroseconds will not be gathered. You can enable it using `sudo sh -c 'echo 1 > /proc/sys/kernel/task_delayacct'` or by using sysctl.

clickhouse-0.clickhouse-pods.demo.svc.cluster.local :) show databases;

SHOW DATABASES

Query id: 2807f810-8375-47b7-80fe-0ebe3df028ad

   ┌─name───────────────┐
1. │ INFORMATION_SCHEMA │
2. │ default            │
3. │ information_schema │
4. │ kubedb_system      │
5. │ system             │
   └────────────────────┘

5 rows in set. Elapsed: 0.001 sec. 

```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo clickhouse-user-auth -o=jsonpath="{.data.username\.prev}" | base64 -d
admin⏎                                                                                                      
$  kubectl get secret -n demo clickhouse-user-auth -o=jsonpath="{.data.password\.prev}" | base64 -d
sG0OKmIim3ZkfhpE⏎                                    
```
Let's confirm that the previous credentials no longer work.
```shell
$  kubectl exec -it -n demo clickhouse-0 -c clickhouse -- bash
clickhouse@clickhouse-0:/$ clickhouse-client -uadmin --password="sG0OKmIim3ZkfhpE"
ClickHouse client version 25.7.1.3997 (official build).
Connecting to localhost:9000 as user admin.
Code: 516. DB::Exception: Received from localhost:9000. DB::Exception: admin: Authentication failed: password is incorrect, or there is no user with such name.. (AUTHENTICATION_FAILED)

clickhouse@clickhouse-0:/$ 

```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete clickhouseopsrequests -n demo chops-rotate-auth-generated chops-rotate-auth-user
$ kubectl delete clickhouse -n demo clickhouse-prod
$ kubectl delete secret -n demo clickhouse-user-auth
$ kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

