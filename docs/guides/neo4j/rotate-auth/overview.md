---
title: Rotating Neo4j Credentials
menu:
  docs_{{ .version }}:
    identifier: neo4j-rotate-auth-overview
    name: Overview
    parent: neo4j-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Neo4j

This guide shows how to rotate Neo4j credentials with `Neo4jOpsRequest`.

## Before You Begin

- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/rotate-auth/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply RotateAuth OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/rotate-auth/ops-request.yaml
kubectl get neo4jopsrequest -n demo neo4j-rotate-auth
```

## Verify

```bash
kubectl describe neo4jopsrequest -n demo neo4j-rotate-auth
kubectl get secret -n demo neo4j-test-auth -o yaml
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-rotate-auth
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
