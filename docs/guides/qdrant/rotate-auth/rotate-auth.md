---
title: Rotate Auth of Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-rotate-auth-description
    name: Rotate Auth
    parent: qdrant-rotate-auth
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication for Qdrant

KubeDB supports rotating authentication credentials (API key) for Qdrant via a `QdrantOpsRequest`. This tutorial will show you how to use KubeDB to rotate authentication credentials.

KubeDB supports two methods for rotating credentials:
- **Operator Generated** — KubeDB generates a new API key automatically.
- **User Defined** — You provide a custom secret with a new API key.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/rotate-auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/rotate-auth) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Qdrant

In this section, we are going to deploy a Qdrant database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
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

Let's create the `Qdrant` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/rotate-auth/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

Now, wait until `qdrant-sample` has status `Ready`:

```bash
$ kubectl get qdrant -n demo
NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    3m22s
```

When Qdrant is deployed, KubeDB creates a secret called `qdrant-sample-auth` (format: `{db-name}-auth`) that stores the API key used for authentication.

```bash
$ kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 --decode
<initial-api-key>
```

## Apply RotateAuth OpsRequest

Now, we are going to create a `QdrantOpsRequest` to rotate the authentication credentials.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: qdrant-sample
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are rotating auth credentials for `qdrant-sample` Qdrant database.
- `spec.type` specifies that we are performing `RotateAuth` on our database.
- `spec.timeout` specifies the timeout for the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#spectimeout)).
- `spec.apply` specifies when to apply the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#specapply)).

Let's create the `QdrantOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/rotate-auth/ops-request.yaml
qdrantopsrequest.ops.kubedb.com/qdops-rotate-auth created
```

## Verify Authentication Rotated

If everything goes well, `KubeDB` ops-manager operator will rotate the authentication credentials of the `Qdrant` database.

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ watch -n 3 kubectl get QdrantOpsRequest -n demo qdops-rotate-auth
Every 3.0s: kubectl get QdrantOpsRequest -n demo qdops-rotate-auth

NAME                TYPE         STATUS       AGE
qdops-rotate-auth   RotateAuth   Successful   2m15s
```

We can see from the above output that the `QdrantOpsRequest` has succeeded. Now let's check if the authentication secret has been updated:

```bash
$ kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 --decode
<new-rotated-api-key>
```

You can see that the API key has been rotated. The new key is different from the initial key. KubeDB has automatically updated the Qdrant instances to use the new credentials.

## Rotate Auth with a Custom Secret

You can also rotate the authentication credentials using a custom secret that you provide:

```yaml
apiVersion: v1
stringData:
  api-key: MyCus0mAPIKey
  read-only-api-key: MyCus0mReadOnlyKey
kind: Secret
metadata:
  name: my-custom-auth-secret
  namespace: demo
type: Opaque
```

> **Note:** The custom auth secret must contain both `api-key` and `read-only-api-key`. If `read-only-api-key` is missing, the `RotateAuth` OpsRequest will not complete.

Let's create the `Secret` we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/rotate-auth/custom-auth-secret.yaml
secret/my-custom-auth-secret created
```

Now, create a `QdrantOpsRequest` with the custom secret reference:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-rotate-auth-custom
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: qdrant-sample
  authentication:
    secretRef:
      name: my-custom-auth-secret
      kind: Secret
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/rotate-auth/ops-custom.yaml
qdrantopsrequest.ops.kubedb.com/qdops-rotate-auth-custom created
```

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete qdrantopsrequest -n demo qdops-rotate-auth qdops-rotate-auth-custom
qdrantopsrequest.ops.kubedb.com "qdops-rotate-auth" deleted
qdrantopsrequest.ops.kubedb.com "qdops-rotate-auth-custom" deleted

$ kubectl delete secret -n demo my-custom-auth-secret
secret "my-custom-auth-secret" deleted

$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```