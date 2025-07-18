---
title: Cassandra Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: cas-horizontal-scaling-overview
    name: Overview
    parent: cas-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra Horizontal Scaling

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Cassandra` cluster replicas of various component such as Combined, Broker, Controller.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator scales up or down `Cassandra` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of Cassandra" src="/docs/images/day-2-operation/cassandra/cas-horizontal-scaling.svg">
<figcaption align="center">Fig: Horizontal scaling process of Cassandra</figcaption>
</figure>

The Horizontal scaling process consists of the following steps:

1. At first, a user creates a `Cassandra` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Cassandra` CR.

3. When the operator finds a `Cassandra` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to scale the various components (ie. ReplicaSet, Shard, ConfigServer, Mongos, etc.) of the `Cassandra` cluster, the user creates a `CassandraOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `CassandraOpsRequest` CR.

6. When it finds a `CassandraOpsRequest` CR, it halts the `Cassandra` object which is referred from the `CassandraOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Cassandra` object during the horizontal scaling process.

7. Then the `KubeDB` Ops-manager operator will scale the related PetSet Pods to reach the expected number of replicas defined in the `CassandraOpsRequest` CR.

8. After the successfully scaling the replicas of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the number of replicas in the `Cassandra` object to reflect the updated state.

9. After the successful scaling of the `Cassandra` replicas, the `KubeDB` Ops-manager operator resumes the `Cassandra` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on horizontal scaling of Cassandra cluster using `CassandraOpsRequest` CRD.