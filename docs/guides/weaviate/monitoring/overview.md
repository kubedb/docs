---
title: Weaviate Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-monitoring-overview
    name: Overview
    parent: weaviate-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Monitoring Overview

This guide will give an overview of how KubeDB supports monitoring for `Weaviate` databases.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)

## How KubeDB Monitoring Works

KubeDB uses Prometheus to monitor `Weaviate` databases. KubeDB operator watches the `Weaviate` CR and sets up monitoring as follows:

1. When a `Weaviate` database is deployed with `spec.monitor` configured, KubeDB creates a dedicated `stats` service (or uses the existing service) with the appropriate annotations for Prometheus scraping.

2. KubeDB supports two monitoring approaches:
   - **Builtin Prometheus** - uses Prometheus' built-in auto-discovery mechanism (`prometheus.io/scrape` annotations on the stats service).
   - **Prometheus Operator** - creates a `ServiceMonitor` CR that is picked up by the Prometheus Operator.

3. The Weaviate stats service exposes Prometheus-compatible metrics at the `/metrics` endpoint, including metrics about objects, memory usage, and gRPC/REST request performance.

4. Prometheus scrapes the metrics from the stats service and makes them available for alerting and dashboards.

In the next docs, we are going to show step-by-step guides on monitoring a Weaviate database using Builtin Prometheus and Prometheus Operator.
