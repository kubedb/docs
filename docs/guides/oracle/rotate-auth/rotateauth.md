---
title: Rotate Auth of Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-rotate-auth-cluster
    name: Cluster
    parent: oracle-rotate-auth
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication for Oracle

KubeDB supports rotating authentication credentials (API key) for Oracle via a `OracleOpsRequest`. This tutorial will show you how to use KubeDB to rotate authentication credentials.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/oracle/rotate-auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/rotate-auth) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Oracle

In this section, we are going to deploy a Oracle database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/rotate-auth/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

Now, wait until `oracle-sample` has status `Ready`:

```bash
$ kubectl get oracle -n demo
NAME             VERSION   STATUS   AGE
oracle-sample    1.17.0    Ready    3m22s
```

When Oracle is deployed, KubeDB creates a secret called `oracle-sample-auth` (format: `{db-name}-auth`) that stores the API key used for authentication.

```bash
$ kubectl get secret -n demo oracle-sample-auth -o jsonpath='{.data.api-key}' | base64 --decode
<initial-api-key>
```

## Apply RotateAuth OpsRequest

Now, we are going to create a `OracleOpsRequest` to rotate the authentication credentials.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: oracle-sample
```

Here,

- `spec.databaseRef.name` specifies that we are rotating auth credentials for `oracle-sample` Oracle database.
- `spec.type` specifies that we are performing `RotateAuth` on our database.

Let's create the `OracleOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/rotate-auth/ops.yaml
oracleopsrequest.ops.kubedb.com/qdops-rotate-auth created
```

## Verify Authentication Rotated

If everything goes well, `KubeDB` ops-manager operator will rotate the authentication credentials of the `Oracle` database.

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ watch -n 3 kubectl get OracleOpsRequest -n demo qdops-rotate-auth
Every 3.0s: kubectl get OracleOpsRequest -n demo qdops-rotate-auth

NAME                TYPE         STATUS       AGE
qdops-rotate-auth   RotateAuth   Successful   2m15s
```

We can see from the above output that the `OracleOpsRequest` has succeeded. Now let's check if the authentication secret has been updated:

```bash
$ kubectl get secret -n demo oracle-sample-auth -o jsonpath='{.data.api-key}' | base64 --decode
<new-rotated-api-key>
```

You can see that the API key has been rotated. The new key is different from the initial key. KubeDB has automatically updated the Oracle instances to use the new credentials.

## Rotate Auth with a Custom Secret

You can also rotate the authentication credentials using a custom secret that you provide:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-rotate-auth-custom
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: oracle-sample
  authentication:
    secretRef:
      name: my-custom-auth-secret
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/rotate-auth/ops-custom.yaml
oracleopsrequest.ops.kubedb.com/qdops-rotate-auth-custom created
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo qdops-rotate-auth qdops-rotate-auth-custom
kubectl delete oracle -n demo oracle-sample
kubectl delete ns demo
```