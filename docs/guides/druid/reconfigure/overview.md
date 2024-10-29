---
title: Reconfiguring Druid
menu:
  docs_{{ .version }}:
    identifier: guides-druid-reconfigure-overview
    name: Overview
    parent: guides-druid-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Druid

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `Druid` components such as Combined, Broker, Controller, etc.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/kafka/concepts/kafka.md)
    - [DruidOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How Reconfiguring Druid Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `Druid` components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of Druid" src="/docs/guides/druid/reconfigure/images/reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of Druid</figcaption>
</figure>

The Reconfiguring Druid process consists of the following steps:

1. At first, a user creates a `Druid` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Druid` CR.

3. When the operator finds a `Druid` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the various components (ie. Coordinators, Overlords, Historicals, MiddleManagers, Brokers, Routers) of the `Druid`, the user creates a `DruidOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `DruidOpsRequest` CR.

6. When it finds a `DruidOpsRequest` CR, it halts the `Druid` object which is referred from the `DruidOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Druid` object during the reconfiguring process.

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `MogoDBOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `DruidOpsRequest` CR.

9. After the successful reconfiguring of the `Druid` components, the `KubeDB` Ops-manager operator resumes the `Druid` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on reconfiguring Druid components using `DruidOpsRequest` CRD.