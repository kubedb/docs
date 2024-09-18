---
title: RabbitMQ
menu:
  docs_{{ .version }}:
    identifier: guides-rabbitmq-readme
    name: RabbitMQ
    parent: guides-rabbitmq
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

| Features                                                      | Availability |
|---------------------------------------------------------------|:------------:|
| Clustering                                                    |   &#10003;   |
| Custom Configuration                                          |   &#10003;   |
| Custom Plugin configurations                                  |   &#10003;   |
| Monitoring using Prometheus and Grafana                       |   &#10003;   |
| Builtin Prometheus Discovery                                  |   &#10003;   |
| Operator managed Prometheus Discovery                         |   &#10003;   |
| Authentication & Authorization (TLS)                          |   &#10003;   |
| Externally manageable Auth Secret                             |   &#10003;   |
| Persistent volume                                             |   &#10003;   |
| Grafana Dashboards (Alerts and Monitoring)                    |   &#10003;   |
| Pre-Enabled Dashboard ( Management UI )                       |   &#10003;   |
| Pre-Enabled utility plugins ( Shovel, Federation )            |   &#10003;   |
| Pre-Enabled Protocols with web dispatch ( AMQP, MQTT, STOMP ) |   &#10003;   |
| Automated Vertical & Horizontal Scaling                       |   &#10003;   |
| Automated Volume Expansion                                    |   &#10003;   |
| Autoscaling ( Compute resources & Storage )                   |   &#10003;   |
| Reconfigurable Health Checker                                 |   &#10003;   |
| Reconfigurable TLS Certificates (Add, Remove, Rotate, Update) |   &#10003;   |

## Supported RabbitMQ Versions

KubeDB supports the following RabbitMQ Versions.
- `3.12.12`
- `3.13.2`

## Life Cycle of a RabbitMQ Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p text-align="center">
    <img alt="lifecycle"  src="/docs/guides/rabbitmq/images/rabbitmq-lifecycle.png" >
</p>

## User Guide

- [Quickstart RabbitMQ](/docs/guides/rabbitmq/quickstart/quickstart.md) with KubeDB Operator.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).