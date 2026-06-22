---
title: Neo4j Monitoring Overview
description: Neo4j Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-monitoring-overview
    name: Overview
    parent: neo4j-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Neo4j with KubeDB

KubeDB has native support for monitoring via [Prometheus](https://prometheus.io/). You can use builtin [Prometheus](https://github.com/prometheus/prometheus) scraper or [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) to monitor KubeDB-managed Neo4j clusters. This guide shows how Neo4j monitoring works with KubeDB and how to configure `Neo4j` CR to enable monitoring.

## Overview

KubeDB uses an exporter sidecar to expose Prometheus metrics from Neo4j pods. The following diagram shows the logical monitoring flow for Neo4j with KubeDB.

<p align="center">
  <img alt="Database Monitoring Flow" src="/docs/images/concepts/monitoring/database-monitoring-overview.svg">
</p>

When a user creates a `Neo4j` CR with `spec.monitor` configured, KubeDB provisions the database and creates a dedicated stats service named `{neo4j-name}-stats` for monitoring. Prometheus scrapes metrics from this stats service.

## Configure Monitoring

To enable monitoring for Neo4j, configure the `spec.monitor` section. KubeDB provides the following options:

| Field | Type | Uses |
| --- | --- | --- |
| `spec.monitor.agent` | `Required` | Type of monitoring agent. Supported values: `prometheus.io/builtin` or `prometheus.io/operator`. |
| `spec.monitor.prometheus.serviceMonitor.labels` | `Optional` | Labels for `ServiceMonitor` object. |
| `spec.monitor.prometheus.serviceMonitor.interval` | `Optional` | Metrics scraping interval. |

## Sample Configuration

A sample YAML for a Neo4j cluster with monitoring enabled using Prometheus operator is shown below.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: sample-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  deletionPolicy: WipeOut
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here, `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create monitoring resources for Prometheus operator. KubeDB creates a `ServiceMonitor` object with the configured labels, and the Prometheus server discovers and scrapes Neo4j metrics through `{neo4j-name}-stats` service.

## Next Steps

- Monitor Neo4j using [Builtin Prometheus](/docs/guides/neo4j/monitoring/using-builtin-prometheus.md).
- Monitor Neo4j using [Prometheus Operator](/docs/guides/neo4j/monitoring/using-prometheus-operator.md).
- Configure alerting and dashboards from scraped metrics in Prometheus/Grafana.
