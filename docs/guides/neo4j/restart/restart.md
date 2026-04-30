---
title: Restart Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-restart-overview
    name: Restart Neo4j
    parent: neo4j-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Neo4j

This guide shows how to restart Neo4j pods using `Neo4jOpsRequest`.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured.
- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/restart/ops-request.yaml`.
- Create an isolated namespace:

```bash
kubectl create ns demo
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply Restart OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/restart/ops-request.yaml
```

## Verify

```bash
kubectl get neo4jopsrequest -n demo neo4j-restart
kubectl describe neo4jopsrequest -n demo neo4j-restart
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-restart
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
