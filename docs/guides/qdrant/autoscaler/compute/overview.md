---
title: Qdrant Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-compute-overview
    name: Overview
    parent: qdrant-autoscaler-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `QdrantAutoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [QdrantAutoscaler](/docs/guides/qdrant/concepts/autoscaler.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Qdrant` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Compute Auto Scaling process of Qdrant" src="/docs/guides/qdrant/images/qdrant-compute-autoscaling.png">
<figcaption align="center">Fig: Compute Auto Scaling process of Qdrant</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Qdrant` Custom Resource Object (CRO).

2. `KubeDB` Provisioner operator watches the `Qdrant` CRO.

3. When the operator finds a `Qdrant` CRO, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the `Qdrant` database the user creates a `QdrantAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `QdrantAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `QdrantAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `QdrantOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `QdrantOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `QdrantOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling of various Qdrant database using `QdrantAutoscaler` CRD.
