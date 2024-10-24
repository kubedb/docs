---
title: Microsoft SQL Server Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: ms-scaling-vertical-overview
    name: Overview
    parent: ms-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling MSSQLServer

This guide will give you an overview of how KubeDB Ops Manager updates the resources(for example Memory, CPU etc.) of the `MSSQLServer`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how the `KubeDB` Ops Manager used to update the resources of the `MSSQLServer`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of MSSQLServer" src="/docs/images/day-2-operation/mssqlserver/ms-vertical-scaling.svg">
<figcaption align="center">Fig: Vertical scaling process of MSSQLServer</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `MSSQLServer` CR.

2. `KubeDB` community operator watches for the `MSSQLServer` CR.

3. When it finds one, it creates a `PetSet` and related necessary stuff like secret, service, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `MSSQLServer` database the user creates a `MSSQLServerOpsRequest` CR.

5. `KubeDB` Ops Manager watches for `MSSQLServerOpsRequest`.

6. When it finds one, it halts the `MSSQLServer` object so that the `KubeDB` Provisioner operator doesn't perform any operation on the `MSSQLServer` during the scaling process.

7. Then the KubeDB Ops-manager operator will update resources of the PetSet's Pods to reach desired state.

8. After successful updating of the resources of the PetSet's Pods, the `KubeDB` Ops Manager updates the `MSSQLServer` object resources to reflect the updated state.

9. After successful updating of the `MSSQLServer` resources, the `KubeDB` Ops Manager resumes the `MSSQLServer` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on updating resources of MSSQLServer database using vertical scaling operation.
