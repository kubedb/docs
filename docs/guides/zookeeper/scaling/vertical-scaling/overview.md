---
title: ZooKeeper Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: zk-vertical-scaling-overview
    name: Overview
    parent: zk-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ZooKeeper Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `ZooKeeper` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
    - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `ZooKeeper` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of ZooKeeper" src="/docs/images/day-2-operation/zookeeper/zk-vertical-scaling.svg">
<figcaption align="center">Fig: Vertical scaling process of ZooKeeper</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `ZooKeeper` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `ZooKeeper` CR.

3. When the operator finds a `ZooKeeper` CR, it creates required number of `Petsets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `ZooKeeper` database the user creates a `ZooKeeperOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ZooKeeperOpsRequest` CR.

6. When it finds a `ZooKeeperOpsRequest` CR, it halts the `ZooKeeper` object which is referred from the `ZooKeeperOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `ZooKeeper` object during the vertical scaling process.

7. Then the `KubeDB` Ops-manager operator will update resources of the Petset Pods to reach desired state.

8. After the successful update of the resources of the Petset's replica, the `KubeDB` Ops-manager operator updates the `ZooKeeper` object to reflect the updated state.

9. After the successful update  of the `ZooKeeper` resources, the `KubeDB` Ops-manager operator resumes the `ZooKeeper` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the [next](/docs/guides/zookeeper/scaling/vertical-scaling/vertical-scaling.md) docs, we are going to show a step by step guide on updating resources of ZooKeeper database using `ZooKeeperOpsRequest` CRD.