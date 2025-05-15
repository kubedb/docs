---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: kf-rotate-auth-overview
    name: Overview
    parent: kf-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Kafka

This guide will give an overview on how KubeDB Ops-manager operator Rotate Authentication configuration.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How Reconfiguring Kafka TLS Configuration Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures TLS of a `Kafka`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of Kafka" src="/docs/images/day-2-operation/kafka/kf-rotate-auth.svg">
<figcaption align="center">Fig: Rotate Auth process of Kafka</figcaption>
</figure>

The Rotate Kafka Authentication process consists of the following steps:

1. At first, a user creates a `Kafka` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Kafka` CRO.

3. When the operator finds a `Kafka` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to rotate the Authentication configuration of the `Kafka`, the user creates a `KafkaOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `KafkaOpsRequest` CR.

6. When it finds a `KafkaOpsRequest` CR, it pauses the `Kafka` object which is referred from the `KafkaOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Kafka` object during the rotating Authentication process.

7. Then the `KubeDB` Ops-manager operator will update necessary configuration based on the Ops Request yaml to update credentials.

8. Then the `KubeDB` Ops-manager operator will restart all the Pods of the database so that they restart with the new Authentication ENVs or other configuration defined in the `KafkaOpsRequest` CR.

9. After the successful rotating of the `Kafka` Authentication, the `KubeDB` Ops-manager operator resumes the `Kafka` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on rotating Authentication configuration of a Kafka using `KafkaOpsRequest` CRD.