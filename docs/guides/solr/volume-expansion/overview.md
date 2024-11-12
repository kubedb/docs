---
title: Solr Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: sl-volume-expansion-overview
    name: Overview
    parent: sl-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Volume Expansion

This guide will give an overview on how KubeDB Ops-manager operator expand the volume of various component of `Solr` like:. (Combined and Topology).

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops-manager operator expand the volumes of `Solr` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of Solr" src="/docs/images/day-2-operation/Solr/kf-volume-expansion.svg">
<figcaption align="center">Fig: Volume Expansion process of Solr</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `Solr` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Solr` CR.

3. When the operator finds a `Solr` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the petset configuration. This Persistent Volume will be expanded by the `KubeDB` Ops-manager operator.

5. Then, in order to expand the volume of the various components (ie. Combined, Broker, Controller) of the `Solr`, the user creates a `SolrOpsRequest` CR with desired information.

6. `KubeDB` Ops-manager operator watches the `SolrOpsRequest` CR.

7. When it finds a `SolrOpsRequest` CR, it halts the `Solr` object which is referred from the `SolrOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Solr` object during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator will expand the persistent volume to reach the expected size defined in the `SolrOpsRequest` CR.

9. After the successful Volume Expansion of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the new volume size in the `Solr` object to reflect the updated state.

10. After the successful Volume Expansion of the `Solr` components, the `KubeDB` Ops-manager operator resumes the `Solr` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on Volume Expansion of various Solr database components using `SolrOpsRequest` CRD.
