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
    - [Ignite](/docs/guides/ignite/concepts/ignite/index.md)
    - [IgniteOpsRequest](/docs/guides/ignite/concepts/ignite/index.md)

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
  name: sample-ignite
  namespace: demo
spec:
  version: "2.16.0"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/ignite/quickstart/examples/sample-ignite.yaml
ignite.kubedb.com/sample-ignite created
```

Now, wait until sample-ignite has status Ready. i.e,

```shell
$ kubectl get ignite -n demo -w
NAME            VERSION   STATUS   AGE
sample-ignite   2.16.0    Ready    30m
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get ignite -n demo sample-ignite -ojson | jq .spec.authSecret.name
"sample-ignite-auth"
$ kubectl get secret -n demo sample-ignite-auth -o=jsonpath='{.data.username}' | base64 -d
ignite⏎
$ kubectl get secret -n demo sample-ignite-auth -o=jsonpath='{.data.password}' | base64 -d
s)cJQ*iL8wHySpvT⏎
````
Now, you can exec into the pod `sample-ignite` and connect to the database using `username` and `password`
```shell
$ kubectl exec -it -n demo sample-ignite-0 -- bash
```
If you can access the data and run queries, it means the secrets are working correctly.

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
    name: sample-ignite
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on
`sample-ignite` database.
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
igops-rotate-auth-generated   RotateAuth   Successful   6m28s
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
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   sample-ignite
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-11T06:43:21Z
    Message:               Controller has started to Progress the IgniteOpsRequest: demo/igops-rotate-auth-generated
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-07-11T06:43:24Z
    Message:               Successfully generated new credentials
    Reason:                patchedSecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-11T06:44:12Z
    Message:               Successfully rotate Ignite auth for IgniteOpsRequest: demo/igops-rotate-auth-generated
    Reason:                UpdateCredential
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-11T06:44:12Z
    Message:               Controller has successfully rotate Ignite auth secret demo/igops-rotate-auth-generated
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Phase:                   Successful
```

**Verify Auth is rotated**
```shell
$ kubectl get ignite -n demo sample-ignite -ojson | jq .spec.authSecret.name
"sample-ignite-auth"
$ kubectl get secret -n demo sample-ignite-auth -o=jsonpath='{.data.username}' | base64 -d
ignite⏎
$ kubectl get secret -n demo sample-ignite-auth -o=jsonpath='{.data.password}' | base64 -d
gTJJMdgpKy9U(Eqi⏎
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo sample-ignite-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
ignite⏎
$ kubectl get secret -n demo sample-ignite-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
s)cJQ*iL8wHySpvT⏎
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.

#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

> Note: You cannot change the database `username`, but you can update the `password` while keeping the existing `username`.

```shell
$ kubectl create secret generic sample-ignite-auth-user -n demo \
   --type=kubernetes.io/basic-auth \
   --from-literal=username=ignite \
   --from-literal=password=testpassword
secret/sample-ignite-auth-user created
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
    name: sample-ignite
  authentication:
    secretRef:
      kind: Secret
      name: sample-ignite-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `sample-ignite` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on `Ignite`.
- `spec.authentication.secretRef.name` specifies that we are using `sample-ignite-auth-user` as `spec.authSecret.name` for authentication.

Let's create the `IgniteOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/ignite/rotate-auth/overview/examples/rotate-auth-user.yaml
igniteopsrequest.ops.kubedb.com/igops-rotate-auth-user created
```
Let's wait for `IgniteOpsRequest` to be Successful. Run the following command to watch `IgniteOpsRequest` CRO:

```shell
$ kubectl get igniteopsrequest -n demo
NAME                          TYPE         STATUS       AGE
igops-rotate-auth-generated   RotateAuth   Successful   100s
igops-rotate-auth-user        RotateAuth   Successful   62s
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
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  sample-ignite-auth-user
  Database Ref:
    Name:   sample-ignite
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-14T06:56:25Z
    Message:               Controller has started to Progress the IgniteOpsRequest: demo/igops-rotate-auth-user
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-07-14T06:56:28Z
    Message:               Successfully referenced the user provided authSecret
    Reason:                patchedSecret
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-14T06:56:45Z
    Message:               Successfully rotate Ignite auth for IgniteOpsRequest: demo/igops-rotate-auth-user
    Reason:                UpdateCredential
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-14T06:56:45Z
    Message:               Controller has successfully rotate Ignite auth secret demo/igops-rotate-auth-user
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Phase:                   Successful
```

**Verify auth is rotate**
```shell
$ kubectl get ignite -n demo sample-ignite -ojson | jq .spec.authSecret.name
"sample-ignite-auth-user"
$ kubectl get secret -n demo sample-ignite-auth-user -o=jsonpath='{.data.username}' | base64 -d
ignite⏎
$ kubectl get secret -n demo sample-ignite-auth-user -o=jsonpath='{.data.password}' | base64 -d
testpassword⏎
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo sample-ignite-auth-user -o go-template='{{ index .data "username.prev" }}' | base64 -d
ignite⏎
$ kubectl get secret -n demo sample-ignite-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
gTJJMdgpKy9U(Eqi⏎
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete igniteopsrequest igops-rotate-auth-generated igops-rotate-auth-user -n demo
igniteopsrequest.ops.kubedb.com "igops-rotate-auth-generated" deleted
igniteopsrequest.ops.kubedb.com "igops-rotate-auth-user" deleted
$ kubectl delete secret -n demo sample-ignite-auth-user
secret "sample-ignite-auth-user" deleted
$ kubectl delete secret -n demo sample-ignite-auth
secret "sample-ignite-auth" deleted
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
