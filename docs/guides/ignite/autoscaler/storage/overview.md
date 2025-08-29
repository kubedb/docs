---
title: Ignite Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: ig-autoscaling-storage-overview
    name: Overview
    parent: ig-autoscaling-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ignite Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `Igniteautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteAutoscaler](/docs/guides/ignite/concepts/autoscaler.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Ignite` database components. Open the image in a new tab to see the enlarged version.


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Ignite` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Ignite` CR.

3. When the operator finds a `Ignite` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

- Each StatefulSet creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration.

4. Then, in order to set up storage autoscaling of the `Ignite` cluster, the user creates a `IgniteAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `IgniteAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `IgniteOpsRequest` to expand the storage of the database. 
   
7. `KubeDB` Ops-manager operator watches the `IgniteOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the database component as specified on the `IgniteOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various Ignite database components using `IgniteAutoscaler` CRD.
