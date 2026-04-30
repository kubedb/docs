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

# Qdrant TLS

This guide shows the key TLS considerations for Qdrant.

## Before You Begin

- Install `cert-manager` in your cluster.
- Deploy Qdrant first using the [quickstart guide](/docs/guides/qdrant/quickstart/quickstart.md).

## Configure TLS

Qdrant TLS configuration is available via `spec.tls`.

- Enable client and p2p certificate handling.
- Manage cert issuance using cert-manager issuer references.

## Verify

```bash
kubectl get qdrant -n demo qdrant-sample -o yaml
kubectl get secret -n demo
```

## Next Steps

- Test both client and peer communication after enabling TLS.
- Use the dedicated [Reconfigure TLS guide](/docs/guides/qdrant/reconfigure-tls/overview.md) for certificate rotation.
