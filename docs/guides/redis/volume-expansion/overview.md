---
title: Redis Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: rd-volume-expansion-overview
    name: Overview
    parent: rd-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis Volume Expansion

This guide will give an overview on how KubeDB Ops Manager expand the volume of `Redis`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops Manager expand the volumes of `Redis` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of Redis" src="/docs/images/day-2-operation/redis/rd-volume-expansion.svg">
<figcaption align="center">Fig: Volume Expansion process of Redis</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `Redis` Custom Resource (CR).

2. `KubeDB` Community operator watches the `Redis` CR.

3. When the operator finds a `Redis` CR, it creates required `PetSet` and related necessary stuff like secrets, services, etc.

4. The petSet creates Persistent Volumes according to the Volume Claim Template provided in the petset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, in order to expand the volume of the `Redis` database the user creates a `RedisOpsRequest` CR with desired information.

6. `KubeDB` Enterprise operator watches the `RedisOpsRequest` CR.

7. When it finds a `RedisOpsRequest` CR, it pauses the `Redis` object which is referred from the `RedisOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `Redis` object during the volume expansion process.

8. Then the `KubeDB` Enterprise operator will expand the persistent volume to reach the expected size defined in the `RedisOpsRequest` CR.

9. After the successful expansion of the volume of the related PetSet Pods, the `KubeDB` Enterprise operator updates the new volume size in the `Redis` object to reflect the updated state.

10. After the successful Volume Expansion of the `Redis`, the `KubeDB` Enterprise operator resumes the `Redis` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on Volume Expansion of various Redis database using `RedisOpsRequest` CRD.
