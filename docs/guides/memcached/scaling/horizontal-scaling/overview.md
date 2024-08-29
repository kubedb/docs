---
title: Memcached Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: mc-horizontal-scaling-overview
    name: Overview
    parent: mc-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Memcached Horizontal Scaling

This guide will give an overview on how KubeDB Ops Manager scales up or down of `Memcached` database for both the number of replicas and shards.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Memcached](/docs/guides/memcached/concepts/memcached.md)
  - [MemcachedOpsRequest](/docs/guides/memcached/concepts/memcached-opsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops Manager scales up or down `Memcached` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of Memcached" src="/docs/images/memcached/memcached-horizontal-scaling.png">
<figcaption align="center">Fig: Horizontal scaling process of Memcached</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Memcached` Custom Resource (CR).

2. `KubeDB` Community operator watches the `Memcached` CR.

3. When the operator finds a `Memcached` CR, it creates required number of `PetSets` and related necessary stuff like appbinding, services, etc.

4. Then, in order to scale the number of replica for the `Memcached` database the user creates a `MemcachedOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `MemcachedOpsRequest` CR.

6. When it finds a `MemcachedOpsRequest` CR, it halts the `Memcached` object which is referred from the `MemcachedOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `Memcached` object during the scaling process.

7. Then the Memcached Ops-manager operator will scale the related PetSet Pods to reach the expected number of replicas defined in the MemcachedOpsRequest CR.

8. After the successful scaling the replicas of the related PetSet Pods, the KubeDB Ops-manager operator updates the number of replicas in the Memcached object to reflect the updated state.

9. After successfully updating of `Memcached` object, the `KubeDB` Enterprise operator resumes the `Memcached` object so that the `KubeDB` Community operator can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating of a Memcached database using scale operation.

## Next Steps

- Learn how to horizontally scale [Memcached](/docs/guides/memcached/scaling/horizontal-scaling/memcached.md)