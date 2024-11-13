---
title: Solr Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: sl-scaling-vertical-overview
    name: Overview
    parent: sl-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `Solr`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `Solr`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of Solr" src="/docs/images/day-2-operation/solr/vertical-scaling.svg">
<figcaption align="center">Fig: Vertical scaling process of Solr</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `Solr` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Solr` CR.

3. When the operator finds a `Solr` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `Solr` cluster, the user creates a `SolrOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `SolrOpsRequest` CR.

6. When it finds a `SolrOpsRequest` CR, it halts the `Solr` object which is referred from the `SolrOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Solr` object during the vertical scaling process.

7. Then the `KubeDB` Ops-manager operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Ops-manager operator updates the `Solr` object to reflect the updated state.

9. After the successful update  of the `Solr` resources, the `KubeDB` Ops-manager operator resumes the `Solr` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of Solr database using `SolrOpsRequest` CRD.