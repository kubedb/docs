---
title: Rotate Auth of Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-rotate-auth-cluster
    name: Cluster
    parent: neo4j-rotate-auth
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Neo4j

This guide shows how to rotate database authentication secrets for Neo4j using `Neo4jOpsRequest`.

## Rotate Credentials

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: neo4j-prod
```

```bash
$ kubectl apply -f neo4j-rotate-auth.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-rotate-auth created
```

## Verify

```bash
$ kubectl get neo4jopsrequest -n demo neo4j-rotate-auth
NAME                TYPE         STATUS       AGE
neo4j-rotate-auth   RotateAuth   Successful   2m
```
