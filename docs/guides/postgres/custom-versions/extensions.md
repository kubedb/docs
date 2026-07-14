---
title: PostgreSQL Extensions (Extension-enabled Versions)
menu:
  docs_{{ .version }}:
    identifier: pg-extensions-custom-versions
    name: Extensions
    parent: pg-custom-versions-postgres
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL Extensions

KubeDB ships a set of **extension-enabled** PostgreSQL images (the `-ext` versions). These are the
same official PostgreSQL images, but with a curated set of popular extensions compiled in, so you
can enable them with a plain `CREATE EXTENSION` — no custom image build required.

Every `-ext` image bundles the following extensions:

| Extension | `CREATE EXTENSION` name | Version | Needs `shared_preload_libraries`? | What it is |
|---|---|---|---|---|
| pgvector | `vector` | 0.8.2 | No | Vector similarity search (embeddings) |
| PostGIS | `postgis` | 3.6.2 | No | Geospatial types & functions |
| pg_repack | `pg_repack` | 1.5.3 | No | Rebuild tables/indexes without long locks |
| pg_cron | `pg_cron` | 1.6 | **Yes** | In-database cron job scheduler |
| pgaudit | `pgaudit` | tracks PG major | **Yes** (recommended) | Session/object audit logging |
| pg_stat_statements | `pg_stat_statements` | 1.12 | **Yes** (already preloaded) | SQL execution statistics |

> **Availability:** The `-ext` PostgresVersions are shipped by default from **KubeDB `v2026.7.10`**
> onwards. If you are on an earlier release, you can still use them by creating the PostgresVersion
> objects by hand — see [Using extensions on older KubeDB releases](#using-extensions-on-older-kubedb-releases).

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

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres/extensions](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres/extensions) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Available extension-enabled versions

List the `-ext` PostgresVersions available in your cluster:

```bash
$ kubectl get postgresversions | grep -E 'NAME|ext'
NAME                      VERSION   DISTRIBUTION   DB_IMAGE                                                                DEPRECATED   AGE
16.13-bookworm-ext        16.13     KubeDB         ghcr.io/appscode-images/postgres:16.13-bookworm-ext                                  10h
16.13-ext                 16.13     KubeDB         ghcr.io/appscode-images/postgres:16.13-alpine-ext                                    10h
17.9-bookworm-ext         17.9      KubeDB         ghcr.io/appscode-images/postgres:17.9-bookworm-ext                                   10h
17.9-ext                  17.9      KubeDB         ghcr.io/appscode-images/postgres:17.9-alpine-ext                                     10h
18.3-bookworm-ext         18.3      KubeDB         ghcr.io/appscode-images/postgres:18.3-bookworm-ext                                   10h
18.3-ext                  18.3      KubeDB         ghcr.io/appscode-images/postgres:18.3-alpine-ext                                     10h
```

The `-ext` versions come in two base OS flavours: `*-ext` (Alpine) and `*-bookworm-ext` (Debian
Bookworm). Pick whichever matches your fleet; the bundled extensions are identical.

## Choose the extensions you need

You do **not** have to use every extension. Two things decide what you must do:

1. **Extensions that need to be preloaded** (`pg_cron`, `pgaudit`, `pg_stat_statements`) must be listed
   in `shared_preload_libraries`. KubeDB already preloads `pg_stat_statements` by default, so you only
   need a custom configuration if you want `pg_cron` and/or `pgaudit`.
2. **Extensions that do NOT need preloading** (`pgvector`, `PostGIS`, `pg_repack`) work with just
   `CREATE EXTENSION` — for these you can skip the configuration Secret entirely.

So:

