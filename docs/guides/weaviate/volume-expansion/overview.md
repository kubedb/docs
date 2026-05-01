---
title: Expanding Weaviate Storage
menu:
  docs_{{ .version }}:
    identifier: weaviate-volume-expansion-overview
    name: Overview
    parent: weaviate-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Volume Expansion

This guide will give an overview of how KubeDB Ops-manager expands the volume of a `Weaviate` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)

## How Volume Expansion Works

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `Weaviate` Custom Resource (CR).

2. `KubeDB-Provisioner` operator watches the `Weaviate` CR.

3. When the operator finds a `Weaviate` CR, it creates a `StatefulSet` and related necessary stuff like pods, PVCs, secrets, services, etc.

4. Each StatefulSet creates a Persistent Volume according to the volume claim template. This Persistent Volume will be expanded by the `KubeDB` Ops-manager operator.

5. Then, in order to expand the volume of the `Weaviate` database, the user creates a `WeaviateOpsRequest` CR with the desired new storage size.

6. `KubeDB` Ops-manager operator watches the `WeaviateOpsRequest` CR.

7. When it finds a `WeaviateOpsRequest` CR, it pauses the `Weaviate` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Weaviate` during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator expands the persistent volumes to the expected size defined in the `WeaviateOpsRequest` CR.

9. After the successful expansion of the volumes, the `KubeDB` Ops-manager updates the new volume size in the `Weaviate` object to reflect the updated state.

10. After the successful Volume Expansion, the `KubeDB` Ops-manager resumes the `Weaviate` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on Volume Expansion of a Weaviate database using `WeaviateOpsRequest` CRD.
