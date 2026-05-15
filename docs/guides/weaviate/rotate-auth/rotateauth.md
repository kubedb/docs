---
title: Rotate Auth of Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-rotate-auth-cluster
    name: Cluster
    parent: weaviate-rotate-auth
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Weaviate

`RotateAuth` updates authentication credentials used by Weaviate. This guide shows how to trigger rotation using `WeaviateOpsRequest`.

## Before You Begin

- Ensure KubeDB and Ops Manager are installed.
- Review [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) and [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md).
- Use namespace `demo` for this guide.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy Weaviate

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
weaviate.kubedb.com/weaviate-sample created

$ kubectl get weaviate -n demo weaviate-sample -w
NAME              VERSION   STATUS   AGE
weaviate-sample   1.33.1    Ready    2m
```

## Rotate Authentication Using Operator Generated Credentials

Apply the sample ops request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: weaviate-sample
  timeout: 5m
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/rotate-auth/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-rotate-auth created
```

## Verify Credential Rotation

Check request completion:

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-rotate-auth
NAME                   TYPE         STATUS       AGE
weaviate-rotate-auth   RotateAuth   Successful   2m
```

Inspect operation details:

```bash
$ kubectl describe weaviateopsrequest -n demo weaviate-rotate-auth
```

Inspect auth secret after rotation:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.authSecret.name}{"\n"}'
$ kubectl get secret -n demo weaviate-sample-auth -o yaml
```

If your environment stores previous credentials, you can inspect previous values from secret keys such as `username.prev` and `password.prev`.

## Optional: Use User-Provided Credentials

You can rotate to user-provided credentials by creating a `kubernetes.io/basic-auth` secret and referencing it from `spec.authSecret` in the ops request.

```bash
kubectl create secret generic weaviate-user-auth -n demo \
  --type=kubernetes.io/basic-auth \
  --from-literal=username=admin \
  --from-literal=password='strong-pass'
```

Then apply a `WeaviateOpsRequest` with `type: RotateAuth` and `spec.authSecret.name: weaviate-user-auth`.

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-rotate-auth
kubectl delete weaviate -n demo weaviate-sample
kubectl delete secret -n demo weaviate-user-auth --ignore-not-found
kubectl delete ns demo
```