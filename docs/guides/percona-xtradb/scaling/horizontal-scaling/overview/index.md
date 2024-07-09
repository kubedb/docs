---
title: PerconaXtraDB Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-scaling-horizontal-overview
    name: Overview
    parent: guides-perconaxtradb-scaling-horizontal
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB Horizontal Scaling

This guide will give an overview on how KubeDB Ops Manager scales up or down `PerconaXtraDB Cluster`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb/)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest/)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops Manager scales up or down `PerconaXtraDB` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of PerconaXtraDB" src="/docs/guides/percona-xtradb/scaling/horizontal-scaling/overview/images/horizontal-scaling.jpg">
<figcaption align="center">Fig: Horizontal scaling process of PerconaXtraDB</figcaption>
</figure>

The Horizontal scaling process consists of the following steps:

1. At first, a user creates a `PerconaXtraDB` Custom Resource (CR).

2. `KubeDB` Community operator watches the `PerconaXtraDB` CR.

3. When the operator finds a `PerconaXtraDB` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to scale the `PerconaXtraDB` database the user creates a `PerconaXtraDBOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `PerconaXtraDBOpsRequest` CR.

6. When it finds a `PerconaXtraDBOpsRequest` CR, it pauses the `PerconaXtraDB` object which is referred from the `PerconaXtraDBOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `PerconaXtraDB` object during the horizontal scaling process.  

7. Then the `KubeDB` Enterprise operator will scale the related PetSet Pods to reach the expected number of replicas defined in the `PerconaXtraDBOpsRequest` CR.

8. After the successfully scaling the replicas of the PetSet Pods, the `KubeDB` Enterprise operator updates the number of replicas in the `PerconaXtraDB` object to reflect the updated state.

9. After the successful scaling of the `PerconaXtraDB` replicas, the `KubeDB` Enterprise operator resumes the `PerconaXtraDB` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on horizontal scaling of PerconaXtraDB database using `PerconaXtraDBOpsRequest` CRD.
