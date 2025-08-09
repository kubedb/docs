---
title: Rotate Authentication Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-rotateauth-details
    name: Guide
    parent: pp-rotateauth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---


# Rotate Authentication of Pgpool

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Pgpool` user's authentication credentials using a `PgpoolOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential..
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `username` and `password`, and then reference this Secret in the `PgpoolOpsRequest`.

> Note: Before you begin, you have to create `Pgpool` CRD with the help of [this](/docs/guides/pgpool/quickstart/quickstart.md).
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get Pgpool -n pool quick-pgpool -ojson | jq .spec.authSecret.name
"quick-pgpool-auth"
$ kubectl get secrets -n pool quick-pgpool-auth -o jsonpath='{.data.\username}' | base64 -d
pcp⏎               
$ kubectl get secrets -n pool quick-pgpool-auth -o jsonpath='{.data.\password}' | base64 -d
gZoAOjr0iUkH07ku⏎                                                  
````

## Create RotateAuth PgpoolOpsRequest


#### 1. Using operator generated credentials:

In order to rotate authentication to the Pgpool using operator generated, we have to create a `PgpoolOpsRequest` CR with `RotateAuth` type. Below is the YAML of the `PgpoolOpsRequest` CR that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind:  PgpoolOpsRequest
metadata:
  name: pgpops-rotate-auth-generated
  namespace: pool
spec:
  type: RotateAuth
  databaseRef:
    name: quick-pgpool
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `quick-pgpool` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Pgpool.

Let's create the `PgpoolOpsRequest` CR we have shown above,
```shell
 $kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/rotateauth/rotateauth.yaml
 pgpoolopsrequest.ops.kubedb.com/pgpops-rotate-auth-generated created
```
Let's wait for `PgpoolOpsrequest` to be `Successful`. Run the following command to watch `PgpoolOpsrequest` CR
```shell
 $ kubectl get PgpoolOpsRequest -n pool
NAME                           TYPE         STATUS       AGE
pgpops-rotate-auth-generated   RotateAuth   Successful   52s
```
If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe pgpoolopsrequest -n pool pgpops-rotate-auth-generated
Name:         pgpops-rotate-auth-generated
Namespace:    pool
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2025-08-01T08:48:44Z
  Generation:          1
  Resource Version:    653764
  UID:                 3822eb18-a4ab-497e-901a-1b7ddb76a516
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   quick-pgpool
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-01T08:48:44Z
    Message:               Pgpool ops request has started to rotate auth for pgpool
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-01T08:48:47Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-08-01T08:48:47Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-01T08:48:48Z
    Message:               Successfully updated petsets rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-01T08:49:33Z
    Message:               Successfully Restarted Pods With New User
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-01T08:48:53Z
    Message:               get pod; ConditionStatus:True; PodName:quick-pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--quick-pgpool-0
    Last Transition Time:  2025-08-01T08:48:53Z
    Message:               evict pod; ConditionStatus:True; PodName:quick-pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--quick-pgpool-0
    Last Transition Time:  2025-08-01T08:49:28Z
    Message:               check pod running; ConditionStatus:True; PodName:quick-pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--quick-pgpool-0
    Last Transition Time:  2025-08-01T08:49:33Z
    Message:               Successfully completed the reconfigure for Pgpool
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                            Age    From                         Message
  ----     ------                                                            ----   ----                         -------
  Normal   Starting                                                          3m21s  KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: pool/pgpops-rotate-auth-generated
  Normal   Starting                                                          3m21s  KubeDB Ops-manager Operator  Pausing Pgpool databse: pool/quick-pgpool
  Normal   Successful                                                        3m21s  KubeDB Ops-manager Operator  Successfully paused Pgpool database: pool/quick-pgpool for PgpoolOpsRequest: pgpops-rotate-auth-generated
  Normal   VersionUpdate                                                     3m18s  KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                     3m17s  KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  get pod; ConditionStatus:True; PodName:quick-pgpool-0             3m12s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:quick-pgpool-0
  Warning  evict pod; ConditionStatus:True; PodName:quick-pgpool-0           3m12s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:quick-pgpool-0
  Warning  check pod running; ConditionStatus:False; PodName:quick-pgpool-0  3m7s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:quick-pgpool-0
  Warning  check pod running; ConditionStatus:True; PodName:quick-pgpool-0   2m37s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:quick-pgpool-0
  Normal   RestartPods                                                       2m32s  KubeDB Ops-manager Operator  Successfully Restarted Pods With New User
  Normal   Starting                                                          2m32s  KubeDB Ops-manager Operator  Resuming Pgpool database: pool/quick-pgpool
  Normal   Successful                                                        2m32s  KubeDB Ops-manager Operator  Successfully resumed Pgpool database: pool/quick-pgpool for PgpoolOpsRequest: pgpops-rotate-auth-generated
```

**Verify Auth is rotated**
```shell
$  kubectl get Pgpool -n pool quick-pgpool -ojson | jq .spec.authSecret.name
"quick-pgpool-auth"
$ kubectl get secrets -n pool quick-pgpool-auth -o jsonpath='{.data.\username}' | base64 -d
pcp⏎     
$ kubectl get secrets -n pool quick-pgpool-auth -o jsonpath='{.data.\password}' | base64 -d
h1yPX0CjgGXNjpKY⏎                                                                      
```
Also, there will be two more new keys in the secret that stores the previous credentials. The key is `authData.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n pool quick-pgpool-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
pcp⏎         
$ kubectl get secret -n pool quick-pgpool-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
gZoAOjr0iUkH07ku⏎                                                              
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

