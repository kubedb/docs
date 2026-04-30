---
title: Expanding Neo4j Storage
menu:
  docs_{{ .version }}:
    identifier: neo4j-volume-expansion-overview
    name: Overview
    parent: neo4j-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Volume Expansion for Neo4j

This guide shows how to expand Neo4j persistent storage.

## Before You Begin

- Ensure StorageClass supports `allowVolumeExpansion: true`.
- Use the example files from `docs/examples/neo4j/quickstart/neo4j.yaml` and `docs/examples/neo4j/volume-expansion/ops-request.yaml`.

```bash
kubectl create ns demo
kubectl get storageclass
```

## Deploy Neo4j

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Apply VolumeExpansion OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/volume-expansion/ops-request.yaml
kubectl get neo4jopsrequest -n demo neo4j-volume-expand
```

## Verify

```bash
kubectl describe neo4jopsrequest -n demo neo4j-volume-expand
kubectl get pvc -n demo
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-volume-expand
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```
