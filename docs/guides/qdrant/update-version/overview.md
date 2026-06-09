---
title: Updating Qdrant Version Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-update-version-overview
    name: Overview
    parent: qdrant-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Qdrant Version Overview

This guide will give you an overview on how KubeDB Ops-manager operator update the version of `Qdrant` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How update version Process Works

The following diagram shows how KubeDB Ops-manager operator update the version of `Qdrant`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of Qdrant" src="/docs/guides/qdrant/images/qdrant-update-version.png">
<figcaption align="center">Fig: Updating Process of Qdrant</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Qdrant` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `Qdrant` CR.

3. When the operator finds a `Qdrant` CR, it creates `PetSets` and related necessary resources like secrets, services, etc.

4. Then, in order to update the version of the `Qdrant` the user creates a `QdrantOpsRequest` CR with the desired version.

5. `KubeDB` Ops-manager operator watches the `QdrantOpsRequest` CR.

6. When it finds a `QdrantOpsRequest` CR, it halts the `Qdrant` object which is referred from the `QdrantOpsRequest`. So, the `KubeDB` Provisioner operator doesn't perform any operations on the `Qdrant` object during the updating process.

7. By looking at the target version from `QdrantOpsRequest` CR, `KubeDB` Ops-manager operator updates the images of the `PetSet`.

8. After successfully updating the `PetSet` and their `Pods` images, the `KubeDB` Ops-manager operator updates the image of the `Qdrant` object to reflect the updated state of the database.

9. After successfully updating of `Qdrant` object, the `KubeDB` Ops-manager operator resumes the `Qdrant` object so that the `KubeDB` Provisioner operator can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating of a Qdrant using UpdateVersion operation.
