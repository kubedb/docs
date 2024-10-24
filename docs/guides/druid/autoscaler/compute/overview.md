---
title: Druid Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-druid-autoscaler-compute-overview
    name: Overview
    parent: guides-druid-autoscaler-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `kafkaautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Druid](/docs/guides/kafka/concepts/kafka.md)
  - [DruidAutoscaler](/docs/guides/kafka/concepts/kafkaautoscaler.md)
  - [DruidOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Druid` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of Druid" src="/docs/guides/druid/autoscaler/compute/images/compute-autoscaling.png">
<figcaption align="center">Fig: Compute Auto Scaling process of Druid</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Druid` Custom Resource Object (CRO).

2. `KubeDB` Provisioner operator watches the `Druid` CRO.

3. When the operator finds a `Druid` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the various components (ie. Coordinators, Overlords, Historicals, MiddleManagers, Brokers, Routers) of the `Druid` cluster the user creates a `DruidAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `DruidAutoscaler` CRO.

6. `KubeDB` Autoscaler operator generates recommendation using the modified version of kubernetes [official recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender) for different components of the database, as specified in the `DruidAutoscaler` CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `DruidOpsRequest` CRO to scale the database to match the recommendation generated.

8. `KubeDB` Ops-manager operator watches the `DruidOpsRequest` CRO.

9. Then the `KubeDB` Ops-manager operator will scale the database component vertically as specified on the `DruidOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of various Druid database components using `DruidAutoscaler` CRD.
