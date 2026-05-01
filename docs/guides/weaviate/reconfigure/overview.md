---
title: Reconfiguring Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-reconfigure-overview
    name: Overview
    parent: weaviate-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Weaviate

Use `WeaviateOpsRequest` with type `Reconfigure` to change runtime configuration for a running Weaviate database.

## Before You Begin

- Install KubeDB and Ops Manager from [here](/docs/setup/README.md).
- Review [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) and [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md) concepts.
- Use the example files from `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/reconfigure/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Weaviate

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

When the database is `Ready`, continue with the detailed reconfiguration workflow in [Reconfigure Weaviate](/docs/guides/weaviate/reconfigure/reconfigure.md).

## Verify

```bash
kubectl get weaviateopsrequest -n demo weaviate-reconfigure
kubectl describe weaviateopsrequest -n demo weaviate-reconfigure
kubectl get weaviate -n demo weaviate-sample -o yaml
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-reconfigure
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
