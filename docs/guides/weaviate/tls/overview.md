---
title: Weaviate TLS Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-tls-overview
    name: Overview
    parent: weaviate-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate TLS/SSL Encryption

This guide will give an overview of how KubeDB supports TLS/SSL encryption for `Weaviate` databases.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)

## How TLS Works for Weaviate

KubeDB uses `cert-manager` to manage TLS certificates for Weaviate databases. The TLS configuration process consists of the following steps:

1. At first, a user creates a `ClusterIssuer` or `Issuer` using `cert-manager`.

2. The user then creates a `Weaviate` CR with the `spec.tls` field configured, pointing to the `Issuer` or `ClusterIssuer`.

3. `KubeDB-Provisioner` operator watches the `Weaviate` CR.

4. When the operator finds a `Weaviate` CR with `spec.tls` configured, it requests TLS certificates from `cert-manager` using the specified issuer.

5. `cert-manager` creates the certificates and stores them in a `Secret`.

6. `KubeDB-Provisioner` operator creates the `StatefulSet` with the TLS secrets mounted, enabling encrypted communication.

7. The `Weaviate` database nodes use these certificates for encrypted client-to-server and peer-to-peer communication.

KubeDB supports the following TLS configurations for Weaviate:

- **Add TLS** — Enable TLS on an existing non-TLS Weaviate database using a `WeaviateOpsRequest`.
- **Rotate TLS** — Rotate the existing TLS certificates to refresh expiring certificates.
- **Remove TLS** — Remove TLS from an existing TLS-enabled Weaviate database.

In the next doc, we are going to show a step-by-step guide on configuring TLS for a Weaviate database.
