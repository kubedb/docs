---
title: Updating Weaviate Version
menu:
  docs_{{ .version }}:
    identifier: weaviate-update-version-overview
    name: Overview
    parent: weaviate-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Weaviate Version

This guide will give you an overview of how KubeDB Ops-manager updates the version of a `Weaviate` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)

## How the Update Process Works

The updating process consists of the following steps:

1. At first, a user creates a `Weaviate` CR.

2. `KubeDB-Provisioner` operator watches for the `Weaviate` CR.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the version of the `Weaviate` database, the user creates a `WeaviateOpsRequest` CR with the desired target version.

5. `KubeDB-ops-manager` operator watches for `WeaviateOpsRequest`.

6. When it finds one, it pauses the `Weaviate` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Weaviate` during the updating process.

7. By looking at the target version from the `WeaviateOpsRequest` CR, the `KubeDB-ops-manager` operator updates the images of the `StatefulSet` for the new version.

8. After successful update of the `StatefulSet` and its Pod images, the `KubeDB-ops-manager` updates the image of the `Weaviate` object to reflect the updated cluster state.

9. After successful update of the `Weaviate` object, the `KubeDB` Ops-manager resumes the `Weaviate` object so that the `KubeDB-Provisioner` can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating a Weaviate database using the `UpdateVersion` operation.
