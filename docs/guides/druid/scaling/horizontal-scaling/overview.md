---
title: Druid Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-druid-scaling-horizontal-scaling-overview
    name: Overview
    parent: guides-druid-scaling-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid Horizontal Scaling

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Druid` cluster replicas of various component such as Coordinators, Overlords, Historicals, MiddleManager, Brokers and Routers.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator scales up or down `Druid` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of Druid" src="/docs/guides/druid/scaling/horizontal-scaling/images/dr-horizontal-scaling.png" width="1000" height="350">
  <figcaption align="center">Fig: Horizontal scaling process of Druid</figcaption>
</figure>

The Horizontal scaling process consists of the following steps:

1. At first, a user creates a `Druid` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Druid` CR.

3. When the operator finds a `Druid` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to scale the various components (i.e. Coordinators, Overlords, Historicals, MiddleManagers, Brokers, Routers) of the `Druid` cluster, the user creates a `DruidOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `DruidOpsRequest` CR.

6. When it finds a `DruidOpsRequest` CR, it halts the `Druid` object which is referred from the `DruidOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Druid` object during the horizontal scaling process.

7. Then the `KubeDB` Ops-manager operator will scale the related PetSet Pods to reach the expected number of replicas defined in the `DruidOpsRequest` CR.

8. After the successfully scaling the replicas of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the number of replicas in the `Druid` object to reflect the updated state.

9. After the successful scaling of the `Druid` replicas, the `KubeDB` Ops-manager operator resumes the `Druid` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on horizontal scaling of Druid cluster using `DruidOpsRequest` CRD.