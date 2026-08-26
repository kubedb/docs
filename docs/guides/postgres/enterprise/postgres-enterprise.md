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

**Postgres Enterprise by AppsCode** is a PostgreSQL distribution built and
supported by AppsCode. Unlike the community images, its `postgres` binary
verifies a license certificate at startup and refuses to run without a valid
one. This guide walks through obtaining that certificate, giving it to KubeDB,
and running a licensed database end to end.

{{< notice type="warning" message="Two different licenses are involved and they are not interchangeable. The **KubeDB platform license** (passed as `global.license` when you install the KubeDB operator) unlocks KubeDB features. The **database license** described on this page is consumed by the PostgreSQL binary itself. A cluster can have a perfectly valid KubeDB license and still fail to start a Postgres Enterprise database because the database license is missing." >}}

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

You can inspect what you were issued before using it. The certificate's
`notAfter` date is when the database will stop accepting it:

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

`spec.license.secretRef.key` may be omitted, in which case KubeDB's mutating
webhook fills in `license.pem`. Spelling it out is still the safer habit: the
CRD schema itself marks the field required, so a path that bypasses the
mutating webhook rejects the object rather than defaulting it.

## Step 5 — Confirm the license was accepted

Three independent checks, in increasing order of detail.

**The database reports the enterprise build.** A community image would answer
`PostgreSQL 16.9 ...` here:

```bash
$ kubectl exec -it -n demo pg-enterprise-0 -c postgres -- psql -U postgres -tAc 'SELECT version();'
Postgres Enterprise by AppsCode 16.9 on x86_64-pc-linux-musl, compiled by gcc (Alpine 14.2.0) 14.2.0, 64-bit
```

The rebranding is cosmetic and deliberately does not touch the numeric version,
so extensions and client drivers that gate on it behave exactly as on community
PostgreSQL:

```bash
$ kubectl exec -it -n demo pg-enterprise-0 -c postgres -- psql -U postgres -tAc 'SHOW server_version; SHOW server_version_num;'
16.9
160009
```

**The server logged its acceptance.** One line is emitted per verification. A
first boot logs two — one for the `initdb` bootstrap phase, which is itself
license-checked, and one for the server proper:

```bash
$ kubectl logs -n demo pg-enterprise-0 -c postgres | grep -i 'license accepted'
2026-08-26 06:04:25.059 UTC [64] LOG:  license accepted: id (serial) 1884894798739465099, CN e366e720-f748-4ca4-b68d-a3bc2dba87f4, features postgres-enterprise, plan postgres-enterprise, product line postgres, tier enterprise, expires 2026-09-23 03:44:53Z (27 days remaining), certificate SHA-256 D2:3D:90:71:6F:8C:87:0C:AE:9B:...
```

The `id (serial)` and the SHA-256 fingerprint identify exactly which certificate
was accepted, which is how you confirm a renewal actually took effect rather
than the old file still being in place.

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

This reads a snapshot the postmaster placed in shared memory when it verified
the certificate, so it costs nothing to query and cannot disagree with what the
server actually accepted at startup.

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

## Where the license lives in the pod

Worth knowing when debugging. KubeDB projects the Secret to a fixed path
**outside** the data directory and points the server at it with the `PGLICENSE`
environment variable:

```bash
$ kubectl exec -n demo pg-enterprise-0 -c postgres -- printenv PGLICENSE
/etc/pg-license/license.pem

$ kubectl exec -n demo pg-enterprise-0 -c postgres -- ls -lL /etc/pg-license/license.pem
-r--r-----    1 root     postgres      1375 Aug 26 06:04 /etc/pg-license/license.pem
```

The certificate is mounted read-only, owned by `root` and readable by the
`postgres` group. Note the `-L`: like every Kubernetes Secret volume, the
visible entry is a symlink into a `..data` directory, so a plain `ls -l` shows
the link rather than the file's own mode.

Two consequences follow from the path being outside `PGDATA`:

- The license is **not** part of the data volume, so it is absent from
  snapshots and backups of `PGDATA` and cannot leak through them.
- Restoring a backup does not carry a license with it. The restored `Postgres`
  object needs its own `spec.license`.

Only the `postgres` container gets the license. The init container and the
`pg-coordinator` sidecar have neither the mount nor the variable — the
coordinator drives `pg_rewind` and `pg_ctl` by executing them *inside* the
`postgres` container, so it inherits the license from there.

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
supplying a `spec.configSecret`. Without it they refuse clearly rather than
failing obscurely:

