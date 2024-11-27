---
title: PgBouncer Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: pb-horizontal-scaling-overview
    name: Overview
    parent: pb-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PgBouncer Horizontal Scaling

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `PgBouncer` replicas of PetSet.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator scales up or down `PgBouncer` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of PgBouncer" src="/docs/images/day-2-operation/pgbouncer/horizontal-scaling.png">
<figcaption align="center">Fig: Horizontal scaling process of PgBouncer</figcaption>
</figure>

The Horizontal scaling process consists of the following steps:

1. At first, a user creates a `PgBouncer` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `PgBouncer` CR.

3. When the operator finds a `PgBouncer` CR, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to scale the `PetSet` of the `PgBouncer` database the user creates a `PgBouncerOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `PgBouncerOpsRequest` CR.

6. When it finds a `PgBouncerOpsRequest` CR, it pauses the `PgBouncer` object which is referred from the `PgBouncerOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `PgBouncer` object during the horizontal scaling process.  

7. Then the `KubeDB` Ops-manager operator will scale the related PetSet Pods to reach the expected number of replicas defined in the `PgBouncerOpsRequest` CR.

8. After the successfully scaling the replicas of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the number of replicas in the `PgBouncer` object to reflect the updated state.

9. After the successful scaling of the `PgBouncer` replicas, the `KubeDB` Ops-manager operator resumes the `PgBouncer` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on horizontal scaling of PgBouncer using `PgBouncerOpsRequest` CRD.