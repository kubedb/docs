---
title: HanaDB TLS Overview
menu:
  docs_{{ .version }}:
    identifier: hanadb-tls-overview
    name: Overview
    parent: hanadb-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# HanaDB TLS

KubeDB supports TLS for HanaDB through `spec.tls`. When TLS is configured, KubeDB uses cert-manager to issue certificates and configures SAP HANA SQL traffic, KubeDB client connections, and the metrics exporter with TLS.

## Prerequisites

Install cert-manager before creating a TLS-enabled HanaDB.

## HanaDB TLS Fields

HanaDB uses the common KubeDB TLS configuration under `spec.tls`.

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: hdb-ca-issuer
```

`spec.tls.issuerRef` must reference an `Issuer` or `ClusterIssuer` that cert-manager can use to sign certificates.

## Day-2 TLS Changes

To add TLS to an existing HanaDB or rotate existing certificates, use a `HanaDBOpsRequest` of type `ReconfigureTLS`. See [Reconfigure HanaDB TLS](/docs/guides/hanadb/reconfigure-tls/reconfigure-tls.md).
