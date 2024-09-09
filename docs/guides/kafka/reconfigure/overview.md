---
title: Reconfiguring Kafka
menu:
  docs_{{ .version }}:
    identifier: kf-reconfigure-overview
    name: Overview
    parent: kf-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Kafka

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `Kafka` components such as Combined, Broker, Controller, etc.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How Reconfiguring Kafka Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `Kafka` components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of Kafka" src="/docs/images/day-2-operation/kafka/kf-reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of Kafka</figcaption>
</figure>

The Reconfiguring Kafka process consists of the following steps:

1. At first, a user creates a `Kafka` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Kafka` CR.

3. When the operator finds a `Kafka` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the various components (ie. Combined, Broker) of the `Kafka`, the user creates a `KafkaOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `KafkaOpsRequest` CR.

6. When it finds a `KafkaOpsRequest` CR, it halts the `Kafka` object which is referred from the `KafkaOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Kafka` object during the reconfiguring process.

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `MogoDBOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `KafkaOpsRequest` CR.

9. After the successful reconfiguring of the `Kafka` components, the `KubeDB` Ops-manager operator resumes the `Kafka` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring Kafka components using `KafkaOpsRequest` CRD.