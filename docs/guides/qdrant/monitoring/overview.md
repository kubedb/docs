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

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Monitoring Overview

This guide will give an overview of how KubeDB supports monitoring for `Qdrant` databases.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)

## How KubeDB Monitoring Works

KubeDB uses Prometheus to monitor `Qdrant` databases. KubeDB operator watches the `Qdrant` CR and sets up monitoring as follows:

1. When a `Qdrant` database is deployed with `spec.monitor` configured, KubeDB creates a dedicated `stats` service (or uses the existing service) with the appropriate annotations for Prometheus scraping.

2. KubeDB supports two monitoring approaches:
   - **Builtin Prometheus** — uses Prometheus' built-in auto-discovery mechanism (`prometheus.io/scrape` annotations on the stats service).
   - **Prometheus Operator** — creates a `ServiceMonitor` CR that is picked up by the Prometheus Operator.

3. The Qdrant stats service exposes Prometheus-compatible metrics at the `/metrics` endpoint, including metrics about collections, vectors, memory usage, and gRPC/REST request performance.

4. Prometheus scrapes the metrics from the stats service and makes them available for alerting and dashboards.

In the next docs, we are going to show step-by-step guides on monitoring a Qdrant database using Builtin Prometheus and Prometheus Operator.
