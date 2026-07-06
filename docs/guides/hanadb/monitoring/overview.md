---
title: HanaDB Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-monitoring-overview
    name: Overview
    parent: guides-hanadb-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring HanaDB

KubeDB exposes Prometheus metrics for HanaDB through a bundled
[hanadb_exporter](https://github.com/kubedb/hanadb-exporter) sidecar. You enable it through the
`spec.monitor` field of the `HanaDB` object.

## How it works

When `spec.monitor` is set, KubeDB adds an `exporter` container to the database pods and a `<db>-stats`
Service that exposes the metrics endpoint (default port `9668`, path `/metrics`). Two agents are
supported:

- `prometheus.io/builtin` — annotates the stats Service so a Prometheus server that scrapes by
  annotation discovers the target. See [Using Builtin Prometheus](/docs/guides/hanadb/monitoring/using-builtin-prometheus.md).
- `prometheus.io/operator` — creates a `ServiceMonitor` for the
  [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator). See
  [Using Prometheus Operator](/docs/guides/hanadb/monitoring/using-prometheus-operator.md).

## The spec.monitor field

```yaml
spec:
  monitor:
    agent: prometheus.io/operator   # or prometheus.io/builtin
    prometheus:
      exporter:
        port: 9668
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

- `spec.monitor.agent` selects the monitoring agent.
- `spec.monitor.prometheus.exporter.port` is the exporter port (`9668`).
- `spec.monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` of your
  Prometheus Operator instance (commonly `release: prometheus`).

## Next Steps

- [Using Builtin Prometheus](/docs/guides/hanadb/monitoring/using-builtin-prometheus.md).
- [Using Prometheus Operator](/docs/guides/hanadb/monitoring/using-prometheus-operator.md).
