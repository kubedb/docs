---
title: SingleStore Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: sdb-auto-scaling-overview
    name: Overview
    parent: sdb-compute-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `singlestoreautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStoreAutoscaler](/docs/guides/singlestore/concepts/autoscaler.md)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `SingleStore` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of SingleStore" src="/docs/images/singlestore/compute-process.svg">
<figcaption align="center">Fig: Compute Auto Scaling process of SingleStore</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `SingleStore` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `SingleStore` CRO.

3. When the operator finds a `SingleStore` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the various components (ie. Aggregator, Leaf, Standalone) of the `SingleStore` database the user creates a `SingleStoreAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `SingleStoreAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `SingleStoreAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `SingleStoreOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `SingleStoreOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `SingleStoreOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of various SingleStore database components using `SingleStoreAutoscaler` CRD.