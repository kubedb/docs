---
title: Druid
menu:
  docs_{{ .version }}:
    identifier: guides-druid-readme
    name: Druid
    parent: guides-druid
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/druid/
aliases:
  - /docs/{{ .version }}/guides/druid/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Druid Features


| Features                                                                           | Availability |
|------------------------------------------------------------------------------------|:-----:|
| Clustering                                                                         |   &#10003; |
| Druid Dependency Management (MySQL, PostgreSQL and ZooKeeper)                      |   &#10003; |
| Authentication & Authorization                                                     |   &#10003; |
| Custom Configuration                                                               |   &#10003; |
| Backup/Recovery: Instant, Scheduled ( [KubeStash](https://kubestash.com/))         |   &#10003; |
| Monitoring with Prometheus & Grafana                                               |   &#10003; |
| Builtin Prometheus Discovery                                                       |   &#10003; |
| Using Prometheus operator                                                          |   &#10003; |
| Externally manageable Auth Secret                                                  |   &#10003; |
| Reconfigurable Health Checker                                                      |   &#10003; |
| Persistent volume                                                                  |   &#10003; | 
| Dashboard ( Druid Web Console )                                                    |   &#10003; |
| Automated Version Update                                                           |  &#10003;  |
| Automatic Vertical Scaling                                                         |  &#10003;  |
| Automated Horizontal Scaling                                                       |  &#10003;  |
| Automated db-configure Reconfiguration                                             |  &#10003;  |
| TLS: Add, Remove, Update, Rotate ( [Cert Manager](https://cert-manager.io/docs/) ) |  &#10003;  |
| Automated Reprovision                                                              |  &#10003;  |
| Automated Volume Expansion                                                         |  &#10003;  |
| Autoscaling (vertically)                                                           |  &#10003;  |

## Supported Druid Versions

KubeDB supports The following Druid versions.
- `28.0.1`
- `30.0.1`

> The listed DruidVersions are tested and provided as a part of the installation process (ie. catalog chart), but you are open to create your own [DruidVersion](/docs/guides/druid/concepts/druidversion.md) object with your custom Druid image.

## Lifecycle of Druid Object

<!---
ref : https://cacoo.com/diagrams/bbB63L6KRIbPLl95/9A5B0
--->

<p align="center">
<img alt="lifecycle"  src="/docs/images/druid/Druid-CRD-Lifecycle.png">
</p>

## User Guide 
- [Quickstart Druid](/docs/guides/druid/quickstart/guide/index.md) with KubeDB Operator.
- [Druid Clustering](/docs/guides/druid/clustering/overview/index.md) with KubeDB Operator.
- [Backup & Restore](/docs/guides/druid/backup/overview/index.md) Druid databases using KubeStash.
- Start [Druid with Custom Config](/docs/guides/druid/configuration/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/druid/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Detail concepts of [DruidVersion object](/docs/guides/druid/concepts/druidversion.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).