---
title: MySQL Database Migration Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-migration-database
    name: MySQL Database Migration
    parent: guides-mysql-migration
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL Database Migration

This guide will show you how to use `KubeDB` Migrator to migrate an existing `MySQL` database — such as one running on AWS RDS or any external instance — entirely into a KubeDB-managed `MySQL` with minimal downtime. 

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` operator with the Migrator operator enabled in your cluster following the steps [here](/docs/operatormanual/migration/).

- The source `MySQL` instance must be network-reachable from within your Kubernetes cluster.

- The source `MySQL` instance must have binary logging enabled with `binlog_format=ROW` and `binlog_row_image=FULL`. The database user provided for migration must have replication privileges.

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/mysql/concepts/appbinding/)
    - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
    - [Migrator](/docs/guides/mysql/concepts/migrator/)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Database

We will use an **AWS RDS MySQL** instance as the source. Connect to it as the admin user to verify the prerequisites and set up the migration user and test data.

<details>
<summary><b>Configuring your source instance.</b></summary>

<br> **Self-hosted MySQL** <br>

Add the following to your `my.cnf` and restart MySQL:
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

**AWS RDS MySQL** <br>

Enable [Automated Backups](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_LogAccess.MySQL.BinaryFormat.html) on the instance (this activates binary logging). Set `binlog_format = ROW` and `binlog_row_image = FULL` in an RDS Parameter Group. Then grant replication privileges via SQL as shown above.

<br> <br> **Azure Database for MySQL** <br>

Binary logging is enabled automatically when backups are on. Set `binlog_format` and `binlog_row_image` under **Server Parameters** in the [Azure Portal](https://learn.microsoft.com/en-us/azure/mysql/flexible-server/concepts-read-replicas). Then grant replication privileges via SQL as shown above.

<br> <br> **Google Cloud SQL for MySQL** <br>

Enable binary logging under **Backups** in the [Cloud Console](https://cloud.google.com/sql/docs/mysql/replication/create-replica), then set `binlog_format = ROW` and `binlog_row_image = FULL` under **Database flags**.

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

-- For schema migration (mysqldump) and snapshot
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
$ kubectl create secret generic source-mysql-auth -n demo \
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

> **Note:** For mTLS, also include the client certificate and key by appending <br> `--from-file=tls.crt=$CERT_PATH/tls.crt` <br> `--from-file=tls.key=$CERT_PATH/tls.key` <br> to the command above.

Now create an `AppBinding` with the necessary information. The Migrator operator reads the source MySQL connection information from this AppBinding CR. Use the following YAML to create your AppBinding:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: source-mysql
  namespace: demo
spec:
  type: mysql
  version: "8.4.8"
  clientConfig:
    url: "mysql://<rds-endpoint>.rds.amazonaws.com:3306"
  secret:
    name: source-mysql-auth
  tlsSecret: # omit if TLS is disabled
    name: ca-secret
```

Here,

- `spec.clientConfig.url` is the connection URL of the source MySQL instance.
- `spec.secret.name` is the reference to the secret we created earlier, containing the MySQL authentication information.

> For a `KubeDB`-managed database, an `AppBinding` is created by default. So there is no need to create one for the target database.

## Create Target MySQL Database

KubeDB implements a `MySQL` CRD to define the specification of a MySQL database. Follow the `MySQL` object to create the target database.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: target-mysql
  namespace: demo
spec:
  version: "8.4.8"
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 30Gi
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/migration/target-mysql.yaml
mysql.kubedb.com/target-mysql created
```
> Note: Adjust the `resources.requests.storage` based on source database.

Wait untill target-mysql has status `Ready`

## Apply Migrator CR

To Migrate database we have to create a `Migrator` CR. Below is the YAML of the `Migrator` CR that we are going to create,

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: mysql-migrate
  namespace: demo
spec:
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
  source:
    mysql:
      connectionInfo:
        appBinding:
          name: source-mysql
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
    mysql:
      connectionInfo:
        appBinding:
          name: target-mysql
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/migration/mysql-migrate.yaml
migrator.migrator.kubedb.com/mysql-migrate created
```

Here we scope the migration to the `shop` database (`schema.database: [shop]`), enable both the bulk snapshot and CDC streaming phases, and cap connections at 100 on each side. For a full description of every field, see the [Migrator CRD reference](/docs/guides/mysql/concepts/migrator/).

## Watch Migration Progress

Let's wait for the `LAG` to reach near zero. Run the following command to watch `Migrator` CR:

```bash
Every 2.0s: kubectl get migrator -n demo 

NAME            PHASE     DBTYPE   STAGE       LAG   PROGRESS   AGE
mysql-migrate   Running   mysql    Streaming   0B    100%       4h36m
```

### Verify initial snapshot on target

Once the migrator reaches the `Streaming` stage, exec into the KubeDB target pod and confirm all seed rows were copied over:

```bash
$ kubectl exec -it -n demo target-mysql-0 -- mysql -u root -p<root-password>
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

With the migrator still running, connect to the **source RDS** instance and run some DML:

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

Now delete the `Migrator` CR to stop the migration process:

```bash
$ kubectl delete migrator -n demo mysql-migrate
migrator.migrator.kubedb.com "mysql-migrate" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `MySQL` database. The migration is complete.
