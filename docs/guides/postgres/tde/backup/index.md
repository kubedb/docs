---
title: Backup & Restore a TDE-Encrypted Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-tde-backup
    name: Backup & Restore
    parent: guides-postgres-tde
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start with the [KubeDB documentation](/docs/README.md).

# Backup & Restore a TDE-Encrypted Postgres

A TDE-encrypted Postgres cannot be backed up or restored with the community
backup/restore path: the physical backup tool (`pg_basebackup`) and the WAL
reader used for continuous archiving cannot read `pg_tde`-encrypted files or
its custom WAL records. KubeDB and [KubeStash](https://kubestash.com) ship
`pg_tde`-aware equivalents for both the logical and physical paths, so backup
and restore work the same way you would expect, once you point at the right
catalog entries.

Read the [TDE overview](/docs/guides/postgres/tde/overview/index.md) and
[TDE guide](/docs/guides/postgres/tde/guide/index.md) first.

## Logical backup & restore (KubeStash)

[KubeStash logical backup](/docs/guides/postgres/backup/kubestash/logical/index.md)
uses `pg_dump`/`pg_dumpall` to back up and `psql` to restore, which talk to
Postgres over the normal protocol and never touch the on-disk files directly,
so a TDE cluster backs up exactly like a community one -- no extra
configuration needed on the
`BackupConfiguration`/`RestoreSession` side. The `postgres-addon` used by
KubeStash detects the Percona distribution automatically.

The one thing to plan for is the **restore target**:

- Restoring a TDE-encrypted dump into another TDE (Percona) cluster works as
  expected; data is re-encrypted under the target cluster's own principal key.
- Restoring a TDE-encrypted dump into a **non-TDE (community) Postgres** is
  rejected: the dump contains `pg_tde`/`tde_heap` DDL that a community server
  cannot execute. The restore addon detects this up front and fails loudly
  instead of partially applying the dump.

## Physical backup, continuous archiving & PITR

Follow the [continuous archiving and PITR guide](/docs/guides/postgres/pitr/archiver.md)
for the general `PostgresArchiver` setup (`BackupStorage`, `RetentionPolicy`,
`PostgresArchiver`, `spec.archiver.ref` on the `Postgres` object). For a TDE
cluster, the differences are:

- **Full/base backups** use `pg_tde_basebackup` instead of `pg_basebackup`
  (selected automatically whenever the database image is a Percona
  distribution -- no field to set).
- **WAL archiving and PITR recovery** need a `pg_tde`-aware archiver image,
  because reading the commit LSN out of the WAL stream (for archiving) and
  replaying it (for recovery) both require registering `pg_tde`'s custom WAL
  resource manager, and -- when `spec.tde.encryptWAL: true` -- decrypting the
  WAL itself. Use a `PostgresVersion` whose `spec.archiver.walg.image` points
  at the Percona build of the archiver image (matching the `-percona` catalog
  entries used elsewhere in TDE, e.g. `17.9-percona`), not the community
  build. The KubeDB installer ships the correct image per version; you only
  need to make sure you deployed the `-percona` `PostgresVersion`, the same
  one used to create the database.
- This applies whether or not `spec.tde.encryptWAL` is enabled: even with WAL
  encryption off, `pg_tde` still writes custom WAL records that a community
  WAL reader does not recognize.

With the Percona archiver image in place, PITR works the same for both
`encryptWAL: false` and `encryptWAL: true` clusters. Either way, the original
cluster's principal key must still be resolvable through its key provider
(Vault, KMIP, or file) at restore time: decrypting the archived WAL and the
base backup's `pg_tde` internal keys both depend on it, regardless of whether
WAL encryption itself was on. Restore a `Postgres` with
`spec.init.archiver.recoveryTimestamp` exactly as shown in the
[PITR guide](/docs/guides/postgres/pitr/archiver.md#restore-postgresql).

## Pulling the Percona TDE images

The Percona Server for PostgreSQL image (and, if your environment mirrors
images through a private registry, the matching archiver/backup plugin
images) may not be publicly pullable in every environment. If your registry
requires credentials, create a `Secret` of type
`kubernetes.io/dockerconfigjson` and reference it via
`spec.podTemplate.spec.imagePullSecrets` on the `Postgres` object. WAL archiving
runs as part of the same Postgres Pod, so it reuses that secret too and needs no
separate configuration. KubeStash's own `BackupConfiguration`
and `RestoreSession` Jobs are created by the KubeStash operator, not the Postgres
operator, and do **not** inherit `Postgres.spec.podTemplate.spec.imagePullSecrets`.
If your registry requires credentials for the `postgres-addon` image used by
those Jobs, set `imagePullSecrets` on their own `jobTemplate.template.spec`
separately:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: tde-postgres
  namespace: demo
spec:
  version: "17.9-percona"
  podTemplate:
    spec:
      imagePullSecrets:
      - name: image-pull-secret
  # ...
```

See [using a private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md)
for how KubeDB propagates `imagePullSecrets` to the database Pods themselves.
For KubeStash `BackupConfiguration`/`RestoreSession` Jobs, configure
`imagePullSecrets` on each object's own `jobTemplate.template.spec` instead, as
described above.

## Next Steps

- Review the [TDE overview](/docs/guides/postgres/tde/overview/index.md) for
  the key hierarchy and limitations.
- Follow the [TDE guide](/docs/guides/postgres/tde/guide/index.md) to deploy
  an encrypted Postgres, rotate the principal key, and enable WAL encryption.
- Learn about [logical backup & restore with KubeStash](/docs/guides/postgres/backup/kubestash/logical/index.md).
- Learn about [continuous archiving and point-in-time recovery](/docs/guides/postgres/pitr/archiver.md).
