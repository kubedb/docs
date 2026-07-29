---
title: Milvus Update Version Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-update-version-overview
    name: Overview
    parent: milvus-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update Version of Milvus

This guide will give an overview on how the KubeDB Ops-manager operator updates the version of a `Milvus` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

## How Update Version Process Works

The set of Milvus versions KubeDB can run is described by `MilvusVersion` catalog objects. Updating moves a running database from its current `MilvusVersion` to a newer, non-deprecated one.

A `MilvusOpsRequest` of type `UpdateVersion` names the target version:

```yaml
spec:
  type: UpdateVersion
  updateVersion:
    targetVersion: 2.6.11
```

The flow is:

1. A user creates a `MilvusOpsRequest` of type `UpdateVersion`.
2. The operator validates that `targetVersion` is an existing, non-deprecated `MilvusVersion`, then pauses the `Milvus` database.
3. The operator updates the container images in the PetSets to the target version.
4. Pods are restarted (standalone: the single pod; distributed: each role) so they come up on the new image.
5. The operator resumes the database and marks the `MilvusOpsRequest` as `Successful`.

The Recommendation Engine also generates an `UpdateVersion` recommendation automatically when a newer catalog version becomes available for a running database.

In the next doc, we will see a step-by-step guide on updating the version of a Milvus database.
