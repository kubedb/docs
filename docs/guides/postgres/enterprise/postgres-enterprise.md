---
title: Run Postgres Enterprise by AppsCode
menu:
  docs_{{ .version }}:
    identifier: pg-enterprise-guide
    name: Run Postgres Enterprise
    parent: pg-enterprise-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Postgres Enterprise by AppsCode

**Postgres Enterprise by AppsCode** is a commercially supported PostgreSQL
distribution built by AppsCode. It is licensed per deployment: the server
verifies a license certificate issued to you, so running it takes one extra
piece of setup compared with the community images. This guide covers that setup
end to end — obtaining the certificate, handing it to KubeDB, and confirming the
database is running under it.

{{< notice type="info" message="Two separate licenses are in play, and it is worth knowing which is which. The **KubeDB platform license**, passed as `global.license` when you install the operator, covers KubeDB itself. The **database license** on this page is issued for Postgres Enterprise and is read by the database server. You need both, and one does not substitute for the other." >}}

## Before You Begin

- A Kubernetes cluster with KubeDB installed. If you don't have one, follow the
  [installation instructions](/docs/setup/README.md).
- `kubectl` configured against that cluster.
- A namespace to work in. This guide uses `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

> The YAML manifests used here are available in the
> [docs repository](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres/enterprise).

## Step 1 — Obtain a database license

Request a license for the Postgres Enterprise distribution from
[appscode.com/issue-license?p=postgres](https://appscode.com/issue-license?p=postgres).
You will receive an X.509 certificate in PEM form. Save it locally, for example
as `license.pem`.

It is worth checking what you were issued before using it. `notAfter` is the end
of your license term:

```bash
$ openssl x509 -in license.pem -noout -subject -dates
subject=C = postgres, ST = enterprise, O = postgres-enterprise, OU = postgres-enterprise, CN = e366e720-f748-4ca4-b68d-a3bc2dba87f4
notBefore=Aug 24 03:44:53 2026 GMT
notAfter=Sep 23 03:44:53 2026 GMT
```

The `CN` is your license ID; quote it when contacting AppsCode support. The
`O` field must be `postgres-enterprise` — that is what marks the certificate as
valid for this product.

## Step 2 — Find the licensed versions

Licensed builds are shipped as `PostgresVersion` objects whose `spec.distribution`
is `AppsCode`:

```bash
$ kubectl get postgresversion -o custom-columns='NAME:.metadata.name,VERSION:.spec.version,DISTRIBUTION:.spec.distribution,LICENSE:.spec.license.required,IMAGE:.spec.db.image' | grep AppsCode
16.9-appscode                16.9      AppsCode       true      ghcr.io/appscode-images/postgres-enterprise:16.9-alpine
16.9-appscode-bookworm       16.9      AppsCode       true      ghcr.io/appscode-images/postgres-enterprise:16.9-bookworm
16.9-appscode-bookworm-ext   16.9      AppsCode       true      ghcr.io/appscode-images/postgres-enterprise:16.9-bookworm-ext
16.9-appscode-ext            16.9      AppsCode       true      ghcr.io/appscode-images/postgres-enterprise:16.9-alpine-ext
```

The four variants differ only in base OS and bundled extensions:

| Version | Base OS | Runs as UID | Bundled extensions |
|---|---|---|---|
| `16.9-appscode` | Alpine | 70 | — |
| `16.9-appscode-bookworm` | Debian bookworm | 999 | — |
| `16.9-appscode-ext` | Alpine | 70 | pgvector, PostGIS, pg_repack, pg_cron, pgaudit |
| `16.9-appscode-bookworm-ext` | Debian bookworm | 999 | pgvector, PostGIS, pg_repack, pg_cron, pgaudit |

`LICENSE` being `true` (`spec.license.required`) is what makes `spec.license`
mandatory on every `Postgres` object that uses the version.

## Step 3 — Store the license in a Secret

The certificate is delivered to the database through a Secret in the **same
namespace** as the `Postgres` object:

```bash
$ kubectl create secret generic pg-enterprise-license -n demo \
    --from-file=license.pem=./license.pem
