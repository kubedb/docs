---
title: Oracle Monitoring Overview
description: Oracle Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-monitoring-overview
    name: Overview
    parent: guides-oracle-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Oracle with KubeDB

KubeDB has native support for monitoring via [Prometheus](https://prometheus.io/). You can use the [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) to monitor KubeDB managed Oracle databases. This tutorial will show you how database monitoring works with KubeDB and how to configure the Oracle CRD to enable monitoring.

## Overview

KubeDB collects Oracle metrics using the free, public **Oracle AI Database Metrics Exporter**, which gathers standard Oracle database metrics and supports custom metrics collection. The metrics can be visualized in Grafana through flexible dashboards, enabling users to monitor database health and performance.

When a user creates an Oracle CRD with the `spec.monitor` section configured, the KubeDB operator provisions the database and runs the metrics exporter alongside the database pod. It also creates a dedicated stats service with the name `{oracle-crd-name}-stats` for monitoring. A Prometheus server can scrape metrics using this stats service.

## Configure Monitoring

In order to enable monitoring for an Oracle database, you have to configure the `spec.monitor` section. KubeDB provides the following options to configure the `spec.monitor` section:

|                Field                               |    Type    |                                                       Uses                                                       |
| -------------------------------------------------- | ---------- | --------------------------------------------------------------------------------------------------------------- |
| `spec.monitor.agent`                               | `Required` | Type of the monitoring agent used to monitor this database. Use `prometheus.io/operator`.                       |
| `spec.monitor.prometheus.exporter.port`            | `Optional` | Port number where the exporter serves metrics (defaults to `9161`).                                             |
| `spec.monitor.prometheus.exporter.resources`       | `Optional` | Resources required by the exporter container.                                                                   |
| `spec.monitor.prometheus.exporter.securityContext` | `Optional` | Security options the exporter should run with.                                                                  |
| `spec.monitor.prometheus.serviceMonitor.labels`    | `Optional` | Labels for the `ServiceMonitor` CRD.                                                                            |
| `spec.monitor.prometheus.serviceMonitor.interval`  | `Optional` | Interval at which metrics should be scraped.                                                                    |

## Sample Configuration

A sample YAML for an Oracle CRD with the `spec.monitor` section configured to enable monitoring with the [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) is shown below.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: standalone-monitoring
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9161
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 256Mi
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
```

Here, we have specified that we are going to monitor this server using the Prometheus operator through `spec.monitor.agent: prometheus.io/operator`. KubeDB will create a `ServiceMonitor` CRD with the `release: prometheus` label so that the Prometheus server discovers and scrapes it.

## Next Steps

- Learn how to monitor an Oracle database with KubeDB using the [Prometheus operator](/docs/guides/oracle/monitoring/using-prometheus-operator.md).
- Detail concepts of the [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

> ## ⚠️ Legal Notice
>
> Oracle® and Oracle Database® are registered trademarks of Oracle Corporation.
> KubeDB is not affiliated with, endorsed by, or sponsored by Oracle Corporation.
>
> KubeDB provides only orchestration and management tooling for Kubernetes.
> It does not distribute, bundle, ship, or include any Oracle Database software or binaries.
>
> Users must provide their own Oracle container images and hold valid Oracle licenses.
> Users are solely responsible for compliance with Oracle’s licensing terms, including all rules regarding containers, Docker, and Kubernetes environments.
>
> KubeDB makes no representations or warranties regarding Oracle licensing compliance.
