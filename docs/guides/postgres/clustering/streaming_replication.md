---
title: Using Postgres Streaming Replication
menu:
  docs_{{ .version }}:
    identifier: pg-streaming-replication-clustering
    name: Streaming Replication
    parent: pg-clustering-postgres
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Streaming Replication

Streaming Replication lets one or more *standby* servers stay up to date with a *primary* by shipping
and replaying [WAL](https://www.postgresql.org/docs/current/wal-intro.html) records continuously. The
standby connects to the primary, which streams WAL records as they are generated.

KubeDB supports two replication modes, controlled by `spec.streamingMode`:

- **`Asynchronous`** (default) — the primary does not wait for standbys before acknowledging a commit.
  Lowest latency, but a failover can lose the last few transactions that had not yet reached a standby.
- **`Synchronous`** — the primary waits for one or more standbys to confirm a commit before returning
  success. Stronger durability at the cost of some commit latency. This is configured with the
  `spec.synchronousReplicationConfig` API described [below](#synchronous-streaming-replication).

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured
to communicate with your cluster. If you do not already have a cluster, you can create one by using
[kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps
[here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres/clustering](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres/clustering) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Asynchronous Streaming Replication (default)

The example below deploys a three-node PostgreSQL cluster with the default asynchronous streaming
replication.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  version: "18.3"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

In this example:

- The `Postgres` object creates three PostgreSQL servers, indicated by the **`replicas`** field.
- One server becomes the *primary* and the other two become *standby* servers.

Create the object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/clustering/ha-postgres.yaml
postgres.kubedb.com/ha-postgres created
```

KubeDB operator creates three Pods. Each Pod has two containers: `postgres` and the `pg-coordinator`
sidecar that runs the leader election and manages replication.

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=ha-postgres" -L kubedb.com/role
NAME            READY   STATUS    RESTARTS   AGE   ROLE
ha-postgres-0   2/2     Running   0          55s   primary
ha-postgres-1   2/2     Running   0          48s   standby
ha-postgres-2   2/2     Running   0          41s   standby
```

Here:

- Pod `ha-postgres-0` is the *primary*, indicated by the label `kubedb.com/role=primary`.
- Pods `ha-postgres-1` and `ha-postgres-2` are *standby* servers, labelled `kubedb.com/role=standby`.

KubeDB creates the following Services for the cluster:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=ha-postgres"
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                               AGE
ha-postgres           ClusterIP   10.43.19.110   <none>        5432/TCP,2379/TCP                     62s
ha-postgres-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP,2384/TCP   62s
ha-postgres-standby   ClusterIP   10.43.64.57    <none>        5432/TCP                              62s
```

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=ha-postgres" -o=custom-columns=NAME:.metadata.name,SELECTOR:.spec.selector
NAME                  SELECTOR
ha-postgres           map[... kubedb.com/role:primary ...]
ha-postgres-pods      map[... (all pods) ...]
ha-postgres-standby   map[... kubedb.com/role:standby ...]
```

Here:

- Service `ha-postgres` always targets the current *primary* (selector includes `kubedb.com/role=primary`).
  Use it for read/write traffic.
- Service `ha-postgres-standby` targets the *standby* Pods (`kubedb.com/role=standby`). Use it for
  read-only traffic when running *hot* standbys.
- Service `ha-postgres-pods` is a headless service targeting every Pod.

Retrieve the credentials to connect:

```bash
$ kubectl get secret -n demo ha-postgres-auth -o jsonpath='{.data.username}' | base64 -d
postgres
$ kubectl get secret -n demo ha-postgres-auth -o jsonpath='{.data.password}' | base64 -d
```

You can check `pg_stat_replication` on the primary to see who is streaming:

```bash
$ kubectl exec -n demo ha-postgres-0 -c postgres -- psql -U postgres \
    -c "select application_name, client_addr, state, sync_state from pg_stat_replication order by 1;"
 application_name | client_addr |   state   | sync_state
------------------+-------------+-----------+------------
 ha-postgres-1    | 10.42.0.176 | streaming | async
 ha-postgres-2    | 10.42.0.180 | streaming | async
(2 rows)
```

Both standbys are streaming with `sync_state = async` — this is asynchronous replication.

### Automatic failover

If the *primary* fails, the `pg-coordinator` promotes a healthy *standby* to *primary*. Let's delete
the primary Pod and watch the failover:

```bash
$ kubectl delete pod -n demo ha-postgres-0
pod "ha-postgres-0" deleted
```

Within a few seconds a new primary is elected (≈14s in this run). After things settle:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=ha-postgres" -L kubedb.com/role
NAME            READY   STATUS    RESTARTS   AGE   ROLE
ha-postgres-0   2/2     Running   0          20m   standby
ha-postgres-1   2/2     Running   0          21m   primary
ha-postgres-2   2/2     Running   0          21m   standby
```

Pod `ha-postgres-1` is now the *primary*; the old primary `ha-postgres-0` rejoined as a *standby*.
The `ha-postgres` Service automatically follows the new primary, so applications reconnect without
configuration changes.

```bash
$ kubectl exec -n demo ha-postgres-1 -c postgres -- psql -U postgres \
    -c "select application_name, client_addr, state, sync_state from pg_stat_replication order by 1;"
 application_name | client_addr |   state   | sync_state
------------------+-------------+-----------+------------
 ha-postgres-0    | 10.42.0.182 | streaming | async
 ha-postgres-2    | 10.42.0.180 | streaming | async
(2 rows)
```

## Synchronous Streaming Replication

With synchronous replication, a commit on the primary does not return until the required number of
standbys confirm they received (or applied) the transaction. KubeDB exposes full control over
PostgreSQL's `synchronous_standby_names` and `synchronous_commit` through
`spec.synchronousReplicationConfig`.

> **Availability:** `spec.synchronousReplicationConfig` is available from **KubeDB `v2026.7.10`**. It is
> rendered by the database init scripts, which require **`postgres-init` image `>= 0.20.0`**. Confirm
> your chosen version points to a compatible init image:
>
> ```bash
> $ kubectl get postgresversion 17.4 -o jsonpath='{.spec.initContainer.image}'
> ghcr.io/kubedb/postgres-init:0.20.0
> ```
>
> If a version still references an older init image, synchronous config falls back to the legacy
> `ANY 1` behaviour — pick a version whose `initContainer.image` is `>= 0.20.0`.

### The `synchronousReplicationConfig` API

| Field | Type | Default | Description |
|---|---|---|---|
| `mode` | `Any` \| `First` | `Any` | `Any` = quorum (`ANY N`); `First` = priority order (`FIRST N`). |
| `numSyncReplicas` | integer | `1` | The `N` in `ANY N` / `FIRST N`. Must be `>= 1` and `< spec.replicas` (or `<= len(standbyNames)` when names are given). |
| `commitLevel` | `On` \| `RemoteApply` \| `RemoteWrite` \| `Local` \| `Off` | `RemoteWrite` | Maps to `synchronous_commit`. |
| `standbyNames` | `[]string` | auto (all pods, ascending) | Explicit, ordered list of standby `application_name`s. For `First` mode the order is the priority. No empty strings or duplicates. Mutually exclusive with `useWildcard`. |
| `useWildcard` | bool | `false` | Use `*` to match any connected standby. Renders `... (*)`. Mutually exclusive with `standbyNames`. |

These fields render `synchronous_standby_names` as `<MODE> <numSyncReplicas> (<names or *>)`.

> ⚠️ **YAML gotcha:** `On` and `Off` are parsed as booleans in YAML. Always **quote** them:
> `commitLevel: "On"`. An unquoted `commitLevel: On` is rejected by the admission webhook.

### Example 1 — Quorum (`Any`)

Wait for **any 2** of the standbys to acknowledge each commit:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sync-postgres
  namespace: demo
spec:
  version: "17.4"
  replicas: 3
  standbyMode: Hot
  streamingMode: Synchronous
  synchronousReplicationConfig:
    mode: Any            # Any (quorum) | First (priority-ordered)
    numSyncReplicas: 2   # wait for this many standbys before a commit returns
    commitLevel: RemoteWrite  # On | RemoteApply | RemoteWrite | Local | Off
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/clustering/sync-postgres.yaml
postgres.kubedb.com/sync-postgres created
```

Once `Ready`, inspect the rendered configuration on the primary:

```bash
$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres -tAc "SHOW synchronous_standby_names;"
ANY 2 ("sync-postgres-1","sync-postgres-2")

$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres -tAc "SHOW synchronous_commit;"
remote_write

$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres \
    -c "select application_name, state, sync_state from pg_stat_replication order by 1;"
 application_name |   state   | sync_state
------------------+-----------+------------
 sync-postgres-1  | streaming | quorum
 sync-postgres-2  | streaming | quorum
(2 rows)
```

With `Any` (quorum) mode the eligible standbys report `sync_state = quorum`: a commit succeeds as soon
as **any 2** of them acknowledge.

### Example 2 — Priority order (`First` + `standbyNames`)

Use `First` mode with an explicit ordered `standbyNames` list to make `sync-postgres-2` the preferred
synchronous standby, and use the strongest `commitLevel`:

```yaml
spec:
  version: "17.4"
  replicas: 3
  standbyMode: Hot
  streamingMode: Synchronous
  synchronousReplicationConfig:
    mode: First
    numSyncReplicas: 1
    commitLevel: "On"        # quoted — see the YAML gotcha above
    standbyNames:
    - sync-postgres-2        # highest priority
    - sync-postgres-1
```

```bash
$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres -tAc "SHOW synchronous_standby_names;"
FIRST 1 ("sync-postgres-2","sync-postgres-1")

$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres -tAc "SHOW synchronous_commit;"
on

$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres \
    -c "select application_name, sync_state from pg_stat_replication order by 1;"
 application_name | sync_state
------------------+------------
 sync-postgres-1  | potential
 sync-postgres-2  | sync
(2 rows)
```

`sync-postgres-2` is first in the list, so it is the active `sync` standby; `sync-postgres-1` is a
`potential` standby that is promoted into the synchronous set only if `sync-postgres-2` becomes
unavailable.

### Example 3 — Wildcard (`useWildcard`)

When standby `application_name`s are not known in advance, use `useWildcard: true` to accept any
connected standby:

```yaml
spec:
  streamingMode: Synchronous
  synchronousReplicationConfig:
    mode: Any
    numSyncReplicas: 1
    useWildcard: true        # mutually exclusive with standbyNames
```

```bash
$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres -tAc "SHOW synchronous_standby_names;"
ANY 1 (*)

$ kubectl exec -n demo sync-postgres-0 -c postgres -- psql -U postgres \
    -c "select application_name, sync_state from pg_stat_replication order by 1;"
 application_name | sync_state
------------------+------------
 sync-postgres-1  | quorum
 sync-postgres-2  | quorum
(2 rows)
```

### Choosing a `commitLevel`

`commitLevel` maps directly to PostgreSQL's `synchronous_commit` and trades durability against latency:

| `commitLevel` | `synchronous_commit` | Meaning |
|---|---|---|
| `Off` | `off` | Commit returns without even waiting for local WAL flush. Fastest, least durable. |
| `Local` | `local` | Wait for local WAL flush only; standbys are not waited on. |
| `RemoteWrite` | `remote_write` | **(default)** Wait until a synchronous standby has written WAL to its OS buffer (not necessarily fsync'd). Lowest-latency synchronous option. |
| `On` | `on` | Wait until a synchronous standby has flushed WAL to disk. Safe, higher latency. |
| `RemoteApply` | `remote_apply` | Wait until a synchronous standby has *applied* WAL (queries there see the commit). Highest latency. |

> `synchronousReplicationConfig` is applied when the database is provisioned. Provide it in the initial
> `Postgres` spec for the mode you want.

### Failover with synchronous replication

Failover works exactly like the [asynchronous case](#automatic-failover): deleting the primary Pod
causes `pg-coordinator` to promote a standby, and the `sync-postgres` Service follows the new primary.
The difference is durability — because a commit only returned after a synchronous standby acknowledged
it, transactions that were confirmed to the client survive the promotion.

## Warm vs Hot standby

`spec.standbyMode` controls whether standbys accept read connections:

- **`Warm`** (default) — standbys replicate but reject all client connections. Connect to the primary only.
- **`Hot`** — standbys accept read-only queries. Connect to them through the `*-standby` Service.

The `sync-postgres` cluster above uses `standbyMode: Hot`, so its standbys serve reads. A write on a
standby is rejected, while reads succeed:

```bash
$ kubectl exec -n demo sync-postgres-1 -c postgres -- psql -U postgres -c "CREATE DATABASE standby_test;"
ERROR:  cannot execute CREATE DATABASE in a read-only transaction

$ kubectl exec -n demo sync-postgres-1 -c postgres -- psql -U postgres \
    -c "SELECT pg_is_in_recovery(), pg_last_wal_receive_lsn();"
 pg_is_in_recovery | pg_last_wal_receive_lsn
-------------------+-------------------------
 t                 | 0/504F9F8
(1 row)
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo pg/ha-postgres pg/sync-postgres -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo pg/ha-postgres pg/sync-postgres
kubectl delete ns demo
```

## Next Steps

- Learn about the [synchronous replication overview](/docs/guides/postgres/synchronous/index.md).
- Run PostgreSQL with a [custom configuration file](/docs/guides/postgres/configuration/using-config-file.md).
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
```
