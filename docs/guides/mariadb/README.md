---
title: MariaDB
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-overview
    name: MariaDB
    parent: guides-mariadb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/mariadb/
aliases:
  - /docs/{{ .version }}/guides/mariadb/README/
---


> New to KubeDB? Please start [here](/docs/README.md).

## Supported MariaDB Features

| Features                                                         | Availability |
|------------------------------------------------------------------| :----------: |
| Clustering                                                       |   &#10003;   |
| Persistent Volume                                                |   &#10003;   |
| Backup & Recovery (Instant & Scheduled)                          |   &#10003;   |
| Continuous Archiving and Point-in-time Recovery                  |   &#10003;   |
| Initialization (Script & Git Repository)                         |   &#10003;   |
| Custom Configuration                                             |   &#10003;   |
| Using Custom Docker Image                                        |   &#10003;   |
| Monitoring (Prometheus)                                          |   &#10003;   |
| TLS/SSL Encryption                                               |   &#10003;   |
| Horizontal & Vertical Scaling                                    |   &#10003;   |
| Autoscaling (Compute & Storage)                                  |   &#10003;   |
| Reconfigure                                                      |   &#10003;   |
| Update Version                                                   |   &#10003;   |
| Volume Expansion                                                 |   &#10003;   |
| Restart                                                          |   &#10003;   |
| Rotate Authentication                                            |   &#10003;   |
| Failover and Disaster Recovery                                   |   &#10003;   |
| Distributed (Multi-cluster)                                      |   &#10003;   |
| GitOps                                                           |   &#10003;   |
| Custom RBAC                                                      |   &#10003;   |

## Life Cycle of a MariaDB Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/mariadb/images/mariadb-lifecycle.png" >
</p>

## User Guide

- [Quickstart MariaDB](/docs/guides/mariadb/quickstart/overview) with KubeDB Operator.
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb).
- Detail concepts of [MariaDBVersion object](/docs/guides/mariadb/concepts/mariadb-version).
- Create [MariaDB Cluster](/docs/guides/mariadb/clustering/galera-cluster).
- Create [MariaDB with Custom Configuration](/docs/guides/mariadb/configuration/using-config-file).
- Use [Custom RBAC](/docs/guides/mariadb/custom-rbac/using-custom-rbac).
- Use [private Docker registry](/docs/guides/mariadb/private-registry/quickstart) to deploy MySQL with KubeDB.
- Initialize [MariaDB with Script](/docs/guides/mariadb/initialization/using-script).
- Backup and Restore [MariaDB](/docs/guides/mariadb/backup/stash/overview).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
