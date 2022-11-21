---
title: MySQL Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-scaling-vertical-overview
    name: Overview
    parent: guides-mysql-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scaling MySQL

This guide will give you an overview of how KubeDB enterprise operator updates the resources(for example Memory and RAM etc.) of the `MySQL` database server.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/database/index.md)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest/index.md)

## How Vertical Scaling Process Works

The following diagram shows how the KubeDB enterprise operator used to update the resources of the `MySQL` database server. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/guides/mysql/scaling/vertical-scaling/overview/images/my-vertical_scaling.png">
<figcaption align="center">Fig: Vertical scaling process of MySQL</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `MySQL` cr.

2. `KubeDB` community operator watches for the `MySQL` cr.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secret, service, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `MySQL` database the user creates a `MySQLOpsRequest` cr.

5. `KubeDB` enterprise operator watches for `MySQLOpsRequest`.

6. When it finds one, it halts the `MySQL` object so that the `KubeDB` community operator doesn't perform any operation on the `MySQL` during the scaling process.  

7. Then the `KubeDB` enterprise operator will update resources of the StatefulSet replicas to reach the desired state.

8. After successful updating of the resources of the StatefulSet's replica, the `KubeDB` enterprise operator updates the `MySQL` object resources to reflect the updated state.

9. After successful updating of the `MySQL` resources, the `KubeDB` enterprise operator resumes the `MySQL` object so that the `KubeDB` community operator resumes its usual operations.

In the next doc, we are going to show a step by step guide on updating resources of MySQL database using vertical scaling operation.