---
title: Weaviate Rotate Authentication Overview
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

# Rotate Authentication of Weaviate

This guide will give you an overview of how KubeDB Ops Manager rotates the API-key authentication of a `Weaviate` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Weaviate Quickstart](/docs/guides/weaviate/quickstart/quickstart.md)

## How Rotate Authentication Process Works

By default, KubeDB enables API-key authentication for a Weaviate cluster and stores the generated key in a Secret named `<database-name>-auth`. Rotating authentication replaces this key.

The rotate authentication process consists of the following steps:

1. The user creates a `WeaviateOpsRequest` CR of type `RotateAuth` referencing the `Weaviate` database. There are two modes:
   - **Operator-generated key** — when no `spec.authentication.secretRef` is provided, the Ops Manager generates a brand-new random API key.
   - **User-provided key** — when `spec.authentication.secretRef` is provided, the Ops Manager uses the key from the referenced Secret.

2. `KubeDB` Ops Manager watches for the `WeaviateOpsRequest` CR and halts the `Weaviate` object.

3. The Ops Manager updates the auth Secret with the new key. It also keeps the previous key in a `*-PREV` data field so that in-flight clients still using the old key have a short grace window, and it records the `activeFrom` timestamp.

4. The Ops Manager updates the PetSet and restarts the pods one by one so they pick up the new key.

5. After successfully rotating the key, the `KubeDB` Ops Manager resumes the `Weaviate` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on rotating the authentication of a Weaviate cluster.
