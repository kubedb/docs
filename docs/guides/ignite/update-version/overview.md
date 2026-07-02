---
title: Updating Ignite Overview
menu:
  docs_{{ .version }}:
    identifier: ig-update-version-overview
    name: Overview
    parent: ig-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview of Ignite Version Update

This guide will give you an overview on how KubeDB Ops-manager operator update the version of `Ignite` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Ignite](/docs/guides/ignite/concepts/ignite/index.md)
    - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest/index.md)

## How update version Process Works

The following diagram shows how KubeDB Ops-manager operator used to update the version of `Ignite`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of Ignite" src="/docs/images/day-2-operation/ignite/ig-version-update.svg">
<figcaption align="center">Fig: updating Process of Ignite</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates an `Ignite` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `Ignite` CR.

3. When the operator finds an `Ignite` CR, it creates required number of `PetSets` and other kubernetes native resources like secrets, services, etc.

4. Then, in order to update the version of the `Ignite` database the user creates an `IgniteOpsRequest` CR with the desired version.

5. `KubeDB` Ops-manager operator watches the `IgniteOpsRequest` CR.

6. When it finds an `IgniteOpsRequest` CR, it halts the `Ignite` object which is referred from the `IgniteOpsRequest`. So, the `KubeDB` Provisioner operator doesn't perform any operations on the `Ignite` object during the updating process.

7. By looking at the target version from `IgniteOpsRequest` CR, `KubeDB` Ops-manager operator updates the images of all the `PetSets`.

8. After successfully updating the `PetSets` and their `Pods` images, the `KubeDB` Ops-manager operator updates the version field of the `Ignite` object to reflect the updated state of the database.

9. After successfully updating of `Ignite` object, the `KubeDB` Ops-manager operator resumes the `Ignite` object so that the `KubeDB` Provisioner operator can resume its usual operations.

In the [next](/docs/guides/ignite/update-version/update-version.md) doc, we are going to show a step-by-step guide on updating of an Ignite database using updateVersion operation.
