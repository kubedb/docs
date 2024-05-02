---
title: Druid
menu:
  docs_{{ .version }}:
    identifier: dr-readme-druid
    name: Druid
    parent: dr-druid-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/druid/
aliases:
  - /docs/{{ .version }}/guides/druid/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Druid Features


| Features                             | Availability |
|--------------------------------------|:------------:|
| Clustering                           |   &#10003;   |
| Authentication                       |   &#10003;   |
| Custom Configuration                 |   &#10003;   |
| Monitoring with Prometheus & Grafana |   &#10003;   |
| Builtin Prometheus Discovery         |   &#10003;   |
| Using Prometheus operator            |   &#10003;   |
| Externally manageable Auth Secret    |   &#10003;   |
| Reconfigurable Health Checker        |   &#10003;   |
| Persistent volume                    |   &#10003;   | 
| Druid Web Console                    |   &#10003;   |

## Supported Druid Versions

KubeDB supports The following Druid versions.
- `25.0.0`
- `28.0.1`

> The listed DruidVersions are tested and provided as a part of the installation process (ie. catalog chart), but you are open to create your own [DruidVersion](/docs/guides/druid/concepts/catalog.md) object with your custom Druid image.

## Lifecycle of Druid Object

<!---
ref : https://cacoo.com/diagrams/bbB63L6KRIbPLl95/9A5B0
--->

<p align="center">
<img alt="lifecycle"  src="/docs/images/druid/Druid-CRD-Lifecycle.png">
</p>

## User Guide 
- [Quickstart Druid](/docs/guides/druid/quickstart/overview/index.md) with KubeDB Operator.

[//]: # (- Druid Clustering supported by KubeDB)

[//]: # (  - [Topology Clustering]&#40;/docs/guides/druid/clustering/topology-cluster/index.md&#41;)

[//]: # (- Use [kubedb cli]&#40;/docs/guides/druid/cli/cli.md&#41; to manage databases like kubectl for Kubernetes.)

[//]: # (- Detail concepts of [Druid object]&#40;/docs/guides/druid/concepts/druid.md&#41;.)

[//]: # (- Want to hack on KubeDB? Check our [contribution guidelines]&#40;/docs/CONTRIBUTING.md&#41;.)