```
ERROR:  pgaudit must be loaded via shared_preload_libraries
ERROR:  unrecognized configuration parameter "cron.database_name"
```

See the [extensions guide](/docs/guides/postgres/custom-versions/extensions.md)
for how to supply that configuration.

## Renewing the license

Licenses expire. Within 30 days of expiry, each verification emits a warning
next to the acceptance line:

```bash
$ kubectl logs -n demo pg-enterprise-0 -c postgres | grep -i license
2026-08-26 06:04:25.059 UTC [64] LOG:  license accepted: id (serial) 1884894798739465099, CN e366e720-f748-4ca4-b68d-a3bc2dba87f4, features postgres-enterprise, plan postgres-enterprise, product line postgres, tier enterprise, expires 2026-09-23 03:44:53Z (27 days remaining), certificate SHA-256 D2:3D:90:71:6F:8C:87:0C:AE:9B:...
2026-08-26 06:04:25.087 UTC [68] WARNING:  license for Postgres Enterprise by AppsCode expires in 27 days
```

Do not rely on that warning to remind you. Verification happens at startup, and
the periodic re-checks that follow are silent while they keep succeeding — so a
database that has been up for weeks logs the warning exactly once, at the boot
where it first applied. `days_remaining` from `appscode_license_info()` carries
the same number, is queryable at any time, and is the right thing to alert on:

```bash
$ kubectl exec -n demo pg-enterprise-0 -c postgres -- psql -U postgres -tAc \
    'SELECT days_remaining FROM appscode_license_info();'
27
```

The server does not simply trust the certificate it validated at startup. A
background worker **re-reads the license file from disk** roughly once a minute
and re-verifies it. That is what makes an in-place renewal possible: replace the
file and the next re-check picks it up, with no restart and no downtime.

{{< notice type="danger" message="The same re-check is what makes a botched renewal an immediate outage. If the replacement certificate is invalid for any reason, the next re-check fails and the server requests a **fast shutdown** within about a minute. Verify the new certificate locally with `openssl x509 -in license-renewed.pem -noout -dates` **before** putting it into the Secret." >}}

To renew, replace the certificate in the Secret:

```bash
$ kubectl patch secret pg-enterprise-license -n demo --type=merge \
    -p "{\"data\":{\"license.pem\":\"$(base64 -w0 license-renewed.pem)\"}}"
secret/pg-enterprise-license patched
```

`kubectl patch` is used rather than `kubectl apply` only because applying over a
Secret created with `kubectl create` prints a spurious
`missing the kubectl.kubernetes.io/last-applied-configuration annotation`
warning. Either works.

A Secret update reaches the pod's filesystem on the kubelet's own refresh cycle,
not instantly — allow a couple of minutes end to end.

Verifying that a renewal succeeded is less direct than you might expect, because
**a successful periodic re-check logs nothing**. Only failures are logged; a
server that has been up for hours still shows just the one or two
`license accepted` lines from its startup. So there is no positive log line to
wait for.

What that leaves is a reliable negative: an invalid certificate shuts the server
down within about a minute, so a database still serving several minutes after the
Secret changed has accepted the new file.

```bash
$ kubectl get pg -n demo pg-enterprise
NAME            VERSION         STATUS   AGE
pg-enterprise   16.9-appscode   Ready    3h
```

If you want positive confirmation of *which* certificate is in force — after a
renewal window closes, say — restart the pod and read the fresh startup line,
which reports the serial and fingerprint of the file as it is now:

```bash
$ kubectl delete pod -n demo pg-enterprise-0
$ kubectl logs -n demo pg-enterprise-0 -c postgres | grep 'license accepted'
```

### Recovering from a failed renewal

A license-triggered shutdown does **not** self-heal, and this is the single most
confusing failure in this guide, because Kubernetes reports the pod as healthy:

```bash
$ kubectl get pod -n demo
NAME              READY   STATUS    RESTARTS   AGE
pg-enterprise-0   1/1     Running   0          4m

$ kubectl get pg -n demo
NAME            VERSION         STATUS     AGE
pg-enterprise   16.9-appscode   NotReady   4m
```

The container's entrypoint stays alive after the PostgreSQL server shuts down, so
the container never exits, never restarts, and continues to report `1/1 Running`
with a restart count of `0`. Only the `Postgres` object's `NotReady` phase and
the container log tell the truth:

```
LOG:  license verification failed during periodic re-check: license file
      "/etc/pg-license/license.pem" contains no PEM certificate; requesting fast shutdown
LOG:  received fast shutdown request
LOG:  database system is shut down
```

