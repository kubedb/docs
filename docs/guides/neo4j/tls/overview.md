---
title: Neo4j TLS Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-tls-overview
    name: Overview
    parent: neo4j-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Neo4j TLS

This guide shows the key TLS considerations for Neo4j.

## Before You Begin

- Install `cert-manager` in your cluster.
- Deploy Neo4j first using the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md).

## Configure TLS

Neo4j supports protocol-level TLS settings with `spec.tls`.

- Configure bolt, http, and cluster channel settings.
- Use mode values Disabled, TLS, or mTLS per protocol.

## Verify

```bash
kubectl get neo4j -n demo neo4j-test -o yaml
kubectl get secret -n demo
```

## Next Steps

- Test each enabled protocol after certificate changes.
- Use the dedicated [Reconfigure TLS guide](/docs/guides/neo4j/reconfigure-tls/overview.md) for post-deployment certificate rotation.