```shell
$ kubectl create secret generic quick-pp-user-auth -n pool \
        --type=kubernetes.io/basic-auth \
        --from-literal=username=user \
        --from-literal=password=Pgpool2
secret/quick-pp-user-auth created
```
Now create a `PgpoolOpsRequest` with `RotateAuth` type. Below is the YAML of the `PgpoolOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: ppops-rotate-auth-user
  namespace: pool
spec:
  type: RotateAuth
  databaseRef:
    name: quick-pgpool
  authentication:
    secretRef:
      name: quick-pp-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `quick-pgpool`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Pgpool.
- `spec.authentication.secretRef.name` specifies that we are using `quick-pp-user-auth` as `spec.authSecret.name` for authentication.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```shell
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/rotateauth/rotateauthuser.yaml
Pgpoolopsrequest.ops.kubedb.com/pbops-rotate-auth-user created
```
Let’s wait for `PgpoolOpsRequest` to be Successful. Run the following command to watch `PgpoolOpsRequest` CR:

```shell
$ kubectl get PgpoolOpsRequest -n pool
NAME                           TYPE         STATUS       AGE
pgpops-rotate-auth-generated   RotateAuth   Successful   56m
ppops-rotate-auth-user         RotateAuth   Successful   44m
```
We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.
```shell
$  kubectl describe Pgpoolopsrequest -n pool ppops-rotate-auth-user
Name:         ppops-rotate-auth-user
Namespace:    pool
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2025-08-01T08:55:40Z
  Generation:          1
  Resource Version:    654393
  UID:                 5bba7694-a979-4ce0-9fcd-4e8b11287277
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  quick-pp-user-auth
  Database Ref:
    Name:   quick-pgpool
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-01T08:55:40Z
    Message:               Pgpool ops request has started to rotate auth for pgpool
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-01T08:55:43Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-08-01T08:55:43Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-01T08:55:44Z
    Message:               Successfully updated petsets rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-01T08:56:29Z
    Message:               Successfully Restarted Pods With New User
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-01T08:55:49Z
    Message:               get pod; ConditionStatus:True; PodName:quick-pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--quick-pgpool-0
    Last Transition Time:  2025-08-01T08:55:49Z
    Message:               evict pod; ConditionStatus:True; PodName:quick-pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--quick-pgpool-0
    Last Transition Time:  2025-08-01T08:56:24Z
    Message:               check pod running; ConditionStatus:True; PodName:quick-pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--quick-pgpool-0
    Last Transition Time:  2025-08-01T08:56:29Z
    Message:               Successfully completed the reconfigure for Pgpool
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                            Age    From                         Message
  ----     ------                                                            ----   ----                         -------
  Normal   Starting                                                          5m5s   KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: pool/ppops-rotate-auth-user
  Normal   Starting                                                          5m5s   KubeDB Ops-manager Operator  Pausing Pgpool databse: pool/quick-pgpool
  Normal   Successful                                                        5m5s   KubeDB Ops-manager Operator  Successfully paused Pgpool database: pool/quick-pgpool for PgpoolOpsRequest: ppops-rotate-auth-user
  Normal   VersionUpdate                                                     5m2s   KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                     5m1s   KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  get pod; ConditionStatus:True; PodName:quick-pgpool-0             4m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:quick-pgpool-0
  Warning  evict pod; ConditionStatus:True; PodName:quick-pgpool-0           4m56s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:quick-pgpool-0
  Warning  check pod running; ConditionStatus:False; PodName:quick-pgpool-0  4m51s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:quick-pgpool-0
  Warning  check pod running; ConditionStatus:True; PodName:quick-pgpool-0   4m21s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:quick-pgpool-0
  Normal   RestartPods                                                       4m16s  KubeDB Ops-manager Operator  Successfully Restarted Pods With New User
  Normal   Starting                                                          4m16s  KubeDB Ops-manager Operator  Resuming Pgpool database: pool/quick-pgpool
  Normal   Successful                                                        4m16s  KubeDB Ops-manager Operator  Successfully resumed Pgpool database: pool/quick-pgpool for PgpoolOpsRequest: ppops-rotate-auth-user

```
**Verify auth is rotate**
```shell
$ kubectl get pgpool -n pool quick-pgpool -ojson | jq .spec.authSecret.name
"quick-pp-user-auth"
$ kubectl get secrets -n pool quick-pp-user-auth -o jsonpath='{.data.\username}' | base64 -d
user⏎                                                                                       
$ kubectl get secrets -n pool quick-pp-user-auth -o jsonpath='{.data.\password}' | base64 -d
Pgpool2⏎                                                                                                                                        
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$  kubectl get secret -n pool quick-pp-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
pcp⏎                                                                                                                                                            
$ kubectl get secret -n pool quick-pp-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
gZoAOjr0iUkH07ku⏎                                
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Pgpoolopsrequest pgpops-rotate-auth-generated ppops-rotate-auth-user -n pool
Pgpoolopsrequest.ops.kubedb.com "pgpops-rotate-auth-generated" "ppops-rotate-auth-user" deleted
$ kubectl delete secret -n pool quick-pp-user-auth
secret "quick-pp-user-auth" deleted
$ kubectl delete secret -n pool   quick-pgpool-auth
secret "quick-pgpool-auth " deleted
```

## Next Steps

- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
