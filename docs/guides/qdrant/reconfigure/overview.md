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

This guide will give an overview of how KubeDB Ops-manager reconfigures a `Qdrant` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How Reconfiguration Works

The Reconfiguration process consists of the following steps:

1. At first, a user creates a `Qdrant` CR.

2. `KubeDB-Provisioner` operator watches the `Qdrant` CR.

3. When the operator finds a `Qdrant` CR, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `Qdrant` database, the user creates a `QdrantOpsRequest` CR with the new configuration. The user can provide the new configuration either via a new config secret, via `applyConfig`, or by removing the custom configuration (reverting to defaults).

5. `KubeDB` Ops-manager operator watches the `QdrantOpsRequest` CR.

6. When it finds a `QdrantOpsRequest` CR, it pauses the `Qdrant` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Qdrant` during the reconfiguration process.

7. Then the `KubeDB` Ops-manager operator updates the configuration secret and restarts the pods in a rolling fashion to apply the new configuration.

8. After the successful configuration update, the `KubeDB` Ops-manager updates the `Qdrant` object to reflect the updated configuration state.

9. After the successful reconfiguration, the `KubeDB` Ops-manager resumes the `Qdrant` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on reconfiguring a Qdrant database using `QdrantOpsRequest` CRD.
