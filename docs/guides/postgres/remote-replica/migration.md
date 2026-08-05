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
| **Non-standard source ports are honored via the AppBinding** | Every source-facing connection (seed, streaming, monitor, recovery) uses `spec.clientConfig.service.port` | Set `port:` in the AppBinding (Step 2); it defaults to 5432 when unset |
| **A `LOGIN`-able `postgres` role must exist in the source catalog before cutover** | Promotion connects locally *as* `postgres` to re-key passwords | One-liner on the source, replicates automatically (Step 5) |
| **The source's OS/libc must match the KubeDB image's** | Physical replication carries text indexes ordered by the source's libc collations; running them under a different libc (glibc↔musl) risks silent index corruption. **A distribution mismatch is not supported.** | Pick a `PostgresVersion` whose `spec.db.image` uses the same base as the source (the `Official` 17.4 image is Alpine/musl). Verify with the check in Step 5 |
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
      name: 10.2.0.30          # source address
      path: /
      port: 5432               # honored: set to the source's actual port
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

**Distribution / libc match** — the source and the KubeDB image must be built against the
same libc. Compare `version()` on both sides; the platform triple must match (here:
`x86_64-pc-linux-musl` on both). If they differ, stop and pick a matching `PostgresVersion`
— a mismatch is not supported:

```bash
kubectl exec -n <source-ns> <source-pod> -- psql -U <user> -d postgres -tAc "SELECT version();"
kubectl exec -n demo pg-mig-0 -c postgres -- psql -U migrator -d postgres -tAc "SELECT version();"
```

```
PostgreSQL 17.4 on x86_64-pc-linux-musl, compiled by gcc (Alpine 14.2.0) 14.2.0, 64-bit
PostgreSQL 17.4 on x86_64-pc-linux-musl, compiled by gcc (Alpine 14.2.0) 14.2.0, 64-bit
```

With matching distributions the replica connects without any collation-version warning —
if you see `database "postgres" has no actual collation version, but a version was
recorded` on every connection, the libc differs and this migration path does not apply.

**Extensions:** `SELECT extname FROM pg_extension;` on the source; anything beyond what the
KubeDB image ships must be resolved before you rely on this method.

## Step 6: lossless cutover

A migration is only successful if the migrated database contains **every row the source
ever acknowledged**. This runbook makes that a verified gate, not an assumption: cutover
is forbidden until a content fingerprint of the source and the replica are identical.

Two practical notes before the steps:

- Run the replica-side verification queries **over the local unix socket as the source's
  own application user** (`admin` in this guide). That role exists in the copied catalog
  and can read its own tables; the replica's `local ... trust` pg_hba line means no
  password is needed. The replication user typically cannot `SELECT` from the
  application's tables, and the final `postgres` password does not exist yet.
- On a libc-mismatched migration every `psql` invocation prints
  `WARNING: database "postgres" has no actual collation version ...` on stderr. It is
  harmless here and fixed in Step 7 — but don't let it confuse scripts that merge stderr
  into stdout.

**1. Stop application writes at the source — and verify they stopped.** Do not trust
"the app was told to disconnect": killing a client does not kill server-side sessions
(in our testing, a "killed" writer kept inserting for ten more minutes). The source's
WAL position and row counts are the truth — both must be frozen across two samples:

```bash
# on the source, twice, 2 s apart; proceed only when BOTH are identical
SELECT pg_current_wal_lsn();
SELECT count(*) FROM writes;    -- your busiest table(s)
```

**2. Freeze the source fingerprint.** Order-independent, content-sensitive, and constant
memory, so it works on tables of any size (extend the pattern to every table you care
about):

```sql
SELECT (SELECT count(*) FROM payload)
  ||'|'|| (SELECT coalesce(sum(hashtextextended(id::text||data, 0)), 0) FROM payload)
  ||'|'|| (SELECT count(*) FROM writes)
  ||'|'|| (SELECT coalesce(sum(hashtextextended(id::text||origin||ts::text, 0)), 0) FROM writes);
```

Record the result — this is the value the migrated database must reproduce.

**3. Wait for the replica to apply everything** (compare against the frozen LSN from
step 1; `>=`, not equality — the source still emits checkpoint WAL after quiescing):

```bash
kubectl exec -n demo pg-mig-0 -c postgres -- psql -U admin -d postgres -tAc \
  "SELECT pg_wal_lsn_diff(pg_last_wal_replay_lsn(), '<frozen-lsn>') >= 0;"   # wait for: t
```

**4. THE ZERO-LOSS GATE.** Run the *same* fingerprint query on the replica, over the
socket as the application user:

```bash
kubectl exec -n demo pg-mig-0 -c postgres -- psql -U admin -d postgres -tAc "<fingerprint SQL>"
```

- **Identical** → every acknowledged row is on the replica; proceed.
- **Different** → **do not cut over.** Nothing is lost — the replica is still streaming
  and the source is intact. Find what is still moving (a second application? a cron?)
  and return to step 1.

