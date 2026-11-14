---
title: YugabyteDB YSQL Connection Manager
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-connection-manager-guide
    name: Connection Manager
    parent: yugabytedb-connection-manager
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB YSQL Connection Manager

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

The YSQL Connection Manager is the connection pooler built into YugabyteDB (`--enable_ysql_conn_mgr`). It lets many client connections share a smaller number of PostgreSQL backends. KubeDB exposes it as `spec.ysql.connectionManager`:

- `enabled`: starts the Connection Manager.
- `maxClientConnections`: `--ysql_conn_mgr_max_client_connections`, the number of client connections it accepts. YugabyteDB defaults to 10000. The minimum is 1.

When it is enabled, the tservers move PostgreSQL to **port 6433** inside the pods and the Connection Manager listens on **5433**, so clients keep using the usual YSQL port.

## Enable it at creation

The webhook requires `spec.ysql.enabled` for the Connection Manager. This example also turns on webserver TLS, so it needs a cert-manager `Issuer` named `yb-ca-issuer` (see the [TLS guide](/docs/guides/yugabytedb/tls/configure/index.md)):

```yaml
apiVersion: kubedb.com/v1alpha2
kind: YugabyteDB
metadata:
  name: yb-cm
  namespace: demo
spec:
  version: 2025.2.8
  ysql:
    enabled: true
    connectionManager:
      enabled: true
      maxClientConnections: 10000
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: yb-ca-issuer
    webserver: true
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/connection-manager/yugabytedb-connection-manager.yaml
```

## Change it on a running database

After the database exists, `spec.ysql.connectionManager` can only be changed by a `Reconfigure` request with the `spec.connectionManager` field. The pods restart one at a time (step `ApplyConnectionManager`). `spec.connectionManager` is only valid in a `Reconfigure` request. It can be given next to or instead of `spec.configuration`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-enable-connection-manager
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: yb-demo
  # PostgreSQL moves to port 6433 inside the pods and the Connection Manager listens on 5433.
  connectionManager:
    enabled: true
    maxClientConnections: 5000
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/connection-manager/reconfigure-connection-manager.yaml
```

The webhook rejects the request when YSQL is not enabled, when `maxClientConnections` is below 1, and when the setting equals the current one.

## Things to keep in mind

- Because PostgreSQL moves to port 6433, anything that talked to the pods on 5433 now reaches the Connection Manager, not PostgreSQL. Whether KubeDB's own health checks and SQL access work behind the Connection Manager was not verified.
- The Connection Manager is a YugabyteDB feature with its own limits; read the upstream documentation before relying on it.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-enable-connection-manager
$ kubectl delete yb -n demo yb-cm
```
