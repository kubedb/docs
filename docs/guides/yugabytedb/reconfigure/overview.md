---
title: Reconfiguring YugabyteDB
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-reconfigure-overview
    name: Overview
    parent: yugabytedb-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring YugabyteDB

For a `YugabyteDBOpsRequest` with `spec.type: Reconfigure`, KubeDB Ops-manager:

1. Validates `spec.configuration` (a `configSecret`, inline `applyConfig` values, or `removeCustomConfig`). The keys are `master.flags`, `tserver.flags`, `ysql_pg.conf` and `ysql_hba.conf`; flags owned by the operator are rejected.
2. Merges the new configuration into the `<name>-config` Secret that holds the flag files (condition `ApplyConfiguration`).
3. Decides what needs a restart. With `restart: auto` (the default), a flag that YugabyteDB can change at runtime, or a PostgreSQL parameter that only needs a reload, is applied to the running servers (condition `SetRuntimeFlags`) and no pod restarts. If a changed flag cannot be set at runtime, the masters (`RollMasters`) and tservers (`RollTServers`) it affects are rolled with the rolling engine. `restart: "true"` always rolls, `restart: "false"` never does.
4. Marks the request `Successful`.

## Next Step

[Reconfigure YugabyteDB](/docs/guides/yugabytedb/reconfigure/reconfigure.md).
