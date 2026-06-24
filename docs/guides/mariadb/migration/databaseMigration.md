---
title: MariaDB Database Migration Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-migration-database
    name: MariaDB Database Migration
    parent: guides-mariadb-migration
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDB Database Migration

This guide will show you how to use `KubeDB` Migrator to migrate an existing `MariaDB` database — such as one running on AWS RDS or any external instance — entirely into a KubeDB-managed `MariaDB` with minimal downtime.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` operator with the Migrator operator enabled in your cluster following the steps [here](/docs/operatormanual/migration/).

- The source `MariaDB` instance must be network-reachable from within your Kubernetes cluster.

- The source `MariaDB` instance must have binary logging enabled with `binlog_format=ROW` and `binlog_row_image=FULL`. The database user provided for migration must have replication privileges. There is no single procedure to configure this — it depends on your deployment environment.

  <details>
  <summary>How to configure this on your source instance?</summary>

  **Self-hosted MariaDB** <br>

  Add the following to your `my.cnf` and restart MariaDB:
  ```ini
  [mysqld]
  log_bin          = mysql-bin
  binlog_format    = ROW
  binlog_row_image = FULL
  ```

  Then grant the required privileges to the migration user:
  ```sql
  GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO '<migration-user>'@'%';
  FLUSH PRIVILEGES;
  ```

  **AWS RDS MariaDB** <br>

  Enable [Automated Backups](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_LogAccess.MySQL.BinaryFormat.html) on the instance (this activates binary logging). Set `binlog_format = ROW` and `binlog_row_image = FULL` in an RDS Parameter Group. Then grant replication privileges via SQL as shown above.

  <br> <br> **Azure Database for MariaDB** <br>

  Binary logging is enabled automatically when backups are on. Set `binlog_format` and `binlog_row_image` under **Server Parameters** in the [Azure Portal](https://learn.microsoft.com/en-us/azure/mariadb/concepts-read-replicas). Then grant replication privileges via SQL as shown above.

  See the official [MariaDB Binary Log](https://mariadb.com/kb/en/binary-log/) docs for more details.

  </details>

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/mariadb/concepts/appbinding/)
    - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Connection Information

First, create an authentication secret to communicate with the source MariaDB database:

```bash
$ kubectl create secret generic source-mariadb-auth -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=<username> \
                --from-literal=password=<password>
```

If your database has TLS enabled, create a secret with the CA certificate:

```bash
kubectl create secret generic ca-secret \
  --from-file=ca.crt=$CERT_PATH/ca.crt \
  --namespace=demo
```

> **Note:** For mTLS, also include the client certificate and key by appending <br> `--from-file=tls.crt=$CERT_PATH/tls.crt` <br> `--from-file=tls.key=$CERT_PATH/tls.key` <br> to the command above.

Now create an `AppBinding` with the necessary information. The Migrator operator reads the source MariaDB connection information from this AppBinding CR. Use the following YAML to create your AppBinding:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: source-mariadb
  namespace: demo
spec:
  type: mariadb
  version: "10.5.23"
  clientConfig:
    url: "mariadb://host:port"
  secret:
    name: source-mariadb-auth
  tlsSecret: # omit if TLS is disabled
    name: ca-secret
```

Here,

- `spec.clientConfig.url` is the connection URL of the source MariaDB instance.
- `spec.secret.name` is the reference to the secret we created earlier, containing the MariaDB authentication information.

> For a `KubeDB`-managed database, an `AppBinding` is created by default. So there is no need to create one for the target database.

## Create Target MariaDB Database

KubeDB implements a `MariaDB` CRD to define the specification of a MariaDB database. Follow the `MariaDB` object to create the target database.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: target-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 20Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/migration/target-mariadb.yaml
mariadb.kubedb.com/target-mariadb created
```

> Note: Adjust the `resources.requests.storage` based on the source database size.

Wait until `target-mariadb` has status `Ready`.

## Apply Migrator CR

To migrate the database we have to create a `Migrator` CR. Below is the YAML of the `Migrator` CR that we are going to create:

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: mariadb-migrate
  namespace: demo
spec:
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
  source:
    mariadb:
      connectionInfo:
        appBinding:
          name: source-mariadb
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
      schema:
        enabled: true
        database: [] # database to include
        excludeDatabase: [] # database to exclude
      snapshot:
        enabled: true
        pipeline:
          workers: 3
          sinkers: 4
          buffer: 12
          write_batch_size: 200
          read_batch_size: 1000
      streaming:
        enabled: true

  target:
    mariadb:
      connectionInfo:
        appBinding:
          name: target-mariadb
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
```

Here,

**`spec.source` / `spec.target` — connectionInfo:**
- `appBinding.name` / `appBinding.namespace` — references the `AppBinding` for the source or target MariaDB instance.
- `dbName` — the internal database used as the initial connection entry point.
- `maxConnections` — limits the number of concurrent connections the migrator opens to this MariaDB instance.

**`spec.source.schema` — schema migration phase:**
- `enabled: true` — enables the schema migration phase.
- `database` — list of databases to include; empty means all databases except system database(mysql,information_schema,performance_schema,sys).
- `excludeDatabase` — list of databases to exclude from migration.

**`spec.source.snapshot` — bulk snapshot phase:**
- `enabled: true` — enables the initial bulk snapshot phase.
- `pipeline.workers` — number of parallel workers, each processing a separate table concurrently.
- `pipeline.sinkers` — number of parallel write workers pushing data to the target for each worker.
- `pipeline.buffer` — size of the in-memory queue (in records) between readers and writers.
- `pipeline.read_batch_size` — number of rows fetched per read batch from the source.
- `pipeline.write_batch_size` — number of rows written per batch to the target.

**`spec.source.streaming` — CDC streaming phase:**
- `enabled: true` — enables change-data capture streaming after the snapshot completes, keeping the target continuously in sync with ongoing changes on the source.

## Watch Migration Progress

Let's wait for the `LAG` to reach near zero. Run the following command to watch `Migrator` CR:

```bash
Every 2.0s: kubectl get migrator -n demo

NAME              PHASE     DBTYPE    STAGE       LAG   PROGRESS   AGE
mariadb-migrate   Running   mariadb   Streaming   0B               4h36m
```

## Cutover

Once the `LAG` drops to near zero, stop all writes to the source database. Wait until the `LAG` reaches exactly zero — at that point both databases are fully in sync.

Now delete the `Migrator` CR to stop the migration process:

```bash
$ kubectl delete migrator -n demo mariadb-migrate
migrator.migrator.kubedb.com "mariadb-migrate" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `MariaDB` database. The migration is complete.