- **Only want pgvector / PostGIS / pg_repack?** Skip to [Deploy PostgreSQL](#deploy-postgresql) and
  drop the `configSecret` line — no Secret needed.
- **Want pg_cron and/or pgaudit?** Create the configuration Secret below, keeping only the libraries
  you actually want (always keep `pg_stat_statements`, it is KubeDB's default).

## Create the configuration Secret

KubeDB applies custom PostgreSQL configuration through a Secret containing a `user.conf` file (see
[Using Custom Configuration File](/docs/guides/postgres/configuration/using-config-file.md) for details).
Here we use it to append `pg_cron` and `pgaudit` to `shared_preload_libraries`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/extensions/pg-extensions-config.yaml
secret/pg-extensions-config created
```

Below is the Secret we just created:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pg-extensions-config
  namespace: demo
stringData:
  # KubeDB preloads pg_stat_statements by default. Keep it and append only the
  # extensions you actually want that REQUIRE preloading (pg_cron, pgaudit).
  # pgvector, PostGIS and pg_repack do NOT need to be listed here.
  user.conf: |-
    shared_preload_libraries='pg_stat_statements,pg_cron,pgaudit'
    cron.database_name='postgres'
```

> **Tailoring the list:** drop `pg_cron` if you don't want the scheduler (and remove the
> `cron.database_name` line too), or drop `pgaudit` if you don't need auditing. Just never remove
> `pg_stat_statements` — it is enabled by default and other tooling relies on it.
> `cron.database_name` tells the `pg_cron` background worker which database to run jobs in.

## Deploy PostgreSQL

Now create a Postgres object that uses an `-ext` version and references the Secret via
`spec.configSecret`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/extensions/pg-extensions.yaml
postgres.kubedb.com/pg-extensions created
```

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-extensions
  namespace: demo
spec:
  version: "18.3-ext" # an extension-enabled PostgresVersion
  replicas: 1
  configSecret:
    name: pg-extensions-config # omit this if you only need pgvector / PostGIS / pg_repack
  storage:
    storageClassName: "standard" # change to your cluster's StorageClass
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Wait for the database to become `Ready`:

```bash
$ kubectl get pg -n demo pg-extensions
NAME            VERSION    STATUS   AGE
pg-extensions   18.3-ext   Ready    29s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=pg-extensions
NAME              READY   STATUS    RESTARTS   AGE
pg-extensions-0   1/1     Running   0          24s
```

Confirm the preloaded libraries were applied:

```bash
$ kubectl exec -it -n demo pg-extensions-0 -c postgres -- psql -U postgres -c "SHOW shared_preload_libraries;"
      shared_preload_libraries
------------------------------------
 pg_stat_statements,pg_cron,pgaudit
(1 row)
```

## Enable and use the extensions

Open a `psql` session on the primary pod:

```bash
$ kubectl exec -it -n demo pg-extensions-0 -c postgres -- psql -U postgres
psql (18.3)
Type "help" for help.

postgres=#
```

Create only the extensions you want. The commands below are independent — run just the ones you need.

```sql
postgres=# CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION
postgres=# CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION
postgres=# CREATE EXTENSION IF NOT EXISTS pg_repack;
CREATE EXTENSION
postgres=# CREATE EXTENSION IF NOT EXISTS pg_cron;
CREATE EXTENSION
postgres=# CREATE EXTENSION IF NOT EXISTS pgaudit;
CREATE EXTENSION
postgres=# CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION
```

Verify what is installed with `\dx`:

```sql
postgres=# \dx
                                                     List of installed extensions
        Name        | Version | Default version |   Schema   |                              Description
--------------------+---------+-----------------+------------+------------------------------------------------------------------------
 pg_cron            | 1.6     | 1.6             | pg_catalog | Job scheduler for PostgreSQL
 pg_repack          | 1.5.3   | 1.5.3           | public     | Reorganize tables in PostgreSQL databases with minimal locks
 pg_stat_statements | 1.12    | 1.12            | public     | track planning and execution statistics of all SQL statements executed
 pgaudit            | 18.0    | 18.0            | public     | provides auditing functionality
 plpgsql            | 1.0     | 1.0             | pg_catalog | PL/pgSQL procedural language
 postgis            | 3.6.2   | 3.6.2           | public     | PostGIS geometry and geography spatial types and functions
 vector             | 0.8.2   | 0.8.2           | public     | vector data type and ivfflat and hnsw access methods
(7 rows)
```

### pgvector — vector similarity search

```sql
postgres=# CREATE TABLE items (id bigserial PRIMARY KEY, embedding vector(3));
CREATE TABLE
postgres=# INSERT INTO items (embedding) VALUES ('[1,2,3]'), ('[4,5,6]'), ('[7,8,9]');
INSERT 0 3
-- order rows by (L2) distance to a query vector
postgres=# SELECT id, embedding FROM items ORDER BY embedding <-> '[3,1,2]' LIMIT 3;
 id | embedding
----+-----------
  1 | [1,2,3]
  2 | [4,5,6]
  3 | [7,8,9]
(3 rows)
```

### PostGIS — geospatial data

```sql
postgres=# CREATE TABLE cities (id bigserial PRIMARY KEY, name text, geom geometry(Point,4326));
CREATE TABLE
postgres=# INSERT INTO cities (name, geom) VALUES
  ('Dhaka',    ST_SetSRID(ST_MakePoint(90.4125, 23.8103), 4326)),
  ('New York', ST_SetSRID(ST_MakePoint(-73.9857, 40.7484), 4326));
INSERT 0 2
-- great-circle distance between the two cities, in kilometres
postgres=# SELECT a.name, b.name, round(ST_DistanceSphere(a.geom, b.geom)/1000) AS km
           FROM cities a, cities b WHERE a.name='Dhaka' AND b.name='New York';
 name  |   name   |  km
-------+----------+-------
 Dhaka | New York | 12658
(1 row)
```

### pg_cron — scheduled jobs

`pg_cron` requires the `shared_preload_libraries` entry (added above). Schedule a job and inspect the
`cron.job` catalog:

```sql
postgres=# SELECT cron.schedule('nightly-vacuum', '0 3 * * *', 'VACUUM;');
 schedule
----------
        1
(1 row)

postgres=# SELECT jobid, schedule, command, jobname FROM cron.job;
 jobid | schedule  | command |    jobname
-------+-----------+---------+----------------
     1 | 0 3 * * * | VACUUM; | nightly-vacuum
(1 row)
```

### pgaudit — audit logging

`pgaudit` also requires preloading. Session-level `SET` only lasts for one connection, so persist the
setting on the database (or put `pgaudit.log` directly in `user.conf`):

```sql
postgres=# ALTER DATABASE postgres SET pgaudit.log = 'ddl, write';
ALTER DATABASE
```

Reconnect and confirm, then run an audited statement:

```sql
postgres=# SHOW pgaudit.log;
 pgaudit.log
-------------
 ddl, write
(1 row)

postgres=# CREATE TABLE audit_demo (id int);
CREATE TABLE
postgres=# INSERT INTO audit_demo VALUES (1);
INSERT 0 1
```

pgaudit writes to the PostgreSQL server log. Check the pod logs:

```bash
$ kubectl logs -n demo pg-extensions-0 -c postgres | grep 'AUDIT:'
... LOG:  AUDIT: SESSION,1,1,DDL,CREATE TABLE,TABLE,public.audit_demo,CREATE TABLE audit_demo (id int),<not logged>
... LOG:  AUDIT: SESSION,2,1,WRITE,INSERT,,,INSERT INTO audit_demo VALUES (1),<not logged>
```

### pg_stat_statements — query statistics

This one is preloaded out of the box, so `CREATE EXTENSION` is all you need:

```sql
postgres=# SELECT queryid, calls, left(query, 40) AS query
           FROM pg_stat_statements ORDER BY calls DESC LIMIT 5;
       queryid        | calls |                 query
----------------------+-------+----------------------------------------
 -3688696628780506391 |     6 | SELECT $1
 -8802593197440449731 |     6 | BEGIN READ WRITE
 -7676915344437841334 |     6 | ROLLBACK
  4357125790077000179 |     6 | SELECT now()
 -6953085582145547871 |     1 | CREATE EXTENSION IF NOT EXISTS pgaudit
(5 rows)
```

### pg_repack — online table reorganization

`pg_repack` ships both the extension (installed above) and a client binary that is available inside
the database pod. It needs the target table to have a primary key or a not-null unique key.

```sql
postgres=# CREATE TABLE bloat_demo AS SELECT g AS id, md5(g::text) AS val FROM generate_series(1,100000) g;
SELECT 100000
postgres=# ALTER TABLE bloat_demo ADD PRIMARY KEY (id);
ALTER TABLE
postgres=# \q
```

Run `pg_repack` from inside the pod:

```bash
$ kubectl exec -it -n demo pg-extensions-0 -c postgres -- pg_repack -U postgres -d postgres --table public.bloat_demo
INFO: repacking table "public.bloat_demo"
```

## Using extensions on older KubeDB releases

The `-ext` PostgresVersions are bundled from **KubeDB `v2026.7.10`**. On earlier releases they will not
exist:

```bash
$ kubectl get postgresversions | grep ext
# (no results)
```

Because the extensions live inside the database image, enabling them on an older release only requires
a PostgresVersion whose `spec.db.image` points to an `-ext` image. The safest way is to copy an existing
PostgresVersion of the same PostgreSQL version (so the coordinator/init/exporter images already match
your operator) and change **only** `metadata.name` and `spec.db.image`:

```bash
# start from the official 18.3 version already in your cluster
$ kubectl get postgresversion 18.3 -o yaml > 18.3-ext.yaml

# then edit 18.3-ext.yaml:
#   metadata.name:   18.3   ->   18.3-ext
#   spec.db.image:   ghcr.io/appscode-images/postgres:18.3-alpine
#                ->  ghcr.io/appscode-images/postgres:18.3-alpine-ext
# and remove the runtime fields (status, resourceVersion, uid, creationTimestamp, managedFields)

$ kubectl apply -f 18.3-ext.yaml
postgresversion.catalog.kubedb.com/18.3-ext created
```

Alternatively, ready-made sample PostgresVersion manifests for every `-ext` version are provided in the
[docs/examples/postgres/extensions](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres/extensions)
folder. Apply the one matching your PostgreSQL version, e.g.:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/extensions/postgresversion-18.3-ext.yaml
postgresversion.catalog.kubedb.com/18.3-ext created
```

The available sample files are:

```
postgresversion-16.13-ext.yaml            postgresversion-16.13-bookworm-ext.yaml
postgresversion-17.9-ext.yaml             postgresversion-17.9-bookworm-ext.yaml
postgresversion-18.3-ext.yaml             postgresversion-18.3-bookworm-ext.yaml
```

> The sidecar image tags (`coordinator`, `initContainer`, `exporter`, `courier`, `walg`) in the sample
> files are the ones released with KubeDB `v2026.7.10`. If you are on a different release, prefer the
> copy-and-edit approach above so those images stay in sync with your operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pg/pg-extensions
kubectl delete -n demo secret pg-extensions-config
kubectl delete ns demo
```

## Next Steps

- Learn how to [run PostgreSQL with a custom configuration file](/docs/guides/postgres/configuration/using-config-file.md).
- Set up a [Highly Available PostgreSQL cluster](/docs/guides/postgres/clustering/ha_cluster.md) — extensions created on the primary replicate to the standbys.
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Detail concepts of [PostgresVersion object](/docs/guides/postgres/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
```
