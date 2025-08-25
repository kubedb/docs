---
title: Rotate Authentication guide
menu:
  docs_{{ .version }}:
    identifier: mc-rotate-auth-guide
    name: guide
    parent: rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Rotate Authentication of Memcached

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Memcached`
user's authentication credentials using a `MemcachedOpsRequest`. There are two ways to perform this 
rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the 
existing secret with the new credential, and does not provide the secret details directly to the user.
2. **User Defined:** The user can create their own credentials by defining a Secret of type
`kubernetes.io/basic-auth` containing the desired `username` and `password`, and then reference this Secret in the `MemcachedOpsRequest`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME      STATUS    AGE
demo      Active    1s
```

## Find Available MemcachedVersion

When you have installed KubeDB, it has created `MemcachedVersion` crd for all supported Memcached versions. Check 0

```bash
$ kubectl get memcachedversions
NAME       VERSION   DB_IMAGE                                          DEPRECATED   AGE
1.5        1.5       ghcr.io/kubedb/memcached:1.5                      true         5d19h
1.5-v1     1.5       ghcr.io/kubedb/memcached:1.5-v1                   true         5d19h
1.5.22     1.5.22    ghcr.io/appscode-images/memcached:1.5.22-alpine                5d19h
1.5.4      1.5.4     ghcr.io/kubedb/memcached:1.5.4                    true         5d19h
1.5.4-v1   1.5.4     ghcr.io/kubedb/memcached:1.5.4-v1                 true         5d19h
1.6.22     1.6.22    ghcr.io/appscode-images/memcached:1.6.22-alpine                5d19h
1.6.29     1.6.29    ghcr.io/appscode-images/memcached:1.6.29-alpine                5d19h
1.6.33     1.6.33    ghcr.io/appscode-images/memcached:1.6.33-alpine                5d19h
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/memcached](/docs/examples/memcached/update-version) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Create a Memcached Server

KubeDB implements a `Memcached` CRD to define the specification of a Memcached server. Below is the `Memcached` object created in this tutorial.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: DoNotTerminate
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/quickstart/demo-v1.yaml
memcached.kubedb.com/memcd-quickstart created
```

Now, wait until memcd-quickstart has status Ready. i.e,

```shell
$  kubectl get mc -n demo -w
NAME               VERSION   STATUS   AGE
memcd-quickstart   1.6.22    Ready    17h
```
## Verify Authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get memcached -n demo memcd-quickstart -ojson | jq .spec.authSecret.name
"memcd-quickstart-auth"
$ kubectl get secret -n demo memcd-quickstart-auth -o=jsonpath='{.data.authData}' | base64 -d
user:ikbkjbodeewenrgj
````
Here, `username`is `user` and `password` is `ikbkjbodeewenrgj`

### Connect With Memcached Database Using Credentials

Here, we will connect to Memcached server from local-machine through port-forwarding.
We will connect to `memcd-quickstart-0` pod from local-machine using port-frowarding and it must be running in separate terminal.
```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
```
Now, you can connect to this database using `telnet`.Connect to Memcached from local-machine through telnet.
```shell

~ $ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.

version
CLIENT_ERROR unauthenticated # that means still you can not enter in DB

# Save data Command:
set my_key 0 2592000 21
# Meaning:
# 0       => no flags
# 2592000 => TTL (Time-To-Live) in [s]
# 21       => credential size in bytes

user bvkwwqxekxouudbr  # username password 

# Output:
STORED

#now you can use DB
version
VERSION 1.6.22

# Exit
quit
```
If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth MemcachedOpsRequest

#### 1. Using Operator Generated Credentials:

In order to rotate authentication to the Memcached using operator generated, we have to create a `MemcachedOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `MemcachedOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mcops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: memcd-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `memcd-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Memcached.

Let's create the `MemcachedOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/memcached/rotate-auth/rotate-auth-generated.yaml
 Memcachedopsrequest.ops.kubedb.com/mcops-rotate-auth-generated created
```
Let's wait for `MemcachedOpsrequest` to be `Successful`. Run the following command to watch `MemcachedOpsrequest` CRO
```shell
 $ kubectl get Memcachedopsrequest -n demo
 NAME                          TYPE         STATUS       AGE
 mcops-rotate-auth-generated   RotateAuth   Successful   7m47s
