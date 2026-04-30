---
title: Neo4j Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-horizontal-scaling-overview
    name: Overview
    parent: neo4j-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Horizontal Scaling

This guide shows how to scale Neo4j cluster members horizontally.

## Before You Begin

- Ensure database is healthy (`status.phase=Ready`).
- Use odd replica counts for quorum-sensitive deployments.
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/scaling/horizontal-scaling/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply HorizontalScaling OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/scaling/horizontal-scaling/ops-request.yaml
kubectl get neo4jopsrequest -n demo neo4j-horizontal-scale
```

## Verify

```bash
kubectl describe neo4jopsrequest -n demo neo4j-horizontal-scale
kubectl get neo4j -n demo neo4j-test -o yaml
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-horizontal-scale
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
