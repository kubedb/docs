---
title: DB2
menu:
  docs_{{ .version }}:
    identifier: db2-readme
    name: DB2
    parent: db2-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/db2/
aliases:
  - /docs/{{ .version }}/guides/db2/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports IBM DB2 using the `DB2` Custom Resource Definition. You can declare the desired DB2 configuration, and KubeDB provisions and manages the required Kubernetes resources.

## Supported DB2 Features

| Features                       | Availability |
|--------------------------------|:------------:|
| Standalone DB2 deployment      |   &#10003;   |
| Persistent volume              |   &#10003;   |
| Authentication secret          |   &#10003;   |
| Pod and service customization  |   &#10003;   |
| Health checker                 |   &#10003;   |

## Example DB2 Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: db2
  namespace: demo
spec:
  version: 11.5.8.0
  storageType: Durable
  deletionPolicy: Delete
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

## User Guide

- [Quickstart DB2](/docs/guides/db2/quickstart/quickstart.md) with KubeDB operator.
- [DB2 CRD Concept](/docs/guides/db2/concepts/db2.md).
- [DB2Version CRD Concept](/docs/guides/db2/concepts/catalog.md).
- [DB2OpsRequest CRD Concept](/docs/guides/db2/concepts/opsrequest.md).
- [Configuration](/docs/guides/db2/configuration/overview.md) for auth, pod, service, and health settings.
- [Ops Request](/docs/guides/db2/ops-request/overview.md).