```
If we describe the `MemcachedOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe MemcachedopsRequest -n demo mcops-rotate-auth-generated
Name:         mcops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MemcachedOpsRequest
Metadata:
  Creation Timestamp:  2025-08-21T10:08:37Z
  Generation:          1
  Resource Version:    147775
  UID:                 2466e827-fee0-4f58-89f5-2cd2265cf1a6
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   memcd-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-21T10:08:37Z
    Message:               Memcached Ops Request has started to rotate auth for Memcached
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-21T10:08:40Z
    Message:               Successfully generated new Credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-21T10:08:40Z
    Message:               Successfully Updated PetSets for rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-21T10:08:51Z
    Message:               Restarted pods after rotate auth
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-21T10:08:46Z
    Message:               evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--memcd-quickstart-0
    Last Transition Time:  2025-08-21T10:08:46Z
    Message:               is pod ready; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsPodReady
    Last Transition Time:  2025-08-21T10:08:51Z
    Message:               is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--memcd-quickstart-0
    Last Transition Time:  2025-08-21T10:08:51Z
    Message:               Successfully Rotated Memcached Auth Secret
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age   From                         Message
  ----     ------                                                          ----  ----                         -------
  Normal   PauseDatabase                                                   112s  KubeDB Ops-manager Operator  Pausing Memcached demo/memcd-quickstart
  Normal   VersionUpdate                                                   109s  KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                   109s  KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0     103s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
  Warning  is pod ready; ConditionStatus:False                             103s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False
  Warning  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0  98s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
  Normal   RestartPods                                                     98s   KubeDB Ops-manager Operator  Restarted pods after rotate auth
  Normal   ResumeDatabase                                                  98s   KubeDB Ops-manager Operator  Resuming Memcached demo/memcd-quickstart
  Normal   ResumeDatabase                                                  98s   KubeDB Ops-manager Operator  Successfully resumed Memcached demo/memcd-quickstart
  Normal   Successful                                                      98s   KubeDB Ops-manager Operator  Successfully Rotated Memcached Auth Secret for demo/memcd-quickstart
```

**Verify Auth is rotated**
```shell
$ kubectl get mc -n demo memcd-quickstart -ojson | jq .spec.authSecret.name
"memcd-quickstart-auth"
$ kubectl get secret -n demo memcd-quickstart-auth -o=jsonpath='{.data.authData}' | base64 -d
user:yjf3Oc;ZlSs.iMVO                    
```
**Let's verify whether the credential is working or not:**

We will connect to `memcd-quickstart-0` pod from local-machine using port-frowarding and it must be running in separate terminal.
```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211

```
Now, you can connect to this database using `telnet`.Connect to Memcached from local-machine through telnet.
```shell

$ telnet 127.0.0.1 11211
#Output
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.

# Save data Command:
set my_key 0 2592000 21
# Meaning:
# 0       => no flags
# 2592000 => TTL (Time-To-Live) in [s]
# 21       => credential size in bytes
user yjf3Oc;ZlSs.iMVO  
#ouput
STORED
#command
version
#output
VERSION 1.6.22
#output
quit
Connection closed by foreign host.
```
Your credentials have been rotated successfully, so everything’s working.
Also, there will be two more new keys in the secret that stores the previous credentials. The key is `authData.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo memcd-quickstart-auth -o go-template='{{ index .data "authData.prev" }}' | base64 -d
user:ikbkjbodeewenrgj
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using User Created Credentials

At first, we need to create a secret with `kubernetes.io/basic-auth` type using custom username 
and password. Below is the command to create a secret with `kubernetes.io/basic-auth` type,
You can use the following yaml regarding changing the credentials

```shell
apiVersion: v1
data:
  authData: dXNlcjpwYXNzCg==
kind: Secret
metadata:
   name: mc-new-auth
   namespace: demo
type: Opaque
```
Here,
The `data.authdata` field stores user credentials in `Base64-encoded` format as **username:password**.
Now create a `MemcachedOpsRequest` with `RotateAuth` type. Below is the YAML of the
`MemcachedOpsRequest` that we are going to create,

Let's create the secret 
```shell
kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/memcached/rotate-auth/secret.yaml
secret/mc-new-auth created
```
```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mcops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: memcached-quickstart
  authentication:
    secretRef:
      name: mc-new-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `memcd-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Memcached.
