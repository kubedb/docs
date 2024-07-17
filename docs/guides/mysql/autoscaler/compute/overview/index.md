---
title: MySQL Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-autoscaling-compute-overview
    name: Overview
    parent: guides-mysql-autoscaling-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL Compute Resource Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database compute resources i.e. cpu and memory using `mysqlautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
  - [MySQLAutoscaler](/docs/guides/mysql/concepts/autoscaler)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)

## How Compute Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `MySQL` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Auto Scaling process of MySQL" src="/docs/guides/mysql/autoscaler/compute/overview/images/compute-autoscaling.jpg">
<figcaption align="center">Fig: Auto Scaling process of MySQL</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, the user creates a `MySQL` Custom Resource Object (CRO).

2. `KubeDB` Community operator watches the `MySQL` CRO.

3. When the operator finds a `MySQL` CRO, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to set up autoscaling of the CPU & Memory resources of the `MySQL` database the user creates a `MySQLAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `MySQLAutoscaler` CRO.

6. `KubeDB` Autoscaler operator utilizes the modified version of Kubernetes official [VPA-Recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg) for different components of the database, as specified in the `mysqlautoscaler` CRO.
It generates recommendations based on resource usages, & store them in the `status` section of the autoscaler CRO.

7. If the generated recommendation doesn't match the current resources of the database, then `KubeDB` Autoscaler operator creates a `MySQLOpsRequest` CRO to scale the database to match the recommendation provided by the VPA object.

8. `KubeDB Ops-Manager operator` watches the `MySQLOpsRequest` CRO.

9. Lastly, the `KubeDB Ops-Manager operator` will scale the database component vertically as specified on the `MySQLOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling of MySQL database using `MySQLAutoscaler` CRD.
