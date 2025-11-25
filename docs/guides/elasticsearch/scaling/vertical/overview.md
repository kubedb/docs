---
title: Elasticsearch Vertical Scaling Overview
menu:
    docs_{{ .version }}
        identifier: es-vertical-scalling-overview
        name: Overview
        parent: es-vertical-scalling-elasticsearch
        weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `Elasticsearch`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
- [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
- [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `Elasticsearch`. Open the image in a new tab to see the enlarged version.

<figure align="center">
      <img alt="Vertical scaling process of Elasticsearch" src="/docs/images/elasticsearch/es-vertical-scaling.jpg">
    <figcaption align="center">Fig: Vertical scaling process of Elasticsearch</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `Elasticsearch` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Elasticsearch` CR.

3. When the operator finds a `Elasticsearch` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `Elasticsearch` cluster, the user creates a `ElasticsearchOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ElasticsearchOpsRequest` CR.

6. When it finds a `ElasticsearchOpsRequest` CR, it halts the `Elasticsearch` object which is referred from the `ElasticsearchOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Elasticsearch` object during the vertical scaling process.

7. Then the `KubeDB` Ops-manager operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Ops-manager operator updates the `Elasticsearch` object to reflect the updated state.

9. After the successful update  of the `Elasticsearch` resources, the `KubeDB` Ops-manager operator resumes the `Elasticsearch` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of Elasticsearch database using `ElasticsearchOpsRequest` CRD.