---
title: Redis Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: rd-auto-scaling-overview
    name: Overview
    parent: rd-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `redisautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisAutoscaler](/docs/guides/redis/concepts/autoscaler.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Redis/Valkey` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of Redis" src="/docs/images/redis/rd-compute-autoscaling.svg">
<figcaption align="center">Fig: Compute Auto Scaling process of Redis</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Redis`/`RedisSentinel` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Redis`/`RedisSentinel` CRO.

3. When the operator finds a `Redis`/`RedisSentinel` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the `Redis` database the user creates a `RedisAutoscaler` CRO with desired configuration.

5. Then, in order to set up autoscaling of the `RedisSentinel` database the user creates a `RedisSentinelAutoscaler` CRO with desired configuration.

6. `KubeDB` Autoscaler operator watches the `RedisAutoscaler` && `RedisSentinelAutoscaler` CRO.

7. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for the database, as specified in the `RedisAutoscaler`/`RedisSentinelAutoscaler` CRO.

8. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `RedisOpsRequest`/`RedisSentinelOpsRequest` CRO to scale the database to match the recommendation generated.

9. `KubeDB` Ops-manager operator watches the `RedisOpsRequest`/`RedisSentinelOpsRequest` CRO.

10. Then the `KubeDB` ops-manager operator will scale the database component vertically as specified on the `RedisOpsRequest`/`RedisSentinelOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of various Redis database components using `RedisAutoscaler`/`RedisSentinelAutoscaler` CRD.
