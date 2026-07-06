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

This guide will show you how to use `KubeDB` Migration to migrate an existing `MariaDB` database — such as one running on AWS RDS or any external instance — entirely into a KubeDB-managed `MariaDB` with minimal downtime.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` operator with the courier operator enabled in your cluster following the steps [here](/docs/operatormanual/migration/).

- The source `MariaDB` instance must be network-reachable from within your Kubernetes cluster.

- The source `MariaDB` instance must have binary logging enabled with `binlog_format=ROW` and `binlog_row_image=FULL`. The database user provided for migration must have replication privileges.

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/mariadb/concepts/appbinding/)
    - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
    - [Migration](/docs/guides/mariadb/concepts/migrator/)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Database

We will use an **AWS RDS MariaDB** instance as the source. Connect to it as the admin user to verify the prerequisites and set up the migration user and test data.

<details>
<summary><b>Configuring your source instance.</b></summary>

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

### Verify prerequisites

```bash
$ mysql -h <rds-endpoint>.rds.amazonaws.com -u admin -p
```

```sql
-- Verify binary logging is enabled
SHOW VARIABLES LIKE 'log_bin';
+---------------+-------+
| Variable_name | Value |
+---------------+-------+
| log_bin       | ON    |
+---------------+-------+

-- Verify binlog format and row image
SHOW VARIABLES LIKE 'binlog_format';
+---------------+-------+
| Variable_name | Value |
+---------------+-------+
| binlog_format | ROW   |
+---------------+-------+

SHOW VARIABLES LIKE 'binlog_row_image';
+------------------+-------+
| Variable_name    | Value |
+------------------+-------+
| binlog_row_image | FULL  |
+------------------+-------+
```

### Create a dedicated migration user

Create a dedicated user with the minimum required privileges:

```sql
CREATE USER 'migrator'@'%' IDENTIFIED BY '<password>';

-- For CDC (binlog streaming)
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'migrator'@'%';

-- For schema migration (mariadb-dump) and snapshot
GRANT SELECT, SHOW DATABASES, SHOW VIEW, TRIGGER, EVENT, LOCK TABLES, PROCESS ON *.* TO 'migrator'@'%';

FLUSH PRIVILEGES;

-- Verify
SHOW GRANTS FOR 'migrator'@'%';
```

The `migrator` user is referenced in the Kubernetes secret and AppBinding for the rest of this guide.

### Create table and seed data

```sql
CREATE DATABASE shop;
USE shop;