secret/pg-enterprise-license created
```

The key name is yours to choose; `license.pem` is the default KubeDB assumes when
`spec.license.secretRef.key` is omitted.

## Step 4 — Deploy the database

Reference the Secret from `spec.license`:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-enterprise
  namespace: demo
spec:
  version: "16.9-appscode"
  replicas: 1
  license:
    secretRef:
      name: pg-enterprise-license
      key: license.pem
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f postgres-enterprise.yaml
postgres.kubedb.com/pg-enterprise created
```

Watch it come up:

```bash
$ kubectl get pg -n demo -w
NAME            VERSION         STATUS         AGE
pg-enterprise   16.9-appscode   Provisioning   5s
pg-enterprise   16.9-appscode   Ready          45s
```

`spec.license.secretRef.key` may be omitted — KubeDB defaults it to
`license.pem` — but setting it explicitly is the better habit, and it is
required if you store the certificate under a different key.

## Step 5 — Confirm the license was accepted

Three independent checks, in increasing order of detail.

**The database reports the enterprise build.** A community image would answer
`PostgreSQL 16.9 ...` here:

```bash
$ kubectl exec -it -n demo pg-enterprise-0 -c postgres -- psql -U postgres -tAc 'SELECT version();'
Postgres Enterprise by AppsCode 16.9 on x86_64-pc-linux-musl, compiled by gcc (Alpine 14.2.0) 14.2.0, 64-bit
```

It reports the PostgreSQL version it is built on, so existing drivers, tooling
and extensions work against it unchanged:

```bash
$ kubectl exec -it -n demo pg-enterprise-0 -c postgres -- psql -U postgres -tAc 'SHOW server_version; SHOW server_version_num;'
16.9
160009
```

**The server logged its acceptance.** Each verification records the license it
accepted (a first boot logs two lines, since database initialisation is verified
as well as the server start):

```bash
$ kubectl logs -n demo pg-enterprise-0 -c postgres | grep -i 'license accepted'
2026-08-26 06:04:25.059 UTC [64] LOG:  license accepted: id (serial) 1884894798739465099, CN e366e720-f748-4ca4-b68d-a3bc2dba87f4, features postgres-enterprise, plan postgres-enterprise, product line postgres, tier enterprise, expires 2026-09-23 03:44:53Z (27 days remaining), certificate SHA-256 D2:3D:90:71:6F:8C:87:0C:AE:9B:...
```

The serial and SHA-256 fingerprint identify precisely which certificate is in
use — handy for confirming a renewal, and for support conversations.

**The database can describe its own license.** The build ships a reporting
extension:

```bash
$ kubectl exec -it -n demo pg-enterprise-0 -c postgres -- psql -U postgres -c 'CREATE EXTENSION IF NOT EXISTS appscode_license;'
CREATE EXTENSION

$ kubectl exec -it -n demo pg-enterprise-0 -c postgres -- psql -U postgres -xc 'SELECT * FROM appscode_license_info();'
-[ RECORD 1 ]--+------------------------------------------------------------
license_id     | 1884894798739465099
cn             | e366e720-f748-4ca4-b68d-a3bc2dba87f4
product_line   | postgres
tier           | enterprise
plan           | postgres-enterprise
features       | {postgres-enterprise}
not_before     | 2026-08-24 03:44:53Z
not_after      | 2026-09-23 03:44:53Z
days_remaining | 27
verifies       | t
leaf_sha256    | D2:3D:90:71:6F:8C:87:0C:AE:9B:...
```

This is served from memory, so it is cheap to query and safe to scrape on a
schedule — `days_remaining` in particular is the natural thing to alert on.

## Step 6 — Connect and use it

The licensing is transparent to clients — this is ordinary PostgreSQL:

```bash
$ kubectl get secret -n demo pg-enterprise-auth -o jsonpath='{.data.password}' | base64 -d
```

