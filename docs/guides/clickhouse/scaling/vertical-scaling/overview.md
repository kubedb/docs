---
title: ClickHouse Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: ch-vertical-scaling-overview
    name: Overview
    parent: ch-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouse Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `ClickHouse`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `ClickHouse`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of ClickHouse" src="/docs/images/day-2-operation/clickhouse/vertical-scaling.svg">
<figcaption align="center">Fig: Vertical scaling process of ClickHouse</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `ClickHouse` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `ClickHouse` CR.

3. When the operator finds a `ClickHouse` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `ClickHouse` cluster, the user creates a `ClickHouseOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ClickHouseOpsRequest` CR.

6. When it finds a `ClickHouseOpsRequest` CR, it halts the `ClickHouse` object which is referred from the `ClickHouseOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `ClickHouse` object during the vertical scaling process.

7. Then the `KubeDB` Ops-manager operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Ops-manager operator updates the `ClickHouse` object to reflect the updated state.

9. After the successful update  of the `ClickHouse` resources, the `KubeDB` Ops-manager operator resumes the `ClickHouse` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of ClickHouse database using `ClickHouseOpsRequest` CRD.