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

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB Monitoring

KubeDB has native support for monitoring HanaDB via [Prometheus](https://prometheus.io/). You can use built-in Prometheus discovery or Prometheus Operator to monitor KubeDB-managed HanaDB instances.

## Before You Begin

- Deploy HanaDB first using the [quickstart guide](/docs/guides/hanadb/quickstart/quickstart.md).
- Install Prometheus Operator or another monitoring stack that can scrape Kubernetes services.

## Overview

When you create a HanaDB object with `spec.monitor` configured, KubeDB injects the HanaDB exporter sidecar into the database pod. KubeDB also creates a stats service named `{hanadb-name}-stats` for scraping the exporter endpoint.

## Configure Monitoring

HanaDB monitoring is configured via `spec.monitor`.

| Field                                             | Type       | Uses |
|---------------------------------------------------|------------|------|
| `spec.monitor.agent`                              | `Required` | Monitoring agent type. Use `prometheus.io/builtin` or `prometheus.io/operator`. |
| `spec.monitor.prometheus.exporter.port`           | `Optional` | Port where the exporter sidecar serves metrics. Defaults to `9668`. |
| `spec.monitor.prometheus.exporter.args`           | `Optional` | Arguments passed to the exporter sidecar. |
| `spec.monitor.prometheus.exporter.env`            | `Optional` | Environment variables set in the exporter sidecar. |
| `spec.monitor.prometheus.exporter.resources`      | `Optional` | Resource requirements for the exporter sidecar. |
| `spec.monitor.prometheus.serviceMonitor.labels`   | `Optional` | Labels added to the `ServiceMonitor` for Prometheus selection. |
| `spec.monitor.prometheus.serviceMonitor.interval` | `Optional` | Scrape interval for Prometheus Operator. |

## Sample Configuration

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-prometheus-operator
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9668
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

## Next Steps

- Learn how to monitor HanaDB using [Prometheus Operator](/docs/guides/hanadb/monitoring/using-prometheus-operator.md).
- Learn how to monitor HanaDB using [built-in Prometheus discovery](/docs/guides/hanadb/monitoring/using-builtin-prometheus.md).
