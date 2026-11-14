---
title: YugabyteDB Monitoring Overview
description: YugabyteDB Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-monitoring-overview
    name: Overview
    parent: yugabytedb-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring YugabyteDB with KubeDB

YugabyteDB exposes Prometheus metrics natively, so KubeDB does not add an exporter sidecar. When `spec.monitor` is set, KubeDB creates a Service named `<name>-stats` that selects all pods of the universe, and, for the Prometheus operator, a `ServiceMonitor`.

## What was verified

- The `<name>-stats` Service and the metrics endpoint `/prometheus-metrics` of the master (port 7000) and tserver (port 9000) web servers were checked. For example, `curl localhost:9000/prometheus-metrics` inside a tserver pod returns metrics such as `drive_bytes_written{drive_path="/mnt/disk0",...}`.
- The `ServiceMonitor` was **not** verified in a cluster: the test cluster had no prometheus-operator. It is covered by unit tests that use the shared KubeDB monitoring agent.
- The built-in Prometheus agent (`prometheus.io/builtin`) is supported by the API, but it cannot restrict the master port to master pods, so the Prometheus operator is the recommended way. This was not run either.
- No Grafana dashboard for YugabyteDB ships with this documentation.

## Ports

| Port | Component | Metrics |
|------|-----------|---------|
| 7000 | master web server | `/prometheus-metrics` |
| 9000 | tserver web server | `/prometheus-metrics` |
| 13000 | tserver, YSQL | `/prometheus-metrics` |
| 12000 | tserver, YCQL (only if `spec.ycql.enabled`) | `/prometheus-metrics` |

The ServiceMonitor has one endpoint per Service port. It relabels the series with `group`, `export_type` (`master_export`, `tserver_export`, `ysql_export`, `cql_export`) and `node_prefix`, converts `handler_latency_*` metrics to `rpc_latency{server_type,service_type,service_method}`, and keeps the master port only on pods labelled as masters, so the usual community YugabyteDB dashboards work.

## Configure monitoring (not verified)

```yaml
apiVersion: kubedb.com/v1alpha2
kind: YugabyteDB
metadata:
  name: yb-demo
  namespace: demo
spec:
  version: 2025.2.8
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  # ... topology as in the quickstart
```

| Field | Type | Uses |
| --- | --- | --- |
| `spec.monitor.agent` | `Required` | `prometheus.io/operator` or `prometheus.io/builtin`. |
| `spec.monitor.prometheus.serviceMonitor.labels` | `Optional` | Labels of the `ServiceMonitor`, they must match the selector of your Prometheus. |
| `spec.monitor.prometheus.serviceMonitor.interval` | `Optional` | Scrape interval. |

With TLS enabled, the web servers stay plain HTTP inside the pod network.

## Check the metrics endpoint by hand

```bash
$ kubectl exec -n demo yb-demo-tserver-0 -c yb-tserver -- curl -s localhost:9000/prometheus-metrics | head -3
```

## Next Steps

- [YugabyteDB CRD](/docs/guides/yugabytedb/concepts/yugabytedb.md)
- [Cluster architecture](/docs/guides/yugabytedb/clustering/architecture-overview.md)
