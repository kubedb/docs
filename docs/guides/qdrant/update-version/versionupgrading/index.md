---
title: Upgrade Qdrant Version
menu:
  docs_{{ .version }}:
    identifier: qdrant-version-upgrading
    name: Version Upgrading
    parent: qdrant-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Upgrade Qdrant Version

This guide shows how to upgrade Qdrant version using `QdrantOpsRequest`.

## Before You Begin

- Ensure target version exists in `QdrantVersion` catalog.
- Ensure the database is in `Ready` state.

```bash
$ kubectl get qdrantversions
```

## Apply UpdateVersion OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-version-upgrade
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: qdrant-sample
  updateVersion:
    targetVersion: "1.18.0"
```

```bash
$ kubectl apply -f qdrant-version-upgrade.yaml
qdrantopsrequest.ops.kubedb.com/qdrant-version-upgrade created
```

## Verify

```bash
$ kubectl get qdrantopsrequest -n demo qdrant-version-upgrade
$ kubectl get qdrant -n demo qdrant-sample
```