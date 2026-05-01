---
title: Rotating Weaviate Credentials
menu:
  docs_{{ .version }}:
    identifier: weaviate-rotate-auth-overview
    name: Overview
    parent: weaviate-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Weaviate

Rotate authentication credentials for Weaviate using a `WeaviateOpsRequest` with `type: RotateAuth`.

## Before You Begin

- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Review [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md) concepts.
- Use the example files from `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/rotate-auth/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Weaviate

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

Continue with the complete procedure in [Rotate Auth for Weaviate](/docs/guides/weaviate/rotate-auth/rotateauth.md).

## Verify

```bash
kubectl get secret -n demo weaviate-sample-auth -o yaml
kubectl describe weaviateopsrequest -n demo weaviate-rotate-auth
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-rotate-auth
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
