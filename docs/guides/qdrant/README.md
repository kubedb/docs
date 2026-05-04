---
title: Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-readme
    name: Qdrant
    parent: qdrant-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/qdrant/
aliases:
  - /docs/{{ .version }}/guides/qdrant/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports Qdrant vector databases through the `Qdrant` CRD.

## Supported Qdrant Features

| Features                 | Availability |
|--------------------------|:------------:|
| Standalone provisioning  |   &#10003;   |
| Distributed provisioning |   &#10003;   |
| TLS                      |   &#10003;   |
| Logical Backup           |   &#10003;   |
| Volume Snapshot          |   &#10003;   |
| Monitoring               |   &#10003;   |
| Ops Requests             |   &#10003;   |
| Autoscaler               |   &#10003;   |


## Life Cycle of a Qdrant Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/qdrant/images/qdrant-lifecycle.png" >
</p>


## User Guide

- [Quickstart Qdrant](/docs/guides/qdrant/quickstart/quickstart.md) with KubeDB operator.
- Deploy [Distributed Qdrant](/docs/guides/qdrant/distributed-deployment/overview.md) cluster.
- Detail concepts of [Qdrant Object](/docs/guides/qdrant/concepts/).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
