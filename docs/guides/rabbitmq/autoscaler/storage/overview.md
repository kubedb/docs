---
title: RabbitMQ Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: rm-autoscaling-storage-overview
    name: Overview
    parent: rm-autoscaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQ Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `RabbitMQautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQAutoscaler](/docs/guides/rabbitmq/concepts/autoscaler.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `RabbitMQ` database components. Open the image in a new tab to see the enlarged version.


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `RabbitMQ` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `RabbitMQ` CR.

3. When the operator finds a `RabbitMQ` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

- Each StatefulSet creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration.

4. Then, in order to set up storage autoscaling of the `RabbitMQ` cluster, the user creates a `RabbitMQAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `RabbitMQAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `RabbitMQOpsRequest` to expand the storage of the database. 
   
7. `KubeDB` Ops-manager operator watches the `RabbitMQOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the database component as specified on the `RabbitMQOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various RabbitMQ database components using `RabbitMQAutoscaler` CRD.
