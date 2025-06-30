---
title: Ignite
menu:
  docs_{{ .version }}:
    identifier: ig-readme-ignite
    name: Ignite
    parent: ig-ignite-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/ignite/
aliases:
  - /docs/{{ .version }}/guides/ignite/README/
---
> New to KubeDB? Please start [here](/docs/README.md).
## Supported Ignite Features

| Features                               | Availability |
| ------------------------------------   | :----------: |
| Clustering                             |   &#10003;   |
| Persistent Volume                      |   &#10003;   |
| Initialize using Script                |   &#10003;   |
| Multiple Ignite Versions               |   &#10003;   |
| Custom Configuration                   |   &#10003;   |
| Externally manageable Auth Secret	     |   &#10003;   |
| Reconfigurable Health Checker		       |   &#10003;   |
| Using Custom docker image              |   &#10003;   |
| Builtin Prometheus Discovery           |   &#10003;   |
| Using Prometheus operator              |   &#10003;   |
| Grafana Dashboard                      |   &#10003;   |
| Alert Dashboard	                       |   &#10003;   |



## Life Cycle of a Ignite Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/ignite/ignite-lifecycle.png">
</p>

## User Guide
- [Quickstart Ignite](/docs/guides/ignite/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Ignite server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ignite/monitoring/using-prometheus-operator.md).
- Monitor your Ignite server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ignite/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/ignite/private-registry/using-private-registry.md) to deploy Ignite with KubeDB.
- Detail concepts of [Ignite object](/docs/guides/ignite/concepts/ignite.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).