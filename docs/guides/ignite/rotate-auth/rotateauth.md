---
title: Rotate Authentication Ignite
menu:
  docs_{{ .version }}:
    identifier: guides-ignite-rotate-authentication
    name: Guide
    parent: guides-ignite-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Rotate Authentication of Ignite

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate an `Ignite` user's authentication credentials using an `IgniteOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `password` and then reference this secret in the `IgniteOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Ignite](/docs/guides/ignite/concepts/ignite.md)
    - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Create an Ignite database

KubeDB implements an `Ignite` CRD to define the specification of an Ignite database. Below is the `Ignite` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-quickstart
  namespace: demo
spec:
  replicas: 3
  version: 2.18.0
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/ignite/quickstart/examples/ignite-quickstart.yaml
ignite.kubedb.com/ignite-quickstart created
```

Now, wait until ignite-quickstart has status Ready. i.e,

```shell
$ kubectl get ignite -n demo -w
NAME            VERSION   STATUS   AGE
ignite-quickstart   2.18.0    Ready    30m
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get ignite -n demo ignite-quickstart -ojson | jq .spec.authSecret.name
"ignite-quickstart-auth"
$  kubectl get secret -n demo ignite-quickstart-auth -o=jsonpath='{.data.username}' | base64 -d
ignite⏎                                                                                    
$ kubectl get secret -n demo ignite-quickstart-auth -o=jsonpath='{.data.password}' | base64 -d
EO3AhW7uypPxPsQQ⏎             
````

## Create RotateAuth IgniteOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the `Ignite` using operator generated credentials, we have to create an
`IgniteOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `IgniteOpsRequest` CRO that we
are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: ignite-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on
`ignite-quickstart` database.
- `spec.type` specifies that we are performing `RotateAuth` on Ignite.

Let's create the `IgniteOpsRequest` CR we have shown above,
```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/ignite/rotate-auth/overview/examples/Ignite-rotate-auth-generated.yaml
igniteopsrequest.ops.kubedb.com/igops-rotate-auth-generated created
```
Let's wait for `IgniteOpsRequest` to be `Successful`. Run the following command to watch `IgniteOpsRequest` CRO
```shell
$ kubectl get igniteopsrequest -n demo
NAME                          TYPE         STATUS       AGE
igops-rotate-auth-generated   RotateAuth   Successful   7m7s
```
If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe igniteopsrequest -n demo igops-rotate-auth-generated
Name:         igops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2026-06-11T08:25:20Z
  Generation:          1
  Resource Version:    246724
  UID:                 9190f15b-eaf0-4418-b24f-52209e56c55b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:       ignite-quickstart
  Max Retries:  1
  Timeout:      5m
  Type:         RotateAuth
Status:
  Conditions:
    Last Transition Time:  2026-06-11T08:25:20Z
    Message:               Ignite ops-request has started to rotate auth for Ignite nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2026-06-11T08:25:23Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2026-06-11T08:25:35Z
    Message:               successfully reconciled the Ignite with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-11T08:28:03Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-11T08:25:48Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-0
    Last Transition Time:  2026-06-11T08:25:48Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-0
    Last Transition Time:  2026-06-11T08:25:53Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2026-06-11T08:26:28Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-0
    Last Transition Time:  2026-06-11T08:26:33Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-1
    Last Transition Time:  2026-06-11T08:26:33Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-1
    Last Transition Time:  2026-06-11T08:27:13Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-1
    Last Transition Time:  2026-06-11T08:27:18Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-2
    Last Transition Time:  2026-06-11T08:27:18Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-2
    Last Transition Time:  2026-06-11T08:27:58Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-2
    Last Transition Time:  2026-06-11T08:28:03Z
    Message:               Successfully completed reconfigure Ignite
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age    From                         Message
  ----     ------                                                          ----   ----                         -------
  Normal   Starting                                                        7m30s  KubeDB Ops-manager Operator  Pausing Ignite databse: demo/ignite-quickstart
  Normal   Successful                                                      7m30s  KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-generated
  Normal   UpdatePetSets                                                   7m17s  KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated version
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-0      7m4s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0    7m4s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  running pod; ConditionStatus:False                              6m59s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-0  6m24s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-1      6m19s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-1    6m19s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  running pod; ConditionStatus:False                              6m14s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-1  5m39s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-2      5m34s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-2    5m34s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Warning  running pod; ConditionStatus:False                              5m29s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-2  4m54s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Normal   RestartNodes                                                    4m49s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                        4m49s  KubeDB Ops-manager Operator  Resuming Ignite database: demo/ignite-quickstart
  Normal   Successful                                                      4m49s  KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-generated
```

**Verify Auth is rotated**
```shell
$ kubectl get ignite -n demo ignite-quickstart -ojson | jq .spec.authSecret.name
"ignite-quickstart-auth"
$ kubectl get secret -n demo ignite-quickstart-auth -o=jsonpath='{.data.username}' | base64 -d
ignite⏎                                                                                    
$ kubectl get secret -n demo ignite-quickstart-auth -o=jsonpath='{.data.password}' | base64 -d
is0x8KNq6hGFfmvU⏎                
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo ignite-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
ignite⏎                                                                                    
$ kubectl get secret -n demo ignite-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
EO3AhW7uypPxPsQQ⏎          
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.

#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

> Note: You cannot change the database `username`, but you can update the `password` while keeping the existing `username`.

```shell
$ kubectl create secret generic ignite-quickstart-auth-user -n demo \
   --type=kubernetes.io/basic-auth \
   --from-literal=username=ignite \
   --from-literal=password=testpassword
