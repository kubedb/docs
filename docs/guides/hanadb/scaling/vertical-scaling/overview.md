---
title: HanaDB Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: hanadb-scaling-vertical-overview
    name: Overview
    parent: hanadb-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# Vertical Scaling HanaDB

KubeDB supports updating CPU and memory resources for the HanaDB container with a `HanaDBOpsRequest` of type `VerticalScaling`.

## Before You Begin

You should be familiar with the following KubeDB concepts:

- [HanaDB](/docs/guides/hanadb/concepts/hanadb.md)
- [HanaDBOpsRequest](/docs/guides/hanadb/concepts/opsrequest.md)

## How Vertical Scaling Works

The vertical scaling process consists of the following steps:

1. A user creates a `HanaDB` object.
2. The KubeDB Provisioner provisions the required PetSet, services, secrets, and related resources.
3. To change container resources, the user creates a `HanaDBOpsRequest` with `spec.type: VerticalScaling`.
4. The KubeDB Ops Manager pauses the referenced `HanaDB` object while the operation is running.
5. Ops Manager updates the `HanaDB` pod template and the underlying PetSet resources.
6. Ops Manager restarts HanaDB pods as needed so the new resource requirements take effect.
7. After the operation succeeds, Ops Manager resumes the `HanaDB` object.

See the [Vertical Scaling guide](/docs/guides/hanadb/scaling/vertical-scaling/vertical-scaling.md) for a step-by-step example.
