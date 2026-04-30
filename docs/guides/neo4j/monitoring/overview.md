---
title: Neo4j Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-monitoring-overview
    name: Overview
    parent: neo4j-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Neo4j Monitoring

This guide shows how to enable and verify monitoring for Neo4j.

## Before You Begin

- Deploy Neo4j first using the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md).
- Install Prometheus Operator or another compatible metrics stack.

## Enable Monitoring

Neo4j monitoring can be enabled via `spec.monitor`.

- Integrate with Prometheus operator service monitors.
- Verify target ports for selected Neo4j protocols.

## Verify

```bash
kubectl get neo4j -n demo neo4j-test -o yaml
kubectl get servicemonitor -A
```

## Next Steps

- Add dashboards for cluster health, storage pressure, and protocol-specific traffic.
- Recheck scrape targets after scaling or version updates.
