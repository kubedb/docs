---
title: Updating Elasticsearch Overview
menu:
  docs_{{ .version }}:
    identifier: guides-Elasticsearch-updating-overview
    name: Overview
    parent:  es-updateversion-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Elasticsearch version

This guide will give you an overview of how KubeDB ops manager updates the version of `Elasticsearch` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

## How update Process Works

The following diagram shows how KubeDB KubeDB ops manager used to update the version of `Elasticsearch`. Open the image in a new tab to see the enlarged version.

[//]: # (<figure align="center">)

[//]: # (  <img alt="Elasticsearch update Flow" src="/docs/guides/Elasticsearch/update-version/overview/images/pg-updating.png">)

[//]: # (<figcaption align="center">Fig: updating Process of Elasticsearch</figcaption>)

[//]: # (</figure>)

The updating process consists of the following steps:

1. At first, a user creates a `Elasticsearch` cr.

2. `KubeDB-Provisioner` operator watches for the `Elasticsearch` cr.

3. When it finds one, it creates a `PetSet` and related necessary stuff like secret, service, etc.

4. Then, in order to update the version of the `Elasticsearch` database the user creates a `ElasticsearchOpsRequest` cr with the desired version.

5. `KubeDB-ops-manager` operator watches for `ElasticsearchOpsRequest`.

6. When it finds one, it Pauses the `Elasticsearch` object so that the `KubeDB-Provisioner` operator doesn't perform any operation on the `Elasticsearch` during the updating process.

7. By looking at the target version from `ElasticsearchOpsRequest` cr, In case of major update `KubeDB-ops-manager` does some pre-update steps as we need old bin and lib files to update from current to target Elasticsearch version.

8. Then By looking at the target version from `ElasticsearchOpsRequest` cr, `KubeDB-ops-manager` operator updates the images of the `PetSet` for updating versions.

9. After successful upgradation of the `PetSet` and its `Pod` images, the `KubeDB-ops-manager` updates the image of the `Elasticsearch` object to reflect the updated cluster state.

10. After successful upgradation of `Elasticsearch` object, the `KubeDB` ops manager resumes the `Elasticsearch` object so that the `KubeDB-provisioner` can resume its usual operations.

In the next doc, we are going to show a step by step guide on updating of a Elasticsearch database using update operation.