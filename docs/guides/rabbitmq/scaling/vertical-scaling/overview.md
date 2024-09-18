---
title: RabbitMQ Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: rm-vertical-scaling-overview
    name: Overview
    parent: rm-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQ Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `RabbitMQ` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `RabbitMQ` database. Open the image in a new tab to see the enlarged version.

The vertical scaling process consists of the following steps:

1. At first, a user creates a `RabbitMQ` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `RabbitMQ` CR.

3. When the operator finds a `RabbitMQ` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `RabbitMQ` database the user creates a `RabbitMQOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `RabbitMQOpsRequest` CR.

6. When it finds a `RabbitMQOpsRequest` CR, it halts the `RabbitMQ` object which is referred from the `RabbitMQOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `RabbitMQ` object during the vertical scaling process.  

7. Then the `KubeDB` Ops-manager operator will update resources of the StatefulSet Pods to reach desired state.

8. After the successful update of the resources of the StatefulSet's replica, the `KubeDB` Ops-manager operator updates the `RabbitMQ` object to reflect the updated state.

9. After the successful update  of the `RabbitMQ` resources, the `KubeDB` Ops-manager operator resumes the `RabbitMQ` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of RabbitMQ database using `RabbitMQOpsRequest` CRD.