```bash
$ kubectl exec -it -n demo pg-enterprise-0 -c postgres -- psql -U postgres
postgres=# CREATE TABLE tenants (id serial PRIMARY KEY, name text);
CREATE TABLE
postgres=# INSERT INTO tenants (name) VALUES ('acme'), ('globex');
INSERT 0 2
postgres=# SELECT * FROM tenants;
 id |  name
----+--------
  1 | acme
  2 | globex
(2 rows)
```

## Where the license lives

KubeDB mounts the certificate read-only at a fixed path **outside** the data
directory, and points the server at it. Two useful consequences:

- The license is not part of the data volume, so it stays out of snapshots and
  backups of `PGDATA`.
- Because of that, a database restored from a backup needs its own
  `spec.license` — it does not inherit one from the backup.

## Running a highly available cluster

Nothing about licensing changes for HA; set `spec.replicas` and KubeDB mounts
the same Secret on every replica.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-enterprise-ha
  namespace: demo
spec:
  version: "16.9-appscode"
  replicas: 3
  standbyMode: Hot
  license:
    secretRef:
      name: pg-enterprise-license
      key: license.pem
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

Each replica verifies the license independently at startup, and again whenever
it restarts or is rebuilt during a failover. A single Secret covers the whole
cluster.

## Using the extension-enabled images

Licensing is identical for the `-ext` images — swap `spec.version` for
`16.9-appscode-ext` (or `16.9-appscode-bookworm-ext`) and keep the same
`spec.license` block unchanged. What differs is the bundled extensions.

`pgvector`, `PostGIS` and `pg_repack` are ready to use immediately:

```bash
$ kubectl exec -n demo <your-ext-db>-0 -c postgres -- psql -U postgres -c 'CREATE EXTENSION vector;'
CREATE EXTENSION
```

`pg_cron` and `pgaudit` are background-worker extensions: the library has to be
in `shared_preload_libraries` before the extension can be created, which means
supplying a `spec.configSecret`. Until it is, they say so directly:

```
ERROR:  pgaudit must be loaded via shared_preload_libraries
ERROR:  unrecognized configuration parameter "cron.database_name"
```

See the [extensions guide](/docs/guides/postgres/custom-versions/extensions.md)
for how to supply that configuration.

## Renewing the license

Licenses are issued for a fixed term, so plan a renewal before the current
certificate runs out. `appscode_license_info()` reports how long you have, and is
the right thing to put behind an alert:

```bash
$ kubectl exec -n demo pg-enterprise-0 -c postgres -- psql -U postgres -tAc \
    'SELECT not_after, days_remaining FROM appscode_license_info();'
2026-09-23 03:44:53Z|27
```

Within 30 days of expiry the server also notes it in the log at startup:

```
WARNING:  license for Postgres Enterprise by AppsCode expires in 27 days
```

**Renewal is in place and online.** The server re-reads the license file
periodically rather than only trusting what it read at boot, so replacing the
certificate is picked up on its own — no restart, no downtime, no ops request:

```bash
$ kubectl patch secret pg-enterprise-license -n demo --type=merge \
    -p "{\"data\":{\"license.pem\":\"$(base64 -w0 license-renewed.pem)\"}}"
secret/pg-enterprise-license patched
```

Check the replacement is the certificate you expect before you apply it —
`openssl x509 -in license-renewed.pem -noout -dates` — since the server will
start using it without being asked to.

The update reaches the pod on the kubelet's normal refresh cycle, so allow a
couple of minutes. The database serving normally after that is your confirmation
that the new certificate was accepted; if you want to see its serial and
fingerprint recorded explicitly, restart the pod and read the startup line:

```bash
$ kubectl delete pod -n demo pg-enterprise-0
$ kubectl logs -n demo pg-enterprise-0 -c postgres | grep 'license accepted'
```

### If a renewal is rejected

An unusable certificate is not silently ignored — the server stops rather than
run unlicensed, and the `Postgres` object goes `NotReady` with the reason in the
`postgres` container log. Note that the pod itself keeps reporting `1/1 Running`,
so check `kubectl get pg` rather than the pod. Restore a valid certificate and
delete the pod:

