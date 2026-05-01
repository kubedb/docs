---
title: Weaviate Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-vertical-scaling-overview
    name: Overview
    parent: weaviate-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Vertical Scaling

This guide will give an overview of how KubeDB Ops-manager updates the CPU and memory resources of `Weaviate` database nodes.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)

## How Vertical Scaling Works

The Vertical Scaling process consists of the following steps:

1. At first, a user creates a `Weaviate` CR.

2. `KubeDB-Provisioner` operator watches the `Weaviate` CR.

3. When the operator finds a `Weaviate` CR, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the CPU and memory resources of the `Weaviate` database nodes, the user creates a `WeaviateOpsRequest` CR with the desired resource specifications.

5. `KubeDB` Ops-manager operator watches the `WeaviateOpsRequest` CR.

6. When it finds a `WeaviateOpsRequest` CR, it pauses the `Weaviate` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Weaviate` during the scaling process.

7. Then the `KubeDB` Ops-manager operator updates the resources of the `StatefulSet` pods to the desired values defined in the `WeaviateOpsRequest` CR.

8. After the successful resource update of the pods, the `KubeDB` Ops-manager updates the resource specifications in the `Weaviate` object to reflect the updated state.

9. After the successful Vertical Scaling, the `KubeDB` Ops-manager resumes the `Weaviate` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on Vertical Scaling of a Weaviate database using `WeaviateOpsRequest` CRD.
