---
title: Oracle Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-volume-expansion-overview
    name: Overview
    parent: guides-oracle-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Oracle Volume Expansion

This guide will give an overview on how KubeDB Ops-manager operator expands the volume of an `Oracle` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops-manager operator expands the volumes of `Oracle` database components. Open the image in a new tab to see the enlarged version.

The Volume Expansion process consists of the following steps:

1. At first, a user creates an `Oracle` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `Oracle` CR.

3. When the operator finds an `Oracle` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc. The database pods use `PersistentVolumeClaim`s for durable storage.

4. Each `PetSet` creates `PersistentVolumeClaim`s according to the `volumeClaimTemplate` provided in the `PetSet`.

5. Then, in order to expand the volume of the `Oracle` database the user creates an `OracleOpsRequest` CR with the desired volume size.

6. `KubeDB` Ops-manager operator watches the `OracleOpsRequest` CR.

7. When it finds an `OracleOpsRequest` CR, it pauses the `Oracle` object so that the `KubeDB` Provisioner operator doesn't perform any operations on the `Oracle` object during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator will expand the persistent volume to reach the expected size, defined in the `OracleOpsRequest` CR. Volume expansion can be performed either `Online` (without stopping the pod) or `Offline` (the pod is recreated after the PVC is resized), as defined in `spec.volumeExpansion.mode`.

9. After the successful Volume Expansion of the related `PetSet`s, the `KubeDB` Ops-manager operator updates the `Oracle` object to reflect the updated state.

> **Note:** Volume expansion is only supported for storage classes whose underlying provisioner allows it (`allowVolumeExpansion: true`). For example, `local-path` does **not** support volume expansion, while a CSI driver such as `longhorn` does.

In the next docs, we are going to show a step by step guide on Volume Expansion of various Oracle database components using `OracleOpsRequest` CRD.

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
