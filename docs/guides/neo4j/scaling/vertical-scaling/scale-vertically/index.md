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
    name: neo4j-prod
  verticalScaling:
    server:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "1"
        memory: "2Gi"
```

```bash
$ kubectl apply -f neo4j-vertical-scale.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-vertical-scale created
```
