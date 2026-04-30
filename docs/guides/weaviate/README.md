---
title: Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-readme
    name: Weaviate
    parent: weaviate-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/weaviate/
aliases:
  - /docs/{{ .version }}/guides/weaviate/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports Weaviate through the `Weaviate` CRD.

## Supported Weaviate Features

| Features                  | Availability |
|---------------------------|:------------:|
| Standalone provisioning   |   &#10003;   |
| Cluster provisioning      |   &#10003;   |
| Ops Requests              |      No      |

## Supported Ops Requests

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD. The existing ops pages are placeholder documentation only.

## Example Weaviate Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

## User Guide

- [Quickstart Weaviate](/docs/guides/weaviate/quickstart/quickstart.md) with KubeDB operator.
- [Weaviate CRD Concept](/docs/guides/weaviate/concepts/weaviate.md).
- [WeaviateVersion CRD Concept](/docs/guides/weaviate/concepts/catalog.md).
- [WeaviateOpsRequest CRD Concept](/docs/guides/weaviate/concepts/opsrequest.md).
- [Private Registry](/docs/guides/weaviate/private-registry/using-private-registry.md)
- [Custom RBAC](/docs/guides/weaviate/custom-rbac/using-custom-rbac.md)
- [Custom Configuration](/docs/guides/weaviate/configuration/using-config-file.md)
- [Ops Request](/docs/guides/weaviate/ops-request/overview.md) for current documentation status.
- [Reconfigure](/docs/guides/weaviate/reconfigure/overview.md)
- [Restart](/docs/guides/weaviate/restart/restart.md)
- [Rotate Auth](/docs/guides/weaviate/rotate-auth/overview.md)
- [Update Version](/docs/guides/weaviate/update-version/overview.md)
- [Volume Expansion](/docs/guides/weaviate/volume-expansion/overview.md)
- [Vertical Scaling](/docs/guides/weaviate/scaling/vertical-scaling/overview.md)