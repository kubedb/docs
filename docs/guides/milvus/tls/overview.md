---
title: Milvus TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-tls-overview
    name: Overview
    parent: milvus-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus TLS/SSL Encryption

KubeDB supports providing TLS/SSL encryption for `Milvus`. This guide will give you an overview of how it works.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)

## How TLS/SSL Configures in Milvus

KubeDB uses [cert-manager](https://cert-manager.io/) to provision and manage the certificates used by Milvus. You point the database at a cert-manager `Issuer` or `ClusterIssuer` through `spec.tls.issuerRef`, and KubeDB requests the required certificates on your behalf.

```yaml
spec:
  tls:
    issuerRef:
      name: milvus-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    external:
      mode: mTLS
    internal:
      mode: TLS
    certificates:
      - alias: server
      - alias: client
```

- **`spec.tls.issuerRef`** references the cert-manager `Issuer`/`ClusterIssuer` that signs the certificates.
- **`spec.tls.external`** controls client-facing traffic. `mode` can be `Disabled`, `TLS` (server-side only) or `mTLS` (mutual TLS — clients must present a certificate).
- **`spec.tls.internal`** controls inter-component traffic between Milvus roles. `mode` can be `Disabled` or `TLS`.
- **`spec.tls.certificates`** lets you customize the two certificate aliases used by Milvus:
  - `server` — the server certificate, stored in the `<db>-server-cert` secret.
  - `client` — the client certificate, stored in the `<db>-client-cert` secret.

When TLS is enabled, KubeDB:

1. Requests the `server` and `client` certificates from cert-manager.
2. Mounts them, together with the CA, into every Milvus pod at `/milvus/tls` (`ca.pem`, `server.pem`, `server.key`, `client.pem`, `client.key`).
3. Renders the appropriate `internaltls`, `tlsMode` and `common.security` settings into `milvus.yaml`.
4. Sets the connection scheme in the `AppBinding` to `https`.

In the next doc, we will see a step-by-step guide on deploying a TLS-secured Milvus database.
