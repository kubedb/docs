---
title: Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: etcd-rotate-auth-details
    name: Guide
    parent: etcd-rotate-authentication
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Etcd

**Rotate Authentication** is a feature of the KubeDB Ops-manager that lets you rotate the etcd
credentials of a running cluster with an `EtcdOpsRequest`. There are two ways to perform the
rotation:

1. **Operator generated:** the KubeDB operator generates a random password, applies it to etcd and
   updates the existing auth Secret with it.
2. **User defined:** you create your own `kubernetes.io/basic-auth` Secret containing the desired
   `password`, and reference it from the `EtcdOpsRequest`.

> **etcd rotates live.** Unlike most KubeDB day-2 operations, this one does **not** restart any pod.
> etcd RBAC users live inside the keyspace, so a password change takes effect on the next
> authentication. See the [overview](/docs/guides/etcd/rotate-authentication/overview.md#why-rotating-etcd-credentials-does-not-restart-anything)
> for the details.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Rotate Authentication Overview](/docs/guides/etcd/rotate-authentication/overview.md)

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure the Provisioner and Ops-manager operators are installed with `featureGates.Etcd=true`.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create an Etcd cluster

KubeDB implements an `Etcd` CRD to define the specification of an etcd cluster.

You can apply this yaml file:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-quickstart
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/rotate-auth/etcd.yaml
etcd.kubedb.com/etcd-quickstart created
```

Now, wait until `etcd-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get etcd -n demo -w
NAME              VERSION   STATUS   AGE
etcd-quickstart   3.6.4     Ready    3m
```

## Verify authentication

The credentials live in the auth Secret referenced by `spec.authSecret`. It is named
`<db-name>-auth` by default and holds a `username` and a `password` key. The username is always
`root` for a KubeDB provisioned etcd cluster.

```bash
$ kubectl get etcd -n demo etcd-quickstart -ojson | jq .spec.authSecret.name
"etcd-quickstart-auth"

$ kubectl get secret -n demo etcd-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
root
$ kubectl get secret -n demo etcd-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
yFj_WnVA9rxfQlLt
```

Now you can exec into a member pod and confirm that those credentials authenticate:

```bash
$ kubectl exec -it -n demo etcd-quickstart-0 -c etcd -- etcdctl \
    --endpoints=http://127.0.0.1:2379 --user=root:yFj_WnVA9rxfQlLt put foo bar
OK
$ kubectl exec -it -n demo etcd-quickstart-0 -c etcd -- etcdctl \
    --endpoints=http://127.0.0.1:2379 --user=root:yFj_WnVA9rxfQlLt get foo
foo
bar
```

If the read and the write succeed, the credentials in the Secret are working.

Note the pod ages before rotating, so that you can convince yourself afterwards that nothing was
restarted:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-quickstart
NAME                READY   STATUS    RESTARTS   AGE
etcd-quickstart-0   1/1     Running   0          8m
etcd-quickstart-1   1/1     Running   0          7m
etcd-quickstart-2   1/1     Running   0          7m
```

## Create a RotateAuth EtcdOpsRequest

### 1. Using operator generated credentials

To rotate the credentials with an operator generated password, create an `EtcdOpsRequest` of type
`RotateAuth` and leave `spec.authentication` unset. Below is the YAML of the `EtcdOpsRequest` we are
going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcdops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: etcd-quickstart
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are rotating the authentication of the `etcd-quickstart`
  cluster.
- `spec.type` specifies that we are performing a `RotateAuth` operation.
- `spec.authentication` is omitted, which is what tells the operator to generate the new password
  itself.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/rotate-auth/rotate-auth-generated.yaml
