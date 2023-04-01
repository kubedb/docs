---
title: Redis Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: rd-horizontal-scaling-overview
    name: Overview
    parent: rd-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Redis Horizontal Scaling

This guide will give an overview on how KubeDB Enterprise operator scales up or down of `Redis` cluster database for both the number of replicas and masters.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Enterprise operator scales up or down `Redis` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of Redis" src="/docs/images/day-2-operation/redis/rd-horizontal_scaling.svg">
<figcaption align="center">Fig: Horizontal scaling process of Redis</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Redis`/`RedisSentinel` Custom Resource (CR).

2. `KubeDB` Community operator watches the `Redis` and `RedisSentinel` CR.

3. When the operator finds a `Redis`/`RedisSentinel` CR, it creates required number of `StatefulSets` and related necessary stuff like appbinding, services, etc.

4. Then, in order to scale the number of replica or master for the `Redis` cluster database the user creates a `RedisOpsRequest` CR with desired information.

5. Then, in order to scale the number of replica for the `RedisSentinel` instance the user creates a `RedisSentinelOpsRequest` CR with desired information.

6. `KubeDB` Enterprise operator watches the `RedisOpsRequest` and `RedisSentinelOpsRequest` CR.

7. When it finds a `RedisOpsRequest` CR, it halts the `Redis` object which is referred from the `RedisOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `Redis` object during the scaling process.

8. When it finds a `RedisSentinelOpsRequest` CR, it halts the `RedisSentinel` object which is referred from the `RedisSentinelOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `RedisSentinel` object during the scaling process.

9. Then the Redis Ops-manager operator will scale the related StatefulSet Pods to reach the expected number of masters and/or replicas defined in the RedisOpsRequest or RedisSentinelOpsRequest CR.

10. After the successful scaling the replicas  of the related StatefulSet Pods, the KubeDB Ops-manager operator updates the number of replicas/masters in the Redis/RedisSentinel object to reflect the updated state.

11. After successfully updating of `Redis`/`RedisSentinel` object, the `KubeDB` Enterprise operator resumes the `Redis`/`RedisSentinel` object so that the `KubeDB` Community operator can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating of a Redis database using scale operation.