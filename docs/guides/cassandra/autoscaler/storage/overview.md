---
title: Cassandra Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: cas-storage-auto-scaling-overview
    name: Overview
    parent: cas-storage-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `Cassandraautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
  - [CassandraAutoscaler](/docs/guides/cassandra/concepts/cassandraautoscaler.md)
  - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Cassandra` database components. Open the image in a new tab to see the enlarged version.


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Cassandra` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Cassandra` CR.

3. When the operator finds a `Cassandra` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

- Each StatefulSet creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration.

4. Then, in order to set up storage autoscaling of the `Cassandra` cluster, the user creates a `CassandraAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `CassandraAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `CassandraOpsRequest` to expand the storage of the database. 
   
7. `KubeDB` Ops-manager operator watches the `CassandraOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the database component as specified on the `CassandraOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various Cassandra database components using `CassandraAutoscaler` CRD.
