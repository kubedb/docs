---
title: Solr Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: sl-storage-autoscaling-overview
    name: Overview
    parent: sl-storage-autoscaling-solr
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Storage SolrAutoscaler

This guide will give an overview on how KubeDB Autoscaler operator autoscales the Solr storage using `Solrautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrAutoscaler](/docs/guides/solr/concepts/autoscaler.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)

## How Storage Autoscaling Works

The Auto Scaling process consists of the following steps:

<p align="center">
  <img alt="Compute Autoscaling Flow"  src="/docs/images/day-2-operation/solr/storage-autoscaling.svg">
</p>

1. At first, a user creates a `Solr` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Solr` CR.

3. When the operator finds a `Solr` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

- Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the petset configuration.

4. Then, in order to set up storage autoscaling of the various components of the `Solr` database the user creates a `SolrAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `SolrAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `SolrOpsRequest` to expand the storage of the database.

7. `KubeDB` Ops-manager operator watches the `SolrOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the database component as specified on the `SolrOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling storage of various Solr database components using `SolrAutoscaler` CRD.
