---
title: Solr
menu:
  docs_{{ .version }}:
    identifier: sl-readme-solr
    name: Solr
    parent: sl-solr-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/solr/
aliases:
  - /docs/{{ .version }}/guides/solr/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

### Overview

Solr is an open-source, Java-based, information retrieval library with support for limited relational, graph, statistical, data analysis or storage related use cases. Solr is designed to drive powerful document retrieval or analytical applications involving unstructured data, semi-structured data or a mix of unstructured and structured data. Solr is highly reliable, scalable and fault tolerant, providing distributed indexing, replication and load-balanced querying, automated failover and recovery, centralized configuration and more. Solr powers the search and navigation features of many of the world's largest internet sites.

## Supported Solr Features
| Features                             | Availability |
|--------------------------------------|:------------:|
| Clustering                           |   &#10003;   |
| Customized Docker Image              |   &#10003;   |
| Authentication & Autorization        |   &#10003;   | 
| Reconfigurable Health Checker        |   &#10003;   |
| Custom Configuration                 |   &#10003;   | 
| Grafana Dashboards                   |   &#10003;   | 
| Externally manageable Auth Secret    |   &#10003;   |
| Persistent Volume                    |   &#10003;   |
| Monitoring with Prometheus & Grafana |   &#10003;   |
| Builtin Prometheus Discovery         |   &#10003;   | 
| Alert Dashboard                      |   &#10003;   |
| Using Prometheus operator            |   &#10003;   |
| Dashboard ( Solr UI )                |   &#10003;   |

## Life Cycle of a Solr Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/solr/quickstart/overview/images/Lifecycle-of-a-solr-instance.png">
</p>

## User Guide

- [Quickstart Solr](/docs/guides/solr/quickstart/overview/index.md) with KubeDB Operator.
- Detail Concept of [Solr Object](/docs/guides/solr/concepts/solr.md).


## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).