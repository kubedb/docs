---
title: Reconfiguring Neo4j TLS
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-tls-overview
    name: Overview
    parent: neo4j-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Neo4j TLS

This guide shows how to rotate or update TLS materials of a Neo4j database.

## Before You Begin

- Install `cert-manager` in your cluster.
- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/reconfigure-tls/ops-request.yaml`.
- Use namespace `demo` for isolation.

```bash
kubectl create ns demo
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply ReconfigureTLS OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/reconfigure-tls/ops-request.yaml
kubectl get neo4jopsrequest -n demo neo4j-reconfigure-tls
```

## Verify

```bash
kubectl describe neo4jopsrequest -n demo neo4j-reconfigure-tls
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-reconfigure-tls
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
