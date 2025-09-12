---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: rd-rotate-auth-overview
    name: Overview
    parent: rd-rotate-auth-redis
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Redis

This guide will give an overview on how KubeDB Ops-manager operator Rotate Authentication configuration.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Redis](/docs/guides/redis/concepts/redis.md)
    - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)

## How Rotate Redis Authentication Configuration Process Works

[//]: # (The following diagram shows how KubeDB Ops-manager operator Rotate Authentication of a `Redis`. Open the image in a new tab to see the enlarged version.)

[//]: # ()
[//]: # (<figure align="center">)

[//]: # (  <img alt="Rotate Authentication process of Redis" src="/docs/images/day-2-operation/Redis/kf-rotate-auth.svg">)

[//]: # (<figcaption align="center">Fig: Rotate Auth process of Redis</figcaption>)

[//]: # (</figure>)

The authentication rotation process for Redis using KubeDB involves the following steps:

1. A user first creates a `Redis` Custom Resource Object (CRO).

2. The `KubeDB Provisioner operator` continuously watches for `Redis` CROs.

3. When the operator detects a `Redis` CR, it provisions the required `PetSets`, along with related resources such as secrets, services, and other dependencies.

4. To initiate authentication rotation, the user creates a `RedisOpsRequest` CR with the desired configuration.

5. The `KubeDB Ops-manager` operator watches for `RedisOpsRequest` CRs.

6. Upon detecting a `RedisOpsRequest`, the operator pauses the referenced `Redis` object, ensuring that the Provisioner
   operator does not perform any operations during the authentication rotation process.

7. The `Ops-manager` operator then updates the necessary configuration (such as credentials) based on the provided `RedisOpsRequest` specification.

8. After applying the updated configuration, the operator restarts all `Redis` Pods so they come up with the new authentication environment variables and settings.

9. Once the authentication rotation is completed successfully, the operator resumes the `Redis` object, allowing the Provisioner operator to continue its usual operations.

In the next section, we will walk you through a step-by-step guide to rotating Redis authentication using the `RedisOpsRequest` CRD.