Putting a valid certificate back is necessary but **not sufficient** — the server
is not restarted by the entrypoint. Fix the Secret, then delete the pod:

```bash
$ kubectl delete pod -n demo pg-enterprise-0
pod "pg-enterprise-0" deleted
```

The database returns to `Ready` about 45 seconds later. Data is untouched: the
shutdown was a clean fast shutdown with a completed checkpoint, not a crash.

## Troubleshooting

Licensing failures land in one of three places, and knowing which saves time:

| Symptom | Where it failed |
|---|---|
| `kubectl apply` returns an error, no object created | admission webhook |
| Pod stuck in `Init:0/1` | kubelet, building the Secret volume |
| Pod `Running` but `Postgres` is `NotReady` | the PostgreSQL server itself |

For the third case, the `postgres` container log is the only place the reason
appears:

```bash
$ kubectl logs -n demo pg-enterprise-0 -c postgres | grep -iE 'license|FATAL'
```

### The Postgres object is rejected outright

Requesting a licensed version without a license:

```
Error from server (Forbidden): admission webhook "postgreswebhook.validators.kubedb.com" denied the request:
PostgresVersion "16.9-appscode" requires spec.license (a licensed AppsCode Postgres Enterprise build);
set spec.license.secretRef to a Secret containing the license certificate under key "license.pem"
```

KubeDB refuses rather than creating a database that could only crash-loop. The
mirror image is refused too — pointing `spec.license` at a community version that
does not need one:

```
Error from server (Forbidden): admission webhook "postgreswebhook.validators.kubedb.com" denied the request:
spec.license is set but PostgresVersion "16.9" is not a licensed distribution (spec.license.required is not set)
```

### The pod stays in `Init:0/1`

The kubelet cannot build the license volume, so no container starts. Check the
pod's events:

```bash
$ kubectl describe pod -n demo pg-enterprise-0 | grep -A2 MountVolume
```

Either the Secret is absent:

```
MountVolume.SetUp failed for volume "pg-license" : secret "pg-enterprise-license" not found
```

or the key inside it does not match `spec.license.secretRef.key`:

```
MountVolume.SetUp failed for volume "pg-license" : references non-existent secret key: license.pem
```

Compare the two directly:

```bash
$ kubectl get pg -n demo pg-enterprise -o jsonpath='{.spec.license.secretRef}'
{"key":"license.pem","name":"pg-enterprise-license"}

$ kubectl get secret -n demo pg-enterprise-license -o go-template='{{range $k,$v := .data}}{{$k}}{{"\n"}}{{end}}'
license.pem
```

Both failures are self-healing: fix the Secret and the kubelet completes the
mount on its next retry, with no need to delete the pod.

### The license file is not a certificate

```
FATAL:  license file "/etc/pg-license/license.pem" contains no PEM certificate
DETAIL:  License path: "/etc/pg-license/license.pem"; mode: single-user; failure: no-pem.
```

The Secret key holds something that is not a PEM certificate — commonly the
wrapper the license arrived in (an email body, an HTML page) rather than the
certificate itself. Check locally before loading it:

```bash
$ openssl x509 -in license.pem -noout -subject
```

### The certificate was not issued by AppsCode

```
FATAL:  license certificate chain does not verify against the AppsCode license CA: self-signed certificate
DETAIL:  License path: "/etc/pg-license/license.pem"; mode: single-user; failure: chain-invalid.
```

The file is a well-formed certificate but does not chain to the AppsCode
licensing CA. Verification is by cryptographic signature, so a self-signed
certificate that merely copies the AppsCode subject or issuer names fails here
just the same. The text after the colon varies with how the chain broke
(`self-signed certificate`, `unable to get local issuer certificate`); the
`failure: chain-invalid` detail is the stable marker. Request a real license.

### The certificate has expired

The `FATAL` message names expiry as the reason, and the `DETAIL` line carries its
own `failure:` marker, in the same shape as the two cases above. Renew the
certificate as described above. A database whose license expires while it is
running does not keep running: the periodic re-check fails and the server shuts
itself down. Alert on `days_remaining` and renew ahead of time.

### `"/var/pv/data" is not a valid data directory`

Seen immediately after one of the `FATAL` messages above, on a **brand new**
database. The license check also covers the single-user phase of `initdb`
(`mode: single-user` in the `DETAIL` line), so a first-boot license failure
aborts bootstrap and leaves a partially initialised data directory behind. It is
a consequence of the license failure, not a separate problem — fix the license
and delete the pod; `initdb` runs again from scratch.

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