etcdopsrequest.ops.kubedb.com/etcdops-rotate-auth-generated created
```

Let's wait for the `EtcdOpsRequest` to be `Successful`. Run the following command to watch it,

```bash
$ kubectl get etcdopsrequest -n demo
NAME                            TYPE         STATUS       AGE
etcdops-rotate-auth-generated   RotateAuth   Successful   1m
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcdops-rotate-auth-generated
Name:         etcdops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-02-10T11:24:10Z
  Generation:          1
  Resource Version:    546623
  UID:                 97a07133-c98e-457c-9249-85c0c690a82e
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   etcd-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2026-02-10T11:24:10Z
    Message:               Etcd ops request has started to rotate the authentication
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-02-10T11:24:13Z
    Message:               Successfully staged the new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2026-02-10T11:24:18Z
    Message:               Successfully applied the new credentials to the etcd RBAC store
    Observed Generation:   1
    Reason:                EtcdCredentialApplied
    Status:                True
    Type:                  EtcdCredentialApplied
    Last Transition Time:  2026-02-10T11:24:20Z
    Message:               Successfully updated the Etcd auth secret reference
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-02-10T11:24:21Z
    Message:               Successfully rotated the etcd auth secret
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age  From                         Message
  ----    ------                 ---- ----                         -------
  Normal  PauseDatabase          1m   KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-quickstart
  Normal  UpdateCredential       1m   KubeDB Ops-manager Operator  Successfully staged the new credentials
  Normal  EtcdCredentialApplied  63s  KubeDB Ops-manager Operator  Successfully applied the new credentials to the etcd RBAC store
  Normal  UpdateDatabase         61s  KubeDB Ops-manager Operator  Successfully updated the Etcd auth secret reference
  Normal  ResumeDatabase         61s  KubeDB Ops-manager Operator  Resuming Etcd demo/etcd-quickstart
  Normal  Successful             60s  KubeDB Ops-manager Operator  Successfully rotated the etcd auth secret
```

Notice what is **absent** from that condition list: there is no `RestartEtcdPods`, no
`EvictPod--<pod>` and no `CheckPodReady--<pod>`. The whole rotation is three writes —
`UpdateCredential` (stage), `EtcdCredentialApplied` (apply to etcd's RBAC store), `UpdateDatabase`
(promote and repoint the `Etcd` object).

**Verify that no pod was restarted**

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-quickstart
NAME                READY   STATUS    RESTARTS   AGE
etcd-quickstart-0   1/1     Running   0          11m
etcd-quickstart-1   1/1     Running   0          10m
etcd-quickstart-2   1/1     Running   0          10m
```

The `AGE` column keeps counting up from the original creation and `RESTARTS` is still `0` — the
members were never evicted.

**Verify that the auth is rotated**

```bash
$ kubectl get etcd -n demo etcd-quickstart -ojson | jq .spec.authSecret.name
"etcd-quickstart-auth"

$ kubectl get secret -n demo etcd-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
root
$ kubectl get secret -n demo etcd-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
zGB9GF!NXwI.2HP9
```

The superseded credential is kept in the same Secret under the `username.prev` and `password.prev`
keys, for rollback purposes:

```bash
$ kubectl get secret -n demo etcd-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root
$ kubectl get secret -n demo etcd-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
yFj_WnVA9rxfQlLt
```

The staging keys (`username.next` / `password.next`) are only present while the rotation is in
flight; once the new password has been promoted they are removed again.

The operator also stamps the moment the credential went live, both as an annotation on the Secret
and as `spec.authSecret.activeFrom` on the `Etcd` object:

```bash
$ kubectl get etcd -n demo etcd-quickstart -o jsonpath='{.spec.authSecret.activeFrom}'
2026-02-10T11:24:13Z
```

Finally, confirm the new password really is the live one inside etcd:

```bash
$ kubectl exec -it -n demo etcd-quickstart-0 -c etcd -- etcdctl \
    --endpoints=http://127.0.0.1:2379 --user='root:zGB9GF!NXwI.2HP9' get foo
foo
bar
```

### 2. Using user created credentials

First, create a Secret of type `kubernetes.io/basic-auth` holding the username and the password you
want the cluster to use. Both keys must be present and non-empty, otherwise the ops request fails
validation.

> **Note:** keep the username as `root`, which is the user KubeDB provisions and the one the
> operator's health checker authenticates with. If the Secret names a *different* user, the operator
> creates that etcd user and grants it the `root` role rather than renaming the existing one, which
> leaves the old user behind in etcd's RBAC store.

```bash
$ kubectl create secret generic etcd-quickstart-user-auth -n demo \
                                                --type=kubernetes.io/basic-auth \
                                                --from-literal=username=root \
                                                --from-literal=password=etcd-secret
secret/etcd-quickstart-user-auth created
```

Now create an `EtcdOpsRequest` of type `RotateAuth` that references it. Below is the YAML of the
`EtcdOpsRequest` we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcdops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: etcd-quickstart
  authentication:
    secretRef:
      kind: Secret
      name: etcd-quickstart-user-auth
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are rotating the authentication of the `etcd-quickstart`
  cluster.
- `spec.type` specifies that we are performing a `RotateAuth` operation.
- `spec.authentication.secretRef.name` names the Secret whose `password` should become the live
  credential. After the rotation, this Secret becomes the cluster's `spec.authSecret`, marked
  `externallyManaged: true`.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/rotate-auth/rotate-auth-user.yaml
