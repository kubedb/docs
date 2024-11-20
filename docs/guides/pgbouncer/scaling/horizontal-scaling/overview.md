---
title: Pgpool Horizontal Scaling Overview
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

# Pgpool Horizontal Scaling

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Pgpool` replicas of PetSet.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Pgpool](/docs/guides/pgpool/concepts/pgpool.md)
  - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator scales up or down `Pgpool` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling process of Pgpool" src="/docs/images/day-2-operation/pgpool/pb-horizontal-scaling.png">
<figcaption align="center">Fig: Horizontal scaling process of Pgpool</figcaption>
</figure>

The Horizontal scaling process consists of the following steps:

1. At first, a user creates a `Pgpool` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Pgpool` CR.

3. When the operator finds a `Pgpool` CR, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to scale the `PetSet` of the `Pgpool` database the user creates a `PgpoolOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `PgpoolOpsRequest` CR.

6. When it finds a `PgpoolOpsRequest` CR, it pauses the `Pgpool` object which is referred from the `PgpoolOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Pgpool` object during the horizontal scaling process.  

7. Then the `KubeDB` Ops-manager operator will scale the related PetSet Pods to reach the expected number of replicas defined in the `PgpoolOpsRequest` CR.

8. After the successfully scaling the replicas of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the number of replicas in the `Pgpool` object to reflect the updated state.

9. After the successful scaling of the `Pgpool` replicas, the `KubeDB` Ops-manager operator resumes the `Pgpool` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on horizontal scaling of Pgpool using `PgpoolOpsRequest` CRD.