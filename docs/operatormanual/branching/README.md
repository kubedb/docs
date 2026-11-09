---
title: Database Branching
menu:
  docs_{{ .version }}:
    identifier: branching
    name: Branching
    parent: operatormanual
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
url: /docs/{{ .version }}/operatormanual/branching/
aliases:
  - /docs/{{ .version }}/operatormanual/branching/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Database Branching

KubeDB Courier creates independent database copies from storage snapshots. A `Branch` resource describes the source database, the target, and optional steps such as password reset, post-actions, and scheduled refresh.

Start with the [overview](/docs/operatormanual/branching/overview.md), then configure Courier for your environment:

- [Same-cluster setup and configuration](/docs/operatormanual/branching/setup.md).
- [Cross-cluster setup and configuration](/docs/operatormanual/branching/setup-ocm.md) with an existing Open Cluster Management (OCM) installation.

Read the feature guides for [snapshot configuration](/docs/operatormanual/branching/snapshots.md), [refresh](/docs/operatormanual/branching/refresh.md), [ResetRootPassword](/docs/operatormanual/branching/reset-root-password.md), and [post-actions](/docs/operatormanual/branching/post-actions.md).

For complete manifests and database-specific behavior, follow the [PostgreSQL branching guides](/docs/guides/postgres/branch/overview/index.md).