- `spec.authentication.secretRef.name` specifies that we are using `mc-new-auth` as `spec.authSecret.name` for authentication.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/memcached/rotate-auth/rotate-auth-user.yaml
Memcachedopsrequest.ops.kubedb.com/mcops-rotate-auth-user created
```
Let’s wait for `MemcachedOpsRequest` to be Successful. Run the following command to watch `MemcachedOpsRequest` CRO:

```shell
$ kubectl get Memcachedopsrequest -n demo
NAME                          TYPE         STATUS       AGE
mcops-rotate-auth-user        RotateAuth   Successful   7m44s
```
We can see from the above output that the `MemcachedOpsRequest` has succeeded. If we describe the `MemcachedOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Memcachedopsrequest -n demo mcops-rotate-auth-user
Name:         mcops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MemcachedOpsRequest
Metadata:
  Creation Timestamp:  2025-08-21T11:39:01Z
  Generation:          1
  Resource Version:    159639
  UID:                 6d289952-fd05-4311-985e-9a7fa7845953
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  mc-new-auth
  Database Ref:
    Name:   memcd-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-21T11:39:01Z
    Message:               Memcached Ops Request has started to rotate auth for Memcached
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-21T11:39:01Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-21T11:39:01Z
    Message:               Successfully Updated PetSets for rotate auth type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-21T11:39:11Z
    Message:               Restarted pods after rotate auth
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-21T11:39:06Z
    Message:               evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--memcd-quickstart-0
    Last Transition Time:  2025-08-21T11:39:06Z
    Message:               is pod ready; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsPodReady
    Last Transition Time:  2025-08-21T11:39:11Z
    Message:               is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--memcd-quickstart-0
    Last Transition Time:  2025-08-21T11:39:11Z
    Message:               Successfully Rotated Memcached Auth Secret
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age   From                         Message
  ----     ------                                                          ----  ----                         -------
  Normal   VersionUpdate                                                   13m   KubeDB Ops-manager Operator  Updating PetSets
  Normal   VersionUpdate                                                   13m   KubeDB Ops-manager Operator  Successfully Updated PetSets
  Warning  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0     13m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
  Warning  is pod ready; ConditionStatus:False                             13m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False
  Warning  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0  13m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
  Normal   RestartPods                                                     13m   KubeDB Ops-manager Operator  Restarted pods after rotate auth
  Normal   ResumeDatabase                                                  13m   KubeDB Ops-manager Operator  Resuming Memcached demo/memcd-quickstart
  Normal   ResumeDatabase                                                  13m   KubeDB Ops-manager Operator  Successfully resumed Memcached demo/memcd-quickstart
  Normal   Successful                                                      13m   KubeDB Ops-manager Operator  Successfully Rotated Memcached Auth Secret for demo/memcd-quickstart

```
**Verify auth is rotate**
```shell
$ kubectl get mc -n demo memcd-quickstart -ojson | jq .spec.authSecret.name
"mc-new-auth"
$  kubectl get secret -n demo mc-new-auth -o=jsonpath='{.data.authData}' | base64 -d
user:pass                                                                                   
```
**Let's verify whether the credential is working or not:**

We will connect to `memcd-quickstart-0` pod from local-machine using port-frowarding and it must be running in separate terminal.
```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211

```
Now, you can connect to this database using `telnet`.Connect to Memcached from local-machine through telnet.
```shell

$ telnet 127.0.0.1 11211
#Output
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.

# Save data Command:
set my_key 0 2592000 9
# Meaning:
# 0       => no flags
# 2592000 => TTL (Time-To-Live) in [s]
# 9      => credential size in bytes
user pass
#ouput
STORED
#command
version
#output
VERSION 1.6.22
#output
quit
Connection closed by foreign host.
```
Your credentials have been rotated successfully, so everything’s working.

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo mc-new-auth -o go-template='{{ index .data "authData.prev" }}' | base64 -d
user:yjf3Oc;ZlSs.iMVO
                  
```

The above output shows that the password has been updated successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning Up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Memcachedopsrequest mcops-rotate-auth-generated mcops-rotate-auth-user -n demo
Memcachedopsrequest.ops.kubedb.com "mcops-rotate-auth-generated" "mcops-rotate-auth-user" deleted
$ kubectl delete secret -n demo mc-new-auth
secret "mc-new-auth" deleted
$ kubectl delete secret -n demo  memcd-quickstart-auth
secret "memcd-quickstart-auth" deleted
```

## Next Steps

- Detail concepts of [Memcached object](/docs/guides/memcached/concepts/memcached.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
