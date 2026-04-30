---
title: HanaDB Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: hanadb-monitoring-overview
    name: Overview
    parent: hanadb-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# HanaDB Monitoring

This guide shows how to enable and verify monitoring for HanaDB.

## Before You Begin

- Deploy HanaDB first using the [quickstart guide](/docs/guides/hanadb/quickstart/quickstart.md).
- Install Prometheus Operator or another monitoring stack that can scrape Kubernetes targets.

## Enable Monitoring

HanaDB monitoring is configured via `spec.monitor`.

- Use Prometheus operator-based scraping where available.
- Ensure metrics endpoints are reachable from the monitoring namespace.

## Verify

```bash
kubectl get hanadb -n demo hana-cluster -o yaml
kubectl get servicemonitor -A
```

## Next Steps

- Add alerts for pod health, storage usage, and replication status.
- Review your Prometheus targets after every database topology change.
