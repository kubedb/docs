---
title: Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-rotate-auth-details
    name: Guide
    parent: guides-oracle-rotate-authentication
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Oracle

KubeDB supports rotating the authentication credentials (the database password) of an Oracle database through an `OracleOpsRequest` of type `RotateAuth`. This guide shows both ways of rotating: using operator generated credentials and using your own credentials.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/rotate-auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/rotate-auth) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`.

## Create an Oracle database

In this section, we are going to deploy an Oracle standalone database. Below is the YAML of the `Oracle` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sa-sample
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above and wait until it is `Ready`.

## Verify authentication

KubeDB stores the database credentials in a Secret named `<db-name>-auth`. For our database that is `oracle-sa-sample-auth`,

```bash
$ kubectl get secret -n demo oracle-sa-sample-auth -o jsonpath='{.data.username}' | base64 -d
sys
$ kubectl get secret -n demo oracle-sa-sample-auth -o jsonpath='{.data.password}' | base64 -d
LbK!aQQ3zkcOC3~u
```

> **Note:** The privileged user for Oracle is `SYS`. Oracle does **not** allow renaming the `SYS` user, so rotate authentication changes the **password** only.

## Create RotateAuth OracleOpsRequest

There are two ways to rotate the authentication.

#### 1. Using operator generated credentials

To rotate the password using an operator generated credential, create an `OracleOpsRequest` of type `RotateAuth` without referencing any secret. Below is the YAML,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: standalone-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: oracle-sa-sample
  apply: IfReady
```

Here,

- `spec.type` specifies that we are performing a `RotateAuth` operation.
- `spec.databaseRef.name` specifies the database `oracle-sa-sample`.
- `spec.apply: IfReady` tells the operator to apply the operation only when the database is ready.

Let's create the `OracleOpsRequest`,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/rotate-auth/standalone-rotate-auth.yaml
oracleopsrequest.ops.kubedb.com/standalone-rotate-auth created
```

Let's wait until the `OracleOpsRequest` becomes `Successful`,

```bash
$ kubectl get oracleopsrequest -n demo standalone-rotate-auth
NAME                     TYPE         STATUS       AGE
standalone-rotate-auth   RotateAuth   Successful   2m7s
```

The full progress is shown by `kubectl describe`,

```bash
$ kubectl describe oracleopsrequest -n demo standalone-rotate-auth
Name:         standalone-rotate-auth
Namespace:    demo
...
Status:
  Conditions:
    Last Transition Time:  2026-06-22T19:01:09Z
    Message:               Oracle ops-request has started to rotate auth for oracle nodes
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2026-06-22T19:03:17Z
    Message:               Successfully generated new credentials
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2026-06-22T19:01:26Z
    Message:               successfully reconciled the oracle with new auth credentials
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-22T19:02:17Z
    Message:               Successfully restarted all nodes
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-22T19:02:12Z
    Message:               Pod oracle-sa-sample-0 restarted and healthy
    Status:                True
    Type:                  RestartedPod--oracle-sa-sample-0
    Last Transition Time:  2026-06-22T19:02:24Z
    Message:               Successfully completed rotate-auth for Oracle
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Phase:                   Successful
```

**Verify auth is rotated:**

After the operation succeeds, the operator has generated a new password and stored it in the auth secret. The previous credential is kept under the `authData.prev` keys so applications have a grace window to migrate,

```bash
$ kubectl get secret -n demo oracle-sa-sample-auth -o jsonpath='{.data.password}' | base64 -d
VYWX2Wu!Sx1JdqKl
```

The auth secret now also holds the previous credentials under the `.prev` keys,

```bash
$ kubectl get secret -n demo oracle-sa-sample-auth -o json | jq '.data | keys'
[
  "password",
  "password.prev",
  "username",
  "username.prev"
]

$ kubectl get secret -n demo oracle-sa-sample-auth -o jsonpath='{.data.password\.prev}' | base64 -d
LbK!aQQ3zkcOC3~u
```

Finally, let's confirm the new password works by connecting to the database,

```bash
$ kubectl exec -n demo oracle-sa-sample-0 -c oracle -- bash -lc \
    "echo -e 'SELECT USER FROM DUAL;\nexit;' | sqlplus -s sys/<new-password>@localhost:1521/ORCL as sysdba"

USER
------------------------------
SYS
```

#### 2. Using user created credentials

If you want to set a specific password, first create a Secret of type `kubernetes.io/basic-auth` containing the `username` (`sys`) and your desired `password`:

```bash
$ kubectl create secret generic oracle-user-auth -n demo \
    --type=kubernetes.io/basic-auth \
    --from-literal=username=sys \
    --from-literal=password='New-Strong-Pass-123'
secret/oracle-user-auth created
```

> **Note:** The `username` must remain `sys`; only the password can change.

Then create an `OracleOpsRequest` that references the secret through `spec.authentication.secretRef.name`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: standalone-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: oracle-sa-sample
  authentication:
    secretRef:
      kind: Secret
      name: oracle-user-auth
  apply: IfReady
```

```bash
$ kubectl create -f standalone-rotate-auth-user.yaml
oracleopsrequest.ops.kubedb.com/standalone-rotate-auth-user created
```

Once the ops request succeeds, the database password is updated to the value from your `oracle-user-auth` secret, and the `Oracle` object's `spec.authSecret` is pointed at it.

## Rotating authentication for a DataGuard cluster

The same `OracleOpsRequest` works for a DataGuard cluster — point `spec.databaseRef.name` at the DataGuard database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: dataguard-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: oracle-dg-sample
  apply: IfReady
```

The operator updates the credential on the primary, reconciles the PetSets, and performs a rolling restart so every DataGuard pod (primary, standbys, observer) picks up the new credential.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo standalone-rotate-auth
kubectl patch -n demo oracle/oracle-sa-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo oracle-sa-sample
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Monitor your Oracle database with KubeDB using [Prometheus operator](/docs/guides/oracle/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
