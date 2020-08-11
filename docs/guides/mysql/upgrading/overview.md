---
title: Upgrading MySQL Overview
menu:
  docs_{{ .version }}:
    identifier: my-upgrading-overview
    name: Overview
    parent: my-upgrading-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="Upgrading is an Enterprise feature of KubeDB. You must have KubeDB Enterprise operator installed to test this feature." >}}

# Upgrading MySQL version Overview

This guide will give you an overview on how KubeDB enterprise operator upgrade the version of `MySQL` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/concepts/databases/mysql.md)
  - [MySQLOpsRequest](/docs/concepts/day-2-operations/mysqlopsrequest.md)

## How Upgrade Process Works

The following diagram shows how KubeDB enterprise operator used to upgrade the version of `MySQL`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/images/day-2-operation/mysql/my-upgrading.png">
<figcaption align="center">Fig: Upgrading Process of MySQL</figcaption>
</figure>

The upgrading process consists of the following steps:

1. At first, a user creates a `MySQL` CR.

2. `KubeDB` community operator watches for the `MySQL` CR.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secret, service, etc.

4. Then, in order to upgrade the version of the `MySQL` database the user creates a `MySQLOpsRequest` cr with the desired version.

5. `KubeDB` enterprise operator watches for `MySQLOpsRequest`.

6. When it finds one, it pauses the `MySQL` object so that the `KubeDB` community operator doesn't perform any operation on the `MySQL` during the upgrading process.  

7. By looking at the target version from `MySQLOpsRequest` cr, `KubeDB` enterprise operator takes one of the following steps:
   - either update the images of the `StatefulSet` for upgrading between patch/minor versions.
   - or creates a new `StatefulSet` using targeted image for upgrading between major versions.

8. After successful upgradation of the `StatefulSet` and its `Pod` images, the `KubeDB` enterprise operator updates the image of the `MySQL` object to reflect the updated cluster state.

9. After successful upgradation of `MySQL` object, the `KubeDB` enterprise operator resumes the `MySQL` object so that the `KubeDB` community operator can resume it's usual operations.

In the next doc, we are going to show a step by step guide on upgrading of a MySQL database using upgrade operation.