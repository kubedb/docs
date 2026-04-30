---
title: Qdrant Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-monitoring-overview
    name: Overview
    parent: qdrant-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Qdrant Monitoring

This guide shows how to enable and verify monitoring for Qdrant.

## Before You Begin

- Deploy Qdrant first using the [quickstart guide](/docs/guides/qdrant/quickstart/quickstart.md).
- Install Prometheus Operator or another compatible metrics stack.

## Enable Monitoring

Qdrant monitoring can be configured through `spec.monitor`.

- Enable metric scraping and verify health endpoints.
- Track shard status and node health in distributed mode.

## Verify

```bash
kubectl get qdrant -n demo qdrant-sample -o yaml
kubectl get servicemonitor -A
```

## Next Steps

- Add dashboards for shard balance, request latency, and storage growth.
- Recheck metrics after scaling or version updates.
