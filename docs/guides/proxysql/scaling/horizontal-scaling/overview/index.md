---
title: ProxySQL Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-scaling-horizontal-overview
    name: Overview
    parent: guides-proxysql-scaling-horizontal
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# ProxySQL Horizontal Scaling

This guide will give an overview on how KubeDB Enterprise operator scales up or down `ProxySQL Cluster`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/)
    - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest/)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Enterprise operator scales up or down `ProxySQL` components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of ProxySQL" src="/docs/guides/proxysql/scaling/horizontal-scaling/overview/images/horizontal-scaling.png">
<figcaption align="center">Fig: Horizontal scaling process of ProxySQL</figcaption>
</figure>

The Horizontal scaling process consists of the following steps:

1. At first, a user creates a `ProxySQL` Custom Resource (CR).

2. `KubeDB` Community operator watches the `ProxySQL` CR.

3. When the operator finds a `ProxySQL` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to scale the `ProxySQL` the user creates a `ProxySQLOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `ProxySQLOpsRequest` CR.

6. When it finds a `ProxySQLOpsRequest` CR, it pauses the `ProxySQL` object which is referred from the `ProxySQLOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `ProxySQL` object during the horizontal scaling process.

7. Then the `KubeDB` Enterprise operator will scale the related StatefulSet Pods to reach the expected number of replicas defined in the `ProxySQLOpsRequest` CR.

8. After the successfully scaling the replicas of the StatefulSet Pods, the `KubeDB` Enterprise operator updates the number of replicas in the `ProxySQL` object to reflect the updated state.

9. After the successful scaling of the `ProxySQL` replicas, the `KubeDB` Enterprise operator resumes the `ProxySQL` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on horizontal scaling of ProxySQL database using `ProxySQLOpsRequest` CRD.