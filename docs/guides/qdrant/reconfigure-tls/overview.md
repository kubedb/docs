---
title: Reconfiguring Qdrant TLS
menu:
  docs_{{ .version }}:
    identifier: qdrant-reconfigure-tls-overview
    name: Overview
    parent: qdrant-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure TLS for Qdrant

This guide shows how to rotate or reconfigure TLS materials for Qdrant.

## Before You Begin

- Install `cert-manager` in your cluster.
- Ensure KubeDB and Ops-manager are installed.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/reconfigure-tls/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply ReconfigureTLS OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure-tls/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-reconfigure-tls
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-reconfigure-tls
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-reconfigure-tls
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
