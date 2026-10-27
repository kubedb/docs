---
title: YugabyteDB Vertical Scaling
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-vertical-scaling-overview
    name: Overview
    parent: yugabytedb-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Vertical Scaling

This page describes how KubeDB Ops-manager changes the CPU and memory of the masters and tservers.

For a `YugabyteDBOpsRequest` with `spec.type: VerticalScaling`, Ops-manager:

1. Validates the request and pauses the reconciliation of the database.
2. Writes the new resources (`spec.verticalScaling.master` and `spec.verticalScaling.tserver`) into the PetSets (condition `UpdateResources`). The PetSets use `OnDelete`, so nothing restarts yet.
3. Recomputes `--memory_limit_hard_bytes` from the new memory limit and the catalog percentages (85% for masters and 70% for tservers by default) and restarts the pods with the rolling engine: masters first (`RollMasters`), then tservers one at a time (`RollTServers`). Memory limits of YugabyteDB are static, so even an in-place resize needs a process restart; `mode` defaults to `Restart`.
4. Resumes the database and marks the request `Successful`.

## Next Step

[Scale YugabyteDB vertically](/docs/guides/yugabytedb/scaling/vertical-scaling/scale-vertically/).