**5. Promote.** Remove `spec.remoteReplica`; for `spec.authSecret`, one rule:

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

**6. Redirect writes, then prove zero loss.** The database is migrated when a write is
accepted with the final credentials:

```bash
PGPASSWORD=$(kubectl get secret pg-mig-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
kubectl exec -n demo pg-mig-0 -c postgres -- env PGPASSWORD="$PGPASSWORD" \
  psql -h 127.0.0.1 -U postgres -d postgres -c \
  "INSERT INTO writes(origin) VALUES ('first-write-after-migration') RETURNING id;"
```

Then re-run the fingerprint query on the migrated database, scoped to exclude
post-cutover writes — it must equal the value frozen in step 2. If your write traffic
carries server-side timestamps, the migration's write gap is computable from the data
itself, immune to clock skew between clusters:

```sql
SELECT min(ts) FILTER (WHERE origin = 'first-write-after-migration')
     - max(ts) FILTER (WHERE origin <> 'first-write-after-migration') FROM writes;
```

**Measured on the runs behind this guide** (single-replica, ~1–2 GB databases, same-LAN
clusters, a ~5 TPS writer running until cutover): write gap from last committed source
transaction to first accepted write on the migrated database **37.63 s**, measured from
server-side row timestamps, with all 2904 in-flight-era transactions verified present by
fingerprint. The full procedure is in the verification section below.

## Verifying the cutover: zero transaction loss and measured downtime

Run a writer against the source during the warm phase and prove afterwards that every
committed transaction arrived and how long writes were unavailable. Everything below is
from a live run: a ~5 TPS writer, 2904 transactions committed before cutover.

**Writer** (on the source; stopped by creating a sentinel file — never rely on killing a
client, server-side sessions survive it):

```bash
kubectl exec -n <source-ns> <source-pod> -- bash -c 'rm -f /tmp/stopw
cat > /tmp/writer.sh <<"EOF"
#!/bin/bash
while [ ! -f /tmp/stopw ]; do
  psql -U <appuser> -d postgres -qc "INSERT INTO writes(origin) VALUES ('"'"'writer'"'"');"
  sleep 0.2
done
EOF
chmod +x /tmp/writer.sh; nohup /tmp/writer.sh >/tmp/writer.log 2>&1 &'
```

**Stop and verify quiesce** — the source's WAL position must be identical across two
samples; only then is the snapshot below the full truth:

```bash
kubectl exec -n <source-ns> <source-pod> -- touch /tmp/stopw
# run twice, 2s apart; proceed only when both values are identical
kubectl exec -n <source-ns> <source-pod> -- psql -U <user> -tAc "SELECT pg_current_wal_lsn();"
```

**Snapshot the source truth** (count, highest id, and a fingerprint over every id):

```bash
kubectl exec -n <source-ns> <source-pod> -- psql -U <user> -d postgres -tAc   "SELECT count(*)||'|'||max(id)||'|'||md5(string_agg(id::text,',' ORDER BY id)) FROM writes;"
```

```
2904|2904|9b409e931a89ac8134e47cdcb91da247
```

**Wait for the replica to apply everything** (`>=` against the frozen LSN — the source
still emits checkpoint WAL after quiescing, so never compare for equality):

```bash
kubectl exec -n demo pg-mig-0 -c postgres -- psql -U migrator -d postgres -tAc   "SELECT pg_wal_lsn_diff(pg_last_wal_replay_lsn(), '<frozen-lsn>') >= 0;"   # wait for: t
```

Then cut over as in Step 6. Once the first write is accepted, compare:

```bash
PGPASSWORD=$(kubectl get secret pg-mig-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
kubectl exec -n demo pg-mig-0 -c postgres -- env PGPASSWORD="$PGPASSWORD"   psql -h 127.0.0.1 -U postgres -d postgres -tAc   "SELECT count(*)||'|'||max(id)||'|'||md5(string_agg(id::text,',' ORDER BY id)) FROM writes WHERE origin='writer';"
```

```
2904|2904|9b409e931a89ac8134e47cdcb91da247      <- identical to the source snapshot: zero loss
```

**Downtime, from server-side row timestamps** (skew-free — both rows were stamped by a
database clock):

```bash
kubectl exec -n demo pg-mig-0 -c postgres -- env PGPASSWORD="$PGPASSWORD"   psql -h 127.0.0.1 -U postgres -d postgres -tAc   "SELECT round(extract(epoch FROM (SELECT min(ts) FROM writes WHERE origin<>'writer')
                              - (SELECT max(ts) FROM writes WHERE origin='writer'))::numeric,2)||' s';"
```

```
37.63 s
```

The measured gap is dominated by the pod recreate and promotion, not by data size. Note
that the first post-cutover `id` may jump ahead (2921 in this run, after 2904): PostgreSQL
sequences advance in cached increments across a promotion. That is normal sequence
behavior, not lost rows — the fingerprint comparison above is the loss check.

## Step 7: after the cutover

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
