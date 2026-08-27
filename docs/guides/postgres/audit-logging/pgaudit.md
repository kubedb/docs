---
title: Audit Logging with pgAudit
menu:
  docs_{{ .version }}:
    identifier: pg-audit-logging-pgaudit
    name: Audit Logging with pgAudit
    parent: pg-audit-logging-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Audit Logging with pgAudit

Regulated workloads — core banking, payments, anything under PCI DSS, SOX or a
central bank's IT security circular — generally have to answer three questions
about the database after the fact: **who did it, what exactly did they do, and
when**. PostgreSQL's own `log_statement` cannot answer them at an acceptable
cost, because it is all-or-nothing.

[pgAudit](https://github.com/pgaudit/pgaudit) can. This guide builds a
production-shaped audit configuration on a KubeDB-managed cluster, proves each
control with a real query and the log record it produced, and covers the
operational parts that decide whether the trail is admissible: log volume,
collection, and what happens on a standby.

The walkthrough uses a licensed `16.9-appscode-bookworm-ext` build, but nothing
here is specific to it — pgAudit ships in every KubeDB `-ext` PostgreSQL image.

## Before You Begin

- A Kubernetes cluster with KubeDB installed ([setup](/docs/setup/README.md)).
- A PostgreSQL version whose image bundles pgAudit — any `-ext` variant:

  ```bash
  $ kubectl get postgresversion | grep -- -ext
  ```

- A namespace to work in. This guide uses `bank`.

  ```bash
  $ kubectl create ns bank
  namespace/bank created
  ```

> Manifests used here live in the
> [docs repository](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres/audit-logging).

{{< notice type="note" message="`16.9-appscode-bookworm-ext` is a licensed [Postgres Enterprise by AppsCode](/docs/guides/postgres/enterprise/postgres-enterprise.md) build and additionally needs `spec.license`. If you are using a community `-ext` version instead, drop the `license` block from the manifests below; everything else is identical." >}}

## What to audit, and what not to

The instinct is to audit everything. On a core banking database that produces a
trail nobody can afford to store or search, so the useful configuration splits
into two halves.

**Session auditing** records whole classes of statement for every user. Use it
for everything that *changes* state:

| Class | Covers | Why a bank wants it |
|---|---|---|
| `write` | `INSERT`, `UPDATE`, `DELETE`, `TRUNCATE` | The ledger changed — by whom, to what |
| `ddl` | `CREATE`, `ALTER`, `DROP` | Schema change control |
| `role` | `GRANT`, `REVOKE`, `CREATE ROLE` | Privilege escalation, segregation of duties |
| `function` | `DO` blocks, function calls | Logic executed inside the database |
| `read` | `SELECT`, `COPY … TO` | **Usually too expensive — see below** |

**Object auditing** records reads, but only of the objects you nominate. This is
the half that makes read auditing affordable: instead of logging every `SELECT`
on the database, you log only those touching the columns that actually matter —
national ID numbers, card numbers, balances.

The configuration below uses session auditing for all four write-ish classes and
object auditing for reads.

## Step 1 — Write the audit configuration

pgAudit is a shared library, so it has to be preloaded; it cannot simply be
created as an extension. KubeDB passes custom PostgreSQL settings through a
Secret referenced by `spec.configSecret`, under the key `user.conf`.

`audit.conf`:

```ini
# ---------------------------------------------------------------------------
# Audit engine
# ---------------------------------------------------------------------------
shared_preload_libraries = 'pgaudit'

# Session audit logging: every statement that changes data, schema, privileges
# or executable code. Reads are deliberately NOT in this list.
pgaudit.log = 'ddl, role, write, function'

# Do not audit reads of the system catalogs; psql \d and every ORM's
# introspection would otherwise drown the trail.
pgaudit.log_catalog = off

# Audit records go to the server log only, never to the client. A client that
# can read its own audit records can also learn what is being audited.
pgaudit.log_client = off
pgaudit.log_level = log

# Record the statement text and its bound parameter values.
pgaudit.log_parameter = on
pgaudit.log_statement = on

# One audit record per relation touched, so a multi-table statement does not
# collapse into a single ambiguous entry.
pgaudit.log_relation = on
pgaudit.log_statement_once = off

# Object audit logging: SELECTs are audited only on objects this role has been
# granted access to.
pgaudit.role = 'auditor'

# ---------------------------------------------------------------------------
# Log record content -- an audit record is only evidence if it names the actor
# ---------------------------------------------------------------------------
log_line_prefix = '%m [%p] %q%u@%d %r app=%a sid=%c/%l '
log_error_verbosity = default

# pgaudit supersedes log_statement; leaving both on double-logs everything.
log_statement = 'none'

# Successful connection logging is off deliberately -- see "Log volume" below.
# Failed authentication is still logged regardless of this setting.
log_connections = off
log_disconnections = off
```

The `log_line_prefix` is not decoration. Without `%u` (user), `%r` (source
address) and `%c` (session ID) an audit record says what happened but not who
did it or from where, which is exactly the part an auditor asks about.

```bash
$ kubectl create secret generic bank-pg-audit-config -n bank \
    --from-file=user.conf=./audit.conf
secret/bank-pg-audit-config created
```

## Step 2 — Deploy the cluster

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: bank-pg
  namespace: bank
spec:
  version: "16.9-appscode-bookworm-ext"
  replicas: 3
  standbyMode: Hot
  license:                       # licensed builds only
    secretRef:
      name: bank-pg-license
      key: license.pem
  configSecret:
    name: bank-pg-audit-config
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f bank-pg.yaml
postgres.kubedb.com/bank-pg created

$ kubectl get pg -n bank -w
NAME      VERSION                      STATUS         AGE
bank-pg   16.9-appscode-bookworm-ext   Provisioning   15s
bank-pg   16.9-appscode-bookworm-ext   Ready          60s
```

Confirm the settings are actually in force before trusting them. Assert on
`pg_settings`, never on the file you supplied — a rejected or misspelled
parameter is a `WARNING` at startup, not an error, and leaves the policy
silently inert:

```bash
$ kubectl exec -n bank bank-pg-0 -c postgres -- psql -U postgres -c \
    "SELECT name, setting, source FROM pg_settings WHERE name LIKE 'pgaudit%' ORDER BY name;"
              name              |          setting           |       source
--------------------------------+----------------------------+--------------------
 pgaudit.log                    | ddl, role, write, function | configuration file
 pgaudit.log_catalog            | off                        | configuration file
 pgaudit.log_client             | off                        | configuration file
 pgaudit.log_level              | log                        | configuration file
 pgaudit.log_parameter          | on                         | configuration file
 pgaudit.log_parameter_max_size | 0                          | default
 pgaudit.log_relation           | on                         | configuration file
 pgaudit.log_rows               | off                        | default
 pgaudit.log_statement          | on                         | configuration file
 pgaudit.log_statement_once     | off                        | configuration file
 pgaudit.role                   | auditor                    | configuration file
```

`source = configuration file` is the thing to look for. Anything you set that
still reads `default` did not take effect.

## Step 3 — Create the extension and the audit role

Preloading the library is what makes auditing happen; `CREATE EXTENSION` adds
the SQL-level objects and is still required for a supported installation.

```bash
$ kubectl exec -it -n bank bank-pg-0 -c postgres -- psql -U postgres
```

```sql
CREATE EXTENSION IF NOT EXISTS pgaudit;

-- The audit role is never logged into. It exists only so that GRANTs to it
-- mark which objects pgAudit should record reads of.
CREATE ROLE auditor NOLOGIN;
```

## Step 4 — Nominate the sensitive data

This is the whole of object auditing: grant the `auditor` role read access to
what must be watched. Column-level grants work, which is what keeps the trail
narrow.

The examples below assume a core banking schema of customers, accounts, card
details and a ledger. If you want to follow along exactly,
[`schema.sql`](https://github.com/kubedb/docs/blob/{{< param "info.version" >}}/docs/examples/postgres/audit-logging/schema.sql)
creates that schema together with the extension, the audit role, the grants in
this step and the roles in the next one:

```bash
$ kubectl exec -i -n bank bank-pg-0 -c postgres -- psql -U postgres < schema.sql
```

The two grants that drive object auditing are these:

```sql
-- PII: only these two columns are sensitive, not the whole row
GRANT SELECT (national_id, date_of_birth) ON core.customers TO auditor;

-- Card data: the entire table is sensitive
GRANT SELECT ON core.card_details TO auditor;
```

Reads of `core.customers.full_name` alone now go unaudited, while any read that
touches `national_id` produces a record. Nothing else needs configuring.

## Step 5 — Give actions a named actor

An audit trail whose every record says `postgres` is worth very little. Create
real roles for the humans and services that touch the database:

```sql
CREATE ROLE teller LOGIN PASSWORD '<strong-password>';
CREATE ROLE payments_svc LOGIN PASSWORD '<strong-password>';

GRANT USAGE ON SCHEMA core TO teller, payments_svc;
GRANT SELECT ON core.customers, core.accounts, core.card_details TO teller;
GRANT SELECT, INSERT ON core.transactions TO payments_svc;
GRANT SELECT, UPDATE ON core.accounts TO payments_svc;
```

pgAudit redacts credentials in the records it writes, so role management is
itself auditable without leaking secrets:

```
AUDIT: SESSION,3,1,ROLE,CREATE ROLE,,,CREATE ROLE teller LOGIN PASSWORD <REDACTED>,<none>
```

## Reading an audit record

Every record is a CSV payload after the `AUDIT:` marker:

```
AUDIT: SESSION,1,1,WRITE,INSERT,TABLE,core.transactions,"INSERT …","1,-12500.00,wire_out"
       │       │ │ │     │      │     │                 │          └─ parameter values
       │       │ │ │     │      │     │                 └─ statement text
       │       │ │ │     │      │     └─ object name
       │       │ │ │     │      └─ object type
       │       │ │ │     └─ command
       │       │ │ └─ class
       │       │ └─ substatement number
       │       └─ statement number within the session
       └─ SESSION (class-based) or OBJECT (grant-based)
```

`SESSION` versus `OBJECT` in the first field tells you which half of the
configuration produced the record — useful when tuning.

## Use cases, with the records they produce

Each of the following was run against the cluster built above; the log lines are
the actual output.

### A teller reads a customer's national ID

```bash
$ psql -U teller -c "SELECT full_name, national_id FROM core.customers WHERE id=1;"
```

```
2026-08-27 07:03:30.013 UTC [2184] teller@postgres 127.0.0.1(57684) app=psql sid=6a8fe142.888/1 LOG:  AUDIT: OBJECT,1,1,READ,SELECT,TABLE,core.customers,"SELECT full_name, national_id FROM core.customers WHERE id=1;",<none>
```

`OBJECT` class, and the prefix names the actor (`teller`), the source address and
the session. This is the record that answers "who looked at this customer?"

### The same teller reads only non-sensitive columns

```bash
$ psql -U teller -c "SELECT id, full_name FROM core.customers;"
```

**No audit record is produced.** This is the point of object auditing — the
trail stays small enough to be searchable, because unremarkable reads do not
enter it.

### A teller reads a card number

```bash
$ psql -U teller -c "SELECT pan FROM core.card_details;"
```

```
2026-08-27 07:03:30.218 UTC [2198] teller@postgres 127.0.0.1(57740) app=psql sid=6a8fe142.896/1 LOG:  AUDIT: OBJECT,1,1,READ,SELECT,TABLE,core.card_details,SELECT pan FROM core.card_details;,<none>
```

### A service account books a transfer

Parameterised, as a real application would issue it:

```sql
INSERT INTO core.transactions (account_id, amount, kind) VALUES ($1,$2,$3);
-- bound: 1, -12500.00, 'wire_out'
```

```
2026-08-27 07:03:30.322 UTC [2206] payments_svc@postgres 127.0.0.1(57770) app=psql sid=6a8fe142.89e/1 LOG:  AUDIT: SESSION,1,1,WRITE,INSERT,TABLE,core.transactions,"INSERT INTO core.transactions (account_id, amount, kind) VALUES ($1,$2,$3)
	;","1,-12500.00,wire_out"
```

Note the final field: `1,-12500.00,wire_out`. Because `pgaudit.log_parameter` is
on, the record contains the values actually written, not just the statement
template. Without it you would know a transfer was booked but not for how much.

### A DBA changes the schema

```bash
$ psql -U postgres -c "ALTER TABLE core.transactions ADD COLUMN channel text;"
```

```
2026-08-27 07:03:30.428 UTC [2214] postgres@postgres [local] app=psql sid=6a8fe142.8a6/1 LOG:  AUDIT: SESSION,1,1,DDL,ALTER TABLE,TABLE,core.transactions,ALTER TABLE core.transactions ADD COLUMN channel text;,<none>
```

### A privilege is revoked

```bash
$ psql -U postgres -c "REVOKE UPDATE ON core.accounts FROM teller;"
```

```
2026-08-27 07:03:30.529 UTC [2221] postgres@postgres [local] app=psql sid=6a8fe142.8ad/1 LOG:  AUDIT: SESSION,1,1,ROLE,REVOKE,TABLE,,REVOKE UPDATE ON core.accounts FROM teller;,<none>
```

### An unauthorised write is attempted

```bash
$ psql -U teller -c "DELETE FROM core.transactions;"
ERROR:  permission denied for table transactions
```

```
2026-08-27 07:03:30.642 UTC [2228] teller@postgres 127.0.0.1(57808) app=psql sid=6a8fe142.8b4/1 ERROR:  permission denied for table transactions
```

The statement never executed, so pgAudit does not record it — but the `ERROR`
carries the same prefix, so the attempt is still attributable. Alert on these:
a legitimate application does not routinely attempt what it is not entitled to.

### Authentication fails

```
2026-08-27 07:03:09.765 UTC [1996] teller@postgres 10.42.0.67(33904) app=[unknown] sid=6a8fe12d.7cc/1 FATAL:  password authentication failed for user "teller"
```

This is logged by PostgreSQL itself, not pgAudit, and **is not affected by
`log_connections = off`** — which is what makes turning that setting off safe.

{{< notice type="info" message="Testing this yourself needs care. KubeDB's generated `pg_hba.conf` maps the local socket and `127.0.0.1` to `trust`, so a wrong password supplied through `kubectl exec … psql` succeeds and logs nothing. Drive failed-login tests pod to pod against a peer's IP, where `md5` applies." >}}

## Log volume, and why `log_connections` is off

pgAudit audits **statements, not rows**. A single statement writing a thousand
rows produces one record:

```sql
INSERT INTO core.transactions (account_id, amount, kind)
SELECT 1, -1.00, 'load_test' FROM generate_series(1,1000);
```

→ one audit record, under 300 bytes. Whereas 200 individual `INSERT`s produced
51,987 bytes, or about **260 bytes per audited statement**. Size retention
against statement rates, not row counts.

That measurement is also why the configuration disables successful connection
logging. With `log_connections = on` on this three-replica cluster:

| | records per minute |
|---|---:|
| `connection authorized` / `disconnection` | 601 |
| pgAudit records | 5 |
| total log lines | 1120 |

Of 241 connections sampled in one minute, 234 were KubeDB's own health probes
and the replication connections — `user=postgres` and `application_name=pg_isready`.
The audit trail was 0.4% of its own log. With the setting off, the same
three-replica cluster logs **nothing at all while idle**, so every line that does
appear is worth reading — and failed authentication is still captured.

If a control you are held to explicitly requires logging successful sessions,
turn it back on — but filter the operator's health-check connections in your log
pipeline and budget for the volume, which is around 1.6M lines per day per pod
even on an idle cluster.

## Collecting the trail

Two properties of this setup decide whether the trail survives to be audited.

**Each pod audits only its own traffic.** A standby verifies and applies the same
configuration, and audits reads served locally:

```bash
$ kubectl exec -n bank bank-pg-2 -c postgres -- psql -U postgres -tAc \
    "SELECT pg_is_in_recovery(), current_setting('pgaudit.role');"
t|auditor
```

```
2026-08-27 07:04:03.348 UTC [1057] teller@postgres 127.0.0.1(42000) app=psql sid=6a8fe163.421/1 LOG:  AUDIT: OBJECT,1,1,READ,SELECT,TABLE,core.customers,SELECT national_id FROM core.customers WHERE id=1;,<none>
```

So a complete trail has to be collected from **every** pod, not just the primary.
If reporting traffic is routed to standbys, those reads exist only in that
standby's log.

**pgAudit writes to the PostgreSQL log and nowhere else.** It creates no tables,
so the trail cannot be queried over SQL:

```bash
$ kubectl exec -n bank bank-pg-0 -c postgres -- psql -U postgres -tAc \
    "SELECT count(*) FROM pg_class WHERE relname LIKE '%pgaudit%';"
0
```

In Kubernetes the log is the container's stdout, which means it is rotated by the
kubelet and **discarded when the pod is deleted** — and a pod is deleted on every
restart, version upgrade and scaling operation. A trail left in `kubectl logs` is
not a retained audit trail.

Ship it off the node before you call the setup complete: run a log collector
(Fluent Bit, Vector, Promtail) as a DaemonSet, select the `postgres` container in
the database namespace, and forward to storage with the retention and
immutability your policy requires. Filtering on the `AUDIT:` marker separates the
audit stream from ordinary server logging.

## Changing the policy on a running cluster

Audit requirements change. Update the Secret and apply a `Reconfigure`
ops request; KubeDB rolls the change out replica by replica.

```bash
$ kubectl create secret generic bank-pg-audit-config -n bank \
    --from-file=user.conf=./audit.conf --dry-run=client -o yaml | kubectl apply -f -
secret/bank-pg-audit-config configured
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: bank-pg-tune-audit
  namespace: bank
spec:
  type: Reconfigure
  databaseRef:
    name: bank-pg
  configuration:
    configSecret:
      name: bank-pg-audit-config
```

```bash
$ kubectl apply -f reconfigure.yaml
postgresopsrequest.ops.kubedb.com/bank-pg-tune-audit created

$ kubectl get postgresopsrequest -n bank -w
NAME                 TYPE          STATUS        AGE
bank-pg-tune-audit   Reconfigure   Progressing   30s
bank-pg-tune-audit   Reconfigure   Successful    4m52s
```

Because `shared_preload_libraries` is involved, this is a restarting change —
KubeDB restarts the replicas one at a time, which took just under five minutes
for this three-replica cluster. Confirm the new values with the `pg_settings`
query from Step 2 afterwards. Data, grants and roles are unaffected.

## Tuning notes

- **`pgaudit.log_parameter`** captures the values written, which is what makes a
  `WRITE` record useful — and also means account numbers and other sensitive
  values land in the log. That is a deliberate trade: the audit log inherits the
  sensitivity of the data it describes, so it needs the same access controls.
  Set it to `off` if your policy cannot accommodate that.
- **`pgaudit.log_rows`** (default `off`) appends the affected row count to each
  record — the trailing `7` below. Useful for spotting bulk extraction; it does
  not log the rows themselves.

  ```
  AUDIT: SESSION,1,1,WRITE,INSERT,TABLE,core.transactions,"INSERT … generate_series(1,7);",<none>,7
  ```
- **`read` in `pgaudit.log`** audits every `SELECT` from every user. Reach for
  object auditing first; enable the class only for a specific investigation.
- **`misc_set`** captures `SET` statements, including `SET ROLE`. Worth adding
  where segregation of duties depends on role switching.
- **Object auditing errs towards recording too much.** A statement that does not
  name an audited column can still produce a record — `SELECT count(*)` on a
  table with column-level grants does. For an audit control, over-recording is
  the safe direction.
- **The operator appears in the trail.** KubeDB's own provisioning shows up as
  `postgres` activity, including an `ALTER ROLE` at bootstrap. Expect it, and do
  not mistake it for unexplained superuser access.

## Cleaning up

```bash
$ kubectl delete pg -n bank bank-pg
$ kubectl delete secret -n bank bank-pg-audit-config bank-pg-license
$ kubectl delete ns bank
```

## Next Steps

- [Run a licensed Postgres Enterprise build](/docs/guides/postgres/enterprise/postgres-enterprise.md).
- [Extensions available in the `-ext` images](/docs/guides/postgres/custom-versions/extensions.md).
- [Custom configuration](/docs/guides/postgres/configuration/using-config-file.md) reference.
- [Reconfigure](/docs/guides/postgres/reconfigure/overview.md) ops requests.
- [Monitor your database](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
