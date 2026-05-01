---
title: Updating Qdrant Version
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

# Updating Qdrant Version

This guide will give you an overview of how KubeDB Ops-manager updates the version of a `Qdrant` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How the Update Process Works

The updating process consists of the following steps:

1. At first, a user creates a `Qdrant` CR.

2. `KubeDB-Provisioner` operator watches for the `Qdrant` CR.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the version of the `Qdrant` database, the user creates a `QdrantOpsRequest` CR with the desired target version.

5. `KubeDB-ops-manager` operator watches for `QdrantOpsRequest`.

6. When it finds one, it pauses the `Qdrant` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Qdrant` during the updating process.

7. By looking at the target version from the `QdrantOpsRequest` CR, the `KubeDB-ops-manager` operator updates the images of the `StatefulSet` for the new version.

8. After successful update of the `StatefulSet` and its Pod images, the `KubeDB-ops-manager` updates the image of the `Qdrant` object to reflect the updated cluster state.

9. After successful update of the `Qdrant` object, the `KubeDB` Ops-manager resumes the `Qdrant` object so that the `KubeDB-Provisioner` can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating a Qdrant database using the `UpdateVersion` operation.
