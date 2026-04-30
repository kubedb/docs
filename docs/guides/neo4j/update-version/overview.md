---
title: Updating Neo4j Version
menu:
  docs_{{ .version }}:
    identifier: neo4j-update-version-overview
    name: Overview
    parent: neo4j-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Neo4j Version

This guide shows how to update Neo4j version using `Neo4jOpsRequest`.

## Before You Begin

- Ensure Neo4j database is `Ready`.
- Ensure target version exists in `Neo4jVersion`.
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/update-version/ops-request.yaml`.

```bash
kubectl create ns demo
kubectl get neo4jversions
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply UpdateVersion OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/update-version/ops-request.yaml
kubectl get neo4jopsrequest -n demo neo4j-update-version
kubectl describe neo4jopsrequest -n demo neo4j-update-version
```

## Verify

```bash
kubectl get neo4j -n demo neo4j-test -o jsonpath='{.spec.version}{"\n"}'
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-update-version
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
