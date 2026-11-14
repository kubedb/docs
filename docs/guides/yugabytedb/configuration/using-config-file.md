---
title: Run YugabyteDB with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-using-config-file
    name: Custom Configuration
    parent: yugabytedb-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration

YugabyteDB is configured with flag files for the two roles and with PostgreSQL configuration for YSQL. KubeDB renders them from three sources. A later source overrides an earlier one:

1. the defaults of the `YugabyteDBVersion` (`spec.defaultFlags`) and the flags KubeDB itself needs,
2. your configuration (`spec.configuration.secretName` and `spec.configuration.inline`),
3. flags that the operator owns (addresses, ports, placement, TLS, replication factor, `master_join_existing_universe`, and so on), which are always applied last. A user value for an operator-owned flag is rejected by the webhook or silently dropped.

## Keys

| Key | Applied to |
|-----|------------|
| `master.flags` | `yb-master` flags, one `--flag=value` per line |
| `tserver.flags` | `yb-tserver` flags, one `--flag=value` per line |
| `ysql_pg.conf` | PostgreSQL parameters (`ysql_pg_conf_csv`), one `name = value` per line |
| `ysql_hba.conf` | `pg_hba` rules (`ysql_hba_conf_csv`) |

## At creation (not verified)

You can create a Secret with these keys and reference it from the `YugabyteDB`. This path (configuration at creation time) was not run in the test cluster; the [Reconfigure](/docs/guides/yugabytedb/reconfigure/reconfigure.md) operation below was.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: yb-custom-config
  namespace: demo
stringData:
  tserver.flags: |
    --ysql_max_connections=300
  ysql_pg.conf: |
    work_mem = 8MB
---
apiVersion: kubedb.com/v1alpha2
kind: YugabyteDB
metadata:
  name: yb-demo
  namespace: demo
spec:
  version: 2025.2.8
  configuration:
    secretName: yb-custom-config
  # ... topology as in the quickstart
```

## On a running database (verified)

Use a [Reconfigure ops request](/docs/guides/yugabytedb/reconfigure/reconfigure.md). KubeDB changes the flag at runtime when every changed flag can be set at runtime (a PostgreSQL parameter such as `work_mem`, for instance) and rolls the pods only when a flag needs a restart.

## Next Steps

- [Reconfigure overview](/docs/guides/yugabytedb/reconfigure/overview.md)
- [YugabyteDB CRD](/docs/guides/yugabytedb/concepts/yugabytedb.md)
