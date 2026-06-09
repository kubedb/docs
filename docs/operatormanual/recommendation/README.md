---
title: Recommendation
menu:
  docs_{{ .version }}:
    identifier: readme-recommendation
    name: Recommendation
    parent: recommendation
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
url: /docs/{{ .version }}/operatormanual/recommendation/
aliases:
  - /docs/{{ .version }}/operatormanual/recommendation/README/
---



> New to KubeDB? Please start [here](/docs/README.md).

# Recommendation for KubeDB

Production databases on Kubernetes need regular, careful maintenance — security patches, version upgrades, TLS certificate rotations, and credential rotations. Skipping them risks exposure to known CVEs, expired certificates that break clients, and stale secrets that violate compliance. Doing them by hand is error-prone and easy to forget.

KubeDB solves this by **generating recommendations automatically**, as Kubernetes-native CRDs, whenever a managed database needs a maintenance action. The Supervisor then executes the recommendation either immediately, on operator approval, or inside a scheduled maintenance window — with full status tracked on the resource itself.

## Why it matters

- **Security** — older versions carry known vulnerabilities; expiring TLS certificates cause outages; stale auth secrets are an obvious attack surface.
- **Compliance** — auditors expect documented, repeatable rotation policies.
- **Operational safety** — execution is bounded by deadlines, retries, and (optionally) operator-approved windows, so disruptive ops never run at peak hours.

## How recommendations flow

A `Recommendation` is a custom resource created by the KubeDB **Ops-manager** and reconciled by the **Supervisor**. You need both installed; the easiest path is to enable the Supervisor when installing KubeDB via Helm:

```bash
--set supervisor.enabled=true
```

<p align="center">
<img alt="Recommendation Generation"  src="/docs/operatormanual/recommendation/images/recommendation-generation.png">
</p>

1. The **KubeDB Provisioner** reconciles user-provided database CRs and creates all required resources.
2. Once the database is `Ready`, the **Ops-manager** inspects it and creates a `Recommendation` whenever an action is needed (vulnerable version, certificate near expiry, auth secret near rotation deadline, …).
3. The **Supervisor** watches the Recommendation, applies approval policies, waits for the configured maintenance window, and then creates the corresponding `OpsRequest` (e.g. `UpdateVersion`, `ReconfigureTLS`, `RotateAuth`).
4. The Supervisor watches the OpsRequest and updates the Recommendation status (`Succeeded`, `Failed`, `Skipped`, …) so the whole lifecycle is visible on one object.

## Recommendation types

KubeDB generates three kinds of recommendations:

1. [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md) — upgrade to a patched/newer database version.
2. [TLS Certificate Rotation Recommendation](/docs/operatormanual/recommendation/rotate-tls-recommendation.md) — rotate certificates before expiry.
3. [Authentication Secret Rotation Recommendation](/docs/operatormanual/recommendation/rotate-auth-recommendation.md) — rotate database credentials.

## Configuring scheduling and approval

For automation and execution control, refer to:

- [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md) — complete field reference for the Recommendation CRD.
- [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) — namespace-scoped scheduling for automatic operations.
- [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md) — cluster-wide default maintenance scheduling.
- [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) — link maintenance windows to resources for automatic recommendation execution.

The following pages walk through each recommendation type, show how to approve or reject them, and explain how to automate execution with maintenance windows.
