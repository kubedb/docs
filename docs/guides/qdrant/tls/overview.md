---
title: Qdrant TLS Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-tls-overview
    name: Overview
    parent: qdrant-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant TLS/SSL Encryption

This guide will give an overview of how KubeDB supports TLS/SSL encryption for `Qdrant` databases.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How TLS Works for Qdrant

KubeDB uses `cert-manager` to manage TLS certificates for Qdrant databases. The TLS configuration process consists of the following steps:

1. At first, a user creates a `ClusterIssuer` or `Issuer` using `cert-manager`.

2. The user then creates a `Qdrant` CR with the `spec.tls` field configured, pointing to the `Issuer` or `ClusterIssuer`.

3. `KubeDB-Provisioner` operator watches the `Qdrant` CR.

4. When the operator finds a `Qdrant` CR with `spec.tls` configured, it requests TLS certificates from `cert-manager` using the specified issuer.

5. `cert-manager` creates the certificates and stores them in a `Secret`.

6. `KubeDB-Provisioner` operator creates the `StatefulSet` with the TLS secrets mounted, enabling encrypted communication.

7. The `Qdrant` database nodes use these certificates for encrypted client-to-server and peer-to-peer communication.

KubeDB supports the following TLS configurations for Qdrant:

- **Add TLS** — Enable TLS on an existing non-TLS Qdrant database using a `QdrantOpsRequest`.
- **Rotate TLS** — Rotate the existing TLS certificates to refresh expiring certificates.
- **Remove TLS** — Remove TLS from an existing TLS-enabled Qdrant database.

In the next doc, we are going to show a step-by-step guide on configuring TLS for a Qdrant database.
