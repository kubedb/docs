---
title: RabbitMQ Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: rm-volume-expansion-overview
    name: Overview
    parent: rm-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQ Volume Expansion

This guide will give an overview on how KubeDB Ops-manager operator expand the volume of `RabbitMQ` cluster nodes.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops-manager operator expand the volumes of `RabbitMQ` database components. Open the image in a new tab to see the enlarged version.

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `RabbitMQ` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `RabbitMQ` CR.

3. When the operator finds a `RabbitMQ` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Each StatefulSet creates a Persistent Volume according to the Volume Claim Template provided in the PetSet configuration. This Persistent Volume will be expanded by the `KubeDB` Ops-manager operator.

5. Then, in order to expand the volume the `RabbitMQ` database the user creates a `RabbitMQOpsRequest` CR with desired information.

6. `KubeDB` Ops-manager operator watches the `RabbitMQOpsRequest` CR.

7. When it finds a `RabbitMQOpsRequest` CR, it halts the `RabbitMQ` object which is referred from the `RabbitMQOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `RabbitMQ` object during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator will expand the persistent volume to reach the expected size defined in the `RabbitMQOpsRequest` CR.

9. After the successful Volume Expansion of the related StatefulSet Pods, the `KubeDB` Ops-manager operator updates the new volume size in the `RabbitMQ` object to reflect the updated state.

10. After the successful Volume Expansion of the `RabbitMQ` components, the `KubeDB` Ops-manager operator resumes the `RabbitMQ` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on Volume Expansion of various RabbitMQ database components using `RabbitMQOpsRequest` CRD.
