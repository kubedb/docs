---
title: HanaDB Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: hanadb-volume-expansion-overview
    name: Overview
    parent: hanadb-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# HanaDB Volume Expansion

KubeDB supports expanding HanaDB persistent volumes with a `HanaDBOpsRequest` of type `VolumeExpansion`.

## Before You Begin

You should be familiar with the following KubeDB concepts:

- [HanaDB](/docs/guides/hanadb/concepts/hanadb.md)
- [HanaDBOpsRequest](/docs/guides/hanadb/concepts/opsrequest.md)

The StorageClass used by the HanaDB PVC must support volume expansion.

## How Volume Expansion Works

The volume expansion process consists of the following steps:

1. A user creates a `HanaDB` object with durable storage.
2. The KubeDB Provisioner provisions the required PVCs through the PetSet volume claim template.
3. To expand storage, the user creates a `HanaDBOpsRequest` with `spec.type: VolumeExpansion`.
4. The KubeDB Ops Manager pauses the referenced `HanaDB` object while the operation is running.
5. Ops Manager expands the target PVCs and updates the HanaDB storage specification.
6. After the operation succeeds, Ops Manager resumes the `HanaDB` object.

See the [Volume Expansion guide](/docs/guides/hanadb/volume-expansion/volume-expansion.md) for a step-by-step example.
