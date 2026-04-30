---
title: DocumentDB
menu:
  docs_{{ .version }}:
    identifier: documentdb-readme
    name: DocumentDB
    parent: documentdb-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/documentdb/
aliases:
  - /docs/{{ .version }}/guides/documentdb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports Amazon DocumentDB-compatible workloads through the `DocumentDB` CRD.

## Supported DocumentDB Features

| Features                          | Availability |
|-----------------------------------|:------------:|
| Standalone provisioning           |   &#10003;   |
| Persistent volume                 |   &#10003;   |
| Replicas                          |   &#10003;   |
| Deletion policy                   |   &#10003;   |

## Example DocumentDB Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb
spec:
  version: "pg17-0.109.0"
  storageType: Durable
  deletionPolicy: Delete
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

## User Guide

- [Quickstart DocumentDB](/docs/guides/documentdb/quickstart/quickstart.md) with KubeDB operator.
- [DocumentDB CRD Concept](/docs/guides/documentdb/concepts/documentdb.md).
- [DocumentDBVersion CRD Concept](/docs/guides/documentdb/concepts/catalog.md).
- [DocumentDBOpsRequest CRD Concept](/docs/guides/documentdb/concepts/opsrequest.md).
- [Configuration](/docs/guides/documentdb/configuration/overview.md) for replicas, auth, pod, and service settings.
- [Ops Request](/docs/guides/documentdb/ops-request/overview.md).