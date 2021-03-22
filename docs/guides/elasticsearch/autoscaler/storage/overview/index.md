---
title: Elasticsearch Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: es-storage-auto-scaling-overview
    name: Overview
    parent: es-storage-auto-scaling
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Elasticsearch Storange Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the Elasticsearch storage using `elasticsearchautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
  - [ElasticsearchAutoscaler](/docs/guides/elasticsearch/concepts/autoscaler/index.md)
  - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

## How Storage Autoscaling Works

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Elasticsearch` Custom Resource (CR).

2. `KubeDB` Community operator watches the `Elasticsearch` CR.

3. When the operator finds a `Elasticsearch` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Each StatefulSet creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, to set up storage autoscaling of the various nodes (ie. master, data, ingest, etc.) of the `Elasticsearch` cluster the user creates a `ElasticsearchAutoscaler` CRO with the desired configuration.

6. `KubeDB` Autoscaler operator watches the `ElasticsearchAutoscaler` CRO.

7. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.

8. If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `ElasticsearchOpsRequest` to expand the storage of the database.
   
9. `KubeDB` Enterprise operator watches the `ElasticsearchOpsRequest` CRO.

10. Then the `KubeDB` Enterprise operator will expand the storage of the database component as specified on the `ElasticsearchOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various Elasticsearch database components using `ElasticsearchAutoscaler` CRD.
