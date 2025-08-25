---
title: Hazelcast Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: hz-storage-auto-scaling-overview
    name: Overview
    parent: hz-storage-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Hazelcast Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `hazelcastautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
    - [HazelcastAutoscaler](/docs/guides/hazelcast/concepts/hazelcastautoscaler.md)
    - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Hazelcast` cluster components. Open the image in a new tab to see the enlarged version.

<figure align="center">
    <img alt="Storage Auto Scaling process of Hazelcast" src="/docs/images/day-2-operation/hazelcast/hz-storage-autoscaling.svg">
<figcaption align="center">Fig: Storage Auto Scaling process of Hazelcast</figcaption>
</figure>


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Hazelcast` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Hazelcast` CR.

3. When the operator finds a `Hazelcast` CR, it creates required number of `Statefulsets` and related necessary stuff like secrets, services, etc.

- Each Statefulset creates a Persistent Volume according to the Volume Claim Template provided in the Statefulset configuration.

4. Then, in order to set up storage autoscaling of the various components (ie. Combined, Node) of the `Hazelcast` cluster, the user creates a `HazelcastAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `HazelcastAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the clusters to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `HazelcastOpsRequest` to expand the storage of the database.

7. `KubeDB` Ops-manager operator watches the `HazelcastOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the cluster component as specified on the `HazelcastOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various Hazelcast cluster components using `HazelcastAutoscaler` CRD.
