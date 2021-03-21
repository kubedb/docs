---
title: Elasticsearch Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: es-auto-scaling-overview
    name: Overview
    parent: es-compute-auto-scaling
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Elasticsearch Compute Resource Autoscaling

This guide will give an overview on how the KubeDB Autoscaler operator autoscales the database compute resources i.e. `cpu` and `memory` using `elasticsearchautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
  - [ElasticsearchAutoscaler](/docs/guides/elasticsearch/concepts/autoscaler/index.md)
  - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

## How Compute Autoscaling Works

The Auto Scaling process consists of the following steps:

1. At first, a user creates an `Elasticsearch` Custom Resource Object (CRO).

2. `KubeDB` Community operator watches the `Elasticsearch` CRO.

3. When the operator finds an `Elasticsearch` CRO, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the various nodes (ie. master, data, ingest, etc.) of the `Elasticsearch` database the user creates an `ElasticsearchAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `ElasticsearchAutoscaler` CRO.

6. `KubeDB` Autoscaler operator creates required number of Vertical Pod Autoscaler [VPA](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#intro) for different components of the database, as specified in the `elasticsearchautoscaler` CRO.

7. Then `KubeDB` Autoscaler operator continuously watches the VPA objects for the recommendation.

8. If the VPA generated recommendation doesn't match the current resources of the database, then the `KubeDB` Autoscaler operator creates an `ElasticsearchOpsRequest` CRO to scale the database to match the recommendation provided by the VPA object.

9. `KubeDB` Enterprise operator watches the `ElasticsearchOpsRequest` CRO.

10. Then the `KubeDB` Enterprise operator will scale the database component vertically as specified on the `ElasticsearchOpsRequest` CRO.

In the next docs, we are going to show a detailed-guide on Autoscaling of various Elasticsearch database components using `ElasticsearchAutoscaler` CRD.
