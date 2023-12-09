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

# Elasticsearch Compute Resource Autoscaling

This guide will give an overview on how the KubeDB Autoscaler operator autoscales the database compute resources i.e. `cpu` and `memory` using `elasticsearchautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
  - [ElasticsearchAutoscaler](/docs/guides/elasticsearch/concepts/autoscaler/index.md)
  - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

## How Compute Autoscaling Works

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Elasticsearch` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Elasticsearch` CRO.

3. When the operator finds a `Elasticsearch` CRO, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the various components of the `Elasticsearch` database the user creates a `ElasticsearchAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `ElasticsearchAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `ElasticsearchAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `ElasticsearchOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `ElasticsearchOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `ElasticsearchOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of various Elasticsearch database components using `ElasticsearchAutoscaler` CRD.

