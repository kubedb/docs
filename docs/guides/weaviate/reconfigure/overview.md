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

This guide tracks the reconfiguration workflow documentation for Weaviate.

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD, so the example manifests referenced by older placeholder docs are not CRD-backed.

## Before You Begin

- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Use the example files from `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/reconfigure/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Weaviate

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

See the detailed note in [Reconfigure Weaviate](/docs/guides/weaviate/reconfigure/reconfigure.md).

## Verify

```bash
kubectl describe weaviateopsrequest -n demo weaviate-reconfigure
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-reconfigure
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
