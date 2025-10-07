---
title: ClickHouse Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: ch-auto-scaling-overview
    name: Overview
    parent: ch-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouse Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `ClickHouseautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseAutoscaler](/docs/guides/clickhouse/concepts/clickhouseautoscaler.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)

## How Compute Autoscaling Works

<figure align="center">
  <img alt="Compute AutoScale process of ClickHouse" src="/docs/images/day-2-operation/clickhouse/compute%20autoscaling.svg">
<figcaption align="center">Fig: Compute Auto Scale process of ClickHouse</figcaption>
</figure>

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `ClickHouse` database components. Open the image in a new tab to see the enlarged version.


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `ClickHouse` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `ClickHouse` CRO.

3. When the operator finds a `ClickHouse` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the of the `ClickHouse` cluster the user creates a `ClickHouseAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `ClickHouseAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `ClickHouseAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `ClickHouseOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `ClickHouseOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `ClickHouseOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of various ClickHouse database components using `ClickHouseAutoscaler` CRD.
