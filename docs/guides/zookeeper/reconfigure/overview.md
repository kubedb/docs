---
title: Reconfiguring ZooKeeper
menu:
  docs_{{ .version }}:
    identifier: zk-reconfigure-overview
    name: Overview
    parent: zk-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring ZooKeeper

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `ZooKeeper` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
  - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)

## How does Reconfiguring ZooKeeper Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `ZooKeeper` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of ZooKeeper" src="/docs/images/day-2-operation/zookeeper/zk-reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of ZooKeeper</figcaption>
</figure>

The Reconfiguring ZooKeeper process consists of the following steps:

1. At first, a user creates a `ZooKeeper` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `ZooKeeper` CR.

3. When the operator finds a `ZooKeeper` CR, it creates required number of `Petsets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `ZooKeeper` database the user creates a `ZooKeeperOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ZooKeeperOpsRequest` CR.

6. When it finds a `ZooKeeperOpsRequest` CR, it halts the `ZooKeeper` object which is referred from the `ZooKeeperOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `ZooKeeper` object during the reconfiguring process.  

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `ZooKeeperOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related Petset Pods so that they restart with the new configuration defined in the `ZooKeeperOpsRequest` CR.

9. After the successful reconfiguring of the `ZooKeeper` components, the `KubeDB` Ops-manager operator resumes the `ZooKeeper` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the [next](/docs/guides/zookeeper/reconfigure/reconfigure.md) docs, we are going to show a step by step guide on reconfiguring ZooKeeper database components using `ZooKeeperOpsRequest` CRD.