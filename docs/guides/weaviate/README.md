---
title: Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-readme
    name: Weaviate
    parent: weaviate-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/weaviate/
aliases:
  - /docs/{{ .version }}/guides/weaviate/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

Weaviate is an open-source, AI-native vector database that stores both objects and their vector embeddings, enabling fast semantic search, hybrid search, and retrieval-augmented generation (RAG) at scale. KubeDB supports provisioning and management of Weaviate clusters directly inside Kubernetes, bringing production-grade vector search infrastructure to your cluster with minimal operational overhead. Deploy Weaviate with multiple replicas to achieve horizontal scalability, replication, and high availability for large embedding workloads. KubeDB automates cluster lifecycle management tasks such as deployment, scaling, custom configuration, TLS, authentication rotation, volume expansion, storage migration, and autoscaling, simplifying operations for machine learning and semantic search applications. With seamless Kubernetes integration, users can efficiently run and manage resilient Weaviate deployments for modern AI and retrieval-augmented generation (RAG) workloads.

## Supported Weaviate Features

| Features                        | Availability |
|---------------------------------|:------------:|
| Clustered (Multi-node) provisioning |   &#10003;   |
| Custom Configuration            |   &#10003;   |
| TLS                             |   &#10003;   |
| Authentication (API key)        |   &#10003;   |
| Reconfigure                     |   &#10003;   |
| Reconfigure TLS                 |   &#10003;   |
| Rotate Authentication           |   &#10003;   |
| Restart                         |   &#10003;   |
| Vertical Scaling                |   &#10003;   |
| Horizontal Scaling              |   &#10003;   |
| Volume Expansion                |   &#10003;   |
| Storage Migration               |   &#10003;   |
| Compute Autoscaler              |   &#10003;   |
| Storage Autoscaler              |   &#10003;   |

## Supported Weaviate Versions

KubeDB supports the following Weaviate versions.
- `1.33.1`

> The listed versions are the ones shipped with the `WeaviateVersion` catalog in your KubeDB installation. Run `kubectl get weaviateversions` to see the versions available in your cluster.

## User Guide

- [Quickstart Weaviate](/docs/guides/weaviate/quickstart/quickstart.md) with KubeDB operator.
- Use [custom configuration](/docs/guides/weaviate/configuration/using-config-file.md) for Weaviate.
- Secure your cluster with [TLS](/docs/guides/weaviate/tls/overview.md).
- Run day-2 operations with [Ops Requests](/docs/guides/weaviate/restart/restart.md) and [Autoscaler](/docs/guides/weaviate/autoscaler/compute/compute-autoscale.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
