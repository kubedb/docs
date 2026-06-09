---
title: Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-readme
    name: Qdrant
    parent: qdrant-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/qdrant/
aliases:
  - /docs/{{ .version }}/guides/qdrant/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

Qdrant is a high-performance open-source vector database designed for similarity search and AI-powered applications. KubeDB supports provisioning and management of Qdrant clusters directly inside Kubernetes, enabling scalable and production-ready vector search infrastructure with minimal operational overhead. Deploy Qdrant in distributed mode to achieve horizontal scalability, replication, and high availability for large-scale embedding workloads. KubeDB automates cluster lifecycle management tasks such as deployment, scaling, monitoring, backups, and version upgrades, simplifying operations for machine learning and semantic search applications. With seamless Kubernetes integration, users can efficiently run and manage resilient Qdrant deployments for modern AI and retrieval-augmented generation (RAG) workloads.

## Supported Qdrant Features

| Features                      | Availability |
|-------------------------------|:------------:|
| Standalone Provisioning       |   &#10003;   |
| Distributed Provisioning      |   &#10003;   |
| Custom Configuration          |   &#10003;   |
| TLS/SSL Encryption            |   &#10003;   |
| Backup & Recovery             |   &#10003;   |
| Monitoring (Prometheus)       |   &#10003;   |
| Horizontal & Vertical Scaling |   &#10003;   |
| Volume Expansion              |   &#10003;   |
| Reconfigure                   |   &#10003;   |
| Update Version                |   &#10003;   |
| Restart                       |   &#10003;   |
| Rotate Authentication         |   &#10003;   |
| Autoscaling                   |   &#10003;   |
| StorageClass Migration        |   &#10003;   |

## Supported Qdrant Versions

KubeDB supports the following Qdrant versions.
- `1.15.4`
- `1.16.2`
- `1.17.0`

## Life Cycle of a Qdrant Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/qdrant/images/qdrant-lifecycle.png" style="padding: 20px;" >
</p>


## User Guide

- [Quickstart Qdrant](/docs/guides/qdrant/quickstart/quickstart.md) with KubeDB operator.
- Deploy [Distributed Qdrant](/docs/guides/qdrant/distributed-deployment/overview.md) cluster.
- Configure [Custom Configuration](/docs/guides/qdrant/configuration/using-config-file.md) for Qdrant.
- Configure [TLS/SSL](/docs/guides/qdrant/tls/overview.md) for Qdrant.
- [Backup and Restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant databases using KubeStash.
- [Monitor Qdrant](/docs/guides/qdrant/monitoring/overview.md) with KubeDB.
- [Horizontal Scaling](/docs/guides/qdrant/scaling/horizontal-scaling/overview.md) of Qdrant.
- [Vertical Scaling](/docs/guides/qdrant/scaling/vertical-scaling/overview.md) of Qdrant.
- [Volume Expansion](/docs/guides/qdrant/volume-expansion/overview.md) of Qdrant.
- [Reconfigure](/docs/guides/qdrant/reconfigure/overview.md) Qdrant.
- [Reconfigure TLS](/docs/guides/qdrant/reconfigure-tls/overview.md) for Qdrant.
- [Update Version](/docs/guides/qdrant/update-version/overview.md) of Qdrant.
- [Restart](/docs/guides/qdrant/restart/restart.md) Qdrant.
- [Rotate Authentication](/docs/guides/qdrant/rotate-auth/overview.md) credentials for Qdrant.
- [Compute Autoscaling](/docs/guides/qdrant/autoscaler/compute/overview.md) for Qdrant.
- [Storage Autoscaling](/docs/guides/qdrant/autoscaler/storage/overview.md) for Qdrant.
- [StorageClass Migration](/docs/guides/qdrant/migration/storageMigration.md) for Qdrant.
- Detail concepts of [Qdrant Object](/docs/guides/qdrant/concepts/).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
