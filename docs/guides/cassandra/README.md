---
title: Cassandra
menu:
  docs_{{ .version }}:
    identifier: cas-readme-cassandra
    name: Cassandra
    parent: cas-cassandra-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/cassandra/
aliases:
  - /docs/{{ .version }}/guides/cassandra/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview 

Apache Cassandra is a highly scalable, distributed NoSQL database designed to handle large amounts of data across many servers, offering high availability and no single point of failure. It's known for its ability to handle massive data loads with high performance, making it suitable for applications like social media, financial services, and IoT platforms.
## Supported Cassandra Features

| Features                                                      | Availability |
|---------------------------------------------------------------|:------------:|
| Clustering                                                    |   &#10003;   |
| Custom Configuration                                          |   &#10003;   |
| Backup & Recovery                                             |   &#10003;   |
| Monitoring (Prometheus & Grafana Dashboards)                  |   &#10003;   |
| TLS/SSL Encryption (Add, Remove, Rotate, Update)              |   &#10003;   |
| Auth Secret Management & Rotation                             |   &#10003;   |
| Persistent Volume                                             |   &#10003;   |
| Automated Version Update                                      |   &#10003;   |
| Horizontal & Vertical Scaling                                 |   &#10003;   |
| Automated Volume Expansion                                    |   &#10003;   |
| Autoscaling (Compute & Storage)                               |   &#10003;   |
| Reconfigure                                                   |   &#10003;   |
| Restart                                                       |   &#10003;   |

## Supported Cassandra Versions

KubeDB supports the following Cassandra Versions.
- `4.1.8`
- `5.0.3`

## Life Cycle of a Cassandra Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p text-align="center">
    <img alt="lifecycle"  src="/docs/images/cassandra/lifecycle.png" >
</p>

## User Guide

- [Quickstart Cassandra](/docs/guides/cassandra/quickstart/guide/quickstart.md) with KubeDB Operator.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).