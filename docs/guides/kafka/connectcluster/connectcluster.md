---
title: ConnectCluster
menu:
  docs_{{ .version }}:
    identifier: kf-connectcluster-guides-connectcluster
    name: ConnectCluster
    parent: kf-connectcluster-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

### What is Kafka ConnectCluster?

Kafka Connect Cluster is an awesome tool for building reliable and scalable data pipelines. Its pluggable design makes it possible to build powerful pipelines without writing a single line of code. It can be used to stream changes out of a database and into Kafka, enabling other services to easily react in real-time. 
It has three main components:
1. Workers: Workers are responsible for running and distributing connectors and tasks.
2. Connectors: Connectors are responsible for managing the lifecycle of tasks.
3. Tasks: Tasks are responsible for moving data between Kafka and other systems.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka/connectcluster) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Create a Kafka ConnectCluster
Here, we are going to create a TLS secured Kafka ConnectCluster.

### Create Issuer/ ClusterIssuer

At first, make sure you have [cert-manager](https://cert-manager.io/docs/installation/) installed on your k8s for enabling TLS. KubeDB operator uses cert manager to inject certificates into kubernetes secret & uses them for secure `SASL` encrypted communication among kafka connectcluster worker nodes and client. We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in ConnectCluster. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you CA certificates using openssl.

```bash
openssl req -newkey rsa:2048 -keyout ca.key -nodes -x509 -days 3650 -out ca.crt
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls connectcluster-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: connectcluster-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: connectcluster-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tls/kcc-issuer.yaml
issuer.cert-manager.io/connectcluster-ca-issuer created
```

### Provision TLS secured ConnectCluster

For this tutorial, we are going to use ConnectCluster version `3.6.1` with three worker nodes. To learn more about ConnectCluster CR, please visit [here](/docs/guides/kafka/concepts/connectcluster.md) and visit [here](/docs/guides/kafka/concepts/kafkaconnectorversion.md) to learn about KafkaConnectorVersion CR.

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: ConnectCluster
metadata:
  name: connectcluster-distributed
  namespace: demo
spec:
  version: 3.6.1
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: connectcluster-ca-issuer
  configSecret:
    name: connectcluster-custom-config
  replicas: 3
  connectorPlugins:
  - postgres-2.4.2.final
  - jdbc-2.6.1.final
  kafkaRef:
    name: kafka-prod
    namespace: demo
  deletionPolicy: WipeOut
```
Here,
- `spec.version` - is the name of the KafkaVersion CR. Here, a ConnectCluster of version `3.6.1` will be created.
- `spec.enableSSL` - specifies whether the ConnectCluster should be TLS secured or not.
- `spec.tls.issuerRef` - specifies the name of the Issuer CR. Here, the ConnectCluster will use the `connectcluster-ca-issuer` Issuer to enable SSL/TLS.
- `spec.replicas` - specifies the number of ConnectCluster workers.
- `spec.configSecret` - specifies the name of the secret that contains the custom configuration for the ConnectCluster. Here, the ConnectCluster will use the `connectcluster-custom-config` secret for custom configuration.
- `spec.connectorPlugins` - is the name of the KafkaConnectorVersion CR. Here, mongodb, mysql, postgres, and jdbc connector-plugins will be loaded to the ConnectCluster worker nodes.
- `spec.kafkaRef` specifies the Kafka instance that the ConnectCluster will connect to. Here, the ConnectCluster will connect to the Kafka instance named `kafka-prod` in the `demo` namespace.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete ConnectCluster CR. Deletion policy `WipeOut` will delete the worker pods, secret when the ConnectCluster CR is deleted.

## N.B:
1. If replicas are set to 1, the ConnectCluster will run in standalone mode, you can't scale replica after provision the cluster.
2. If replicas are set to more than 1, the ConnectCluster will run in distributed mode.
3. If you want to run the ConnectCluster in distributed mode with 1 replica, you must set the `CONNECT_CLUSTER_MODE` environment variable to `distributed` in the pod template.
```yaml
spec:
  podTemplate:
    spec:
    containers:
    - name: connect-cluster
      env:
      - name: CONNECT_CLUSTER_MODE
        value: distributed
```

Before creating ConnectCluster, you must need a Kafka cluster. If you don't have a Kafka cluster, you can create one by following this [tutorial](/docs/guides/kafka/clustering/topology-cluster/index.md).
We are using TLS secured Kafka topology cluster `kafka-prod` with 3 brokers and 3 controllers in this tutorial.
We are also running our cluster with custom configuration. The custom configuration is stored in a secret named `connectcluster-custom-config`. So, create the secret with the custom configuration.

Create a file named `config.properties` with the following content:

```bash
$ cat config.properties
key.converter.schemas.enable=true
value.converter.schemas.enable=true
internal.key.converter.schemas.enable=true
internal.value.converter.schemas.enable=true
internal.key.converter=org.apache.kafka.connect.json.JsonConverter
internal.value.converter=org.apache.kafka.connect.json.JsonConverter
key.converter=org.apache.kafka.connect.json.JsonConverter
value.converter=org.apache.kafka.connect.json.JsonConverter
```

Create a secret named `connectcluster-custom-config` with the `config.properties` file:

```bash
$ kubectl create secret generic connectcluster-custom-config --from-file=./config.properties -n demo
```

Let's create the ConnectCluster using the above YAML:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/connectcluster/kcc-distributed.yaml
connectcluster.kafka.kubedb.com/connectcluster-distributed created
```

Watch the bootstrap progress:

```bash
$ kubectl get kcc -n demo -w
NAME                         TYPE                        VERSION   STATUS         AGE
connectcluster-distributed   kafka.kubedb.com/v1alpha1   3.6.1     Provisioning   0s
connectcluster-distributed   kafka.kubedb.com/v1alpha1   3.6.1     Provisioning   33s
.
.
connectcluster-distributed   kafka.kubedb.com/v1alpha1   3.6.1     Ready          97s
```

Hence, the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of ConnectCluster:

```bash
$ kubectl get all,petset,secret -n demo -l 'app.kubernetes.io/instance=connectcluster-distributed'
NAME                               READY   STATUS    RESTARTS   AGE
pod/connectcluster-distributed-0   1/1     Running   0          8m55s
pod/connectcluster-distributed-1   1/1     Running   0          8m52s

NAME                                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/connectcluster-distributed        ClusterIP   10.128.238.9   <none>        8083/TCP   17m
service/connectcluster-distributed-pods   ClusterIP   None           <none>        8083/TCP   17m

NAME                                                      READY   AGE
petset.apps.k8s.appscode.com/connectcluster-distributed   2/2     8m56s

NAME                                                            TYPE                              VERSION   AGE
appbinding.appcatalog.appscode.com/connectcluster-distributed   kafka.kubedb.com/connectcluster   3.6.1     8m56s

NAME                                                      TYPE                       DATA   AGE
secret/connectcluster-distributed-client-connect-cert     kubernetes.io/tls          3      17m
secret/connectcluster-distributed-config                  Opaque                     1      17m
secret/connectcluster-distributed-connect-cred            kubernetes.io/basic-auth   2      17m
secret/connectcluster-distributed-connect-keystore-cred   Opaque                     3      17m
secret/connectcluster-distributed-kafka-client-cred       Opaque                     5      17m
secret/connectcluster-distributed-server-connect-cert     kubernetes.io/tls          5      17m
```

We are going to use the `postgres` source connector to stream data from a Postgres database to Kafka and the `jdbc` sink connector to stream data from Kafka to MySQL database.
To do that, we need to create a Postgres database. You can create a Postgres database by following this [tutorial](/docs/guides/postgres/quickstart/quickstart.md).
We also need a MySQL database. You can create a MySQL database by following this [tutorial](/docs/guides/mysql/quickstart/index.md).
> Note: Make sure you have a Postgres database must be running with `wal_level=logical`. You can check the `wal_level` by connecting to the Postgres database using the `psql` command-line tool.
```bash
postgres-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# show wal_level;
 wal_level 
-----------
 logical
(1 row)
```

### Create Postgres Source Connector

To create a Postgres source connector with KubeDB `Connector` CR, you need to create a secret that contains the Postgres database credentials and the connector configuration. The secret should have the following configuration with filename `config.properties`:
```bash
$ cat config.properties
connector.class=io.debezium.connector.postgresql.PostgresConnector
tasks.max=1
database.hostname=postgres.demo.svc
database.port=5432
database.user=postgres
database.password=l!eQe0JG8mr62tOM
database.dbname=source_database
key.converter=org.apache.kafka.connect.json.JsonConverter
key.converter.schemas.enable=true
value.converter=org.apache.kafka.connect.json.JsonConverter
value.converter.schemas.enable=true
database.whitelist=public.users
database.history.kafka.topic=schema-changes.users
```
Here,
- `connector.class` - specifies the connector class. Here, the Postgres source connector will be used.
- `tasks.max` - specifies the maximum number of tasks that should be created for this connector.
- `database.hostname` - specifies the hostname of the Postgres database. Update the value with the actual hostname of your Postgres database.
- `database.port` - specifies the port of the Postgres database. Update the value with the actual port of your Postgres database.
- `database.user` - specifies the username of the Postgres database. Update the value with the actual username of your Postgres database.
- `database.dbname` - specifies the name of the Postgres datab- `database.password` - specifies the password of the Postgres database. Update the value with the actual password of your Postgres database.
ase. Update the value with the actual name of your Postgres database that you want to stream data from.
- `database.whitelist` - specifies the list of tables that you want to stream data from. Here, the `public.users` table will be streamed. Update the value with the list of tables using comma separated that you want to stream data from.
- `database.history.kafka.topic` - specifies the Kafka topic where the connector will store the schema changes.

Now, create the secret named `postgres-source-connector-config` with the `config.properties` file:

```bash
$ kubectl create secret generic postgres-source-connector-config --from-file=./config.properties -n demo
```

Now, create a `Connector` CR to create the Postgres source connector:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: Connector
metadata:
  name: postgres-source-connector
  namespace: demo
spec:
  configSecret:
    name: postgres-source-connector-config
  connectClusterRef:
    name: connectcluster-distributed
    namespace: demo
  deletionPolicy: WipeOut
```

```bash
NAME                        TYPE                        CONNECTCLUSTER               STATUS    AGE
postgres-source-connector   kafka.kubedb.com/v1alpha1   connectcluster-distributed   Pending   0s
postgres-source-connector   kafka.kubedb.com/v1alpha1   connectcluster-distributed   Pending   1s
.
.
postgres-source-connector   kafka.kubedb.com/v1alpha1   connectcluster-distributed   Running   3s
```

The Postgres source connector is successfully created and running. Let's test it.

## Insert Data into Postgres Database

Insert some data into the `public.users` table of the Postgres database named `source_database` to stream the data to Kafka using the Postgres source connector.
Use the following commands to insert data into the `public.users` table:

```bash
postgres-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# CREATE DATABASE source_database;
CREATE DATABASE
postgres=# \c source_database
You are now connected to database "source_database" as user "postgres".
source_database=# CREATE TABLE users (
source_database(#     id SERIAL PRIMARY KEY,
source_database(#     name VARCHAR(100),
source_database(#     email VARCHAR(100),
source_database(#     age INT
source_database(# );
CREATE TABLE
source_database=# INSERT INTO users (name, email, age) VALUES
source_database-#     ('John Doe', 'john@example.com', 30),
source_database-#     ('Jane Smith', 'jane@example.com', 25),
source_database-#     ('Alice Johnson', 'alice@example.com', 35),
source_database-#     ('Bob Brown', 'bob@example.com', 40);
INSERT 0 4
source_database=# SELECT * FROM users;
 id |     name      |       email       | age 
----+---------------+-------------------+-----
  1 | John Doe      | john@example.com  |  30
  2 | Jane Smith    | jane@example.com  |  25
  3 | Alice Johnson | alice@example.com |  35
  4 | Bob Brown     | bob@example.com   |  40
(4 rows)
```

## Check Data from Kafka
Exec into one of the kafka brokers in interactive mode. Run consumer command to check the data in the topic.

```bash
~ $ kubectl exec -it -n demo kafka-prod-broker-1 -- bash
kafka@kafka-prod-broker-1:~$ kafka-console-consumer.sh --bootstrap-server localhost:9092 --consumer.config config/clientauth.properties --topic postgres.public.users --from-beginning

{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"before"},{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"after"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"version"},{"type":"string","optional":false,"field":"connector"},{"type":"string","optional":false,"field":"name"},{"type":"int64","optional":false,"field":"ts_ms"},{"type":"string","optional":true,"name":"io.debezium.data.Enum","version":1,"parameters":{"allowed":"true,last,false,incremental"},"default":"false","field":"snapshot"},{"type":"string","optional":false,"field":"db"},{"type":"string","optional":true,"field":"sequence"},{"type":"int64","optional":false,"field":"ts_us"},{"type":"int64","optional":false,"field":"ts_ns"},{"type":"string","optional":false,"field":"schema"},{"type":"string","optional":false,"field":"table"},{"type":"int64","optional":true,"field":"txId"},{"type":"int64","optional":true,"field":"lsn"},{"type":"int64","optional":true,"field":"xmin"}],"optional":false,"name":"io.debezium.connector.postgresql.Source","field":"source"},{"type":"string","optional":false,"field":"op"},{"type":"int64","optional":true,"field":"ts_ms"},{"type":"int64","optional":true,"field":"ts_us"},{"type":"int64","optional":true,"field":"ts_ns"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"id"},{"type":"int64","optional":false,"field":"total_order"},{"type":"int64","optional":false,"field":"data_collection_order"}],"optional":true,"name":"event.block","version":1,"field":"transaction"}],"optional":false,"name":"postgres.public.users.Envelope","version":2},"payload":{"before":null,"after":{"id":1,"name":"John Doe","email":"john@example.com","age":30},"source":{"version":"2.6.1.Final","connector":"postgresql","name":"postgres","ts_ms":1715071299083,"snapshot":"first","db":"source_database","sequence":"[null,\"43621864\"]","ts_us":1715071299083564,"ts_ns":1715071299083564000,"schema":"public","table":"users","txId":1554,"lsn":43621864,"xmin":null},"op":"r","ts_ms":1715071299247,"ts_us":1715071299247213,"ts_ns":1715071299247213680,"transaction":null}}
{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"before"},{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"after"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"version"},{"type":"string","optional":false,"field":"connector"},{"type":"string","optional":false,"field":"name"},{"type":"int64","optional":false,"field":"ts_ms"},{"type":"string","optional":true,"name":"io.debezium.data.Enum","version":1,"parameters":{"allowed":"true,last,false,incremental"},"default":"false","field":"snapshot"},{"type":"string","optional":false,"field":"db"},{"type":"string","optional":true,"field":"sequence"},{"type":"int64","optional":false,"field":"ts_us"},{"type":"int64","optional":false,"field":"ts_ns"},{"type":"string","optional":false,"field":"schema"},{"type":"string","optional":false,"field":"table"},{"type":"int64","optional":true,"field":"txId"},{"type":"int64","optional":true,"field":"lsn"},{"type":"int64","optional":true,"field":"xmin"}],"optional":false,"name":"io.debezium.connector.postgresql.Source","field":"source"},{"type":"string","optional":false,"field":"op"},{"type":"int64","optional":true,"field":"ts_ms"},{"type":"int64","optional":true,"field":"ts_us"},{"type":"int64","optional":true,"field":"ts_ns"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"id"},{"type":"int64","optional":false,"field":"total_order"},{"type":"int64","optional":false,"field":"data_collection_order"}],"optional":true,"name":"event.block","version":1,"field":"transaction"}],"optional":false,"name":"postgres.public.users.Envelope","version":2},"payload":{"before":null,"after":{"id":2,"name":"Jane Smith","email":"jane@example.com","age":25},"source":{"version":"2.6.1.Final","connector":"postgresql","name":"postgres","ts_ms":1715071299083,"snapshot":"true","db":"source_database","sequence":"[null,\"43621864\"]","ts_us":1715071299083564,"ts_ns":1715071299083564000,"schema":"public","table":"users","txId":1554,"lsn":43621864,"xmin":null},"op":"r","ts_ms":1715071299249,"ts_us":1715071299249635,"ts_ns":1715071299249635836,"transaction":null}}
{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"before"},{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"after"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"version"},{"type":"string","optional":false,"field":"connector"},{"type":"string","optional":false,"field":"name"},{"type":"int64","optional":false,"field":"ts_ms"},{"type":"string","optional":true,"name":"io.debezium.data.Enum","version":1,"parameters":{"allowed":"true,last,false,incremental"},"default":"false","field":"snapshot"},{"type":"string","optional":false,"field":"db"},{"type":"string","optional":true,"field":"sequence"},{"type":"int64","optional":false,"field":"ts_us"},{"type":"int64","optional":false,"field":"ts_ns"},{"type":"string","optional":false,"field":"schema"},{"type":"string","optional":false,"field":"table"},{"type":"int64","optional":true,"field":"txId"},{"type":"int64","optional":true,"field":"lsn"},{"type":"int64","optional":true,"field":"xmin"}],"optional":false,"name":"io.debezium.connector.postgresql.Source","field":"source"},{"type":"string","optional":false,"field":"op"},{"type":"int64","optional":true,"field":"ts_ms"},{"type":"int64","optional":true,"field":"ts_us"},{"type":"int64","optional":true,"field":"ts_ns"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"id"},{"type":"int64","optional":false,"field":"total_order"},{"type":"int64","optional":false,"field":"data_collection_order"}],"optional":true,"name":"event.block","version":1,"field":"transaction"}],"optional":false,"name":"postgres.public.users.Envelope","version":2},"payload":{"before":null,"after":{"id":3,"name":"Alice Johnson","email":"alice@example.com","age":35},"source":{"version":"2.6.1.Final","connector":"postgresql","name":"postgres","ts_ms":1715071299083,"snapshot":"true","db":"source_database","sequence":"[null,\"43621864\"]","ts_us":1715071299083564,"ts_ns":1715071299083564000,"schema":"public","table":"users","txId":1554,"lsn":43621864,"xmin":null},"op":"r","ts_ms":1715071299249,"ts_us":1715071299249846,"ts_ns":1715071299249846409,"transaction":null}}
{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"before"},{"type":"struct","fields":[{"type":"int32","optional":false,"default":0,"field":"id"},{"type":"string","optional":true,"field":"name"},{"type":"string","optional":true,"field":"email"},{"type":"int32","optional":true,"field":"age"}],"optional":true,"name":"postgres.public.users.Value","field":"after"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"version"},{"type":"string","optional":false,"field":"connector"},{"type":"string","optional":false,"field":"name"},{"type":"int64","optional":false,"field":"ts_ms"},{"type":"string","optional":true,"name":"io.debezium.data.Enum","version":1,"parameters":{"allowed":"true,last,false,incremental"},"default":"false","field":"snapshot"},{"type":"string","optional":false,"field":"db"},{"type":"string","optional":true,"field":"sequence"},{"type":"int64","optional":false,"field":"ts_us"},{"type":"int64","optional":false,"field":"ts_ns"},{"type":"string","optional":false,"field":"schema"},{"type":"string","optional":false,"field":"table"},{"type":"int64","optional":true,"field":"txId"},{"type":"int64","optional":true,"field":"lsn"},{"type":"int64","optional":true,"field":"xmin"}],"optional":false,"name":"io.debezium.connector.postgresql.Source","field":"source"},{"type":"string","optional":false,"field":"op"},{"type":"int64","optional":true,"field":"ts_ms"},{"type":"int64","optional":true,"field":"ts_us"},{"type":"int64","optional":true,"field":"ts_ns"},{"type":"struct","fields":[{"type":"string","optional":false,"field":"id"},{"type":"int64","optional":false,"field":"total_order"},{"type":"int64","optional":false,"field":"data_collection_order"}],"optional":true,"name":"event.block","version":1,"field":"transaction"}],"optional":false,"name":"postgres.public.users.Envelope","version":2},"payload":{"before":null,"after":{"id":4,"name":"Bob Brown","email":"bob@example.com","age":40},"source":{"version":"2.6.1.Final","connector":"postgresql","name":"postgres","ts_ms":1715071299083,"snapshot":"last","db":"source_database","sequence":"[null,\"43621864\"]","ts_us":1715071299083564,"ts_ns":1715071299083564000,"schema":"public","table":"users","txId":1554,"lsn":43621864,"xmin":null},"op":"r","ts_ms":1715071299250,"ts_us":1715071299250200,"ts_ns":1715071299250200576,"transaction":null}}
```

Data from the `public.users` table of the Postgres database is successfully streamed to the Kafka topic named `postgres.public.users`.

Now, let's create a JDBC sink connector to stream data from the Kafka topic to a MySQL database.

### Create JDBC Sink Connector(MySQL)

To create a JDBC sink connector with KubeDB `Connector` CR, you need to create a secret that contains the MySQL database credentials and the connector configuration. The secret should have the following configuration with filename `config.properties`:
```bash
$ cat config.properties
heartbeat.interval.ms=3000
autoReconnect=true
connector.class=io.debezium.connector.jdbc.JdbcSinkConnector
tasks.max=1
connection.url=jdbc:mysql://mysql-demo.demo.svc:3306/sink_database
connection.username=root
connection.password=wBehTM*AjtEXk8Ig
insert.mode=upsert
delete.enabled=true
primary.key.mode=record_key
schema.evolution=basic
database.time_zone=UTC
auto.evolve=true
quote.identifiers=true
auto.create=true
value.converter.schemas.enable=true
value.converter=org.apache.kafka.connect.json.JsonConverter
table.name.format=${topic}
topics=postgres.public.users
pk.mode=kafka
```
Here,
- `heartbeat.interval.ms` - specifies the interval in milliseconds at which the connector should send heartbeat messages to the database.
- `autoReconnect` - specifies whether the connector should automatically reconnect to the database in case of a connection failure.
- `connector.class` - specifies the connector class. Here, the JDBC sink connector will be used.
- `tasks.max` - specifies the maximum number of tasks that should be created for this connector.
- `connection.url` - specifies the JDBC URL of the MySQL database. Update the value with the actual JDBC URL of your MySQL database.
- `connection.username` - specifies the username of the MySQL database. Update the value with the actual username of your MySQL database.
- `connection.password` - specifies the password of the MySQL database. Update the value with the actual password of your MySQL database.
- `insert.mode` - specifies the strategy used to `insert` events into the database.(update/upsert/insert)
- `auto.create` - specifies whether the connector should automatically create the table in the MySQL database if it does not exist.
- `table.name.format` - specifies the format of the table name. Here, the table name will be the same as the Kafka topic name.
- `topics` - specifies the Kafka topic from which the connector will stream data to the MySQL database.



Now, create the secret named `mysql-sink-connector-config` with the `config.properties` file:

```bash
$ kubectl create secret generic mysql-sink-connector-config --from-file=./config.properties -n demo
```

Before creating connector, create the database `sink_database` in MySQL database which is mentioned in the `connection.url` of the `config.properties` file. Example: `jdbc:mysql://<host>:<port>/<sink-db-name>`.
Connect to the MySQL database using the `mysql` command-line tool and create the `sink_database` database using following commands:

