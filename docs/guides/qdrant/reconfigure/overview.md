---
title: Reconfiguring Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-reconfigure-overview
    name: Overview
    parent: qdrant-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Qdrant

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `Qdrant` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How Reconfiguring Qdrant Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `Qdrant` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of Qdrant" src="/docs/guides/qdrant/images/qdrant-reconfigure.png">
<figcaption align="center">Fig: Reconfiguring process of Qdrant</figcaption>
</figure>

The Reconfiguring Qdrant process consists of the following steps:

1. At first, a user creates a `Qdrant` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `Qdrant` CR.

3. When the operator finds a `Qdrant` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `Qdrant` database the user creates a `QdrantOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `QdrantOpsRequest` CR.

6. When it finds a `QdrantOpsRequest` CR, it halts the `Qdrant` object which is referred from the `QdrantOpsRequest`. So, the `KubeDB` Provisioner operator doesn't perform any operations on the `Qdrant` object during the reconfiguring process.

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `QdrantOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `QdrantOpsRequest` CR.

9. After the successful reconfiguring of the `Qdrant`, the `KubeDB` Ops-manager operator resumes the `Qdrant` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on reconfiguring Qdrant database using `QdrantOpsRequest` CR.
