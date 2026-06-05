---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-rotate-auth-overview
    name: Overview
    parent: qdrant-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Qdrant

This guide will give an overview on how KubeDB Ops-manager operator Rotate Authentication configuration.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How Rotate Qdrant Authentication Configuration Process Works

The authentication rotation process for Qdrant using KubeDB involves the following steps:

1. A user first creates a `Qdrant` Custom Resource Object (CRO).

2. The `KubeDB Provisioner operator` continuously watches for `Qdrant` CROs.

3. When the operator detects a `Qdrant` CR, it provisions the required `PetSets`, along with related resources such as secrets, services, and other dependencies.

4. To initiate authentication rotation, the user creates a `QdrantOpsRequest` CR with the desired configuration.

5. The `KubeDB Ops-manager` operator watches for `QdrantOpsRequest` CRs.

6. Upon detecting a `QdrantOpsRequest`, the operator pauses the referenced `Qdrant` object, ensuring that the Provisioner operator does not perform any operations during the authentication rotation process.

7. The `Ops-manager` operator then updates the necessary configuration (such as credentials) based on the provided `QdrantOpsRequest` specification.

8. After applying the updated configuration, the operator restarts all `Qdrant` Pods so they come up with the new authentication environment variables and settings.

9. Once the authentication rotation is completed successfully, the operator resumes the `Qdrant` object, allowing the Provisioner operator to continue its usual operations.

In the next section, we will walk you through a step-by-step guide to rotating Qdrant authentication using the `QdrantOpsRequest` CRD.
