---
title: ClickHouse Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: ch-storage-auto-scaling-overview
    name: Overview
    parent: ch-storage-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouse Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `ClickHouseAutoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseAutoscaler](/docs/guides/clickhouse/concepts/clickhouseautoscaler.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)

## How Storage Autoscaling Works

<figure align="center">
  <img alt="Storage AutoScale process of ClickHouse" src="/docs/images/day-2-operation/clickhouse/storage%20autoscaling.svg">
<figcaption align="center">Fig: Storage Auto Scale process of ClickHouse</figcaption>
</figure>

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `ClickHouse` database components. Open the image in a new tab to see the enlarged version.


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `ClickHouse` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `ClickHouse` CR.

3. When the operator finds a `ClickHouse` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

- Each StatefulSet creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration.

4. Then, in order to set up storage autoscaling of the `ClickHouse` cluster, the user creates a `ClickHouseAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `ClickHouseAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `ClickHouseOpsRequest` to expand the storage of the database.

7. `KubeDB` Ops-manager operator watches the `ClickHouseOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the database component as specified on the `ClickHouseOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various ClickHouse database components using `ClickHouseAutoscaler` CRD.
