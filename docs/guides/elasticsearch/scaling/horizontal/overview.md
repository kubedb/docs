---
title: Elasticsearch Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: es-horizontal-scalling-overview
    name: Overview
    parent: es-horizontal-scalling-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Horizontal Scaling

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Elasticsearch` cluster replicas of various component such as Combined, Broker, Controller.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator scales up or down `Elasticsearch` database components. Open the image in a new tab to see the enlarged version.

[//]: # (<figure align="center">)

[//]: # (  <img alt="Horizontal scaling process of Elasticsearch" src="/docs/images/day-2-operation/Elasticsearch/kf-horizontal-scaling.svg">)

[//]: # (<figcaption align="center">Fig: Horizontal scaling process of Elasticsearch</figcaption>)

[//]: # (</figure>)

The Horizontal scaling process consists of the following steps:

1. At first, a user creates a `Elasticsearch` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Elasticsearch` CR.

3. When the operator finds a `Elasticsearch` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to scale the various components of the `Elasticsearch` cluster, the user creates a `ElasticsearchOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ElasticsearchOpsRequest` CR.

6. When it finds a `ElasticsearchOpsRequest` CR, it halts the `Elasticsearch` object which is referred from the `ElasticsearchOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Elasticsearch` object during the horizontal scaling process.

7. Then the `KubeDB` Ops-manager operator will scale the related PetSet Pods to reach the expected number of replicas defined in the `ElasticsearchOpsRequest` CR.

8. After the successfully scaling the replicas of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the number of replicas in the `Elasticsearch` object to reflect the updated state.

9. After the successful scaling of the `Elasticsearch` replicas, the `KubeDB` Ops-manager operator resumes the `Elasticsearch` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on horizontal scaling of Elasticsearch cluster using `ElasticsearchOpsRequest` CRD.