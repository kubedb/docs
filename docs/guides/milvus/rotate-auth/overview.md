---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-rotate-auth-overview
    name: Overview
    parent: milvus-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Milvus

This guide will give an overview on how the KubeDB Ops-manager operator rotates the authentication credentials of a `Milvus` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

## How Rotate Authentication Process Works

Milvus authentication is enabled by default (`spec.disableSecurity` defaults to `false`). When you do not provide `spec.authSecret`, KubeDB auto-generates a `kubernetes.io/basic-auth` secret named `<db>-auth` containing a `root` user and a random password.

A `MilvusOpsRequest` of type `RotateAuth` rotates that credential. There are two modes:

1. **Operator-generated password** — omit `spec.authentication`. The operator generates a new random password and updates the existing auth secret in place.
2. **User-supplied credentials** — set `spec.authentication.secretRef.name` to a `Secret` (with `username`/`password` keys) you created. The operator switches the database to use your secret.

The flow is:

1. A user creates a `MilvusOpsRequest` of type `RotateAuth`.
2. The operator validates the request and pauses the `Milvus` database.
3. The credential is updated dynamically inside the running Milvus, then the rendered configuration and PetSets are reconciled to reference the new secret.
4. Pods are restarted to ensure every component uses the new credential.
5. The operator resumes the database and marks the `MilvusOpsRequest` as `Successful`.

## Relationship with the Recommendation Engine

Two fields on `spec.authSecret` drive automatic auth-rotation recommendations:

- **`rotateAfter`** — the maximum age of the credential. Once the secret is older than this duration, the Recommendation Engine emits a `RotateAuth` recommendation.
- **`activeFrom`** — the timestamp from which the current credential is considered active (also stamped on the secret via the `kubedb.com/auth-active-from` annotation). It is the reference point `rotateAfter` is measured from.

See the [Recommendation Engine guide](/docs/guides/milvus/recommendation/guide.md) for the end-to-end flow.

In the next doc, we will see a step-by-step guide on rotating authentication of a Milvus database.
