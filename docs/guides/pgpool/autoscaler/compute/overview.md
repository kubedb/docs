---
title: Pgpool Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: pp-auto-scaling-overview
    name: Overview
    parent: pp-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Pgpool Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `pgpoolautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Pgpool](/docs/guides/pgpool/concepts/pgpool.md)
  - [PgpoolAutoscaler](/docs/guides/pgpool/concepts/autoscaler.md)
  - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Pgpool`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of Pgpool" src="/docs/images/day-2-operation/pgpool/compute-process.png">
<figcaption align="center">Fig: Compute Auto Scaling process of Pgpool</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Pgpool` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Pgpool` CRO.

3. When the operator finds a `Pgpool` CRO, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of `Pgpool`, the user creates a `PgpoolAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `PgpoolAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `PgpoolAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `PgpoolOpsRequest` CRO to scale the pgpool to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `PgpoolOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the pgpool vertically as specified on the `PgpoolOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of Pgpool using `PgpoolAutoscaler` CRD.