```bash
bash-4.4$ mysql -uroot -p'wBehTM*AjtEXk8Ig'
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 16
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create database sink_database;
Query OK, 1 row affected (0.02 sec)
```

Now, create a `Connector` CR to create the MySQL sink connector:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: Connector
metadata:
  name: mysql-sink-connector
  namespace: demo
spec:
  configSecret:
    name: mysql-sink-connector-config
  connectClusterRef:
    name: connectcluster-distributed
    namespace: demo
  deletionPolicy: WipeOut
```

```bash
NAME                        TYPE                        CONNECTCLUSTER               STATUS    AGE
mysql-sink-connector        kafka.kubedb.com/v1alpha1   connectcluster-distributed   Pending   0s
mysql-sink-connector        kafka.kubedb.com/v1alpha1   connectcluster-distributed   Pending   0s
.
.
mysql-sink-connector        kafka.kubedb.com/v1alpha1   connectcluster-distributed   Running   1s
```

The JDBC sink connector is successfully created and running. Let's test it by checking the data in the MySQL database using the `mysql` command-line tool as follows:

```bash
bash-4.4$ mysql -uroot -p'wBehTM*AjtEXk8Ig'
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 16
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> use sink_database;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
mysql> show tables;
+-------------------------+
| Tables_in_sink_database |
+-------------------------+
| postgres_public_users   |
+-------------------------+
1 row in set (0.01 sec)

