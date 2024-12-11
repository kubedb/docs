---
title: Reconfiguring Solr
menu:
  docs_{{ .version }}:
    identifier: sl-reconfigure-overview
    name: Overview
    parent: sl-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Solr

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `Solr` components such as Combined, Broker, Controller, etc.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)

## How Reconfiguring Solr Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `Solr` components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of Solr" src="/docs/images/day-2-operation/solr/reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of Solr</figcaption>
</figure>

The Reconfiguring Solr process consists of the following steps:

1. At first, a user creates a `Solr` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Solr` CR.

3. When the operator finds a `Solr` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the various components (ie. Combined, Broker) of the `Solr`, the user creates a `SolrOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `SolrOpsRequest` CR.

6. When it finds a `SolrOpsRequest` CR, it halts the `Solr` object which is referred from the `SolrOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Solr` object during the reconfiguring process.

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `MogoDBOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `SolrOpsRequest` CR.

9. After the successful reconfiguring of the `Solr` components, the `KubeDB` Ops-manager operator resumes the `Solr` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring Solr components using `SolrOpsRequest` CRD.