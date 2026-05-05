---
title: Expand Neo4j Volume
menu:
  docs_{{ .version }}:
    identifier: neo4j-volume-expansion-cluster
    name: Cluster
    parent: neo4j-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand Neo4j Volume

This guide shows how to expand storage for Neo4j using `Neo4jOpsRequest`.

## Volume Expansion Request

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: neo4j-test
  volumeExpansion:
    mode: "Online"
    server: 4Gi
```

```bash
$ kubectl apply -f neo4j-volume-expansion.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-volume-expansion created
```