secret/ignite-quickstart-auth-user created
```
Now create an `IgniteOpsRequest` with `RotateAuth` type. Below is the YAML of the `IgniteOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: ignite-quickstart
  authentication:
    secretRef:
      kind: Secret
      name: ignite-quickstart-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `ignite-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on `Ignite`.
- `spec.authentication.secretRef.name` specifies that we are using `ignite-quickstart-auth-user` as `spec.authSecret.name` for authentication.

Let's create the `IgniteOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/ignite/rotate-auth/overview/examples/rotate-auth-user.yaml
igniteopsrequest.ops.kubedb.com/igops-rotate-auth-user created
```
Let's wait for `IgniteOpsRequest` to be Successful. Run the following command to watch `IgniteOpsRequest` CRO:

```shell
$ kubectl get igniteopsrequest -n demo
NAME                          TYPE         STATUS       AGE
igops-rotate-auth-generated   RotateAuth   Successful   15m
igops-rotate-auth-user        RotateAuth   Successful   4m34s
```
We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe igniteopsrequest -n demo igops-rotate-auth-user
Name:         igops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2026-06-11T08:36:04Z
  Generation:          1
  Resource Version:    247262
  UID:                 68890c94-59d4-4d7d-ba4b-b5bf502cd998
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      API Group:  
      Kind:       Secret
      Name:       ignite-quickstart-auth-user
  Database Ref:
    Name:       ignite-quickstart
  Max Retries:  1
  Timeout:      5m
  Type:         RotateAuth
Status:
  Conditions:
    Last Transition Time:  2026-06-11T08:36:04Z
    Message:               Ignite ops-request has started to rotate auth for Ignite nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2026-06-11T08:36:07Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2026-06-11T08:36:19Z
    Message:               successfully reconciled the Ignite with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-11T08:38:47Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-11T08:36:32Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-0
    Last Transition Time:  2026-06-11T08:36:32Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-0
    Last Transition Time:  2026-06-11T08:36:37Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2026-06-11T08:37:12Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-0
    Last Transition Time:  2026-06-11T08:37:17Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-1
    Last Transition Time:  2026-06-11T08:37:17Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-1
    Last Transition Time:  2026-06-11T08:37:57Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-1
    Last Transition Time:  2026-06-11T08:38:02Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-2
    Last Transition Time:  2026-06-11T08:38:02Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-2
    Last Transition Time:  2026-06-11T08:38:42Z
    Message:               running pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--ignite-quickstart-2
    Last Transition Time:  2026-06-11T08:38:47Z
    Message:               Successfully completed reconfigure Ignite
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age    From                         Message
  ----     ------                                                          ----   ----                         -------
  Normal   Starting                                                        4m53s  KubeDB Ops-manager Operator  Pausing Ignite databse: demo/ignite-quickstart
  Normal   Successful                                                      4m53s  KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-user
  Normal   UpdatePetSets                                                   4m40s  KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated version
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-0      4m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0    4m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  running pod; ConditionStatus:False                              4m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-0  3m47s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-1      3m42s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-1    3m42s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  running pod; ConditionStatus:False                              3m37s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-1  3m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-2      2m57s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-2    2m57s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Warning  running pod; ConditionStatus:False                              2m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:ignite-quickstart-2  2m17s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Normal   RestartNodes                                                    2m12s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                        2m12s  KubeDB Ops-manager Operator  Resuming Ignite database: demo/ignite-quickstart
  Normal   Successful                                                      2m12s  KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/ignite-quickstart for IgniteOpsRequest: igops-rotate-auth-user
```

**Verify auth is rotate**
```shell
$ kubectl get ignite -n demo ignite-quickstart -ojson | jq .spec.authSecret.name
"ignite-quickstart-auth-user"
$ kubectl get secret -n demo ignite-quickstart-auth-user -o=jsonpath='{.data.username}' | base64 -d
ignite⏎                                                                                    
$ kubectl get secret -n demo ignite-quickstart-auth-user -o=jsonpath='{.data.password}' | base64 -d
testpassword⏎                
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo ignite-quickstart-auth-user -o go-template='{{ index .data "username.prev" }}' | base64 -d
ignite⏎
$ kubectl get secret -n demo ignite-quickstart-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
is0x8KNq6hGFfmvU⏎            
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete igniteopsrequest igops-rotate-auth-generated igops-rotate-auth-user -n demo
igniteopsrequest.ops.kubedb.com "igops-rotate-auth-generated" deleted
igniteopsrequest.ops.kubedb.com "igops-rotate-auth-user" deleted
$ kubectl delete secret -n demo ignite-quickstart-auth-user
secret "ignite-quickstart-auth-user" deleted
$ kubectl delete secret -n demo ignite-quickstart-auth
secret "ignite-quickstart-auth" deleted
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
