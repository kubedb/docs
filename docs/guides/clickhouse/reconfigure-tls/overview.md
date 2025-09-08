---
title: Reconfiguring TLS/SSL
menu:
  docs_{{ .version }}:
    identifier: ch-reconfigure-tls-overview
    name: Overview
    parent: ch-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of ClickHouse

This guide will give an overview on how KubeDB Ops-manager operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of `ClickHouse`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)

## How Reconfiguring ClickHouse TLS Configuration Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures TLS of a `ClickHouse`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of ClickHouse" src="/docs/images/day-2-operation/clickhouse/reconfigureTLS.svg">
<figcaption align="center">Fig: Reconfiguring TLS process of ClickHouse</figcaption>
</figure>

The Reconfiguring ClickHouse TLS process consists of the following steps:

1. At first, a user creates a `ClickHouse` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `ClickHouse` CRO.

3. When the operator finds a `ClickHouse` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `ClickHouse` database the user creates a `ClickHouseOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ClickHouseOpsRequest` CR.

6. When it finds a `ClickHouseOpsRequest` CR, it pauses the `ClickHouse` object which is referred from the `ClickHouseOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `ClickHouse` object during the reconfiguring TLS process.

7. Then the `KubeDB` Ops-manager operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Ops-manager operator will restart all the Pods of the database so that they restart with the new TLS configuration defined in the `ClickHouseOpsRequest` CR.

9. After the successful reconfiguring of the `ClickHouse` TLS, the `KubeDB` Ops-manager operator resumes the `ClickHouse` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on reconfiguring TLS configuration of a ClickHouse database using `ClickHouseOpsRequest` CRD.