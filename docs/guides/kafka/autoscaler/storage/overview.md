---
title: Kafka Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: mg-storage-auto-scaling-overview
    name: Overview
    parent: mg-storage-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `kafkaautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [KafkaAutoscaler](/docs/guides/kafka/concepts/kafkaautoscaler.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Kafka` cluster components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Storage Auto Scaling process of Kafka" src="/docs/images/mongodb/storage-process.svg">
<figcaption align="center">Fig: Storage Auto Scaling process of Kafka</figcaption>
</figure>


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Kafka` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Kafka` CR.

3. When the operator finds a `Kafka` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

- Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the petset configuration.

4. Then, in order to set up storage autoscaling of the various components (ie. Combined, Broker, Controller.) of the `Kafka` cluster, the user creates a `KafkaAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `KafkaAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the clusters to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `KafkaOpsRequest` to expand the storage of the database.

7. `KubeDB` Ops-manager operator watches the `KafkaOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the cluster component as specified on the `KafkaOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various Kafka cluster components using `KafkaAutoscaler` CRD.