CREATE TABLE orders (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  customer_name VARCHAR(100) NOT NULL,
  product       VARCHAR(100) NOT NULL,
  quantity      INT          NOT NULL DEFAULT 1,
  status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
  created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO orders (customer_name, product, quantity, status) VALUES
  ('Alice', 'Laptop',     1, 'shipped'),
  ('Bob',   'Headphones', 2, 'pending'),
  ('Carol', 'Keyboard',   3, 'delivered');

SELECT * FROM orders;
+----+---------------+------------+----------+-----------+---------------------+
| id | customer_name | product    | quantity | status    | created_at          |
+----+---------------+------------+----------+-----------+---------------------+
|  1 | Alice         | Laptop     |        1 | shipped   | 2026-06-29 08:00:00 |
|  2 | Bob           | Headphones |        2 | pending   | 2026-06-29 08:00:01 |
|  3 | Carol         | Keyboard   |        3 | delivered | 2026-06-29 08:00:02 |
+----+---------------+------------+----------+-----------+---------------------+
3 rows in set (0.00 sec)
```

## Prepare Source Connection Information

First, create an authentication secret using the `migrator` user credentials:

```bash
$ kubectl create secret generic source-mariadb-auth -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=migrator \
                --from-literal=password=<password>
```

If your database has TLS enabled, create a secret with the CA certificate:

```bash
kubectl create secret generic ca-secret \
  --from-file=ca.crt=$CERT_PATH/ca.crt \
  --namespace=demo
```

> **Note:** For mTLS, also include the client certificate and key by appending <br> `--from-file=tls.crt=$CERT_PATH/tls.crt` <br> 
`--from-file=tls.key=$CERT_PATH/tls.key` <br> 
to the command above.

Now create an `AppBinding` with the necessary information. The courier operator reads the source MariaDB connection information from this AppBinding CR. Use the following YAML to create your AppBinding:

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
    url: "mariadb://<rds-endpoint>.rds.amazonaws.com:3306"
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

## Apply Migration CR

To migrate the database we have to create a `Migration` CR. Below is the YAML of the `Migration` CR that we are going to create:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Migration
metadata:
  name: mariadb-migrate
  namespace: demo
spec:
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
        database:
          - shop
        excludeDatabase: []
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/migration/mariadb-migrate.yaml
migration.courier.kubedb.com/mariadb-migrate created
```

Here we scope the migration to the `shop` database (`schema.database: [shop]`), enable both the bulk snapshot and CDC streaming phases, and cap connections at 100 on each side. For a full description of every field, see the [Migration CRD reference](/docs/guides/mariadb/concepts/migrator/).

## Watch Migration Progress

Let's wait for the `LAG` to reach near zero. Run the following command to watch `Migration` CR:

```bash
Every 2.0s: kubectl get migration -n demo

NAME              PHASE     DBTYPE    STAGE       LAG   PROGRESS   AGE
mariadb-migrate   Running   mariadb   Streaming   0B    100%       4h36m
```

### Verify initial snapshot on target

Once the migration reaches the `Streaming` stage, exec into the KubeDB target pod and confirm all seed rows were copied over:

```bash
$ kubectl exec -it -n demo target-mariadb-0 -- mysql -u root -p<root-password>
```

```sql
USE shop;
SELECT * FROM orders;
+----+---------------+------------+----------+-----------+---------------------+
| id | customer_name | product    | quantity | status    | created_at          |
+----+---------------+------------+----------+-----------+---------------------+
|  1 | Alice         | Laptop     |        1 | shipped   | 2026-06-29 08:00:00 |
|  2 | Bob           | Headphones |        2 | pending   | 2026-06-29 08:00:01 |
|  3 | Carol         | Keyboard   |        3 | delivered | 2026-06-29 08:00:02 |
+----+---------------+------------+----------+-----------+---------------------+
3 rows in set (0.00 sec)
```

### Test live CDC streaming

With the migration still running, connect to the **source RDS** instance and run some DML:

```bash
$ mysql -h <rds-endpoint>.rds.amazonaws.com -u migrator -p
```

```sql
-- Insert a new order
INSERT INTO shop.orders (customer_name, product, quantity, status)
VALUES ('Dave', 'Mouse', 1, 'pending');

-- Mark Bob's headphones as delivered
UPDATE shop.orders SET status = 'delivered' WHERE id = 2;

-- Remove the already-shipped laptop order
DELETE FROM shop.orders WHERE id = 1;
```

Wait a few seconds for the binlog events to propagate, then re-query the **target**:

```sql
USE shop;
SELECT * FROM orders;
+----+---------------+------------+----------+-----------+---------------------+
| id | customer_name | product    | quantity | status    | created_at          |
+----+---------------+------------+----------+-----------+---------------------+
|  2 | Bob           | Headphones |        2 | delivered | 2026-06-29 08:00:01 |
|  3 | Carol         | Keyboard   |        3 | delivered | 2026-06-29 08:00:02 |
|  4 | Dave          | Mouse      |        1 | pending   | 2026-06-29 08:10:00 |
+----+---------------+------------+----------+-----------+---------------------+
3 rows in set (0.00 sec)
```

The INSERT, UPDATE, and DELETE are all reflected on the target — CDC streaming is working correctly.

## Cutover

Once the `LAG` drops to near zero, stop all writes to the source database. Wait until the `LAG` reaches exactly zero — at that point both databases are fully in sync.

Now delete the `Migration` CR to stop the migration process:

```bash
$ kubectl delete migration -n demo mariadb-migrate
migration.courier.kubedb.com "mariadb-migrate" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `MariaDB` database. The migration is complete.
