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
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 2Gi
    cluster:
      - name: appscode-cluster
        shards: 2
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: clickhouse
                resources:
                  limits:
                    memory: 4Gi
                  requests:
                    cpu: 500m
                    memory: 2Gi
            initContainers:
              - name: clickhouse-init
                resources:
                  limits:
                    memory: 1Gi
                  requests:
                    cpu: 500m
                    memory: 1Gi
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
➤ kubectl get clickhouse -n demo -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   37s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   49s
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m1s
```

We can verify from the above output that authentication is enabled for this cluster. By default, KubeDB operator create default credentials for the ClickHouse cluster. The default credentials are stored in a secret named `<clickhouse-name>-auth` in the same namespace as the ClickHouse cluster. You can find the secret by running the following command:

```bash
➤ kubectl get ch -n demo clickhouse-prod -ojson | jq .spec.authSecret.name
"clickhouse-prod-auth"
➤ kubectl get secret -n demo clickhouse-prod-auth -o=jsonpath='{.data.username}' | base64 -d
admin                                                                     
➤ kubectl get secret -n demo clickhouse-prod-auth -o=jsonpath='{.data.password}' | base64 -d
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
    name: clickhouse-prod
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
➤ kubectl get clickhouseopsrequest -n demo 
NAME                          TYPE         STATUS       AGE
chops-rotate-auth-generated   RotateAuth   Successful   3m19s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
➤ kubectl describe  chops -n demo chops-rotate-auth-generated 
Name:         chops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T09:47:04Z
  Generation:          1
  Resource Version:    798956
  UID:                 e3495f4f-a0a2-4cbb-bc8a-67bca41381ad
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-25T09:47:04Z
    Message:               ClickHouse ops-request has started to rotate auth for clickhouse nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-25T09:47:16Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-25T09:47:21Z
    Message:               successfully reconciled the ClickHouse with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T09:47:21Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T09:50:01Z
    Message:               Successfully restarted all pods
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-25T09:47:26Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T09:47:26Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T09:47:31Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-25T09:48:01Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T09:48:01Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T09:48:41Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T09:48:41Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T09:49:21Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T09:49:21Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T09:50:01Z
    Message:               Successfully completed reconfigure clickhouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age    From                         Message
  ----     ------                                                                               ----   ----                         -------
  Normal   Starting                                                                             4m     KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-rotate-auth-generated
  Normal   Starting                                                                             4m     KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           4m     KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-rotate-auth-generated
  Warning  reconcile; ConditionStatus:True                                                      3m43s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      3m43s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      3m43s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        3m43s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with updated version
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    3m38s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  3m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   3m33s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    3m3s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  3m3s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    2m23s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  2m23s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    103s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  103s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         63s    KubeDB Ops-manager Operator  Successfully restarted all pods
  Normal   Starting                                                                             63s    KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           63s    KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-rotate-auth-generated

```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
➤ kubectl get ch -n demo clickhouse-prod -ojson | jq .spec.authSecret.name
"clickhouse-prod-auth"
➤ kubectl get secret -n demo clickhouse-prod-auth -o=jsonpath='{.data.username}' | base64 -d
admin⏎                                                                    
➤ kubectl get secret -n demo clickhouse-prod-auth -o=jsonpath='{.data.password}' | base64 -d
qj4zcj2JEPGJDpf5
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
➤ kubectl get secret -n demo clickhouse-prod-auth -o=jsonpath="{.data.username\.prev}" | base64 -d
admin⏎                                                                    
➤ kubectl get secret -n demo clickhouse-prod-auth -o=jsonpath="{.data.password\.prev}" | base64 -d
St9402lDFuk9LgDo
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
    name: ch
  authentication:
    secretRef:
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
➤ kubectl get clickhouseopsrequest -n demo chops-rotate-auth-user 
NAME                     TYPE         STATUS       AGE
chops-rotate-auth-user   RotateAuth   Successful   4m1s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-rotate-auth-user 
Name:         chops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T09:58:14Z
  Generation:          1
  Resource Version:    800494
  UID:                 bad01b6c-8799-4b76-85c2-9ec3220abe2e
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  clickhouse-user-auth
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-25T09:58:14Z
    Message:               ClickHouse ops-request has started to rotate auth for clickhouse nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-25T09:58:17Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-25T09:58:22Z
    Message:               successfully reconciled the ClickHouse with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T09:58:22Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T10:00:22Z
    Message:               Successfully restarted all pods
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-25T09:58:27Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T09:58:27Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T09:58:32Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-25T09:58:57Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T09:58:57Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T09:59:17Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T09:59:17Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T09:59:42Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T09:59:42Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T10:00:22Z
    Message:               Successfully completed reconfigure clickhouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age    From                         Message
  ----     ------                                                                               ----   ----                         -------
  Normal   Starting                                                                             4m43s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-rotate-auth-user
  Normal   Starting                                                                             4m43s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           4m43s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-rotate-auth-user
  Warning  reconcile; ConditionStatus:True                                                      4m35s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      4m35s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      4m35s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        4m35s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with updated version
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    4m30s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  4m30s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   4m25s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    4m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  4m     KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    3m40s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  3m40s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    3m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  3m15s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         2m35s  KubeDB Ops-manager Operator  Successfully restarted all pods
  Normal   Starting                                                                             2m35s  KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           2m35s  KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-rotate-auth-user
```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
➤ kubectl get ch -n demo clickhouse-prod -ojson | jq .spec.authSecret.name
"clickhouse-user-auth"
➤ kubectl get secret -n demo clickhouse-user-auth -o=jsonpath='{.data.username}' | base64 -d
clickhouse                                                                
➤ kubectl get secret -n demo clickhouse-user-auth -o=jsonpath='{.data.password}' | base64 -d
clickhouse-secret
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
➤ kubectl get secret -n demo clickhouse-user-auth -o=jsonpath="{.data.username\.prev}" | base64 -d
admin⏎                                                                   
➤ kubectl get secret -n demo clickhouse-user-auth -o=jsonpath="{.data.password\.prev}" | base64 -d
qj4zcj2JEPGJDpf5
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete clickhouseopsrequests -n demo chops-rotate-auth-generated chops-rotate-auth-user
kubectl delete clickhouse -n demo clickhouse-prod
kubectl delete secret -n demo clickhouse-user-auth
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

