---
title: MSSQLServer Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: ms-volume-expansion-overview
    name: Overview
    parent: ms-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServer Volume Expansion

This guide will give an overview on how KubeDB Ops Manager expand the volume of `MSSQLServer`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/redis/concepts/redis.md)
  - [MSSQLServerOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops Manager expand the volumes of `MSSQLServer` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of MSSQLServer" src="/docs/images/day-2-operation/redis/ms-volume-expansion.svg">
<figcaption align="center">Fig: Volume Expansion process of MSSQLServer</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `MSSQLServer` Custom Resource (CR).

2. `KubeDB` Community operator watches the `MSSQLServer` CR.

3. When the operator finds a `MSSQLServer` CR, it creates required `PetSet` and related necessary stuff like secrets, services, etc.

4. The petSet creates Persistent Volumes according to the Volume Claim Template provided in the petset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, in order to expand the volume of the `MSSQLServer` database the user creates a `MSSQLServerOpsRequest` CR with desired information.

6. `KubeDB` Enterprise operator watches the `MSSQLServerOpsRequest` CR.

7. When it finds a `MSSQLServerOpsRequest` CR, it pauses the `MSSQLServer` object which is referred from the `MSSQLServerOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `MSSQLServer` object during the volume expansion process.

8. Then the `KubeDB` Enterprise operator will expand the persistent volume to reach the expected size defined in the `MSSQLServerOpsRequest` CR.

9. After the successful expansion of the volume of the related PetSet Pods, the `KubeDB` Enterprise operator updates the new volume size in the `MSSQLServer` object to reflect the updated state.

10. After the successful Volume Expansion of the `MSSQLServer`, the `KubeDB` Enterprise operator resumes the `MSSQLServer` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on Volume Expansion of various MSSQLServer database using `MSSQLServerOpsRequest` CRD.
