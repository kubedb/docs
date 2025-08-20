---
title: Hazelcast
menu:
  docs_{{ .version }}:
    identifier: hz-readme-hazelcast
    name: Hazelcast
    parent: hz-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/hazelcast/
aliases:
  - /docs/{{ .version }}/guides/hazelcast/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

### Overview

Hazelcast is an open-source, Java-based, information retrieval library with support for limited relational, graph, statistical, data analysis or storage related use cases. Hazelcast is designed to drive powerful document retrieval or analytical applications involving unstructured data, semi-structured data or a mix of unstructured and structured data. Hazelcast is highly reliable, scalable and fault tolerant, providing distributed indexing, replication and load-balanced querying, automated failover and recovery, centralized configuration and more. Hazelcast powers the search and navigation features of many of the world's largest internet sites.

## Supported Hazelcast Features
| Features                                                                           | Availability |
|------------------------------------------------------------------------------------|:------------:|
| Clustering                                                                         |   &#10003;   |
| Customized Docker Image                                                            |   &#10003;   |
| Authentication & Autorization                                                      |   &#10003;   | 
| Reconfigurable Health Checker                                                      |   &#10003;   |
| Custom Configuration                                                               |   &#10003;   | 
| Grafana Dashboards                                                                 |   &#10003;   | 
| Externally manageable Auth Secret                                                  |   &#10003;   |
| Persistent Volume                                                                  |   &#10003;   |
| Monitoring with Prometheus & Grafana                                               |   &#10003;   |
| Builtin Prometheus Discovery                                                       |   &#10003;   | 
| Alert Dashboard                                                                    |   &#10003;   |
| Using Prometheus operator                                                          |   &#10003;   |

## Life Cycle of a Hazelcast Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/hazelcast/quickstart/overview/images/Lifecycle-of-a-hazelcast-instance.png">
</p>

## User Guide

- [Quickstart Hazelcast](/docs/guides/hazelcast/quickstart/overview/index.md) with KubeDB Operator.
- Detail Concept of [Hazelcast Object](/docs/guides/hazelcast/concepts/hazelcast.md).


## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).