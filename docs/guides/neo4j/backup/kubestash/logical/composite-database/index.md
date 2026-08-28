---
title: Backup and Restore Neo4j Composite Databases and Aliases
description: Backup and restore Neo4j composite databases and aliases using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-neo4j-composite-database-logical-backup-stashv2
    name: Composite Databases and Aliases
    parent: guides-neo4j-logical-backup-stashv2
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore Neo4j Composite Databases and Aliases

This guide shows you how to back up and restore physical databases, composite databases, composite aliases, and standalone aliases with KubeStash. It also explains the separate credential step required to restore stored-native remote aliases.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl` to access it.
- Install [KubeDB](/docs/setup/README.md), [KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash), and the [KubeStash kubectl plugin](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- Use a KubeDB-supported Neo4j Enterprise version that supports composite databases. This guide uses `2025.12.1`.
- Prepare an S3-compatible bucket and its access credentials.
- Install a Java Development Kit that provides `keytool` on the machine where you run the setup commands.
- Make sure the source and restore-target Neo4j deployments can reach the remote Neo4j Bolt endpoint.
- Read the [Neo4j backup and restore overview](/docs/guides/neo4j/backup/kubestash/overview/index.md) if you are new to this process.

Create the namespace used throughout this guide:

```bash
$ kubectl create namespace demo
```

> **Note:** The example manifests are in [the composite-database examples directory](/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples). Replace every placeholder before applying a manifest.

## Architecture and What Gets Backed Up

This walkthrough uses three distinct Neo4j deployments:

- `source-neo4j` owns the `movies` physical database, the `media` composite database, and the alias catalog that KubeStash backs up.
- `remote-neo4j` owns the `reviews` physical database. `source-neo4j` reaches it through a remote alias. Its data is outside this backup's scope.
- `restored-neo4j` receives the physical database backup and reconstructed catalog.

KubeStash discovers the catalog through Neo4j's `system` database. A backup contains physical database backup artifacts and metadata for composite database definitions, composite aliases, and standalone local and remote aliases. Depending on the alias, this metadata includes its location, target database, URL, username, credential type, driver settings, and properties.

The password for a stored-native remote alias is intentionally **not** stored in backup metadata. You must supply it separately during restore.

The database selector has the following behavior:

- The default `*` selector backs up all visible physical databases and captures all composite and standalone alias definitions.
- Selecting a composite database explicitly automatically includes physical databases referenced by its local aliases.
- A remote alias definition is captured, but the data in its remote target database is not backed up.
- A local alias target must be selected and restored successfully before its composite database and alias can be restored.
- Database and alias exclusions are honored. Do not exclude a physical database required by a selected local alias.
- The `system` database is used for catalog discovery; it is not restored as a normal physical database.

During restore, KubeStash performs these operations in dependency order:

1. Restore the selected physical databases.
2. Wait for local alias target databases to become visible and online.
3. Create or replace composite databases.
4. Restore local and remote composite aliases.
5. Restore standalone aliases.
6. Verify the reconstructed composite and alias catalog.

Restore preflight checks catalog conflicts and required remote credentials before destructive restore work begins. Without overwrite, a conflicting physical database, composite database, or alias causes the restore to fail. With overwrite enabled, KubeStash replaces catalog definitions where supported. It may temporarily detach aliases that depend on a physical database being replaced and recreate them after that database is restored.

## Deploy the Source and Remote Neo4j Instances

### Generate the Remote Alias Encryption Key

Neo4j reversibly encrypts stored-native remote alias credentials in the `system` database. Before creating such an alias, generate a 256-bit AES key and store it in a password-protected PKCS12 keystore:

```bash
$ KEYSTORE_PASSWORD=$(openssl rand -base64 32)
$ keytool -genseckey -keyalg AES -keysize 256 -storetype PKCS12 \
    -keystore neo4j-remote-alias-keystore.p12 -alias neo \
    -storepass "$KEYSTORE_PASSWORD"
```

Create a Kubernetes Secret containing the keystore and its password. The Secret must exist before you create either Neo4j resource that references it:

```bash
$ kubectl create secret generic neo4j-remote-alias-keystore -n demo \
    --from-file=aes=neo4j-remote-alias-keystore.p12 \
    --from-literal=password="$KEYSTORE_PASSWORD"
```

The `source-neo4j` and `restored-neo4j` manifests reference the Secret through this configuration:

```yaml
spec:
  configuration:
    remoteAliasKeystore:
      keystoreRef:
        name: neo4j-remote-alias-keystore
        key: aes
      passwordRef:
        name: neo4j-remote-alias-keystore
        key: password
      keyName: neo
