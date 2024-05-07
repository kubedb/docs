---
title: Kafka
menu:
  docs_{{ .version }}:
    identifier: kf-readme-kafka
    name: Kafka
    parent: kf-kafka-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/kafka/
aliases:
  - /docs/{{ .version }}/guides/kafka/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Kafka Features

| Features                                                                           | Kafka    | ConnectCluster |
|------------------------------------------------------------------------------------|----------|----------------|
| Clustering - Combined (shared controller and broker nodes)                         | &#10003; | &#45;          |
| Clustering - Topology (dedicated controllers and broker nodes)                     | &#10003; | &#45;          |
| Custom Configuration                                                               | &#10003; | &#10003;       |
| Automated Version Update                                                           | &#10003; | &#10007;       |
| Automatic Vertical Scaling                                                         | &#10003; | &#10007;       |
| Automated Horizontal Scaling                                                       | &#10003; | &#10007;       |
| Automated Volume Expansion                                                         | &#10003; | &#45;          |
| Custom Docker Image                                                                | &#10003; | &#10003;       |
| Authentication & Authorization                                                     | &#10003; | &#10003;       |
| TLS: Add, Remove, Update, Rotate ( [Cert Manager](https://cert-manager.io/docs/) ) | &#10003; | &#10003;       |
| Reconfigurable Health Checker                                                      | &#10003; | &#10003;       |
| Externally manageable Auth Secret                                                  | &#10003; | &#10003;       |
| Pre-Configured JMX Exporter for Metrics                                            | &#10003; | &#10003;       |
| Monitoring with Prometheus & Grafana                                               | &#10003; | &#10003;       |
| Autoscaling (vertically, volume)	                                                  | &#10003; | &#10007;       |
| Custom Volume                                                                      | &#10003; | &#10003;       |
| Persistent Volume                                                                  | &#10003; | &#45;          |
| Connectors                                                                         | &#45;    | &#10003;       |

## Lifecycle of Kafka Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p align="center">
<img alt="lifecycle"  src="/docs/images/kafka/kafka-crd-lifecycle.png">
</p>

## Lifecycle of ConnectCluster Object

<p align="center">
<img alt="lifecycle"  src="/docs/images/kafka/connectcluster/connectcluster-crd-lifecycle.png">
</p>

## Supported Kafka Versions

KubeDB supports The following Kafka versions. Supported version are applicable for Kraft mode or Zookeeper-less releases:
- `3.3.2`
- `3.4.1`
- `3.5.1`
- `3.5.2`
- `3.6.0`
- `3.6.1`

> The listed KafkaVersions are tested and provided as a part of the installation process (ie. catalog chart), but you are open to create your own [KafkaVersion](/docs/guides/kafka/concepts/kafkaversion.md) object with your custom Kafka image.

## Supported KafkaConnector Versions

| Connector Plugin     | Type   | Version     | Connector Class                                            |
|----------------------|--------|-------------|------------------------------------------------------------|
| mongodb-1.11.0       | Source | 1.11.0      | com.mongodb.kafka.connect.MongoSourceConnector             |
| mongodb-1.11.0       | Sink   | 1.11.0      | com.mongodb.kafka.connect.MongoSinkConnector               |
| mysql-2.4.2.final    | Source | 2.4.2.Final | io.debezium.connector.mysql.MySqlConnector                 |
| postgres-2.4.2.final | Source | 2.4.2.Final | io.debezium.connector.postgresql.PostgresConnector         |
| jdbc-2.6.1.final     | Sink   | 2.6.1.Final | io.debezium.connector.jdbc.JdbcSinkConnector               |
| s3-2.15.0            | Sink   | 2.15.0      | io.aiven.kafka.connect.s3.AivenKafkaConnectS3SinkConnector |
| gcs-0.13.0           | Sink   | 0.13.0      | io.aiven.kafka.connect.gcs.GcsSinkConnector                |


## User Guide 
- [Quickstart Kafka](/docs/guides/kafka/quickstart/overview/kafka/index.md) with KubeDB Operator.
- [Quickstart ConnectCluster](/docs/guides/kafka/quickstart/overview/connectcluster/index.md) with KubeDB Operator.
- Kafka Clustering supported by KubeDB
  - [Combined Clustering](/docs/guides/kafka/clustering/combined-cluster/index.md)
  - [Topology Clustering](/docs/guides/kafka/clustering/topology-cluster/index.md)
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus and Grafana](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Detail concepts of [ConnectCluster object](/docs/guides/kafka/concepts/connectcluster.md).
- Detail concepts of [Connector object](/docs/guides/kafka/concepts/connector.md).
- Detail concepts of [KafkaVersion object](/docs/guides/kafka/concepts/kafkaversion.md).
- Detail concepts of [KafkaConnectorVersion object](/docs/guides/kafka/concepts/kafkaconnectorversion.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).