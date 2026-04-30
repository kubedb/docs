---
title: Scale Neo4j Horizontally
menu:
  docs_{{ .version }}:
    identifier: neo4j-scale-horizontally
    name: Scale Horizontally
    parent: neo4j-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling for Neo4j

This guide shows how to scale a Neo4j database horizontally using `Neo4jOpsRequest`.

## Horizontal Scaling Request

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-horizontal-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: neo4j-prod
  horizontalScaling:
    server: 3
```

```bash
$ kubectl apply -f neo4j-horizontal-scale.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-horizontal-scale created
```

## Verify

```bash
$ kubectl get neo4j -n demo neo4j-prod
NAME         VERSION   STATUS   AGE
neo4j-prod   2025.11.2 Ready    10m

$ kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-prod
NAME           READY   STATUS    RESTARTS   AGE
neo4j-prod-0   1/1     Running   0          10m
neo4j-prod-1   1/1     Running   0          4m
neo4j-prod-2   1/1     Running   0          4m
```
