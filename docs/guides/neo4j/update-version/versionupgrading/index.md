---
title: Upgrade Neo4j Version
menu:
  docs_{{ .version }}:
    identifier: neo4j-version-upgrading
    name: Version Upgrading
    parent: neo4j-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Upgrade Neo4j Version

This guide shows how to upgrade Neo4j using `Neo4jOpsRequest`.

## Update Version Request

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-version-upgrade
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: neo4j-prod
  updateVersion:
    targetVersion: "2025.11.3"
```

```bash
$ kubectl apply -f neo4j-version-upgrade.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-version-upgrade created
```
