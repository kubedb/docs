---
title: Cassandra Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: cas-vertical-scaling-overview
    name: Overview
    parent: cas-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `Cassandra`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `Cassandra`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of Cassandra" src="/docs/images/day-2-operation/cassandra/verticalScale.svg">
<figcaption align="center">Fig: Vertical scaling process of Cassandra</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `Cassandra` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Cassandra` CR.

3. When the operator finds a `Cassandra` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `Cassandra` cluster, the user creates a `CassandraOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `CassandraOpsRequest` CR.

6. When it finds a `CassandraOpsRequest` CR, it halts the `Cassandra` object which is referred from the `CassandraOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Cassandra` object during the vertical scaling process.

7. Then the `KubeDB` Ops-manager operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Ops-manager operator updates the `Cassandra` object to reflect the updated state.

9. After the successful update  of the `Cassandra` resources, the `KubeDB` Ops-manager operator resumes the `Cassandra` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of Cassandra database using `CassandraOpsRequest` CRD.