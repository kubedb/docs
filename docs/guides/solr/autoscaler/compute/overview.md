---
title: Solr Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: sl-computer-autoscaling-overview
    name: Overview
    parent: sl-compute-autoscaling-solr
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Compute Resource Autoscaling

This guide will give an overview on how the KubeDB Autoscaler operator autoscales the database compute resources i.e. `cpu` and `memory` using `Solrautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrAutoscaler](/docs/guides/solr/concepts/autoscaler.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)

## How Compute Autoscaling Works

The Auto Scaling process consists of the following steps:

<p align="center">
  <img alt="Compute Autoscaling Flow"  src="/docs/images/day-2-operation/solr/compute-autoscaling.svg">
</p>

1. At first, a user creates a `Solr` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Solr` CRO.

3. When the operator finds a `Solr` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the various components of the `Solr` database the user creates a `SolrAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `SolrAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `SolrAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `SolrOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `SolrOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `SolrOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of various Solr database components using `SolrAutoscaler` CRD.

