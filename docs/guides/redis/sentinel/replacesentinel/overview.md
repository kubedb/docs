---
title: Replace Sentinel Overview
menu:
  docs_{{ .version }}:
    identifier: replace-sentinel-overview
    name: Overview
    parent: rd-replace-sentinel
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Replace Sentinel Overview

This guide will give you an overview on how KubeDB Ops Manager replaces sentinel of a Redis instance.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisSentinel](/docs/guides/redis/concepts/redissentinel.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)

## How Replace Sentinel Process Works

The following diagram shows how KubeDB Ops Manager replaces sentinel of `Redis`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of MongoDB" src="/docs/images/day-2-operation/redis/replace-sentinel.svg">
<figcaption align="center">Fig: Replace Sentinel Process of Redis</figcaption>
</figure>

The replace sentinel process consists of the following steps:

1. At first, a user creates `Redis` and `RedisSentinel` Custom Resource (CR).

2. `KubeDB` Community operator watches the `Redis` and `RedisSentinel` CR.

3. When the operator finds a `Redis` CR, it creates a `PetSet` and related necessary stuff like appbinding, services, etc.

4. When the operator finds a `RedisSentinel` CR, it creates a `PetSet` and related necessary stuff like appbinding, services, etc.

5. Then, in order to replace the sentinel of the `Redis` database the user creates a new `RedisSentinel` object which will replace the old one and `RedisOpsRequest` CR with reference to the newly created sentinel.

6. `KubeDB` Enterprise operator watches the `RedisOpsRequest` CR.

7. When it finds a `RedisOpsRequest` CR, it pauses the `Redis` object which is referred from the `RedisOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `Redis` object during the updating process.  

8. By looking at the target sentinel reference from `RedisOpsRequest` CR, `KubeDB` Enterprise operator removes current sentinel and add desired sentinel to the `Redis` object. Then it may delete the old `RedisSentinel` object if it is orphaned and user requested to remove it.

9. After successfully updating `Redis` object, the `KubeDB` Enterprise operator resumes the `Redis` object so that the `KubeDB` Community operator can resume its usual operations.

In the next docs, we are going to show a step-by-step guide on replacing sentinel of Redis database components using `RedisOpsRequest` CRD.

## Next Steps

- Replace Sentinel of your database with a new [Sentinel](/docs/guides/redis/sentinel/replacesentinel/replace-sentinel.md)