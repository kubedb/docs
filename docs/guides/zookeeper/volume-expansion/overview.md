---
title: ZooKeeper Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: zk-volume-expansion-overview
    name: Overview
    parent: zk-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ZooKeeper Volume Expansion

This guide will give an overview on how KubeDB Ops-manager operator expand the volume of `ZooKeeper` cluster nodes.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
    - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops-manager operator expand the volumes of `ZooKeeper` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of ZooKeeper" src="/docs/images/day-2-operation/zookeeper/zk-volume-expansion.svg">
<figcaption align="center">Fig: Volume Expansion process of ZooKeeper</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `ZooKeeper` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `ZooKeeper` CR.

3. When the operator finds a `ZooKeeper` CR, it creates required number of `Petsets` and related necessary stuff like secrets, services, etc.

4. Each petset creates a Persistent Volume according to the Volume Claim Template provided in the PetSet configuration. This Persistent Volume will be expanded by the `KubeDB` Ops-manager operator.

5. Then, in order to expand the volume the `ZooKeeper` database the user creates a `ZooKeeperOpsRequest` CR with desired information.

6. `KubeDB` Ops-manager operator watches the `ZooKeeperOpsRequest` CR.

7. When it finds a `ZooKeeperOpsRequest` CR, it halts the `ZooKeeper` object which is referred from the `ZooKeeperOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `ZooKeeper` object during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator will expand the persistent volume to reach the expected size defined in the `ZooKeeperOpsRequest` CR.

9. After the successful Volume Expansion of the related Petset Pods, the `KubeDB` Ops-manager operator updates the new volume size in the `ZooKeeper` object to reflect the updated state.

10. After the successful Volume Expansion of the `ZooKeeper` components, the `KubeDB` Ops-manager operator resumes the `ZooKeeper` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the [next](/docs/guides/zookeeper/volume-expansion/volume-expansion.md) docs, we are going to show a step-by-step guide on Volume Expansion of various ZooKeeper database components using `ZooKeeperOpsRequest` CRD.
