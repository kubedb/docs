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


| Features                                                       | Community | Enterprise |
|----------------------------------------------------------------|:---------:|:----------:|
| Clustering - Combined (shared controller and broker nodes)     | &#10003;  |  &#10003;  |
| Clustering - Topology (dedicated controllers and broker nodes) | &#10003;  |  &#10003;  |
| Custom Docker Image                                            | &#10003;  |  &#10003;  |
| Authentication & Authorization                                 | &#10003;  |  &#10003;  |
| Persistent Volume                                              | &#10003;  |  &#10003;  |
| Custom Volume                                                  | &#10003;  |  &#10003;  |
| TLS: using ( [Cert Manager](https://cert-manager.io/docs/) )   | &#10007;  |  &#10003;  |
| Reconfigurable Health Checker                                  | &#10003;  |  &#10003;  |
| Externally manageable Auth Secret                              | &#10003;  |  &#10003;  |

## Supported Kafka Versions

KubeDB supports The following Kafka versions. Supported version are applicable for Kraft mode or Zookeeper-less releases:
- `3.3.0`

> The listed KafkaVersions are tested and provided as a part of the installation process (ie. catalog chart), but you are open to create your own [KafkaVersion](/docs/guides/kafka/concepts/catalog/index.md) object with your custom Kafka image.

## Lifecycle of Kafka Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p align="center">
<img alt="lifecycle"  src="/docs/guides/kafka/images/Kafka-CRD-Lifecycle.png">
</p>

## User Guide 
- [Quickstart Kafka](/docs/guides/kafka/quickstart/overview/index.md) with KubeDB Operator.
- Kafka Clustering supported by KubeDB
  - [Combined Clustering](/docs/guides/kafka/clustering/combined-cluster/index.md)
  - [Topology Clustering](/docs/guides/kafka/clustering/topology-cluster/index.md)
- Use [kubedb cli](/docs/guides/elasticsearch/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).