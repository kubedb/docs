---
title: Updating YugabyteDB Version
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-update-version-overview
    name: Overview
    parent: yugabytedb-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating YugabyteDB Version

A `YugabyteDBOpsRequest` with `spec.type: UpdateVersion` performs a rolling upgrade to the version named by `spec.updateVersion.targetVersion` (a `YugabyteDBVersion`). The source version must be allowed by the `updateConstraints` of the target version, and the YSQL major version must be the same (true for every path in the catalog).

Conditions, in order:

1. `Preflight`: all masters and tservers alive, no under-replicated tablets.
2. `RecordAutoFlagsConfig`: records the AutoFlags configuration before the upgrade.
3. `UpdatePetSets`: writes the new image into the PetSets (`OnDelete`, so no pod restarts yet).
4. `RollMasters` and `RollTServers`: the rolling engine, one pod at a time, with leader drain for the tservers.
5. `PromoteAutoFlags`: promotes the new version's AutoFlags once every pod runs the new version.
6. `Soak`: watches the universe for `spec.updateVersion.soakDuration` (default 10m). Until the upgrade is finalized it can still be rolled back.
7. `FinalizeUpgrade`: finalizes automatically after the soak.

## Next Step

[Upgrade YugabyteDB version](/docs/guides/yugabytedb/update-version/versionupgrading/).
