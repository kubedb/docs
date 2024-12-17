---
title: FerretDB Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: fr-auto-scaling-overview
    name: Overview
    parent: fr-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# FerretDB Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `FerretdbAutoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md)
    - [FerretDBAutoscaler](/docs/guides/ferretdb/concepts/autoscaler.md)
    - [FerretDBOpsRequest](/docs/guides/ferretdb/concepts/opsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `FerretDB`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of FerretDB" src="/docs/images/ferretdb/fr-compute-autoscaling.svg">
<figcaption align="center">Fig: Compute Auto Scaling process of FerretDB</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `FerretDB` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `FerretDB` CRO.

3. When the operator finds a `FerretDB` CRO, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of `FerretDB`, the user creates a `FerretDBAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `FerretDBAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `FerretDBAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `FerretDBOpsRequest` CRO to scale the ferretdb to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `FerretDBOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the ferretdb vertically as specified on the `FerretDBOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of FerretDB using `FerretDBAutoscaler` CRD.
