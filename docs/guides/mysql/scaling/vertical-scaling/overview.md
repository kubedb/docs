---
title: MySQL Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: my-vertical-scaling-overview
    name: Overview
    parent: my-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# Vertical Scaling MySQL

This guide will show you how KubeDB enterprise operator used to update resources(for example Memory and RAM etc.) of nodes of `MySQL`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/concepts/databases/mysql.md)
  - [MySQLOpsRequest](/docs/concepts/day-2-operations/mysqlopsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB enterprise operator used to update the resources of the nodes of the `MySQL` group replication. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/images/day-2-operation/ops_req-vertical_scale%20.svg">
<figcaption align="center">Fig: Vertical scaling process of MySQL group replication</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `MySQL` crd.

2. `KubeDB` community operator watches for `MySQL` crd.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secret, service, etc.

4. When the user sees the `MySQL` object has arrived in the `ready` state, she creates a `MySQLOpsRequest` crd which specifies the `MySQL` object reference and information about container resources like `CPU`, `Memory` etc that will be applied for all container.

5. `KubeDB` enterprise operator watches for `MySQLOpsRequest`.

6. When it finds a specific one, it pauses the `MySQL` object so that the `KubeDB` community operator doesn't perform any operation on the `MySQL` during the scaling process.  

7. Then the `KubeDB` enterprise operator will update resources of the StatefulSet replicas to reach desired state.

8. After successful updating of the resources of the StatefulSet's replica, the `KubeDB` enterprise operator updates the `MySQL` object resources.

9. After successful updating of the `MySQL` resources, the `KubeDB` enterprise operator resumes the `MySQL` object so that the `KubeDB` community operator resumes it's actual operations.

At each of the above steps, the `KubeDB` enterprise operator updates the `status` section of the `MySQLOpsRequest`.