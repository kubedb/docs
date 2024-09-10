---
title: Memcached Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: mc-auto-scaling-overview
    name: Overview
    parent: mc-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Memcached Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `Memcachedautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Memcached](/docs/guides/memcached/concepts/memcached.md)
  - [MemcachedAutoscaler](/docs/guides/memcached/concepts/memcached-autoscaler.md)
  - [MemcachedOpsRequest](/docs/guides/memcached/concepts/memcached-opsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Memcached` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of Memcached" src="/docs/images/memcached/memcached-autoscaling-compute.png">
<figcaption align="center">Fig: Compute Auto Scaling process of Memcached</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, user creates a `Memcached` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Memcached` CRO.

3. When the operator finds a `Memcached` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the `Memcached` database the user creates a `MemcachedAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `MemcachedAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for the database, as specified in the `MemcachedAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `MemcachedOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `MemcachedOpsRequest` CRO.

9. Then the `KubeDB` ops-manager operator will scale the database component vertically as specified on the `MemcachedOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of various Memcached database components using `MemcachedAutoscaler` CRD.
