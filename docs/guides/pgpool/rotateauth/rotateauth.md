---
title: Rotate Authentication Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-rotateauth-details
    name: Rotate Authentication Pgpool
    parent: pp-rotateauth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---


# Rotate Authentication of Pgpool

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Pgpool` user's authentication credentials using a `PgpoolOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential, and does not provide the secret details directly to the user.
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

In order to rotate authentication to the Pgpool using operator generated, we have to create a `PgpoolOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `PgpoolOpsRequest` CRO that we are going to create,
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
Let's wait for `PgpoolOpsrequest` to be `Successful`. Run the following command to watch `PgpoolOpsrequest` CRO
```shell
 $ kubectl get PgpoolOpsRequest -n pool
NAME                           TYPE         STATUS       AGE
pgpops-rotate-auth-generated   RotateAuth   Successful   52s
```
If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.
```shell
$  kubectl describe pgpoolopsrequest -n pool pgpops-rotate-auth-generated
Name:         pgpops-rotate-auth-generated
Namespace:    pool
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T09:39:25Z
  Generation:          1
  Resource Version:    127114
  UID:                 18c24f33-d69b-4ed7-a581-82284df130ae
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   quick-pgpool
  Timeout:  5m
  Type:     RotateAuth
Status:
  Phase:  Pending
Events:   <none>
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
> Note: The `username` must be fixed as `sa`. The `password` must include uppercase letters, lowercase letters, and numbers
```shell
$  kubectl create secret generic quick-pb-user-auth -n pool \
        --type=kubernetes.io/basic-auth \
        --from-literal=username=user \
        --from-literal=password=Pgpool2
secret/quick-pb-user-auth created
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
      name: quick-pb-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `quick-pgpool`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Pgpool.
- `spec.authentication.secretRef.name` specifies that we are using `quick-pb-user-auth` as `spec.authSecret.name` for authentication.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```shell
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/rotateauth/rotateauthuser.yaml
Pgpoolopsrequest.ops.kubedb.com/pbops-rotate-auth-user created
```
Let’s wait for `PgpoolOpsRequest` to be Successful. Run the following command to watch `PgpoolOpsRequest` CRO:

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
  Creation Timestamp:  2025-07-18T09:53:51Z
  Generation:          1
  Resource Version:    128231
  UID:                 aac4055e-889e-49cd-9124-ea9f5cdd88a1
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  quick-pb-user-auth
  Database Ref:
    Name:   quick-pgpool
  Timeout:  5m
  Type:     RotateAuth
Status:
  Phase:  Pending
Events:   <none>

```
**Verify auth is rotate**
```shell
$ kubectl get pgpool -n pool quick-pgpool -ojson | jq .spec.authSecret.name
"quick-pb-user-auth"
$ kubectl get secrets -n pool quick-pb-user-auth -o jsonpath='{.data.\username}' | base64 -d
user⏎                                                                                       
$ kubectl get secrets -n pool quick-pb-user-auth -o jsonpath='{.data.\password}' | base64 -d
Pgpool2⏎                                                                                                                                        
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$  kubectl get secret -n pool quick-pgpool-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
pcp⏎                                                                                                                                                            
$ kubectl get secret -n pool quick-pgpool-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
gZoAOjr0iUkH07ku⏎                                
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Pgpoolopsrequest pgpops-rotate-auth-generated ppops-rotate-auth-user -n pool
Pgpoolopsrequest.ops.kubedb.com "pgpops-rotate-auth-generated" "ppops-rotate-auth-user" deleted
$ kubectl delete secret -n pool quick-pb-user-auth
secret "quick-pb-user-auth" deleted
$ kubectl delete secret -n pool   quick-pgpool-auth
secret "quick-pgpool-auth " deleted
```

## Next Steps

- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