```

Here, `aes` is the Secret data key containing the PKCS12 file, `password` contains its password, and `keyName` must match the alias passed to `keytool` (`neo` in this example). Keep the keystore and password secure. Do not commit either one to source control.

### Create the Neo4j Instances

Create the source instance:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/source-neo4j.yaml
```

Create the separate instance that will host the remote database:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/remote-neo4j.yaml
```

Both manifests request three Neo4j replicas and durable `2Gi` storage, and use `WipeOut` as the deletion policy. The source manifest also configures the remote alias keystore. Wait until both databases are ready:

```bash
$ kubectl wait --for=jsonpath='{.status.phase}'=Ready neo4j/source-neo4j neo4j/remote-neo4j -n demo --timeout=10m
```

Store the generated administrator passwords in shell variables. These commands do not print the values:

```bash
$ SOURCE_PASS=$(kubectl get secret source-neo4j-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
$ REMOTE_PASS=$(kubectl get secret remote-neo4j-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
```

## Create Sample Physical Databases

Administrative commands run against the `system` database. Create `movies` on the source and `reviews` on the remote instance:

```bash
$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d system -u neo4j -p "$SOURCE_PASS" \
    "CREATE DATABASE movies WAIT;"

$ kubectl exec -n demo remote-neo4j-0 -- cypher-shell -d system -u neo4j -p "$REMOTE_PASS" \
    "CREATE DATABASE reviews WAIT;"
```

Add data to each physical database:

```bash
$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d movies -u neo4j -p "$SOURCE_PASS" \
    "CREATE (:Movie {title: 'Example Movie'});"

$ kubectl exec -n demo remote-neo4j-0 -- cypher-shell -d reviews -u neo4j -p "$REMOTE_PASS" \
    "CREATE (:Review {summary: 'Example Review'});"
```

## Create a Composite Database

Create the `media` composite database on the source instance:

```bash
$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d system -u neo4j -p "$SOURCE_PASS" \
    "CREATE COMPOSITE DATABASE media WAIT;"
```

## Create a Local Alias

Create `media.movies` as a constituent local alias. Also create `movies-local` to demonstrate a standalone local alias:

```bash
$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d system -u neo4j -p "$SOURCE_PASS" \
    "CREATE ALIAS media.movies FOR DATABASE movies;
     CREATE ALIAS `movies-local` FOR DATABASE movies;"
```

The `media` prefix makes `media.movies` a constituent of the composite database. `movies-local` has no composite namespace and is therefore standalone.

## Create a Stored-Native Remote Alias

With `spec.configuration.remoteAliasKeystore` configured, create `media.reviews` and a standalone alias named `reviews-remote`. The service URL addresses the separate `remote-neo4j` deployment. This in-cluster example uses the `neo4j://` scheme, so its driver settings explicitly disable TLS enforcement. Use a secure `neo4j+s://` URL in production.

```bash
$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d system -u neo4j -p "$SOURCE_PASS" \
    "CREATE ALIAS media.reviews FOR DATABASE reviews
       AT 'neo4j://remote-neo4j.demo.svc:7687'
       USER neo4j PASSWORD '$REMOTE_PASS'
       DRIVER {ssl_enforced: false}
       PROPERTIES {purpose: 'reviews'};
     CREATE ALIAS `reviews-remote` FOR DATABASE reviews
       AT 'neo4j://remote-neo4j.demo.svc:7687'
       USER neo4j PASSWORD '$REMOTE_PASS'
       DRIVER {ssl_enforced: false};"
```

Do not place the real password in a manifest, documentation, shell history, or source control. The shell expands `REMOTE_PASS` only when this command runs.

## Verify the Source Catalog

Use the `system` database for catalog queries:

```bash
$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d system -u neo4j -p "$SOURCE_PASS" \
    "SHOW DATABASES
     YIELD name, type, currentStatus
     RETURN name, type, currentStatus
     ORDER BY name;"

$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d system -u neo4j -p "$SOURCE_PASS" \
    "SHOW ALIASES FOR DATABASE
     YIELD name, composite, database, location, url, user
     RETURN name, composite, database, location, url, user
     ORDER BY name;"
```

Confirm that `movies` is `standard` and online, `media` is `composite`, both `media.*` aliases belong to `media`, and the two standalone aliases have a null `composite` value. The local aliases must target `movies`; the remote aliases must show location `remote`, database `reviews`, the configured URL, and user `neo4j`.

Verify that both composite constituents can be queried:

```bash
$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d media -u neo4j -p "$SOURCE_PASS" \
    "USE media.movies MATCH (m:Movie) RETURN m.title;"

$ kubectl exec -n demo source-neo4j-0 -- cypher-shell -d media -u neo4j -p "$SOURCE_PASS" \
    "USE media.reviews MATCH (r:Review) RETURN r.summary;"
```

## Configure BackupStorage and RetentionPolicy

Create the S3 credential Secret as described in the [basic logical backup guide](/docs/guides/neo4j/backup/kubestash/logical/index.md#prepare-backend). Then edit the placeholders in `backupstorage.yaml` and apply it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/backupstorage.yaml
```

The provider block selects S3 and supplies its bucket, region, endpoint, credential Secret, and object prefix. `usagePolicy` permits repositories in all namespaces to use this storage. `deletionPolicy: Delete` removes stored backup data when the `BackupStorage` is deleted.

Create the retention policy:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/retentionpolicy.yaml
```

This policy retains the last five successful and two failed snapshots for at most two months. Its usage policy allows all namespaces.

## Create BackupConfiguration

Apply the backup configuration:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/backupconfiguration.yaml
```

The target is `source-neo4j`. The backend refers to `s3-storage` and `demo-retention`. The session runs every five minutes and writes to repository `s3-neo4j-composite-repo` under `/neo4j-composite`. The `logical-backup` task uses its default `*` selector, so it backs up all visible physical databases and captures every composite and standalone alias definition.

To select only `media`, set `databases: "media"` in the task parameters. KubeStash will also select `movies` because `media.movies` depends on it. Do not exclude `movies` from that backup.

## Verify the Backup and Snapshot

Wait for the configuration and its first backup to succeed:

```bash
$ kubectl get backupconfiguration -n demo source-neo4j-backup
$ kubectl get backupsession -n demo -w
```

Then inspect the repository and snapshot:

```bash
$ kubectl get repository -n demo s3-neo4j-composite-repo
$ kubectl get snapshot -n demo -l kubestash.com/repo-name=s3-neo4j-composite-repo
```

The `BackupConfiguration` and repository should become `Ready`, and the `BackupSession` and snapshot should reach `Succeeded`. Snapshot names and timings are generated, so this guide does not show fabricated output. Save the snapshot name if you prefer to restore a fixed snapshot instead of `latest`.

## Deploy the Restore Target

Create the empty restore target:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/restored-neo4j.yaml
```

The target uses the same Neo4j version, topology, and remote alias keystore configuration. `spec.init.waitForInitialRestore: true` keeps initial database startup coordinated with the restore. The restore Job will seed `restored-neo4j-0` through its PVC. The keystore configuration lets the restored Neo4j instance encrypt credentials while KubeStash recreates stored-native remote aliases.

## Create the Remote Alias Credential Secret

Stored-native alias passwords are never included in backup metadata. Create this Secret in the same namespace as the `RestoreSession` before starting the restore:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: neo4j-remote-alias-credentials
  namespace: demo
type: Opaque
stringData:
  credentials.yaml: |
    media.reviews:
      password: "<remote-database-password>"
    reviews-remote:
      password: "<remote-database-password>"
```

Replace the placeholders locally and apply the file. Do not commit real values:

```bash
$ kubectl apply -f remote-alias-credentials.yaml
```

The Secret must contain a `credentials.yaml` key. Its value is a map keyed by the exact, complete alias name returned by `SHOW ALIASES FOR DATABASE`. Every selected stored-native remote alias needs a non-empty `password`. The username, URL, driver settings, and properties come from backup metadata; only the password comes from this Secret. The remote endpoint must be reachable from `restored-neo4j`. OIDC credential-forwarding aliases do not use stored passwords and do not need entries.

The committed [example Secret](/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/remote-alias-credentials.yaml) contains placeholders only; download or copy it before substituting credentials.

## Create RestoreSession

Create the restore session after the credential Secret exists:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/composite-database/examples/restoresession.yaml
```

The target is `restored-neo4j`, and the data source selects the latest snapshot from `s3-neo4j-composite-repo`. The `logical-backup-restore` task uses `restored-neo4j-0` as the seed server. `remoteAliasCredentialsSecret` identifies the Secret that supplies stored-native remote alias passwords. The restore Job mounts `data-restored-neo4j-0` at `/data` with subpath `data` and runs as the Neo4j user `7474`, ensuring correct ownership of restored files.

The addon's default restore arguments enable overwrite. If overwrite is disabled, any conflicting physical database, composite database, or alias fails preflight. Review the target catalog before restoring into a non-empty deployment.

Watch the restore:

```bash
$ kubectl get restoresession -n demo source-neo4j-restore -w
```

Continue only after its phase is `Succeeded` and `restored-neo4j` is `Ready`.

## Verify the Restored Physical Databases, Composite Database, and Aliases

Load the restore-target password without printing it:

```bash
$ RESTORED_PASS=$(kubectl get secret restored-neo4j-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
```

Verify database type and status:

```bash
$ kubectl exec -n demo restored-neo4j-0 -- cypher-shell -d system -u neo4j -p "$RESTORED_PASS" \
    "SHOW DATABASES
     YIELD name, type, currentStatus
     RETURN name, type, currentStatus
     ORDER BY name;"
```

Confirm that `movies` is online and that `media` has type `composite`. Next, compare alias metadata with the source catalog:

```bash
$ kubectl exec -n demo restored-neo4j-0 -- cypher-shell -d system -u neo4j -p "$RESTORED_PASS" \
    "SHOW ALIASES FOR DATABASE
     YIELD name, composite, database, location, url, user
     RETURN name, composite, database, location, url, user
     ORDER BY name;"
```

Confirm the expected composite membership, target, location, URL, and username for all four aliases. Finally, query both composite constituents:

```bash
$ kubectl exec -n demo restored-neo4j-0 -- cypher-shell -d media -u neo4j -p "$RESTORED_PASS" \
    "USE media.movies MATCH (m:Movie) RETURN m.title;"

$ kubectl exec -n demo restored-neo4j-0 -- cypher-shell -d media -u neo4j -p "$RESTORED_PASS" \
    "USE media.reviews MATCH (r:Review) RETURN r.summary;"

$ kubectl get restoresession -n demo source-neo4j-restore
```

The local query reads restored `movies` data. The remote query reads the live `reviews` database on `remote-neo4j`; it does not prove that remote data was part of the backup. The final command must report the `RestoreSession` phase as `Succeeded`.

## Limitations and Troubleshooting

- **Missing `remoteAliasCredentialsSecret`:** A restore selecting a stored-native remote alias fails preflight. Add the parameter under the restore task and point it to a Secret in the `RestoreSession` namespace.
- **Missing `credentials.yaml`:** Recreate the Secret with a key named exactly `credentials.yaml`.
- **Alias absent from the map:** Add the complete name shown by `SHOW ALIASES FOR DATABASE`, including its composite prefix. For example, `media.reviews` is not interchangeable with `reviews`.
- **Empty password:** Provide a non-empty password for every selected stored-native remote alias.
- **Remote endpoint unreachable:** Verify DNS, network policy, TLS settings, service availability, and connectivity from the restore-target pods. Metadata restoration does not make the remote endpoint available.
- **Local target excluded:** Restore the physical target of every selected local alias. If `movies` is excluded or fails, `media.movies` and its dependent composite catalog cannot be restored correctly.
- **Existing name conflict:** Without overwrite, conflicting physical databases, composite databases, and aliases fail the restore. With overwrite, KubeStash replaces supported catalog definitions and may temporarily detach and recreate dependent aliases.
- **Remote metadata versus data:** KubeStash captures a remote alias definition, not the data stored in the remote `reviews` database. Back up that remote Neo4j instance separately.
- **`system` database:** KubeStash reads it to discover the catalog, but does not restore it as an ordinary physical database.
- **Remote passwords:** Passwords are deliberately never written to backup metadata. They must come from the restore-time Secret. OIDC credential-forwarding aliases are the exception because they do not store a remote password.

For any failure, inspect the restore Job logs and compare both catalogs:

```bash
$ kubectl get jobs -n demo -l kubestash.com/invoker-name=source-neo4j-restore
$ kubectl logs -n demo job/<restore-job-name>
```

Run `SHOW DATABASES` and `SHOW ALIASES FOR DATABASE` against the source and target `system` databases to identify missing targets, conflicts, or mismatched alias names.

## Cleanup

Delete the tutorial resources when you no longer need them:

```bash
$ kubectl delete restoresession -n demo source-neo4j-restore
$ kubectl delete backupconfiguration -n demo source-neo4j-backup
$ kubectl delete backupstorage -n demo s3-storage
$ kubectl delete retentionpolicy -n demo demo-retention
$ kubectl delete secret -n demo neo4j-remote-alias-credentials neo4j-remote-alias-keystore s3-secret
$ kubectl delete neo4j -n demo restored-neo4j source-neo4j remote-neo4j
```

Because these Neo4j resources use `deletionPolicy: WipeOut`, deleting them also removes their database storage. The `BackupStorage` uses `deletionPolicy: Delete`, so deleting it removes its stored backup data.
