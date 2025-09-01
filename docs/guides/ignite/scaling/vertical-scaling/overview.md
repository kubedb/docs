---
title: Ignite Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: ig-vertical-scaling-overview
    name: Overview
    parent: ig-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ignite Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `Ignite` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `Ignite`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of Ignite" src="/docs">
<figcaption align="center">Fig: Vertical scaling process of Ignite</figcaption>
</figure>


The vertical scaling process consists of the following steps:

1. At first, a user creates a `Ignite` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Ignite` CR.

3. When the operator finds a `Ignite` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `Ignite` database the user creates a `IgniteOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `IgniteOpsRequest` CR.

6. When it finds a `IgniteOpsRequest` CR, it halts the `Ignite` object which is referred from the `IgniteOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Ignite` object during the vertical scaling process.  

7. Then the `KubeDB` Ops-manager operator will update resources of the StatefulSet Pods to reach desired state.

8. After the successful update of the resources of the StatefulSet's replica, the `KubeDB` Ops-manager operator updates the `Ignite` object to reflect the updated state.

9. After the successful update  of the `Ignite` resources, the `KubeDB` Ops-manager operator resumes the `Ignite` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of Ignite database using `IgniteOpsRequest` CRD.