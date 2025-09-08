---
title: Reconfiguring ClickHouse
menu:
  docs_{{ .version }}:
    identifier: ch-reconfigure-overview
    name: Overview
    parent: ch-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring ClickHouse

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `ClickHouse` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)

## How Reconfiguring ClickHouse Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `ClickHouse` components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of ClickHouse" src="/docs/images/day-2-operation/clickhouse/reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of ClickHouse</figcaption>
</figure>

The Reconfiguring ClickHouse process consists of the following steps:

1. At first, a user creates a `ClickHouse` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `ClickHouse` CR.

3. When the operator finds a `ClickHouse` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `ClickHouse` database the user creates a `ClickHouseOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ClickHouseOpsRequest` CR.

6. When it finds a `ClickHouseOpsRequest` CR, it halts the `ClickHouse` object which is referred from the `ClickHouseOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `ClickHouse` object during the reconfiguring process.

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `ClickHouseOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `ClickHouseOpsRequest` CR.

9. After the successful reconfiguring of the `ClickHouse` components, the `KubeDB` Ops-manager operator resumes the `ClickHouse` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring ClickHouse components using `ClickHouseOpsRequest` CRD.