---
title: Neo4j Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-vertical-scaling-overview
    name: Overview
    parent: neo4j-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Vertical Scaling

This guide shows how to update CPU and memory for Neo4j pods.

## Before You Begin

- Ensure database is healthy and all pods are running.
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/scaling/vertical-scaling/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply VerticalScaling OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/scaling/vertical-scaling/ops-request.yaml
kubectl get neo4jopsrequest -n demo neo4j-vertical-scale
```

## Verify

```bash
kubectl describe neo4jopsrequest -n demo neo4j-vertical-scale
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-vertical-scale
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
