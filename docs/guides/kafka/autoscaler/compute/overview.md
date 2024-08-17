---
title: Kafka Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: mg-auto-scaling-overview
    name: Overview
    parent: mg-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `kafkaautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Kafka](/docs/guides/mongodb/concepts/mongodb.md)
  - [KafkaAutoscaler](/docs/guides/kafka/concepts/autoscaler.md)
  - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Kafka` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of Kafka" src="/docs/images/mongodb/compute-process.svg">
<figcaption align="center">Fig: Compute Auto Scaling process of Kafka</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Kafka` Custom Resource Object (CRO).

2. `KubeDB` Provisioner operator watches the `Kafka` CRO.

3. When the operator finds a `Kafka` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the various components (ie. Combined, Broker, Controller) of the `Kafka` cluster the user creates a `KafkaAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `KafkaAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `KafkaAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `KafkaOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `KafkaOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `KafkaOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of various Kafka database components using `KafkaAutoscaler` CRD.
