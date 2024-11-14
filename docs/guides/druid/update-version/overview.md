---
title: Update Version Overview
menu:
  docs_{{ .version }}:
    identifier: guides-druid-update-version-overview
    name: Overview
    parent: guides-druid-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Druid Update Version Overview

This guide will give you an overview on how KubeDB Ops-manager operator update the version of `Druid`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)

## How update version Process Works

The following diagram shows how KubeDB Ops-manager operator used to update the version of `Druid`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of Druid" src="/docs/guides/druid/update-version/images/dr-update-version.png">
<figcaption align="center">Fig: updating Process of Druid</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Druid` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Druid` CR.

3. When the operator finds a `Druid` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the version of the `Druid` database the user creates a `DruidOpsRequest` CR with the desired version.

5. `KubeDB` Ops-manager operator watches the `DruidOpsRequest` CR.

6. When it finds a `DruidOpsRequest` CR, it halts the `Druid` object which is referred from the `DruidOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Druid` object during the updating process.

7. By looking at the target version from `DruidOpsRequest` CR, `KubeDB` Ops-manager operator updates the images of all the `PetSets`.

8. After successfully updating the `PetSets` and their `Pods` images, the `KubeDB` Ops-manager operator updates the image of the `Druid` object to reflect the updated state of the database.

9. After successfully updating of `Druid` object, the `KubeDB` Ops-manager operator resumes the `Druid` object so that the `KubeDB` Provisioner  operator can resume its usual operations.

In the next doc, we are going to show a step by step guide on updating of a Druid database using updateVersion operation.