---
title: Upgrading MariaDB Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-upgrading-overview
    name: Overview
    parent: guides-mariadb-upgrading
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrading MariaDB version Overview

This guide will give you an overview on how KubeDB Enterprise operator upgrade the version of `MariaDB` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)

## How Upgrade Process Works

The following diagram shows how KubeDB Enterprise operator used to upgrade the version of `MariaDB`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Upgrading Process of MariaDB" src="/docs/guides/mariadb/upgrading/overview/images/mdops-upgrade.jpeg">
<figcaption align="center">Fig: Upgrading Process of MariaDB</figcaption>
</figure>

The upgrading process consists of the following steps:

1. At first, a user creates a `MariaDB` Custom Resource (CR).

2. `KubeDB` Community operator watches the `MariaDB` CR.

3. When the operator finds a `MariaDB` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to upgrade the version of the `MariaDB` database the user creates a `MariaDBOpsRequest` CR with the desired version.

5. `KubeDB` Enterprise operator watches the `MariaDBOpsRequest` CR.

6. When it finds a `MariaDBOpsRequest` CR, it halts the `MariaDB` object which is referred from the `MariaDBOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `MariaDB` object during the upgrading process.  

7. By looking at the target version from `MariaDBOpsRequest` CR, `KubeDB` Enterprise operator updates the images of all the `StatefulSets`. After each image update, the operator performs some checks such as if the oplog is synced and database size is almost same or not.

8. After successfully updating the `StatefulSets` and their `Pods` images, the `KubeDB` Enterprise operator updates the image of the `MariaDB` object to reflect the updated state of the database.

9. After successfully updating of `MariaDB` object, the `KubeDB` Enterprise operator resumes the `MariaDB` object so that the `KubeDB` Community operator can resume its usual operations.

In the next doc, we are going to show a step by step guide on upgrading of a MariaDB database using upgrade operation.