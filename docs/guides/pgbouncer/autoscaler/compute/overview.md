---
title: PgBouncer Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: pb-auto-scaling-overview
    name: Overview
    parent: pb-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PgBouncer Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `pgbouncerautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerAutoscaler](/docs/guides/pgbouncer/concepts/autoscaler.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `PgBouncer`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of PgBouncer" src="/docs/images/day-2-operation/pgbouncer/autoscaling.png">
<figcaption align="center">Fig: Compute Auto Scaling process of PgBouncer</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `PgBouncer` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `PgBouncer` CRO.

3. When the operator finds a `PgBouncer` CRO, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of `PgBouncer`, the user creates a `PgBouncerAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `PgBouncerAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `PgBouncerAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `PgBouncerOpsRequest` CRO to scale the pgbouncer to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `PgBouncerOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the pgbouncer vertically as specified on the `PgBouncerOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of PgBouncer using `PgBouncerAutoscaler` CRD.
