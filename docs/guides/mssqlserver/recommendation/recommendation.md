---
title: MSSQLServer Recommendation Overview
menu:
  docs_{{ .version }}:
    identifier: mssqlserver-recommendation-overview
    name: Recommendation Overview
    parent: mssqlserver-recommendation
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServer Recommendation

## Overview

A `Recommendation` is a Kubernetes-native CRD created by the **KubeDB Ops-Manager** and reconciled by the **KubeDB Supervisor**. For a MSSQLServer instance managed by KubeDB, the Ops-Manager watches the database's state and emits a Recommendation whenever it detects an action you should take — a newer version, an expiring TLS certificate, or an authentication secret nearing its rotation deadline.

Nothing runs until the Recommendation is approved — either by you (`status.approvalStatus: Approved`) or automatically through an `ApprovalPolicy` bound to a `MaintenanceWindow`. Once approved, the Supervisor creates the corresponding `MSSQLServerOpsRequest` and tracks it to completion.

This page is the **MSSQLServer-specific intro**: which recommendations apply to MSSQLServer and which spec fields trigger them. For prerequisites, Helm flags that control generation timing, and the full Recommendation lifecycle, see:

- [Recommendation Configuration](/docs/operatormanual/recommendation/configuration.md) — prerequisites, Supervisor CRD install, and all Helm flags.
- [Recommendation Overview](/docs/operatormanual/recommendation) — architecture and lifecycle walkthrough.

<p align="center">
  <img alt="Recommendation Lifecycle" src="/docs/operatormanual/recommendation/images/recommendation-generation.png">
</p>

---

## Relevant KubeDB concepts

- [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
- [MSSQLServerRotateAuth](/docs/guides/mssqlserver/rotate-auth/overview.md)
- [MSSQLServerReconfigureTLS](/docs/guides/mssqlserver/tls/overview.md)
- [MSSQLServerUpdateVersion](/docs/guides/mssqlserver/update-version/overview.md)

---

## Recommendation types for MSSQLServer

| Type                              | Triggered when                                                                    | Walkthrough                                                                                                  |
| --------------------------------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| **Version Update**                | A newer major, minor, or patch version becomes available                          | [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)        |
| **Same-Version Update**           | The container image for your *current* version is refreshed (e.g. security patch) | [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)        |
| **TLS Certificate Rotation**      | An issued certificate is approaching its expiry threshold                         | [TLS Certificate Rotation Recommendation](/docs/operatormanual/recommendation/rotate-tls-recommendation.md)  |
| **Authentication Secret Rotation** | The auth secret is approaching its `rotateAfter` deadline                         | [Authentication Secret Rotation Recommendation](/docs/operatormanual/recommendation/rotate-auth-recommendation.md) |

---

## Triggers specific to MSSQLServer

This section shows the minimal MSSQLServer CR fields that cause each recommendation to be generated. For deeper, end-to-end walkthroughs use the links in the table above.

### Authentication Secret Rotation

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-recommendation
  namespace: demo
spec:
  version: "2022-cu12"
  authSecret:
    kind: Secret
    name: mssqlserver-auth
    rotateAfter: 1h
```

In this configuration:

* The `rotateAfter` field defines how long the authentication secret remains valid

KubeDB monitors the configured lifecycle and generates a RotateAuth Recommendation based on the following conditions:

* If the secret lifespan is greater than one month, a recommendation is generated when less than one month of validity remains

* If the secret lifespan is less than one month, a recommendation is generated when approximately one-third of its validity remains

Once approved, KubeDB creates an opsrequest to rotate the credentials automatically, ensuring:

* No expired credentials

* Improved security posture

* Reduced manual intervention

### TLS Certificate Rotation

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-recommendation
  namespace: demo
spec:
  version: "2022-cu12"
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssqlserver-ca-issuer
    certificates:
      - alias: client
        duration: 1h20m
      - alias: server
        duration: 2h10m
```

In this configuration:

* The `spec.tls.certificates.duration` field defines how long each certificate remains valid

KubeDB monitors the configured lifecycle and generates a RotateTLS Recommendation based on the following conditions:

* If the certificate duration is greater than one month, a recommendation is generated when less than one month of validity remains

* If the certificate duration is less than one month, a recommendation is generated when approximately one-third of its validity remains

Once approved, KubeDB creates an opsrequest to reconfigure TLS automatically, ensuring:

* Continuous secure communication

* No unexpected certificate expiry

* Seamless certificate renewal

### Version Update

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-recommendation
  namespace: demo
spec:
  version: "2022-cu12"
```

In this configuration:

* KubeDB monitors the running version of the database

KubeDB monitors the configured lifecycle and generates a VersionUpdate Recommendation based on the following conditions:

* If a newer container image is available for the current version, a recommendation is generated

* If a patch version is released, a recommendation is generated

* If a newer minor or major version becomes available, a recommendation is generated

* If changes are introduced in the existing version image (e.g., security fixes or image updates without a version bump), a recommendation is generated

For example: Recommending version update from `2022-cu12` to `2022-cu14`

Once approved, KubeDB creates an opsrequest to perform the version upgrade automatically, ensuring:

* Timely adoption of security patches and fixes

* Access to new features and improvements

* Consistent performance and stability across deployments

### Same-Version Update

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-recommendation
  namespace: demo
spec:
  version: "2022-cu12"
```

In this configuration:

* KubeDB monitors the container image of the current database version

KubeDB monitors the configured lifecycle and generates a SameVersionUpdate Recommendation based on the following conditions:

* If the container image backing the current version is updated (e.g., security patches or rebuilds without a version change), a recommendation is generated

Once approved, KubeDB creates an opsrequest to update the running workload automatically, ensuring:

* Security patches are applied without requiring a version upgrade

* Consistency with the latest available container image

* Improved reliability and maintainability

---

For prerequisites, Helm configuration flags, and the full cross-database Recommendation lifecycle, see the [Recommendation Configuration](/docs/operatormanual/recommendation/configuration.md) and [Recommendation Overview](/docs/operatormanual/recommendation) in the operator manual.
