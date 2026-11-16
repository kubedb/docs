---
title: Reconfigure YugabyteDB
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-reconfigure-guide
    name: Reconfigure
    parent: yugabytedb-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure YugabyteDB

This guide changes the configuration of a running `yb-demo` universe in two ways: a runtime PostgreSQL parameter (no restart) and a tserver flag that needs a restart.

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/reconfigure](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

Deploy `yb-demo` from the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md).

## Change a runtime parameter (no restart)

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-reconf-runtime
  namespace: demo
spec:
  type: Reconfigure
  databaseRef: {name: yb-demo}
  configuration:
    applyConfig:
      ysql_pg.conf: |
        work_mem = 8MB
    restart: auto
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure/ops-request-runtime.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-reconf-runtime created
```

Measured under load: **Successful in 40 s, no pod restarted**, `show work_mem` returned `8MB`, 75,690 successful calls, 0 failed calls, longest period without success 0.21 s.

## Change a flag that needs a restart

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-reconf-restart
  namespace: demo
spec:
  type: Reconfigure
  databaseRef: {name: yb-demo}
  configuration:
    applyConfig:
      tserver.flags: |
        --ysql_max_connections=300
    restart: auto
  rolling: {tserverWait: 20s}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure/ops-request-restart.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-reconf-restart created
```

Measured under load: **Successful in 241 s**, only the tservers were rolled, the flag appeared in the `tserver.flags` of the config Secret, 326,738 successful calls, 5 dropped connections to restarted pods, no other failure, longest period without success 0.44 s.

## Notes

- The inline keys are `master.flags`, `tserver.flags`, `ysql_pg.conf` and `ysql_hba.conf`. A `configSecret` (a Secret you create with the same keys) is also accepted. Only the inline form above was run.
- `removeCustomConfig: true` removes the previous custom configuration. This was not run.
- Flags that the operator owns (addresses, ports, placement, TLS) are rejected or dropped.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-demo-reconf-runtime yb-demo-reconf-restart
```
