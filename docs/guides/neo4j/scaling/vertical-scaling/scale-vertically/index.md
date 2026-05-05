---
title: Scale Neo4j Vertically
menu:
  docs_{{ .version }}:
    identifier: neo4j-scale-vertically
    name: Scale Vertically
    parent: neo4j-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling for Neo4j

This guide shows how to vertically scale Neo4j using `Neo4jOpsRequest`.

## Vertical Scaling Request

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: neo4j-test
  verticalScaling:
    server:
      resources:
        limits:
          cpu: 1500m
          memory: 4Gi
        requests:
          cpu: 700m
          memory: 4Gi
```

```bash
$ kubectl apply -f neo4j-vertical-scale.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-vertical-scale created
```
