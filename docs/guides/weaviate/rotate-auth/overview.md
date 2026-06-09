---
title: Rotating Weaviate Credentials
menu:
  docs_{{ .version }}:
    identifier: weaviate-rotate-auth-overview
    name: Overview
    parent: weaviate-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotating Weaviate Authentication Credentials

This guide will give an overview of how KubeDB Ops-manager rotates the authentication credentials of a `Weaviate` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)

## How Rotate Auth Works

The Rotate Auth process consists of the following steps:

1. At first, a user creates a `Weaviate` CR.

2. `KubeDB-Provisioner` operator watches the `Weaviate` CR.

3. When the operator finds a `Weaviate` CR, it creates a `StatefulSet` and generates an `authSecret` containing the initial API key for the Weaviate database. This API key is injected into Weaviate pods via the `AUTHENTICATION_APIKEY_ALLOWED_KEYS` environment variable.

4. Then, in order to rotate the authentication credentials, the user creates a `WeaviateOpsRequest` CR with `type: RotateAuth`. The user can optionally provide a new custom secret, or let KubeDB auto-generate a new API key.

5. `KubeDB` Ops-manager operator watches the `WeaviateOpsRequest` CR.

6. When it finds a `WeaviateOpsRequest` CR, it pauses the `Weaviate` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Weaviate` during the credential rotation process.

7. Then the `KubeDB` Ops-manager operator generates a new API key (or uses the provided secret), updates the `authSecret`, and restarts the pods in a rolling fashion to apply the new credentials.

8. After the successful credential rotation, the `KubeDB` Ops-manager updates the `Weaviate` object to reflect the updated auth state.

9. After the successful Rotate Auth, the `KubeDB` Ops-manager resumes the `Weaviate` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on rotating authentication credentials of a Weaviate database using `WeaviateOpsRequest` CRD.
