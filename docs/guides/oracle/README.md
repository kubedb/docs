---
title: Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-readme
    name: Oracle
    parent: guides-oracle
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/oracle/
aliases:
  - /docs/{{ .version }}/guides/oracle/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Oracle Features

| Features                                          | Availability |
|---------------------------------------------------|:------------:|
| Deployment Modes (Standalone & DataGuard)         |   &#10003;   |
| Physical Standby & Synchronous Replication        |   &#10003;   |
| Automatic Failover (FSFO)                         |   &#10003;   |
| Persistent Volume                                 |   &#10003;   |
| Resource Management (CPU/Memory)                  |   &#10003;   |
| Deletion Policy                                   |   &#10003;   |

## Life Cycle of a Oracle Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/oracle/oracle_lifecycle.png">
</p>

## User Guide

- [Quickstart Oracle](/docs/guides/oracle/quickstart/guide.md) with KubeDB Operator.
- [Oracle CRD Concepts](/docs/guides/oracle/concepts/oracle.md) - Understand the Oracle CRD specification.
- [Failover and Disaster Recovery](/docs/guides/oracle/failover/overview.md) - Data Guard based HA and auto-failover.