mysql> select * from postgres_public_users;
+----+---------------+-------------------+------+
| id | name          | email             | age  |
+----+---------------+-------------------+------+
|  1 | John Doe      | john@example.com  |   30 |
|  2 | Jane Smith    | jane@example.com  |   25 |
|  3 | Alice Johnson | alice@example.com |   35 |
|  4 | Bob Brown     | bob@example.com   |   40 |
+----+---------------+-------------------+------+
4 rows in set (0.00 sec)
```

We can see that the data from the Kafka topic named `postgres.public.users` is successfully streamed to the MySQL database named `sink_database` to the table named `postgres_public_users`. The table has created automatically by the JDBC sink connector which is mentioned in the `table.name.format` and `auto.creare` property of the `config.properties` file. 

You can customize the connector configuration by updating the `config.properties` file. The connector will automatically reload the configuration and apply the changes. Also, you can update the using Kafka ConnectCluster Rest API through the `connectcluster-distributed` service.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo connectcluster connectcluster-distributed -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
connectcluster.kafka.kubedb.com/connectcluster-distributed patched

$ kubectl delete kf connectcluster-distributed  -n demo
connectcluster.kafka.kubedb.com "connectcluster-distributed" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for the production environment. You can follow these tips to avoid them.

1 **Use `deletionPolicy: Delete`**. It is nice to be able to resume the cluster from the previous one. So, we preserve auth `Secrets`. If you don't want to resume the cluster, you can just use `spec.deletionPolicy: WipeOut`. It will clean up every resource that was created with the ConnectCluster CR. For more details, please visit [here](/docs/guides/kafka/concepts/connectcluster.md#specdeletionpolicy).

## Next Steps

- [Quickstart Kafka](/docs/guides/kafka/quickstart/kafka/index.md) with KubeDB Operator.
- [Quickstart ConnectCluster](/docs/guides/kafka/connectcluster/quickstart.md) with KubeDB Operator.
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [ConnectCluster object](/docs/guides/kafka/concepts/connectcluster.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
