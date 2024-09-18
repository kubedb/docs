---
title: Reconfiguring RabbitMQ
menu:
  docs_{{ .version }}:
    identifier: rm-reconfigure-overview
    name: Overview
    parent: rm-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring RabbitMQ

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `RabbitMQ` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)

## How does Reconfiguring RabbitMQ Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `RabbitMQ` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of RabbitMQ" src="/docs/guides/rabbitmq/images/reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of RabbitMQ</figcaption>
</figure>

The Reconfiguring RabbitMQ process consists of the following steps:

1. At first, a user creates a `RabbitMQ` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `RabbitMQ` CR.

3. When the operator finds a `RabbitMQ` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `RabbitMQ` database the user creates a `RabbitMQOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `RabbitMQOpsRequest` CR.

6. When it finds a `RabbitMQOpsRequest` CR, it halts the `RabbitMQ` object which is referred from the `RabbitMQOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `RabbitMQ` object during the reconfiguring process.  

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `MogoDBOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related StatefulSet Pods so that they restart with the new configuration defined in the `RabbitMQOpsRequest` CR.

9. After the successful reconfiguring of the `RabbitMQ` components, the `KubeDB` Ops-manager operator resumes the `RabbitMQ` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring RabbitMQ database components using `RabbitMQOpsRequest` CRD.