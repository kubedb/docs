---
title: Update Version Overview
menu:
  docs_{{ .version }}:
    identifier: sl-update-version-overview
    name: Overview
    parent: sl-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Update Version Overview

This guide will give you an overview on how KubeDB Ops-manager operator update the version of `Solr`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)

## How update version Process Works

The following diagram shows how KubeDB Ops-manager operator used to update the version of `Solr`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of Solr" src="/docs/images/day-2-operation/Solr/kf-update-version.svg">
<figcaption align="center">Fig: updating Process of Solr</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Solr` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Solr` CR.

3. When the operator finds a `Solr` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the version of the `Solr` database the user creates a `SolrOpsRequest` CR with the desired version.

5. `KubeDB` Ops-manager operator watches the `SolrOpsRequest` CR.

6. When it finds a `SolrOpsRequest` CR, it halts the `Solr` object which is referred from the `SolrOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Solr` object during the updating process.

7. By looking at the target version from `SolrOpsRequest` CR, `KubeDB` Ops-manager operator updates the images of all the `PetSets`.

8. After successfully updating the `PetSets` and their `Pods` images, the `KubeDB` Ops-manager operator updates the image of the `Solr` object to reflect the updated state of the database.

9. After successfully updating of `Solr` object, the `KubeDB` Ops-manager operator resumes the `Solr` object so that the `KubeDB` Provisioner  operator can resume its usual operations.

In the next doc, we are going to show a step by step guide on updating of a Solr database using updateVersion operation.