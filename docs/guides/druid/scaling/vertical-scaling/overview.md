---
title: Druid Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-druid-scaling-vertical-scaling-overview
    name: Overview
    parent: guides-druid-scaling-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `Druid`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/kafka/concepts/kafka.md)
    - [DruidOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `Druid`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of Druid" src="/docs/guides/druid/scaling/horizontal-scaling/images/dr-horizontal-scaling.png">
<figcaption align="center">Fig: Vertical scaling process of Druid</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `Druid` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Druid` CR.

3. When the operator finds a `Druid` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `Druid` cluster, the user creates a `DruidOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `DruidOpsRequest` CR.

6. When it finds a `DruidOpsRequest` CR, it halts the `Druid` object which is referred from the `DruidOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Druid` object during the vertical scaling process.

7. Then the `KubeDB` Ops-manager operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Ops-manager operator updates the `Druid` object to reflect the updated state.

9. After the successful update  of the `Druid` resources, the `KubeDB` Ops-manager operator resumes the `Druid` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of Druid database using `DruidOpsRequest` CRD.