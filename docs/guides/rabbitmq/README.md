---
title: RabbitMQ
menu:
  docs_{{ .version }}:
    identifier: rm-guides-readme
    name: RabbitMQ
    parent: rm-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/rabbitmq/
aliases:
  - /docs/{{ .version }}/guides/rabbitmq/README/
---
> New to KubeDB? Please start [here](/docs/README.md).

# Overview 

RabbitMQ is a robust and flexible open-source message broker software that facilitates communication between distributed applications. It implements the Advanced Message Queuing Protocol (AMQP) standard, ensuring reliable messaging across various platforms and languages. With its support for multiple messaging protocols (MQTT, STOMP etc.) and delivery patterns (Fanout, Direct, Exchange etc.), RabbitMQ enables seamless integration and scalability for modern microservices architectures. It provides features such as message persistence, clustering, and high availability, making it a preferred choice for handling asynchronous communication and decoupling components in enterprise systems.

## Supported RabbitMQ Features

| Features                                                        | Availability |
|-----------------------------------------------------------------|:------------:|
| Clustering                                                      |   &#10003;   |
| Custom Configuration (Config File, PodTemplate, Plugins)        |   &#10003;   |
| Monitoring with Prometheus & Grafana (Builtin & Operator)       |   &#10003;   |
| Authentication & Authorization (Managed Auth Secret, Rotation)  |   &#10003;   |
| TLS/SSL Encryption (Configure & Reconfigure)                    |   &#10003;   |
| Persistent Volume                                               |   &#10003;   |
| Pre-Enabled Dashboard (Management UI)                           |   &#10003;   |
| Pre-Enabled Utility Plugins (Shovel, Federation)                |   &#10003;   |
| Pre-Enabled Protocols via Web Dispatch (AMQP, MQTT, STOMP)      |   &#10003;   |
| Horizontal & Vertical Scaling                                   |   &#10003;   |
| Volume Expansion                                                |   &#10003;   |
| Autoscaling (Compute & Storage)                                 |   &#10003;   |
| Reconfigure Cluster                                             |   &#10003;   |
| Reconfigurable Health Checker                                   |   &#10003;   |
| Version Update                                                  |   &#10003;   |
| Rolling Restart                                                 |   &#10003;   |

## Supported RabbitMQ Versions

KubeDB supports the following RabbitMQ Versions.
- `3.12.12`
- `3.13.2`
- `4.0.4`

## Life Cycle of a RabbitMQ Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p text-align="center">
    <img alt="lifecycle"  src="/docs/guides/rabbitmq/images/rabbitmq-lifecycle.png" >
</p>

## User Guide

- [Quickstart RabbitMQ](/docs/guides/rabbitmq/quickstart/quickstart.md) with KubeDB Operator.
- [Run RabbitMQ with Custom Configuration](/docs/guides/rabbitmq/configuration/using-config-file.md)
- [Run RabbitMQ with Custom PodTemplate](/docs/guides/rabbitmq/configuration/using-podtemplate.md)
- [Monitor RabbitMQ with Builtin Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md)
- [Monitor RabbitMQ with Prometheus Operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md)
- [Configure TLS/SSL for RabbitMQ](/docs/guides/rabbitmq/tls/tls.md)
- [Reconfigure RabbitMQ TLS/SSL Encryption](/docs/guides/rabbitmq/reconfigure-tls/reconfigure-tls.md)
- [Reconfigure RabbitMQ Cluster](/docs/guides/rabbitmq/reconfigure/reconfigure.md)
- [Horizontal Scale RabbitMQ](/docs/guides/rabbitmq/scaling/horizontal-scaling/horizontal-scaling.md)
- [Vertical Scale RabbitMQ](/docs/guides/rabbitmq/scaling/vertical-scaling/vertical-scaling.md)
- [Expand Volume of RabbitMQ](/docs/guides/rabbitmq/volume-expansion/volume-expansion.md)
- [Autoscale RabbitMQ Compute Resources](/docs/guides/rabbitmq/autoscaler/compute/compute-autoscale.md)
- [Autoscale RabbitMQ Storage](/docs/guides/rabbitmq/autoscaler/storage/storage-autoscale.md)
- [Update RabbitMQ Version](/docs/guides/rabbitmq/update-version/update-version.md)
- [Restart RabbitMQ](/docs/guides/rabbitmq/restart/restart.md)
- [Rotate RabbitMQ Authentication Credentials](/docs/guides/rabbitmq/rotate-auth/guide.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).