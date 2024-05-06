---
title: Solr
menu:
  docs_{{ .version }}:
    identifier: sl-readme-solr
    name: Solr
    parent: sl-solr-guides
    weight: 8
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/solr/
aliases:
  - /docs/{{ .version }}/guides/solr/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Solr Features
| Features                           | Availability |
|------------------------------------|:------------:|
| Clustering                         |   &#10003;   |
| Authentication & Autorization      |   &#10003;   | 
| Custom Configuration               |   &#10003;   | 
| Grafana Dashboards                 |   &#10003;   | 
| Externally manageable Auth Secret  |   &#10003;   |
| Persistent Volume                  |   &#10003;   |
| Builtin Prometheus Discovery       |   &#10003;   | 
| Using Prometheus operator          |   &#10003;   |
| Solr Builtin UI                    |   &#10003;   |

## Life Cycle of a Solr Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/solr/quickstart/overview/images/Lifecycle-of-a-solr-instance.png">
</p>

## User Guide

- [Quickstart Solr](/docs/guides/solr/quickstart/overview/index.md) with KubeDB Operator.
- Detail Concept of [Solr Object](/docs/guides/solr/concepts/solr.md).


## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).