---
title: Milvus Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-monitoring-overview
    name: Overview
    parent: milvus-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Milvus Monitoring

This guide shows how to enable and verify monitoring for Milvus deployments.

## Before You Begin

- Deploy Milvus first using the [quickstart guide](/docs/guides/milvus/quickstart/quickstart.md).
- Install Prometheus Operator or another compatible metrics collection stack.

## Enable Monitoring

Milvus supports monitoring integration through `spec.monitor`.

- Configure monitoring agent and service monitor selectors.
- Validate scrape targets for proxy and distributed components.

## Verify

```bash
kubectl get milvus -n demo milvus-cluster -o yaml
kubectl get servicemonitor -A
```

## Next Steps

- Review which Milvus components need dedicated scrape targets in distributed mode.
- Add dashboards for latency, indexing throughput, and node health.