etcdopsrequest.ops.kubedb.com/etcdops-rotate-auth-user created
```

Let's wait for the `EtcdOpsRequest` to be `Successful`. Run the following command to watch it,

```bash
$ kubectl get etcdopsrequest -n demo
NAME                            TYPE         STATUS       AGE
etcdops-rotate-auth-generated   RotateAuth   Successful   19h
etcdops-rotate-auth-user        RotateAuth   Successful   1m
```

We can see from the above output that the `EtcdOpsRequest` has succeeded. If we describe it, we get
an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcdops-rotate-auth-user
Name:         etcdops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-02-11T06:45:44Z
  Generation:          1
  Resource Version:    562328
  UID:                 d25c3d36-cc15-4c82-8fe4-64e5ffc1467c
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Kind:  Secret
      Name:  etcd-quickstart-user-auth
  Database Ref:
    Name:   etcd-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2026-02-11T06:45:44Z
    Message:               Etcd ops request has started to rotate the authentication
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-02-11T06:45:47Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2026-02-11T06:45:52Z
    Message:               Successfully applied the new credentials to the etcd RBAC store
    Observed Generation:   1
    Reason:                EtcdCredentialApplied
    Status:                True
    Type:                  EtcdCredentialApplied
    Last Transition Time:  2026-02-11T06:45:54Z
    Message:               Successfully updated the Etcd auth secret reference
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-02-11T06:45:55Z
    Message:               Successfully rotated the etcd auth secret
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age  From                         Message
  ----    ------                 ---- ----                         -------
  Normal  PauseDatabase          1m   KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-quickstart
  Normal  UpdateCredential       1m   KubeDB Ops-manager Operator  Successfully referenced the user provided authSecret
  Normal  EtcdCredentialApplied  63s  KubeDB Ops-manager Operator  Successfully applied the new credentials to the etcd RBAC store
  Normal  UpdateDatabase         61s  KubeDB Ops-manager Operator  Successfully updated the Etcd auth secret reference
  Normal  ResumeDatabase         61s  KubeDB Ops-manager Operator  Resuming Etcd demo/etcd-quickstart
  Normal  Successful             60s  KubeDB Ops-manager Operator  Successfully rotated the etcd auth secret
```

Again, no pod was restarted — the condition list is the same three steps as before, with the
`UpdateCredential` message reflecting that an existing, user provided Secret was referenced instead
of a new password being generated.

**Verify that the auth is rotated**

The `Etcd` object now points at your Secret:

```bash
$ kubectl get etcd -n demo etcd-quickstart -ojson | jq .spec.authSecret
{
  "activeFrom": "2026-02-11T06:45:47Z",
  "externallyManaged": true,
  "kind": "Secret",
  "name": "etcd-quickstart-user-auth"
}

$ kubectl get secret -n demo etcd-quickstart-user-auth -o jsonpath='{.data.username}' | base64 -d
root
$ kubectl get secret -n demo etcd-quickstart-user-auth -o jsonpath='{.data.password}' | base64 -d
etcd-secret
```

The credential that was superseded is recorded inside your Secret, again under `username.prev` and
`password.prev`:

```bash
$ kubectl get secret -n demo etcd-quickstart-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
root
$ kubectl get secret -n demo etcd-quickstart-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
zGB9GF!NXwI.2HP9
```

And the new password authenticates against the running cluster:

```bash
$ kubectl exec -it -n demo etcd-quickstart-0 -c etcd -- etcdctl \
    --endpoints=http://127.0.0.1:2379 --user=root:etcd-secret get foo
foo
bar
```

> The previously used `etcd-quickstart-auth` Secret is left in place; KubeDB simply stops referring
> to it. Delete it yourself once you are sure you no longer need the old credential for a rollback.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete etcdopsrequest -n demo etcdops-rotate-auth-generated etcdops-rotate-auth-user
etcdopsrequest.ops.kubedb.com "etcdops-rotate-auth-generated" deleted
etcdopsrequest.ops.kubedb.com "etcdops-rotate-auth-user" deleted
$ kubectl delete etcd -n demo etcd-quickstart
etcd.kubedb.com "etcd-quickstart" deleted
$ kubectl delete secret -n demo etcd-quickstart-user-auth
secret "etcd-quickstart-user-auth" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Detail concepts of [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Detail concepts of [EtcdOpsRequest object](/docs/guides/etcd/concepts/etcdopsrequest.md).
- Restart the members of a cluster with [Restart](/docs/guides/etcd/restart/restart.md).
- Change the etcd tuning knobs of a running cluster with [Reconfigure](/docs/guides/etcd/reconfigure/reconfigure.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
