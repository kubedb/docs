---
title: Migration
menu:
  docs_{{ .version }}:
    identifier: migration
    name: Migration
    parent: operatormanual
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
url: /docs/{{ .version }}/operatormanual/migration/
aliases:
  - /docs/{{ .version }}/operatormanual/migration/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Migrate To KubeDB

KubeDB Migration lets you move an existing database — such as a MySQL instance running on AWS RDS or any external host — entirely into a KubeDB-managed database. The migration runs in the background while your source database continues to serve live traffic, and you only cut over once the streaming lag has dropped to a few megabytes — at which point stopping writes to the source drains the remaining lag to zero almost instantly.

## Why It Matters

- **No maintenance window** — the source database stays fully operational and accepting reads and writes before the cutover
- **Minimal downtime cutover** — you only stop writes to the source for the brief moment it takes for streaming lag to reach zero, then immediately redirect application endpoints to the new KubeDB-managed database.
- **Source stays untouched** — KubeDB never modifies the source database; you remain in control of when and whether to cut over.

## Setup

### Fresh Install

Add `--set kubedb-migrator.enabled=true` to the standard [KubeDB helm install](/docs/setup/install/kubedb.md) guide.

### Upgrade Existing Install

The Migrator CRD is required to apply the Migrator CR. Helm upgrade command doesnt apply CRD. So apply it manually:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubedb/apimachinery/refs/heads/master/crds/migrator.kubedb.com_migrators.yaml
```

Add `--set kubedb-migrator.enabled=true` to the [Kubedb helm upgrade](/docs/setup/upgrade/index.md) 


## Migration Steps

1. **Create a Migrator CR** with your source connection details and target KubeDB database reference. Migration starts automatically.
2. **Source stays live** — the KubeDB Migrator first copies the schema, then takes an initial bulk snapshot of your data. Your source database continues to accept reads and writes throughout.
3. **Streaming phase begins** — once the snapshot is complete, KubeDB streams ongoing changes from the source using CDC (change-data capture). 
4. **Stop writes to the source** — when the streaming lag approaches zero, stop source database write
5. **Wait for lag to reach zero** — once lag is exactly zero, the two databases are fully in sync.
6. **Switch endpoints** — update your application connection string to point to the new KubeDB-managed database.
7. **Downtime is minimal** — the only downtime is the window between steps 4 and 6, which is typically just a few minutes.

## Supported Database

The following database has migration support.

[PostgreSQL](/docs/guides/postgres/migration/databaseMigration.md)

[MySQL](/docs/guides/mysql/migration/databaseMigration.md)

[MariaDB](/docs/guides/mariadb/migration/databaseMigration.md)
