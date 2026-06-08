---
title: Recommendation
menu:
  docs_{{ .version }}:
    identifier: recommendation
    name: Recommendation
    parent: operatormanual
    weight: 10
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

# Recommendation for KubeDB

KubeDB **Recommendations** are Kubernetes-native, declarative suggestions for routine database maintenance — version updates, TLS certificate rotations, and authentication secret rotations — generated automatically by the KubeDB Ops-manager and executed by the Supervisor.

Use the pages below to learn the model, configure scheduling, and walk through each recommendation type end-to-end.

## Concepts & configuration

- [Overview](/docs/operatormanual/recommendation/overview.md) — what a Recommendation is and how it flows through the system
- [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md) — complete field reference for the Recommendation CRD
- [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) — namespace-scoped scheduling for automatic operations
- [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md) — cluster-wide default scheduling
- [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) — link maintenance windows to resources for automatic execution

## Recommendation types

1. [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)
2. [TLS Certificate Rotation Recommendation](/docs/operatormanual/recommendation/rotate-tls-recommendation.md)
3. [Authentication Secret Rotation Recommendation](/docs/operatormanual/recommendation/rotate-auth-recommendation.md)
