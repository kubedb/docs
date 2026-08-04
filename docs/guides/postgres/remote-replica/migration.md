---
title: Migrate a Self-Managed PostgreSQL into KubeDB
menu:
  docs_{{ .version }}:
    identifier: pg-remote-replica-migration
    name: Migration from Self-Managed
    parent: pg-remote-replica
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Migrate a Self-Managed PostgreSQL into KubeDB

This guide migrates a **self-managed** PostgreSQL — one KubeDB does not manage: a bare
StatefulSet, a VM in another datacenter, a hardened installation whose superuser you will
never see — into a KubeDB-managed cluster with seconds of write downtime, using the remote
replica mechanism: seed with `pg_basebackup`, stream until the lag is zero, stop writes,
cut over.

Everything below was executed end to end against two source flavours:

- **bare**: stock `postgres:17.4`, superuser credentials shared with us
- **hardened**: `scram-sha-256`, a replication-only user, a restricted `pg_hba.conf`, and
  **no `postgres` role at all** (`initdb` run as a different superuser)

> Note: YAML files used in this tutorial are stored in [docs/guides/postgres/remote-replica/migration-yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/migration-yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Limitations — read first

| Limitation | Why | Way out |
|---|---|---|
| **Self-managed sources only** | Managed services (RDS, Cloud SQL, …) do not expose physical replication to external standbys | Logical replication (different machinery, not this guide) |
| **Major version must match** | Physical standby; WAL format is per-major | Pick the matching `version:` from the KubeDB catalog; upgrade after migration |
| **`kubectl-dba remote-config` cannot be used** | It lists the source's pods by label and `kubectl exec`s into them — a foreign source has no pods | Hand-craft the AppBinding + secret (Step 2 of this guide) |
| **The source must be reachable at port 5432** | The generated connection strings and `primary_conninfo` do not carry a port | Expose the source at `:5432` on the address you put in the AppBinding |
| **A `LOGIN`-able `postgres` role must exist in the source catalog before cutover** | Promotion connects locally *as* `postgres` to re-key passwords | One-liner on the source, replicates automatically (Step 5) |
| **libc / collation mismatch** | A glibc-built source seeded onto a musl-based image records a collation version the new libc cannot confirm; text indexes are ordered by the old libc | Post-cutover: clear `datcollversion`, `REINDEX` collation-sensitive indexes (Step 7) |
| **Extensions** | WAL replays fine, but queries touching extension objects need the `.so` present in the KubeDB image | Inventory `pg_extension` on the source first; anything not in the image blocks this method |
| **Tablespaces with absolute paths** | Paths from the source host do not exist in the container | Consolidate to the default tablespace first |
| **No replication slot for steady-state streaming** | The replica streams slotless; if it disconnects longer than the source retains WAL, it cannot resume | Set `wal_keep_size` on the source to cover your longest tolerable outage. The initial seed itself needs **no** WAL retention config: `pg_basebackup -Xs` backs it with a temporary slot it creates and drops itself |

## Step 1: prepare the source (the source DBA does this)

A dedicated replication user — superuser **not** required:

```sql
CREATE ROLE migrator LOGIN REPLICATION PASSWORD '<migrator-password>';
```

Two `pg_hba.conf` lines — one for the WAL stream, one because KubeDB's monitor queries
`pg_stat_replication` over a normal connection to the `postgres` database:

```
host  replication  migrator  <replica-source-cidr>  scram-sha-256
host  postgres     migrator  <replica-source-cidr>  scram-sha-256
```

then `SELECT pg_reload_conf();`.

**Don't guess `<replica-source-cidr>`** — NAT between the clusters decides what the source
sees. Deploy the replica first (Step 3) and read the address out of the source's own log:

```
FATAL:  no pg_hba.conf entry for replication connection from host "10.42.0.17", user "migrator", no encryption
```

That `host` value is the address to allowlist. The replica retries the seed forever, so
fixing `pg_hba.conf` after the fact needs no restart of anything — the seed simply proceeds
on the next retry.

Requirements that are defaults on modern PostgreSQL: `wal_level = replica`,
`max_wal_senders` ≥ 2 free (the `-Xs` seed briefly uses two), `listen_addresses` covering
the ingress path.

## Step 2: hand-craft the AppBinding

On the **destination** cluster (`remote-config` cannot generate this for a foreign source):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: source-pg-auth
  namespace: demo
stringData:
  username: migrator
  password: "<migrator-password>"
type: kubernetes.io/basic-auth
---
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: source-pg
  namespace: demo
  labels:
    app.kubernetes.io/name: postgreses.kubedb.com
spec:
  clientConfig:
    service:
      name: 10.2.0.30          # source address; must serve PostgreSQL on 5432
      path: /
      port: 5432
      query: sslmode=disable   # TLS source: verify-ca + a tlsSecret carrying the SOURCE's CA
      scheme: postgresql
  secret:
    apiGroup: ""
    kind: Secret
    name: source-pg-auth
  type: kubedb.com/postgres
  version: "17.4"
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/migration-yamls/source-appbinding.yaml
```

## Step 3: deploy the warm replica

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-mig
  namespace: demo
spec:
  remoteReplica:
    sourceRef:
      name: source-pg
      namespace: demo
  authSecret:
    name: source-pg-auth       # yes — the migration user's credentials; see below
  healthChecker:
    disableWriteCheck: true
  clientAuthMode: md5
  standbyMode: Hot
  replicas: 1
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 10Gi
  storageType: Durable
  deletionPolicy: Halt
  version: "17.4"
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/migration-yamls/pg-mig.yaml
kubectl wait pg pg-mig -n demo --for=jsonpath='{.status.phase}'=Ready --timeout=900s
```

**Why `authSecret` points at the migration user:** `pg_basebackup` copies the source's
`pg_authid` wholesale, so after the seed the replica's passwords are the *source's*
passwords. KubeDB's health checker authenticates with the authSecret — against a hardened
source whose superuser password you don't have, the only credentials guaranteed to work in
the copied catalog are the migration user's. With them, the CR reports `Ready` throughout
the warm phase. (Against a bare source that shared its `postgres` password, an authSecret
with `username: postgres` works the same way.)

The seed streams WAL concurrently with the copy (`pg_basebackup -Xs`, temporary slot,
self-cleaning), so the source needs no WAL-retention configuration for it.

## Step 4: watch the lag

The coordinator sidecar logs the replica's byte lag behind the source. Sampling starts at
5s; every consecutive zero-lag reading doubles the interval up to 300s, and any non-zero
reading resets it to 5s — so a catching-up replica (the phase you actually watch) is
sampled every 5 seconds:

```bash
kubectl logs -f pg-mig-0 -n demo -c pg-coordinator | grep LagMonitor
```

```
[LagMonitor] Pod pg-mig-0: lag=0 B (in sync with source); next check in 10s
[LagMonitor] Pod pg-mig-0: lag=0 B (in sync with source); next check in 20s
[LagMonitor] Pod pg-mig-0: lag=48681472 B behind source; next check in 5s
[LagMonitor] Pod pg-mig-0: lag=0 B (in sync with source); next check in 10s
```

## Step 5: pre-cutover checks (while streaming, zero risk)

The replica is readable, so every check runs against live data.

**The `postgres` role.** Promotion runs `ALTER USER postgres … PASSWORD` connecting locally
*as* `postgres`. If the source's `initdb` used another superuser name, that role does not
exist and promotion cannot complete:

```bash
kubectl exec -n demo pg-mig-0 -c postgres -- \
  psql -U migrator -d postgres -tAc "SELECT count(*) FROM pg_roles WHERE rolname='postgres';"
```

If `0`, have the source DBA run — it replicates within seconds, re-check to confirm:

```sql
CREATE ROLE postgres LOGIN SUPERUSER;
```

**Collation-sensitive indexes** (matters whenever source and replica differ in libc — a
glibc VM or Debian-based image onto KubeDB's musl-based image always does):

```bash
kubectl exec -n demo pg-mig-0 -c postgres -- psql -U migrator -d postgres -tAc \
 "SELECT count(*) FROM pg_index i WHERE EXISTS (SELECT 1 FROM unnest(i.indcollation) col
   WHERE col <> 0 AND col NOT IN (SELECT oid FROM pg_collation WHERE collname IN ('C','POSIX')));"
```

Remember the number — it is your post-cutover `REINDEX` workload (Step 7).

**Extensions:** `SELECT extname FROM pg_extension;` on the source; anything beyond what the
KubeDB image ships must be resolved before you rely on this method.

## Step 6: cutover

**1. Stop application writes at the source — and verify they stopped.** Do not trust
"the app was told to disconnect"; killing a client does not necessarily kill server-side
sessions. The source's WAL position is the truth — it must be frozen:

```bash
# run twice, 2s apart, on the source; proceed only when identical
SELECT pg_current_wal_lsn();
```

**2. Wait for the replica to apply everything** (compare against the frozen LSN from step 1;
`>=`, not equality — the source still emits checkpoint WAL after quiescing):

```bash
kubectl exec -n demo pg-mig-0 -c postgres -- psql -U migrator -d postgres -tAc \
  "SELECT pg_wal_lsn_diff(pg_last_wal_replay_lsn(), '<frozen-lsn>') >= 0;"   # wait for: t
```

**3. Promote.** Remove `spec.remoteReplica`; for `spec.authSecret`, one rule:

- authSecret's `username` **is** `postgres` → keep it. Its password becomes the superuser
  password at promotion.
- authSecret's `username` is **not** `postgres` (the hardened case) → **remove it too**.
  The operator generates `<name>-auth` with user `postgres` and a fresh password, and
  promotion re-keys the copied catalog to it. You never needed the source's superuser
  password at any point.

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/migration-yamls/pg-mig-standalone.yaml
kubectl delete pod pg-mig-0 -n demo
```

**4. Redirect writes and confirm.** The database is migrated when a write is accepted with
the final credentials:

```bash
PGPASSWORD=$(kubectl get secret pg-mig-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
kubectl exec -n demo pg-mig-0 -c postgres -- env PGPASSWORD="$PGPASSWORD" \
  psql -h 127.0.0.1 -U postgres -d postgres -c "SELECT NOT pg_is_in_recovery();"
```

**Measured on the runs behind this guide** (single-replica, ~1–2 GB databases, same-LAN
clusters): write gap from last replicated write to first accepted write **43 s** (bare,
measured from server-side row timestamps); cutover-initiated to first accepted write with
the operator-generated password **37 s** (hardened). The time is dominated by the pod
recreate and promotion, not by data size — the data was already there.

## Step 7: after the cutover

**Collation record** — on a libc-mismatched migration every connection warns
`database "postgres" has no actual collation version, but a version was recorded`, and
`ALTER DATABASE … REFRESH COLLATION VERSION` fails with `invalid collation version change`
because the new libc reports no version to refresh *to*. Clear the recorded version, and
rebuild whatever Step 5's index count found:

```sql
UPDATE pg_database SET datcollversion = NULL WHERE datcollversion IS NOT NULL;
-- for each collation-sensitive index found in Step 5:
REINDEX INDEX <name>;   -- or REINDEX DATABASE if the count was large
```

**Hygiene** — the migration user came along in the copied catalog:

```sql
DROP OWNED BY migrator; DROP ROLE migrator;
```

The source's other roles and databases are all present — that is the migration payload.
From here the database is a normal KubeDB Postgres: scale it, enable TLS, attach a remote
replica of its own for DR.

## Next Steps

- [Remote Replica overview](/docs/guides/postgres/remote-replica/remotereplica.md)
- [Cross-Cluster DR with Bidirectional Failover](/docs/guides/postgres/remote-replica/advanced-setup.md)
- [Zero-Data-Loss Cross-Cluster Replication](/docs/guides/postgres/remote-replica/synchronous.md)
