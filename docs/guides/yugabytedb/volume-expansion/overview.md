---
title: YugabyteDB Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-volume-expansion-overview
    name: Overview
    parent: yugabytedb-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Volume Expansion

> **Not verified end to end.** The operation is implemented but it could not be run to completion: the storage class of the test cluster (kind, `rancher.io/local-path`) has `allowVolumeExpansion: false` and cannot resize volumes.

A `YugabyteDBOpsRequest` with `spec.type: VolumeExpansion` grows the PVCs of the masters, the tservers, or both (`spec.volumeExpansion.master` and `.tserver`). The `mode` is `Online` (the pods keep running while the CSI driver resizes the volumes) or `Offline` (the pods are rolled so the file system can be resized).

Steps seen in the test: `OrphanPetSetPods`, `UpdatePetSets` (the PetSets are recreated with the new storage size and the pods are left alone), `ReadyPetSets`, then `UpdateVolumes`.

## Requirements

- A StorageClass with `allowVolumeExpansion: true` and a CSI driver that supports resizing.
- Free capacity in the backing storage.

## Next Step

[Expand YugabyteDB volumes](/docs/guides/yugabytedb/volume-expansion/volume-expansion.md).
