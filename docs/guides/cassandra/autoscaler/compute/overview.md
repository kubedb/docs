---
title: Cassandra Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: cas-auto-scaling-overview
    name: Overview
    parent: cas-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `Cassandraautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
  - [CassandraAutoscaler](/docs/guides/cassandra/concepts/cassandraautoscaler.md)
  - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Cassandra` database components. Open the image in a new tab to see the enlarged version.


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Cassandra` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Cassandra` CRO.

3. When the operator finds a `Cassandra` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the of the `Cassandra` cluster the user creates a `CassandraAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `CassandraAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `CassandraAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `CassandraOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `CassandraOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `CassandraOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of various Cassandra database components using `CassandraAutoscaler` CRD.
