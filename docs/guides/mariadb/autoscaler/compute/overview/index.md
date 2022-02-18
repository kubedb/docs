---
title: MariaDB Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-autoscaling-compute-overview
    name: Overview
    parent: guides-mariadb-autoscaling-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# MariaDB Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `mariadbautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBAutoscaler](/docs/guides/mariadb/concepts/autoscaler)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `MariaDB` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Auto Scaling process of MariaDB" src="/docs/guides/mariadb/autoscaler/compute/overview/images/mdas-compute.jpg">
<figcaption align="center">Fig: Auto Scaling process of MariaDB</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `MariaDB` Custom Resource Object (CRO).

2. `KubeDB` Community operator watches the `MariaDB` CRO.

3. When the operator finds a `MariaDB` CRO, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the `MariaDB` database the user creates a `MariaDBAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `MariaDBAutoscaler` CRO.

6. `KubeDB` Autoscaler operator creates required number of Vertical Pod Autoscaler [VPA](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#intro) for different components of the database, as specified in the `mariadbautoscaler` CRO.

7. Then `KubeDB` Autoscaler operator continuously watches the VPA objects for recommendation.

8. If the VPA generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `MariaDBOpsRequest` CRO to scale the database to match the recommendation provided by the VPA object.

9. `KubeDB` Enterprise operator watches the `MariaDBOpsRequest` CRO.

10. Then the `KubeDB` Enterprise operator will scale the database component vertically as specified on the `MariaDBOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of MariaDB database using `MariaDBAutoscaler` CRD.
