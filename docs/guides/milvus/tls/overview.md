---
title: Milvus TLS Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-tls-overview
    name: Overview
    parent: milvus-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Milvus TLS

This guide shows the main TLS considerations for a Milvus deployment.

## Before You Begin

- Install `cert-manager` in your cluster.
- Deploy Milvus first using the [quickstart guide](/docs/guides/milvus/quickstart/quickstart.md).

## Configure TLS

Milvus TLS setup is managed with cert-manager and KubeDB TLS settings.

- Use a trusted issuer for server certificates.
- Rotate certificates in a controlled maintenance window.

## Verify

```bash
kubectl get milvus -n demo milvus-cluster -o yaml
kubectl get secret -n demo
```

## Next Steps

- Test client connectivity with TLS enabled before routing production traffic.
- Document the certificate rotation process for your team.
