---
title: Rotating Qdrant Credentials
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

# Rotating Qdrant Authentication Credentials

This guide will give an overview of how KubeDB Ops-manager rotates the authentication credentials of a `Qdrant` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How Rotate Auth Works

The Rotate Auth process consists of the following steps:

1. At first, a user creates a `Qdrant` CR.

2. `KubeDB-Provisioner` operator watches the `Qdrant` CR.

3. When the operator finds a `Qdrant` CR, it creates a `StatefulSet` and generates an `authSecret` containing the initial API key for the Qdrant database.

4. Then, in order to rotate the authentication credentials, the user creates a `QdrantOpsRequest` CR with `type: RotateAuth`. The user can optionally provide a new custom secret, or let KubeDB auto-generate a new API key.

5. `KubeDB` Ops-manager operator watches the `QdrantOpsRequest` CR.

6. When it finds a `QdrantOpsRequest` CR, it pauses the `Qdrant` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Qdrant` during the credential rotation process.

7. Then the `KubeDB` Ops-manager operator generates a new API key (or uses the provided secret), updates the `authSecret`, and restarts the pods in a rolling fashion to apply the new credentials.

8. After the successful credential rotation, the `KubeDB` Ops-manager updates the `Qdrant` object to reflect the updated auth state.

9. After the successful Rotate Auth, the `KubeDB` Ops-manager resumes the `Qdrant` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on rotating authentication credentials of a Qdrant database using `QdrantOpsRequest` CRD.
