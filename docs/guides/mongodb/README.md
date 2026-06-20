---
title: MongoDB
menu:
  docs_{{ .version }}:
    identifier: mg-readme-mongodb
    name: MongoDB
    parent: mg-mongodb-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/mongodb/
aliases:
  - /docs/{{ .version }}/guides/mongodb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported MongoDB Features


| Features                                                                           | Availability |
|------------------------------------------------------------------------------------|:------------:|
| Clustering (Replication & Sharding)                                                |   &#10003;   |
| Arbiter & Hidden Node Support                                                      |   &#10003;   |
| Custom Configuration                                                               |   &#10003;   |
| Using Custom Docker Image                                                          |   &#10003;   |
| Initialization (Script, Git Repository & Snapshot)                                 |   &#10003;   |
| Authentication & Authorization                                                     |   &#10003;   |
| Rotate Authentication Credentials                                                  |   &#10003;   |
| Persistent Volume                                                                  |   &#10003;   |
| Backup & Recovery (Instant & Scheduled)                                            |   &#10003;   |
| Continuous Archiving & Point-in-time Recovery                                      |   &#10003;   |
| Monitoring (Prometheus)                                                            |   &#10003;   |
| Automated Version Update                                                           |   &#10003;   |
| Horizontal & Vertical Scaling                                                      |   &#10003;   |
| Autoscaling (Compute & Storage)                                                    |   &#10003;   |
| Automated Volume Expansion                                                         |   &#10003;   |
| Reconfiguration                                                                    |   &#10003;   |
| TLS/SSL Encryption ( [Cert Manager](https://cert-manager.io/docs/) )               |   &#10003;   |
| Automated Reprovision                                                              |   &#10003;   |
| Restart                                                                            |   &#10003;   |
| Custom RBAC                                                                        |   &#10003;   |
| Schema Manager                                                                     |   &#10003;   |
| Encryption at Rest with Vault KMIP                                                 |   &#10003;   |
| External Connections (MongoDB Horizon)                                             |   &#10003;   |
| GitOps                                                                             |   &#10003;   |
| Failure & Disaster Recovery                                                        |   &#10003;   |


## Life Cycle of a MongoDB Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/mongodb/quick-start.png">
</p>

## User Guide

- [Quickstart MongoDB](/docs/guides/mongodb/quickstart/quickstart.md) with KubeDB Operator.
- [MongoDB Replicaset](/docs/guides/mongodb/clustering/replicaset.md) with KubeDB Operator.
- [MongoDB Sharding](/docs/guides/mongodb/clustering/sharding.md) with KubeDB Operator.
- [Backup & Restore](/docs/guides/mongodb/backup/stash/overview/index.md) MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Start [MongoDB with Custom Config](/docs/guides/mongodb/configuration/using-config-file.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