```bash
$ kubectl delete pod -n demo pg-enterprise-0
```

The database is back about 45 seconds later, with data intact — the stop is a
clean shutdown, not a crash.

## Troubleshooting

License problems surface in one of three places. Knowing which one narrows the
search immediately:

| Symptom | Look at |
|---|---|
| `kubectl apply` returns an error, no object created | the message it printed |
| Pod stuck in `Init:0/1` | the pod's events |
| Pod `Running` but `Postgres` is `NotReady` | the `postgres` container log |

### The Postgres object is rejected

KubeDB validates the license reference up front, so a mistake here is caught
before anything is provisioned. Requesting a licensed version without a license:

```
admission webhook "postgreswebhook.validators.kubedb.com" denied the request:
PostgresVersion "16.9-appscode" requires spec.license (a licensed AppsCode Postgres Enterprise build);
set spec.license.secretRef to a Secret containing the license certificate under key "license.pem"
```

And the reverse — setting `spec.license` on a version that does not take one:

```
admission webhook "postgreswebhook.validators.kubedb.com" denied the request:
spec.license is set but PostgresVersion "16.9" is not a licensed distribution (spec.license.required is not set)
```

### The pod stays in `Init:0/1`

The Secret named in `spec.license` cannot be mounted yet. Check the pod's events:

```bash
$ kubectl describe pod -n demo pg-enterprise-0 | grep -A2 MountVolume
```

Either the Secret does not exist:

```
MountVolume.SetUp failed for volume "pg-license" : secret "pg-enterprise-license" not found
```

or its key does not match `spec.license.secretRef.key`:

```
MountVolume.SetUp failed for volume "pg-license" : references non-existent secret key: license.pem
```

Compare the two:

```bash
$ kubectl get pg -n demo pg-enterprise -o jsonpath='{.spec.license.secretRef}'
{"key":"license.pem","name":"pg-enterprise-license"}

$ kubectl get secret -n demo pg-enterprise-license -o go-template='{{range $k,$v := .data}}{{$k}}{{"\n"}}{{end}}'
license.pem
```

Both cases recover on their own: create or correct the Secret and the mount
completes on the next retry, with no need to delete the pod.

### The certificate is not accepted

The server reports the reason in the `postgres` container log, with a `DETAIL`
line naming the specific failure:

```bash
$ kubectl logs -n demo pg-enterprise-0 -c postgres | grep -iE 'license|FATAL'
```

| Log message | Meaning |
|---|---|
| `contains no PEM certificate` (`failure: no-pem`) | The Secret holds something other than the certificate — often the email or page it arrived in. |
| `chain does not verify against the AppsCode license CA` (`failure: chain-invalid`) | A well-formed certificate that AppsCode did not issue. |
| expiry named in the message | The term has ended; renew it. |

In each case, confirm what you have locally, put the right certificate in the
Secret, and delete the pod:

```bash
$ openssl x509 -in license.pem -noout -subject -dates
```

If you believe the certificate is correct and it is still refused, send the `CN`
from that output to AppsCode support.

## Cleaning up

```bash
$ kubectl delete pg -n demo pg-enterprise pg-enterprise-ha
$ kubectl delete secret -n demo pg-enterprise-license
$ kubectl delete ns demo
```

`deletionPolicy: WipeOut` was used above so the PVCs go with the object. With
the default `Delete` or with `Halt`, remove leftover PVCs yourself.

## Next Steps

- [Postgres object](/docs/guides/postgres/concepts/postgres.md) reference,
  including [`spec.license`](/docs/guides/postgres/concepts/postgres.md#speclicense).
- [PostgresVersion](/docs/guides/postgres/concepts/catalog.md) reference,
  including `spec.distribution` and `spec.license.required`.
- [Run a highly available cluster](/docs/guides/postgres/clustering/ha_cluster.md).
- [Monitor your database](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- [Backup and restore](/docs/guides/postgres/backup/kubestash/overview/index.md).
