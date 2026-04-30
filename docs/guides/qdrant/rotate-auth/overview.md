---
title: Rotating Qdrant Credentials
menu:
  docs_{{ .version }}:
    identifier: qdrant-rotate-auth-overview
    name: Overview
    parent: qdrant-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Qdrant

This guide shows how to rotate Qdrant authentication credentials.

## Before You Begin

- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/rotate-auth/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply RotateAuth OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/rotate-auth/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-rotate-auth
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-rotate-auth
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-rotate-auth
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
