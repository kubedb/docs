---
title: ProxySQL Monitoring Overview
description: ProxySQL Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-monitoring-overview
    name: Overview
    parent: guides-proxysql-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring ProxySQL with KubeDB

KubeDB has native support for monitoring via [Prometheus](https://prometheus.io/). You can use builtin [Prometheus](https://github.com/prometheus/prometheus) scraper or [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) to monitor KubeDB managed databases. This tutorial will show you how database monitoring works with KubeDB and how to configure Database crd to enable monitoring.

## Overview

KubeDB uses builtin proxysql [exporter](https://proxysql.com/documentation/prometheus-exporter/) to export Prometheus metrics for KubeDB ProxySQL. Following diagram shows the logical flow of proxysql server monitoring with KubeDB.

<p align="center">
  <img alt="ProxySQL Monitoring Flow"  src="/docs/guides/proxysql/monitoring/overview/images/monitoring.png">
</p>

When a user creates a proxysql crd with `spec.monitor` section configured, KubeDB operator provisions the respective server and enables the `admin-restapi_enabled`. It also creates a dedicated stats service with name `{database-crd-name}-stats` for monitoring. Prometheus server can scrape metrics using this stats service.

## Configure Monitoring

In order to enable monitoring for a database, you have to configure `spec.monitor` section. KubeDB provides following options to configure `spec.monitor` section for proxysql:

|                Field                               |    Type    |                                                                                     Uses                                                       |
| -------------------------------------------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `spec.monitor.agent`                               | `Required` | Type of the monitoring agent that will be used to monitor this database. It can be `prometheus.io/builtin` or `prometheus.io/operator`. |
| `spec.monitor.prometheus.serviceMonitor.labels`    | `Optional` | Labels for `ServiceMonitor` crd.                                                                                                               |
| `spec.monitor.prometheus.serviceMonitor.interval`  | `Optional` | Interval at which metrics should be scraped.                                                                                                   |

## Sample Configuration

A sample YAML for Redis crd with `spec.monitor` section configured to enable monitoring with [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) is shown below.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 3
  mode: GroupReplication
  backend:
    name: mysql-server-test
  syncUsers: false
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here, we have specified that we are going to monitor this server using Prometheus operator through `spec.monitor.agent: prometheus.io/operator`. KubeDB will create a `ServiceMonitor` crd in `monitoring` namespace and this `ServiceMonitor` will have `k8s-app: prometheus` label.

## Next Steps

- Learn how to monitor Elasticsearch database with KubeDB using [builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md) and using [Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Learn how to monitor PostgreSQL database with KubeDB using [builtin-Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md) and using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Learn how to monitor MySQL database with KubeDB using [builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md) and using [Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Learn how to monitor MongoDB database with KubeDB using [builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md) and using [Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Learn how to monitor Redis server with KubeDB using [builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md) and using [Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Learn how to monitor Memcached server with KubeDB using [builtin-Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md) and using [Prometheus operator](/docs/guides/memcached/monitoring/using-prometheus-operator.md).
