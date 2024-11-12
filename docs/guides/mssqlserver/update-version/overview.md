---
title: Updating MSSQLServer Overview
menu:
  docs_{{ .version }}:
    identifier: ms-updating-overview
    name: Overview
    parent: mssql-updating
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating MSSQLServer version Overview

This guide will give you an overview on how KubeDB Ops-manager operator update the version of `MSSQLServer` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)

## How update version Process Works

The following diagram shows how KubeDB Ops-manager operator update the version of `MSSQLServer`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of MSSQLServer" src="/docs/images/day-2-operation/mssqlserver/ms-update-version.png">
<figcaption align="center">Fig: Updating Process of MSSQLServer</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `MSSQLServer` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `MSSQLServer` CR.

3. When the operator finds a `MSSQLServer` CR, it creates `PetSets` and related necessary resources like secrets, services, etc.

4. Then, in order to update the version of the `MSSQLServer` the user creates a `MSSQLServerOpsRequest` CR with the desired version.

5. `KubeDB` Ops-manager operator watches the `MSSQLServerOpsRequest` CR.

6. When it finds a `MSSQLServerOpsRequest` CR, it halts the `MSSQLServer` object which is referred from the `MSSQLServerOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `MSSQLServer` object during the updating process.  

7. By looking at the target version from `MSSQLServerOpsRequest` CR, `KubeDB` Ops-manager operator updates the images of the `PetSet`.

8. After successfully updating the `PetSet` and their `Pods` images, the `KubeDB` Ops-manager operator updates the image of the `MSSQLServer` object to reflect the updated state of the database.

9. After successfully updating of `MSSQLServer` object, the `KubeDB` Ops-manager operator resumes the `MSSQLServer` object so that the `KubeDB` Provisioner  operator can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating of a MSSQLServer using UpdateVersion operation.