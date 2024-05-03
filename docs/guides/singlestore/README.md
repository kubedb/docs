---
title: SingleStore
menu:
  docs_{{ .version }}:
    identifier: guides-singlestore-readme
    name: SingleStore
    parent: guides-singlestore
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/singlestore/
aliases:
  - /docs/{{ .version }}/guides/singlestore/README/
---
> New to KubeDB? Please start [here](/docs/README.md).

# Overview 

SingleStore, a distributed SQL database for real-time analytics, transactional workloads, and operational applications. With its in-memory processing and scalable architecture, SingleStore enables organizations to achieve high-performance and low-latency data processing across diverse data sets, making it ideal for modern data-intensive applications and analytical workflows. 

## Supported SingleStore Features

| Features                                                | Availability |
|---------------------------------------------------------|:------------:|
| Clustering                                              |   &#10003;   |
| Authentication & Authorization                          |   &#10003;   |
| Initialize using Script (\*.sql, \*sql.gz and/or \*.sh) |   &#10003;   |
| Custom Configuration                                    |   &#10003;   |
| TLS                                                     |   &#10003;   |
| Monitoring with Prometheus & Grafana                    |   &#10003;   |
| Builtin Prometheus Discovery                            |   &#10003;   |
| Using Prometheus operator                               |   &#10003;   |
| Externally manageable Auth Secret                       |   &#10003;   |
| Reconfigurable Health Checker                           |   &#10003;   |
| Persistent volume                                       |   &#10003;   | 
| SingleStore Studio (UI)                                 |   &#10003;   |


## Supported SingleStore Versions

KubeDB supports the following SingleSore Versions.
- `8.1.32`
- `8.5.7`

## Life Cycle of a SingleStore Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/singlestore/images/singlestore-lifecycle.png" >
</p>

## User Guide

- [Quickstart SingleStore](/docs/guides/singlestore/quickstart/quickstart.md) with KubeDB Operator.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).