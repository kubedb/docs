---
title: MariaDB Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: my-horizontal-scaling-overview
    name: Overview
    parent: my-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Horizontal Scaling Overview

This guide will give you an overview of how KubeDB enterprise operator scales up/down the number of members of a `MariaDB` group replication.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb.md)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB enterprise operator used to scale up the number of members of a `MariaDB` group replication. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/images/day-2-operation/mariadb/my-horizontal_scaling.png">
<figcaption align="center">Fig: Horizontal scaling process of MariaDB group replication</figcaption>
</figure>

The horizontal scaling process consists of the following steps:

1. At first, a user creates a `MariaDB` cr.

2. `KubeDB` community operator watches for the `MariaDB` cr.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secret, service, etc.

4. Then, in order to scale the cluster, the user creates a `MariaDBOpsRequest` cr with the desired number of members after scaling.

5. `KubeDB` enterprise operator watches for `MariaDBOpsRequest`.

6. When it finds one, it halts the `MariaDB` object so that the `KubeDB` community operator doesn't perform any operation on the `MariaDB` during the scaling process.  

7. Then the `KubeDB` enterprise operator will scale the StatefulSet replicas to reach the expected number of members for the group replication.

8. After successful scaling of the StatefulSet's replica, the `KubeDB` enterprise operator updates the `spec.replicas` field of `MariaDB` object to reflect the updated cluster state.

9. After successful scaling of the `MariaDB` replicas, the `KubeDB` enterprise operator resumes the `MariaDB` object so that the `KubeDB` community operator can resume its usual operations.

In the next doc, we are going to show a step by step guide on scaling of a MariaDB group replication using Horizontal Scaling.