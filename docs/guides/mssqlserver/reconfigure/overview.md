---
title: Reconfiguring MSSQLServer
menu:
  docs_{{ .version }}:
    identifier: ms-reconfigure-overview
    name: Overview
    parent: ms-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring SQL Server

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `MSSQLServer` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)

## How Reconfiguring MSSQLServer Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `MSSQLServer` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of MSSQLServer" src="/docs/images/day-2-operation/mssqlserver/ms-reconfigure.png">
<figcaption align="center">Fig: Reconfiguring process of MSSQLServer</figcaption>
</figure>

The Reconfiguring MSSQLServer process consists of the following steps:

1. At first, a user creates a `MSSQLServer` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `MSSQLServer` CR.

3. When the operator finds a `MSSQLServer` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `MSSQLServer` database the user creates a `MSSQLServerOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `MSSQLServerOpsRequest` CR.

6. When it finds a `MSSQLServerOpsRequest` CR, it halts the `MSSQLServer` object which is referred from the `MSSQLServerOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `MSSQLServer` object during the reconfiguring process.  

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `MSSQLServerOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `MSSQLServerOpsRequest` CR.

9. After the successful reconfiguring of the `MSSQLServer`, the `KubeDB` Ops-manager operator resumes the `MSSQLServer` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on reconfiguring MSSQLServer database using `MSSQLServerOpsRequest` CR.