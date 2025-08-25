---
title: Hazelcast Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: hz-horizontal-scaling-overview
    name: Overview
    parent: hz-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Hazelcast Horizontal Scaling

This guide will give an overview on how KubeDB Ops Manager scales up or down of `Hazelcast` database cluster members.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops Manager scales up or down `Hazelcast` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
    <img alt="Horizontal scaling process of Hazelcast" src="/docs/images/day-2-operation/hazelcast/hz-horizontal-scaling.svg">
<figcaption align="center">Fig: Horizontal scaling process of Hazelcast</figcaption>
</figure>

The scaling process consists of the following steps:

1. At first, a user creates a `Hazelcast` Custom Resource (CR).

2. `KubeDB` Community operator watches the `Hazelcast` CR.

3. When the operator finds a `Hazelcast` CR, it creates required number of `StatefulSets` and related necessary stuff like appbinding, services, etc.

4. Then, in order to scale the cluster, the user creates a `HazelcastOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `HazelcastOpsRequest` CR.

6. When it finds a `HazelcastOpsRequest` CR, it halts the `Hazelcast` object which is referred from the `HazelcastOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `Hazelcast` object during the scaling process.  

7. Then the `KubeDB` Enterprise operator will scale the related PetSets to reach the expected number of members defined in the `HazelcastOpsRequest` CR.

8. After the successfully scaling the StatefulSets replicas, the `KubeDB` Enterprise operator updates the number of members in the `Hazelcast` object to reflect the updated state of the database.

9. After the successful scaling of `Hazelcast` members, the `KubeDB` Enterprise operator resumes the `Hazelcast` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on horizontal scaling of Hazelcast database using `HazelcastOpsRequest` CRD.
