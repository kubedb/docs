---
title: Reconfiguring Ignite
menu:
  docs_{{ .version }}:
    identifier: ig-reconfigure-overview
    name: Overview
    parent: ig-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Ignite

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `Ignite` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)

## How does Reconfiguring Ignite Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `Ignite` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of Ignite" src="/docs/guides/ignite/images/reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of Ignite</figcaption>
</figure>

The Reconfiguring Ignite process consists of the following steps:

1. At first, a user creates a `Ignite` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Ignite` CR.

3. When the operator finds a `Ignite` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `Ignite` database the user creates a `IgniteOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `IgniteOpsRequest` CR.

6. When it finds a `IgniteOpsRequest` CR, it halts the `Ignite` object which is referred from the `IgniteOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Ignite` object during the reconfiguring process.  

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `MogoDBOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `IgniteOpsRequest` CR.

9. After the successful reconfiguring of the `Ignite` components, the `KubeDB` Ops-manager operator resumes the `Ignite` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring Ignite database components using `IgniteOpsRequest` CRD.