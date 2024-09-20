---
title: RabbitMQ Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: rm-autoscaling-compute-overview
    name: Overview
    parent: rm-autoscaling-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQ Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `RabbitMQautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQAutoscaler](/docs/guides/rabbitmq/concepts/autoscaler.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `RabbitMQ` database components. Open the image in a new tab to see the enlarged version.


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `RabbitMQ` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `RabbitMQ` CRO.

3. When the operator finds a `RabbitMQ` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the of the `RabbitMQ` cluster the user creates a `RabbitMQAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `RabbitMQAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `RabbitMQAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `RabbitMQOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `RabbitMQOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `RabbitMQOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of various RabbitMQ database components using `RabbitMQAutoscaler` CRD.
