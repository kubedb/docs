---
title: Reconfiguring Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-overview
    name: Overview
    parent: neo4j-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Neo4j

This guide shows how to reconfigure Neo4j using `Neo4jOpsRequest`.

## Before You Begin

- Be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/reconfigure/ops-request.yaml`.
- Create namespace:

```bash
kubectl create ns demo
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply Reconfigure OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/reconfigure/ops-request.yaml
```

## Verify

```bash
kubectl get neo4jopsrequest -n demo neo4j-reconfigure
kubectl describe neo4jopsrequest -n demo neo4j-reconfigure
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-reconfigure
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
