---
title: Druid Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: guides-druid-volume-expansion-overview
    name: Overview
    parent: guides-druid-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid Volume Expansion

This guide will give an overview on how KubeDB Ops-manager operator expand the volume of various component of `Druid` like:. (Combined and Topology).

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops-manager operator expand the volumes of `Druid` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of Druid" src="/docs/guides/druid/volume-expansion/images/druid-volume-expansion.png">
<figcaption align="center">Fig: Volume Expansion process of Druid</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `Druid` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Druid` CR.

3. When the operator finds a `Druid` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the petset configuration. This Persistent Volume will be expanded by the `KubeDB` Ops-manager operator.

5. Then, in order to expand the volume of the druid data components (ie. Historicals, MiddleManagers) of the `Druid`, the user creates a `DruidOpsRequest` CR with desired information.

6. `KubeDB` Ops-manager operator watches the `DruidOpsRequest` CR.

7. When it finds a `DruidOpsRequest` CR, it halts the `Druid` object which is referred from the `DruidOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Druid` object during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator will expand the persistent volume to reach the expected size defined in the `DruidOpsRequest` CR.

9. After the successful Volume Expansion of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the new volume size in the `Druid` object to reflect the updated state.

10. After the successful Volume Expansion of the `Druid` components, the `KubeDB` Ops-manager operator resumes the `Druid` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on Volume Expansion of various Druid database components using `DruidOpsRequest` CRD.
