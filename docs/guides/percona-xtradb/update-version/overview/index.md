---
title: Updating PerconaXtraDB Overview
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-updating-overview
    name: Overview
    parent: guides-perconaxtradb-updating
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# updating PerconaXtraDB version Overview

This guide will give you an overview on how KubeDB Ops Manager update the version of `PerconaXtraDB` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest)

## How update Process Works

The following diagram shows how KubeDB Ops Manager used to update the version of `PerconaXtraDB`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of PerconaXtraDB" src="/docs/guides/percona-xtradb/update-version/overview/images/pxops-update.jpeg">
<figcaption align="center">Fig: updating Process of PerconaXtraDB</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `PerconaXtraDB` Custom Resource (CR).

2. `KubeDB` Community operator watches the `PerconaXtraDB` CR.

3. When the operator finds a `PerconaXtraDB` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the version of the `PerconaXtraDB` database the user creates a `PerconaXtraDBOpsRequest` CR with the desired version.

5. `KubeDB` Enterprise operator watches the `PerconaXtraDBOpsRequest` CR.

6. When it finds a `PerconaXtraDBOpsRequest` CR, it halts the `PerconaXtraDB` object which is referred from the `PerconaXtraDBOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `PerconaXtraDB` object during the updating process.  

7. By looking at the target version from `PerconaXtraDBOpsRequest` CR, `KubeDB` Enterprise operator updates the images of all the `PetSets`. After each image update, the operator performs some checks such as if the oplog is synced and database size is almost same or not.

8. After successfully updating the `PetSets` and their `Pods` images, the `KubeDB` Enterprise operator updates the image of the `PerconaXtraDB` object to reflect the updated state of the database.

9. After successfully updating of `PerconaXtraDB` object, the `KubeDB` Enterprise operator resumes the `PerconaXtraDB` object so that the `KubeDB` Community operator can resume its usual operations.

In the next doc, we are going to show a step by step guide on updating of a PerconaXtraDB database using update operation.