---
title: Upgrading MySQL Overview
menu:
  docs_{{ .version }}:
    identifier: my-upgrade-overview
    name: overview
    parent: my-upgrading-mysql
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# Upgrade MySQL version

This guide will show you how KubeDB enterprise operator upgrade MySQL version.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/concepts/databases/mysql.md)
  - [MySQLOpsRequest](/docs/concepts/day-2-operations/mysqlopsrequest.md)

## How Upgrade Process Works

The following diagram shows how KubeDB enterprise operator upgrade `MySQL` version. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/images/day-2-operation/ops_req-upgrade.svg">
<figcaption align="center">Fig: Upgrading Process of MySQL</figcaption>
</figure>

The upgrading process consists of the following steps:

1. At first, a user creates a `MySQL` crd.

2. `KubeDB` community operator watches for `MySQL` crd.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secret, service, etc.

4. When the user sees the `MySQL` object has arrived in the `ready` state, she creates a `MySQLOpsRequest` crd which specifies the `MySQL` object reference and target version.

5. `KubeDB` enterprise operator watches for `MySQLOpsRequest`.

6. When it finds a specific one, it pauses the `MySQL` object so that the `KubeDB` community operator doesn't perform any operation on the `MySQL` during the upgrading process.  

7. By looking at the target version from `MySQLOpsRequest` crd, `KubeDB` enterprise operator takes one of the following steps:
   - Update the images of the `StatefulSet` for upgrading between patch/minor versions.
   - Creates a new `StatefulSet` using targeted images for upgrading between major versions.

8. After successful upgradation of `StatefulSet` and its `Pod` images, the `KubeDB` enterprise operator updates the `MySQL` object images.

9. After successful upgradation of `MySQL` object, the `KubeDB` enterprise operator resumes the `MySQL` object so that the `KubeDB` community operator can resumes it's operations.

At each of the above steps, the `KubeDB` enterprise operator updates the `status` section of the `MySQLOpsRequest`.